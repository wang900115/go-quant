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
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/model"
)

type OkxStreamClient struct {
	client *websocket.Conn

	handler    func(message []byte) error
	bufferSize int

	newPriceChan      chan model.PricePoint
	priceIntervalChan chan model.PriceInterval
	orderBookChan     chan model.OrderBook
}

func NewStreamClient(cfg OkxConfig) (*OkxStreamClient, error) {
	if cfg.Callback == nil {
		cfg.Callback = defaultCallback
	}
	if cfg.BufferSize == 0 {
		cfg.BufferSize = defaultBufferSize
	}
	c := &OkxStreamClient{
		handler:           cfg.Callback,
		bufferSize:        cfg.BufferSize,
		newPriceChan:      make(chan model.PricePoint, cfg.BufferSize),
		priceIntervalChan: make(chan model.PriceInterval, cfg.BufferSize),
		orderBookChan:     make(chan model.OrderBook, cfg.BufferSize),
	}
	if cfg.IsTestNet {
		if err := c.connect(wsTestEndPointPublic); err != nil {
			return nil, errInitFailed
		}
	} else {
		if err := c.connect(wsEndPointPublic); err != nil {
			return nil, errInitFailed
		}
	}
	return c, nil
}

func (oc *OkxStreamClient) connect(url string) error {
	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	oc.client = c
	return nil
}

func (oc *OkxStreamClient) Close() error {
	if oc.client != nil {
		return oc.client.Close()
	}
	close(oc.newPriceChan)
	close(oc.priceIntervalChan)
	close(oc.orderBookChan)
	return nil
}

func (oc *OkxStreamClient) ReceiveStream() (<-chan model.PricePoint, <-chan model.PriceInterval, <-chan model.OrderBook) {
	return oc.newPriceChan, oc.priceIntervalChan, oc.orderBookChan
}

func (oc *OkxStreamClient) SubscribeStream(pair model.QuotesPair, channels []string) error {
	var args []map[string]interface{}

	instId := getInstId(pair)
	for _, ch := range channels {
		args = append(args, map[string]interface{}{
			"channel": ch,
			"instId":  instId,
		})
	}

	msg := map[string]interface{}{
		"op":   "subscribe",
		"args": args,
	}
	return oc.client.WriteJSON(msg)
}

func (oc *OkxStreamClient) Dispatch(ctx context.Context) error {
	dispathcers := oc.getDispatchers()
	for {
		select {
		case <-ctx.Done():
			return oc.client.Close()
		default:
			_, message, err := oc.client.ReadMessage()
			if err != nil {
				return err
			}
			if oc.handler != nil {
				if err := oc.handler(message); err != nil {
					return err
				}
			}
			if fn, ok := dispathcers[getMessageType(message)]; ok {
				fn(oc, message)
			}
		}
	}
}

type dispatchFunc func(*OkxStreamClient, []byte)

func (oc *OkxStreamClient) getDispatchers() map[string]dispatchFunc {
	return map[string]dispatchFunc{
		"tickers": func(client *OkxStreamClient, msg []byte) {
			if p, err := parsePricePoint(msg); err == nil {
				model.PushToChan(client.newPriceChan, *p)
			}
		},
		"candle": func(client *OkxStreamClient, msg []byte) {
			if intervals, err := parsePriceInterval(msg); err == nil {
				for _, interval := range intervals {
					model.PushToChan(client.priceIntervalChan, interval)
				}
			}
		},
		"books": func(client *OkxStreamClient, msg []byte) {
			if orderBook, err := parseOrderBook(msg); err == nil {
				model.PushToChan(client.orderBookChan, *orderBook)
			}
		},
	}
}

func getMessageType(msg []byte) string {
	var raw struct {
		Event string `json:"event"`
		Arg   struct {
			Channel string `json:"channel"`
		} `json:"arg"`
	}
	_ = json.Unmarshal(msg, &raw)

	if raw.Event != "" {
		return raw.Event
	}
	return raw.Arg.Channel
}

