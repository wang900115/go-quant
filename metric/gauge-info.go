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

import (
	"encoding/json"
	"sync"
)

// GaugeInfoSnapshot represents a snapshot of the GaugeInfo at a point in time
type GaugeInfoSnapshot GaugeInfoValue

// Value returns the GaugeInfoValue representation of the snapshot
func (g GaugeInfoSnapshot) Value() GaugeInfoValue { return GaugeInfoValue(g) }

// GaugeInfoValue represents key-value pairs stored in GaugeInfo
type GaugeInfoValue map[string]string

// String returns the JSON representation of the GaugeInfoValue
func (g GaugeInfoValue) String() string { data, _ := json.Marshal(g); return string(data) }

// GaugeInfo is a thread-safe structure for storing key-value pairs
type GaugeInfo struct {
	mutex sync.Mutex
	value GaugeInfoValue
}

// NewGaugeInfo creates and returns a new GaugeInfo instance
func NewGaugeInfo() *GaugeInfo {
	return &GaugeInfo{
		value: make(GaugeInfoValue),
	}
}

// Snapshot creates a snapshot of the current GaugeInfo state
func (g *GaugeInfo) Snapshot() GaugeInfoSnapshot {
	return GaugeInfoSnapshot(g.value)
}

// Set sets a key-value pair in the GaugeInfo
func (g *GaugeInfo) Set(key, val string) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.value[key] = val
}
