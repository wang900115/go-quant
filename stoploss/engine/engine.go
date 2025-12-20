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
	"log"
	"time"

	"github.com/wang900115/quant/common/sys"
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/result"
	"github.com/wang900115/quant/stoploss"
)

type strategyEngineError struct{ msg string }

func (e *strategyEngineError) Error() string {
	return e.msg
}

var (
	errNoStrategies = &strategyEngineError{"no strategies registered"}
	errNonsupported = &strategyEngineError{"unsupported strategy type"}
)

type Config struct {
	// Buffer size for internal channels
	BufferSize int
	// Buffer size for results channels
	BufferRSize int
	// Timeout for reading data from channels
	ReadTimeout time.Duration
	// Heartbeat interval for strategy processing goroutines
	CheckInterval time.Duration
	// Heartbeat interval from internal channels
	HeartbeatInterval time.Duration
	// Interval between for failed goroutine retry mechanism
	RetryInterval time.Duration
	// Report Callback func
	ReportCallback func(interface{})
}

func DefaultConfig() Config {
	return Config{
		BufferSize:        2048,
		ReadTimeout:       3 * time.Second,
		CheckInterval:     5 * time.Second,
		HeartbeatInterval: 15 * time.Second,
		RetryInterval:     1 * time.Second,
	}
}

type StrategyEngine struct {
	engine    *sys.Engine
	portfolio *Portfolio
	execution *Execution
	Reporter  *Report
	Metrics   *Metrics
	Config    Config
}

func New(config Config) *StrategyEngine {
	return &StrategyEngine{
		engine:    sys.NewEngine(config.RetryInterval, config.CheckInterval),
		portfolio: NewPortfolio(),
		execution: NewExecutionManager(config.BufferSize, config.BufferRSize),
		Reporter:  NewReport(config.ReportCallback),
		Metrics:   NewMetrics(),
		Config:    config,
	}
}

func (csm *StrategyEngine) RegisterStrategy(name string, strategy interface{}) error {
	switch s := strategy.(type) {
	case stoploss.FixedStopLoss:
		csm.portfolio.RegistFixedStoplossStrategy(name, s)
	case stoploss.DebouncedStopLoss:
		csm.portfolio.RegistDebouncedStoplossStrategy(name, s)
	case stoploss.FixedTakeProfit:
		csm.portfolio.RegistFixedTakeProfitStrategy(name, s)
	case stoploss.DebouncedTakeProfit:
		csm.portfolio.RegistDebouncedTakeProfitStrategy(name, s)
	case stoploss.HybridWithoutTime:
		csm.portfolio.RegistHybridFixedStrategy(name, s)
	case stoploss.HybridWithTime:
		csm.portfolio.RegistHybridDebouncedStrategy(name, s)
	default:
		return errNonsupported
	}
	return nil
}

func (csm *StrategyEngine) Start() error {

	if csm.portfolio.count == 0 {
		return errNoStrategies
	}
	goroutineCount := 0
	if len(csm.portfolio.fixedStoplossStrategies) > 0 {
		goroutineCount++
		csm.engine.SafeGo(func(ctx context.Context) {
			csm.handleFixedStopLoss(ctx)
		}, nil)
	}
	if len(csm.portfolio.DebouncedStoplossStrategies) > 0 {
		goroutineCount++
		csm.engine.SafeGo(func(ctx context.Context) {
			csm.handleDebouncedStopLoss(ctx)
		}, nil)
	}
	if len(csm.portfolio.fixedTakeProfitStrategies) > 0 {
		goroutineCount++
		csm.engine.SafeGo(func(ctx context.Context) {
			csm.handleFixedProfit(ctx)
		}, nil)
	}
	if len(csm.portfolio.DebouncedTakeProfitStrategies) > 0 {
		goroutineCount++
		csm.engine.SafeGo(func(ctx context.Context) {
			csm.handleDebouncedProfit(ctx)
		}, nil)
	}
	if len(csm.portfolio.hybridFixedStrategies) > 0 {
		goroutineCount++
		csm.engine.SafeGo(func(ctx context.Context) {
			csm.handleFixedHybrid(ctx)
		}, nil)
	}
	if len(csm.portfolio.hybridDebouncedStrategies) > 0 {
		goroutineCount++
		csm.engine.SafeGo(func(ctx context.Context) {
			csm.handleDebouncedHybrid(ctx)
		}, nil)
	}

	generalResult, hybridResult := csm.execution.getResult()
	if csm.portfolio.openGeneral {
		csm.engine.SafeGo(func(ctx context.Context) {
			csm.Reporter.ProcessGeneralResult(generalResult, ctx)
		}, nil)
	}

	if csm.portfolio.openHybrid {
		csm.engine.SafeGo(func(ctx context.Context) {
			csm.Reporter.ProcessHybridResult(hybridResult, ctx)
		}, nil)
	}

	// Log the number of started goroutines
	log.Printf("Started %d strategy goroutines + %d reporter goroutines", goroutineCount, func() int {
		count := 0
		if csm.portfolio.openGeneral {
			count++
		}
		if csm.portfolio.openHybrid {
			count++
		}
		return count
	}())

	time.Sleep(30 * time.Second)

	log.Println(csm.Snapshot())
	return nil
}

