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

import "testing"

func TestCounterInt64Clear(t *testing.T) {
	c := NewCounterInt64()
	c.Inc(10)
	c.Clear()
	if count := c.Snapshot().Count(); count != 0 {
		t.Errorf("expected count to be 0 after Clear, got %d", count)
	}
}

func TestCounter(t *testing.T) {
	c := NewCounterInt64()
	if count := c.Snapshot().Count(); count != 0 {
		t.Errorf("expected initial count to be 0, got %d", count)
	}
	c.Dec(1)
	if count := c.Snapshot().Count(); count != -1 {
		t.Errorf("expected count to be -1 after Decrementing by 1, got %d", count)
	}
	c.Inc(5)
	if count := c.Snapshot().Count(); count != 4 {
		t.Errorf("expected count to be 4 after Incrementing by 5, got %d", count)
	}
	c.Dec(2)
	if count := c.Snapshot().Count(); count != 2 {
		t.Errorf("expected count to be 2 after Decrementing by 2, got %d", count)
	}
	c.Inc(3)
	if count := c.Snapshot().Count(); count != 5 {
		t.Errorf("expected count to be 5 after Incrementing by 3, got %d", count)
	}
}

func TestCounterSnapshot(t *testing.T) {
	c := NewCounterInt64()
	c.Inc(7)
	snapshot := c.Snapshot()
	c.Inc(1)
	if snapshot.Count() != 7 {
		t.Errorf("expected snapshot count to be 7, got %d", snapshot.Count())
	}
	if c.Snapshot().Count() != 8 {
		t.Errorf("expected current count to be 8 after Incrementing by 1, got %d", c.Snapshot().Count())
	}
}

func TestCounterFork(t *testing.T) {
	c := NewCounterInt64()
	c.Inc(15)
	forked := c.Fork()
	if count := forked.Snapshot().Count(); count != 15 {
		t.Errorf("expected forked count to be 15, got %d", count)
	}
	forked.Inc(5)
	if count := forked.Snapshot().Count(); count != 20 {
		t.Errorf("expected forked count to be 20 after Incrementing by 5, got %d", count)
	}
}
