package main

import (
	"fmt"
	"log"
	"opentoken-node/api"
	"opentoken-node/core"
	"opentoken-node/global"
	"opentoken-node/initialize"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	initializeSystem()
	// 初始化Gin引擎
	r := gin.Default()

	// 基础健康检查路由
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	// 初始化API路由
	initRoutes(r)

	// 启动 node -> server 代理连接
	core.StartNodeAgent()

	// 托管前端静态资源（用于Windows可交付版本）
	initialize.RegisterFrontendStaticRoutes(r)

	// 启动服务
	host := strings.TrimSpace(global.GLB_CONFIG.System.Host)
	address := fmt.Sprintf(":%d", global.GLB_CONFIG.System.Port)
	if host != "" {
		address = fmt.Sprintf("%s:%d", host, global.GLB_CONFIG.System.Port)
	}
	fmt.Printf("Starting server at %s\n", address)
	log.Fatal(r.Run(address))

}

func initializeSystem() {
	global.GLB_VP = core.Viper() // 初始化Viper
	global.GLB_LOG = core.Zap()  // 初始化zap日志库
	zap.ReplaceGlobals(global.GLB_LOG)
	global.GLB_DB = initialize.Gorm() // gorm连接数据库
	if global.GLB_DB != nil {
		initialize.RegisterTables() // 初始化表
	}

}

// 路由初始化函数
func initRoutes(r *gin.Engine) {
	// 初始化所有API路由
	api.InitRoutes(r)
}
