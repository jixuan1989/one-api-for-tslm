package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/blacklist"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/model"
)

// SessionTokenAuth authenticates relay requests using either:
//  1. Bearer token (standard TokenAuth behavior for SDK/API calls)
//  2. Session cookie (for web frontend — automatically uses the user's system token for billing)
//
// This prevents exposing system tokens to the frontend while still billing correctly.
func SessionTokenAuth() func(c *gin.Context) {
	return func(c *gin.Context) {
		key := c.Request.Header.Get("Authorization")
		key = strings.TrimPrefix(key, "Bearer ")

		// If Bearer token provided, delegate to standard TokenAuth
		if key != "" && key != "session" {
			TokenAuth()(c)
			return
		}

		// No Bearer token — try session cookie
		session := sessions.Default(c)
		userId := session.Get("id")
		status := session.Get("status")
		if userId == nil {
			abortWithMessage(c, http.StatusUnauthorized, "未登录，请先登录或提供 API 令牌")
			return
		}
		uid := userId.(int)
		if status != nil && status.(int) == model.UserStatusDisabled || blacklist.IsUserBanned(uid) {
			abortWithMessage(c, http.StatusForbidden, "用户已被封禁")
			return
		}

		// Determine request model
		requestModel, _ := getRequestModel(c)
		c.Set(ctxkey.RequestModel, requestModel)

		// Find user's system token for this model
		token, err := model.GetUserSystemTokenForModel(uid, requestModel)
		if err != nil {
			abortWithMessage(c, http.StatusForbidden, err.Error())
			return
		}

		c.Set(ctxkey.Id, uid)
		c.Set(ctxkey.TokenId, token.Id)
		c.Set(ctxkey.TokenName, token.Name)
		c.Next()
	}
}