func (csm *StrategyEngine) Snapshot() map[string]int64 {
	return csm.Reporter.Stats()
}

func (csm *StrategyEngine) GetMetrics() map[string]interface{} {
	return csm.Metrics.Stats()
}

func (csm *StrategyEngine) handleFixedStopLoss(ctx context.Context) {
	ticker := time.NewTicker(csm.Config.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("[handleFixedStopLoss] stopped")
			return
		case update := <-csm.execution.fixedStoplossChannel:
			csm.processFixedStopStrategies(update, ctx)
		case <-ticker.C:
			log.Println("[handleFixedStopLoss] heartbeat")
			continue
		}

	}
}

func (csm *StrategyEngine) handleDebouncedStopLoss(ctx context.Context) {
	ticker := time.NewTicker(csm.Config.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("[handleDebouncedStopLoss] stopped")
			return
		case update := <-csm.execution.DebouncedStoplossChannel:
			csm.processDebouncedStopStrategies(update, ctx)
		case <-ticker.C:
			log.Println("[handleDebouncedStopLoss] heartbeat")
			continue
		}
	}
}

func (csm *StrategyEngine) handleFixedProfit(ctx context.Context) {
	ticker := time.NewTicker(csm.Config.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("[handleFixedProfit] stopped")
			return
		case update := <-csm.execution.fixedTakeProfitChannel:
			csm.processFixedProfitStrategies(update, ctx)
		case <-ticker.C:
			log.Println("[handleFixedProfit] heartbeat")
			continue
		}
	}
}

func (csm *StrategyEngine) handleDebouncedProfit(ctx context.Context) {
	ticker := time.NewTicker(csm.Config.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("[handleDebouncedProfit] stopped")
			return
		case update := <-csm.execution.DebouncedTakeProfitChannel:
			csm.processDebouncedProfitStrategies(update, ctx)
		case <-ticker.C:
			log.Println("[handleDebouncedProfit] heartbeat")
			continue
		}
	}
}

func (csm *StrategyEngine) handleFixedHybrid(ctx context.Context) {
	ticker := time.NewTicker(csm.Config.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("[handleFixedHybrid] stopped")
			return
		case update := <-csm.execution.hybridFixedChannel:
			csm.processHybridFixedStrategies(update, ctx)
		case <-ticker.C:
			log.Println("[handleFixedHybrid] heartbeat")
			continue
		}
	}
}

func (csm *StrategyEngine) handleDebouncedHybrid(ctx context.Context) {
	ticker := time.NewTicker(csm.Config.HeartbeatInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Println("[handleDebouncedHybrid] stopped")
			return
		case update := <-csm.execution.hybridDebouncedChannel:
			csm.processHybridDebouncedStrategies(update, ctx)
		case <-ticker.C:
			log.Println("[handleDebouncedHybrid] heartbeat")
			continue
		}
	}
}

func (csm *StrategyEngine) processFixedStopStrategies(update model.PricePoint, ctx context.Context) {
	strategies := csm.portfolio.GetFixedStoplossStrategies()
	for name, strategy := range strategies {
		shouldTrigger, err := strategy.ShouldTriggerStopLoss(update.NewPrice)
		newThreshold, calcErr := strategy.CalculateStopLoss(update.NewPrice)
		if calcErr == nil {
			result := result.NewGeneral(name, model.FIXED, model.STOP_LOSS, update.NewPrice, newThreshold, update.UpdatedAt, time.Duration(0))
			if err == nil {
				result.SetTriggered(shouldTrigger)
			} else {
				result.SetError(err)
			}
			select {
			case csm.execution.generalResults <- *result:
				// successfully sent
			case <-time.After(csm.Config.ReadTimeout):
				csm.Metrics.RecordChannelTimeout(model.FIXED, model.STOP_LOSS)
			case <-ctx.Done():
				return
			}
		}
	}
}

