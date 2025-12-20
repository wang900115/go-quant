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

func TestNewFixedMovingAverageStop(t *testing.T) {
	s, err := NewFixedMovingAverageStop(d(100), d(95), d(0.02), nil)
	if err != nil {
		t.Fatalf("Failed to create MA stop: %v", err)
	}
	if s == nil {
		t.Fatal("Strategy is nil")
	}

	sl, _ := s.GetStopLoss()
	expected := d(93.1) // 95 * (1 - 0.02) = 93.1
	if !sl.Equal(expected) {
		t.Errorf("Expected SL=%v, got %v", expected, sl)
	}
	t.Logf("Entry: 100, MA: 95, Offset: 2%%, Initial Stop Loss: %v", sl)
}

func TestNewFixedMovingAverageProfit(t *testing.T) {
	s, err := NewFixedMovingAverageProfit(d(100), d(105), d(0.03), nil)
	if err != nil {
		t.Fatalf("Failed to create MA profit: %v", err)
	}
	if s == nil {
		t.Fatal("Strategy is nil")
	}

	tp, _ := s.GetTakeProfit()
	expected := d(108.15) // 105 * (1 + 0.03) = 108.15
	if !tp.Equal(expected) {
		t.Errorf("Expected TP=%v, got %v", expected, tp)
	}
	t.Logf("Entry: 100, MA: 105, Offset: 3%%, Initial Take Profit: %v", tp)
}

