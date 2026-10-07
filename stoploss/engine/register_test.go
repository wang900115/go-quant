package engine

import (
	"sync"
	"testing"

	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/stoploss"
)

// --- mock strategy types ---

type mockFixedSL struct{}

func (m *mockFixedSL) CalculateStopLoss(p decimal.Decimal) (decimal.Decimal, error) {
	return p, nil
}
func (m *mockFixedSL) Trigger(evt stoploss.TriggerEvent) error { return nil }
func (m *mockFixedSL) GetStopLoss() (decimal.Decimal, error)   { return decimal.Zero, nil }
func (m *mockFixedSL) ReSetStopLosser(p decimal.Decimal) error { return nil }
func (m *mockFixedSL) Deactivate() error                       { return nil }
func (m *mockFixedSL) ShouldTriggerStopLoss(p decimal.Decimal) (bool, error) {
	return false, nil
}

type mockDebouncedSL struct{}

func (m *mockDebouncedSL) CalculateStopLoss(p decimal.Decimal) (decimal.Decimal, error) {
	return p, nil
}
func (m *mockDebouncedSL) Trigger(evt stoploss.TriggerEvent) error { return nil }
func (m *mockDebouncedSL) GetStopLoss() (decimal.Decimal, error)   { return decimal.Zero, nil }
func (m *mockDebouncedSL) ReSetStopLosser(p decimal.Decimal) error { return nil }
func (m *mockDebouncedSL) Deactivate() error                       { return nil }
func (m *mockDebouncedSL) ShouldTriggerStopLoss(p decimal.Decimal, ts int64) (bool, error) {
	return false, nil
}
func (m *mockDebouncedSL) GetTimeThreshold() (int64, error) { return 0, nil }

type mockFixedTP struct{}

func (m *mockFixedTP) CalculateTakeProfit(p decimal.Decimal) (decimal.Decimal, error) {
	return p, nil
}
func (m *mockFixedTP) Trigger(evt stoploss.TriggerEvent) error   { return nil }
func (m *mockFixedTP) ReSetTakeProfiter(p decimal.Decimal) error { return nil }
func (m *mockFixedTP) GetTakeProfit() (decimal.Decimal, error)   { return decimal.Zero, nil }
func (m *mockFixedTP) Deactivate() error                         { return nil }
func (m *mockFixedTP) ShouldTriggerTakeProfit(p decimal.Decimal) (bool, error) {
	return false, nil
}

type mockDebouncedTP struct{}

func (m *mockDebouncedTP) CalculateTakeProfit(p decimal.Decimal) (decimal.Decimal, error) {
	return p, nil
}
func (m *mockDebouncedTP) Trigger(evt stoploss.TriggerEvent) error   { return nil }
func (m *mockDebouncedTP) ReSetTakeProfiter(p decimal.Decimal) error { return nil }
func (m *mockDebouncedTP) GetTakeProfit() (decimal.Decimal, error)   { return decimal.Zero, nil }
func (m *mockDebouncedTP) Deactivate() error                         { return nil }
func (m *mockDebouncedTP) ShouldTriggerTakeProfit(p decimal.Decimal, ts int64) (bool, error) {
	return false, nil
}
func (m *mockDebouncedTP) GetTimeThreshold() (int64, error) { return 0, nil }

type mockHybridFixed struct{}

func (m *mockHybridFixed) Calculate(p decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	return p, p, nil
}
func (m *mockHybridFixed) Trigger(evt stoploss.TriggerEvent) error { return nil }
func (m *mockHybridFixed) ReSet(p decimal.Decimal) error           { return nil }
func (m *mockHybridFixed) GetTakeProfit() (decimal.Decimal, error) { return decimal.Zero, nil }
func (m *mockHybridFixed) GetStopLoss() (decimal.Decimal, error)   { return decimal.Zero, nil }
func (m *mockHybridFixed) Deactivate() error                       { return nil }
func (m *mockHybridFixed) ShouldTriggerTakeProfit(p decimal.Decimal) (bool, error) {
	return false, nil
}
func (m *mockHybridFixed) ShouldTriggerStopLoss(p decimal.Decimal) (bool, error) {
	return false, nil
}

type mockHybridDebounced struct{}

func (m *mockHybridDebounced) Calculate(p decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	return p, p, nil
}
func (m *mockHybridDebounced) Trigger(evt stoploss.TriggerEvent) error { return nil }
func (m *mockHybridDebounced) ReSet(p decimal.Decimal) error           { return nil }
func (m *mockHybridDebounced) GetTakeProfit() (decimal.Decimal, error) { return decimal.Zero, nil }
func (m *mockHybridDebounced) GetStopLoss() (decimal.Decimal, error)   { return decimal.Zero, nil }
func (m *mockHybridDebounced) Deactivate() error                       { return nil }
func (m *mockHybridDebounced) ShouldTriggerTakeProfit(p decimal.Decimal, ts int64) (bool, error) {
	return false, nil
}
func (m *mockHybridDebounced) ShouldTriggerStopLoss(p decimal.Decimal, ts int64) (bool, error) {
	return false, nil
}
func (m *mockHybridDebounced) GetTimeThreshold() (int64, error) { return 0, nil }

