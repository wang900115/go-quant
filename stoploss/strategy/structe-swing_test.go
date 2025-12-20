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

func TestNewStructureSwingStop_Long(t *testing.T) {
	s, err := NewStructureSwingStop(5, d(100), d(0.02), d(0.95), d(1.10), true, nil)
	if err != nil || s == nil {
		t.Fatalf("Failed to create long strategy: %v", err)
	}
	sl, _ := s.GetStopLoss()
	tp, _ := s.GetTakeProfit()
	t.Logf("Long - Entry:100, SL:%v, TP:%v", sl, tp)
}

func TestNewStructureSwingStop_Short(t *testing.T) {
	s, err := NewStructureSwingStop(5, d(100), d(0.02), d(1.05), d(0.90), false, nil)
	if err != nil || s == nil {
		t.Fatalf("Failed to create short strategy: %v", err)
	}
	sl, _ := s.GetStopLoss()
	tp, _ := s.GetTakeProfit()
	t.Logf("Short - Entry:100, SL:%v, TP:%v", sl, tp)
}

func TestStructureSwing_InvalidParams(t *testing.T) {
	tests := []struct {
		name     string
		lookback int
		swing    decimal.Decimal
		wantErr  bool
	}{
		{"Valid", 5, d(0.02), false},
		{"Invalid lookback", 0, d(0.02), true},
		{"Invalid swing distance", 5, d(0), true},
		{"Negative swing", 5, d(-0.02), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewStructureSwingStop(tt.lookback, d(100), tt.swing, d(0.95), d(1.10), true, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestStructureSwing_LongPosition_UpTrend(t *testing.T) {
	data := GetMockTrendingData()
	s, _ := NewStructureSwingStop(3, data[0].Close, d(0.02), d(0.95), d(1.10), true, nil)

	for i := 1; i < len(data); i++ {
		sl, _ := s.ShouldTriggerStopLoss(data[i].Low)
		tp, _ := s.ShouldTriggerTakeProfit(data[i].High)
		curSL, _ := s.GetStopLoss()
		curTP, _ := s.GetTakeProfit()
		t.Logf("Period %d: Price=%v, SL=%v, TP=%v", i, data[i].Close, curSL, curTP)
		if sl {
			t.Error("SL should not trigger in long uptrend")
			break
		}
		if tp {
			t.Logf("TP triggered at period %d", i)
			break
		}
	}
}

func TestStructureSwing_ShortPosition_UpTrend(t *testing.T) {
	data := GetMockTrendingData()
	s, _ := NewStructureSwingStop(3, data[0].Close, d(0.02), d(1.05), d(0.90), false, nil)

	for i := 1; i < len(data); i++ {
		sl, _ := s.ShouldTriggerStopLoss(data[i].High)
		tp, _ := s.ShouldTriggerTakeProfit(data[i].Low)
		if sl {
			t.Logf("Short SL triggered at period %d (uptrend)", i)
			break
		}
		if tp {
			t.Logf("Short TP triggered at period %d", i)
			break
		}
	}
}

func TestStructureSwing_ReSet(t *testing.T) {
	s, _ := NewStructureSwingStop(5, d(100), d(0.02), d(0.95), d(1.10), true, nil)
	s.ReSet(d(110))
	sl, _ := s.GetStopLoss()
	tp, _ := s.GetTakeProfit()
	t.Logf("After reset to 110: SL=%v, TP=%v", sl, tp)
}

func TestStructureSwing_Deactivate(t *testing.T) {
	s, _ := NewStructureSwingStop(5, d(100), d(0.02), d(0.95), d(1.10), true, nil)
	s.Deactivate()
	_, _, err := s.Calculate(d(100))
	if err != stoploss.ErrStatusInvalid {
		t.Errorf("Expected ErrStatusInvalid, got %v", err)
	}
}

func BenchmarkNewStructureSwingStop(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewStructureSwingStop(5, d(100), d(0.02), d(0.95), d(1.10), true, nil)
	}
}

func BenchmarkStructureSwing_Calculate(b *testing.B) {
	s, _ := NewStructureSwingStop(5, d(100), d(0.02), d(0.95), d(1.10), true, nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Calculate(d(105))
	}
}
