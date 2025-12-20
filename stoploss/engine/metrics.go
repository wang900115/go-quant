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
	"time"

	"github.com/wang900115/quant/metric"
	"github.com/wang900115/quant/model"
)

// Metrics tracks various statistics about the engine's performance work for stream receiver
type Metrics struct {
	// TotalReceived counts the total number of messages received
	TotalReceived *metric.CounterInt64
	// TotalDropped counts the total number of messages dropped
	TotalDropped *metric.CounterInt64

	// FixedStopReceived counts the number of fixed stop loss messages received
	FixedStopReceived *metric.CounterInt64
	// DebouncedStopReceived counts the number of Debounced stop loss messages received
	DebouncedStopReceived *metric.CounterInt64
	// FixedProfitReceived counts the number of fixed profit messages received
	FixedProfitReceived *metric.CounterInt64
	// DebouncedProfitReceived counts the number of Debounced profit messages received
	DebouncedProfitReceived *metric.CounterInt64
	// HybridFixedReceived counts the number of hybrid fixed messages received
	HybridFixedReceived *metric.CounterInt64
	// HybridDebouncedReceived counts the number of hybrid Debounced messages received
	HybridDebouncedReceived *metric.CounterInt64

	// FixedStopDropped counts the number of fixed stop loss messages dropped
	FixedStopDropped *metric.CounterInt64
	// DebouncedStopDropped counts the number of Debounced stop loss messages dropped
	DebouncedStopDropped *metric.CounterInt64
	// FixedProfitDropped counts the number of fixed profit messages dropped
	FixedProfitDropped *metric.CounterInt64
	// DebouncedProfitDropped counts the number of Debounced profit messages dropped
	DebouncedProfitDropped *metric.CounterInt64
	// HybridFixedDropped counts the number of hybrid fixed messages dropped
	HybridFixedDropped *metric.CounterInt64
	// HybridDebouncedDropped counts the number of hybrid Debounced messages dropped
	HybridDebouncedDropped *metric.CounterInt64

	// FixedStopTimeout counts the number of fixed stop loss messages Debounced out
	FixedStopTimeout *metric.CounterInt64
	// DebouncedStopTimeout counts the number of Debounced stop loss messages Debounced out
	DebouncedStopTimeout *metric.CounterInt64
	// FixedProfitTimeout counts the number of fixed profit messages Debounced out
	FixedProfitTimeout *metric.CounterInt64
	// DebouncedProfitTimeout counts the number of Debounced profit messages Debounced out
	DebouncedProfitTimeout *metric.CounterInt64
	// HybridFixedTimeout counts the number of hybrid fixed messages Debounced out
	HybridFixedTimeout *metric.CounterInt64
	// HybridDebouncedTimeout counts the number of hybrid Debounced messages Debounced out
	HybridDebouncedTimeout *metric.CounterInt64

	// StartTime records the time when the metrics tracking started
	StartTime time.Time
}

// NewMetrics creates a new Metrics instance
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime:               time.Now(),
		TotalReceived:           metric.NewCounterInt64(),
		TotalDropped:            metric.NewCounterInt64(),
		FixedStopReceived:       metric.NewCounterInt64(),
		DebouncedStopReceived:   metric.NewCounterInt64(),
		FixedProfitReceived:     metric.NewCounterInt64(),
		DebouncedProfitReceived: metric.NewCounterInt64(),
		HybridFixedReceived:     metric.NewCounterInt64(),
		HybridDebouncedReceived: metric.NewCounterInt64(),
		FixedStopDropped:        metric.NewCounterInt64(),
		DebouncedStopDropped:    metric.NewCounterInt64(),
		FixedProfitDropped:      metric.NewCounterInt64(),
		DebouncedProfitDropped:  metric.NewCounterInt64(),
		HybridFixedDropped:      metric.NewCounterInt64(),
		HybridDebouncedDropped:  metric.NewCounterInt64(),
		FixedStopTimeout:        metric.NewCounterInt64(),
		DebouncedStopTimeout:    metric.NewCounterInt64(),
		FixedProfitTimeout:      metric.NewCounterInt64(),
		DebouncedProfitTimeout:  metric.NewCounterInt64(),
		HybridFixedTimeout:      metric.NewCounterInt64(),
		HybridDebouncedTimeout:  metric.NewCounterInt64(),
	}
}

// RecordReceived increments the total received counter
func (m *Metrics) RecordReceived() {
	m.TotalReceived.Inc(1)
}

// RecordDropped increments the total dropped counter
func (m *Metrics) RecordDropped() {
	m.TotalDropped.Inc(1)
}

