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

package trade

type Category string

const (
	SPOT    Category = "SPOT"
	FUTURES Category = "FUTURES"
	INVERSE Category = "INVERSE"
)

func (c Category) String() string {
	return string(c)
}

type Signal string

const (
	BUY  Signal = "BUY"
	SELL Signal = "SELL"
)

type Type string

const (
	LIMIT  Type = "LIMIT"
	MARKET Type = "MARKET"
)

type Status string

const (
	NEW              Status = "NEW"
	FILLED           Status = "FILLED"
	CANCELED         Status = "CANCELED"
	PARTIALLY_FILLED Status = "PARTIALLY_FILLED"
	UNKNOWN          Status = "UNKNOWN"
)

type TimeInForce string

const (
	GTC TimeInForce = "GTC" // Good Till Cancelled
	IOC TimeInForce = "IOC" // Immediate Or Cancel
	FOK TimeInForce = "FOK" // Fill Or Kill
)
