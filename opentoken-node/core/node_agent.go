package core

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"opentoken-node/global"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

type nodeWireMessage struct {
	Type      string          `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type nodeHelloPayload struct {
	NodeName string                 `json:"node_name"`
	Models   []string               `json:"models"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type nodeDispatchPayload struct {
	Model    string      `json:"model"`
	Messages interface{} `json:"messages"`
	Stream   bool        `json:"stream"`
	Raw      interface{} `json:"raw,omitempty"`
}

type streamChunkPayload struct {
	Delta string `json:"delta"`
}

type streamEndPayload struct {
	FinishReason string `json:"finish_reason,omitempty"`
}

type streamErrorPayload struct {
	Error string `json:"error"`
}

type nodeAgent struct {
	sendMu sync.Mutex
	sendCh chan nodeWireMessage
}

// StartNodeAgent 启动 node 到 server 的长连接代理。
func StartNodeAgent() {
	cfg := global.GLB_CONFIG.NodeAgent
	if !cfg.Enabled {
		global.GLB_LOG.Info("node agent disabled")
		return
	}
	if strings.TrimSpace(cfg.ServerWSURL) == "" || strings.TrimSpace(cfg.Token) == "" {
		global.GLB_LOG.Warn("node agent missing server-ws-url or token, skip start")
		return
	}
	if len(cfg.Models) == 0 {
		global.GLB_LOG.Warn("node agent models empty, skip start")
		return
	}

	agent := &nodeAgent{sendCh: make(chan nodeWireMessage, 256)}
	go agent.runForever()
	global.GLB_LOG.Info("node agent started")
}

func (a *nodeAgent) runForever() {
	for {
		if err := a.runOnce(); err != nil {
			global.GLB_LOG.Error("node agent disconnected", zap.Error(err))
		}
		time.Sleep(backoffDuration())
	}
}

func (a *nodeAgent) runOnce() error {
	u, err := buildWSURL()
	if err != nil {
		return err
	}

	dialer := websocket.Dialer{
		HandshakeTimeout: handshakeTimeout(),
	}
	if global.GLB_CONFIG.UpstreamLLM.InsecureSkipTLS {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	conn, resp, err := dialer.Dial(u, nil)
	if err != nil {
		if resp != nil {
			return fmt.Errorf("dial ws failed: %w, status=%d", err, resp.StatusCode)
		}
		return fmt.Errorf("dial ws failed: %w", err)
	}
	defer conn.Close()

	hello := nodeHelloPayload{
		NodeName: strings.TrimSpace(global.GLB_CONFIG.NodeAgent.NodeName),
		Models:   global.GLB_CONFIG.NodeAgent.Models,
		Metadata: map[string]interface{}{"node_version": "v1", "ts": time.Now().Unix()},
	}
	helloBytes, _ := json.Marshal(hello)
	if err := conn.WriteJSON(nodeWireMessage{Type: "hello", Payload: helloBytes}); err != nil {
		return fmt.Errorf("send hello failed: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go a.writerLoop(ctx, conn)
	go a.pingLoop(ctx)

	for {
		_, data, err := conn.ReadMessage()
		if err != nil {
			return err
		}
		var msg nodeWireMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		switch msg.Type {
		case "hello_ack", "pong":
			continue
		case "dispatch":
			go a.handleDispatch(msg)
		default:
			global.GLB_LOG.Debug("unknown node wire message", zap.String("type", msg.Type))
		}
	}
}

func (a *nodeAgent) writerLoop(ctx context.Context, conn *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-a.sendCh:
			a.sendMu.Lock()
			err := conn.WriteJSON(msg)
			a.sendMu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

func (a *nodeAgent) pingLoop(ctx context.Context) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.send(nodeWireMessage{Type: "ping"})
		}
	}
}

func (a *nodeAgent) handleDispatch(msg nodeWireMessage) {
	if strings.TrimSpace(msg.RequestID) == "" {
		return
	}
	var payload nodeDispatchPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		a.sendStreamError(msg.RequestID, err)
		return
	}
	if err := validateDispatch(payload); err != nil {
		a.sendStreamError(msg.RequestID, err)
		return
	}

	if payload.Stream {
		a.executeStreaming(msg.RequestID, payload)
		return
	}
	a.executeNonStreaming(msg.RequestID, payload)
}

func (a *nodeAgent) executeStreaming(requestID string, payload nodeDispatchPayload) {
	reqBody, err := buildUpstreamBody(payload, true)
	if err != nil {
		a.sendStreamError(requestID, err)
		return
	}

	resp, err := doUpstreamRequest(reqBody)
	if err != nil {
		a.sendStreamError(requestID, err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 16*1024))
		global.GLB_LOG.Error("Upstream Error Response", zap.Int("status", resp.StatusCode), zap.String("body", string(body)))
		a.sendStreamError(requestID, fmt.Errorf("upstream status=%d body=%s", resp.StatusCode, string(body)))
		return
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}
		if !strings.HasPrefix(strings.ToLower(line), "data:") {
			continue
		}
		data := strings.TrimSpace(line[5:])
		if data == "[DONE]" {
			a.sendStreamEnd(requestID, "stop")
			return
		}
		delta := extractDeltaFromOpenAIChunk(data)
		global.GLB_LOG.Debug("Upstream Stream Chunk", zap.String("data", data), zap.String("delta", delta))
		if delta != "" {
			a.sendStreamChunk(requestID, delta)
		}
	}
	if err := scanner.Err(); err != nil {
		a.sendStreamError(requestID, err)
		return
	}
	a.sendStreamEnd(requestID, "stop")
}

