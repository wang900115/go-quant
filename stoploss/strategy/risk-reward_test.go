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

func TestNewRiskRewardRatio(t *testing.T) {
	s, err := NewRiskRewardRatio(d(100), d(0.05), d(0.10), nil)
	if err != nil || s == nil {
		t.Fatalf("Failed to create strategy: %v", err)
	}
	sl, _ := s.GetStopLoss()
	tp, _ := s.GetTakeProfit()
	if !sl.Equal(d(95)) || !tp.Equal(d(110)) {
		t.Errorf("Expected SL=95, TP=110, got SL=%v, TP=%v", sl, tp)
	}
}

func TestRiskRewardRatio_InvalidParams(t *testing.T) {
	tests := []struct {
		name         string
		risk, reward decimal.Decimal
		wantErr      bool
	}{
		{"Valid", d(0.05), d(0.10), false},
		{"Negative risk", d(-0.05), d(0.10), true},
		{"Negative reward", d(0.05), d(-0.10), true},
		{"Risk > 1", d(1.1), d(0.10), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewRiskRewardRatio(d(100), tt.risk, tt.reward, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("wantErr=%v, got err=%v", tt.wantErr, err)
			}
		})
	}
}

func TestRiskRewardRatio_HistoricalData(t *testing.T) {
	upData := GetMockTrendingData()
	downData := GetMockHistoricalData()

	t.Run("UpTrend", func(t *testing.T) {
		s, _ := NewRiskRewardRatio(upData[0].Close, d(0.05), d(0.10), nil)
		for i := 1; i < len(upData); i++ {
			sl, _ := s.ShouldTriggerStopLoss(upData[i].Low)
			tp, _ := s.ShouldTriggerTakeProfit(upData[i].High)
			if sl {
				t.Errorf("SL should not trigger in uptrend at period %d", i)
			}
			if tp {
				t.Logf("TP triggered at period %d", i)
				break
			}
		}
	})

	t.Run("DownTrend", func(t *testing.T) {
		s, _ := NewRiskRewardRatio(downData[0].Close, d(0.05), d(0.10), nil)
		for i := 1; i < len(downData); i++ {
			sl, _ := s.ShouldTriggerStopLoss(downData[i].Low)
			if sl {
				t.Logf("SL triggered at period %d", i)
				break
			}
		}
	})
}

func TestRiskRewardRatio_ReSet(t *testing.T) {
	s, _ := NewRiskRewardRatio(d(100), d(0.05), d(0.10), nil)
	s.ReSet(d(110))
	sl, _ := s.GetStopLoss()
	tp, _ := s.GetTakeProfit()
	if !sl.Equal(d(104.5)) || !tp.Equal(d(121)) {
		t.Errorf("After reset: expected SL=104.5, TP=121, got SL=%v, TP=%v", sl, tp)
	}
}

func TestRiskRewardRatio_Deactivate(t *testing.T) {
	s, _ := NewRiskRewardRatio(d(100), d(0.05), d(0.10), nil)
	s.Deactivate()
	_, _, err := s.Calculate(d(100))
	if err != stoploss.ErrStatusInvalid {
		t.Errorf("Expected ErrStatusInvalid, got %v", err)
	}
}

func BenchmarkNewRiskRewardRatio(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewRiskRewardRatio(d(100), d(0.05), d(0.10), nil)
	}
}

func BenchmarkRiskRewardRatio_Calculate(b *testing.B) {
	s, _ := NewRiskRewardRatio(d(100), d(0.05), d(0.10), nil)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Calculate(d(105))
	}
}
