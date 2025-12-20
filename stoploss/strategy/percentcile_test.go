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

func TestNewFixedPercentStop(t *testing.T) {
	s, err := NewFixedPercentStop(d(100), d(0.05), nil)
	if err != nil {
		t.Fatalf("Failed to create fixed percent stop: %v", err)
	}
	if s == nil {
		t.Fatal("Strategy is nil")
	}

	sl, _ := s.GetStopLoss()
	expected := d(95) // 100 * (1 - 0.05)
	if !sl.Equal(expected) {
		t.Errorf("Expected SL=%v, got %v", expected, sl)
	}
	t.Logf("Entry: 100, Stop Loss: %v (5%%)", sl)
}

func TestNewFixedPercentProfit(t *testing.T) {
	s, err := NewFixedPercentProfit(d(100), d(0.10), nil)
	if err != nil {
		t.Fatalf("Failed to create fixed percent profit: %v", err)
	}
	if s == nil {
		t.Fatal("Strategy is nil")
	}

	tp, _ := s.GetTakeProfit()
	expected := d(110) // 100 * (1 + 0.10)
	if !tp.Equal(expected) {
		t.Errorf("Expected TP=%v, got %v", expected, tp)
	}
	t.Logf("Entry: 100, Take Profit: %v (10%%)", tp)
}