func (csm *StrategyEngine) processDebouncedStopStrategies(update model.PricePoint, ctx context.Context) {
	strategies := csm.portfolio.GetDebouncedStoplossStrategies()
	for name, strategy := range strategies {
		timeThreshold, _ := strategy.GetTimeThreshold()
		shouldTrigger, err := strategy.ShouldTriggerStopLoss(update.NewPrice, update.UpdatedAt.UnixMilli())
		newThreshold, calcErr := strategy.CalculateStopLoss(update.NewPrice)
		if calcErr == nil {
			result := result.NewGeneral(name, model.DEBUNCED, model.STOP_LOSS, update.NewPrice, newThreshold, update.UpdatedAt, time.Duration(timeThreshold))
			if err == nil {
				result.SetTriggered(shouldTrigger)
			} else {
				result.SetError(err)
			}
			select {
			case csm.execution.generalResults <- *result:
				// successfully sent
			case <-time.After(csm.Config.ReadTimeout):
				csm.Metrics.RecordChannelTimeout(model.DEBUNCED, model.STOP_LOSS)
			case <-ctx.Done():
				return
			}
		}
	}
}

func (csm *StrategyEngine) processFixedProfitStrategies(update model.PricePoint, ctx context.Context) {
	strategies := csm.portfolio.GetFixedTakeProfitStrategies()
	for name, strategy := range strategies {
		shouldTrigger, err := strategy.ShouldTriggerTakeProfit(update.NewPrice)
		newThreshold, calcErr := strategy.CalculateTakeProfit(update.NewPrice)
		if calcErr == nil {
			result := result.NewGeneral(name, model.FIXED, model.TAKE_PROFIT, update.NewPrice, newThreshold, update.UpdatedAt, time.Duration(0))
			if err == nil {
				result.SetTriggered(shouldTrigger)
			} else {
				result.SetError(err)
			}
			select {
			case csm.execution.generalResults <- *result:
				// successfully sent
			case <-time.After(csm.Config.ReadTimeout):
				csm.Metrics.RecordChannelTimeout(model.FIXED, model.TAKE_PROFIT)
			case <-ctx.Done():
				return
			}
		}
	}
}

func (csm *StrategyEngine) processDebouncedProfitStrategies(update model.PricePoint, ctx context.Context) {
	strategies := csm.portfolio.GetDebouncedTakeProfitStrategies()
	for name, strategy := range strategies {
		timeThreshold, _ := strategy.GetTimeThreshold()
		shouldTrigger, err := strategy.ShouldTriggerTakeProfit(update.NewPrice, update.UpdatedAt.UnixMilli())
		newThreshold, calcErr := strategy.CalculateTakeProfit(update.NewPrice)
		if calcErr == nil {
			result := result.NewGeneral(name, model.DEBUNCED, model.TAKE_PROFIT, update.NewPrice, newThreshold, update.UpdatedAt, time.Duration(timeThreshold))
			if err == nil {
				result.SetTriggered(shouldTrigger)
			} else {
				result.SetError(err)
			}
			select {
			case csm.execution.generalResults <- *result:
				// successfully sent
			case <-time.After(csm.Config.ReadTimeout):
				csm.Metrics.RecordChannelTimeout(model.DEBUNCED, model.TAKE_PROFIT)
			case <-ctx.Done():
				return
			}
		}
	}
}

