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
	"errors"

	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/stoploss"
)

var (
	errStopLossRateInvalid   = errors.New("stop loss rate must be between 0 and 1")
	errTakeProfitRateInvalid = errors.New("take profit rate must be between 0 and 1")
)

// FixedPercentStop represents a stop loss strategy based on fixed percentage
type FixedPercentStop struct {
	stoploss.BaseResolver
	threshold    decimal.Decimal
	tolerancePct decimal.Decimal
	LastPrice    decimal.Decimal
}

// FixedPercentProfit represents a take profit strategy based on fixed percentage
type FixedPercentProfit struct {
	stoploss.BaseResolver
	threshold    decimal.Decimal
	tolerancePct decimal.Decimal
	LastPrice    decimal.Decimal
}

// DebouncedPercentStop represents a Debounced stop loss strategy based on fixed percentage
type DebouncedPercentStop struct {
	FixedPercentStop
	TriggerTime   int64
	TimeThreshold int64
}

// DebouncedPercentProfit represents a Debounced take profit strategy based on fixed percentage
type DebouncedPercentProfit struct {
	FixedPercentProfit
	TriggerTime   int64
	TimeThreshold int64
}

// NewFixedPercentStop creates a FixedPercentStopLoss base on fixed percentage
func NewFixedPercentStop(entryPrice, stopLossPct decimal.Decimal, callback stoploss.DefaultCallback) (stoploss.FixedStopLoss, error) {
	if stopLossPct.IsNegative() || stopLossPct.GreaterThan(decimal.NewFromInt(1)) {
		return nil, errStopLossRateInvalid
	}
	s := &FixedPercentStop{
		LastPrice:    entryPrice,
		threshold:    entryPrice.Mul(decimal.NewFromInt(1).Sub(stopLossPct)),
		tolerancePct: stopLossPct,
		BaseResolver: stoploss.BaseResolver{
			Active:   true,
			Callback: callback,
		},
	}
	return s, nil
}

// NewFixedPercentProfit creates a FixedPercentTakeProfit based on fixed percentage
func NewFixedPercentProfit(entryPrice, takeProfitPct decimal.Decimal, callback stoploss.DefaultCallback) (stoploss.FixedTakeProfit, error) {
	if takeProfitPct.IsNegative() || takeProfitPct.GreaterThan(decimal.NewFromInt(1)) {
		return nil, errTakeProfitRateInvalid
	}
	s := &FixedPercentProfit{
		LastPrice:    entryPrice,
		threshold:    entryPrice.Mul(decimal.NewFromInt(1).Add(takeProfitPct)),
		tolerancePct: takeProfitPct,
		BaseResolver: stoploss.BaseResolver{
			Active:   true,
			Callback: callback,
		},
	}
	return s, nil
}

func NewDebouncedPercentStop(entryPrice, stopLossPct decimal.Decimal, timeThreshold int64, callback stoploss.DefaultCallback) (stoploss.DebouncedStopLoss, error) {
	if stopLossPct.IsNegative() || stopLossPct.GreaterThan(decimal.NewFromInt(1)) {
		return nil, errStopLossRateInvalid
	}
	s := &DebouncedPercentStop{
		FixedPercentStop: FixedPercentStop{
			LastPrice:    entryPrice,
			threshold:    entryPrice.Mul(decimal.NewFromInt(1).Sub(stopLossPct)),
			tolerancePct: stopLossPct,
			BaseResolver: stoploss.BaseResolver{
				Active:   true,
				Callback: callback,
			},
		},
		TimeThreshold: timeThreshold,
	}
	return s, nil
}

func NewDebouncedPercentProfit(entryPrice, takeProfitPct decimal.Decimal, timeThreshold int64, callback stoploss.DefaultCallback) (stoploss.DebouncedTakeProfit, error) {
	if takeProfitPct.IsNegative() || takeProfitPct.GreaterThan(decimal.NewFromInt(1)) {
		return nil, errTakeProfitRateInvalid
	}
	s := &DebouncedPercentProfit{
		FixedPercentProfit: FixedPercentProfit{
			LastPrice:    entryPrice,
			threshold:    entryPrice.Mul(decimal.NewFromInt(1).Add(takeProfitPct)),
			tolerancePct: takeProfitPct,
			BaseResolver: stoploss.BaseResolver{
				Active:   true,
				Callback: callback,
			},
		},
		TimeThreshold: timeThreshold,
	}
	return s, nil
}

func (t *DebouncedPercentStop) GetTimeThreshold() (int64, error) {
	if !t.Active {
		return 0, stoploss.ErrStatusInvalid
	}
	return t.TimeThreshold, nil
}

func (t *DebouncedPercentProfit) GetTimeThreshold() (int64, error) {
	if !t.Active {
		return 0, stoploss.ErrStatusInvalid
	}
	return t.TimeThreshold, nil
}

