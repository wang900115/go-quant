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
	"testing"

	"github.com/shopspring/decimal"
)

func TestCounterPrecisionClear(t *testing.T) {
	c := NewCounterPrecision(4)
	c.Inc(decimal.NewFromFloat(10.12345678))
	c.Clear()
	if count := c.Snapshot().Count(); !count.Equal(decimal.NewFromFloat(0).Round(4)) {
		t.Errorf("expected count to be 0 after Clear, got %s", count.String())
	}
}

func TestDefaultCounterPrecision(t *testing.T) {
	c := NewCounterPrecision(0)
	c.Inc(decimal.NewFromFloat(1.123456789))
	if count := c.Snapshot().Count(); !count.Equal(decimal.NewFromFloat(1.123456785).Round(8)) {
		t.Errorf("expected count to be 1.12345678 after Incrementing, got %s", count.String())
	}
}

func TestCounterPrecision(t *testing.T) {
	c := NewCounterPrecision(3)
	if count := c.Snapshot().Count(); !count.Equal(decimal.NewFromFloat(0).Round(3)) {
		t.Errorf("expected initial count to be 0, got %s", count.String())
	}
	c.Dec(decimal.NewFromFloat(1.5678))
	if count := c.Snapshot().Count(); !count.Equal(decimal.NewFromFloat(-1.568)) {
		t.Errorf("expected count to be -1.568 after Decrementing by 1.5678, got %s", count.String())
	}
	c.Inc(decimal.NewFromFloat(5.4321))
	if count := c.Snapshot().Count(); !count.Equal(decimal.NewFromFloat(3.864)) {
		t.Errorf("expected count to be 3.864 after Incrementing by 5.4321, got %s", count.String())
	}
	c.Dec(decimal.NewFromFloat(2.1111))
	if count := c.Snapshot().Count(); !count.Equal(decimal.NewFromFloat(1.753)) {
		t.Errorf("expected count to be 1.753 after Decrementing by 2.1111, got %s", count.String())
	}
	c.Inc(decimal.NewFromFloat(4.8765))
	if count := c.Snapshot().Count(); !count.Equal(decimal.NewFromFloat(6.63)) {
		t.Errorf("expected count to be 6.629 after Incrementing by 4.8765, got %s", count.String())
	}
}

func TestCounterPrecisionSnapshot(t *testing.T) {
	c := NewCounterPrecision(5)
	c.Inc(decimal.NewFromFloat(7.123456))
	if count := c.Snapshot().Count(); !count.Equal(decimal.NewFromFloat(7.12346)) {
		t.Errorf("expected count to be 7.12346 after Incrementing by 7.123456, got %s", count.String())
	}
	snapshot := c.Snapshot()
	c.Inc(decimal.NewFromFloat(1.00001))
	if !snapshot.Count().Equal(decimal.NewFromFloat(7.12346)) {
		t.Errorf("expected snapshot count to be 7.12346, got %s", snapshot.Count().String())
	}
	if !c.Snapshot().Count().Equal(decimal.NewFromFloat(8.12347)) {
		t.Errorf("expected current count to be 8.12347 after Incrementing by 1.00001, got %s", c.Snapshot().Count().String())
	}
}

func TestCounterPrecisionFork(t *testing.T) {
	c := NewCounterPrecision(6)
	c.Inc(decimal.NewFromFloat(15.654321))
	forked := c.Fork()
	if count := forked.Snapshot().Count(); !count.Equal(decimal.NewFromFloat(15.654321)) {
		t.Errorf("expected forked count to be 15.654321, got %s", count.String())
	}
	forked.Inc(decimal.NewFromFloat(5.000005))
	if count := forked.Snapshot().Count(); !count.Equal(decimal.NewFromFloat(20.654326)) {
		t.Errorf("expected forked count to be 20.654326 after Incrementing forked by 5.000005, got %s", count.String())
	}
}
