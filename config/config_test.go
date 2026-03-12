package config

import (
	"testing"
	"time"

	"github.com/wang900115/quant/stoploss/engine"
)

func TestNew_Empty(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
}

func TestNew_WithEngineOpt(t *testing.T) {
	cfg := engine.Config{BufferSize: 512, CheckInterval: 3 * time.Second}
	c := New()
	c.apply(c.WithEngine(cfg))
	if c.Engine.BufferSize != 512 {
		t.Errorf("Engine.BufferSize: want 512, got %d", c.Engine.BufferSize)
	}
	if c.Engine.CheckInterval != 3*time.Second {
		t.Errorf("Engine.CheckInterval: want 3s, got %v", c.Engine.CheckInterval)
	}
}

func TestNewFromEnv_EngineBufferSize(t *testing.T) {
	t.Setenv("ENGINE_BUFFER_SIZE", "1024")
	c, err := NewFromEnv()
	if err != nil {
		t.Fatalf("NewFromEnv: %v", err)
	}
	if c.Engine.BufferSize != 1024 {
		t.Errorf("Engine.BufferSize: want 1024, got %d", c.Engine.BufferSize)
	}
}

func TestNewFromEnv_BinanceCredentials(t *testing.T) {
	t.Setenv("BINANCE_API_KEY", "mykey")
	t.Setenv("BINANCE_SECRET", "mysecret")
	c, err := NewFromEnv()
	if err != nil {
		t.Fatalf("NewFromEnv: %v", err)
	}
	if c.Binance.APIKey != "mykey" {
		t.Errorf("Binance.APIKey: want 'mykey', got %q", c.Binance.APIKey)
	}
	if c.Binance.SecretKey != "mysecret" {
		t.Errorf("Binance.SecretKey: want 'mysecret', got %q", c.Binance.SecretKey)
	}
}

func TestNewFromEnv_InvalidBufferSize(t *testing.T) {
	t.Setenv("ENGINE_BUFFER_SIZE", "not-a-number")
	_, err := NewFromEnv()
	if err == nil {
		t.Fatal("expected error for invalid ENGINE_BUFFER_SIZE")
	}
}

func TestValidate_BinanceKeyWithoutSecret(t *testing.T) {
	c := New()
	c.Binance.APIKey = "somekey"
	// SecretKey intentionally left empty
	if err := c.Validate(); err == nil {
		t.Fatal("expected error when APIKey is set but SecretKey is empty")
	}
}

func TestValidate_Valid(t *testing.T) {
	c := New()
	c.Binance.APIKey = "key"
	c.Binance.SecretKey = "secret"
	if err := c.Validate(); err != nil {
		t.Fatalf("Validate should pass with paired credentials: %v", err)
	}
}

func TestValidate_AllEmpty(t *testing.T) {
	c := New()
	if err := c.Validate(); err != nil {
		t.Fatalf("Validate should pass when all credentials are empty: %v", err)
	}
}
