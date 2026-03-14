package db

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.MaxOpenConns != 25 {
		t.Error("max open")
	}
	if cfg.MaxIdleConns != 10 {
		t.Error("max idle")
	}
}
