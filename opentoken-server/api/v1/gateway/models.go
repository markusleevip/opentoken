package gateway

import (
	"net/http"
	"sort"
	"time"

	"opentoken-server/core"

	"github.com/gin-gonic/gin"
)

// OpenAI compatible model struct
type Model struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Created    int64  `json:"created"`
	OwnedBy    string `json:"owned_by"`
}

type ModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// ListModels 返回所有在线节点支持的模型列表（OpenAI 兼容格式）
func ListModels(c *gin.Context) {
	nodes := core.GLB_NODE_HUB.OnlineNodes()

	// 使用 map 去重，收集所有唯一模型
	modelSet := make(map[string]struct{})
	for _, node := range nodes {
		for _, model := range node.Models {
			if model != "" {
				modelSet[model] = struct{}{}
			}
		}
	}

	// 转换为切片并排序
	models := make([]Model, 0, len(modelSet))
	for modelName := range modelSet {
		models = append(models, Model{
			ID:      modelName,
			Object:  "model",
			Created: time.Now().Unix(),
			OwnedBy: "opentoken",
		})
	}

	// 按模型名称排序
	sort.Slice(models, func(i, j int) bool {
		return models[i].ID < models[j].ID
	})

	c.JSON(http.StatusOK, ModelsResponse{
		Object: "list",
		Data:   models,
	})
}
