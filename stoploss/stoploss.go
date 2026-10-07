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

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/model"
)

var (
	ErrStatusInvalid = errors.New("stop loss status is invalid")
	ErrCallBackFail  = errors.New("stop loss callback function failed")
)

// Time-based
type DebouncedStopLoss interface {
	StopLoss
	StopLossCondT
}

// Fixed strategy
type FixedStopLoss interface {
	StopLoss
	StopLossCond
}

// Fixed-ATR
type FixedVolatilityStopLoss interface {
	FixedStopLoss
	UpdateATR(currentATR decimal.Decimal) error
}

// Debounced-ATR
type DebouncedVolatilityStopLoss interface {
	DebouncedStopLoss
	UpdateATR(currentATR decimal.Decimal) error
}

// Fixed-Moving Average
type FixedMAStopLoss interface {
	FixedStopLoss
	SetMA(value decimal.Decimal)
}

// Debounced-Moving Average
type DebouncedMAStopLoss interface {
	DebouncedStopLoss
	SetMA(value decimal.Decimal)
}

// general StopLoss interface
type StopLoss interface {
	CalculateStopLoss(currentPrice decimal.Decimal) (decimal.Decimal, error)
	Trigger(evt TriggerEvent) error
	GetStopLoss() (decimal.Decimal, error)
	ReSetStopLosser(currentPrice decimal.Decimal) error
	Deactivate() error
}

// StopLoss Condition with timestamp
type StopLossCondT interface {
	ShouldTriggerStopLoss(currentPrice decimal.Decimal, timestamp int64) (bool, error)
	GetTimeThreshold() (int64, error)
}

// StopLoss Condition
type StopLossCond interface {
	ShouldTriggerStopLoss(currentPrice decimal.Decimal) (bool, error)
}

// TriggerEvent carries everything a callback needs to know about a
// stop-loss / take-profit hit.
type TriggerEvent struct {
	// Category is model.STOP_LOSS or model.TAKE_PROFIT.
	Category model.StrategyCategory
	// Reason is one of the TRIGGERED_REASON_* constants.
	Reason string
	// HitPrice is the threshold that was crossed.
	HitPrice decimal.Decimal
	// CurrentPrice is the market price that crossed the threshold.
	CurrentPrice decimal.Decimal
	// Timestamp is the trigger time in Unix seconds.
	Timestamp int64
}

// Time returns the event timestamp as time.Time (falls back to now when unset).
func (e TriggerEvent) Time() time.Time {
	if e.Timestamp <= 0 {
		return time.Now()
	}
	return time.Unix(e.Timestamp, 0)
}

// NewStopLossEvent builds a STOP_LOSS TriggerEvent. ts <= 0 means "now".
func NewStopLossEvent(reason string, hit, current decimal.Decimal, ts int64) TriggerEvent {
	return newEvent(model.STOP_LOSS, reason, hit, current, ts)
}

// NewTakeProfitEvent builds a TAKE_PROFIT TriggerEvent. ts <= 0 means "now".
func NewTakeProfitEvent(reason string, hit, current decimal.Decimal, ts int64) TriggerEvent {
	return newEvent(model.TAKE_PROFIT, reason, hit, current, ts)
}

func newEvent(c model.StrategyCategory, reason string, hit, current decimal.Decimal, ts int64) TriggerEvent {
	if ts <= 0 {
		ts = time.Now().Unix()
	}
	return TriggerEvent{Category: c, Reason: reason, HitPrice: hit, CurrentPrice: current, Timestamp: ts}
}

// DefaultCallback is invoked when a strategy fires.
type DefaultCallback func(evt TriggerEvent) error

type BaseResolver struct {
	Active   bool
	Callback DefaultCallback
}

func (b *BaseResolver) Deactivate() error {
	if !b.Active {
		return ErrStatusInvalid
	}
	b.Active = false
	return nil
}

func (b *BaseResolver) Trigger(evt TriggerEvent) error {
	if !b.Active {
		return ErrStatusInvalid
	}
	if b.Callback != nil {
		return b.Callback(evt)
	}
	return nil
}
