package engine

import (
	"testing"

	"github.com/wang900115/quant/model/result"
)

func TestNewExecutionManager(t *testing.T) {
	e := NewExecutionManager(10, 5)
	if e == nil {
		t.Fatal("NewExecutionManager returned nil")
	}
	if cap(e.fixedStoplossChannel) != 10 {
		t.Errorf("fixedStoplossChannel cap: want 10, got %d", cap(e.fixedStoplossChannel))
	}
	if cap(e.DebouncedStoplossChannel) != 10 {
		t.Errorf("DebouncedStoplossChannel cap: want 10, got %d", cap(e.DebouncedStoplossChannel))
	}
	if cap(e.fixedTakeProfitChannel) != 10 {
		t.Errorf("fixedTakeProfitChannel cap: want 10, got %d", cap(e.fixedTakeProfitChannel))
	}
	if cap(e.DebouncedTakeProfitChannel) != 10 {
		t.Errorf("DebouncedTakeProfitChannel cap: want 10, got %d", cap(e.DebouncedTakeProfitChannel))
	}
	if cap(e.hybridFixedChannel) != 10 {
		t.Errorf("hybridFixedChannel cap: want 10, got %d", cap(e.hybridFixedChannel))
	}
	if cap(e.hybridDebouncedChannel) != 10 {
		t.Errorf("hybridDebouncedChannel cap: want 10, got %d", cap(e.hybridDebouncedChannel))
	}
	if cap(e.generalResults) != 5 {
		t.Errorf("generalResults cap: want 5, got %d", cap(e.generalResults))
	}
	if cap(e.hybridResults) != 5 {
		t.Errorf("hybridResults cap: want 5, got %d", cap(e.hybridResults))
	}
}

func TestGetResult(t *testing.T) {
	e := NewExecutionManager(10, 5)
	gen, hyb := e.getResult()
	if gen == nil {
		t.Error("general results channel is nil")
	}
	if hyb == nil {
		t.Error("hybrid results channel is nil")
	}

	// Verify they are readable channels by doing a non-blocking peek
	var _ <-chan result.StrategyGeneralResult = gen
	var _ <-chan result.StrategyHybridResult = hyb
}

func TestCloseChannels(t *testing.T) {
	e := NewExecutionManager(2, 2)
	e.closeChannels()

	// After close, all channels should be closed (receive returns ok=false)
	_, ok := <-e.fixedStoplossChannel
	if ok {
		t.Error("fixedStoplossChannel should be closed")
	}
	_, ok = <-e.DebouncedStoplossChannel
	if ok {
		t.Error("DebouncedStoplossChannel should be closed")
	}
	_, ok = <-e.generalResults
	if ok {
		t.Error("generalResults channel should be closed")
	}
	_, ok = <-e.hybridResults
	if ok {
		t.Error("hybridResults channel should be closed")
	}
}
