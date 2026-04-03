package saas

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/ctxkey"
	oneapimodel "github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/meta"
	"github.com/songquanpeng/one-api/relay/model"
)

type Adaptor struct{}

func (a *Adaptor) Init(meta *meta.Meta) {}

func (a *Adaptor) GetRequestURL(meta *meta.Meta) (string, error) {
	// /timer/api/v1/predictions/upload → /predictions/upload
	path := meta.RequestURLPath
	path = strings.TrimPrefix(path, "/timer/api/v1")
	if path == "" {
		path = "/"
	}
	return fmt.Sprintf("%s%s", meta.BaseURL, path), nil
}

func (a *Adaptor) SetupRequestHeader(c *gin.Context, req *http.Request, meta *meta.Meta) error {
	adaptor.SetupCommonRequestHeader(c, req, meta)

	// Inject user info headers
	userId := c.GetInt(ctxkey.Id)
	username := c.GetString(ctxkey.Username)
	role := c.GetInt(ctxkey.Role)
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)

	req.Header.Set("X-OneApi-User-Id", strconv.Itoa(userId))
	req.Header.Set("X-OneApi-Username", username)
	req.Header.Set("X-OneApi-User-Role", strconv.Itoa(role))
	req.Header.Set("X-OneApi-Timestamp", timestamp)

	// Inject system-forecast token for saas-backend to call forecast via one-api
	forecastToken, err := oneapimodel.GetUserSystemTokenForModel(userId, "sundial")
	if err == nil && forecastToken != nil {
		req.Header.Set("X-OneApi-Forecast-Token", forecastToken.Key)
	}

	// HMAC signature
	if config.SaasBackendSecret != "" {
		payload := fmt.Sprintf("%d:%s:%d:%s", userId, username, role, timestamp)
		mac := hmac.New(sha256.New, []byte(config.SaasBackendSecret))
		mac.Write([]byte(payload))
		req.Header.Set("X-OneApi-Signature", hex.EncodeToString(mac.Sum(nil)))
	}

	// Forward real client IP
	clientIP := c.ClientIP()
	req.Header.Set("X-Forwarded-For", clientIP)
	req.Header.Set("X-Real-IP", clientIP)

	return nil
}

func (a *Adaptor) ConvertRequest(c *gin.Context, relayMode int, request *model.GeneralOpenAIRequest) (any, error) {
	return nil, fmt.Errorf("saas adaptor does not use ConvertRequest")
}

func (a *Adaptor) ConvertImageRequest(request *model.ImageRequest) (any, error) {
	return nil, fmt.Errorf("saas adaptor does not support image requests")
}

func (a *Adaptor) DoRequest(c *gin.Context, meta *meta.Meta, requestBody io.Reader) (*http.Response, error) {
	return adaptor.DoRequestHelper(a, c, meta, requestBody)
}

func (a *Adaptor) DoResponse(c *gin.Context, resp *http.Response, meta *meta.Meta) (usage *model.Usage, err *model.ErrorWithStatusCode) {
	// Transparently proxy the response
	defer resp.Body.Close()
	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return nil, &model.ErrorWithStatusCode{
			Error:      model.Error{Message: "read response failed: " + readErr.Error(), Type: "saas_error"},
			StatusCode: http.StatusInternalServerError,
		}
	}
	if resp.StatusCode >= 400 {
		return nil, &model.ErrorWithStatusCode{
			Error:      model.Error{Message: string(body), Type: "saas_upstream_error"},
			StatusCode: resp.StatusCode,
		}
	}
	for k, vs := range resp.Header {
		for _, v := range vs {
			c.Writer.Header().Set(k, v)
		}
	}
	c.Writer.WriteHeader(resp.StatusCode)
	c.Writer.Write(body)
	return &model.Usage{TotalTokens: 0}, nil // no billing for business requests
}

func (a *Adaptor) GetModelList() []string {
	return []string{"saas-backend"}
}

func (a *Adaptor) GetChannelName() string {
	return "saas-backend"
}
