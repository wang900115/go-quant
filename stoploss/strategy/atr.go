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
	errTimeThresholdInvalid = errors.New("time threshold must be greater than zero")
	errATRStopLossKInvalid  = errors.New("stop loss k must be greater than 0")
	errATRInvalid           = errors.New("ATR must be greater than 0")
)

type ATR interface {
	FixedATRProfit | FixedATRStop | DebouncedATRProfit | DebouncedATRStop
}

// FixedATRStop represents an ATR-based stop loss strategy
type FixedATRStop struct {
	stoploss.BaseResolver
	threshold  decimal.Decimal
	lastPrice  decimal.Decimal
	multiplier decimal.Decimal
	currentATR decimal.Decimal
}

// FixedATRProfit represents an ATR-based take profit strategy
type FixedATRProfit struct {
	stoploss.BaseResolver
	threshold  decimal.Decimal
	lastPrice  decimal.Decimal
	multiplier decimal.Decimal
	currentATR decimal.Decimal
}

// DebouncedATRStop represents a time-based ATR stop loss strategy
type DebouncedATRStop struct {
	FixedATRStop
	TimeThreshold int64
	TriggerTime   int64
}

// GetTimeThreshold returns the time threshold for Debounced strategies
func (t *DebouncedATRStop) GetTimeThreshold() (int64, error) {
	return t.TimeThreshold, nil
}

