package timer

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/model"
)

// ForecastHandler transparently proxies the timer-rest-service response back to the client.
// It reads the upstream response, returns it as-is, and reports a minimal usage for billing.
func ForecastHandler(c *gin.Context, resp *http.Response) (*model.ErrorWithStatusCode, *model.Usage) {
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: "failed to read upstream response: " + err.Error(),
				Type:    "timer_error",
				Code:    "read_response_failed",
			},
			StatusCode: http.StatusInternalServerError,
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return &model.ErrorWithStatusCode{
			Error: model.Error{
				Message: string(body),
				Type:    "timer_upstream_error",
				Code:    "upstream_error",
			},
			StatusCode: resp.StatusCode,
		}, nil
	}

	// Parse response to count forecast tasks for billing
	var forecastResp model.ForecastResponse
	usage := &model.Usage{TotalTokens: 1} // default 1 unit per request
	if json.Unmarshal(body, &forecastResp) == nil && forecastResp.Code == 200 {
		totalPoints := 0
		for _, result := range forecastResp.Data.Results {
			totalPoints += len(result.Data)
		}
		if totalPoints > 0 {
			usage.TotalTokens = totalPoints
		}
	}

	// Write upstream response directly to client
	for k, vs := range resp.Header {
		for _, v := range vs {
			c.Writer.Header().Set(k, v)
		}
	}
	c.Writer.WriteHeader(resp.StatusCode)
	c.Writer.Write(body)

	return nil, usage
}