// RecordChannelSend records a successful send to a specific channel
func (m *Metrics) RecordChannelSend(typ model.StrategyType, channel model.StrategyCategory) {
	switch typ {
	case model.HYBRID_FIXED:
		m.HybridFixedReceived.Inc(1)
	case model.HYBRID_DEBUNCED:
		m.HybridDebouncedReceived.Inc(1)
	case model.FIXED:
		switch channel {
		case model.STOP_LOSS:
			m.FixedStopReceived.Inc(1)
		case model.TAKE_PROFIT:
			m.FixedProfitReceived.Inc(1)
		}
	case model.DEBUNCED:
		switch channel {
		case model.STOP_LOSS:
			m.DebouncedStopReceived.Inc(1)
		case model.TAKE_PROFIT:
			m.DebouncedProfitReceived.Inc(1)
		}
	}
}

// RecordChannelDrop records a dropped message for a specific channel
func (m *Metrics) RecordChannelDrop(typ model.StrategyType, channel model.StrategyCategory) {
	switch typ {
	case model.HYBRID_FIXED:
		m.HybridFixedDropped.Inc(1)
	case model.HYBRID_DEBUNCED:
		m.HybridDebouncedDropped.Inc(1)
	case model.FIXED:
		switch channel {
		case model.STOP_LOSS:
			m.FixedStopDropped.Inc(1)
		case model.TAKE_PROFIT:
			m.FixedProfitDropped.Inc(1)
		}
	case model.DEBUNCED:
		switch channel {
		case model.STOP_LOSS:
			m.DebouncedStopDropped.Inc(1)
		case model.TAKE_PROFIT:
			m.DebouncedProfitDropped.Inc(1)
		}
	}
}

// RecordChannelTimeout records a timeout for a specific channel
func (m *Metrics) RecordChannelTimeout(typ model.StrategyType, channel model.StrategyCategory) {
	switch typ {
	case model.HYBRID_FIXED:
		m.HybridFixedTimeout.Inc(1)
	case model.HYBRID_DEBUNCED:
		m.HybridDebouncedTimeout.Inc(1)
	case model.FIXED:
		switch channel {
		case model.STOP_LOSS:
			m.FixedStopTimeout.Inc(1)
		case model.TAKE_PROFIT:
			m.FixedProfitTimeout.Inc(1)
		}
	case model.DEBUNCED:
		switch channel {
		case model.STOP_LOSS:
			m.DebouncedStopTimeout.Inc(1)
		case model.TAKE_PROFIT:
			m.DebouncedProfitTimeout.Inc(1)
		}
	}
}

// Stats returns a snapshot of current statistics
func (m *Metrics) Stats() map[string]interface{} {
	uptime := time.Since(m.StartTime)

	totalReceived := m.TotalReceived.Snapshot()
	totalDropped := m.TotalDropped.Snapshot()

	dropRate := float64(0)
	if totalReceived > 0 {
		dropRate = float64(totalDropped) / float64(totalReceived) * 100
	}

	return map[string]interface{}{
		"uptime_seconds":    uptime.Seconds(),
		"total_received":    totalReceived,
		"total_dropped":     totalDropped,
		"drop_rate_percent": dropRate,

		"channels": map[string]interface{}{
			"fixed_stop": map[string]int64{
				"received": m.FixedStopReceived.Snapshot().Count(),
				"dropped":  m.FixedStopDropped.Snapshot().Count(),
				"timeout":  m.FixedStopTimeout.Snapshot().Count(),
			},
			"Debounced_stop": map[string]int64{
				"received": m.DebouncedStopReceived.Snapshot().Count(),
				"dropped":  m.DebouncedStopDropped.Snapshot().Count(),
				"timeout":  m.DebouncedStopTimeout.Snapshot().Count(),
			},
			"fixed_profit": map[string]int64{
				"received": m.FixedProfitReceived.Snapshot().Count(),
				"dropped":  m.FixedProfitDropped.Snapshot().Count(),
				"timeout":  m.FixedProfitTimeout.Snapshot().Count(),
			},
			"Debounced_profit": map[string]int64{
				"received": m.DebouncedProfitReceived.Snapshot().Count(),
				"dropped":  m.DebouncedProfitDropped.Snapshot().Count(),
				"timeout":  m.DebouncedProfitTimeout.Snapshot().Count(),
			},
			"hybrid_fixed": map[string]int64{
				"received": m.HybridFixedReceived.Snapshot().Count(),
				"dropped":  m.HybridFixedDropped.Snapshot().Count(),
				"timeout":  m.HybridFixedTimeout.Snapshot().Count(),
			},
			"hybrid_Debounced": map[string]int64{
				"received": m.HybridDebouncedReceived.Snapshot().Count(),
				"dropped":  m.HybridDebouncedDropped.Snapshot().Count(),
				"timeout":  m.HybridDebouncedTimeout.Snapshot().Count(),
			},
		},
	}
}

// GetDropRate returns the current drop rate as a percentage
func (m *Metrics) GetDropRate() float64 {
	received := m.TotalReceived.Snapshot()
	if received == 0 {
		return 0
	}
	dropped := m.TotalDropped.Snapshot()
	return float64(dropped) / float64(received) * 100
}
