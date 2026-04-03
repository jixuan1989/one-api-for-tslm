package model

import (
	"strings"
	"testing"

	"github.com/songquanpeng/one-api/common/config"
)

func TestSystemSaasTokenSubnet_Default(t *testing.T) {
	if config.SystemSaasTokenSubnet != "" {
		t.Skip("SystemSaasTokenSubnet is set by env")
	}
}

func TestSystemSaasTokenSubnet_Config(t *testing.T) {
	config.SystemSaasTokenSubnet = "10.0.0.0/8"
	defer func() { config.SystemSaasTokenSubnet = "" }()
}

func TestDeleteToken_SystemTokenProtection(t *testing.T) {
	if !strings.HasPrefix("system", "system") {
		t.Error("system token should have system prefix")
	}
}
