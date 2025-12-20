// Copyright (C) 2025 Quantive
//
// SPDX-License-Identifier: MIT OR AGPL-3.0-or-later
//
// This file is part of the Decision Engine project.
// You may choose to use this file under the terms of either
// the MIT License or the GNU Affero General Public License v3.0 or later.
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the LICENSE files for more details.

package strategy

import (
	"testing"

	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/stoploss"
)

func TestNewFixedTrailingStop(t *testing.T) {
	s, err := NewFixedTrailingStop(d(100), d(0.05), nil)
	if err != nil {
		t.Fatalf("Failed to create fixed trailing stop: %v", err)
	}
	if s == nil {
		t.Fatal("Strategy is nil")
	}

	sl, _ := s.GetStopLoss()
	expected := d(95) // 100 * (1 - 0.05)
	if !sl.Equal(expected) {
		t.Errorf("Expected SL=%v, got %v", expected, sl)
	}
	t.Logf("Entry: 100, Initial Stop Loss: %v (5%%)", sl)
}

func TestNewFixedTrailingProfit(t *testing.T) {
	s, err := NewFixedTrailingProfit(d(100), d(0.10), nil)
	if err != nil {
		t.Fatalf("Failed to create fixed trailing profit: %v", err)
	}
	if s == nil {
		t.Fatal("Strategy is nil")
	}

	tp, _ := s.GetTakeProfit()
	expected := d(110) // 100 * (1 + 0.10)
	if !tp.Equal(expected) {
		t.Errorf("Expected TP=%v, got %v", expected, tp)
	}
	t.Logf("Entry: 100, Initial Take Profit: %v (10%%)", tp)
}

