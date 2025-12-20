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
	"github.com/shopspring/decimal"
)

type TriggerMode int

const (
	TriggerAny TriggerMode = iota
	TriggerAll
)

type Composite struct {
	mode            TriggerMode
	fixedStop       []FixedStopLoss
	fixedProfit     []FixedTakeProfit
	DebouncedStop   []DebouncedStopLoss
	DebouncedProfit []DebouncedTakeProfit
}

func NewComposite(mode TriggerMode) *Composite {
	return &Composite{
		mode: mode,
	}
}

func (c *Composite) AddDebounced(cond1 DebouncedStopLoss, cond2 DebouncedTakeProfit) {
	c.DebouncedStop = append(c.DebouncedStop, cond1)
	c.DebouncedProfit = append(c.DebouncedProfit, cond2)
}

func (c *Composite) AddFixed(cond1 FixedStopLoss, cond2 FixedTakeProfit) {
	c.fixedStop = append(c.fixedStop, cond1)
	c.fixedProfit = append(c.fixedProfit, cond2)
}

func (c *Composite) ShouldTriggerStopLoss(currentPrice decimal.Decimal, timestamp int64) bool {
	count := 0
	total := len(c.fixedStop) + len(c.DebouncedStop)

	for _, cond := range c.fixedStop {
		if triggered, _ := cond.ShouldTriggerStopLoss(currentPrice); triggered {
			if c.mode == TriggerAny {
				return true
			}
			count++
		} else if c.mode == TriggerAll {
			return false
		}
	}

	for _, cond := range c.DebouncedStop {
		if triggered, _ := cond.ShouldTriggerStopLoss(currentPrice, timestamp); triggered {
			if c.mode == TriggerAny {
				return true
			}
			count++
		} else if c.mode == TriggerAll {
			return false
		}
	}
	return c.mode == TriggerAll && count == total
}

func (c *Composite) ShouldTriggerTakeProfit(currentPrice decimal.Decimal, timestamp int64) bool {
	count := 0
	total := len(c.fixedProfit) + len(c.DebouncedProfit)
	for _, cond := range c.fixedProfit {
		if triggered, _ := cond.ShouldTriggerTakeProfit(currentPrice); triggered {
			if c.mode == TriggerAny {
				return true
			}
			count++
		} else if c.mode == TriggerAll {
			return false
		}
	}

	for _, cond := range c.DebouncedProfit {
		if triggered, _ := cond.ShouldTriggerTakeProfit(currentPrice, timestamp); triggered {
			if c.mode == TriggerAny {
				return true
			}
			count++
		} else if c.mode == TriggerAll {
			return false
		}
	}
	return c.mode == TriggerAll && count == total
}

func (c *Composite) GetMinTakeProfit() (decimal.Decimal, error) {
	return c.getTakeProfit(func(a, b decimal.Decimal) bool {
		return a.LessThan(b)
	})
}

func (c *Composite) GetMaxTakeProfit() (decimal.Decimal, error) {
	return c.getTakeProfit(func(a, b decimal.Decimal) bool {
		return a.GreaterThan(b)
	})
}

func (c *Composite) GetMinStopLoss() (decimal.Decimal, error) {
	return c.getStopLoss(func(a, b decimal.Decimal) bool {
		return a.LessThan(b)
	})
}

func (c *Composite) GetMaxStopLoss() (decimal.Decimal, error) {
	return c.getStopLoss(func(a, b decimal.Decimal) bool {
		return a.GreaterThan(b)
	})
}

func (c *Composite) ResetMode(mode TriggerMode) {
	c.mode = mode
}

func (c *Composite) getStopLoss(
	cmp func(a, b decimal.Decimal) bool,
) (decimal.Decimal, error) {
	var result decimal.Decimal
	init := false

	for _, cond := range c.fixedStop {
		stop, err := cond.GetStopLoss()
		if err != nil {
			continue
		}
		if !init || cmp(stop, result) {
			result = stop
			init = true
		}
	}

	for _, cond := range c.DebouncedStop {
		stop, err := cond.GetStopLoss()
		if err != nil {
			continue
		}
		if !init || cmp(stop, result) {
			result = stop
			init = true
		}
	}

	if !init {
		return decimal.Zero, ErrStatusInvalid
	}
	return result, nil
}

func (c *Composite) getTakeProfit(
	cmp func(a, b decimal.Decimal) bool,
) (decimal.Decimal, error) {
	var result decimal.Decimal
	init := false

	for _, cond := range c.fixedProfit {
		take, err := cond.GetTakeProfit()
		if err != nil {
			continue
		}
		if !init || cmp(take, result) {
			result = take
			init = true
		}
	}

	for _, cond := range c.DebouncedProfit {
		take, err := cond.GetTakeProfit()
		if err != nil {
			continue
		}
		if !init || cmp(take, result) {
			result = take
			init = true
		}
	}

	if !init {
		return decimal.Zero, ErrStatusInvalid
	}
	return result, nil
}
