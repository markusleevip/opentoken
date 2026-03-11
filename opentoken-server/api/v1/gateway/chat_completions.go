package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"opentoken-server/core"
	"opentoken-server/domain/response"

	"github.com/gin-gonic/gin"
)

func ChatCompletions(c *gin.Context) {
	var raw map[string]interface{}
	if err := c.ShouldBindJSON(&raw); err != nil {
		response.FailWithMessage("invalid request body", c)
		return
	}

	model, _ := raw["model"].(string)
	model = strings.TrimSpace(model)
	if model == "" {
		response.FailWithMessage("model is required", c)
		return
	}

	payload := core.NodeDispatchPayload{
		Model:    model,
		Messages: raw["messages"],
		Stream:   readStreamFlag(raw),
		Raw:      raw,
	}

	stream, nodeSession, err := core.GLB_NODE_HUB.DispatchToModel(payload)
	if err != nil {
		response.FailWithDetailed(err.Error(), "route failed", c)
		return
	}
	defer stream.Cancel()

	if payload.Stream {
		handleStreamingResponse(c, stream, nodeSession.NodeName)
		return
	}
	handleNonStreamingResponse(c, stream, model, nodeSession.NodeName)
}

func handleStreamingResponse(c *gin.Context, stream *core.PendingStream, nodeName string) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.String(http.StatusInternalServerError, "stream not supported")
		return
	}

	writeSSE := func(v interface{}) {
		buf, _ := json.Marshal(v)
		_, _ = c.Writer.Write([]byte("data: "))
		_, _ = c.Writer.Write(buf)
		_, _ = c.Writer.Write([]byte("\n\n"))
		flusher.Flush()
	}

	timeout := time.NewTimer(120 * time.Second)
	defer timeout.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-timeout.C:
			writeSSE(map[string]interface{}{"error": "gateway timeout"})
			_, _ = c.Writer.Write([]byte("data: [DONE]\n\n"))
			flusher.Flush()
			return
		case err, ok := <-stream.ErrCh:
			if !ok {
				return
			}
			writeSSE(map[string]interface{}{"error": err.Error()})
			_, _ = c.Writer.Write([]byte("data: [DONE]\n\n"))
			flusher.Flush()
			return
		case chunk, ok := <-stream.ChunkCh:
			if !ok {
				continue
			}
			writeSSE(openAIChunk(nodeName, chunk))
		case _, ok := <-stream.DoneCh:
			if !ok {
				return
			}
			_, _ = c.Writer.Write([]byte("data: [DONE]\n\n"))
			flusher.Flush()
			return
		}
	}
}

func handleNonStreamingResponse(c *gin.Context, stream *core.PendingStream, model string, nodeName string) {
	var sb strings.Builder
	timeout := time.NewTimer(120 * time.Second)
	defer timeout.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-timeout.C:
			response.FailWithMessage("gateway timeout", c)
			return
		case err, ok := <-stream.ErrCh:
			if ok && err != nil {
				response.FailWithDetailed(err.Error(), "node execution failed", c)
				return
			}
		case chunk, ok := <-stream.ChunkCh:
			if ok {
				sb.WriteString(chunk)
			}
		case _, ok := <-stream.DoneCh:
			if !ok {
				continue
			}
			response.OkWithData(map[string]interface{}{
				"model":   model,
				"node":    nodeName,
				"content": sb.String(),
			}, c)
			return
		}
	}
}

func readStreamFlag(raw map[string]interface{}) bool {
	v, ok := raw["stream"]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	if ok {
		return b
	}
	s, ok := v.(string)
	if ok {
		return strings.EqualFold(strings.TrimSpace(s), "true")
	}
	return false
}

func openAIChunk(nodeName string, delta string) map[string]interface{} {
	return map[string]interface{}{
		"id":      fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano()),
		"object":  "chat.completion.chunk",
		"created": time.Now().Unix(),
		"model":   nodeName,
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"delta": map[string]interface{}{"content": delta},
			},
		},
	}
}
