package node

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"opentoken-server/core"
	"opentoken-server/global"
	"opentoken-server/model"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var nodeUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NodeWebSocket(c *gin.Context) {
	token := readNodeToken(c)
	cred, err := findCredentialByToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	conn, err := nodeUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "websocket upgrade failed"})
		return
	}

	_ = conn.SetReadDeadline(time.Now().Add(15 * time.Second))
	_, raw, err := conn.ReadMessage()
	if err != nil {
		_ = conn.Close()
		return
	}

	var firstMsg core.NodeWireMessage
	if err := json.Unmarshal(raw, &firstMsg); err != nil || firstMsg.Type != "hello" {
		_ = conn.WriteJSON(gin.H{"type": "error", "message": "first message must be hello"})
		_ = conn.Close()
		return
	}

	var hello core.NodeHelloPayload
	if err := json.Unmarshal(firstMsg.Payload, &hello); err != nil {
		_ = conn.WriteJSON(gin.H{"type": "error", "message": "invalid hello payload"})
		_ = conn.Close()
		return
	}
	if len(hello.Models) == 0 {
		_ = conn.WriteJSON(gin.H{"type": "error", "message": "hello.models is required"})
		_ = conn.Close()
		return
	}

	nodeName := strings.TrimSpace(hello.NodeName)
	if nodeName == "" {
		nodeName = cred.Name
	}
	session := core.NewNodeSession(cred.ID, nodeName, hello.Models, conn)
	core.GLB_NODE_HUB.RegisterSession(session)
	updateLastSeenAsync(cred.ID)

	go nodeWriterLoop(session)
	_ = session.SendMessage(core.NodeWireMessage{Type: "hello_ack", Payload: mustJSON(map[string]interface{}{"node_id": cred.ID, "node_name": nodeName})})

	_ = conn.SetReadDeadline(time.Time{})
	for {
		_, data, readErr := conn.ReadMessage()
		if readErr != nil {
			break
		}
		var msg core.NodeWireMessage
		if err := json.Unmarshal(data, &msg); err != nil {
			continue
		}
		core.GLB_NODE_HUB.HandleNodeMessage(cred.ID, msg)
		if msg.Type == "ping" {
			_ = session.SendMessage(core.NodeWireMessage{Type: "pong"})
		}
	}

	core.GLB_NODE_HUB.RemoveSession(cred.ID, errors.New("node websocket disconnected"))
}

func nodeWriterLoop(session *core.NodeSession) {
	for msg := range session.Send {
		if err := session.Conn.WriteJSON(msg); err != nil {
			return
		}
	}
}

func readNodeToken(c *gin.Context) string {
	token := strings.TrimSpace(c.Query("token"))
	if token != "" {
		return token
	}
	auth := strings.TrimSpace(c.GetHeader("Authorization"))
	if auth == "" {
		return ""
	}
	parts := strings.SplitN(auth, " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return auth
}

func updateLastSeenAsync(nodeID uint) {
	now := time.Now()
	go func() {
		_ = global.GLB_DB.Model(&model.NodeCredential{}).Where("id = ?", nodeID).Update("last_seen_at", now).Error
	}()
}

func mustJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
