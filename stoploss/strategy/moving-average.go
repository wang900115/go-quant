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
	errMAInvalid     = errors.New("moving average value must be greater than 0")
	errOffsetInvalid = errors.New("offset percentage must be between 0 and 1")
)

// FixedMovingAverageStop represents a moving average based stop loss strategy
type FixedMovingAverageStop struct {
	stoploss.BaseResolver
	threshold     decimal.Decimal
	lastPrice     decimal.Decimal
	movingAverage decimal.Decimal
	offsetPercent decimal.Decimal
}

// FixedMovingAverageProfit represents a moving average based take profit strategy
type FixedMovingAverageProfit struct {
	stoploss.BaseResolver
	threshold     decimal.Decimal
	lastPrice     decimal.Decimal
	movingAverage decimal.Decimal
	offsetPercent decimal.Decimal
}

// DebouncedMovingAverageStop represents a Debounced moving average based stop loss strategy
type DebouncedMovingAverageStop struct {
	FixedMovingAverageStop
	TriggerTime   int64
	TimeThreshold int64
}

// DebouncedMovingAverageProfit represents a Debounced moving average based take profit strategy
type DebouncedMovingAverageProfit struct {
	FixedMovingAverageProfit
	TriggerTime   int64
	TimeThreshold int64
}

// NewFixedMovingAverageStop creates a FixedMAStopLoss based on Moving Average
func NewFixedMovingAverageStop(entryPrice, initialMA, offsetPercent decimal.Decimal, callback stoploss.DefaultCallback) (stoploss.FixedMAStopLoss, error) {
	if initialMA.LessThanOrEqual(decimal.Zero) {
		return nil, errMAInvalid
	}
	if offsetPercent.IsNegative() || offsetPercent.GreaterThan(decimal.NewFromInt(1)) {
		return nil, errOffsetInvalid
	}
	return &FixedMovingAverageStop{
		lastPrice:     entryPrice,
		movingAverage: initialMA,
		offsetPercent: offsetPercent,
		threshold:     initialMA.Mul(decimal.NewFromInt(1).Sub(offsetPercent)),
		BaseResolver: stoploss.BaseResolver{
			Active:   true,
			Callback: callback,
		},
	}, nil
}

// NewFixedMovingAverageProfit creates a FixedMATakeProfit based on Moving Average
func NewFixedMovingAverageProfit(entryPrice, initialMA, offsetPercent decimal.Decimal, callback stoploss.DefaultCallback) (stoploss.FixedMATakeProfit, error) {
	if initialMA.LessThanOrEqual(decimal.Zero) {
		return nil, errMAInvalid
	}
	if offsetPercent.IsNegative() || offsetPercent.GreaterThan(decimal.NewFromInt(1)) {
		return nil, errOffsetInvalid
	}
	return &FixedMovingAverageProfit{
		lastPrice:     entryPrice,
		movingAverage: initialMA,
		offsetPercent: offsetPercent,
		threshold:     initialMA.Mul(decimal.NewFromInt(1).Add(offsetPercent)),
		BaseResolver: stoploss.BaseResolver{
			Active:   true,
			Callback: callback,
		},
	}, nil
}

func NewDebouncedMovingAverageStop(entryPrice, initialMA, offsetPercent decimal.Decimal, timeThreshold int64, callback stoploss.DefaultCallback) (stoploss.DebouncedMAStopLoss, error) {
	if initialMA.LessThanOrEqual(decimal.Zero) {
		return nil, errMAInvalid
	}
	if offsetPercent.IsNegative() || offsetPercent.GreaterThan(decimal.NewFromInt(1)) {
		return nil, errOffsetInvalid
	}
	return &DebouncedMovingAverageStop{
		FixedMovingAverageStop: FixedMovingAverageStop{
			lastPrice:     entryPrice,
			movingAverage: initialMA,
			offsetPercent: offsetPercent,
			threshold:     initialMA.Mul(decimal.NewFromInt(1).Sub(offsetPercent)),
			BaseResolver: stoploss.BaseResolver{
				Active:   true,
				Callback: callback,
			},
		},
		TimeThreshold: timeThreshold,
	}, nil
}

