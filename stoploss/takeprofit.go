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

package stoploss

import "github.com/shopspring/decimal"

type HybridWithoutTime interface {
	Hybrid
	TakeProfitCond
	StopLossCond
}

type HybridWithTime interface {
	Hybrid
	TakeProfitCondT
	StopLossCondT
}

type Hybrid interface {
	Calculate(currentPrice decimal.Decimal) (decimal.Decimal, decimal.Decimal, error)
	Trigger(reason string) error
	ReSet(currentPrice decimal.Decimal) error
	GetTakeProfit() (decimal.Decimal, error)
	GetStopLoss() (decimal.Decimal, error)
	Deactivate() error
}

// Debounced
type DebouncedTakeProfit interface {
	TakeProfit
	TakeProfitCondT
}

// Fixed strategy
type FixedTakeProfit interface {
	TakeProfit
	TakeProfitCond
}

// Fixed-ATR
type FixedVolatilityTakeProfit interface {
	FixedTakeProfit
	UpdateATR(currentATR decimal.Decimal) error
}

// Debounced-ATR
type DebouncedVolatilityTakeProfit interface {
	DebouncedTakeProfit
	UpdateATR(currentATR decimal.Decimal) error
}

// Fixed-Moving Average
type FixedMATakeProfit interface {
	FixedTakeProfit
	SetMA(value decimal.Decimal)
}

// Debounced-Moving Average
type DebouncedMATakeProfit interface {
	DebouncedTakeProfit
	SetMA(value decimal.Decimal)
}

// general TakeProfit interface
type TakeProfit interface {
	CalculateTakeProfit(currentPrice decimal.Decimal) (decimal.Decimal, error)
	Trigger(reason string) error
	ReSetTakeProfiter(currentPrice decimal.Decimal) error
	GetTakeProfit() (decimal.Decimal, error)
	Deactivate() error
}

// TakeProfit Condition
type TakeProfitCond interface {
	ShouldTriggerTakeProfit(currentPrice decimal.Decimal) (bool, error)
}

// TakeProfit Condition with timestamp
type TakeProfitCondT interface {
	ShouldTriggerTakeProfit(currentPrice decimal.Decimal, timestamp int64) (bool, error)
	GetTimeThreshold() (int64, error)
}
