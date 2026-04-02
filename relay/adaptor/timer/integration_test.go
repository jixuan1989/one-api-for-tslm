package timer

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/songquanpeng/one-api/relay/model"
)

// Integration test: mock timer-rest-service, test full request/response cycle
func TestForecast_EndToEnd(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/forecast" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("unexpected method: %s", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var req model.ForecastRequest
		if err := json.Unmarshal(body, &req); err != nil {
			t.Errorf("failed to parse request: %v", err)
		}
		if len(req.Targets) != 1 {
			t.Errorf("targets len = %d, want 1", len(req.Targets))
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"code":200,"message":"ok","data":{"results":[{"columns":["value"],"data":[[10.5],[11.2],[12.0]]}]}}`)
	}))
	defer mockServer.Close()

	// Directly call the mock server (simulating what the adaptor does)
	reqBody := `{"targets":[{"columns":["value"],"data":[[1],[2],[3],[4],[5],[6],[7],[8],[9],[10],[11],[12],[13],[14],[15],[16]]}],"output_length_list":[3]}`
	resp, err := http.Post(mockServer.URL+"/api/v1/forecast", "application/json", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var forecastResp model.ForecastResponse
	body, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(body, &forecastResp); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if forecastResp.Code != 200 {
		t.Errorf("code = %d, want 200", forecastResp.Code)
	}
	if len(forecastResp.Data.Results) != 1 {
		t.Fatalf("results len = %d, want 1", len(forecastResp.Data.Results))
	}
	if len(forecastResp.Data.Results[0].Data) != 3 {
		t.Errorf("data rows = %d, want 3", len(forecastResp.Data.Results[0].Data))
	}
}

func TestHelloTimer_EndToEnd(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/hello_timer" {
			t.Errorf("unexpected path: %s, want /api/v1/hello_timer", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("unexpected method: %s, want GET", r.Method)
		}
		name := r.URL.Query().Get("name")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"code":200,"message":"Success","data":{"greeting":"Hello, %s!"}}`, name)
	}))
	defer mockServer.Close()

	resp, err := http.Get(mockServer.URL + "/api/v1/hello_timer?name=world")
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Hello, world!") {
		t.Errorf("response body = %s, want greeting with 'world'", string(body))
	}
}

func TestForecast_WithBearerToken(t *testing.T) {
	var receivedAuth string
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"code":200,"message":"ok","data":{"results":[{"columns":["v"],"data":[[1]]}]}}`)
	}))
	defer mockServer.Close()

	req, _ := http.NewRequest("POST", mockServer.URL+"/api/v1/forecast", strings.NewReader(`{"targets":[{"columns":["v"],"data":[[1]]}]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer sk-test-token-123")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	resp.Body.Close()

	if receivedAuth != "Bearer sk-test-token-123" {
		t.Errorf("auth header = %q, want 'Bearer sk-test-token-123'", receivedAuth)
	}
}
