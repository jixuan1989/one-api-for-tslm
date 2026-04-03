package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/channeltype"
)

// DistributeSaas routes requests to the SaaS Backend channel.
// Unlike Distribute which selects channel by model, this always picks the SaasBackend channel.
func DistributeSaas() func(c *gin.Context) {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		userId := c.GetInt(ctxkey.Id)
		userGroup, _ := model.CacheGetUserGroup(userId)
		c.Set(ctxkey.Group, userGroup)
		c.Set(ctxkey.RequestModel, "saas-backend")

		channel, err := model.CacheGetRandomSatisfiedChannel(userGroup, "saas-backend", false)
		if err != nil {
			logger.Errorf(ctx, "No SaaS Backend channel available: %v", err)
			abortWithMessage(c, http.StatusServiceUnavailable,
				fmt.Sprintf("未找到可用的 SaaS Backend 渠道，请在管理后台配置"))
			return
		}

		SetupContextForSelectedChannel(c, channel, "saas-backend")

		// Override channel type to SaasBackend if it's not already
		if channel.Type != channeltype.SaasBackend {
			logger.Warnf(ctx, "Channel %d is not SaasBackend type, got %d", channel.Id, channel.Type)
		}

		c.Next()
	}
}
