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
	"context"
	"fmt"

	"github.com/wang900115/quant/metric"
	"github.com/wang900115/quant/model/result"
)

type Report struct {
	generalCount metric.CounterInt64
	hybridCount  metric.CounterInt64
	triggerCount metric.CounterInt64
	errorCount   metric.CounterInt64
	Callback     func(interface{})
}

func NewReport(Callback func(interface{})) *Report {
	return &Report{
		Callback: Callback,
	}
}

func (rp *Report) ProcessGeneralResult(res <-chan result.StrategyGeneralResult, ctx context.Context) {
	rp.generalCount.Inc(1)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("🔄 GeneralResult processor shutting down...")
			return
		case r, ok := <-res:
			if !ok {
				fmt.Println("🔄 GeneralResult channel closed")
				return
			}

			if r.Error != nil {
				rp.errorCount.Inc(1)
				fmt.Printf("🔴 ERROR in %s (%s): %v\n",
					r.StrategyName, r.StrategyType, r.Error)
				continue
			}

			if r.Triggered {
				rp.triggerCount.Inc(1)
				fmt.Printf("🔔 TRIGGER: %s (%s) - %s at price %s\n",
					r.StrategyName, r.StrategyType, r.TriggerType,
					r.LastPrice.String())
				if rp.Callback != nil {
					rp.Callback(r)
				}
			} else {
				fmt.Printf("📊 UPDATE: %s (%s) - threshold: %s, price: %s\n",
					r.StrategyName, r.StrategyType,
					r.Stat.PriceThreshold.String(), r.LastPrice.String())
			}
		}
	}
}

func (rp *Report) ProcessHybridResult(res <-chan result.StrategyHybridResult, ctx context.Context) {
	rp.hybridCount.Inc(1)
	for {
		select {
		case <-ctx.Done():
			fmt.Println("🔄 HybridResult processor shutting down...")
			return
		case r, ok := <-res:
			if !ok {
				fmt.Println("🔄 HybridResult channel closed")
				return
			}

			if r.Error != nil {
				rp.errorCount.Inc(1)
				fmt.Printf("🔴 HYBRID ERROR in %s: %v\n", r.StrategyName, r.Error)
				continue
			}

			if r.Triggered {
				rp.triggerCount.Inc(1)
				fmt.Printf("🔔 HYBRID TRIGGER: %s at price %s stoploss at %s take profit at %s\n",
					r.StrategyName, r.LastPrice.String(), r.StopStat.PriceThreshold.String(), r.ProfitStat.PriceThreshold.String())
				if rp.Callback != nil {
					rp.Callback(r)
				}
			} else {
				fmt.Printf("📊 HYBRID UPDATE: %s at price %s stoploss at %s take profit at %s\n",
					r.StrategyName, r.LastPrice.String(), r.StopStat.PriceThreshold.String(), r.ProfitStat.PriceThreshold.String())
			}
		}
	}
}

func (rp *Report) Stats() map[string]int64 {
	return map[string]int64{
		"general_results": rp.generalCount.Snapshot().Count(),
		"hybrid_results":  rp.hybridCount.Snapshot().Count(),
		"triggers":        rp.triggerCount.Snapshot().Count(),
		"errors":          rp.errorCount.Snapshot().Count(),
	}
}
