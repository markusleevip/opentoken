package api

import (
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
		publicGroup.GET("/node/online", node.ListOnlineNodes)
	}

	// node 长连接
	r.GET("/ws/node", node.NodeWebSocket)

	// Consumer BaseURL（最小闭环）
	r.POST("/v1/chat/completions", gateway.ChatCompletions)
}
