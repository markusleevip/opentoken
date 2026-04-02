package apikey

import (
	"errors"
	"strings"
	"time"

	"opentoken-server/domain/response"
	"opentoken-server/global"
	"opentoken-server/model"
	"opentoken-server/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// getEncryptKey 获取加密密钥
func getEncryptKey() []byte {
	key := global.GLB_CONFIG.Security.APIKeyEncryptKey
	if key == "" {
		// 使用默认密钥（32字节）
		key = "opentoken-default-encrypt-key-32"
	}
	// 确保密钥长度为 32 字节（AES-256）
	if len(key) < 32 {
		key = key + strings.Repeat("0", 32-len(key))
	}
	if len(key) > 32 {
		key = key[:32]
	}
	return []byte(key)
}

type createAPIKeyReq struct {
	Name        string `json:"name" binding:"required,max=64"`
	Description string `json:"description" binding:"max=255"`
}

type createAPIKeyResp struct {
	ID          uint      `json:"id"`
	Name        string    `json:"name"`
	Token       string    `json:"token"`
	TokenPrefix string    `json:"token_prefix"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// CreateAPIKey 创建新的 API Key
func CreateAPIKey(c *gin.Context) {
	var req createAPIKeyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request body: "+err.Error(), c)
		return
	}

	// 生成随机 token (32 bytes = 43 chars base64)
	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		response.FailWithDetailed(err.Error(), "failed to generate token", c)
		return
	}

	// 添加 sk- 前缀
	token = "sk-" + token

	// 计算哈希和前缀
	tokenHash := utils.Sha256Hex(token)
	tokenPrefix := token
	if len(token) > 16 {
		tokenPrefix = token[:16] + "..."
	}

	// 加密存储完整 token
	tokenEnc, err := utils.EncryptAES(token, getEncryptKey())
	if err != nil {
		response.FailWithDetailed(err.Error(), "failed to encrypt token", c)
		return
	}

	apiKey := model.APIKey{
		Name:        strings.TrimSpace(req.Name),
		TokenHash:   tokenHash,
		TokenPrefix: tokenPrefix,
		TokenEnc:    tokenEnc,
		Enabled:     true,
		Description: strings.TrimSpace(req.Description),
	}

	if err := global.GLB_DB.Create(&apiKey).Error; err != nil {
		response.FailWithDetailed(err.Error(), "failed to create api key", c)
		return
	}

	response.OkWithData(createAPIKeyResp{
		ID:          apiKey.ID,
		Name:        apiKey.Name,
		Token:       token,
		TokenPrefix: apiKey.TokenPrefix,
		Enabled:     apiKey.Enabled,
		Description: apiKey.Description,
		CreatedAt:   apiKey.CreatedAt,
	}, c)
}

type apiKeyListItem struct {
	ID          uint       `json:"id"`
	Name        string     `json:"name"`
	TokenPrefix string     `json:"token_prefix"`
	Enabled     bool       `json:"enabled"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
}

// ListAPIKeys 列出所有 API Key（不包含完整 token）
func ListAPIKeys(c *gin.Context) {
	var keys []model.APIKey
	if err := global.GLB_DB.Order("id desc").Find(&keys).Error; err != nil {
		response.FailWithDetailed(err.Error(), "failed to query api keys", c)
		return
	}

	items := make([]apiKeyListItem, 0, len(keys))
	for _, k := range keys {
		items = append(items, apiKeyListItem{
			ID:          k.ID,
			Name:        k.Name,
			TokenPrefix: k.TokenPrefix,
			Enabled:     k.Enabled,
			Description: k.Description,
			CreatedAt:   k.CreatedAt,
			LastUsedAt:  k.LastUsedAt,
		})
	}

	response.OkWithData(items, c)
}

type updateAPIKeyReq struct {
	Name        string `json:"name" binding:"max=64"`
	Description string `json:"description" binding:"max=255"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

// UpdateAPIKey 更新 API Key 信息
func UpdateAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.FailWithMessage("id is required", c)
		return
	}

	var req updateAPIKeyReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request body: "+err.Error(), c)
		return
	}

	var apiKey model.APIKey
	if err := global.GLB_DB.First(&apiKey, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailWithMessage("api key not found", c)
			return
		}
		response.FailWithDetailed(err.Error(), "failed to query api key", c)
		return
	}

	// 更新字段
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = strings.TrimSpace(req.Name)
	}
	if req.Description != "" {
		updates["description"] = strings.TrimSpace(req.Description)
	}
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}

	if len(updates) > 0 {
		if err := global.GLB_DB.Model(&apiKey).Updates(updates).Error; err != nil {
			response.FailWithDetailed(err.Error(), "failed to update api key", c)
			return
		}
	}

	response.OkWithData(apiKeyListItem{
		ID:          apiKey.ID,
		Name:        apiKey.Name,
		TokenPrefix: apiKey.TokenPrefix,
		Enabled:     apiKey.Enabled,
		Description: apiKey.Description,
		CreatedAt:   apiKey.CreatedAt,
		LastUsedAt:  apiKey.LastUsedAt,
	}, c)
}

// DeleteAPIKey 删除 API Key
func DeleteAPIKey(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.FailWithMessage("id is required", c)
		return
	}

	var apiKey model.APIKey
	if err := global.GLB_DB.First(&apiKey, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailWithMessage("api key not found", c)
			return
		}
		response.FailWithDetailed(err.Error(), "failed to query api key", c)
		return
	}

	if err := global.GLB_DB.Delete(&apiKey).Error; err != nil {
		response.FailWithDetailed(err.Error(), "failed to delete api key", c)
		return
	}

	response.OkWithMessage("deleted", c)
}

// FindAPIKeyByToken 根据 token 查找 API Key（内部使用）
func FindAPIKeyByToken(token string) (*model.APIKey, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("missing api key")
	}

	// 移除 Bearer 前缀
	if strings.HasPrefix(strings.ToLower(token), "bearer ") {
		token = token[7:]
	}

	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("invalid api key format")
	}

	tokenHash := utils.Sha256Hex(token)
	var apiKey model.APIKey
	err := global.GLB_DB.Where("token_hash = ? AND enabled = ?", tokenHash, true).First(&apiKey).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid api key")
		}
		return nil, err
	}

	return &apiKey, nil
}

// UpdateAPIKeyLastUsed 更新 API Key 最后使用时间（异步）
func UpdateAPIKeyLastUsed(apiKeyID uint) {
	now := time.Now()
	go func() {
		_ = global.GLB_DB.Model(&model.APIKey{}).Where("id = ?", apiKeyID).Update("last_used_at", now).Error
	}()
}

// GetFullToken 获取 API Key 的完整 Token
type getFullTokenResp struct {
	Token string `json:"token"`
}

func GetFullToken(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.FailWithMessage("id is required", c)
		return
	}

	var apiKey model.APIKey
	if err := global.GLB_DB.First(&apiKey, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailWithMessage("api key not found", c)
			return
		}
		response.FailWithDetailed(err.Error(), "failed to query api key", c)
		return
	}

	// 解密 token
	token, err := utils.DecryptAES(apiKey.TokenEnc, getEncryptKey())
	if err != nil {
		response.FailWithDetailed(err.Error(), "failed to decrypt token", c)
		return
	}

	response.OkWithData(getFullTokenResp{Token: token}, c)
}
