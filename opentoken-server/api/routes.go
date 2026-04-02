package api

import (
	"opentoken-server/api/v1/apikey"
	"opentoken-server/api/v1/gateway"
	"opentoken-server/api/v1/node"
	"opentoken-server/api/v1/version"

	"github.com/gin-gonic/gin"
)

// InitRoutes 初始化所有API路由
func InitRoutes(r *gin.Engine) {
	// 公共接口
	publicGroup := r.Group("/api/v1")
	{
		publicGroup.GET("/version", version.GetVersion)
		publicGroup.POST("/node/credentials", node.CreateNodeCredential)
		publicGroup.GET("/node/credentials", node.ListNodeCredentials)
		publicGroup.GET("/node/credentials/:id/token", node.GetFullToken)
		publicGroup.DELETE("/node/credentials/:id", node.DeleteNodeCredential)
		publicGroup.GET("/node/online", node.ListOnlineNodes)

		// API Key 管理
		publicGroup.GET("/apikeys", apikey.ListAPIKeys)
		publicGroup.POST("/apikeys", apikey.CreateAPIKey)
		publicGroup.GET("/apikeys/:id/token", apikey.GetFullToken)
		publicGroup.PUT("/apikeys/:id", apikey.UpdateAPIKey)
		publicGroup.DELETE("/apikeys/:id", apikey.DeleteAPIKey)
	}

	// node 长连接
	r.GET("/ws/node", node.NodeWebSocket)

	// Consumer BaseURL（最小闭环）
	r.GET("/v1/models", gateway.ListModels)
	r.POST("/v1/chat/completions", gateway.ChatCompletions)
}
