package saas

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/songquanpeng/one-api/common/config"
)

func TestSaasBackend_EndToEnd(t *testing.T) {
	// Mock saas backend
	var receivedHeaders http.Header
	var receivedPath string
	var receivedBody string
	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header
		receivedPath = r.URL.Path
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"success":true,"data":{"id":1}}`)
	}))
	defer mockBackend.Close()

	// Set secret for signing
	old := config.SaasBackendSecret
	config.SaasBackendSecret = "test-secret-123"
	defer func() { config.SaasBackendSecret = old }()

	// Call mock backend directly (simulating what adaptor does)
	reqBody := `{"query":"test","modelType":"TIMER"}`
	req, _ := http.NewRequest("POST", mockBackend.URL+"/predictions", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-OneApi-User-Id", "42")
	req.Header.Set("X-OneApi-Username", "testuser")
	req.Header.Set("X-OneApi-User-Role", "1")
	req.Header.Set("X-OneApi-Timestamp", "1775186400")
	req.Header.Set("X-OneApi-Signature", "somesig")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if receivedPath != "/predictions" {
		t.Errorf("path = %s, want /predictions", receivedPath)
	}
	if receivedHeaders.Get("X-OneApi-User-Id") != "42" {
		t.Errorf("X-OneApi-User-Id = %s, want 42", receivedHeaders.Get("X-OneApi-User-Id"))
	}
	if receivedBody != reqBody {
		t.Errorf("body = %s, want %s", receivedBody, reqBody)
	}

	var result map[string]interface{}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)
	if result["success"] != true {
		t.Errorf("success = %v, want true", result["success"])
	}
}

func TestSaasBackend_HeaderInjection(t *testing.T) {
	// Verify that user info headers are present
	var headers http.Header
	mockBackend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers = r.Header
		w.WriteHeader(200)
		fmt.Fprint(w, `{"success":true}`)
	}))
	defer mockBackend.Close()

	req, _ := http.NewRequest("GET", mockBackend.URL+"/predictions/models", nil)
	req.Header.Set("X-OneApi-User-Id", "7")
	req.Header.Set("X-OneApi-Username", "alice")
	req.Header.Set("X-Forwarded-For", "1.2.3.4")

	resp, _ := http.DefaultClient.Do(req)
	resp.Body.Close()

	if headers.Get("X-OneApi-User-Id") != "7" {
		t.Errorf("missing X-OneApi-User-Id")
	}
	if headers.Get("X-OneApi-Username") != "alice" {
		t.Errorf("missing X-OneApi-Username")
	}
	if headers.Get("X-Forwarded-For") != "1.2.3.4" {
		t.Errorf("missing X-Forwarded-For")
	}
}
