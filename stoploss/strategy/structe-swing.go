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
	errSwingLookbackInvalid = errors.New("swing lookback period must be greater than 0")
	errSwingDistanceInvalid = errors.New("swing distance must be greater than 0")
)

type StructureSwing struct {
	stoploss.BaseResolver
	lastPrice      decimal.Decimal
	lookbackPeriod int
	swingDistance  decimal.Decimal
	stopLoss       decimal.Decimal
	takeProfit     decimal.Decimal
	priceHistory   []decimal.Decimal
	lastSwingLow   decimal.Decimal
	lastSwingHigh  decimal.Decimal
	stopPct        decimal.Decimal
	profitPct      decimal.Decimal
	isLongPosition bool
}

// NewStructureSwingStop creates a new structure swing stop loss strategy
// lastPrice: the entry price of the position
// lookbackPeriod: number of periods to look back for swing identification
// swingDistance: minimum distance between swing points as a percentage
// isLong: true for long positions, false for short positions
func NewStructureSwingStop(lookbackPeriod int, lastPrice, swingDistance, stopPct, profitPct decimal.Decimal, isLong bool, callback stoploss.DefaultCallback) (stoploss.HybridWithoutTime, error) {
	if lookbackPeriod <= 0 {
		return nil, errSwingLookbackInvalid
	}
	if swingDistance.LessThanOrEqual(decimal.Zero) {
		return nil, errSwingDistanceInvalid
	}

	initialStopLoss := lastPrice
	initialTakeProfit := lastPrice

	if isLong {
		// For long positions, initial stop loss below entry, take profit above
		initialStopLoss = lastPrice.Mul(stopPct)
		initialTakeProfit = lastPrice.Mul(profitPct)
	} else {
		// For short positions, initial stop loss above entry, take profit below
		initialStopLoss = lastPrice.Mul(stopPct)
		initialTakeProfit = lastPrice.Mul(profitPct)
	}

	return &StructureSwing{
		lastPrice:      lastPrice,
		lookbackPeriod: lookbackPeriod,
		swingDistance:  swingDistance,
		stopLoss:       initialStopLoss,
		takeProfit:     initialTakeProfit,
		priceHistory:   make([]decimal.Decimal, 0, lookbackPeriod*2),
		lastSwingLow:   lastPrice,
		lastSwingHigh:  lastPrice,
		isLongPosition: isLong,
		stopPct:        stopPct,
		profitPct:      profitPct,
		BaseResolver: stoploss.BaseResolver{
			Active:   true,
			Callback: callback,
		},
	}, nil
}

// UpdatePrice adds a new price to the history and updates swing levels
func (ss *StructureSwing) UpdatePrice(newPrice decimal.Decimal) {
	if !ss.Active {
		return
	}

	// Add to price history
	ss.priceHistory = append(ss.priceHistory, newPrice)

	// Keep only the required lookback period + buffer
	maxHistory := ss.lookbackPeriod * 2
	if len(ss.priceHistory) > maxHistory {
		ss.priceHistory = ss.priceHistory[len(ss.priceHistory)-maxHistory:]
	}

	// Update swing levels if we have enough data
	if len(ss.priceHistory) >= ss.lookbackPeriod {
		ss.updateSwingLevels()
		ss.updateStopLossAndTakeProfit()
	}
}

// updateSwingLevels identifies new swing highs and lows
func (ss *StructureSwing) updateSwingLevels() {
	if len(ss.priceHistory) < ss.lookbackPeriod {
		return
	}

	currentIdx := len(ss.priceHistory) - 1
	lookbackStart := currentIdx - ss.lookbackPeriod + 1
	if lookbackStart < 0 {
		lookbackStart = 0
	}

	// Find swing high and low in the lookback period
	swingHigh := ss.priceHistory[lookbackStart]
	swingLow := ss.priceHistory[lookbackStart]

	for i := lookbackStart; i <= currentIdx; i++ {
		if ss.priceHistory[i].GreaterThan(swingHigh) {
			swingHigh = ss.priceHistory[i]
		}
		if ss.priceHistory[i].LessThan(swingLow) {
			swingLow = ss.priceHistory[i]
		}
	}

	// Update swing levels if they've moved significantly
	swingHighDiff := swingHigh.Sub(ss.lastSwingHigh).Abs()
	swingLowDiff := swingLow.Sub(ss.lastSwingLow).Abs()

	if swingHighDiff.GreaterThan(ss.lastSwingHigh.Mul(ss.swingDistance)) {
		ss.lastSwingHigh = swingHigh
	}

	if swingLowDiff.GreaterThan(ss.lastSwingLow.Mul(ss.swingDistance)) {
		ss.lastSwingLow = swingLow
	}
}