// ShouldTriggerStopLoss checks if the Debounced stop loss should be triggered
func (t *DebouncedATRStop) ShouldTriggerStopLoss(currentPrice decimal.Decimal, currentTimestamp int64) (bool, error) {
	if !t.Active {
		return false, stoploss.ErrStatusInvalid
	}
	if currentPrice.LessThanOrEqual(t.threshold) {
		if t.TriggerTime == 0 {
			t.TriggerTime = currentTimestamp
		} else if currentTimestamp-t.TriggerTime >= t.TimeThreshold {
			err := t.Trigger(stoploss.TRIGGERED_REASON_DEBOUNCED_ATR_STOPLOSS)
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

// ReSetStopLosser resets the stop loss based on the current price
func (t *DebouncedATRStop) ReSetStopLosser(currentPrice decimal.Decimal) error {
	if !t.Active {
		return stoploss.ErrStatusInvalid
	}
	t.lastPrice = currentPrice
	t.threshold = currentPrice.Sub(t.currentATR.Mul(t.multiplier))
	t.Active = true
	t.TriggerTime = 0
	return nil
}

// DebouncedATRProfit represents a time-based ATR take profit strategy
type DebouncedATRProfit struct {
	FixedATRProfit
	TimeThreshold int64
	TriggerTime   int64
}

// UpdateATR updates the current ATR value for take profit
func (a *DebouncedATRProfit) UpdateATR(currentATR decimal.Decimal) error {
	if !a.Active {
		return stoploss.ErrStatusInvalid
	}
	a.currentATR = currentATR
	return nil
}

// GetTimeThreshold returns the time threshold for Debounced strategies
func (t *DebouncedATRProfit) GetTimeThreshold() (int64, error) {
	return t.TimeThreshold, nil
}

// ShouldTriggerTakeProfit checks if the Debounced take profit should be triggered
func (t *DebouncedATRProfit) ShouldTriggerTakeProfit(currentPrice decimal.Decimal, currentTimestamp int64) (bool, error) {
	if !t.Active {
		return false, stoploss.ErrStatusInvalid
	}
	if currentPrice.GreaterThanOrEqual(t.threshold) {
		if t.TriggerTime == 0 {
			t.TriggerTime = currentTimestamp
		} else if currentTimestamp-t.TriggerTime >= t.TimeThreshold {
			err := t.Trigger(stoploss.TRIGGERED_REASON_DEBOUNCED_ATR_TAKEPROFIT)
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

// ReSetTakeProfiter resets the take profit based on the current price
func (t *DebouncedATRProfit) ReSetTakeProfiter(currentPrice decimal.Decimal) error {
	if !t.Active {
		return stoploss.ErrStatusInvalid
	}
	t.lastPrice = currentPrice
	t.threshold = currentPrice.Add(t.currentATR.Mul(t.multiplier))
	t.Active = true
	t.TriggerTime = 0
	return nil
}

// NewFixedATRStop creates a FixedVolatilityStopLoss based on ATR
func NewFixedATRStop(entryPrice, atr, k decimal.Decimal, callback stoploss.DefaultCallback) (stoploss.FixedVolatilityStopLoss, error) {
	if k.LessThanOrEqual(decimal.Zero) {
		return nil, errATRStopLossKInvalid
	}
	if atr.LessThanOrEqual(decimal.Zero) {
		return nil, errATRInvalid
	}
	return &FixedATRStop{
		lastPrice:  entryPrice,
		currentATR: atr,
		multiplier: k,
		threshold:  entryPrice.Sub(atr.Mul(k)),
		BaseResolver: stoploss.BaseResolver{
			Active:   true,
			Callback: callback,
		},
	}, nil
}

// NewFixedATRProfit creates a VolatilityTakeProfit based on ATR
func NewFixedATRProfit(entryPrice, atr, k decimal.Decimal, callback stoploss.DefaultCallback) (stoploss.FixedVolatilityTakeProfit, error) {
	if k.LessThanOrEqual(decimal.Zero) {
		return nil, errATRStopLossKInvalid
	}
	if atr.LessThanOrEqual(decimal.Zero) {
		return nil, errATRInvalid
	}
	return &FixedATRProfit{
		lastPrice:  entryPrice,
		currentATR: atr,
		multiplier: k,
		threshold:  entryPrice.Add(atr.Mul(k)),
		BaseResolver: stoploss.BaseResolver{
			Active:   true,
			Callback: callback,
		},
	}, nil
}

// NewDebouncedATRStop creates a DebouncedVolatilityStopLoss based on ATR and time threshold
func NewDebouncedATRStop(entryPrice, atr, k decimal.Decimal, timeThreshold int64, callback stoploss.DefaultCallback) (stoploss.DebouncedVolatilityStopLoss, error) {
	if k.LessThanOrEqual(decimal.Zero) {
		return nil, errATRStopLossKInvalid
	}
	if atr.LessThanOrEqual(decimal.Zero) {
		return nil, errATRInvalid
	}
	if timeThreshold <= 0 {
		return nil, errTimeThresholdInvalid
	}
	return &DebouncedATRStop{
		FixedATRStop: FixedATRStop{
			lastPrice:  entryPrice,
			currentATR: atr,
			multiplier: k,
			threshold:  entryPrice.Sub(atr.Mul(k)),
			BaseResolver: stoploss.BaseResolver{
				Active:   true,
				Callback: callback,
			},
		},
		TimeThreshold: timeThreshold,
		TriggerTime:   0,
	}, nil
}

// NewDebouncedATRProfit creates a DebouncedVolatilityTakeProfit based on ATR and time threshold
func NewDebouncedATRProfit(entryPrice, atr, k decimal.Decimal, timeThreshold int64, callback stoploss.DefaultCallback) (stoploss.DebouncedVolatilityTakeProfit, error) {
	if k.LessThanOrEqual(decimal.Zero) {
		return nil, errATRStopLossKInvalid
	}
	if atr.LessThanOrEqual(decimal.Zero) {
		return nil, errATRInvalid
	}
	if timeThreshold <= 0 {
		return nil, errTimeThresholdInvalid
	}
	return &DebouncedATRProfit{
		FixedATRProfit: FixedATRProfit{
			lastPrice:  entryPrice,
			currentATR: atr,
			multiplier: k,
			threshold:  entryPrice.Add(atr.Mul(k)),
			BaseResolver: stoploss.BaseResolver{
				Active:   true,
				Callback: callback,
			},
		},
		TimeThreshold: timeThreshold,
		TriggerTime:   0,
	}, nil
}

// UpdateATR updates the current ATR value for stop loss
func (a *FixedATRStop) UpdateATR(currentATR decimal.Decimal) error {
	if !a.Active {
		return stoploss.ErrStatusInvalid
	}
	a.currentATR = currentATR
	return nil
}

// UpdateATR updates the current ATR value for take profit
func (a *FixedATRProfit) UpdateATR(currentATR decimal.Decimal) error {
	if !a.Active {
		return stoploss.ErrStatusInvalid
	}
	a.currentATR = currentATR
	return nil
}

// CalculateStopLoss represents first update last price and calculate stop loss then update threshold
func (a *FixedATRStop) CalculateStopLoss(currentPrice decimal.Decimal) (decimal.Decimal, error) {
	if !a.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	a.lastPrice = currentPrice
	a.threshold = currentPrice.Sub(a.currentATR.Mul(a.multiplier))
	return a.threshold, nil
}

// CalculateTakeProfit represents first update last price and calculate take profit then update threshold
func (a *FixedATRProfit) CalculateTakeProfit(currentPrice decimal.Decimal) (decimal.Decimal, error) {
	if !a.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	a.lastPrice = currentPrice
	a.threshold = currentPrice.Add(a.currentATR.Mul(a.multiplier))
	return a.threshold, nil
}

// ShouldTriggerStopLoss checks if the stop loss should be triggered
func (a *FixedATRStop) ShouldTriggerStopLoss(currentPrice decimal.Decimal) (bool, error) {
	if !a.Active {
		return false, stoploss.ErrStatusInvalid
	}
	if currentPrice.LessThanOrEqual(a.threshold) {
		err := a.Trigger(stoploss.TRIGGERED_REASON_FIXED_ATR_STOPLOSS)
		if err != nil {
			return true, stoploss.ErrCallBackFail
		}
		return true, nil
	}
	return false, nil
}

// ShouldTriggerTakeProfit checks if the take profit should be triggered
func (a *FixedATRProfit) ShouldTriggerTakeProfit(currentPrice decimal.Decimal) (bool, error) {
	if !a.Active {
		return false, stoploss.ErrStatusInvalid
	}
	if currentPrice.GreaterThanOrEqual(a.threshold) {
		err := a.Trigger(stoploss.TRIGGERED_REASON_FIXED_ATR_TAKEPROFIT)
		if err != nil {
			return true, stoploss.ErrCallBackFail
		}
		return true, nil
	}
	return false, nil
}

// GetStopLoss returns the current stop loss threshold
func (a *FixedATRStop) GetStopLoss() (decimal.Decimal, error) {
	if !a.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	return a.threshold, nil
}

// GetTakeProfit returns the current take profit threshold
func (a *FixedATRProfit) GetTakeProfit() (decimal.Decimal, error) {
	if !a.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	return a.threshold, nil
}

// ReSetStopLosser resets the stop loss based on the current price
func (a *FixedATRStop) ReSetStopLosser(currentPrice decimal.Decimal) error {
	if !a.Active {
		return stoploss.ErrStatusInvalid
	}
	a.lastPrice = currentPrice
	a.threshold = currentPrice.Sub(a.currentATR.Mul(a.multiplier))
	a.Active = true
	return nil
}

// ReSetTakeProfiter resets the take profit based on the current price
func (a *FixedATRProfit) ReSetTakeProfiter(currentPrice decimal.Decimal) error {
	if !a.Active {
		return stoploss.ErrStatusInvalid
	}
	a.lastPrice = currentPrice
	a.threshold = currentPrice.Add(a.currentATR.Mul(a.multiplier))
	a.Active = true
	return nil
}
