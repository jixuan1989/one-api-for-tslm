package controller

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	dbmodel "github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

// RelaySaasHelper handles business requests forwarded to SaaS Backend.
// Only write operations (POST/PUT/DELETE) are billed; GET requests are free.
func RelaySaasHelper(c *gin.Context) *model.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := meta.GetByContext(c)
	meta.OriginModelName = "saas-backend"
	meta.ActualModelName = "saas-backend"

	isWrite := c.Request.Method != http.MethodGet

	var preConsumedQuota int64
	var ratio, modelRatio, groupRatio float64

	if isWrite {
		modelRatio = billingratio.GetModelRatio("saas-backend", meta.ChannelType)
		groupRatio = billingratio.GetGroupRatio(meta.Group)
		ratio = modelRatio * groupRatio
		preConsumedQuota = int64(math.Ceil(ratio))
		if preConsumedQuota <= 0 {
			preConsumedQuota = 1
		}

		userQuota, err := dbmodel.CacheGetUserQuota(ctx, meta.UserId)
		if err != nil {
			return openai.ErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
		}
		if userQuota < preConsumedQuota {
			return openai.ErrorWrapper(fmt.Errorf("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
		}
		if err = dbmodel.CacheDecreaseUserQuota(meta.UserId, preConsumedQuota); err != nil {
			return openai.ErrorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
		}
		if err = dbmodel.PreConsumeTokenQuota(meta.TokenId, preConsumedQuota); err != nil {
			return openai.ErrorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
	}

	requestBody, err := common.GetRequestBody(c)
	if err != nil {
		if isWrite {
			returnPreConsumed(preConsumedQuota, meta.TokenId)
		}
		return openai.ErrorWrapper(err, "read_request_body_failed", http.StatusBadRequest)
	}

	adaptor := relay.GetAdaptor(meta.APIType)
	if adaptor == nil {
		if isWrite {
			returnPreConsumed(preConsumedQuota, meta.TokenId)
		}
		return openai.ErrorWrapper(nil, "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(meta)

	resp, doErr := adaptor.DoRequest(c, meta, bytes.NewBuffer(requestBody))
	if doErr != nil {
		if isWrite {
			returnPreConsumed(preConsumedQuota, meta.TokenId)
		}
		logger.Errorf(ctx, "SaaS DoRequest failed: %s", doErr.Error())
		return openai.ErrorWrapper(doErr, "do_request_failed", http.StatusInternalServerError)
	}

	_, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		if isWrite {
			returnPreConsumed(preConsumedQuota, meta.TokenId)
		}
		return respErr
	}

	if isWrite {
		go postConsumeSaasQuota(ctx, meta, ratio, preConsumedQuota, modelRatio, groupRatio)
	}
	return nil
}

func returnPreConsumed(quota int64, tokenId int) {
	if quota > 0 {
		_ = dbmodel.PostConsumeTokenQuota(tokenId, -quota)
	}
}

func postConsumeSaasQuota(ctx context.Context, meta *meta.Meta, ratio float64, preConsumedQuota int64, modelRatio float64, groupRatio float64) {
	quota := int64(math.Ceil(ratio))
	if ratio != 0 && quota <= 0 {
		quota = 1
	}

	quotaDelta := quota - preConsumedQuota
	_ = dbmodel.PostConsumeTokenQuota(meta.TokenId, quotaDelta)
	_ = dbmodel.CacheUpdateUserQuota(ctx, meta.UserId)

	logContent := fmt.Sprintf("SaaS 业务请求 %s 倍率：%.2f × %.2f", meta.RequestURLPath, modelRatio, groupRatio)
	dbmodel.RecordConsumeLog(ctx, &dbmodel.Log{
		UserId:      meta.UserId,
		ChannelId:   meta.ChannelId,
		ModelName:   "saas-backend",
		TokenName:   meta.TokenName,
		Quota:       int(quota),
		Content:     logContent,
		ElapsedTime: helper.CalcElapsedTime(meta.StartTime),
	})
	dbmodel.UpdateUserUsedQuotaAndRequestCount(meta.UserId, quota)
	dbmodel.UpdateChannelUsedQuota(meta.ChannelId, quota)
}
