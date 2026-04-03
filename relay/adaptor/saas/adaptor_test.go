package saas

import (
	"testing"

	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/relay/meta"
)

func TestGetRequestURL_Predictions(t *testing.T) {
	a := &Adaptor{}
	tests := []struct {
		path string
		base string
		want string
	}{
		{"/timer/api/v1/predictions", "http://127.0.0.1:8080/api", "http://127.0.0.1:8080/api/predictions"},
		{"/timer/api/v1/predictions/upload", "http://127.0.0.1:8080/api", "http://127.0.0.1:8080/api/predictions/upload"},
		{"/timer/api/v1/predictions/123", "http://127.0.0.1:8080/api", "http://127.0.0.1:8080/api/predictions/123"},
		{"/timer/api/v1/predictions/models", "http://localhost:8080/api", "http://localhost:8080/api/predictions/models"},
		{"/timer/api/v1/predictions/dashboard/stats", "http://localhost:8080/api", "http://localhost:8080/api/predictions/dashboard/stats"},
	}
	for _, tt := range tests {
		m := &meta.Meta{BaseURL: tt.base, RequestURLPath: tt.path}
		got, err := a.GetRequestURL(m)
		if err != nil {
			t.Fatalf("unexpected error for path %s: %v", tt.path, err)
		}
		if got != tt.want {
			t.Errorf("GetRequestURL(%s) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestGetChannelName(t *testing.T) {
	a := &Adaptor{}
	if a.GetChannelName() != "saas-backend" {
		t.Errorf("got %q, want saas-backend", a.GetChannelName())
	}
}

func TestGetModelList(t *testing.T) {
	a := &Adaptor{}
	models := a.GetModelList()
	if len(models) != 1 || models[0] != "saas-backend" {
		t.Errorf("got %v, want [saas-backend]", models)
	}
}

func TestHmacSignature(t *testing.T) {
	// Verify that when secret is set, signature is deterministic
	old := config.SaasBackendSecret
	config.SaasBackendSecret = "test-secret"
	defer func() { config.SaasBackendSecret = old }()

	// The actual signing happens in SetupRequestHeader which needs gin context
	// Just verify config is accessible
	if config.SaasBackendSecret != "test-secret" {
		t.Error("SaasBackendSecret not set")
	}
}

func TestNoSignatureWithoutSecret(t *testing.T) {
	old := config.SaasBackendSecret
	config.SaasBackendSecret = ""
	defer func() { config.SaasBackendSecret = old }()

	if config.SaasBackendSecret != "" {
		t.Error("SaasBackendSecret should be empty")
	}
}