func (t *DebouncedPercentStop) ShouldTriggerStopLoss(currentPrice decimal.Decimal, currentTime int64) (bool, error) {
	if !t.Active {
		return false, stoploss.ErrStatusInvalid
	}
	if currentPrice.LessThanOrEqual(t.threshold) {
		if t.TriggerTime == 0 {
			t.TriggerTime = currentTime
		}
		if currentTime-t.TriggerTime >= t.TimeThreshold {
			err := t.Trigger(stoploss.TRIGGERED_REASON_DEBOUNCED_PERCENTCILE_STOPLOSS)
			if err != nil {
				return true, stoploss.ErrCallBackFail
			}
			return true, nil
		}
	} else {
		t.TriggerTime = 0
	}
	return false, nil
}

func (t *DebouncedPercentProfit) ShouldTriggerTakeProfit(currentPrice decimal.Decimal, currentTime int64) (bool, error) {
	if !t.Active {
		return false, stoploss.ErrStatusInvalid
	}
	if currentPrice.GreaterThanOrEqual(t.threshold) {
		if t.TriggerTime == 0 {
			t.TriggerTime = currentTime
		}
		if currentTime-t.TriggerTime >= t.TimeThreshold {
			err := t.Trigger(stoploss.TRIGGERED_REASON_DEBOUNCED_PERCENTCILE_TAKEPROFIT)
			if err != nil {
				return true, stoploss.ErrCallBackFail
			}
			return true, nil
		}
	} else {
		t.TriggerTime = 0
	}
	return false, nil
}

// CalculateStopLoss represents first update last price and calculate stop loss then update threshold
func (f *FixedPercentStop) CalculateStopLoss(currentPrice decimal.Decimal) (decimal.Decimal, error) {
	if !f.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	f.LastPrice = currentPrice
	f.threshold = currentPrice.Mul(decimal.NewFromInt(1).Sub(f.tolerancePct))
	return f.threshold, nil
}

// ShouldTriggerStopLoss checks if the stop loss should be triggered
func (f *FixedPercentStop) ShouldTriggerStopLoss(currentPrice decimal.Decimal) (bool, error) {
	if !f.Active {
		return false, stoploss.ErrStatusInvalid
	}
	if currentPrice.LessThanOrEqual(f.threshold) {
		err := f.Trigger(stoploss.TRIGGERED_REASON_FIXED_PERCENTCILE_STOPLOSS)
		if err != nil {
			return true, stoploss.ErrCallBackFail
		}
		return true, nil
	}
	return false, nil
}

// GetStopLoss returns the current stop loss threshold
func (f *FixedPercentStop) GetStopLoss() (decimal.Decimal, error) {
	if !f.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	return f.threshold, nil
}

// ReSetStopLosser resets the stop loss based on the current price
func (f *FixedPercentStop) ReSetStopLosser(currentPrice decimal.Decimal) error {
	if !f.Active {
		return stoploss.ErrStatusInvalid
	}
	f.LastPrice = currentPrice
	f.threshold = currentPrice.Mul(decimal.NewFromInt(1).Sub(f.tolerancePct))
	f.Active = true
	return nil
}

// CalculateTakeProfit represents first update last price and calculate take profit then update threshold
func (f *FixedPercentProfit) CalculateTakeProfit(currentPrice decimal.Decimal) (decimal.Decimal, error) {
	if !f.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	f.LastPrice = currentPrice
	f.threshold = currentPrice.Mul(decimal.NewFromInt(1).Add(f.tolerancePct))
	return f.threshold, nil
}

// ShouldTriggerTakeProfit checks if the take profit should be triggered
func (f *FixedPercentProfit) ShouldTriggerTakeProfit(currentPrice decimal.Decimal) (bool, error) {
	if !f.Active {
		return false, stoploss.ErrStatusInvalid
	}
	if currentPrice.GreaterThanOrEqual(f.threshold) {
		err := f.Trigger(stoploss.TRIGGERED_REASON_FIXED_PERCENTCILE_TAKEPROFIT)
		if err != nil {
			return true, stoploss.ErrCallBackFail
		}
		return true, nil
	}
	return false, nil
}

// GetTakeProfit returns the current take profit threshold
func (f *FixedPercentProfit) GetTakeProfit() (decimal.Decimal, error) {
	if !f.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	return f.threshold, nil
}

// ReSetTakeProfiter resets the take profit based on the current price
func (f *FixedPercentProfit) ReSetTakeProfiter(currentPrice decimal.Decimal) error {
	if !f.Active {
		return stoploss.ErrStatusInvalid
	}
	f.LastPrice = currentPrice
	f.threshold = currentPrice.Mul(decimal.NewFromInt(1).Add(f.tolerancePct))
	f.Active = true
	return nil
}

func (t *DebouncedPercentStop) ReSetStopLosser(currentPrice decimal.Decimal) error {
	if !t.Active {
		return stoploss.ErrStatusInvalid
	}
	t.LastPrice = currentPrice
	t.threshold = currentPrice.Mul(decimal.NewFromInt(1).Sub(t.tolerancePct))
	t.TriggerTime = 0
	t.Active = true
	return nil
}

func (t *DebouncedPercentProfit) ReSetTakeProfiter(currentPrice decimal.Decimal) error {
	if !t.Active {
		return stoploss.ErrStatusInvalid
	}
	t.LastPrice = currentPrice
	t.threshold = currentPrice.Mul(decimal.NewFromInt(1).Add(t.tolerancePct))
	t.TriggerTime = 0
	t.Active = true
	return nil
}