func TestFixedTrailingStop_InvalidParameters(t *testing.T) {
	tests := []struct {
		name        string
		entryPrice  decimal.Decimal
		rate        decimal.Decimal
		expectError bool
	}{
		{"Valid 5% trailing", d(100), d(0.05), false},
		{"Valid 10% trailing", d(200), d(0.10), false},
		{"Negative rate", d(100), d(-0.05), true},
		{"Rate > 1", d(100), d(1.1), true},
		{"Zero rate", d(100), d(0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFixedTrailingStop(tt.entryPrice, tt.rate, nil)
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestFixedTrailingProfit_InvalidParameters(t *testing.T) {
	tests := []struct {
		name        string
		entryPrice  decimal.Decimal
		rate        decimal.Decimal
		expectError bool
	}{
		{"Valid 10% trailing", d(100), d(0.10), false},
		{"Valid 15% trailing", d(200), d(0.15), false},
		{"Negative rate", d(100), d(-0.10), true},
		{"Rate > 1", d(100), d(1.1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFixedTrailingProfit(tt.entryPrice, tt.rate, nil)
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestFixedTrailingStop_TrailingBehavior(t *testing.T) {
	entryPrice := d(100)
	rate := d(0.05)

	s, _ := NewFixedTrailingStop(entryPrice, rate, nil)
	initialSL, _ := s.GetStopLoss() // 95
	t.Logf("Initial Stop Loss: %v", initialSL)

	// Price increases to 110
	s.CalculateStopLoss(d(110))
	sl1, _ := s.GetStopLoss()
	expectedSL1 := d(104.5) // 110 * 0.95
	t.Logf("After price 110: SL=%v (expected %v)", sl1, expectedSL1)

	if !sl1.Equal(expectedSL1) {
		t.Errorf("Expected SL=%v, got %v", expectedSL1, sl1)
	}

	// Stop loss should have moved up
	if sl1.LessThanOrEqual(initialSL) {
		t.Error("Stop loss should trail up when price increases")
	}

	// Price increases to 120
	s.CalculateStopLoss(d(120))
	sl2, _ := s.GetStopLoss()
	expectedSL2 := d(114) // 120 * 0.95
	t.Logf("After price 120: SL=%v (expected %v)", sl2, expectedSL2)

	if !sl2.Equal(expectedSL2) {
		t.Errorf("Expected SL=%v, got %v", expectedSL2, sl2)
	}

	// Stop loss should keep moving up
	if sl2.LessThanOrEqual(sl1) {
		t.Error("Stop loss should continue trailing up")
	}

	// Price drops to 115
	s.CalculateStopLoss(d(115))
	sl3, _ := s.GetStopLoss()
	t.Logf("After price drops to 115: SL=%v", sl3)

	// Stop loss should NOT move down (trailing behavior)
	if !sl3.Equal(sl2) {
		t.Errorf("Stop loss should not decrease: expected %v, got %v", sl2, sl3)
	}
}

func TestFixedTrailingProfit_TrailingBehavior_Short(t *testing.T) {
	entryPrice := d(100)
	rate := d(0.10) // 10%

	s, _ := NewFixedTrailingProfit(entryPrice, rate, nil)
	initialTP, _ := s.GetTakeProfit()
	t.Logf("Initial Take Profit: %v", initialTP)

	expectedInitial := d(110)
	if !initialTP.Equal(expectedInitial) {
		t.Errorf("Initial take profit incorrect: expected %v, got %v", expectedInitial, initialTP)
	}

	s.CalculateTakeProfit(d(85))
	tp1, _ := s.GetTakeProfit()
	expectedTP1 := d(93.5) // 85 * (1 + 0.10)
	t.Logf("After price drops to 85: TP=%v (expected %v)", tp1, expectedTP1)

	if !tp1.Equal(expectedTP1) {
		t.Errorf("Take profit should move down with falling price: expected %v, got %v", expectedTP1, tp1)
	}

	s.CalculateTakeProfit(d(90))
	tp2, _ := s.GetTakeProfit()
	t.Logf("After price rises to 90: TP=%v (should stay at %v)", tp2, tp1)

	if !tp2.Equal(tp1) {
		t.Errorf("Take profit should not move up on price increase: expected %v, got %v", tp1, tp2)
	}
	s.CalculateTakeProfit(d(80))
	tp3, _ := s.GetTakeProfit()
	expectedTP3 := d(88) // 80 * (1 + 0.10)
	t.Logf("After price drops to 80: TP=%v (expected %v)", tp3, expectedTP3)

	if !tp3.Equal(expectedTP3) {
		t.Errorf("Take profit should continue to trail down: expected %v, got %v", expectedTP3, tp3)
	}
}

func TestFixedTrailingStop_HistoricalData_UpTrend(t *testing.T) {
	data := GetMockTrendingData()
	entryPrice := data[0].Close

	s, _ := NewFixedTrailingStop(entryPrice, d(0.05), nil)
	initialSL, _ := s.GetStopLoss()
	t.Logf("UpTrend - Entry: %v, Initial SL: %v", entryPrice, initialSL)

	previousSL := initialSL

	for i := 1; i < len(data); i++ {
		period := data[i]

		// Update stop loss with current price
		s.CalculateStopLoss(period.Close)

		slTriggered, _ := s.ShouldTriggerStopLoss(period.Low)
		currentSL, _ := s.GetStopLoss()

		t.Logf("Period %d: Price=%v, SL=%v (Trailing)", period.Period, period.Close, currentSL)

		// Verify trailing behavior: SL should only move up
		if currentSL.LessThan(previousSL) {
			t.Errorf("Period %d: Stop loss should not decrease: %v -> %v",
				period.Period, previousSL, currentSL)
		}
		previousSL = currentSL

		if slTriggered {
			t.Logf("Period %d: Stop loss triggered at %v", period.Period, period.Low)
			break
		}
	}
}

func TestFixedTrailingStop_HistoricalData_DownTrend(t *testing.T) {
	data := GetMockHistoricalData()
	entryPrice := data[0].Close

	s, _ := NewFixedTrailingStop(entryPrice, d(0.05), nil)
	initialSL, _ := s.GetStopLoss()
	t.Logf("DownTrend - Entry: %v, Initial SL: %v", entryPrice, initialSL)

	for i := 1; i < len(data); i++ {
		period := data[i]

		s.CalculateStopLoss(period.Close)

		slTriggered, _ := s.ShouldTriggerStopLoss(period.Low)
		currentSL, _ := s.GetStopLoss()

		t.Logf("Period %d: Price=%v, SL=%v", period.Period, period.Close, currentSL)

		if slTriggered {
			t.Logf("Period %d: Stop loss triggered at %v", period.Period, period.Low)
			break
		}
	}
}

func TestFixedTrailingProfit_HistoricalData_UpTrend(t *testing.T) {
	data := GetMockTrendingData()
	entryPrice := data[0].Close

	s, _ := NewFixedTrailingProfit(entryPrice, d(0.10), nil)
	initialTP, _ := s.GetTakeProfit()
	t.Logf("UpTrend - Entry: %v, Initial TP: %v", entryPrice, initialTP)

	previousTP := initialTP

	for i := 1; i < len(data); i++ {
		period := data[i]

		s.CalculateTakeProfit(period.Close)

		tpTriggered, _ := s.ShouldTriggerTakeProfit(period.High)
		currentTP, _ := s.GetTakeProfit()

		t.Logf("Period %d: Price=%v, TP=%v (Trailing)", period.Period, period.Close, currentTP)

		// Verify trailing behavior: TP should only move up
		if currentTP.LessThan(previousTP) {
			t.Errorf("Period %d: Take profit should not decrease: %v -> %v",
				period.Period, previousTP, currentTP)
		}
		previousTP = currentTP

		if tpTriggered {
			t.Logf("Period %d: Take profit triggered at %v", period.Period, period.High)
			break
		}
	}
}

func TestFixedTrailingStop_HistoricalData_Consolidation(t *testing.T) {
	data := GetMockConsolidationData()
	entryPrice := data[0].Close

	s, _ := NewFixedTrailingStop(entryPrice, d(0.05), nil)
	initialSL, _ := s.GetStopLoss()
	t.Logf("Consolidation - Entry: %v, Initial SL: %v", entryPrice, initialSL)

	for i := 1; i < len(data); i++ {
		period := data[i]

		s.CalculateStopLoss(period.Close)

		slTriggered, _ := s.ShouldTriggerStopLoss(period.Low)
		currentSL, _ := s.GetStopLoss()

		t.Logf("Period %d: Price=%v, SL=%v", period.Period, period.Close, currentSL)

		if slTriggered {
			t.Logf("Period %d: Stop loss triggered during consolidation", period.Period)
			break
		}
	}
}

func TestFixedTrailingStop_HistoricalData_VolatileMarket(t *testing.T) {
	data := GetMockVolatileData()
	entryPrice := data[0].Close

	s, _ := NewFixedTrailingStop(entryPrice, d(0.05), nil)
	initialSL, _ := s.GetStopLoss()
	t.Logf("Volatile Market - Entry: %v, Initial SL: %v", entryPrice, initialSL)

	for i := 1; i < len(data); i++ {
		period := data[i]

		s.CalculateStopLoss(period.Close)

		slTriggered, _ := s.ShouldTriggerStopLoss(period.Low)
		currentSL, _ := s.GetStopLoss()

		t.Logf("Period %d: Price=%v, SL=%v, ATR=%v (Volatile)",
			period.Period, period.Close, currentSL, period.ATR)

		if slTriggered {
			t.Logf("Period %d: Stop loss triggered in volatile market", period.Period)
			break
		}
	}
}

func TestFixedTrailingStop_HistoricalData_Recovery(t *testing.T) {
	data := GetMockRecoveryData()
	entryPrice := data[0].Close

	s, _ := NewFixedTrailingStop(entryPrice, d(0.05), nil)
	initialSL, _ := s.GetStopLoss()
	t.Logf("Recovery Scenario - Entry: %v, Initial SL: %v", entryPrice, initialSL)

	maxSL := initialSL

	for i := 1; i < len(data); i++ {
		period := data[i]

		s.CalculateStopLoss(period.Close)

		slTriggered, _ := s.ShouldTriggerStopLoss(period.Low)
		currentSL, _ := s.GetStopLoss()

		if currentSL.GreaterThan(maxSL) {
			maxSL = currentSL
		}

		t.Logf("Period %d: Price=%v, SL=%v, MaxSL=%v",
			period.Period, period.Close, currentSL, maxSL)

		if slTriggered {
			t.Logf("Period %d: Stop loss triggered during recovery", period.Period)
			break
		}
	}

	// In recovery scenario, SL should eventually increase
	if !maxSL.GreaterThan(initialSL) {
		t.Log("Note: Stop loss did not trail up during recovery")
	}
}

func TestFixedTrailingStop_ReSetStopLosser(t *testing.T) {
	entryPrice := d(100)
	rate := d(0.05)

	s, _ := NewFixedTrailingStop(entryPrice, rate, nil)

	// Move price up to trail the stop loss
	s.CalculateStopLoss(d(120))
	slBeforeReset, _ := s.GetStopLoss()
	t.Logf("Stop Loss before reset: %v", slBeforeReset)

	// Reset to new entry price
	newEntryPrice := d(110)
	err := s.ReSetStopLosser(newEntryPrice)
	if err != nil {
		t.Errorf("ReSetStopLosser failed: %v", err)
	}

	// New stop loss: 110 * 0.95 = 104.5
	expectedNewSL := d(104.5)
	newSL, _ := s.GetStopLoss()

	if !newSL.Equal(expectedNewSL) {
		t.Errorf("Expected stop loss %v after reset, got %v", expectedNewSL, newSL)
	}

	t.Logf("After ReSet to %v: New Stop Loss: %v", newEntryPrice, newSL)
}

func TestFixedTrailingProfit_ReSetTakeProfiter(t *testing.T) {
	entryPrice := d(100)
	rate := d(0.10)

	s, _ := NewFixedTrailingProfit(entryPrice, rate, nil)

	// Move price up to trail the take profit
	s.CalculateTakeProfit(d(120))
	tpBeforeReset, _ := s.GetTakeProfit()
	t.Logf("Take Profit before reset: %v", tpBeforeReset)

	// Reset to new entry price
	newEntryPrice := d(110)
	err := s.ReSetTakeProfiter(newEntryPrice)
	if err != nil {
		t.Errorf("ReSetTakeProfiter failed: %v", err)
	}

	// New take profit: 110 * 1.10 = 121
	expectedNewTP := d(121)
	newTP, _ := s.GetTakeProfit()

	if !newTP.Equal(expectedNewTP) {
		t.Errorf("Expected take profit %v after reset, got %v", expectedNewTP, newTP)
	}

	t.Logf("After ReSet to %v: New Take Profit: %v", newEntryPrice, newTP)
}

func TestFixedTrailingStop_Deactivate(t *testing.T) {
	entryPrice := d(100)
	rate := d(0.05)

	s, _ := NewFixedTrailingStop(entryPrice, rate, nil)

	err := s.Deactivate()
	if err != nil {
		t.Errorf("Failed to deactivate: %v", err)
	}

	// After deactivation, should return errors
	_, err = s.CalculateStopLoss(entryPrice)
	if err != stoploss.ErrStatusInvalid {
		t.Errorf("Expected ErrStatusInvalid after deactivation, got %v", err)
	}

	_, err = s.ShouldTriggerStopLoss(d(90))
	if err != stoploss.ErrStatusInvalid {
		t.Errorf("Expected ErrStatusInvalid after deactivation, got %v", err)
	}
}

func TestFixedTrailingProfit_Deactivate(t *testing.T) {
	entryPrice := d(100)
	rate := d(0.10)

	s, _ := NewFixedTrailingProfit(entryPrice, rate, nil)

	err := s.Deactivate()
	if err != nil {
		t.Errorf("Failed to deactivate: %v", err)
	}

	// After deactivation, should return errors
	_, err = s.CalculateTakeProfit(entryPrice)
	if err != stoploss.ErrStatusInvalid {
		t.Errorf("Expected ErrStatusInvalid after deactivation, got %v", err)
	}

	_, err = s.ShouldTriggerTakeProfit(d(120))
	if err != stoploss.ErrStatusInvalid {
		t.Errorf("Expected ErrStatusInvalid after deactivation, got %v", err)
	}
}

// Benchmarks
func BenchmarkNewFixedTrailingStop(b *testing.B) {
	entryPrice := d(100)
	rate := d(0.05)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewFixedTrailingStop(entryPrice, rate, nil)
	}
}

func BenchmarkFixedTrailingStop_CalculateStopLoss(b *testing.B) {
	entryPrice := d(100)
	rate := d(0.05)
	s, _ := NewFixedTrailingStop(entryPrice, rate, nil)
	currentPrice := d(110)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.CalculateStopLoss(currentPrice)
	}
}

func BenchmarkFixedTrailingStop_ShouldTriggerStopLoss(b *testing.B) {
	entryPrice := d(100)
	rate := d(0.05)
	currentPrice := d(94)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s, _ := NewFixedTrailingStop(entryPrice, rate, nil)
		_, _ = s.ShouldTriggerStopLoss(currentPrice)
	}
}

func BenchmarkFixedTrailingStop_ReSetStopLosser(b *testing.B) {
	entryPrice := d(100)
	rate := d(0.05)
	s, _ := NewFixedTrailingStop(entryPrice, rate, nil)
	newPrice := d(110)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.ReSetStopLosser(newPrice)
	}
}