func TestFixedPercentStop_InvalidParameters(t *testing.T) {
	tests := []struct {
		name        string
		entryPrice  decimal.Decimal
		pct         decimal.Decimal
		expectError bool
	}{
		{"Valid 5% stop", d(100), d(0.05), false},
		{"Valid 10% stop", d(200), d(0.10), false},
		{"Negative percentage", d(100), d(-0.05), true},
		{"Percentage > 1", d(100), d(1.1), true},
		{"Zero percentage", d(100), d(0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFixedPercentStop(tt.entryPrice, tt.pct, nil)
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestFixedPercentProfit_InvalidParameters(t *testing.T) {
	tests := []struct {
		name        string
		entryPrice  decimal.Decimal
		pct         decimal.Decimal
		expectError bool
	}{
		{"Valid 10% profit", d(100), d(0.10), false},
		{"Valid 15% profit", d(200), d(0.15), false},
		{"Negative percentage", d(100), d(-0.10), true},
		{"Percentage > 1", d(100), d(1.1), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFixedPercentProfit(tt.entryPrice, tt.pct, nil)
			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestFixedPercentStop_CalculateStopLoss(t *testing.T) {
	tests := []struct {
		name       string
		entryPrice decimal.Decimal
		pct        decimal.Decimal
		expectedSL decimal.Decimal
	}{
		{"5% stop from 100", d(100), d(0.05), d(95)},
		{"10% stop from 200", d(200), d(0.10), d(180)},
		{"3% stop from 50", d(50), d(0.03), d(48.5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := NewFixedPercentStop(tt.entryPrice, tt.pct, nil)
			sl, err := s.CalculateStopLoss(tt.entryPrice)
			if err != nil {
				t.Errorf("CalculateStopLoss failed: %v", err)
			}
			if !sl.Equal(tt.expectedSL) {
				t.Errorf("Expected SL=%v, got %v", tt.expectedSL, sl)
			}
		})
	}
}

func TestFixedPercentProfit_CalculateTakeProfit(t *testing.T) {
	tests := []struct {
		name       string
		entryPrice decimal.Decimal
		pct        decimal.Decimal
		expectedTP decimal.Decimal
	}{
		{"10% profit from 100", d(100), d(0.10), d(110)},
		{"15% profit from 200", d(200), d(0.15), d(230)},
		{"5% profit from 50", d(50), d(0.05), d(52.5)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, _ := NewFixedPercentProfit(tt.entryPrice, tt.pct, nil)
			tp, err := s.CalculateTakeProfit(tt.entryPrice)
			if err != nil {
				t.Errorf("CalculateTakeProfit failed: %v", err)
			}
			if !tp.Equal(tt.expectedTP) {
				t.Errorf("Expected TP=%v, got %v", tt.expectedTP, tp)
			}
		})
	}
}

func TestFixedPercentStop_ShouldTriggerStopLoss(t *testing.T) {
	entryPrice := d(100)
	pct := d(0.05)

	s, _ := NewFixedPercentStop(entryPrice, pct, nil)
	sl, _ := s.GetStopLoss() // Should be 95

	tests := []struct {
		name     string
		price    decimal.Decimal
		expected bool
	}{
		{"Price above SL", d(96), false},
		{"Price at SL", d(95), true},
		{"Price below SL", d(94), true},
		{"Price at entry", d(100), false},
		{"Price much below SL", d(90), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			triggered, err := s.ShouldTriggerStopLoss(tt.price)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if triggered != tt.expected {
				t.Errorf("Price %v: expected trigger=%v, got %v (SL=%v)",
					tt.price, tt.expected, triggered, sl)
			}
		})
	}
}

func TestFixedPercentProfit_ShouldTriggerTakeProfit(t *testing.T) {
	entryPrice := d(100)
	pct := d(0.10)

	s, _ := NewFixedPercentProfit(entryPrice, pct, nil)
	tp, _ := s.GetTakeProfit() // Should be 110

	tests := []struct {
		name     string
		price    decimal.Decimal
		expected bool
	}{
		{"Price below TP", d(109), false},
		{"Price at TP", d(110), true},
		{"Price above TP", d(111), true},
		{"Price at entry", d(100), false},
		{"Price much above TP", d(120), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			triggered, err := s.ShouldTriggerTakeProfit(tt.price)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if triggered != tt.expected {
				t.Errorf("Price %v: expected trigger=%v, got %v (TP=%v)",
					tt.price, tt.expected, triggered, tp)
			}
		})
	}
}

func TestFixedPercentStop_HistoricalData_DownTrend(t *testing.T) {
	data := GetMockHistoricalData()
	entryPrice := data[0].Close

	s, _ := NewFixedPercentStop(entryPrice, d(0.05), nil)
	sl, _ := s.GetStopLoss()
	t.Logf("Entry: %v, Stop Loss: %v (5%%)", entryPrice, sl)

	for i := 1; i < len(data); i++ {
		period := data[i]

		slTriggered, _ := s.ShouldTriggerStopLoss(period.Low)

		pnlPercent := period.Close.Sub(entryPrice).Div(entryPrice).Mul(d(100))
		t.Logf("Period %d: Close=%v, Low=%v, PnL=%.2f%%, SL=%v",
			period.Period, period.Close, period.Low, pnlPercent.InexactFloat64(), sl)

		if slTriggered {
			t.Logf("Period %d: Stop loss triggered at %v", period.Period, period.Low)
			break
		}
	}
}

func TestFixedPercentProfit_HistoricalData_UpTrend(t *testing.T) {
	data := GetMockTrendingData()
	entryPrice := data[0].Close

	s, _ := NewFixedPercentProfit(entryPrice, d(0.10), nil)
	tp, _ := s.GetTakeProfit()
	t.Logf("Entry: %v, Take Profit: %v (10%%)", entryPrice, tp)

	for i := 1; i < len(data); i++ {
		period := data[i]

		tpTriggered, _ := s.ShouldTriggerTakeProfit(period.High)

		pnlPercent := period.Close.Sub(entryPrice).Div(entryPrice).Mul(d(100))
		t.Logf("Period %d: Close=%v, High=%v, PnL=%.2f%%, TP=%v",
			period.Period, period.Close, period.High, pnlPercent.InexactFloat64(), tp)

		if tpTriggered {
			t.Logf("Period %d: Take profit triggered at %v", period.Period, period.High)
			break
		}
	}
}

func TestFixedPercentStop_HistoricalData_GradualDecline(t *testing.T) {
	data := GetMockGradualDeclineData()
	entryPrice := data[0].Close

	s, _ := NewFixedPercentStop(entryPrice, d(0.05), nil)
	sl, _ := s.GetStopLoss()
	t.Logf("Gradual Decline - Entry: %v, SL: %v", entryPrice, sl)

	for i := 1; i < len(data); i++ {
		period := data[i]

		slTriggered, _ := s.ShouldTriggerStopLoss(period.Low)

		t.Logf("Period %d: Price=%v, SL=%v", period.Period, period.Close, sl)

		if slTriggered {
			t.Logf("Period %d: Stop loss triggered during gradual decline", period.Period)
			break
		}
	}
}

func TestFixedPercentStop_HistoricalData_SharpDrop(t *testing.T) {
	data := GetMockSharpDropData()
	entryPrice := data[0].Close

	s, _ := NewFixedPercentStop(entryPrice, d(0.05), nil)
	sl, _ := s.GetStopLoss()
	t.Logf("Sharp Drop - Entry: %v, SL: %v", entryPrice, sl)

	for i := 1; i < len(data); i++ {
		period := data[i]

		slTriggered, _ := s.ShouldTriggerStopLoss(period.Low)

		dropPercent := period.Low.Sub(entryPrice).Div(entryPrice).Mul(d(100))
		t.Logf("Period %d: Price=%v, Low=%v, Drop=%.2f%%, SL=%v",
			period.Period, period.Close, period.Low, dropPercent.InexactFloat64(), sl)

		if slTriggered {
			t.Logf("Period %d: Stop loss triggered in sharp drop", period.Period)
			break
		}
	}
}

func TestFixedPercentStop_HistoricalData_Consolidation(t *testing.T) {
	data := GetMockConsolidationData()
	entryPrice := data[0].Close

	s, _ := NewFixedPercentStop(entryPrice, d(0.05), nil)
	sl, _ := s.GetStopLoss()
	t.Logf("Consolidation - Entry: %v, SL: %v", entryPrice, sl)

	triggered := false
	for i := 1; i < len(data); i++ {
		period := data[i]

		slTriggered, _ := s.ShouldTriggerStopLoss(period.Low)

		t.Logf("Period %d: Price=%v, SL=%v", period.Period, period.Close, sl)

		if slTriggered {
			triggered = true
			t.Logf("Period %d: Stop loss triggered during consolidation", period.Period)
			break
		}
	}

	if !triggered {
		t.Log("✓ No stop loss trigger during consolidation (expected)")
	}
}

func TestFixedPercentStop_HistoricalData_VolatileMarket(t *testing.T) {
	data := GetMockVolatileData()
	entryPrice := data[0].Close

	s, _ := NewFixedPercentStop(entryPrice, d(0.05), nil)
	sl, _ := s.GetStopLoss()
	t.Logf("Volatile Market - Entry: %v, SL: %v", entryPrice, sl)

	for i := 1; i < len(data); i++ {
		period := data[i]

		slTriggered, _ := s.ShouldTriggerStopLoss(period.Low)

		t.Logf("Period %d: Price=%v, High=%v, Low=%v, ATR=%v, SL=%v",
			period.Period, period.Close, period.High, period.Low, period.ATR, sl)

		if slTriggered {
			t.Logf("Period %d: Stop loss triggered in volatile market", period.Period)
			break
		}
	}
}

func TestFixedPercentStop_ReSetStopLosser(t *testing.T) {
	entryPrice := d(100)
	pct := d(0.05)

	s, _ := NewFixedPercentStop(entryPrice, pct, nil)

	initialSL, _ := s.GetStopLoss()
	t.Logf("Initial Stop Loss: %v", initialSL)

	// Reset to new price
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

func TestFixedPercentProfit_ReSetTakeProfiter(t *testing.T) {
	entryPrice := d(100)
	pct := d(0.10)

	s, _ := NewFixedPercentProfit(entryPrice, pct, nil)

	initialTP, _ := s.GetTakeProfit()
	t.Logf("Initial Take Profit: %v", initialTP)

	// Reset to new price
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

func TestFixedPercentStop_Deactivate(t *testing.T) {
	entryPrice := d(100)
	pct := d(0.05)

	s, _ := NewFixedPercentStop(entryPrice, pct, nil)

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

func TestFixedPercentProfit_Deactivate(t *testing.T) {
	entryPrice := d(100)
	pct := d(0.10)

	s, _ := NewFixedPercentProfit(entryPrice, pct, nil)

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
func BenchmarkNewFixedPercentStop(b *testing.B) {
	entryPrice := d(100)
	pct := d(0.05)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = NewFixedPercentStop(entryPrice, pct, nil)
	}
}

func BenchmarkFixedPercentStop_CalculateStopLoss(b *testing.B) {
	entryPrice := d(100)
	pct := d(0.05)
	s, _ := NewFixedPercentStop(entryPrice, pct, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.CalculateStopLoss(entryPrice)
	}
}

func BenchmarkFixedPercentStop_ShouldTriggerStopLoss(b *testing.B) {
	entryPrice := d(100)
	pct := d(0.05)
	currentPrice := d(94)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s, _ := NewFixedPercentStop(entryPrice, pct, nil)
		_, _ = s.ShouldTriggerStopLoss(currentPrice)
	}
}

func BenchmarkFixedPercentStop_ReSetStopLosser(b *testing.B) {
	entryPrice := d(100)
	pct := d(0.05)
	s, _ := NewFixedPercentStop(entryPrice, pct, nil)
	newPrice := d(110)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.ReSetStopLosser(newPrice)
	}
}
