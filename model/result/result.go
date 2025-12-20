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

package result

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/model"
)

type StrategyStat struct {
	PriceThreshold decimal.Decimal
}

type StrategyResult interface {
	StrategyHybridResult | StrategyGeneralResult
}

type StrategyGeneralResult struct {
	StrategyName  string
	StrategyType  model.StrategyType
	Triggered     bool
	TriggerType   model.StrategyCategory
	LastPrice     decimal.Decimal
	Stat          StrategyStat
	LastTime      time.Time
	TimeThreshold time.Duration
	Error         error
}

type StrategyHybridResult struct {
	StrategyName  string
	StrategyType  model.StrategyType
	Triggered     bool
	TriggerType   model.StrategyCategory
	LastTime      time.Time
	LastPrice     decimal.Decimal
	TimeThreshold time.Duration
	StopStat      StrategyStat
	ProfitStat    StrategyStat
	Error         error
}

func (sr *StrategyGeneralResult) Marshall() map[string]interface{} {
	return map[string]interface{}{
		"StrategyName":  sr.StrategyName,
		"StrategyType":  sr.StrategyType,
		"Triggered":     sr.Triggered,
		"TriggerType":   sr.TriggerType,
		"Stat":          sr.Stat,
		"LastTime":      sr.LastTime,
		"TimeThreshold": sr.TimeThreshold,
		"Error":         sr.Error,
	}
}

func (sr *StrategyHybridResult) Marshall() map[string]interface{} {
	return map[string]interface{}{
		"StrategyName":  sr.StrategyName,
		"StrategyType":  sr.StrategyType,
		"Triggered":     sr.Triggered,
		"TriggerType":   sr.TriggerType,
		"LastTime":      sr.LastTime,
		"TimeThreshold": sr.TimeThreshold,
		"StopStat":      sr.StopStat,
		"ProfitStat":    sr.ProfitStat,
		"Error":         sr.Error,
	}
}

func NewGeneral(strategyName string, strategyType model.StrategyType, triggerType model.StrategyCategory, lastPrice, priceThreshold decimal.Decimal, lastTime time.Time, timeThreshold time.Duration) *StrategyGeneralResult {
	return &StrategyGeneralResult{
		StrategyName:  strategyName,
		StrategyType:  strategyType,
		TriggerType:   triggerType,
		LastTime:      lastTime,
		TimeThreshold: timeThreshold,
		LastPrice:     lastPrice,
		Stat: StrategyStat{
			PriceThreshold: priceThreshold,
		},
	}
}

func NewHybrid(strategyName string, strategyType model.StrategyType, LastPrice, stopPriceThreshold, profitPriceThreshold decimal.Decimal, lastTime time.Time, timeThreshold time.Duration) *StrategyHybridResult {
	return &StrategyHybridResult{
		StrategyName:  strategyName,
		StrategyType:  strategyType,
		LastTime:      lastTime,
		TimeThreshold: timeThreshold,
		LastPrice:     LastPrice,
		StopStat: StrategyStat{
			PriceThreshold: stopPriceThreshold,
		},
		ProfitStat: StrategyStat{
			PriceThreshold: profitPriceThreshold,
		},
	}
}

func (sr *StrategyGeneralResult) SetError(err error) {
	sr.Error = err
}

func (sr *StrategyGeneralResult) SetTriggered(triggered bool) {
	sr.Triggered = triggered
}

func (sr *StrategyHybridResult) SetTriggered(triggered bool, triggerType model.StrategyCategory) {
	sr.Triggered = triggered
	sr.TriggerType = triggerType
}

func (sr *StrategyHybridResult) SetError(err error) {
	sr.Error = err
}

func (sr *StrategyGeneralResult) String() string {
	errorStr := "<nil>"
	if sr.Error != nil {
		errorStr = sr.Error.Error()
	}
	return fmt.Sprintf("StrategyGeneralResult{StrategyName: %s, StrategyType: %s, Triggered: %t, TriggerType: %s, LastPrice: %s, PriceThreshold: %s, LastTime: %s, TimeThreshold: %s, Error: %v}",
		sr.StrategyName,
		sr.StrategyType,
		sr.Triggered,
		sr.TriggerType,
		sr.LastPrice.String(),
		sr.Stat.PriceThreshold.String(),
		sr.LastTime.String(),
		sr.TimeThreshold.String(),
		errorStr,
	)
}

func (sr *StrategyHybridResult) String() string {
	errorStr := "<nil>"
	if sr.Error != nil {
		errorStr = sr.Error.Error()
	}
	return fmt.Sprintf("StrategyHybridResult{StrategyName: %s, StrategyType: %s, Triggered: %t, TriggerType: %s, LastPrice: %s, StopPriceThreshold: %s, ProfitPriceThreshold: %s, LastTime: %s, TimeThreshold: %s, Error: %s}",
		sr.StrategyName,
		sr.StrategyType,
		sr.Triggered,
		sr.TriggerType,
		sr.LastPrice.String(),
		sr.StopStat.PriceThreshold.String(),
		sr.ProfitStat.PriceThreshold.String(),
		sr.LastTime.String(),
		sr.TimeThreshold.String(),
		errorStr,
	)
}

func (sr *StrategyGeneralResult) JSONMarshall() string {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(sr.Marshall()); err != nil {
		return ""
	}
	return buf.String()
}

func (sr *StrategyHybridResult) JSONMarshall() string {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(sr.Marshall()); err != nil {
		return ""
	}
	return buf.String()
}
