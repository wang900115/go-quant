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

	"github.com/shopspring/decimal"
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
	Trigger(reason string) error
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

type DefaultCallback func(reason string) error

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

func (b *BaseResolver) Trigger(reason string) error {
	if !b.Active {
		return ErrStatusInvalid
	}
	if b.Callback != nil {
		return b.Callback(reason)
	}
	return nil
}
