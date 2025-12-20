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
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/result"
)

type Execution struct {
	fixedStoplossChannel       chan model.PricePoint
	DebouncedStoplossChannel   chan model.PricePoint
	fixedTakeProfitChannel     chan model.PricePoint
	DebouncedTakeProfitChannel chan model.PricePoint
	hybridFixedChannel         chan model.PricePoint
	hybridDebouncedChannel     chan model.PricePoint

	generalResults chan result.StrategyGeneralResult
	hybridResults  chan result.StrategyHybridResult
}

func NewExecutionManager(bufferSize int, bufferRSize int) *Execution {
	return &Execution{
		fixedStoplossChannel:       make(chan model.PricePoint, bufferSize),
		DebouncedStoplossChannel:   make(chan model.PricePoint, bufferSize),
		fixedTakeProfitChannel:     make(chan model.PricePoint, bufferSize),
		DebouncedTakeProfitChannel: make(chan model.PricePoint, bufferSize),
		hybridFixedChannel:         make(chan model.PricePoint, bufferSize),
		hybridDebouncedChannel:     make(chan model.PricePoint, bufferSize),
		generalResults:             make(chan result.StrategyGeneralResult, bufferRSize),
		hybridResults:              make(chan result.StrategyHybridResult, bufferRSize),
	}
}

func (e *Execution) getResult() (<-chan result.StrategyGeneralResult, <-chan result.StrategyHybridResult) {
	return e.generalResults, e.hybridResults
}

func (e *Execution) closeChannels() {
	close(e.fixedStoplossChannel)
	close(e.DebouncedStoplossChannel)
	close(e.fixedTakeProfitChannel)
	close(e.DebouncedTakeProfitChannel)
	close(e.hybridFixedChannel)
	close(e.hybridDebouncedChannel)
	close(e.generalResults)
	close(e.hybridResults)
}
