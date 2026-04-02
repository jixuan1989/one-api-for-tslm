package controller

import (
	"bytes"
	"context"
	"encoding/json"
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

// RelayForecastHelper handles time series forecast requests.
// It transparently proxies the request body to the timer-rest-service backend.
func RelayForecastHelper(c *gin.Context) *model.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := meta.GetByContext(c)

	// Read request body (kept reusable)
	requestBody, err := common.GetRequestBody(c)
	if err != nil {
		return openai.ErrorWrapper(err, "read_request_body_failed", http.StatusBadRequest)
	}

	// Extract model_id for billing
	var forecastReq model.ForecastRequest
	modelName := "sundial"
	if json.Unmarshal(requestBody, &forecastReq) == nil && forecastReq.ModelID != nil {
		modelName = *forecastReq.ModelID
	}
	meta.OriginModelName = modelName
	meta.ActualModelName = modelName

	// Billing: ratio calculation
	modelRatio := billingratio.GetModelRatio(modelName, meta.ChannelType)
	groupRatio := billingratio.GetGroupRatio(meta.Group)
	ratio := modelRatio * groupRatio

	// Pre-consume: estimate 1 unit
	preConsumedQuota := int64(math.Ceil(ratio))
	if preConsumedQuota <= 0 {
		preConsumedQuota = 1
	}

	userQuota, qErr := dbmodel.CacheGetUserQuota(ctx, meta.UserId)
	if qErr != nil {
		return openai.ErrorWrapper(qErr, "get_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota < preConsumedQuota {
		return openai.ErrorWrapper(fmt.Errorf("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}
	qErr = dbmodel.CacheDecreaseUserQuota(meta.UserId, preConsumedQuota)
	if qErr != nil {
		return openai.ErrorWrapper(qErr, "decrease_user_quota_failed", http.StatusInternalServerError)
	}
	qErr = dbmodel.PreConsumeTokenQuota(meta.TokenId, preConsumedQuota)
	if qErr != nil {
		return openai.ErrorWrapper(qErr, "pre_consume_token_quota_failed", http.StatusForbidden)
	}

	// Get adaptor
	adaptor := relay.GetAdaptor(meta.APIType)
	if adaptor == nil {
		returnForecastPreConsumed(preConsumedQuota, meta.TokenId)
		return openai.ErrorWrapper(fmt.Errorf("invalid api type: %d", meta.APIType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(meta)

	// Proxy request
	resp, doErr := adaptor.DoRequest(c, meta, bytes.NewBuffer(requestBody))
	if doErr != nil {
		returnForecastPreConsumed(preConsumedQuota, meta.TokenId)
		logger.Errorf(ctx, "DoRequest failed: %s", doErr.Error())
		go recordForecastFailure(ctx, meta, modelName, "do_request_failed: "+doErr.Error())
		return openai.ErrorWrapper(doErr, "do_request_failed", http.StatusInternalServerError)
	}

	// Handle response
	usage, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		returnForecastPreConsumed(preConsumedQuota, meta.TokenId)
		go recordForecastFailure(ctx, meta, modelName, fmt.Sprintf("upstream_error: %d %s", respErr.StatusCode, respErr.Error.Message))
		return respErr
	}

	// Post-consume quota
	go postConsumeForecastQuota(ctx, usage, meta, modelName, ratio, preConsumedQuota, modelRatio, groupRatio)
	return nil
}

func returnForecastPreConsumed(preConsumedQuota int64, tokenId int) {
	if preConsumedQuota > 0 {
		_ = dbmodel.PostConsumeTokenQuota(tokenId, -preConsumedQuota)
	}
}

func recordForecastFailure(ctx context.Context, meta *meta.Meta, modelName string, errMsg string) {
	dbmodel.RecordConsumeLog(ctx, &dbmodel.Log{
		UserId:      meta.UserId,
		ChannelId:   meta.ChannelId,
		ModelName:   modelName,
		TokenName:   meta.TokenName,
		Quota:       0,
		Content:     "时序预测失败: " + errMsg,
		ElapsedTime: helper.CalcElapsedTime(meta.StartTime),
	})
}

func postConsumeForecastQuota(ctx context.Context, usage *model.Usage, meta *meta.Meta, modelName string, ratio float64, preConsumedQuota int64, modelRatio float64, groupRatio float64) {
	totalTokens := 0
	if usage != nil {
		totalTokens = usage.TotalTokens
	}
	quota := int64(math.Ceil(float64(totalTokens) * ratio))
	if ratio != 0 && quota <= 0 {
		quota = 1
	}
	if totalTokens == 0 {
		quota = 0
	}

	quotaDelta := quota - preConsumedQuota
	err := dbmodel.PostConsumeTokenQuota(meta.TokenId, quotaDelta)
	if err != nil {
		logger.SysError("error consuming token remain quota: " + err.Error())
	}
	err = dbmodel.CacheUpdateUserQuota(ctx, meta.UserId)
	if err != nil {
		logger.SysError("error update user quota cache: " + err.Error())
	}

	logContent := fmt.Sprintf("时序预测 倍率：%.2f × %.2f", modelRatio, groupRatio)
	dbmodel.RecordConsumeLog(ctx, &dbmodel.Log{
		UserId:      meta.UserId,
		ChannelId:   meta.ChannelId,
		ModelName:   modelName,
		TokenName:   meta.TokenName,
		Quota:       int(quota),
		Content:     logContent,
		ElapsedTime: helper.CalcElapsedTime(meta.StartTime),
	})
	dbmodel.UpdateUserUsedQuotaAndRequestCount(meta.UserId, quota)
	dbmodel.UpdateChannelUsedQuota(meta.ChannelId, quota)
}