func parsePricePoint(msg []byte) (*model.PricePoint, error) {
	var resp struct {
		Arg struct {
			InstID string `json:"instId"`
		} `json:"arg"`
		Data []struct {
			Last string `json:"last"`
			Ts   string `json:"ts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, nil
	}
	price, _ := strconv.ParseFloat(resp.Data[0].Last, 64)
	tsInt, _ := strconv.ParseInt(resp.Data[0].Ts, 10, 64)
	return &model.PricePoint{
		NewPrice:  decimal.NewFromFloat(price),
		UpdatedAt: time.UnixMilli(tsInt),
	}, nil
}

func parsePriceInterval(msg []byte) ([]model.PriceInterval, error) {
	var resp struct {
		Arg struct {
			InstID  string `json:"instId"`
			Channel string `json:"channel"`
		} `json:"arg"`
		Data [][]string `json:"data"`
	}
	if err := json.Unmarshal(msg, &resp); err != nil {
		return nil, err
	}

	intervals := make([]model.PriceInterval, 0, len(resp.Data))
	duration := parseCandleDuration(resp.Arg.Channel)

	for _, d := range resp.Data {
		if len(d) < 6 {
			continue
		}

		tsInt, _ := strconv.ParseInt(d[0], 10, 64)
		open, _ := strconv.ParseFloat(d[1], 64)
		high, _ := strconv.ParseFloat(d[2], 64)
		low, _ := strconv.ParseFloat(d[3], 64)
		closeP, _ := strconv.ParseFloat(d[4], 64)
		vol, _ := strconv.ParseFloat(d[5], 64)

		openTime := time.UnixMilli(tsInt)
		closeTime := openTime.Add(duration)

		intervals = append(intervals, model.PriceInterval{
			OpenTime:         openTime.Format(time.RFC3339),
			CloseTime:        closeTime.Format(time.RFC3339),
			OpeningPrice:     decimal.NewFromFloat(open),
			HighestPrice:     decimal.NewFromFloat(high),
			LowestPrice:      decimal.NewFromFloat(low),
			ClosingPrice:     decimal.NewFromFloat(closeP),
			Volume:           decimal.NewFromFloat(vol),
			IntervalDuration: duration,
		})
	}

	return intervals, nil
}

func parseOrderBook(msg []byte) (*model.OrderBook, error) {
	var resp struct {
		Arg struct {
			InstID string `json:"instId"`
		} `json:"arg"`
		Data []struct {
			Asks [][]string `json:"asks"`
			Bids [][]string `json:"bids"`
			Ts   string     `json:"ts"`
		} `json:"data"`
	}
	if err := json.Unmarshal(msg, &resp); err != nil {
		return nil, err
	}

	if len(resp.Data) == 0 {
		return nil, nil
	}

	data := resp.Data[0]
	var asks []model.OrderBookAsk
	var bids []model.OrderBookBid
	for _, a := range data.Asks {
		if len(a) < 2 {
			continue
		}
		price, _ := strconv.ParseFloat(a[0], 64)
		qty, _ := strconv.ParseFloat(a[1], 64)
		asks = append(asks, model.OrderBookAsk{Price: decimal.NewFromFloat(price), Quantity: decimal.NewFromFloat(qty)})
	}
	for _, b := range data.Bids {
		if len(b) < 2 {
			continue
		}
		price, _ := strconv.ParseFloat(b[0], 64)
		qty, _ := strconv.ParseFloat(b[1], 64)
		bids = append(bids, model.OrderBookBid{Price: decimal.NewFromFloat(price), Quantity: decimal.NewFromFloat(qty)})
	}
	tsInt, _ := strconv.ParseInt(data.Ts, 10, 64)

	return &model.OrderBook{
		Symbol: resp.Arg.InstID,
		Time:   time.UnixMilli(tsInt),
		Bids:   bids,
		Asks:   asks,
	}, nil
}

func parseCandleDuration(channel string) time.Duration {
	switch {
	case strings.Contains(channel, "1m"):
		return time.Minute
	case strings.Contains(channel, "3m"):
		return 3 * time.Minute
	case strings.Contains(channel, "5m"):
		return 5 * time.Minute
	case strings.Contains(channel, "15m"):
		return 15 * time.Minute
	case strings.Contains(channel, "30m"):
		return 30 * time.Minute
	case strings.Contains(channel, "1H"):
		return time.Hour
	case strings.Contains(channel, "2H"):
		return 2 * time.Hour
	case strings.Contains(channel, "4H"):
		return 4 * time.Hour
	case strings.Contains(channel, "6H"):
		return 6 * time.Hour
	case strings.Contains(channel, "12H"):
		return 12 * time.Hour
	case strings.Contains(channel, "1D"):
		return 24 * time.Hour
	case strings.Contains(channel, "1W"):
		return 7 * 24 * time.Hour
	case strings.Contains(channel, "1M"):
		return 30 * 24 * time.Hour
	default:
		return 0
	}
}
