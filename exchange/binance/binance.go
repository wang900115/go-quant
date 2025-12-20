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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/trade"
)

const (
	spotEndpoint          = "https://api.binance.com"
	futuresEndpoint       = "https://fapi.binance.com"
	inverseEndpoint       = "https://dapi.binance.com"
	spotWsEndpoint        = "wss://stream.binance.com:9443/ws"
	spotTestWsEndpoint    = "wss://testnet.binance.vision/ws"
	futuresWsEndpoint     = "wss://fstream.binance.com/ws"
	futuresTestWsEndpoint = "wss://testnet.binancefuture.com/ws"
	inverseWsEndpoint     = "wss://dstream.binance.com/ws"
	inverseTestWsEndpoint = "wss://testnet.binancefuture.com/dws"
)

var defaultCallback = func(message []byte) error {
	return nil
}

const (
	defaultTimeout      = 10 * time.Second
	defaultBufferSize   = 100
	defaultTradeTimeout = 15 * time.Second
)

var (
	errBinanceNoData  = errors.New("binance: no data returned")
	errResponseFailed = errors.New("binance: response failed")
	errInvalidPair    = errors.New("binance: invalid trading pair")
	errInitFailed     = errors.New("binance: initialization failed")
	errNonAssetFound  = errors.New("binance: no such asset found")
)

type BinanceConfig struct {
	IstestNet      bool
	PublicTimeout  time.Duration
	PrivateTimeout time.Duration
	BufferSize     int
	Callback       func(message []byte) error

	APIKey    string
	SecretKey string

	RetryInterval       time.Duration
	HealthCheckInterval time.Duration
}

type BinanceSingleClient struct {
	client *http.Client
}

func NewSingleClient(cfg BinanceConfig) *BinanceSingleClient {
	timeout := cfg.PublicTimeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	return &BinanceSingleClient{
		client: &http.Client{Timeout: timeout},
	}
}

func (bc *BinanceSingleClient) GetPrice(ctx context.Context, pair model.QuotesPair) (*model.PricePoint, error) {
	url, err := decideRoute(pair)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := bc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errResponseFailed
	}
	var raw struct {
		Symbol string `json:"symbol"`
		Price  string `json:"price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if len(raw.Price) == 0 {
		return nil, errBinanceNoData
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

func (bc *BinanceSingleClient) GetKlines(ctx context.Context, pair model.QuotesPair, interval string, limit int) ([]model.PriceInterval, error) {
	url, err := decideRouteKlines(pair, interval, limit)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := bc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errResponseFailed
	}
	// Binance response struct: [[openTime, open, high, low, close, volume, closeTime, ...], ...]
	var rawKlines [][]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&rawKlines); err != nil {
		return nil, fmt.Errorf("failed to decode klines: %w", err)
	}
	intervals := make([]model.PriceInterval, 0, limit)
	for _, kline := range rawKlines {
		if len(kline) < 6 {
			return nil, errBinanceNoData
		}
		openTimestamp := int64(kline[0].(float64))
		openTime := time.UnixMilli(openTimestamp)
		openPrice, err := decimal.NewFromString(kline[1].(string))
		if err != nil {
			return nil, err
		}
		highPrice, err := decimal.NewFromString(kline[2].(string))
		if err != nil {
			return nil, err
		}
		lowPrice, err := decimal.NewFromString(kline[3].(string))
		if err != nil {
			return nil, err
		}
		closePrice, err := decimal.NewFromString(kline[4].(string))
		if err != nil {
			return nil, err
		}
		volume, err := decimal.NewFromString(kline[5].(string))
		if err != nil {
			return nil, err
		}
		closeTimestamp := int64(kline[6].(float64))
		closeTime := time.UnixMilli(closeTimestamp)
		duration := time.Duration(closeTimestamp-openTimestamp) * time.Millisecond
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

func (bc *BinanceSingleClient) GetOrderBook(ctx context.Context, pair model.QuotesPair, limit int) (*model.OrderBook, error) {
	url, err := decideOrderBookRoute(pair, limit)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := bc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errResponseFailed
	}
	var raw struct {
		LastUpdateID int             `json:"lastUpdateId"`
		Bids         [][]interface{} `json:"bids"`
		Asks         [][]interface{} `json:"asks"`
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
	orderBook := &model.OrderBook{
		Symbol: pair.Symbol(),
		Time:   time.Now(),
		Bids:   bids,
		Asks:   asks,
	}
	return orderBook, nil
}

type BinanceClient struct {
	*BinanceSingleClient
	*BinanceStreamClient
	*BinanceTradeClient
}

func New(cfg BinanceConfig) *BinanceClient {
	binanceStreamClient, err := NewStreamClient(cfg)
	if err != nil {
		panic(err)
	}
	binanceTradeClient, err := NewTradeClient(cfg)
	if err != nil {
		panic(err)
	}
	return &BinanceClient{
		BinanceSingleClient: NewSingleClient(cfg),
		BinanceStreamClient: binanceStreamClient,
		BinanceTradeClient:  binanceTradeClient,
	}
}

func decideRoute(pair model.QuotesPair) (string, error) {
	symbol := fmt.Sprintf("%s%s", pair.Base, pair.Quote)
	switch pair.Category {
	case trade.SPOT:
		return fmt.Sprintf("%s/api/v3/ticker/price?symbol=%s", spotEndpoint, symbol), nil
	case trade.FUTURES:
		return fmt.Sprintf("%s/fapi/v1/ticker/price?symbol=%s", futuresEndpoint, symbol), nil
	case trade.INVERSE:
		symbol := symbol[:len(symbol)-1] + "_PERP"
		return fmt.Sprintf("%s/dapi/v1/ticker/price?symbol=%s", inverseEndpoint, symbol), nil
	default:
		return "", errInvalidPair
	}
}

func decideRouteKlines(pair model.QuotesPair, interval string, limit int) (string, error) {
	symbol := fmt.Sprintf("%s%s", pair.Base, pair.Quote)
	switch pair.Category {
	case trade.SPOT:
		return fmt.Sprintf("%s/api/v3/klines?symbol=%s&interval=%s&limit=%d", spotEndpoint, symbol, interval, limit), nil
	case trade.FUTURES:
		return fmt.Sprintf("%s/fapi/v1/klines?symbol=%s&interval=%s&limit=%d", futuresEndpoint, symbol, interval, limit), nil
	case trade.INVERSE:
		symbol = symbol[:len(symbol)-1] + "_PERP"
		return fmt.Sprintf("%s/dapi/v1/klines?symbol=%s&interval=%s&limit=%d", inverseEndpoint, symbol, interval, limit), nil
	default:
		return "", errInvalidPair
	}
}

func decideOrderBookRoute(pair model.QuotesPair, limit int) (string, error) {
	symbol := fmt.Sprintf("%s%s", pair.Base, pair.Quote)
	switch pair.Category {
	case trade.SPOT:
		return fmt.Sprintf("%s/api/v3/depth?symbol=%s&limit=%d", spotEndpoint, symbol, limit), nil
	case trade.FUTURES:
		return fmt.Sprintf("%s/fapi/v1/depth?symbol=%s&limit=%d", futuresEndpoint, symbol, limit), nil
	case trade.INVERSE:
		symbol = symbol[:len(symbol)-1] + "_PERP"
		return fmt.Sprintf("%s/dapi/v1/depth?symbol=%s&limit=%d", inverseEndpoint, symbol, limit), nil
	default:
		return "", errInvalidPair
	}
}
