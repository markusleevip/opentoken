package api

import (
	"github.com/gin-gonic/gin"

	"opentoken-node/api/v1/version"
)

// InitRoutes 初始化所有API路由
func InitRoutes(r *gin.Engine) {

	// 公共接口
	publicGroup := r.Group("/api/v1")
	{
		publicGroup.GET("/version", version.GetVersion)
	}

}
