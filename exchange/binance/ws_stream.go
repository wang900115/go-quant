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
	"fmt"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/trade"
)

type BinanceStreamClient struct {
	spotClient    *websocket.Conn
	futuresClient *websocket.Conn
	inverseClient *websocket.Conn

	handler    func(message []byte) error
	bufferSize int

	newPriceChan      chan model.PricePoint
	priceIntervalChan chan model.PriceInterval
	orderBookChan     chan model.OrderBook
}

func NewStreamClient(cfg BinanceConfig) (*BinanceStreamClient, error) {
	if cfg.Callback == nil {
		cfg.Callback = defaultCallback
	}
	if cfg.BufferSize == 0 {
		cfg.BufferSize = defaultBufferSize
	}
	c := &BinanceStreamClient{
		handler:           cfg.Callback,
		bufferSize:        cfg.BufferSize,
		newPriceChan:      make(chan model.PricePoint, cfg.BufferSize),
		priceIntervalChan: make(chan model.PriceInterval, cfg.BufferSize),
		orderBookChan:     make(chan model.OrderBook, cfg.BufferSize),
	}
	var err error
	if cfg.IstestNet {
		c.spotClient, err = c.connect(spotTestWsEndpoint)
		if err != nil {
			return nil, errInitFailed
		}
		c.futuresClient, err = c.connect(futuresTestWsEndpoint)
		if err != nil {
			return nil, errInitFailed
		}
		c.inverseClient, err = c.connect(inverseTestWsEndpoint)
		if err != nil {
			return nil, errInitFailed
		}
	} else {
		c.spotClient, err = c.connect(spotWsEndpoint)
		if err != nil {
			return nil, errInitFailed
		}
		c.futuresClient, err = c.connect(futuresWsEndpoint)
		if err != nil {
			return nil, errInitFailed
		}
		c.inverseClient, err = c.connect(inverseWsEndpoint)
		if err != nil {
			return nil, errInitFailed
		}
	}
	return c, nil
}

func (bc *BinanceStreamClient) connect(endpoint string) (*websocket.Conn, error) {
	conn, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (bc *BinanceStreamClient) ReceiveStream() (<-chan model.PricePoint, <-chan model.PriceInterval, <-chan model.OrderBook) {
	return bc.newPriceChan, bc.priceIntervalChan, bc.orderBookChan
}

func (bc *BinanceStreamClient) SubscribeStream(pair model.QuotesPair, streamType []string) error {
	client, err := bc.getClient(pair)
	if err != nil {
		return err
	}
	params := []string{}
	for _, st := range streamType {
		params = append(params, fmt.Sprintf("%s@%s", strings.ToLower(pair.Base.String()+pair.Quote.String()), st))
	}

	msg := map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": params,
		"id":     time.Now().Unix(),
	}
	return client.WriteJSON(msg)
}

func (bc *BinanceStreamClient) Dispatch(ctx context.Context) error {
	dispatchers := bc.getDispatchers()
	clients := []*websocket.Conn{bc.spotClient, bc.futuresClient, bc.inverseClient}

	// Start goroutine for each WebSocket connection
	for _, client := range clients {
		go func(conn *websocket.Conn) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					_, message, err := conn.ReadMessage()
					if err != nil {
						return
					}
					if bc.handler != nil {
						bc.handler(message)
					}
					if fn, ok := dispatchers[getMessageType(message)]; ok {
						fn(bc, message)
					}
				}
			}
		}(client)
	}

	<-ctx.Done()
	return nil
}

func (bc *BinanceStreamClient) Close() error {
	if err := bc.spotClient.Close(); err != nil {
		return err
	}
	if err := bc.futuresClient.Close(); err != nil {
		return err
	}
	if err := bc.inverseClient.Close(); err != nil {
		return err
	}
	close(bc.newPriceChan)
	close(bc.priceIntervalChan)
	close(bc.orderBookChan)
	return nil
}

type dispatchFunc func(*BinanceStreamClient, []byte)

