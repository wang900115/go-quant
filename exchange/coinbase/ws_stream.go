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
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/model"
)

type CoinbaseStreamClient struct {
	client *websocket.Conn

	handler    func(message []byte) error
	bufferSize int

	newPriceChan      chan model.PricePoint
	priceIntervalChan chan model.PriceInterval
	orderBookChan     chan model.OrderBook
}

func NewStreamClient(cfg CoinbaseConfig) (*CoinbaseStreamClient, error) {
	if cfg.Callback == nil {
		cfg.Callback = defaultCallback
	}
	if cfg.BufferSize == 0 {
		cfg.BufferSize = defaultBufferSize
	}

	c := &CoinbaseStreamClient{
		handler:           cfg.Callback,
		bufferSize:        cfg.BufferSize,
		newPriceChan:      make(chan model.PricePoint, cfg.BufferSize),
		priceIntervalChan: make(chan model.PriceInterval, cfg.BufferSize),
		orderBookChan:     make(chan model.OrderBook, cfg.BufferSize),
	}
	if cfg.IstestNet {
		if err := c.connect(testSpotWsEndpoint); err != nil {
			return nil, errInitFailed
		}
	} else {
		if err := c.connect(spotWsEndpoint); err != nil {
			return nil, errInitFailed
		}
	}
	return c, nil
}

func (cc *CoinbaseStreamClient) connect(url string) error {
	var dialer websocket.Dialer
	conn, _, err := dialer.Dial(url, nil)
	if err != nil {
		return err
	}
	cc.client = conn
	return nil
}

func (cc *CoinbaseStreamClient) ReceiveStream() (<-chan model.PricePoint, <-chan model.PriceInterval, <-chan model.OrderBook) {
	return cc.newPriceChan, cc.priceIntervalChan, cc.orderBookChan
}

func (cc *CoinbaseStreamClient) SubscribeStream(pair model.QuotesPair, channelsType []string) error {
	symbol := fmt.Sprintf("%s-%s", pair.Base, pair.Quote)
	subscribeMsg := map[string]interface{}{
		"type":        "subscribe",
		"product_ids": []string{symbol},
		"channels":    channelsType,
	}
	return cc.client.WriteJSON(subscribeMsg)
}

func (cc *CoinbaseStreamClient) Dispatch(ctx context.Context) error {
	dispatchers := cc.getDispatchers()
	for {
		select {
		case <-ctx.Done():
			return cc.client.Close()
		default:
			_, message, err := cc.client.ReadMessage()
			if err != nil {
				return err
			}
			if cc.handler != nil {
				cc.handler(message)
			}
			if fn, ok := dispatchers[getMessageType(message)]; ok {
				fn(cc, message)
			}
		}
	}
}

func (cc *CoinbaseStreamClient) Close() error {
	if err := cc.client.Close(); err != nil {
		return err
	}
	close(cc.newPriceChan)
	close(cc.priceIntervalChan)
	close(cc.orderBookChan)
	return nil
}

type dispatchFunc func(*CoinbaseStreamClient, []byte)

func (cc *CoinbaseStreamClient) getDispatchers() map[string]dispatchFunc {
	return map[string]dispatchFunc{
		"ticker": func(client *CoinbaseStreamClient, msg []byte) {
			if p, err := parsePricePoint(msg); err == nil {
				model.PushToChan(client.newPriceChan, *p)
			}
		},
		"candles": func(client *CoinbaseStreamClient, msg []byte) {
			if intervals, err := parsePriceInterval(msg); err == nil {
				for _, interval := range intervals {
					model.PushToChan(client.priceIntervalChan, interval)
				}
			}
		},
		"level2": func(client *CoinbaseStreamClient, msg []byte) {
			if orderBook, err := parseOrderBook(msg); err == nil {
				model.PushToChan(client.orderBookChan, *orderBook)
			}
		},
	}
}

func parsePricePoint(msg []byte) (*model.PricePoint, error) {
	var raw struct {
		Type      string `json:"type"`
		ProductID string `json:"product_id"`
		Price     string `json:"price"`
	}
	if err := json.Unmarshal(msg, &raw); err != nil {
		return nil, err
	}
	p, err := decimal.NewFromString(raw.Price)
	if err != nil {
		return nil, err
	}
	return &model.PricePoint{
		NewPrice:  p,
		UpdatedAt: time.Now(),
	}, nil
}

func parsePriceInterval(msg []byte) ([]model.PriceInterval, error) {
	var raw [][]interface{}
	if err := json.Unmarshal(msg, &raw); err != nil {
		return nil, err
	}

	intervals := make([]model.PriceInterval, 0, len(raw))
	for _, candle := range raw {
		if len(candle) < 6 {
			continue
		}

		t := int64(candle[0].(float64))
		openTime := time.Unix(t, 0)

		openPrice, _ := decimal.NewFromString(fmt.Sprintf("%v", candle[3]))
		highPrice, _ := decimal.NewFromString(fmt.Sprintf("%v", candle[2]))
		lowPrice, _ := decimal.NewFromString(fmt.Sprintf("%v", candle[1]))
		closePrice, _ := decimal.NewFromString(fmt.Sprintf("%v", candle[4]))
		volume, _ := decimal.NewFromString(fmt.Sprintf("%v", candle[5]))

		intervals = append(intervals, model.PriceInterval{
			OpenTime:         openTime.Format(time.RFC3339),
			OpeningPrice:     openPrice,
			HighestPrice:     highPrice,
			LowestPrice:      lowPrice,
			ClosingPrice:     closePrice,
			Volume:           volume,
			CloseTime:        openTime.Add(time.Minute).Format(time.RFC3339),
			IntervalDuration: time.Minute,
		})
	}

	return intervals, nil
}

func parseDecimalFromInterface(v interface{}) (decimal.Decimal, error) {
	switch val := v.(type) {
	case string:
		return decimal.NewFromString(val)
	case float64:
		return decimal.NewFromFloat(val), nil
	case int:
		return decimal.NewFromInt(int64(val)), nil
	case int64:
		return decimal.NewFromInt(val), nil
	default:
		return decimal.Zero, errNotValidType
	}
}

func parseOrderBook(msg []byte) (*model.OrderBook, error) {
	var raw struct {
		Bids [][]interface{} `json:"bids"`
		Asks [][]interface{} `json:"asks"`
	}
	if err := json.Unmarshal(msg, &raw); err != nil {
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
		Time: time.Now(),
		Bids: bids,
		Asks: asks,
	}, nil
}
