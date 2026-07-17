package config

import (
	"testing"
	"time"
)

func TestLoadTransferTimeout(t *testing.T) {
	t.Setenv("OPENPROJECT_TRANSFER_TIMEOUT", "2m30s")
	if timeout := Load().TransferTimeout; timeout != 2*time.Minute+30*time.Second {
		t.Fatalf("TransferTimeout = %s, want 2m30s", timeout)
	}
}

func TestLoadTransferTimeoutFallsBackForInvalidValue(t *testing.T) {
	t.Setenv("OPENPROJECT_TRANSFER_TIMEOUT", "invalid")
	if timeout := Load().TransferTimeout; timeout != 5*time.Minute {
		t.Fatalf("TransferTimeout = %s, want 5m", timeout)
	}
}