func (bc *BinanceStreamClient) getDispatchers() map[string]dispatchFunc {
	return map[string]dispatchFunc{
		"24hrTicker": func(client *BinanceStreamClient, msg []byte) {
			if p, err := parsePricePoint(msg); err == nil {
				model.PushToChan(client.newPriceChan, *p)
			}
		},
		"kline": func(client *BinanceStreamClient, msg []byte) {
			if intervals, err := parsePriceInterval(msg); err == nil {
				for _, interval := range intervals {
					model.PushToChan(client.priceIntervalChan, interval)
				}
			}
		},
		"depthUpdate": func(client *BinanceStreamClient, msg []byte) {
			if ob, err := parseOrderBook(msg); err == nil {
				model.PushToChan(client.orderBookChan, *ob)
			}
		},
	}
}

func (bc *BinanceStreamClient) getClient(pair model.QuotesPair) (*websocket.Conn, error) {
	switch pair.Category {
	case trade.SPOT:
		return bc.spotClient, nil
	case trade.FUTURES:
		return bc.futuresClient, nil
	case trade.INVERSE:
		return bc.inverseClient, nil
	default:
		return nil, errInvalidPair
	}
}

func getMessageType(msg []byte) string {
	// Important: Must declare both 'e' and 'E' fields to avoid Go JSON parsing bug
	// when both lowercase and uppercase keys exist in the same JSON
	var raw struct {
		E  string `json:"e"` // Event type
		ET int64  `json:"E"` // Event time (must declare to avoid parsing confusion)
	}
	if err := json.Unmarshal(msg, &raw); err != nil {
		return ""
	}
	return raw.E
}

func parsePricePoint(msg []byte) (*model.PricePoint, error) {
	// Must declare both lowercase and uppercase fields to avoid Go JSON parsing bug
	// when JSON contains keys that differ only in case (e.g., "c" and "C", "e" and "E")
	var raw struct {
		Event     string `json:"e"` // event type
		Symbol    string `json:"s"` // symbol
		Price     string `json:"c"` // close price
		Time      int64  `json:"E"` // event time
		CloseTime int64  `json:"C"` // close time (must declare to avoid conflict with lowercase "c")
	}
	if err := json.Unmarshal(msg, &raw); err != nil {
		return nil, err
	}
	if raw.Price == "" {
		return nil, errBinanceNoData
	}
	price, err := decimal.NewFromString(raw.Price)
	if err != nil {
		return nil, err
	}
	return &model.PricePoint{
		NewPrice:  price,
		UpdatedAt: time.UnixMilli(raw.Time),
	}, nil
}

func parsePriceInterval(msg []byte) ([]model.PriceInterval, error) {
	var raw struct {
		K struct {
			StartTime int64  `json:"t"`
			EndTime   int64  `json:"T"`
			Open      string `json:"o"`
			Close     string `json:"c"`
			High      string `json:"h"`
			Low       string `json:"l"`
			Volume    string `json:"v"`
		} `json:"k"`
	}
	if err := json.Unmarshal(msg, &raw); err != nil {
		return nil, err
	}

	open, err := decimal.NewFromString(raw.K.Open)
	if err != nil {
		return nil, err
	}
	closeP, err := decimal.NewFromString(raw.K.Close)
	if err != nil {
		return nil, err
	}
	high, err := decimal.NewFromString(raw.K.High)
	if err != nil {
		return nil, err
	}
	low, err := decimal.NewFromString(raw.K.Low)
	if err != nil {
		return nil, err
	}
	volume, err := decimal.NewFromString(raw.K.Volume)
	if err != nil {
		return nil, err
	}

	interval := model.PriceInterval{
		OpenTime:         time.UnixMilli(raw.K.StartTime).Format(time.RFC3339),
		CloseTime:        time.UnixMilli(raw.K.EndTime).Format(time.RFC3339),
		OpeningPrice:     open,
		ClosingPrice:     closeP,
		HighestPrice:     high,
		LowestPrice:      low,
		Volume:           volume,
		IntervalDuration: time.Duration(raw.K.EndTime-raw.K.StartTime) * time.Millisecond,
	}

	return []model.PriceInterval{interval}, nil
}

func parseOrderBook(msg []byte) (*model.OrderBook, error) {
	var raw struct {
		Symbol string          `json:"s"`
		Bids   [][]interface{} `json:"b"`
		Asks   [][]interface{} `json:"a"`
		Time   int64           `json:"E"`
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
		Symbol: raw.Symbol,
		Time:   time.UnixMilli(raw.Time),
		Bids:   bids,
		Asks:   asks,
	}, nil
}
