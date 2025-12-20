//go:build integration
// +build integration

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

package exchange

import (
	"context"
	"testing"

	"github.com/wang900115/quant/exchange/binance"
	"github.com/wang900115/quant/exchange/coinbase"
	"github.com/wang900115/quant/exchange/okx"
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/currency"
	"github.com/wang900115/quant/model/trade"
)

func TestProvider_GetPrice(t *testing.T) {
	ps := New()
	ps.Register(model.BINANCE, binance.New(binance.BinanceConfig{}))
	ps.Register(model.COINBASE, coinbase.New(coinbase.CoinbaseConfig{}))
	ps.Register(model.OKX, okx.New(okx.OkxConfig{}))

	tests := []struct {
		name string
		pair model.QuotesPair
	}{
		{
			name: "Binance SPOT BTC-USDT",
			pair: model.QuotesPair{
				ExchangeID: model.BINANCE,
				Base:       currency.BTCSymbol,
				Quote:      currency.USDTSymbol,
				Category:   trade.SPOT,
			},
		},
		{
			name: "Coinbase SPOT BTC-USDT",
			pair: model.QuotesPair{
				ExchangeID: model.COINBASE,
				Base:       currency.BTCSymbol,
				Quote:      currency.USDTSymbol,
				Category:   trade.SPOT,
			},
		},
		{
			name: "OKX SPOT BTC-USDT",
			pair: model.QuotesPair{
				ExchangeID: model.OKX,
				Base:       currency.BTCSymbol,
				Quote:      currency.USDTSymbol,
				Category:   trade.SPOT,
			},
		},
		{
			name: "Binance FUTURES BTC-USDT",
			pair: model.QuotesPair{
				ExchangeID: model.BINANCE,
				Base:       currency.BTCSymbol,
				Quote:      currency.USDTSymbol,
				Category:   trade.FUTURES,
			},
		},
		{
			name: "OKX FUTURES ETH-USDT",
			pair: model.QuotesPair{
				ExchangeID: model.OKX,
				Base:       currency.ETHSymbol,
				Quote:      currency.USDTSymbol,
				Category:   trade.FUTURES,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, err := ps.GetPrice(context.Background(), tt.pair)
			if err != nil {
				t.Fatalf("Failed to get price for %s: %v", tt.pair.Symbol(), err)
			}

			if price == nil {
				t.Fatalf("Price is nil for %s", tt.pair.Symbol())
			}

			if price.NewPrice.IsZero() || price.NewPrice.IsNegative() {
				t.Errorf("Invalid price for %s: %v", tt.pair.Symbol(), price.NewPrice)
			}

			t.Logf("Price for %s: %s at %v", tt.pair.Symbol(), price.NewPrice.String(), price.UpdatedAt)
		})
	}
}

func TestProvider_ListProviders(t *testing.T) {
	ps := New()

	// Initially empty
	providers := ps.ListProviders()
	if len(providers) != 0 {
		t.Errorf("Expected 0 providers, got %d", len(providers))
	}

	// Register providers
	ps.Register(model.BINANCE, binance.New(binance.BinanceConfig{}))
	ps.Register(model.COINBASE, coinbase.New(coinbase.CoinbaseConfig{}))
	ps.Register(model.OKX, okx.New(okx.OkxConfig{}))

	providers = ps.ListProviders()
	if len(providers) != 3 {
		t.Errorf("Expected 3 providers, got %d", len(providers))
	}

	// Check each provider exists
	found := make(map[model.ExchangeId]bool)
	for _, p := range providers {
		found[p.ID] = true
		t.Logf("Found provider: %s", p.Name)
	}

	if !found[model.BINANCE] {
		t.Error("Binance provider not found")
	}
	if !found[model.COINBASE] {
		t.Error("Coinbase provider not found")
	}
	if !found[model.OKX] {
		t.Error("OKX provider not found")
	}
}

