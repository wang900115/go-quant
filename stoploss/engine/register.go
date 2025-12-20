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

package engine

import (
	"sync"

	"github.com/wang900115/quant/stoploss"
)

type Portfolio struct {
	mutex                         sync.Mutex
	fixedStoplossStrategies       map[string]stoploss.FixedStopLoss
	DebouncedStoplossStrategies   map[string]stoploss.DebouncedStopLoss
	fixedTakeProfitStrategies     map[string]stoploss.FixedTakeProfit
	DebouncedTakeProfitStrategies map[string]stoploss.DebouncedTakeProfit
	hybridFixedStrategies         map[string]stoploss.HybridWithoutTime
	hybridDebouncedStrategies     map[string]stoploss.HybridWithTime
	openGeneral                   bool
	openHybrid                    bool
	count                         int
}

func NewPortfolio() *Portfolio {
	return &Portfolio{
		fixedStoplossStrategies:       make(map[string]stoploss.FixedStopLoss),
		DebouncedStoplossStrategies:   make(map[string]stoploss.DebouncedStopLoss),
		fixedTakeProfitStrategies:     make(map[string]stoploss.FixedTakeProfit),
		DebouncedTakeProfitStrategies: make(map[string]stoploss.DebouncedTakeProfit),
		hybridFixedStrategies:         make(map[string]stoploss.HybridWithoutTime),
		hybridDebouncedStrategies:     make(map[string]stoploss.HybridWithTime),
		openGeneral:                   false,
		openHybrid:                    false,
		count:                         0,
	}
}

func (p *Portfolio) RegistFixedStoplossStrategy(name string, strategy stoploss.FixedStopLoss) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.fixedStoplossStrategies[name] = strategy
	p.openGeneral = true
	p.count++
}

func (p *Portfolio) RegistDebouncedStoplossStrategy(name string, strategy stoploss.DebouncedStopLoss) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.DebouncedStoplossStrategies[name] = strategy
	p.openGeneral = true
	p.count++
}

func (p *Portfolio) RegistFixedTakeProfitStrategy(name string, strategy stoploss.FixedTakeProfit) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.fixedTakeProfitStrategies[name] = strategy
	p.openGeneral = true
	p.count++
}

func (p *Portfolio) RegistDebouncedTakeProfitStrategy(name string, strategy stoploss.DebouncedTakeProfit) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.DebouncedTakeProfitStrategies[name] = strategy
	p.openGeneral = true
	p.count++
}

func (p *Portfolio) RegistHybridFixedStrategy(name string, strategy stoploss.HybridWithoutTime) {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.hybridFixedStrategies[name] = strategy
	p.openHybrid = true
	p.count++
}

func (p *Portfolio) RegistHybridDebouncedStrategy(name string, strategy stoploss.HybridWithTime) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.hybridDebouncedStrategies[name] = strategy
	p.openHybrid = true
	p.count++
}

func (p *Portfolio) GetFixedStoplossStrategies() map[string]stoploss.FixedStopLoss {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Return a copy to avoid race conditions
	copyMap := make(map[string]stoploss.FixedStopLoss)
	for k, v := range p.fixedStoplossStrategies {
		copyMap[k] = v
	}
	return copyMap
}

func (p *Portfolio) GetDebouncedStoplossStrategies() map[string]stoploss.DebouncedStopLoss {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Return a copy
	copyMap := make(map[string]stoploss.DebouncedStopLoss)
	for k, v := range p.DebouncedStoplossStrategies {
		copyMap[k] = v
	}
	return copyMap
}

func (p *Portfolio) GetFixedTakeProfitStrategies() map[string]stoploss.FixedTakeProfit {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Return a copy to avoid race conditions
	copyMap := make(map[string]stoploss.FixedTakeProfit)
	for k, v := range p.fixedTakeProfitStrategies {
		copyMap[k] = v
	}
	return copyMap
}

func (p *Portfolio) GetDebouncedTakeProfitStrategies() map[string]stoploss.DebouncedTakeProfit {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Return a copy to avoid race conditions
	copyMap := make(map[string]stoploss.DebouncedTakeProfit)
	for k, v := range p.DebouncedTakeProfitStrategies {
		copyMap[k] = v
	}
	return copyMap
}

func (p *Portfolio) GetHybridStrategies() map[string]stoploss.HybridWithoutTime {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Return a copy to avoid race conditions
	copyMap := make(map[string]stoploss.HybridWithoutTime)
	for k, v := range p.hybridFixedStrategies {
		copyMap[k] = v
	}
	return copyMap
}

func (p *Portfolio) GetHybridDebouncedStrategies() map[string]stoploss.HybridWithTime {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	// Return a copy to avoid race conditions
	copyMap := make(map[string]stoploss.HybridWithTime)
	for k, v := range p.hybridDebouncedStrategies {
		copyMap[k] = v
	}
	return copyMap
}