func NewDebouncedMovingAverageProfit(entryPrice, initialMA, offsetPercent decimal.Decimal, timeThreshold int64, callback stoploss.DefaultCallback) (stoploss.DebouncedMATakeProfit, error) {
	if initialMA.LessThanOrEqual(decimal.Zero) {
		return nil, errMAInvalid
	}
	if offsetPercent.IsNegative() || offsetPercent.GreaterThan(decimal.NewFromInt(1)) {
		return nil, errOffsetInvalid
	}
	return &DebouncedMovingAverageProfit{
		FixedMovingAverageProfit: FixedMovingAverageProfit{
			lastPrice:     entryPrice,
			movingAverage: initialMA,
			offsetPercent: offsetPercent,
			threshold:     initialMA.Mul(decimal.NewFromInt(1).Add(offsetPercent)),
			BaseResolver: stoploss.BaseResolver{
				Active:   true,
				Callback: callback,
			},
		},
		TimeThreshold: timeThreshold,
	}, nil
}

// SetMA sets the current moving average value for stop loss
func (ma *FixedMovingAverageStop) SetMA(value decimal.Decimal) {
	if !ma.Active {
		return
	}
	ma.movingAverage = value
}

// SetMA sets the current moving average value for take profit
func (ma *FixedMovingAverageProfit) SetMA(value decimal.Decimal) {
	if !ma.Active {
		return
	}
	ma.movingAverage = value
}

// SetMA sets the current moving average value for Debounced stop loss
func (ma *DebouncedMovingAverageStop) SetMA(value decimal.Decimal) {
	if !ma.Active {
		return
	}
	ma.movingAverage = value
}

// SetMA sets the current moving average value for Debounced take profit
func (ma *DebouncedMovingAverageProfit) SetMA(value decimal.Decimal) {
	if !ma.Active {
		return
	}
	ma.movingAverage = value
}

// CalculateStopLoss represents first update last price and calculate stop loss then update threshold
func (ma *FixedMovingAverageStop) CalculateStopLoss(currentPrice decimal.Decimal) (decimal.Decimal, error) {
	if !ma.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	ma.lastPrice = currentPrice
	ma.threshold = ma.movingAverage.Mul(decimal.NewFromInt(1).Sub(ma.offsetPercent))
	return ma.threshold, nil
}

// CalculateTakeProfit represents first update last price and calculate take profit then update threshold
func (ma *FixedMovingAverageProfit) CalculateTakeProfit(currentPrice decimal.Decimal) (decimal.Decimal, error) {
	if !ma.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	ma.lastPrice = currentPrice
	ma.threshold = ma.movingAverage.Mul(decimal.NewFromInt(1).Add(ma.offsetPercent))
	return ma.threshold, nil
}

// CalculateStopLoss represents first update last price and calculate stop loss then update threshold
func (ma *DebouncedMovingAverageStop) CalculateStopLoss(currentPrice decimal.Decimal) (decimal.Decimal, error) {
	if !ma.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	ma.lastPrice = currentPrice
	ma.threshold = ma.movingAverage.Mul(decimal.NewFromInt(1).Sub(ma.offsetPercent))
	return ma.threshold, nil
}

// CalculateTakeProfit represents first update last price and calculate take profit then update threshold
func (ma *DebouncedMovingAverageProfit) CalculateTakeProfit(currentPrice decimal.Decimal) (decimal.Decimal, error) {
	if !ma.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	ma.lastPrice = currentPrice
	ma.threshold = ma.movingAverage.Mul(decimal.NewFromInt(1).Add(ma.offsetPercent))
	return ma.threshold, nil
}

// ShouldTriggerStopLoss checks if the stop loss should be triggered
func (ma *FixedMovingAverageStop) ShouldTriggerStopLoss(currentPrice decimal.Decimal) (bool, error) {
	if !ma.Active {
		return false, stoploss.ErrStatusInvalid
	}

	if currentPrice.LessThanOrEqual(ma.threshold) {
		err := ma.Trigger(stoploss.TRIGGERED_REASON_FIXED_MA_STOPLOSS)
		if err != nil {
			return true, stoploss.ErrCallBackFail
		}
		return true, nil
	}
	return false, nil
}

// ShouldTriggerTakeProfit checks if the take profit should be triggered
func (ma *FixedMovingAverageProfit) ShouldTriggerTakeProfit(currentPrice decimal.Decimal) (bool, error) {
	if !ma.Active {
		return false, stoploss.ErrStatusInvalid
	}
	if currentPrice.GreaterThanOrEqual(ma.threshold) {
		err := ma.Trigger(stoploss.TRIGGERED_REASON_FIXED_MA_TAKEPROFIT)
		if err != nil {
			return true, stoploss.ErrCallBackFail
		}
		return true, nil
	}
	return false, nil
}

