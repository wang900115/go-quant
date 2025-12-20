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

func TestNewFixedATRStop(t *testing.T) {
	s, err := NewFixedATRStop(d(100), d(2), d(2), nil)
	if err != nil || s == nil {
		t.Fatalf("Failed to create ATR stop: %v", err)
	}
	sl, _ := s.GetStopLoss()
	expected := d(96) // 100 - (2 * 2)
	if !sl.Equal(expected) {
		t.Errorf("Expected SL=%v, got %v", expected, sl)
	}
}

func TestNewFixedATRProfit(t *testing.T) {
	s, err := NewFixedATRProfit(d(100), d(2), d(3), nil)
	if err != nil || s == nil {
		t.Fatalf("Failed to create ATR profit: %v", err)
	}
	tp, _ := s.GetTakeProfit()
	expected := d(106) // 100 + (3 * 2)
	if !tp.Equal(expected) {
		t.Errorf("Expected TP=%v, got %v", expected, tp)
	}
}

func TestFixedATRStop_InvalidParams(t *testing.T) {
	tests := []struct {
		name    string
		atr, k  decimal.Decimal
		wantErr bool
	}{
		{"Valid", d(2), d(2), false},
		{"Negative ATR", d(-2), d(2), true},
		{"Negative K", d(2), d(-2), true},
		{"Zero K", d(2), d(0), true},
		{"Zero ATR", d(0), d(2), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFixedATRStop(d(100), tt.atr, tt.k, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestFixedATRProfit_InvalidParams(t *testing.T) {
	tests := []struct {
		name    string
		atr, k  decimal.Decimal
		wantErr bool
	}{
		{"Valid", d(2), d(3), false},
		{"Negative ATR", d(-2), d(3), true},
		{"Negative K", d(2), d(-3), true},
		{"Zero K", d(2), d(0), true},
		{"Zero ATR", d(0), d(3), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewFixedATRProfit(d(100), tt.atr, tt.k, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestFixedATRStop_UpdateATR(t *testing.T) {
	s, _ := NewFixedATRStop(d(100), d(2), d(2), nil)
	initialSL, _ := s.GetStopLoss()

	s.UpdateATR(d(3))
	s.CalculateStopLoss(d(100))
	newSL, _ := s.GetStopLoss()
	expectedSL := d(94) // 100 - (2 * 3)

	if !newSL.Equal(expectedSL) {
		t.Errorf("After UpdateATR: expected SL=%v, got %v", expectedSL, newSL)
	}
	if newSL.GreaterThanOrEqual(initialSL) {
		t.Error("SL should decrease with higher ATR")
	}
}

func TestFixedATRStop_HistoricalData(t *testing.T) {
	data := GetMockHistoricalData()
	s, _ := NewFixedATRStop(data[0].Close, data[0].ATR, d(2), nil)

	for i := 1; i < len(data); i++ {
		s.UpdateATR(data[i].ATR)
		sl, _ := s.ShouldTriggerStopLoss(data[i].Low)
		curSL, _ := s.GetStopLoss()
		t.Logf("Period %d: Price=%v, SL=%v, ATR=%v", i, data[i].Close, curSL, data[i].ATR)
		if sl {
			t.Logf("SL triggered at period %d", i)
			break
		}
	}
}

func TestFixedATRStop_VolatileMarket(t *testing.T) {
	data := GetMockVolatileData()
	s, _ := NewFixedATRStop(data[0].Close, data[0].ATR, d(1.5), nil)

	for i := 1; i < len(data); i++ {
		s.UpdateATR(data[i].ATR)
		sl, _ := s.ShouldTriggerStopLoss(data[i].Low)
		curSL, _ := s.GetStopLoss()
		t.Logf("Period %d: Price=%v, SL=%v, ATR=%v (volatile)", i, data[i].Close, curSL, data[i].ATR)
		if sl {
			t.Logf("SL triggered in volatile market at period %d", i)
			break
		}
	}
}

func TestFixedATRStop_ReSetStopLosser(t *testing.T) {
	s, _ := NewFixedATRStop(d(100), d(2), d(2), nil)
	s.ReSetStopLosser(d(110))
	newSL, _ := s.GetStopLoss()
	expected := d(106) // 110 - (2 * 2)
	if !newSL.Equal(expected) {
		t.Errorf("Expected SL=%v, got %v", expected, newSL)
	}
}

func TestFixedATRStop_Deactivate(t *testing.T) {
	s, _ := NewFixedATRStop(d(100), d(2), d(2), nil)
	s.Deactivate()
	_, err := s.CalculateStopLoss(d(100))
	if err != stoploss.ErrStatusInvalid {
		t.Errorf("Expected ErrStatusInvalid, got %v", err)
	}
}

func BenchmarkNewFixedATRStop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewFixedATRStop(d(100), d(2), d(2), nil)
	}
}

func BenchmarkFixedATRStop_UpdateATR(b *testing.B) {
	s, _ := NewFixedATRStop(d(100), d(2), d(2), nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.UpdateATR(d(2.5))
	}
}
