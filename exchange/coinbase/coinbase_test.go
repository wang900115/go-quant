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

package coinbase

import (
	"context"
	"testing"
	"time"

	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/currency"
	"github.com/wang900115/quant/model/trade"
)

// TestCoinbaseSingleClient_GetPrice tests fetching current price from Coinbase
func TestCoinbaseSingleClient_GetPrice(t *testing.T) {
	cfg := CoinbaseConfig{
		IstestNet:     false,
		PublicTimeout: 10 * time.Second,
	}
	client := NewSingleClient(cfg)

	testCases := []struct {
		name    string
		pair    model.QuotesPair
		wantErr bool
	}{
		{
			name: "SPOT BTC-USD",
			pair: model.QuotesPair{
				Base:     currency.BTCSymbol,
				Quote:    currency.USDSymbol,
				Category: trade.SPOT,
			},
			wantErr: false,
		},
		{
			name: "SPOT ETH-USD",
			pair: model.QuotesPair{
				Base:     currency.ETHSymbol,
				Quote:    currency.USDSymbol,
				Category: trade.SPOT,
			},
			wantErr: false,
		},
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
			name: "SPOT ETH-BTC",
			pair: model.QuotesPair{
				Base:     currency.ETHSymbol,
				Quote:    currency.BTCSymbol,
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

// TestCoinbaseSingleClient_GetKlines tests fetching kline data
func TestCoinbaseSingleClient_GetKlines(t *testing.T) {
	cfg := CoinbaseConfig{
		IstestNet:     false,
		PublicTimeout: 10 * time.Second,
	}
	client := NewSingleClient(cfg)

	testCases := []struct {
		name        string
		pair        model.QuotesPair
		granularity string
		limit       int
		wantErr     bool
	}{
		{
			name: "3600 second granularity BTC-USD",
			pair: model.QuotesPair{
				Base:     currency.BTCSymbol,
				Quote:    currency.USDSymbol,
				Category: trade.SPOT,
			},
			granularity: "3600", // 1 hour
			limit:       10,
			wantErr:     false,
		},
		{
			name: "900 second granularity ETH-USD",
			pair: model.QuotesPair{
				Base:     currency.ETHSymbol,
				Quote:    currency.USDSymbol,
				Category: trade.SPOT,
			},
			granularity: "900", // 15 minutes
			limit:       5,
			wantErr:     false,
		},
		{
			name: "86400 second granularity BTC-USD",
			pair: model.QuotesPair{
				Base:     currency.BTCSymbol,
				Quote:    currency.USDSymbol,
				Category: trade.SPOT,
			},
			granularity: "86400", // 1 day
			limit:       7,
			wantErr:     false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			intervals, err := client.GetKlines(ctx, tc.pair, tc.granularity, tc.limit)

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

			// Coinbase may return different number of candles
			t.Logf("Received %d intervals (requested %d)", len(intervals), tc.limit)

			// Verify first interval structure
			if len(intervals) > 0 {
				first := intervals[0]
				if first.OpeningPrice.IsZero() || first.HighestPrice.IsZero() || first.LowestPrice.IsZero() || first.ClosingPrice.IsZero() {
					t.Error("expected non-zero OHLC values")
				}

				t.Logf("First kline: O:%s H:%s L:%s C:%s V:%s",
					first.OpeningPrice.String(),
					first.HighestPrice.String(),
					first.LowestPrice.String(),
					first.ClosingPrice.String(),
					first.Volume.String(),
				)
			}
		})
	}
}

// TestCoinbaseSingleClient_GetOrderBook tests fetching order book data
func TestCoinbaseSingleClient_GetOrderBook(t *testing.T) {
	cfg := CoinbaseConfig{
		IstestNet:     false,
		PublicTimeout: 10 * time.Second,
	}
	client := NewSingleClient(cfg)

	testCases := []struct {
		name    string
		pair    model.QuotesPair
		level   int
		wantErr bool
	}{
		{
			name: "BTC-USD Level 1",
			pair: model.QuotesPair{
				Base:     currency.BTCSymbol,
				Quote:    currency.USDSymbol,
				Category: trade.SPOT,
			},
			level:   1,
			wantErr: false,
		},
		{
			name: "ETH-USD Level 2",
			pair: model.QuotesPair{
				Base:     currency.ETHSymbol,
				Quote:    currency.USDSymbol,
				Category: trade.SPOT,
			},
			level:   2,
			wantErr: false,
		},
		{
			name: "BTC-USDT Level 2",
			pair: model.QuotesPair{
				Base:     currency.BTCSymbol,
				Quote:    currency.USDTSymbol,
				Category: trade.SPOT,
			},
			level:   2,
			wantErr: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			orderBook, err := client.GetOrderBook(ctx, tc.pair, tc.level)

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

// TestCoinbaseStreamClient_Connection tests WebSocket connection
func TestCoinbaseStreamClient_Connection(t *testing.T) {
	cfg := CoinbaseConfig{
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

	if client.client == nil {
		t.Error("client connection is nil")
	}

	t.Log("Successfully connected to Coinbase WebSocket")
}

// TestCoinbaseStreamClient_Subscribe tests subscribing to streams
func TestCoinbaseStreamClient_Subscribe(t *testing.T) {
	cfg := CoinbaseConfig{
		IstestNet:  false,
		BufferSize: 100,
	}

	client, err := NewStreamClient(cfg)
	if err != nil {
		t.Fatalf("failed to create stream client: %v", err)
	}
	pair := model.QuotesPair{
		Base:     currency.BTCSymbol,
		Quote:    currency.USDSymbol,
		Category: trade.SPOT,
	}

	testCases := []struct {
		name         string
		channelTypes []string
	}{
		{
			name:         "Ticker channel",
			channelTypes: []string{"ticker"},
		},
		{
			name:         "Matches channel",
			channelTypes: []string{"matches"},
		},
		{
			name:         "Level2 channel",
			channelTypes: []string{"level2"},
		},
		{
			name:         "Multiple channels",
			channelTypes: []string{"ticker", "matches"},
		},
	}
	defer client.Close()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := client.SubscribeStream(pair, tc.channelTypes)
			if err != nil {
				t.Fatalf("failed to subscribe: %v", err)
			}
			t.Logf("Successfully subscribed to %v", tc.channelTypes)
		})
	}
}

// TestCoinbaseStreamClient_ReceiveData tests receiving stream data
func TestCoinbaseStreamClient_ReceiveData(t *testing.T) {
	cfg := CoinbaseConfig{
		IstestNet:  false,
		BufferSize: 100,
	}

	client, err := NewStreamClient(cfg)
	if err != nil {
		t.Fatalf("failed to create stream client: %v", err)
	}

	pair := model.QuotesPair{
		Base:     currency.BTCSymbol,
		Quote:    currency.USDSymbol,
		Category: trade.SPOT,
	}

	defer client.Close()

	// Subscribe to ticker channel
	err = client.SubscribeStream(pair, []string{"ticker"})
	if err != nil {
		t.Fatalf("failed to subscribe: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	priceChan, _, _ := client.ReceiveStream()

	// Start dispatcher
	go func() {
		if err := client.Dispatch(ctx); err != nil && err != context.DeadlineExceeded {
			t.Logf("dispatcher error: %v", err)
		}
	}()

	// Collect price updates
	receivedCount := 0
	maxUpdates := 5

	for receivedCount < maxUpdates {
		select {
		case price := <-priceChan:
			receivedCount++
			t.Logf("Update #%d - Price: %s at %v", receivedCount, price.NewPrice.String(), price.UpdatedAt)
		case <-ctx.Done():
			t.Fatalf("timeout waiting for updates, received %d/%d", receivedCount, maxUpdates)
		}
	}

	if receivedCount < maxUpdates {
		t.Errorf("expected at least %d updates, got %d", maxUpdates, receivedCount)
	}
}

// TestCoinbaseStreamClient_MultipleChannels tests subscribing to multiple channel types
func TestCoinbaseStreamClient_MultipleChannels(t *testing.T) {
	cfg := CoinbaseConfig{
		IstestNet:  false,
		BufferSize: 100,
	}

	client, err := NewStreamClient(cfg)
	if err != nil {
		t.Fatalf("failed to create stream client: %v", err)
	}

	pair := model.QuotesPair{
		Base:     currency.BTCSymbol,
		Quote:    currency.USDSymbol,
		Category: trade.SPOT,
	}

	channels := []string{"ticker", "matches", "level2"}

	err = client.SubscribeStream(pair, channels)
	if err != nil {
		t.Fatalf("failed to subscribe to multiple channels: %v", err)
	}

	t.Logf("Successfully subscribed to %d channels: %v", len(channels), channels)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	priceChan, _, orderBookChan := client.ReceiveStream()

	// Start dispatcher
	go func() {
		if err := client.Dispatch(ctx); err != nil && err != context.DeadlineExceeded {
			t.Logf("dispatcher error: %v", err)
		}
	}()

	// Collect data from different channels
	priceCount := 0
	orderBookCount := 0
	maxEach := 2

	for priceCount < maxEach || orderBookCount < maxEach {
		select {
		case price := <-priceChan:
			if priceCount < maxEach {
				priceCount++
				t.Logf("Ticker update #%d: %s", priceCount, price.NewPrice.String())
			}
		case ob := <-orderBookChan:
			if orderBookCount < maxEach {
				orderBookCount++
				t.Logf("OrderBook update #%d: %d bids, %d asks", orderBookCount, len(ob.Bids), len(ob.Asks))
			}
		case <-ctx.Done():
			t.Logf("Timeout - received %d price updates and %d orderbook updates", priceCount, orderBookCount)
			return
		}
	}

	t.Logf("Successfully received data from multiple channels")
}

// TestCoinbaseConfig_DefaultValues tests configuration with default values
func TestCoinbaseConfig_DefaultValues(t *testing.T) {
	cfg := CoinbaseConfig{
		IstestNet: false,
	}

	client := NewSingleClient(cfg)
	if client == nil {
		t.Fatal("expected client, got nil")
	}

	if client.client.Timeout != defaultTimeout {
		t.Errorf("expected default timeout %v, got %v", defaultTimeout, client.client.Timeout)
	}

	t.Log("Default configuration values applied correctly")
}

// TestCoinbaseConfig_CustomValues tests configuration with custom values
func TestCoinbaseConfig_CustomValues(t *testing.T) {
	customTimeout := 5 * time.Second
	customBufferSize := 50

	cfg := CoinbaseConfig{
		IstestNet:     false,
		PublicTimeout: customTimeout,
		BufferSize:    customBufferSize,
	}

	client := NewSingleClient(cfg)
	if client == nil {
		t.Fatal("expected client, got nil")
	}

	if client.client.Timeout != customTimeout {
		t.Errorf("expected custom timeout %v, got %v", customTimeout, client.client.Timeout)
	}

	streamClient, err := NewStreamClient(cfg)
	if err != nil {
		t.Fatalf("failed to create stream client: %v", err)
	}

	if streamClient.bufferSize != customBufferSize {
		t.Errorf("expected custom buffer size %d, got %d", customBufferSize, streamClient.bufferSize)
	}

	t.Log("Custom configuration values applied correctly")
}
