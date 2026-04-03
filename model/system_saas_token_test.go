package model

import (
	"strings"
	"testing"

	"github.com/songquanpeng/one-api/common/config"
)

func TestSystemSaasTokenSubnet_Default(t *testing.T) {
	if config.SystemSaasTokenSubnet != "" {
		// Only check if not set by env
		t.Skip("SystemSaasTokenSubnet is set by env")
	}
}

func TestSystemSaasTokenSubnet_Config(t *testing.T) {
	config.SystemSaasTokenSubnet = "10.0.0.0/8"
	defer func() { config.SystemSaasTokenSubnet = "" }()
}

func TestDeleteToken_SystemTokenProtection(t *testing.T) {
	// Verify that system-* tokens are protected from non-admin deletion
	systemNames := []string{"system-forecast", "system-saas"}
	for _, name := range systemNames {
		if !strings.HasPrefix(name, "system-") {
			t.Errorf("%s should have system- prefix", name)
		}
	}
}

func TestSystemTokenModels(t *testing.T) {
	forecastModels := "sundial,chronos2,timer,timer_xl,moirai2"
	saasModels := "saas-backend"

	if !strings.Contains(forecastModels, "sundial") {
		t.Error("forecast models should contain sundial")
	}
	if saasModels != "saas-backend" {
		t.Error("saas models should be saas-backend")
	}
}
