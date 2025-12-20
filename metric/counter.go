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

package metric

import "sync/atomic"

// counterInt64Snapshot represents a static snapshot of a counter's value.
type counterInt64Snapshot int64

// Count returns the value of the counter's static snapshot for int64.
func (c counterInt64Snapshot) Count() int64 {
	return int64(c)
}

// CounterInt64 is a thread-safe counter for int64 values.
type CounterInt64 atomic.Int64

// NewCounterInt64 creates and returns a new instance of CounterInt64.
func NewCounterInt64() *CounterInt64 {
	return new(CounterInt64)
}

// Inc increments the counter by the specified delta.
func (c *CounterInt64) Inc(delta int64) {
	(*atomic.Int64)(c).Add(delta)
}

// Dec decrements the counter by the specified delta.
func (c *CounterInt64) Dec(delta int64) {
	(*atomic.Int64)(c).Add(-delta)
}

// Clear resets the counter to zero.
func (c *CounterInt64) Clear() {
	(*atomic.Int64)(c).Store(0)
}

// Fork creates a new CounterInt64 instance but shares the same value.
func (c *CounterInt64) Fork() *CounterInt64 {
	var forked atomic.Int64
	forked.Store((*atomic.Int64)(c).Load())
	return (*CounterInt64)(&forked)
}

// Snapshot returns a static snapshot of the current counter value
func (c *CounterInt64) Snapshot() counterInt64Snapshot {
	return counterInt64Snapshot((*atomic.Int64)(c).Load())
}