func TestProvider_UnregisterProvider(t *testing.T) {
	ps := New()
	ps.Register(model.BINANCE, binance.New(binance.BinanceConfig{}))
	ps.Register(model.OKX, okx.New(okx.OkxConfig{}))

	// Verify both are registered
	providers := ps.ListProviders()
	if len(providers) != 2 {
		t.Errorf("Expected 2 providers, got %d", len(providers))
	}

	// Unregister one
	ps.Unregister(model.BINANCE)

	providers = ps.ListProviders()
	if len(providers) != 1 {
		t.Errorf("Expected 1 provider after unregister, got %d", len(providers))
	}

	// Try to get price from unregistered provider
	pair := model.QuotesPair{
		ExchangeID: model.BINANCE,
		Base:       currency.BTCSymbol,
		Quote:      currency.USDTSymbol,
		Category:   trade.SPOT,
	}

	_, err := ps.GetPrice(context.Background(), pair)
	if err == nil {
		t.Error("Expected error when getting price from unregistered provider")
	}

	if err != errMissingProvider {
		t.Errorf("Expected errMissingProvider, got %v", err)
	}
}

func TestProvider_GetKlines(t *testing.T) {
	ps := New()
	ps.Register(model.BINANCE, binance.New(binance.BinanceConfig{}))
	ps.Register(model.COINBASE, coinbase.New(coinbase.CoinbaseConfig{}))
	ps.Register(model.OKX, okx.New(okx.OkxConfig{}))

	tests := []struct {
		name     string
		pair     model.QuotesPair
		interval string
		limit    int
	}{
		{
			name: "Binance SPOT 1m",
			pair: model.QuotesPair{
				ExchangeID: model.BINANCE,
				Base:       currency.BTCSymbol,
				Quote:      currency.USDTSymbol,
				Category:   trade.SPOT,
			},
			interval: "1m",
			limit:    5,
		},
		{
			name: "OKX SPOT 5m",
			pair: model.QuotesPair{
				ExchangeID: model.OKX,
				Base:       currency.ETHSymbol,
				Quote:      currency.USDTSymbol,
				Category:   trade.SPOT,
			},
			interval: "5m",
			limit:    10,
		},
		{
			name: "Coinbase SPOT 1h",
			pair: model.QuotesPair{
				ExchangeID: model.COINBASE,
				Base:       currency.BTCSymbol,
				Quote:      currency.USDTSymbol,
				Category:   trade.SPOT,
			},
			interval: "1h",
			limit:    10,
		},
		{
			name: "Binance FUTURES 15m",
			pair: model.QuotesPair{
				ExchangeID: model.BINANCE,
				Base:       currency.BTCSymbol,
				Quote:      currency.USDTSymbol,
				Category:   trade.FUTURES,
			},
			interval: "15m",
			limit:    20,
		},
		{
			name: "OKX FUTURES 1H",
			pair: model.QuotesPair{
				ExchangeID: model.OKX,
				Base:       currency.ETHSymbol,
				Quote:      currency.USDTSymbol,
				Category:   trade.FUTURES,
			},
			interval: "1H",
			limit:    10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			klines, err := ps.GetKlines(context.Background(), tt.pair, tt.interval, tt.limit)
			if err != nil {
				t.Fatalf("Failed to get klines for %s: %v", tt.pair.Symbol(), err)
			}

			if len(klines) == 0 {
				t.Fatalf("No klines returned for %s", tt.pair.Symbol())
			}

			if len(klines) > tt.limit {
				t.Errorf("Expected at most %d klines, got %d", tt.limit, len(klines))
			}

			// Validate first kline
			first := klines[0]
			if first.OpeningPrice.IsZero() || first.HighestPrice.IsZero() || first.LowestPrice.IsZero() || first.ClosingPrice.IsZero() {
				t.Errorf("Invalid OHLC values in first kline: O:%v H:%v L:%v C:%v",
					first.OpeningPrice, first.HighestPrice, first.LowestPrice, first.ClosingPrice)
			}

			if first.HighestPrice.LessThan(first.LowestPrice) {
				t.Errorf("High (%v) is less than Low (%v)", first.HighestPrice, first.LowestPrice)
			}

			t.Logf("Klines for %s [%s]: %d klines, first: O:%v H:%v L:%v C:%v V:%v",
				tt.pair.Symbol(), tt.interval, len(klines),
				first.OpeningPrice, first.HighestPrice, first.LowestPrice, first.ClosingPrice, first.Volume)
		})
	}
}

