package controller

import (
	"bytes"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

// RelaySaasHelper handles business requests forwarded to SaaS Backend.
// No billing — just transparent proxy with user info injection (handled by saas adaptor).
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
		return openai.ErrorWrapper(doErr, "do_request_failed", http.StatusInternalServerError)
	}

	_, respErr := adaptor.DoResponse(c, resp, meta)
	return respErr
}
