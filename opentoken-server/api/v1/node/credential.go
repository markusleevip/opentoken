package node

import (
	"errors"
	"strings"

	"opentoken-server/core"
	"opentoken-server/domain/response"
	"opentoken-server/global"
	"opentoken-server/model"
	"opentoken-server/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// getEncryptKey 获取加密密钥（与 apikey 共用）
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

type createCredentialReq struct {
	Name string `json:"name"`
}

type createCredentialResp struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Token       string `json:"token"`
	TokenPrefix string `json:"token_prefix"`
	Enabled     bool   `json:"enabled"`
}

func CreateNodeCredential(c *gin.Context) {
	var req createCredentialReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.FailWithMessage("invalid request body", c)
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "node-" + utils.GenerateUID()
	}

	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		response.FailWithDetailed(err.Error(), "failed to generate token", c)
		return
	}

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

	cred := model.NodeCredential{
		Name:        name,
		TokenHash:   tokenHash,
		TokenPrefix: tokenPrefix,
		TokenEnc:    tokenEnc,
		Enabled:     true,
	}

	if err := global.GLB_DB.Create(&cred).Error; err != nil {
		response.FailWithDetailed(err.Error(), "failed to create node credential", c)
		return
	}

	response.OkWithData(createCredentialResp{
		ID:          cred.ID,
		Name:        cred.Name,
		Token:       token,
		TokenPrefix: cred.TokenPrefix,
		Enabled:     cred.Enabled,
	}, c)
}

func ListNodeCredentials(c *gin.Context) {
	credentials := make([]model.NodeCredential, 0)
	if err := global.GLB_DB.Order("id desc").Find(&credentials).Error; err != nil {
		response.FailWithDetailed(err.Error(), "failed to query credentials", c)
		return
	}
	response.OkWithData(credentials, c)
}

// GetFullToken 获取节点凭证的完整 Token
func GetFullToken(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.FailWithMessage("id is required", c)
		return
	}

	var cred model.NodeCredential
	if err := global.GLB_DB.First(&cred, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailWithMessage("credential not found", c)
			return
		}
		response.FailWithDetailed(err.Error(), "failed to query credential", c)
		return
	}

	// 解密 token
	token, err := utils.DecryptAES(cred.TokenEnc, getEncryptKey())
	if err != nil {
		response.FailWithDetailed(err.Error(), "failed to decrypt token", c)
		return
	}

	response.OkWithData(gin.H{"token": token}, c)
}

func ListOnlineNodes(c *gin.Context) {
	response.OkWithData(core.GLB_NODE_HUB.OnlineNodes(), c)
}

// DeleteNodeCredential 删除节点凭证
func DeleteNodeCredential(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.FailWithMessage("id is required", c)
		return
	}

	var cred model.NodeCredential
	if err := global.GLB_DB.First(&cred, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.FailWithMessage("credential not found", c)
			return
		}
		response.FailWithDetailed(err.Error(), "failed to query credential", c)
		return
	}

	// 检查节点是否在线（如果在线，删除后会被断开）
	// TODO: 可选 - 如果节点在线，可以先断开连接再删除

	if err := global.GLB_DB.Delete(&cred).Error; err != nil {
		response.FailWithDetailed(err.Error(), "failed to delete credential", c)
		return
	}

	response.OkWithMessage("deleted", c)
}

func findCredentialByToken(token string) (*model.NodeCredential, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("missing node token")
	}
	hash := utils.Sha256Hex(token)
	var cred model.NodeCredential
	err := global.GLB_DB.Where("token_hash = ? AND enabled = ?", hash, true).First(&cred).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid node token")
		}
		return nil, err
	}
	return &cred, nil
}