func (csm *StrategyEngine) processHybridFixedStrategies(update model.PricePoint, ctx context.Context) {
	strategies := csm.portfolio.GetHybridStrategies()
	for name, strategy := range strategies {
		shouldTriggerSL, errSL := strategy.ShouldTriggerStopLoss(update.NewPrice)
		shouldTriggerTP, errTP := strategy.ShouldTriggerTakeProfit(update.NewPrice)
		newStop, newProfit, calcErr := strategy.Calculate(update.NewPrice)
		if calcErr == nil {
			result := result.NewHybrid(name, model.HYBRID_FIXED, update.NewPrice, newStop, newProfit, update.UpdatedAt, time.Duration(0))
			if errSL == nil && shouldTriggerSL {
				result.SetTriggered(true, model.STOP_LOSS)
			} else if errTP == nil && shouldTriggerTP {
				result.SetTriggered(true, model.TAKE_PROFIT)
			} else if errSL != nil {
				result.SetError(errSL)
			} else if errTP != nil {
				result.SetError(errTP)
			}

			select {
			case csm.execution.hybridResults <- *result:
				// successfully sent
			case <-time.After(csm.Config.ReadTimeout):
				csm.Metrics.RecordChannelTimeout(model.HYBRID_FIXED, "")
			case <-ctx.Done():
				return
			}
		}
	}
}

func (csm *StrategyEngine) processHybridDebouncedStrategies(update model.PricePoint, ctx context.Context) {
	strategies := csm.portfolio.GetHybridStrategies()
	for name, strategy := range strategies {
		shouldTriggerSL, errSL := strategy.ShouldTriggerStopLoss(update.NewPrice)
		shouldTriggerTP, errTP := strategy.ShouldTriggerTakeProfit(update.NewPrice)
		newStop, newProfit, calcErr := strategy.Calculate(update.NewPrice)
		if calcErr == nil {
			result := result.NewHybrid(name, model.HYBRID_DEBUNCED, update.NewPrice, newStop, newProfit, update.UpdatedAt, time.Duration(0))
			if errSL == nil && shouldTriggerSL {
				result.SetTriggered(true, model.STOP_LOSS)
			} else if errTP == nil && shouldTriggerTP {
				result.SetTriggered(true, model.TAKE_PROFIT)
			} else if errSL != nil {
				result.SetError(errSL)
			} else if errTP != nil {
				result.SetError(errTP)
			}

			select {
			case csm.execution.hybridResults <- *result:
				// successfully sent
			case <-time.After(csm.Config.ReadTimeout):
				csm.Metrics.RecordChannelTimeout(model.HYBRID_DEBUNCED, "")
			case <-ctx.Done():
				return
			}
		}
	}
}

func (csm *StrategyEngine) Collect(pricePoint model.PricePoint, callback func()) {
	csm.Metrics.RecordReceived()

	if len(csm.portfolio.fixedStoplossStrategies) > 0 {
		dataFeedWithMetrics(pricePoint, csm.execution.fixedStoplossChannel, model.FIXED, model.STOP_LOSS, csm.Metrics, callback)
	}
	if len(csm.portfolio.DebouncedStoplossStrategies) > 0 {
		dataFeedWithMetrics(pricePoint, csm.execution.DebouncedStoplossChannel, model.DEBUNCED, model.STOP_LOSS, csm.Metrics, callback)
	}
	if len(csm.portfolio.fixedTakeProfitStrategies) > 0 {
		dataFeedWithMetrics(pricePoint, csm.execution.fixedTakeProfitChannel, model.FIXED, model.TAKE_PROFIT, csm.Metrics, callback)
	}
	if len(csm.portfolio.DebouncedTakeProfitStrategies) > 0 {
		dataFeedWithMetrics(pricePoint, csm.execution.DebouncedTakeProfitChannel, model.DEBUNCED, model.TAKE_PROFIT, csm.Metrics, callback)
	}
	if len(csm.portfolio.hybridFixedStrategies) > 0 {
		dataFeedWithMetrics(pricePoint, csm.execution.hybridFixedChannel, model.HYBRID_FIXED, "", csm.Metrics, callback)
	}
	if len(csm.portfolio.hybridDebouncedStrategies) > 0 {
		dataFeedWithMetrics(pricePoint, csm.execution.hybridDebouncedChannel, model.HYBRID_DEBUNCED, "", csm.Metrics, callback)
	}
}

func (csm *StrategyEngine) Stop() {
	csm.engine.Stop()
	csm.execution.closeChannels()
}

func dataFeedWithMetrics(pricePoint model.PricePoint, channel chan model.PricePoint, typ model.StrategyType, category model.StrategyCategory, metrics *Metrics, callback func()) {
	select {
	case channel <- pricePoint:
		metrics.RecordChannelSend(typ, category)
	default:
		// Channel is full
		metrics.RecordChannelDrop(typ, category)
		metrics.RecordDropped()
		callback()
	}
}
