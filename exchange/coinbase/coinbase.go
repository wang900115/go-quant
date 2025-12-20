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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/common/parse"
	"github.com/wang900115/quant/model"
)

const (
	spotEndpoint       = "https://api.exchange.coinbase.com"
	spotWsEndpoint     = "wss://ws-feed.exchange.coinbase.com"
	spotWsPEndpoint    = "wss://advanced-trade-ws.coinbase.com"
	testSpotEndpoint   = "https://api-public.sandbox.exchange.coinbase.com"
	testSpotWsEndpoint = "wss://ws-feed-public.sandbox.exchange.coinbase.com"
)

var defaultCallback = func(message []byte) error {
	return nil
}

const (
	defaultTimeout    = 10 * time.Second
	defaultBufferSize = 100
)

var (
	errCoinbaseNoData = errors.New("coinbase: no data returned")
	errResponseFailed = errors.New("coinbase: response failed")
	errNotValidType   = errors.New("coinbase: not valid type")
	errInitFailed     = errors.New("coinbase: initialization failed")
	errNonAssetFound  = errors.New("coinbase: no such asset found")
)

type CoinbaseSingleClient struct {
	client *http.Client
}

func NewSingleClient(cfg CoinbaseConfig) *CoinbaseSingleClient {
	timeout := cfg.PublicTimeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	return &CoinbaseSingleClient{
		client: &http.Client{Timeout: timeout},
	}
}

func (cc *CoinbaseSingleClient) GetPrice(ctx context.Context, pair model.QuotesPair) (*model.PricePoint, error) {
	symbol := fmt.Sprintf("%s-%s", pair.Base, pair.Quote)
	url := fmt.Sprintf("%s/products/%s/ticker", spotEndpoint, symbol)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := cc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errResponseFailed
	}
	var raw struct {
		Price string `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if len(raw.Price) == 0 {
		return nil, errCoinbaseNoData
	}
	price, err := decimal.NewFromString(raw.Price)
	if err != nil {
		return nil, err
	}
	data := &model.PricePoint{
		NewPrice:  price,
		UpdatedAt: time.Now(),
	}
	return data, nil
}

func (cc *CoinbaseSingleClient) GetKlines(ctx context.Context, pair model.QuotesPair, granularity string, limit int) ([]model.PriceInterval, error) {
	symbol := fmt.Sprintf("%s-%s", pair.Base, pair.Quote)

	granularityInt := int(parse.ParseInterval(granularity).Seconds())
	if !validInterval(granularityInt) {
		return nil, errNotValidType
	}
	endTime := time.Now()
	startTime := endTime.Add(-time.Duration(int64(limit)*int64(granularityInt)) * time.Second)

	url := fmt.Sprintf("%s/products/%s/candles?granularity=%d&start=%s&end=%s",
		spotEndpoint, symbol, granularityInt,
		startTime.Format(time.RFC3339),
		endTime.Format(time.RFC3339))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := cc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errResponseFailed
	}
	// Coinbase response: [[time, low, high, open, close, volume], ...]
	var rawCandles [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawCandles); err != nil {
		return nil, fmt.Errorf("failed to decode candles: %w", err)
	}

	intervals := make([]model.PriceInterval, 0, limit)

	for _, candle := range rawCandles {
		if len(candle) < 6 {
			return nil, errCoinbaseNoData
		}

		// Coinbase response: [time, low, high, open, close, volume]
		// time is Unix timestamp (seconds)
		var timestamp int64
		switch t := candle[0].(type) {
		case float64:
			timestamp = int64(t)
		case int64:
			timestamp = t
		default:
			return nil, errNotValidType
		}
		openTime := time.Unix(timestamp, 0)

		lowPrice, err := parseDecimalFromInterface(candle[1])
		if err != nil {
			return nil, err
		}
		highPrice, err := parseDecimalFromInterface(candle[2])
		if err != nil {
			return nil, err
		}
		openPrice, err := parseDecimalFromInterface(candle[3])
		if err != nil {
			return nil, err
		}
		closePrice, err := parseDecimalFromInterface(candle[4])
		if err != nil {
			return nil, err
		}
		volume, err := parseDecimalFromInterface(candle[5])
		if err != nil {
			return nil, err
		}

		duration := time.Duration(granularityInt) * time.Second
		closeTime := openTime.Add(duration)
		intervals = append(intervals, model.PriceInterval{
			OpenTime:         openTime.Format(time.RFC3339),
			OpeningPrice:     openPrice,
			HighestPrice:     highPrice,
			LowestPrice:      lowPrice,
			ClosingPrice:     closePrice,
			Volume:           volume,
			CloseTime:        closeTime.Format(time.RFC3339),
			IntervalDuration: duration,
		})
	}

	return intervals, nil
}

func (cc *CoinbaseSingleClient) GetOrderBook(ctx context.Context, pair model.QuotesPair, limit int) (*model.OrderBook, error) {
	symbol := fmt.Sprintf("%s-%s", pair.Base, pair.Quote)
	url := fmt.Sprintf("%s/products/%s/book?level=2", spotEndpoint, symbol)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := cc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errResponseFailed
	}

	var raw struct {
		Bids [][]interface{} `json:"bids"`
		Asks [][]interface{} `json:"asks"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	bids, err := model.ParseOrderEntries[model.OrderBookBid](raw.Bids)
	if err != nil {
		return nil, err
	}
	asks, err := model.ParseOrderEntries[model.OrderBookAsk](raw.Asks)
	if err != nil {
		return nil, err
	}

	return &model.OrderBook{
		Symbol: pair.Symbol(),
		Time:   time.Now(),
		Bids:   bids,
		Asks:   asks,
	}, nil
}

type CoinbaseConfig struct {
	IstestNet      bool
	PublicTimeout  time.Duration
	PrivateTimeout time.Duration
	BufferSize     int
	Callback       func(message []byte) error

	APIKey     string
	SecretKey  string
	Passphrase string

	RetryInterval       time.Duration
	HealthCheckInterval time.Duration
}

type CoinbaseClient struct {
	*CoinbaseSingleClient
	*CoinbaseStreamClient
	*CoinbaseTradeClient
}

func New(config CoinbaseConfig) *CoinbaseClient {
	streamClient, err := NewStreamClient(config)
	if err != nil {
		panic(err)
	}
	tradeClient, err := NewTradeClient(config)
	if err != nil {
		panic(err)
	}
	return &CoinbaseClient{
		CoinbaseSingleClient: NewSingleClient(config),
		CoinbaseStreamClient: streamClient,
		CoinbaseTradeClient:  tradeClient,
	}
}

func getMessageType(msg []byte) string {
	var raw struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(msg, &raw); err != nil {
		return ""
	}
	return raw.Type
}

func validInterval(granularity int) bool {
	switch granularity {
	case 60, 300, 900, 3600, 21600, 86400:
		return true
	default:
		return false
	}
}