func TestFixedMovingAverageStop_InvalidParameters(t *testing.T) {
	tests := []struct {
		name        string
		entryPrice  decimal.Decimal
		initialMA   decimal.Decimal
		offset      decimal.Decimal
		expectError bool
	}{
		{"Valid MA and offset", d(100), d(95), d(0.02), false},
		{"Negative MA", d(100), d(-95), d(0.02), true},
		{"Zero MA", d(100), d(0), d(0.02), true},
		{"Negative offset", d(100), d(95), d(-0.02), true},
		{"Offset > 1", d(100), d(95), d(1.5), true},
		{"Zero offset", d(100), d(95), d(0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFixedMovingAverageStop(tt.entryPrice, tt.initialMA, tt.offset, nil)
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestFixedMovingAverageProfit_InvalidParameters(t *testing.T) {
	tests := []struct {
		name        string
		entryPrice  decimal.Decimal
		initialMA   decimal.Decimal
		offset      decimal.Decimal
		expectError bool
	}{
		{"Valid MA and offset", d(100), d(105), d(0.03), false},
		{"Negative MA", d(100), d(-105), d(0.03), true},
		{"Zero MA", d(100), d(0), d(0.03), true},
		{"Negative offset", d(100), d(105), d(-0.03), true},
		{"Offset > 1", d(100), d(105), d(1.5), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFixedMovingAverageProfit(tt.entryPrice, tt.initialMA, tt.offset, nil)
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestFixedMovingAverageStop_SetMA(t *testing.T) {
	entryPrice := d(100)
	initialMA := d(95)
	offset := d(0.02)

	s, _ := NewFixedMovingAverageStop(entryPrice, initialMA, offset, nil)
	initialSL, _ := s.GetStopLoss()
	t.Logf("Initial Stop Loss (MA=%v): %v", initialMA, initialSL)

	// Update MA to new value
	newMA := d(98)
	s.SetMA(newMA)
	s.CalculateStopLoss(d(100))
	newSL, _ := s.GetStopLoss()
	expectedSL := d(96.04) // 98 * 0.98
	t.Logf("After SetMA to %v: New Stop Loss: %v (expected %v)", newMA, newSL, expectedSL)

	if !newSL.Equal(expectedSL) {
		t.Errorf("Expected SL=%v after MA update, got %v", expectedSL, newSL)
	}
}

func TestFixedMovingAverageProfit_SetMA(t *testing.T) {
	entryPrice := d(100)
	initialMA := d(105)
	offset := d(0.03)

	s, _ := NewFixedMovingAverageProfit(entryPrice, initialMA, offset, nil)
	initialTP, _ := s.GetTakeProfit()
	t.Logf("Initial Take Profit (MA=%v): %v", initialMA, initialTP)

	// Update MA to new value
	newMA := d(110)
	s.SetMA(newMA)
	s.CalculateTakeProfit(d(100))
	newTP, _ := s.GetTakeProfit()
	expectedTP := d(113.3) // 110 * 1.03
	t.Logf("After SetMA to %v: New Take Profit: %v (expected %v)", newMA, newTP, expectedTP)

	if !newTP.Equal(expectedTP) {
		t.Errorf("Expected TP=%v after MA update, got %v", expectedTP, newTP)
	}
}

func TestFixedMovingAverageStop_Calculate(t *testing.T) {
	entryPrice := d(100)
	initialMA := d(95)
	offset := d(0.05)

	s, _ := NewFixedMovingAverageStop(entryPrice, initialMA, offset, nil)

	// Test calculation with different MA values
	tests := []struct {
		name       string
		newMA      decimal.Decimal
		expectedSL decimal.Decimal
	}{
		{"MA increases to 98", d(98), d(93.1)}, // 98 * 0.95
		{"MA increases to 100", d(100), d(95)}, // 100 * 0.95
		{"MA decreases to 92", d(92), d(87.4)}, // 92 * 0.95
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s.SetMA(tt.newMA)
			sl, _ := s.CalculateStopLoss(d(100))

			if !sl.Equal(tt.expectedSL) {
				t.Errorf("Expected SL=%v, got %v", tt.expectedSL, sl)
			}
			t.Logf("MA=%v -> SL=%v", tt.newMA, sl)
		})
	}
}

func TestFixedMovingAverageStop_ShouldTrigger(t *testing.T) {
	entryPrice := d(100)
	initialMA := d(95)
	offset := d(0.05)

	s, _ := NewFixedMovingAverageStop(entryPrice, initialMA, offset, nil)
	s.CalculateStopLoss(entryPrice)
	sl, _ := s.GetStopLoss()
	t.Logf("Stop Loss: %v", sl)

	tests := []struct {
		name          string
		currentPrice  decimal.Decimal
		shouldTrigger bool
	}{
		{"Price well above SL", d(100), false},
		{"Price slightly above SL", d(91), false},
		{"Price at SL", sl, true},
		{"Price below SL", d(89), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh instance for each test
			s, _ := NewFixedMovingAverageStop(entryPrice, initialMA, offset, nil)
			s.CalculateStopLoss(entryPrice)

			triggered, _ := s.ShouldTriggerStopLoss(tt.currentPrice)
			if triggered != tt.shouldTrigger {
				t.Errorf("Price %v: expected trigger=%v, got %v", tt.currentPrice, tt.shouldTrigger, triggered)
			}
		})
	}
}

func TestFixedMovingAverageProfit_ShouldTrigger(t *testing.T) {
	entryPrice := d(100)
	initialMA := d(105)
	offset := d(0.05)

	s, _ := NewFixedMovingAverageProfit(entryPrice, initialMA, offset, nil)
	s.CalculateTakeProfit(entryPrice)
	tp, _ := s.GetTakeProfit()
	t.Logf("Take Profit: %v", tp)

	tests := []struct {
		name          string
		currentPrice  decimal.Decimal
		shouldTrigger bool
	}{
		{"Price well below TP", d(100), false},
		{"Price slightly below TP", d(109), false},
		{"Price at TP", tp, true},
		{"Price above TP", d(112), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh instance for each test
			s, _ := NewFixedMovingAverageProfit(entryPrice, initialMA, offset, nil)
			s.CalculateTakeProfit(entryPrice)

			triggered, _ := s.ShouldTriggerTakeProfit(tt.currentPrice)
			if triggered != tt.shouldTrigger {
				t.Errorf("Price %v: expected trigger=%v, got %v", tt.currentPrice, tt.shouldTrigger, triggered)
			}
		})
	}
}

func TestFixedMovingAverageStop_HistoricalData_UpTrend(t *testing.T) {
	data := GetMockTrendingData()
	entryPrice := data[0].Close
	initialMA := entryPrice.Mul(d(0.98)) // MA slightly below entry
	offset := d(0.02)

	s, _ := NewFixedMovingAverageStop(entryPrice, initialMA, offset, nil)
	initialSL, _ := s.GetStopLoss()
	t.Logf("UpTrend - Entry: %v, Initial MA: %v, SL: %v", entryPrice, initialMA, initialSL)

	for i := 1; i < len(data); i++ {
		period := data[i]

		// Simulate MA moving up with price (simplified)
		simpleMA := period.Close.Mul(d(0.98))
		s.SetMA(simpleMA)
		s.CalculateStopLoss(period.Close)

		slTriggered, _ := s.ShouldTriggerStopLoss(period.Low)
		currentSL, _ := s.GetStopLoss()

		t.Logf("Period %d: Price=%v, MA=%v, SL=%v", period.Period, period.Close, simpleMA, currentSL)

		if slTriggered {
			t.Logf("Period %d: Stop loss triggered at %v", period.Period, period.Low)
			break
		}
	}
}

func TestFixedMovingAverageStop_HistoricalData_DownTrend(t *testing.T) {
	data := GetMockHistoricalData()
	entryPrice := data[0].Close
	initialMA := entryPrice.Mul(d(0.98))
	offset := d(0.02)

	s, _ := NewFixedMovingAverageStop(entryPrice, initialMA, offset, nil)
	initialSL, _ := s.GetStopLoss()
	t.Logf("DownTrend - Entry: %v, Initial MA: %v, SL: %v", entryPrice, initialMA, initialSL)

	for i := 1; i < len(data); i++ {
		period := data[i]

		// MA follows price down
		simpleMA := period.Close.Mul(d(0.98))
		s.SetMA(simpleMA)
		s.CalculateStopLoss(period.Close)

		slTriggered, _ := s.ShouldTriggerStopLoss(period.Low)
		currentSL, _ := s.GetStopLoss()

		t.Logf("Period %d: Price=%v, MA=%v, SL=%v", period.Period, period.Close, simpleMA, currentSL)

		if slTriggered {
			t.Logf("Period %d: Stop loss triggered at %v", period.Period, period.Low)
			break
		}
	}
}

func TestFixedMovingAverageProfit_HistoricalData_UpTrend(t *testing.T) {
	data := GetMockTrendingData()
	entryPrice := data[0].Close
	initialMA := entryPrice.Mul(d(1.02)) // MA slightly above entry
	offset := d(0.03)

	s, _ := NewFixedMovingAverageProfit(entryPrice, initialMA, offset, nil)
	initialTP, _ := s.GetTakeProfit()
	t.Logf("UpTrend - Entry: %v, Initial MA: %v, TP: %v", entryPrice, initialMA, initialTP)

	for i := 1; i < len(data); i++ {
		period := data[i]

		// MA moves up with price
		simpleMA := period.Close.Mul(d(1.02))
		s.SetMA(simpleMA)
		s.CalculateTakeProfit(period.Close)

		tpTriggered, _ := s.ShouldTriggerTakeProfit(period.High)
		currentTP, _ := s.GetTakeProfit()

		t.Logf("Period %d: Price=%v, MA=%v, TP=%v", period.Period, period.Close, simpleMA, currentTP)

		if tpTriggered {
			t.Logf("Period %d: Take profit triggered at %v", period.Period, period.High)
			break
		}
	}
}

func TestFixedMovingAverageStop_HistoricalData_Consolidation(t *testing.T) {
	data := GetMockConsolidationData()
	entryPrice := data[0].Close
	initialMA := entryPrice
	offset := d(0.03)

	s, _ := NewFixedMovingAverageStop(entryPrice, initialMA, offset, nil)
	initialSL, _ := s.GetStopLoss()
	t.Logf("Consolidation - Entry: %v, MA: %v, SL: %v", entryPrice, initialMA, initialSL)

	for i := 1; i < len(data); i++ {
		period := data[i]

		// MA stays relatively flat during consolidation
		simpleMA := period.Close.Mul(d(0.99))
		s.SetMA(simpleMA)
		s.CalculateStopLoss(period.Close)

		slTriggered, _ := s.ShouldTriggerStopLoss(period.Low)
		currentSL, _ := s.GetStopLoss()

		t.Logf("Period %d: Price=%v, MA=%v, SL=%v", period.Period, period.Close, simpleMA, currentSL)

		if slTriggered {
			t.Logf("Period %d: Stop loss triggered during consolidation", period.Period)
			break
		}
	}
}

func TestFixedMovingAverageStop_HistoricalData_VolatileMarket(t *testing.T) {
	data := GetMockVolatileData()
	entryPrice := data[0].Close
	initialMA := entryPrice
	offset := d(0.05)

	s, _ := NewFixedMovingAverageStop(entryPrice, initialMA, offset, nil)
	initialSL, _ := s.GetStopLoss()
	t.Logf("Volatile Market - Entry: %v, MA: %v, SL: %v", entryPrice, initialMA, initialSL)

	for i := 1; i < len(data); i++ {
		period := data[i]

		// MA lags price in volatile market
		simpleMA := period.Close.Mul(d(0.97))
		s.SetMA(simpleMA)
		s.CalculateStopLoss(period.Close)

		slTriggered, _ := s.ShouldTriggerStopLoss(period.Low)
		currentSL, _ := s.GetStopLoss()

		t.Logf("Period %d: Price=%v, MA=%v, SL=%v, ATR=%v (Volatile)",
			period.Period, period.Close, simpleMA, currentSL, period.ATR)

		if slTriggered {
			t.Logf("Period %d: Stop loss triggered in volatile market", period.Period)
			break
		}
	}
}

func TestFixedMovingAverageStop_ReSetStopLosser(t *testing.T) {
	entryPrice := d(100)
	initialMA := d(95)
	offset := d(0.05)

	s, _ := NewFixedMovingAverageStop(entryPrice, initialMA, offset, nil)

	// Update MA
	s.SetMA(d(98))
	s.CalculateStopLoss(d(105))
	slBeforeReset, _ := s.GetStopLoss()
	t.Logf("Stop Loss before reset: %v", slBeforeReset)

	// Reset with new entry price
	newEntryPrice := d(110)
	err := s.ReSetStopLosser(newEntryPrice)
	if err != nil {
		t.Errorf("ReSetStopLosser failed: %v", err)
	}

	// New stop loss should be based on current MA
	expectedNewSL := d(93.1) // 98 * 0.95
	newSL, _ := s.GetStopLoss()

	if !newSL.Equal(expectedNewSL) {
		t.Errorf("Expected stop loss %v after reset, got %v", expectedNewSL, newSL)
	}

	t.Logf("After ReSet to %v: New Stop Loss: %v", newEntryPrice, newSL)
}

func TestFixedMovingAverageProfit_ReSetTakeProfiter(t *testing.T) {
	entryPrice := d(100)
	initialMA := d(105)
	offset := d(0.05)

	s, _ := NewFixedMovingAverageProfit(entryPrice, initialMA, offset, nil)

	// Update MA
	s.SetMA(d(110))
	s.CalculateTakeProfit(d(100))
	tpBeforeReset, _ := s.GetTakeProfit()
	t.Logf("Take Profit before reset: %v", tpBeforeReset)

	// Reset with new entry price
	newEntryPrice := d(105)
	err := s.ReSetTakeProfiter(newEntryPrice)
	if err != nil {
		t.Errorf("ReSetTakeProfiter failed: %v", err)
	}

	// New take profit should be based on current MA
	expectedNewTP := d(115.5) // 110 * 1.05
	newTP, _ := s.GetTakeProfit()

	if !newTP.Equal(expectedNewTP) {
		t.Errorf("Expected take profit %v after reset, got %v", expectedNewTP, newTP)
	}

	t.Logf("After ReSet to %v: New Take Profit: %v", newEntryPrice, newTP)
}

func TestFixedMovingAverageStop_Deactivate(t *testing.T) {
	entryPrice := d(100)
	initialMA := d(95)
	offset := d(0.05)

	s, _ := NewFixedMovingAverageStop(entryPrice, initialMA, offset, nil)

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

func TestFixedMovingAverageProfit_Deactivate(t *testing.T) {
	entryPrice := d(100)
	initialMA := d(105)
	offset := d(0.05)

	s, _ := NewFixedMovingAverageProfit(entryPrice, initialMA, offset, nil)

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
func BenchmarkNewFixedMovingAverageStop(b *testing.B) {
	entryPrice := d(100)
	initialMA := d(95)
	offset := d(0.05)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewFixedMovingAverageStop(entryPrice, initialMA, offset, nil)
	}
}

func BenchmarkFixedMovingAverageStop_SetMA(b *testing.B) {
	entryPrice := d(100)
	initialMA := d(95)
	offset := d(0.05)
	s, _ := NewFixedMovingAverageStop(entryPrice, initialMA, offset, nil)
	newMA := d(98)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SetMA(newMA)
	}
}

func BenchmarkFixedMovingAverageStop_CalculateStopLoss(b *testing.B) {
	entryPrice := d(100)
	initialMA := d(95)
	offset := d(0.05)
	s, _ := NewFixedMovingAverageStop(entryPrice, initialMA, offset, nil)
	currentPrice := d(105)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.CalculateStopLoss(currentPrice)
	}
}

func BenchmarkFixedMovingAverageStop_ShouldTriggerStopLoss(b *testing.B) {
	entryPrice := d(100)
	initialMA := d(95)
	offset := d(0.05)
	currentPrice := d(90)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s, _ := NewFixedMovingAverageStop(entryPrice, initialMA, offset, nil)
		_, _ = s.ShouldTriggerStopLoss(currentPrice)
	}
}
