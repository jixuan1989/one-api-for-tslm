package controller

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/ctxkey"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	dbmodel "github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

// RelaySaasHelper handles business requests forwarded to SaaS Backend.
// No billing — but records usage log for auditing.
func RelaySaasHelper(c *gin.Context) *model.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := meta.GetByContext(c)

	requestBody, err := common.GetRequestBody(c)
	if err != nil {
		return openai.ErrorWrapper(err, "read_request_body_failed", http.StatusBadRequest)
	}

	adaptor := relay.GetAdaptor(meta.APIType)
	if adaptor == nil {
		return openai.ErrorWrapper(nil, "invalid_api_type", http.StatusBadRequest)
	}
	adaptor.Init(meta)

	resp, doErr := adaptor.DoRequest(c, meta, bytes.NewBuffer(requestBody))
	if doErr != nil {
		logger.Errorf(ctx, "SaaS DoRequest failed: %s", doErr.Error())
		go recordSaasLog(c, meta, "失败: "+doErr.Error())
		return openai.ErrorWrapper(doErr, "do_request_failed", http.StatusInternalServerError)
	}

	_, respErr := adaptor.DoResponse(c, resp, meta)
	if respErr != nil {
		go recordSaasLog(c, meta, fmt.Sprintf("失败: %d %s", respErr.StatusCode, respErr.Error.Message))
		return respErr
	}

	go recordSaasLog(c, meta, "成功")
	return nil
}

func recordSaasLog(c *gin.Context, meta *meta.Meta, content string) {
	ctx := c.Request.Context()
	userId := c.GetInt(ctxkey.Id)
	path := c.Request.URL.Path
	logContent := fmt.Sprintf("SaaS 业务请求 %s: %s", path, content)
	dbmodel.RecordConsumeLog(ctx, &dbmodel.Log{
		UserId:      userId,
		ChannelId:   meta.ChannelId,
		ModelName:   "saas-backend",
		TokenName:   "session",
		Quota:       0,
		Content:     logContent,
		ElapsedTime: helper.CalcElapsedTime(meta.StartTime),
	})
}
