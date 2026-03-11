package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"opentoken-server/domain/response"
	"opentoken-server/global"
	"opentoken-server/utils"
)

const claimsContextKey = "auth_claims"

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" {
			c.JSON(http.StatusUnauthorized, response.Response{Code: 401, Data: nil, Msg: "missing Authorization header"})
			c.Abort()
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			c.JSON(http.StatusUnauthorized, response.Response{Code: 401, Data: nil, Msg: "invalid Authorization header"})
			c.Abort()
			return
		}

		claims, err := utils.ParseAndValidateJWT(global.GLB_CONFIG.JWT.SigningKey, strings.TrimSpace(parts[1]))
		if err != nil {
			c.JSON(http.StatusUnauthorized, response.Response{Code: 401, Data: nil, Msg: err.Error()})
			c.Abort()
			return
		}
		c.Set(claimsContextKey, claims)
		c.Next()
	}
}

func GetClaims(c *gin.Context) *utils.JWTClaims {
	v, ok := c.Get(claimsContextKey)
	if !ok {
		return nil
	}
	claims, _ := v.(*utils.JWTClaims)
	return claims
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := GetClaims(c)
		if claims == nil {
			c.JSON(http.StatusUnauthorized, response.Response{Code: 401, Data: nil, Msg: "unauthorized"})
			c.Abort()
			return
		}
		if !isAdminUsername(claims.Username) {
			c.JSON(http.StatusForbidden, response.Response{Code: 403, Data: nil, Msg: "admin only"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func isAdminUsername(username string) bool {
	username = strings.TrimSpace(username)
	if username == "" {
		return false
	}
	for _, u := range global.GLB_CONFIG.System.AdminUser {
		if strings.EqualFold(strings.TrimSpace(u), username) {
			return true
		}
	}
	return false
}
