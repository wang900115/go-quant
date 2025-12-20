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
	"github.com/shopspring/decimal"
)

// PriceData represents price data for a specific time period
type PriceData struct {
	High   decimal.Decimal
	Low    decimal.Decimal
	Close  decimal.Decimal
	ATR    decimal.Decimal
	Period int
}

// d is a helper function to create decimal.Decimal from float64
func d(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

// GetMockHistoricalData
// price data simulating a BTC downtrend
func GetMockHistoricalData() []PriceData {
	return []PriceData{
		// Period 1: entry point
		{High: d(50000), Low: d(49000), Close: d(49500), ATR: d(500), Period: 1},
		// Period 2: price increases
		{High: d(50500), Low: d(49500), Close: d(50200), ATR: d(520), Period: 2},
		// Period 3: continues to rise
		{High: d(51000), Low: d(50000), Close: d(50800), ATR: d(540), Period: 3},
		// Period 4: slight pullback
		{High: d(51200), Low: d(50200), Close: d(50500), ATR: d(560), Period: 4},
		// Period 5: again rises
		{High: d(52000), Low: d(50500), Close: d(51800), ATR: d(580), Period: 5},
		// Period 6: sharp decline begins
		{High: d(52000), Low: d(50000), Close: d(50200), ATR: d(650), Period: 6},
		// Period 7: continues to decline
		{High: d(50500), Low: d(48500), Close: d(49000), ATR: d(700), Period: 7},
		// Period 8: triggers stop-loss (approx. -4% drop)
		{High: d(49500), Low: d(47000), Close: d(47500), ATR: d(750), Period: 8},
	}
}

// GetMockConsolidationData
// price moves within a narrow range, testing stop-loss stability
func GetMockConsolidationData() []PriceData {
	return []PriceData{
		{High: d(100), Low: d(98), Close: d(99), ATR: d(1.5), Period: 1},
		{High: d(101), Low: d(99), Close: d(100.5), ATR: d(1.5), Period: 2},
		{High: d(102), Low: d(100), Close: d(101), ATR: d(1.6), Period: 3},
		{High: d(101.5), Low: d(99.5), Close: d(100), ATR: d(1.5), Period: 4},
		{High: d(100.5), Low: d(98.5), Close: d(99.5), ATR: d(1.5), Period: 5},
		{High: d(101), Low: d(99), Close: d(100), ATR: d(1.5), Period: 6},
		{High: d(100.5), Low: d(98.5), Close: d(99.8), ATR: d(1.5), Period: 7},
	}
}

// GetMockTrendingData
// price steadily increases, testing stop-loss trailing functionality
func GetMockTrendingData() []PriceData {
	return []PriceData{
		{High: d(100), Low: d(98), Close: d(99), ATR: d(2), Period: 1},
		{High: d(102), Low: d(99), Close: d(101), ATR: d(2.1), Period: 2},
		{High: d(104), Low: d(101), Close: d(103), ATR: d(2.2), Period: 3},
		{High: d(106), Low: d(103), Close: d(105), ATR: d(2.3), Period: 4},
		{High: d(108), Low: d(105), Close: d(107), ATR: d(2.4), Period: 5},
		{High: d(110), Low: d(107), Close: d(109), ATR: d(2.5), Period: 6},
		{High: d(112), Low: d(109), Close: d(111), ATR: d(2.6), Period: 7},
		{High: d(115), Low: d(111), Close: d(114), ATR: d(2.7), Period: 8},
	}
}

// GetMockVolatileData
// price experiences significant ups and downs, testing stop-loss adaptability
func GetMockVolatileData() []PriceData {
	return []PriceData{
		{High: d(100), Low: d(95), Close: d(98), ATR: d(3.5), Period: 1},
		{High: d(105), Low: d(97), Close: d(103), ATR: d(4.0), Period: 2},
		{High: d(108), Low: d(102), Close: d(102), ATR: d(4.2), Period: 3},
		{High: d(104), Low: d(96), Close: d(97), ATR: d(4.5), Period: 4},
		{High: d(100), Low: d(93), Close: d(95), ATR: d(5.0), Period: 5},
		{High: d(98), Low: d(88), Close: d(90), ATR: d(5.5), Period: 6},
	}
}

// GetMockGradualDeclineData
// price gradually declines, testing stop-loss trailing functionality
func GetMockGradualDeclineData() []PriceData {
	return []PriceData{
		{High: d(100), Low: d(99), Close: d(99.5), ATR: d(1.0), Period: 1},
		{High: d(99.5), Low: d(98.5), Close: d(99), ATR: d(1.0), Period: 2},
		{High: d(99), Low: d(98), Close: d(98.5), ATR: d(1.0), Period: 3},
		{High: d(98.5), Low: d(97.5), Close: d(98), ATR: d(1.0), Period: 4},
		{High: d(98), Low: d(97), Close: d(97.5), ATR: d(1.0), Period: 5},
		{High: d(97.5), Low: d(96.5), Close: d(97), ATR: d(1.0), Period: 6},
		{High: d(97), Low: d(96), Close: d(96.5), ATR: d(1.0), Period: 7},
		{High: d(96.5), Low: d(95.5), Close: d(96), ATR: d(1.0), Period: 8},
		{High: d(96), Low: d(95), Close: d(95.5), ATR: d(1.0), Period: 9},
		{High: d(95.5), Low: d(94.5), Close: d(95), ATR: d(1.0), Period: 10},
	}
}

// GetMockSharpDropData
// price sharply drops, testing stop-loss reaction
func GetMockSharpDropData() []PriceData {
	return []PriceData{
		{High: d(100), Low: d(99), Close: d(99.5), ATR: d(1.0), Period: 1},
		{High: d(100.5), Low: d(99.5), Close: d(100), ATR: d(1.0), Period: 2},
		{High: d(101), Low: d(100), Close: d(100.5), ATR: d(1.0), Period: 3},
		// Price sharply drops
		{High: d(100), Low: d(90), Close: d(91), ATR: d(8.0), Period: 4},
		{High: d(92), Low: d(88), Close: d(89), ATR: d(9.0), Period: 5},
	}
}

// GetMockRecoveryData
// price drops first then rebounds, testing stop-loss reset functionality
func GetMockRecoveryData() []PriceData {
	return []PriceData{
		{High: d(100), Low: d(98), Close: d(99), ATR: d(1.5), Period: 1},
		{High: d(99), Low: d(96), Close: d(97), ATR: d(2.0), Period: 2},
		{High: d(98), Low: d(94), Close: d(95), ATR: d(2.5), Period: 3},
		// Price rebounds
		{High: d(97), Low: d(95), Close: d(96.5), ATR: d(2.0), Period: 4},
		{High: d(99), Low: d(96.5), Close: d(98), ATR: d(1.8), Period: 5},
		{High: d(101), Low: d(98), Close: d(100), ATR: d(1.6), Period: 6},
		{High: d(103), Low: d(100), Close: d(102), ATR: d(1.5), Period: 7},
	}
}