// --- tests ---

func TestNewPortfolio(t *testing.T) {
	p := NewPortfolio()
	if p.fixedStoplossStrategies == nil {
		t.Error("fixedStoplossStrategies map not initialized")
	}
	if p.DebouncedStoplossStrategies == nil {
		t.Error("DebouncedStoplossStrategies map not initialized")
	}
	if p.fixedTakeProfitStrategies == nil {
		t.Error("fixedTakeProfitStrategies map not initialized")
	}
	if p.DebouncedTakeProfitStrategies == nil {
		t.Error("DebouncedTakeProfitStrategies map not initialized")
	}
	if p.hybridFixedStrategies == nil {
		t.Error("hybridFixedStrategies map not initialized")
	}
	if p.hybridDebouncedStrategies == nil {
		t.Error("hybridDebouncedStrategies map not initialized")
	}
	if p.openGeneral {
		t.Error("openGeneral should be false initially")
	}
	if p.openHybrid {
		t.Error("openHybrid should be false initially")
	}
	if p.count != 0 {
		t.Errorf("count should be 0, got %d", p.count)
	}
}

func TestRegistFixedStoploss(t *testing.T) {
	p := NewPortfolio()
	p.RegistFixedStoplossStrategy("s1", &mockFixedSL{})
	if p.count != 1 {
		t.Errorf("count should be 1, got %d", p.count)
	}
	if !p.openGeneral {
		t.Error("openGeneral should be true after registering fixed stop loss")
	}
	if _, ok := p.fixedStoplossStrategies["s1"]; !ok {
		t.Error("strategy not stored")
	}
}

func TestRegistDebouncedStoploss(t *testing.T) {
	p := NewPortfolio()
	p.RegistDebouncedStoplossStrategy("s1", &mockDebouncedSL{})
	if p.count != 1 {
		t.Errorf("count should be 1, got %d", p.count)
	}
	if !p.openGeneral {
		t.Error("openGeneral should be true")
	}
}

func TestRegistFixedTakeProfit(t *testing.T) {
	p := NewPortfolio()
	p.RegistFixedTakeProfitStrategy("s1", &mockFixedTP{})
	if p.count != 1 {
		t.Errorf("count should be 1, got %d", p.count)
	}
	if !p.openGeneral {
		t.Error("openGeneral should be true")
	}
}

func TestRegistDebouncedTakeProfit(t *testing.T) {
	p := NewPortfolio()
	p.RegistDebouncedTakeProfitStrategy("s1", &mockDebouncedTP{})
	if p.count != 1 {
		t.Errorf("count should be 1, got %d", p.count)
	}
	if !p.openGeneral {
		t.Error("openGeneral should be true")
	}
}

func TestRegistHybridFixed(t *testing.T) {
	p := NewPortfolio()
	p.RegistHybridFixedStrategy("h1", &mockHybridFixed{})
	if p.count != 1 {
		t.Errorf("count should be 1, got %d", p.count)
	}
	if !p.openHybrid {
		t.Error("openHybrid should be true after registering hybrid fixed")
	}
	if p.openGeneral {
		t.Error("openGeneral should remain false for hybrid-only portfolio")
	}
}

func TestRegistHybridDebounced(t *testing.T) {
	p := NewPortfolio()
	p.RegistHybridDebouncedStrategy("h1", &mockHybridDebounced{})
	if p.count != 1 {
		t.Errorf("count should be 1, got %d", p.count)
	}
	if !p.openHybrid {
		t.Error("openHybrid should be true")
	}
}

func TestRegistConcurrent(t *testing.T) {
	p := NewPortfolio()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			name := string(rune('a' + n%26))
			p.RegistFixedStoplossStrategy(name, &mockFixedSL{})
		}(i)
	}
	wg.Wait()
	if p.count != 50 {
		t.Errorf("expected count=50 after 50 concurrent registrations, got %d", p.count)
	}
}

func TestGetterReturnsCopy(t *testing.T) {
	p := NewPortfolio()
	p.RegistFixedStoplossStrategy("original", &mockFixedSL{})

	m := p.GetFixedStoplossStrategies()
	m["injected"] = &mockFixedSL{}

	m2 := p.GetFixedStoplossStrategies()
	if _, found := m2["injected"]; found {
		t.Error("mutating the returned map should not affect internal state")
	}
}