func TestProvider_GetOrderBook(t *testing.T) {
	ps := New()
	ps.Register(model.COINBASE, coinbase.New(coinbase.CoinbaseConfig{}))
	ps.Register(model.BINANCE, binance.New(binance.BinanceConfig{}))
	ps.Register(model.OKX, okx.New(okx.OkxConfig{}))

	tests := []struct {
		name  string
		pair  model.QuotesPair
		limit int
	}{
		{
			name: "Binance SPOT BTC-USDT depth 5",
			pair: model.QuotesPair{
				ExchangeID: model.BINANCE,
				Base:       currency.BTCSymbol,
				Quote:      currency.USDTSymbol,
				Category:   trade.SPOT,
			},
			limit: 5,
		},
		{
			name: "Coinbase SPOT BTC-USDT depth 10",
			pair: model.QuotesPair{
				ExchangeID: model.COINBASE,
				Base:       currency.BTCSymbol,
				Quote:      currency.USDTSymbol,
				Category:   trade.SPOT,
			},
			limit: 10,
		},
		{
			name: "OKX SPOT ETH-USDT depth 5",
			pair: model.QuotesPair{
				ExchangeID: model.OKX,
				Base:       currency.ETHSymbol,
				Quote:      currency.USDTSymbol,
				Category:   trade.SPOT,
			},
			limit: 5,
		},
		{
			name: "Binance FUTURES BTC-USDT depth 5",
			pair: model.QuotesPair{
				ExchangeID: model.BINANCE,
				Base:       currency.BTCSymbol,
				Quote:      currency.USDTSymbol,
				Category:   trade.FUTURES,
			},
			limit: 5,
		},
		{
			name: "OKX FUTURES ETH-USDT depth 10",
			pair: model.QuotesPair{
				ExchangeID: model.OKX,
				Base:       currency.ETHSymbol,
				Quote:      currency.USDTSymbol,
				Category:   trade.FUTURES,
			},
			limit: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orderBook, err := ps.GetOrderBook(context.Background(), tt.pair, tt.limit)
			if err != nil {
				t.Fatalf("Failed to get order book for %s: %v", tt.pair.Symbol(), err)
			}

			if orderBook == nil {
				t.Fatalf("Order book is nil for %s", tt.pair.Symbol())
			}

			if len(orderBook.Bids) == 0 {
				t.Errorf("No bids in order book for %s", tt.pair.Symbol())
			}

			if len(orderBook.Asks) == 0 {
				t.Errorf("No asks in order book for %s", tt.pair.Symbol())
			}

			// Validate bids are sorted (highest to lowest)
			for i := 1; i < len(orderBook.Bids); i++ {
				if orderBook.Bids[i-1].Price.LessThan(orderBook.Bids[i].Price) {
					t.Errorf("Bids not properly sorted at index %d: %v < %v",
						i, orderBook.Bids[i-1].Price, orderBook.Bids[i].Price)
				}
			}

			// Validate asks are sorted (lowest to highest)
			for i := 1; i < len(orderBook.Asks); i++ {
				if orderBook.Asks[i-1].Price.GreaterThan(orderBook.Asks[i].Price) {
					t.Errorf("Asks not properly sorted at index %d: %v > %v",
						i, orderBook.Asks[i-1].Price, orderBook.Asks[i].Price)
				}
			}

			// Best bid should be less than best ask
			if len(orderBook.Bids) > 0 && len(orderBook.Asks) > 0 {
				bestBid := orderBook.Bids[0].Price
				bestAsk := orderBook.Asks[0].Price
				if bestBid.GreaterThanOrEqual(bestAsk) {
					t.Errorf("Best bid (%v) >= Best ask (%v)", bestBid, bestAsk)
				}

				t.Logf("Order Book for %s: %d bids, %d asks, Best Bid: %v @ %v, Best Ask: %v @ %v",
					tt.pair.Symbol(), len(orderBook.Bids), len(orderBook.Asks),
					orderBook.Bids[0].Quantity, bestBid, orderBook.Asks[0].Quantity, bestAsk)
			}
		})
	}
}