// updateStopLossAndTakeProfit adjusts stop loss and take profit based on swing levels
func (ss *StructureSwing) updateStopLossAndTakeProfit() {
	if ss.isLongPosition {
		// For long positions: stop loss below swing low, take profit above swing high
		newStopLoss := ss.lastSwingLow.Mul(decimal.NewFromFloat(0.99)) // slightly below swing low
		if newStopLoss.GreaterThan(ss.stopLoss) {
			ss.stopLoss = newStopLoss // Only move stop loss up (trailing)
		}

		// Take profit above swing high
		ss.takeProfit = ss.lastSwingHigh.Mul(decimal.NewFromFloat(1.01))
	} else {
		// For short positions: stop loss above swing high, take profit below swing low
		newStopLoss := ss.lastSwingHigh.Mul(decimal.NewFromFloat(1.01)) // slightly above swing high
		if newStopLoss.LessThan(ss.stopLoss) {
			ss.stopLoss = newStopLoss // Only move stop loss down (trailing)
		}

		// Take profit below swing low
		ss.takeProfit = ss.lastSwingLow.Mul(decimal.NewFromFloat(0.99))
	}
}

// Calculate represents update last price and recalculate stop loss and take profit based on structure swings
func (ss *StructureSwing) Calculate(currentPrice decimal.Decimal) (decimal.Decimal, decimal.Decimal, error) {
	if !ss.Active {
		return decimal.Zero, decimal.Zero, stoploss.ErrStatusInvalid
	}

	// Update price history with current price
	ss.UpdatePrice(currentPrice)

	return ss.stopLoss, ss.takeProfit, nil
}

func (ss *StructureSwing) ShouldTriggerStopLoss(currentPrice decimal.Decimal) (bool, error) {
	if !ss.Active {
		return false, stoploss.ErrStatusInvalid
	}

	// Update price before checking
	ss.UpdatePrice(currentPrice)

	var triggered bool
	if ss.isLongPosition {
		// For long positions, stop loss triggers when price falls below stop loss
		triggered = currentPrice.LessThanOrEqual(ss.stopLoss)
	} else {
		// For short positions, stop loss triggers when price rises above stop loss
		triggered = currentPrice.GreaterThanOrEqual(ss.stopLoss)
	}

	if triggered {
		err := ss.Trigger(stoploss.TRIGGERED_REASON_STRUCTURE_SWING_STOPLOSS)
		if err != nil {
			return true, stoploss.ErrCallBackFail
		}
		return true, nil
	}
	return false, nil
}

func (ss *StructureSwing) ShouldTriggerTakeProfit(currentPrice decimal.Decimal) (bool, error) {
	if !ss.Active {
		return false, stoploss.ErrStatusInvalid
	}

	// Update price before checking
	ss.UpdatePrice(currentPrice)

	var triggered bool
	if ss.isLongPosition {
		// For long positions, take profit triggers when price rises above take profit
		triggered = currentPrice.GreaterThanOrEqual(ss.takeProfit)
	} else {
		// For short positions, take profit triggers when price falls below take profit
		triggered = currentPrice.LessThanOrEqual(ss.takeProfit)
	}

	if triggered {
		err := ss.Trigger(stoploss.TRIGGERED_REASON_STRUCTURE_SWING_TAKEPROFIT)
		if err != nil {
			return true, stoploss.ErrCallBackFail
		}
		return true, nil
	}
	return false, nil
}

func (ss *StructureSwing) GetStopLoss() (decimal.Decimal, error) {
	if !ss.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	return ss.stopLoss, nil
}

func (ss *StructureSwing) GetTakeProfit() (decimal.Decimal, error) {
	if !ss.Active {
		return decimal.Zero, stoploss.ErrStatusInvalid
	}
	return ss.takeProfit, nil
}

func (ss *StructureSwing) ReSet(currentPrice decimal.Decimal) error {
	if !ss.Active {
		return stoploss.ErrStatusInvalid
	}

	ss.lastPrice = currentPrice
	ss.lastSwingLow = currentPrice
	ss.lastSwingHigh = currentPrice
	ss.priceHistory = make([]decimal.Decimal, 0, ss.lookbackPeriod*2)

	if ss.isLongPosition {
		ss.stopLoss = currentPrice.Mul(ss.stopPct)
		ss.takeProfit = currentPrice.Mul(ss.profitPct)
	} else {
		ss.stopLoss = currentPrice.Mul(ss.stopPct)
		ss.takeProfit = currentPrice.Mul(ss.profitPct)
	}

	ss.Active = true
	return nil
}
