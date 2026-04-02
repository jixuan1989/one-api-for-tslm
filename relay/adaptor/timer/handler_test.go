package timer

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/songquanpeng/one-api/relay/model"
)

func TestForecastHandler_Success(t *testing.T) {
	body := `{"code":200,"message":"ok","data":{"results":[{"columns":["value"],"data":[[1.5],[2.5],[3.5]]}]}}`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	errResp, usage := ForecastHandler(c, resp)
	if errResp != nil {
		t.Fatalf("unexpected error: %v", errResp.Error.Message)
	}
	if usage == nil {
		t.Fatal("usage is nil")
	}
	// 3 data points
	if usage.TotalTokens != 3 {
		t.Errorf("TotalTokens = %d, want 3", usage.TotalTokens)
	}
	// Response body should be written to client
	if w.Code != http.StatusOK {
		t.Errorf("status code = %d, want 200", w.Code)
	}
	var result model.ForecastResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	if result.Code != 200 {
		t.Errorf("response code = %d, want 200", result.Code)
	}
}

func TestForecastHandler_UpstreamError(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusInternalServerError,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{"error":"model not found"}`)),
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	errResp, _ := ForecastHandler(c, resp)
	if errResp == nil {
		t.Fatal("expected error response for 500 upstream")
	}
	if errResp.StatusCode != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", errResp.StatusCode)
	}
}

func TestForecastHandler_EmptyResults(t *testing.T) {
	body := `{"code":200,"message":"ok","data":{"results":[]}}`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	errResp, usage := ForecastHandler(c, resp)
	if errResp != nil {
		t.Fatalf("unexpected error: %v", errResp.Error.Message)
	}
	// No data points, default 1
	if usage.TotalTokens != 1 {
		t.Errorf("TotalTokens = %d, want 1 (default)", usage.TotalTokens)
	}
}

func TestForecastHandler_MultipleResults(t *testing.T) {
	body := `{"code":200,"message":"ok","data":{"results":[{"columns":["v"],"data":[[1],[2]]},{"columns":["v"],"data":[[3],[4],[5]]}]}}`
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}

	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	errResp, usage := ForecastHandler(c, resp)
	if errResp != nil {
		t.Fatalf("unexpected error: %v", errResp.Error.Message)
	}
	// 2 + 3 = 5 data points
	if usage.TotalTokens != 5 {
		t.Errorf("TotalTokens = %d, want 5", usage.TotalTokens)
	}
}
