package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

type NodeWireMessage struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type NodeHelloPayload struct {
	NodeName string                 `json:"node_name"`
	Models   []string               `json:"models"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type NodeDispatchPayload struct {
	Model    string      `json:"model"`
	Messages interface{} `json:"messages"`
	Stream   bool        `json:"stream"`
	Raw      interface{} `json:"raw,omitempty"`
}

type NodeStreamChunkPayload struct {
	Delta string `json:"delta"`
}

type NodeStreamEndPayload struct {
	FinishReason string `json:"finish_reason,omitempty"`
}

type NodeStreamErrorPayload struct {
	Error string `json:"error"`
}

type OnlineNode struct {
	NodeID      uint      `json:"node_id"`
	NodeName    string    `json:"node_name"`
	Models      []string  `json:"models"`
	ConnectedAt time.Time `json:"connected_at"`
}

type NodeSession struct {
	NodeID      uint
	NodeName    string
	models      map[string]struct{}
	ConnectedAt time.Time
	Conn        *websocket.Conn
	Send        chan NodeWireMessage
}

func NewNodeSession(nodeID uint, nodeName string, models []string, conn *websocket.Conn) *NodeSession {
	m := make(map[string]struct{}, len(models))
	for _, model := range models {
		k := normalizeModelName(model)
		if k != "" {
			m[k] = struct{}{}
		}
	}
	return &NodeSession{
		NodeID:      nodeID,
		NodeName:    strings.TrimSpace(nodeName),
		models:      m,
		ConnectedAt: time.Now(),
		Conn:        conn,
		Send:        make(chan NodeWireMessage, 64),
	}
}

func (s *NodeSession) SupportsModel(model string) bool {
	_, ok := s.models[normalizeModelName(model)]
	return ok
}

func (s *NodeSession) Snapshot() OnlineNode {
	models := make([]string, 0, len(s.models))
	for m := range s.models {
		models = append(models, m)
	}
	return OnlineNode{
		NodeID:      s.NodeID,
		NodeName:    s.NodeName,
		Models:      models,
		ConnectedAt: s.ConnectedAt,
	}
}

func (s *NodeSession) SendMessage(msg NodeWireMessage) error {
	select {
	case s.Send <- msg:
		return nil
	case <-time.After(3 * time.Second):
		return errors.New("node send queue timeout")
	}
}

type pendingRequest struct {
	nodeID  uint
	chunkCh chan string
	doneCh  chan NodeStreamEndPayload
	errCh   chan error
}

type PendingStream struct {
	RequestID string
	ChunkCh   <-chan string
	DoneCh    <-chan NodeStreamEndPayload
	ErrCh     <-chan error
	hub       *NodeHub
}

func (p *PendingStream) Cancel() {
	if p == nil || p.hub == nil {
		return
	}
	p.hub.FailRequest(p.RequestID, errors.New("request cancelled by server"))
}

type NodeHub struct {
	mu       sync.RWMutex
	sessions map[uint]*NodeSession
	pending  map[string]*pendingRequest
	seq      uint64
}

var GLB_NODE_HUB = NewNodeHub()

func NewNodeHub() *NodeHub {
	return &NodeHub{
		sessions: map[uint]*NodeSession{},
		pending:  map[string]*pendingRequest{},
	}
}

func (h *NodeHub) RegisterSession(session *NodeSession) {
	h.mu.Lock()
	old := h.sessions[session.NodeID]
	h.sessions[session.NodeID] = session
	h.mu.Unlock()

	if old != nil {
		_ = old.Conn.Close()
		close(old.Send)
	}
}

func (h *NodeHub) RemoveSession(nodeID uint, reason error) {
	h.mu.Lock()
	session, ok := h.sessions[nodeID]
	if ok {
		delete(h.sessions, nodeID)
	}
	for reqID, pending := range h.pending {
		if pending.nodeID == nodeID {
			err := reason
			if err == nil {
				err = errors.New("node disconnected")
			}
			h.closePendingLocked(reqID, pending, err, nil)
		}
	}
	h.mu.Unlock()

	if ok {
		_ = session.Conn.Close()
		close(session.Send)
	}
}

func (h *NodeHub) OnlineNodes() []OnlineNode {
	h.mu.RLock()
	defer h.mu.RUnlock()
	out := make([]OnlineNode, 0, len(h.sessions))
	for _, s := range h.sessions {
		out = append(out, s.Snapshot())
	}
	return out
}

func (h *NodeHub) SelectNodeByModel(model string) (*NodeSession, error) {
	model = normalizeModelName(model)
	if model == "" {
		return nil, errors.New("model is required")
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, s := range h.sessions {
		if s.SupportsModel(model) {
			return s, nil
		}
	}
	return nil, fmt.Errorf("no online node supports model: %s", model)
}

func (h *NodeHub) DispatchToModel(payload NodeDispatchPayload) (*PendingStream, *NodeSession, error) {
	session, err := h.SelectNodeByModel(payload.Model)
	if err != nil {
		return nil, nil, err
	}

	requestID := h.nextRequestID()
	pending := &pendingRequest{
		nodeID:  session.NodeID,
		chunkCh: make(chan string, 64),
		doneCh:  make(chan NodeStreamEndPayload, 1),
		errCh:   make(chan error, 1),
	}

	h.mu.Lock()
	h.pending[requestID] = pending
	h.mu.Unlock()

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		h.FailRequest(requestID, err)
		return nil, nil, err
	}

	wire := NodeWireMessage{Type: "dispatch", RequestID: requestID, Payload: payloadBytes}
	select {
	case session.Send <- wire:
		return &PendingStream{
			RequestID: requestID,
			ChunkCh:   pending.chunkCh,
			DoneCh:    pending.doneCh,
			ErrCh:     pending.errCh,
			hub:       h,
		}, session, nil
	case <-time.After(5 * time.Second):
		h.FailRequest(requestID, errors.New("dispatch timeout"))
		return nil, nil, errors.New("node dispatch timeout")
	}
}

func (h *NodeHub) HandleNodeMessage(nodeID uint, msg NodeWireMessage) {
	switch msg.Type {
	case "stream_chunk":
		h.handleStreamChunk(msg)
	case "stream_end":
		h.handleStreamEnd(msg)
	case "stream_error":
		h.handleStreamError(msg)
	case "ping":
		_ = nodeID
	default:
		_ = nodeID
	}
}

func (h *NodeHub) FailRequest(requestID string, err error) {
	h.mu.Lock()
	pending, ok := h.pending[requestID]
	if ok {
		h.closePendingLocked(requestID, pending, err, nil)
	}
	h.mu.Unlock()
}

func (h *NodeHub) handleStreamChunk(msg NodeWireMessage) {
	if strings.TrimSpace(msg.RequestID) == "" {
		return
	}
	var payload NodeStreamChunkPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.FailRequest(msg.RequestID, err)
		return
	}

	h.mu.RLock()
	pending, ok := h.pending[msg.RequestID]
	h.mu.RUnlock()
	if !ok {
		return
	}
	select {
	case pending.chunkCh <- payload.Delta:
	default:
	}
}

func (h *NodeHub) handleStreamEnd(msg NodeWireMessage) {
	if strings.TrimSpace(msg.RequestID) == "" {
		return
	}
	payload := NodeStreamEndPayload{}
	if len(msg.Payload) > 0 {
		_ = json.Unmarshal(msg.Payload, &payload)
	}

	h.mu.Lock()
	pending, ok := h.pending[msg.RequestID]
	if ok {
		h.closePendingLocked(msg.RequestID, pending, nil, &payload)
	}
	h.mu.Unlock()
}

func (h *NodeHub) handleStreamError(msg NodeWireMessage) {
	if strings.TrimSpace(msg.RequestID) == "" {
		return
	}
	payload := NodeStreamErrorPayload{Error: "node stream error"}
	if len(msg.Payload) > 0 {
		_ = json.Unmarshal(msg.Payload, &payload)
	}
	h.FailRequest(msg.RequestID, errors.New(strings.TrimSpace(payload.Error)))
}

func (h *NodeHub) closePendingLocked(requestID string, pending *pendingRequest, err error, end *NodeStreamEndPayload) {
	delete(h.pending, requestID)
	if end != nil {
		select {
		case pending.doneCh <- *end:
		default:
		}
	}
	if err != nil {
		select {
		case pending.errCh <- err:
		default:
		}
	}
	close(pending.chunkCh)
	close(pending.doneCh)
	close(pending.errCh)
}

func (h *NodeHub) nextRequestID() string {
	seq := atomic.AddUint64(&h.seq, 1)
	return fmt.Sprintf("req_%d_%d", time.Now().UnixMilli(), seq)
}

func normalizeModelName(model string) string {
	return strings.ToLower(strings.TrimSpace(model))
}
