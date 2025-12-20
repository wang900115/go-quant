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

package binance

import (
	"context"
	"testing"
	"time"

	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/currency"
	"github.com/wang900115/quant/model/trade"
)

// TestBinanceSingleClient_GetPrice tests fetching current price from Binance
func TestBinanceSingleClient_GetPrice(t *testing.T) {
	cfg := BinanceConfig{
		IstestNet:     false,
		PublicTimeout: 10 * time.Second,
	}
	client := NewSingleClient(cfg)

	testCases := []struct {
		name     string
		pair     model.QuotesPair
		wantErr  bool
		errCheck func(error) bool
	}{
		{
			name: "SPOT BTC-USDT",
			pair: model.QuotesPair{
				Base:     currency.BTCSymbol,
				Quote:    currency.USDTSymbol,
				Category: trade.SPOT,
			},
			wantErr: false,
		},
		{
			name: "FUTURES BTC-USDT",
			pair: model.QuotesPair{
				Base:     currency.BTCSymbol,
				Quote:    currency.USDTSymbol,
				Category: trade.FUTURES,
			},
			wantErr: false,
		},
		{
			name: "SPOT ETH-USDT",
			pair: model.QuotesPair{
				Base:     currency.ETHSymbol,
				Quote:    currency.USDTSymbol,
				Category: trade.SPOT,
			},
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			price, err := client.GetPrice(ctx, tc.pair)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if price == nil {
				t.Fatal("expected price point, got nil")
			}

			if price.NewPrice.IsZero() {
				t.Error("expected non-zero price")
			}

			t.Logf("%s Price: %s at %v", tc.name, price.NewPrice.String(), price.UpdatedAt)
		})
	}
}

// TestBinanceSingleClient_GetKlines tests fetching kline data
func TestBinanceSingleClient_GetKlines(t *testing.T) {
	cfg := BinanceConfig{
		IstestNet:     false,
		PublicTimeout: 10 * time.Second,
	}
	client := NewSingleClient(cfg)

	testCases := []struct {
		name     string
		pair     model.QuotesPair
		interval string
		limit    int
		wantErr  bool
	}{
		{
			name: "SPOT 1h BTC-USDT",
			pair: model.QuotesPair{
				Base:     currency.BTCSymbol,
				Quote:    currency.USDTSymbol,
				Category: trade.SPOT,
			},
			interval: "1h",
			limit:    10,
			wantErr:  false,
		},
		{
			name: "FUTURES 15m ETH-USDT",
			pair: model.QuotesPair{
				Base:     currency.ETHSymbol,
				Quote:    currency.USDTSymbol,
				Category: trade.FUTURES,
			},
			interval: "15m",
			limit:    5,
			wantErr:  false,
		},
		{
			name: "SPOT 1d BTC-USDT",
			pair: model.QuotesPair{
				Base:     currency.BTCSymbol,
				Quote:    currency.USDTSymbol,
				Category: trade.SPOT,
			},
			interval: "1d",
			limit:    7,
			wantErr:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			intervals, err := client.GetKlines(ctx, tc.pair, tc.interval, tc.limit)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(intervals) == 0 {
				t.Fatal("expected intervals, got empty slice")
			}

			if len(intervals) > tc.limit {
				t.Errorf("expected max %d intervals, got %d", tc.limit, len(intervals))
			}

			// Verify first interval structure
			first := intervals[0]
			if first.OpeningPrice.IsZero() || first.HighestPrice.IsZero() || first.LowestPrice.IsZero() || first.ClosingPrice.IsZero() {
				t.Error("expected non-zero OHLC values")
			}

			t.Logf("Received %d klines, first: O:%s H:%s L:%s C:%s V:%s",
				len(intervals),
				first.OpeningPrice.String(),
				first.HighestPrice.String(),
				first.LowestPrice.String(),
				first.ClosingPrice.String(),
				first.Volume.String(),
			)
		})
	}
}

// TestBinanceSingleClient_GetOrderBook tests fetching order book data
func TestBinanceSingleClient_GetOrderBook(t *testing.T) {
	cfg := BinanceConfig{
		IstestNet:     false,
		PublicTimeout: 10 * time.Second,
	}
	client := NewSingleClient(cfg)

	testCases := []struct {
		name    string
		pair    model.QuotesPair
		limit   int
		wantErr bool
	}{
		{
			name: "SPOT BTC-USDT depth 5",
			pair: model.QuotesPair{
				Base:     currency.BTCSymbol,
				Quote:    currency.USDTSymbol,
				Category: trade.SPOT,
			},
			limit:   5,
			wantErr: false,
		},
		{
			name: "FUTURES ETH-USDT depth 10",
			pair: model.QuotesPair{
				Base:     currency.ETHSymbol,
				Quote:    currency.USDTSymbol,
				Category: trade.FUTURES,
			},
			limit:   10,
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			orderBook, err := client.GetOrderBook(ctx, tc.pair, tc.limit)

			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if orderBook == nil {
				t.Fatal("expected order book, got nil")
			}

			if len(orderBook.Bids) == 0 {
				t.Error("expected bids, got empty")
			}

			if len(orderBook.Asks) == 0 {
				t.Error("expected asks, got empty")
			}

			t.Logf("Order Book - Symbol: %s, Bids: %d, Asks: %d",
				orderBook.Symbol,
				len(orderBook.Bids),
				len(orderBook.Asks),
			)

			if len(orderBook.Bids) > 0 {
				t.Logf("Best Bid: %s @ %s", orderBook.Bids[0].Quantity.String(), orderBook.Bids[0].Price.String())
			}
			if len(orderBook.Asks) > 0 {
				t.Logf("Best Ask: %s @ %s", orderBook.Asks[0].Quantity.String(), orderBook.Asks[0].Price.String())
			}
		})
	}
}

