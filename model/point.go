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
	"errors"
	"fmt"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/model/currency"
	"github.com/wang900115/quant/model/trade"
)

type PriceTick interface {
	PricePoint | PriceInterval | OrderBook
}

type QuotesPair struct {
	ExchangeID ExchangeId
	Base       currency.CurrencySymbol
	Quote      currency.CurrencySymbol
	Category   trade.Category
}

func (qp QuotesPair) Symbol() string {
	return fmt.Sprintf("%s/%s", qp.Base, qp.Quote)
}

type PriceInterval struct {
	OpenTime         string
	CloseTime        string
	OpeningPrice     decimal.Decimal
	HighestPrice     decimal.Decimal
	LowestPrice      decimal.Decimal
	ClosingPrice     decimal.Decimal
	Volume           decimal.Decimal
	IntervalDuration time.Duration
}

func (pi PriceInterval) String() string {
	return fmt.Sprintf("OpenTime: %s, CloseTime: %s, Duration: %s, Open: %s, Close: %s, High: %s, Low: %s, Volume: %s",
		pi.OpenTime, pi.CloseTime, pi.IntervalDuration.String(),
		pi.OpeningPrice.String(), pi.ClosingPrice.String(),
		pi.HighestPrice.String(), pi.LowestPrice.String(),
		pi.Volume.String())
}

func (pi *PriceInterval) CalculateIntervalDuration() error {
	openTime, err := time.Parse(time.RFC3339, pi.OpenTime)
	if err != nil {
		return err
	}
	closeTime, err := time.Parse(time.RFC3339, pi.CloseTime)
	if err != nil {
		return err
	}
	pi.IntervalDuration = closeTime.Sub(openTime)
	return nil
}

type PricePoint struct {
	NewPrice  decimal.Decimal
	UpdatedAt time.Time
}

var (
	errInValidOrderBookType = errors.New("invalid order book type")
	errNoOrderBookEntries   = errors.New("no order book entries")
)

type OrderBook struct {
	Symbol string
	Time   time.Time
	Bids   []OrderBookBid
	Asks   []OrderBookAsk
}

type OrderBookBase struct {
	Price    decimal.Decimal
	Quantity decimal.Decimal
}

type OrderBookBid OrderBookBase
type OrderBookAsk OrderBookBase

type OrderBookEntry interface {
	OrderBookBid | OrderBookAsk
}

func ParseOrderEntries[T OrderBookEntry](entries [][]interface{}) ([]T, error) {
	result := make([]T, 0, len(entries))
	for _, entry := range entries {
		if len(entry) < 2 {
			return nil, errNoOrderBookEntries
		}
		priceStr, ok1 := entry[0].(string)
		qtyStr, ok2 := entry[1].(string)
		if !ok1 || !ok2 {
			return nil, errInValidOrderBookType
		}
		price, err := decimal.NewFromString(priceStr)
		if err != nil {
			return nil, err
		}
		quantity, err := decimal.NewFromString(qtyStr)
		if err != nil {
			return nil, err
		}
		entry := OrderBookBase{
			Price:    price,
			Quantity: quantity,
		}
		result = append(result, T(entry))
	}
	return result, nil
}

func PushToChan[T PriceTick](ch chan T, data T) {
	select {
	case ch <- data:
	default:
		// drop if channel is full
	}
}