// ShouldTriggerStopLoss checks if the stop loss should be triggered
func (ma *DebouncedMovingAverageStop) ShouldTriggerStopLoss(currentPrice decimal.Decimal, currentTime int64) (bool, error) {
	if !ma.Active {
		return false, stoploss.ErrStatusInvalid
	}

	if currentPrice.LessThanOrEqual(ma.threshold) {
		if ma.TriggerTime == 0 {
			ma.TriggerTime = currentTime
		}
		if currentTime-ma.TriggerTime >= ma.TimeThreshold {
			err := ma.Trigger(stoploss.TRIGGERED_REASON_DEBOUNCED_MA_STOPLOSS)
			if err != nil {
				return true, stoploss.ErrCallBackFail
			}
			return true, nil
		}
	} else {
		ma.TriggerTime = 0
	}
	return false, nil
}

// ShouldTriggerTakeProfit checks if the take profit should be triggered
func (ma *DebouncedMovingAverageProfit) ShouldTriggerTakeProfit(currentPrice decimal.Decimal, currentTime int64) (bool, error) {
	if !ma.Active {
		return false, stoploss.ErrStatusInvalid
	}
	if currentPrice.GreaterThanOrEqual(ma.threshold) {
		if ma.TriggerTime == 0 {
			ma.TriggerTime = currentTime
		}
		if currentTime-ma.TriggerTime >= ma.TimeThreshold {
			err := ma.Trigger(stoploss.TRIGGERED_REASON_DEBOUNCED_MA_TAKEPROFIT)
			if err != nil {
				return true, stoploss.ErrCallBackFail
			}
			return true, nil
		}
	} else {
		ma.TriggerTime = 0
	}
	return false, nil
}

// GetTimeThreshold returns the time threshold for Debounced strategies
func (ma *DebouncedMovingAverageStop) GetTimeThreshold() (int64, error) {
	if !ma.Active {
		return 0, stoploss.ErrStatusInvalid
	}
	return ma.TimeThreshold, nil
}

// GetTimeThreshold returns the time threshold for Debounced strategies
func (ma *DebouncedMovingAverageProfit) GetTimeThreshold() (int64, error) {
	if !ma.Active {
		return 0, stoploss.ErrStatusInvalid
	}
	return ma.TimeThreshold, nil
}

// GetStopLoss returns the current stop loss threshold
func (ma *FixedMovingAverageStop) GetStopLoss() (decimal.Decimal, error) {
	if !ma.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	return ma.threshold, nil
}

// GetTakeProfit returns the current take profit threshold
func (ma *FixedMovingAverageProfit) GetTakeProfit() (decimal.Decimal, error) {
	if !ma.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	return ma.threshold, nil
}

// ReSetStopLosser resets the stop loss based on the current price
func (ma *FixedMovingAverageStop) ReSetStopLosser(currentPrice decimal.Decimal) error {
	if !ma.Active {
		return stoploss.ErrStatusInvalid
	}

	ma.lastPrice = currentPrice
	ma.threshold = ma.movingAverage.Mul(decimal.NewFromInt(1).Sub(ma.offsetPercent))
	ma.Active = true
	return nil
}

// ReSetTakeProfiter resets the take profit based on the current price
func (ma *FixedMovingAverageProfit) ReSetTakeProfiter(currentPrice decimal.Decimal) error {
	if !ma.Active {
		return stoploss.ErrStatusInvalid
	}
	ma.lastPrice = currentPrice
	ma.threshold = ma.movingAverage.Mul(decimal.NewFromInt(1).Add(ma.offsetPercent))
	ma.Active = true
	return nil
}

// ReSetStopLosser resets the stop loss based on the current price
func (ma *DebouncedMovingAverageStop) ReSetStopLosser(currentPrice decimal.Decimal) error {
	if !ma.Active {
		return stoploss.ErrStatusInvalid
	}

	ma.lastPrice = currentPrice
	ma.threshold = ma.movingAverage.Mul(decimal.NewFromInt(1).Sub(ma.offsetPercent))
	ma.Active = true
	ma.TriggerTime = 0
	return nil
}

// ReSetTakeProfiter resets the take profit based on the current price
func (ma *DebouncedMovingAverageProfit) ReSetTakeProfiter(currentPrice decimal.Decimal) error {
	if !ma.Active {
		return stoploss.ErrStatusInvalid
	}
	ma.lastPrice = currentPrice
	ma.threshold = ma.movingAverage.Mul(decimal.NewFromInt(1).Add(ma.offsetPercent))
	ma.Active = true
	ma.TriggerTime = 0
	return nil
}
