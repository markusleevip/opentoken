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

type createCredentialReq struct {
	Name string `json:"name"`
}

type createCredentialResp struct {
	ID      uint   `json:"id"`
	Name    string `json:"name"`
	Token   string `json:"token"`
	Enabled bool   `json:"enabled"`
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

	cred := model.NodeCredential{
		Name:      name,
		TokenHash: utils.Sha256Hex(token),
		Enabled:   true,
	}

	if err := global.GLB_DB.Create(&cred).Error; err != nil {
		response.FailWithDetailed(err.Error(), "failed to create node credential", c)
		return
	}

	response.OkWithData(createCredentialResp{
		ID:      cred.ID,
		Name:    cred.Name,
		Token:   token,
		Enabled: cred.Enabled,
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

func ListOnlineNodes(c *gin.Context) {
	response.OkWithData(core.GLB_NODE_HUB.OnlineNodes(), c)
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
