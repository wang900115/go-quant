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

func TestGaugeInfoJsonString(t *testing.T) {
	g := NewGaugeInfo()
	g.Set("key1", "value1")
	jsonString := g.Snapshot().Value().String()
	expected := `{"key1":"value1"}`
	if jsonString != expected {
		t.Errorf("Expected JSON string: %s, got: %s", expected, jsonString)
	}
}
