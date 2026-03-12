package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/result"
)

func TestNewReport_NilCallback(t *testing.T) {
	r := NewReport(nil)
	stats := r.Stats()
	for k, v := range stats {
		if v != 0 {
			t.Errorf("stats[%q] should be 0, got %d", k, v)
		}
	}
}

func TestProcessGeneralResult_Triggered(t *testing.T) {
	var called bool
	r := NewReport(func(v interface{}) { called = true })

	ch := make(chan result.StrategyGeneralResult, 1)
	res := result.NewGeneral("s1", model.FIXED, model.STOP_LOSS,
		decimal.NewFromFloat(100), decimal.NewFromFloat(90),
		time.Now(), 0)
	res.SetTriggered(true)
	ch <- *res

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go r.ProcessGeneralResult(ch, ctx)
	time.Sleep(100 * time.Millisecond)

	if !called {
		t.Error("callback should have been called on triggered result")
	}
	if r.Stats()["triggers"] != 1 {
		t.Errorf("triggerCount should be 1, got %d", r.Stats()["triggers"])
	}
}

func TestProcessGeneralResult_NotTriggered(t *testing.T) {
	var called bool
	r := NewReport(func(v interface{}) { called = true })

	ch := make(chan result.StrategyGeneralResult, 1)
	res := result.NewGeneral("s1", model.FIXED, model.STOP_LOSS,
		decimal.NewFromFloat(100), decimal.NewFromFloat(90),
		time.Now(), 0)
	res.SetTriggered(false)
	ch <- *res

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go r.ProcessGeneralResult(ch, ctx)
	time.Sleep(100 * time.Millisecond)

	if called {
		t.Error("callback should NOT be called on non-triggered result")
	}
	if r.Stats()["triggers"] != 0 {
		t.Errorf("triggerCount should be 0, got %d", r.Stats()["triggers"])
	}
}

func TestProcessGeneralResult_Error(t *testing.T) {
	var called bool
	r := NewReport(func(v interface{}) { called = true })

	ch := make(chan result.StrategyGeneralResult, 1)
	res := result.NewGeneral("s1", model.FIXED, model.STOP_LOSS,
		decimal.NewFromFloat(100), decimal.NewFromFloat(90),
		time.Now(), 0)
	res.SetError(errors.New("strategy error"))
	ch <- *res

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go r.ProcessGeneralResult(ch, ctx)
	time.Sleep(100 * time.Millisecond)

	if called {
		t.Error("callback should NOT be called on errored result")
	}
	if r.Stats()["errors"] != 1 {
		t.Errorf("errorCount should be 1, got %d", r.Stats()["errors"])
	}
}

func TestProcessGeneralResult_ContextCancel(t *testing.T) {
	r := NewReport(nil)
	ch := make(chan result.StrategyGeneralResult, 1)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		r.ProcessGeneralResult(ch, ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// success
	case <-time.After(1 * time.Second):
		t.Error("ProcessGeneralResult did not exit after context cancellation")
	}
}

func TestProcessGeneralResult_ClosedChannel(t *testing.T) {
	r := NewReport(nil)
	ch := make(chan result.StrategyGeneralResult)

	ctx := context.Background()
	done := make(chan struct{})
	go func() {
		r.ProcessGeneralResult(ch, ctx)
		close(done)
	}()

	close(ch)

	select {
	case <-done:
		// success
	case <-time.After(1 * time.Second):
		t.Error("ProcessGeneralResult did not exit after channel close")
	}
}

func TestStats(t *testing.T) {
	r := NewReport(nil)
	stats := r.Stats()
	expectedKeys := []string{"general_results", "hybrid_results", "triggers", "errors"}
	for _, k := range expectedKeys {
		if _, ok := stats[k]; !ok {
			t.Errorf("Stats() missing key %q", k)
		}
	}
}
