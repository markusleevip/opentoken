package version

import (
	"opentoken-server/domain/response"
	"opentoken-server/global"
	"opentoken-server/model"

	"github.com/gin-gonic/gin"
)

// GetVersion 获取全局配置
func GetVersion(c *gin.Context) {
	version := model.Version{}
	global.GLB_DB.First(&version)
	response.OkWithData(version, c)

}
