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

package model

import (
	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/model/currency"
	"github.com/wang900115/quant/model/trade"
)

type OrderRequest struct {
	// Symbol format: "BTC/USD" or "ETHUSDT" depending on exchange conventions
	Symbol string
	// "BUY" or "SELL"
	Side trade.Signal
	// "LIMIT", "MARKET", etc.
	Type trade.Type
	// Optional: price for LIMIT orders
	Price decimal.Decimal
	// Quantity to buy/sell
	Quantity decimal.Decimal
	// Optional: time in force policy
	TimeInForce trade.TimeInForce
	// Optional: client order ID for tracking
	ClientOrderID string
}

type OrderResult struct {
	// Unique identifier for the order
	OrderID string
	// Optional: user-defined client order ID
	ClientOrderID string
	// The trading pair symbol
	Symbol string
	// Status of the order: "NEW", "FILLED", etc.
	Status trade.Status
	// Price at which the order was placed (was executed for  orders)
	ExecutedQty decimal.Decimal
}

type OrderDetail struct {
	// Unique identifier for the order
	OrderID string
	// The trading pair symbol
	Symbol string
	// Price at which the order was placed
	Price decimal.Decimal
	// Original quantity requested
	OrigQty decimal.Decimal
	// Quantity that has been executed
	ExecutedQty decimal.Decimal
	// Current status of the order
	Status trade.Status
	// "BUY" or "SELL"
	Side trade.Signal
	// "LIMIT", "MARKET", etc.
	Type trade.Type
	// Checkout time in milliseconds since epoch
	UpdateTime int64
}

type AssetBalance struct {
	// The asset symbol, e.g., "BTC"
	Asset currency.CurrencySymbol
	// Amount of the asset that is free for trading
	Free decimal.Decimal
	// Amount of the asset that is locked in orders
	Locked decimal.Decimal
}

type OrderEvent struct {
	// Unique identifier for the order
	OrderID string
	// The trading pair symbol
	Symbol string
	// Current status of the order
	Status trade.Status
	// Price at which the order was placed
	FilledQty decimal.Decimal
	// Quantity that was last filled
	LastQty decimal.Decimal
	// "BUY" or "SELL"
	Side trade.Signal
	// "LIMIT", "MARKET", etc.
	Type       trade.Type
	UpdateTime int64
}