func (a *nodeAgent) executeNonStreaming(requestID string, payload nodeDispatchPayload) {
	reqBody, err := buildUpstreamBody(payload, false)
	if err != nil {
		a.sendStreamError(requestID, err)
		return
	}

	resp, err := doUpstreamRequest(reqBody)
	if err != nil {
		a.sendStreamError(requestID, err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		global.GLB_LOG.Error("Upstream Error Response", zap.Int("status", resp.StatusCode), zap.String("body", string(body)))
		a.sendStreamError(requestID, fmt.Errorf("upstream status=%d body=%s", resp.StatusCode, string(body)))
		return
	}
	global.GLB_LOG.Info("Upstream Success Response", zap.Int("status", resp.StatusCode), zap.String("body", string(body)))

	text := extractContentFromOpenAIResponse(body)
	if text == "" {
		text = string(body)
	}
	a.sendStreamChunk(requestID, text)
	a.sendStreamEnd(requestID, "stop")
}

func (a *nodeAgent) send(msg nodeWireMessage) {
	select {
	case a.sendCh <- msg:
	default:
		global.GLB_LOG.Warn("node send channel full", zap.String("type", msg.Type))
	}
}

func (a *nodeAgent) sendStreamChunk(requestID string, delta string) {
	b, _ := json.Marshal(streamChunkPayload{Delta: delta})
	a.send(nodeWireMessage{Type: "stream_chunk", RequestID: requestID, Payload: b})
}

func (a *nodeAgent) sendStreamEnd(requestID string, reason string) {
	b, _ := json.Marshal(streamEndPayload{FinishReason: reason})
	a.send(nodeWireMessage{Type: "stream_end", RequestID: requestID, Payload: b})
}

func (a *nodeAgent) sendStreamError(requestID string, err error) {
	if err == nil {
		err = errors.New("unknown upstream error")
	}
	b, _ := json.Marshal(streamErrorPayload{Error: err.Error()})
	a.send(nodeWireMessage{Type: "stream_error", RequestID: requestID, Payload: b})
}

func doUpstreamRequest(body []byte) (*http.Response, error) {
	up := global.GLB_CONFIG.UpstreamLLM
	base := strings.TrimRight(strings.TrimSpace(up.BaseURL), "/")
	if base == "" {
		return nil, errors.New("upstream-llm.base-url is required")
	}
	chatPath := strings.TrimSpace(up.ChatPath)
	if chatPath == "" {
		chatPath = "/v1/chat/completions"
	}
	if !strings.HasPrefix(chatPath, "/") {
		chatPath = "/" + chatPath
	}

	timeout := time.Duration(up.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}

	tr := &http.Transport{}
	if up.InsecureSkipTLS {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	client := &http.Client{Timeout: timeout, Transport: tr}

	reqURL := base + chatPath
	global.GLB_LOG.Info("Sending upstream request",
		zap.String("url", reqURL),
		zap.String("body", string(body)),
	)

	req, err := http.NewRequest(http.MethodPost, reqURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if strings.TrimSpace(up.APIKey) != "" {
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(up.APIKey))
	}
	return client.Do(req)
}

func buildUpstreamBody(payload nodeDispatchPayload, forceStream bool) ([]byte, error) {
	body := make(map[string]interface{})

	if m, ok := payload.Raw.(map[string]interface{}); ok && len(m) > 0 {
		for k, v := range m {
			body[k] = v
		}
	}
	if len(body) == 0 {
		body["model"] = payload.Model
		body["messages"] = payload.Messages
	}

	mappedModel := mapModelName(payload.Model)
	if mappedModel == "" {
		mappedModel = payload.Model
	}
	body["model"] = mappedModel
	body["stream"] = forceStream

	return json.Marshal(body)
}

func mapModelName(model string) string {
	model = strings.TrimSpace(model)
	if model == "" {
		return ""
	}
	for k, v := range global.GLB_CONFIG.UpstreamLLM.ModelMap {
		if strings.EqualFold(strings.TrimSpace(k), model) {
			return strings.TrimSpace(v)
		}
	}
	return model
}

func extractDeltaFromOpenAIChunk(raw string) string {
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &obj); err != nil {
		return ""
	}
	choices, ok := obj["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return ""
	}
	first, ok := choices[0].(map[string]interface{})
	if !ok {
		return ""
	}
	delta, ok := first["delta"].(map[string]interface{})
	if !ok {
		return ""
	}
	content, _ := delta["content"].(string)
	return content
}

func extractContentFromOpenAIResponse(raw []byte) string {
	var obj map[string]interface{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		return ""
	}
	choices, ok := obj["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return ""
	}
	first, ok := choices[0].(map[string]interface{})
	if !ok {
		return ""
	}
	message, ok := first["message"].(map[string]interface{})
	if !ok {
		return ""
	}
	content, _ := message["content"].(string)
	return content
}

func validateDispatch(payload nodeDispatchPayload) error {
	if strings.TrimSpace(payload.Model) == "" {
		return errors.New("dispatch.model is required")
	}
	if payload.Messages == nil {
		return errors.New("dispatch.messages is required")
	}
	return nil
}

func buildWSURL() (string, error) {
	base := strings.TrimSpace(global.GLB_CONFIG.NodeAgent.ServerWSURL)
	if base == "" {
		return "", errors.New("node-agent.server-ws-url is required")
	}
	u, err := url.Parse(base)
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("token", strings.TrimSpace(global.GLB_CONFIG.NodeAgent.Token))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func backoffDuration() time.Duration {
	sec := global.GLB_CONFIG.NodeAgent.ReconnectBackoffSec
	if sec <= 0 {
		sec = 5
	}
	return time.Duration(sec) * time.Second
}

func handshakeTimeout() time.Duration {
	sec := global.GLB_CONFIG.NodeAgent.HandshakeTimeoutSec
	if sec <= 0 {
		sec = 10
	}
	return time.Duration(sec) * time.Second
}
