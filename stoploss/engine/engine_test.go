package engine

import (
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/model"
)

func TestNew(t *testing.T) {
	cfg := DefaultConfig()
	e := New(cfg)
	if e == nil {
		t.Fatal("New returned nil")
	}
	if e.portfolio == nil {
		t.Error("portfolio not initialized")
	}
	if e.execution == nil {
		t.Error("execution not initialized")
	}
	if e.Reporter == nil {
		t.Error("Reporter not initialized")
	}
	if e.Metrics == nil {
		t.Error("Metrics not initialized")
	}
	if e.engine == nil {
		t.Error("internal sys.Engine not initialized")
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.BufferSize == 0 {
		t.Error("BufferSize should be non-zero")
	}
	if cfg.ReadTimeout == 0 {
		t.Error("ReadTimeout should be non-zero")
	}
	if cfg.CheckInterval == 0 {
		t.Error("CheckInterval should be non-zero")
	}
}

func TestRegisterStrategy_AllTypes(t *testing.T) {
	cfg := DefaultConfig()
	cfg.HeartbeatInterval = 1 * time.Second

	tests := []struct {
		name     string
		strategy interface{}
	}{
		{"fixed_sl", &mockFixedSL{}},
		{"debounced_sl", &mockDebouncedSL{}},
		{"fixed_tp", &mockFixedTP{}},
		{"debounced_tp", &mockDebouncedTP{}},
		{"hybrid_fixed", &mockHybridFixed{}},
		{"hybrid_debounced", &mockHybridDebounced{}},
	}

	e := New(cfg)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := e.RegisterStrategy(tt.name, tt.strategy); err != nil {
				t.Errorf("RegisterStrategy(%q): unexpected error: %v", tt.name, err)
			}
		})
	}
}

func TestRegisterStrategy_Unknown(t *testing.T) {
	e := New(DefaultConfig())
	err := e.RegisterStrategy("unknown", "not a strategy")
	if err == nil {
		t.Fatal("expected error for unsupported strategy type")
	}
	if err != errNonsupported {
		t.Errorf("expected errNonsupported, got %v", err)
	}
}

func TestCollect_DropOnFullChannel(t *testing.T) {
	cfg := DefaultConfig()
	cfg.BufferSize = 1
	cfg.BufferRSize = 1

	e := New(cfg)
	e.RegisterStrategy("s1", &mockFixedSL{})

	// Fill the channel
	e.execution.fixedStoplossChannel <- model.PricePoint{
		NewPrice:  decimal.NewFromFloat(100),
		UpdatedAt: time.Now(),
	}

	dropped := false
	e.Collect(model.PricePoint{
		NewPrice:  decimal.NewFromFloat(101),
		UpdatedAt: time.Now(),
	}, func() {
		dropped = true
	})

	if !dropped {
		t.Error("callback should have been called when channel is full (drop)")
	}
	if e.Metrics.TotalDropped.Snapshot().Count() == 0 {
		t.Error("TotalDropped metric should have incremented")
	}
}