// TestBinanceStreamClient_Connection tests WebSocket connection
func TestBinanceStreamClient_Connection(t *testing.T) {
	cfg := BinanceConfig{
		IstestNet:  false,
		BufferSize: 100,
		Callback: func(message []byte) error {
			t.Logf("Received message: %s", string(message))
			return nil
		},
	}

	client, err := NewStreamClient(cfg)
	if err != nil {
		t.Fatalf("failed to create stream client: %v", err)
	}
	defer client.Close()
	if client.spotClient == nil {
		t.Error("spot client is nil")
	}
	if client.futuresClient == nil {
		t.Error("futures client is nil")
	}
	if client.inverseClient == nil {
		t.Error("inverse client is nil")
	}

	t.Log("Successfully connected to all Binance WebSocket endpoints")
}

// TestBinanceStreamClient_Subscribe tests subscribing to streams
func TestBinanceStreamClient_Subscribe(t *testing.T) {
	cfg := BinanceConfig{
		IstestNet:  false,
		BufferSize: 100,
	}

	client, err := NewStreamClient(cfg)
	if err != nil {
		t.Fatalf("failed to create stream client: %v", err)
	}
	defer client.Close()
	pair := model.QuotesPair{
		Base:     currency.BTCSymbol,
		Quote:    currency.USDTSymbol,
		Category: trade.SPOT,
	}

	testCases := []struct {
		name        string
		streamTypes []string
	}{
		{
			name:        "Ticker stream",
			streamTypes: []string{"ticker"},
		},
		{
			name:        "Trade stream",
			streamTypes: []string{"trade"},
		},
		{
			name:        "Multiple streams",
			streamTypes: []string{"ticker", "trade", "depth"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.SubscribeStream(pair, tc.streamTypes)
			if err != nil {
				t.Fatalf("failed to subscribe: %v", err)
			}
			t.Logf("Successfully subscribed to %v", tc.streamTypes)
		})
	}
}

// TestBinanceStreamClient_ReceiveData tests receiving stream data
func TestBinanceStreamClient_ReceiveData(t *testing.T) {
	cfg := BinanceConfig{
		IstestNet:  false,
		BufferSize: 100,
	}
	client, err := NewStreamClient(cfg)
	if err != nil {
		t.Fatalf("failed to create stream client: %v", err)
	}
	defer client.Close()

	pair := model.QuotesPair{
		Base:     currency.BTCSymbol,
		Quote:    currency.USDTSymbol,
		Category: trade.SPOT,
	}

	err = client.SubscribeStream(pair, []string{"ticker"})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	priceChan, _, _ := client.ReceiveStream()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start dispatcher
	go func() {
		if err := client.Dispatch(ctx); err != nil && err != context.DeadlineExceeded {
			t.Logf("dispatcher error: %v", err)
		}
	}()

	receivedCount := 0
	maxUpdates := 2

	for receivedCount < maxUpdates {
		select {
		case price := <-priceChan:
			receivedCount++
			t.Logf("Update #%d - Price: %s at %v", receivedCount, price.NewPrice.String(), price.UpdatedAt)
		case <-ctx.Done():
			t.Fatalf("timeout waiting for updates, received %d/%d", receivedCount, maxUpdates)
		}
	}
}

// TestBinanceStreamClient_MultipleCategories tests streaming from different categories
func TestBinanceStreamClient_MultipleCategories(t *testing.T) {
	cfg := BinanceConfig{
		IstestNet:  false,
		BufferSize: 100,
	}

	client, err := NewStreamClient(cfg)
	if err != nil {
		t.Fatalf("failed to create stream client: %v", err)
	}

	testCases := []struct {
		name string
		pair model.QuotesPair
	}{
		{
			name: "SPOT",
			pair: model.QuotesPair{
				Base:     currency.BTCSymbol,
				Quote:    currency.USDTSymbol,
				Category: trade.SPOT,
			},
		},
		{
			name: "FUTURES",
			pair: model.QuotesPair{
				Base:     currency.BTCSymbol,
				Quote:    currency.USDTSymbol,
				Category: trade.FUTURES,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.SubscribeStream(tc.pair, []string{"ticker"})
			if err != nil {
				t.Fatalf("failed to subscribe to %s: %v", tc.name, err)
			}
			t.Logf("Successfully subscribed to %s category", tc.name)
		})
	}
}
