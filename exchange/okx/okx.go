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

package okx

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/common/parse"
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/trade"
)

const (
	endPoint          = "https://www.okx.com"
	wsEndPointPublic  = "wss://ws.okx.com:8443/ws/v5/public"
	wsEndPointPrivate = "wss://ws.okx.com:8443/ws/v5/private"

	wsTestEndPointPublic  = "wss://wspap.okx.com:8443/ws/v5/public?brokerId=9999"
	wsTestEndPointPrivate = "wss://wspap.okx.com:8443/ws/v5/private?brokerId=9999"
)

var defaultCallback = func(message []byte) error {
	return nil
}

const (
	defaultTimeout    = 10 * time.Second
	defaultBufferSize = 100
)

var (
	errOkxNoData      = errors.New("okx: no data returned")
	errResponseFailed = errors.New("okx: response failed")
	errNotValidType   = errors.New("okx: not valid type")
	errInitFailed     = errors.New("okx: initialization failed")
	errNonAssetFound  = errors.New("okx: no such asset found")
)

type OkxSingleClient struct {
	httpClient *http.Client
}

func NewSingleClient(cfg OkxConfig) *OkxSingleClient {
	timeout := cfg.PublicTimeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	return &OkxSingleClient{
		httpClient: &http.Client{Timeout: timeout},
	}
}

func (oc *OkxSingleClient) GetPrice(ctx context.Context, pair model.QuotesPair) (*model.PricePoint, error) {
	instId := getInstId(pair)
	url := fmt.Sprintf("%s/api/v5/market/ticker?instId=%s", endPoint, instId)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := oc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errResponseFailed
	}
	var raw struct {
		Code string `json:"code"`
		Msg  string `json:"msg"`
		Data []struct {
			Last string `json:"last"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if len(raw.Data) == 0 {
		return nil, errOkxNoData
	}
	price, err := decimal.NewFromString(raw.Data[0].Last)
	if err != nil {
		return nil, err
	}
	return &model.PricePoint{
		NewPrice:  price,
		UpdatedAt: time.Now(),
	}, nil
}

// interval: 1m, 3m, 5m, 15m, 30m, 1H, 2H, 4H, 6H, 12H, 1D, 1W, 1M, 3M
func (oc *OkxSingleClient) GetKlines(ctx context.Context, pair model.QuotesPair, interval string, limit int) ([]model.PriceInterval, error) {
	instId := getInstId(pair)
	url := fmt.Sprintf("%s/api/v5/market/candles?instId=%s&bar=%s&limit=%d",
		endPoint, instId, interval, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := oc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errResponseFailed
	}
	var raw struct {
		Code string     `json:"code"`
		Msg  string     `json:"msg"`
		Data [][]string `json:"data"` // [ts, open, high, low, close, vol, volCcy, volCcyQuote, confirm]
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode klines: %w", err)
	}

	if raw.Code != "0" {
		return nil, fmt.Errorf("okx api error: code=%s, msg=%s", raw.Code, raw.Msg)
	}

	intervals := make([]model.PriceInterval, 0, limit)

	for _, kline := range raw.Data {
		if len(kline) < 6 {
			return nil, errOkxNoData
		}

		// OKX response: [ts, open, high, low, close, vol, volCcy, volCcyQuote, confirm]
		// ts is milliseconds timestamp
		timestamp, err := decimal.NewFromString(kline[0])
		if err != nil {
			return nil, err
		}
		openTime := time.UnixMilli(timestamp.IntPart())

		openPrice, err := decimal.NewFromString(kline[1])
		if err != nil {
			return nil, err
		}
		highPrice, err := decimal.NewFromString(kline[2])
		if err != nil {
			return nil, err
		}
		lowPrice, err := decimal.NewFromString(kline[3])
		if err != nil {
			return nil, err
		}
		closePrice, err := decimal.NewFromString(kline[4])
		if err != nil {
			return nil, err
		}
		volume, err := decimal.NewFromString(kline[5])
		if err != nil {
			return nil, err
		}

		// 從interval推導時長
		duration := parse.ParseInterval(interval)
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

func (oc *OkxSingleClient) GetOrderBook(ctx context.Context, pair model.QuotesPair, limit int) (*model.OrderBook, error) {
	instId := fmt.Sprintf("%s-%s", pair.Base, pair.Quote)
	url := fmt.Sprintf("%s/api/v5/market/books?instId=%s&sz=%d", endPoint, instId, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := oc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errResponseFailed
	}

	var raw struct {
		Code string `json:"code"`
		Data []struct {
			Bids [][]interface{} `json:"bids"`
			Asks [][]interface{} `json:"asks"`
			Ts   string          `json:"ts"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if raw.Code != "0" || len(raw.Data) == 0 {
		return nil, errResponseFailed
	}
	ts, _ := strconv.ParseInt(raw.Data[0].Ts, 10, 64)
	bids, err := model.ParseOrderEntries[model.OrderBookBid](raw.Data[0].Bids)
	if err != nil {
		return nil, err
	}
	asks, err := model.ParseOrderEntries[model.OrderBookAsk](raw.Data[0].Asks)
	if err != nil {
		return nil, err
	}
	return &model.OrderBook{
		Symbol: pair.Symbol(),
		Time:   time.UnixMilli(ts),
		Bids:   bids,
		Asks:   asks,
	}, nil
}

type OkxConfig struct {
	IsTestNet      bool
	PublicTimeout  time.Duration
	PrivateTimeout time.Duration
	BufferSize     int
	Callback       func(message []byte) error
	APIKey         string
	SecretKey      string
	Passphrase     string

	RetryInterval       time.Duration
	HealthCheckInterval time.Duration
}

type OkxClient struct {
	*OkxSingleClient
	*OkxStreamClient
	*OkxTradeClient
}

func New(config OkxConfig) *OkxClient {
	streamClient, err := NewStreamClient(config)
	if err != nil {
		panic(err)
	}
	return &OkxClient{
		OkxSingleClient: NewSingleClient(config),
		OkxStreamClient: streamClient,
	}
}

func getInstId(pair model.QuotesPair) string {
	base := fmt.Sprintf("%s-%s", pair.Base, pair.Quote)
	switch pair.Category {
	case trade.SPOT:
		return base
	case trade.FUTURES:
		return base + "-SWAP"
	default:
		return base
	}
}
