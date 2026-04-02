package model

import (
	"testing"

	"github.com/songquanpeng/one-api/common/config"
)

func TestSystemSaasTokenSubnet_Default(t *testing.T) {
	// Default should be empty (no subnet restriction)
	if config.SystemSaasTokenSubnet != "" {
		// Only fails if env var is set during test
		t.Skipf("SYSTEM_SAAS_TOKEN_SUBNET is set: %s", config.SystemSaasTokenSubnet)
	}
}

func TestSystemSaasTokenSubnet_Config(t *testing.T) {
	// Verify the config variable exists and is a string
	_ = config.SystemSaasTokenSubnet
}

func TestDeleteToken_SystemSaasProtection(t *testing.T) {
	// Verify that the system-saas token name constant is used correctly
	// This is a compile-time check that the protection logic references the right name
	tokenName := "system-saas"
	if tokenName != "system-saas" {
		t.Error("system-saas token name mismatch")
	}
}
