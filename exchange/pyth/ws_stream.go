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

package pyth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/model"
)

var errStreamClosed = errors.New("pyth: stream is closed")

type PythStreamClient struct {
	endpoint string
	feeds    map[string]string
	token    string
	handler  func(message []byte) error

	mu          sync.Mutex
	conn        *websocket.Conn
	subscribed  map[string]struct{}
	closed      bool
	newPrice    chan model.PricePoint
	pricePeriod chan model.PriceInterval
	orderBook   chan model.OrderBook
}

func NewStreamClient(cfg PythConfig) *PythStreamClient {
	bufferSize := cfg.BufferSize
	if bufferSize <= 0 {
		bufferSize = defaultBufferSize
	}
	endpoint := strings.TrimSpace(cfg.HermesWSEndpoint)
	if endpoint == "" {
		endpoint = defaultHermesWSEndpoint
	}
	feeds := make(map[string]string, len(cfg.FeedIDs))
	for symbol, feedID := range cfg.FeedIDs {
		feeds[strings.ToUpper(strings.TrimSpace(symbol))] = strings.TrimSpace(feedID)
	}
	return &PythStreamClient{
		endpoint:    endpoint,
		feeds:       feeds,
		token:       strings.TrimSpace(cfg.AccessToken),
		handler:     cfg.Callback,
		subscribed:  make(map[string]struct{}),
		newPrice:    make(chan model.PricePoint, bufferSize),
		pricePeriod: make(chan model.PriceInterval, bufferSize),
		orderBook:   make(chan model.OrderBook, bufferSize),
	}
}

func (c *PythStreamClient) ReceiveStream() (<-chan model.PricePoint, <-chan model.PriceInterval, <-chan model.OrderBook) {
	return c.newPrice, c.pricePeriod, c.orderBook
}

func (c *PythStreamClient) SubscribeStream(pair model.QuotesPair, _ []string) error {
	feedID, ok := c.feeds[strings.ToUpper(pair.Symbol())]
	if !ok || feedID == "" {
		return fmt.Errorf("%w: %s", errFeedNotFound, pair.Symbol())
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return errStreamClosed
	}
	if _, exists := c.subscribed[feedID]; exists {
		return nil
	}
	if c.conn != nil {
		if err := c.conn.WriteJSON(map[string]interface{}{
			"type":    "subscribe",
			"ids":     []string{feedID},
			"verbose": true,
			"binary":  true,
		}); err != nil {
			return err
		}
	}
	c.subscribed[feedID] = struct{}{}
	return nil
}

func (c *PythStreamClient) Dispatch(ctx context.Context) error {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return errStreamClosed
	}
	if c.conn == nil {
		endpoint := c.endpoint
		if c.token != "" {
			u, err := url.Parse(endpoint)
			if err != nil {
				c.mu.Unlock()
				return fmt.Errorf("pyth: invalid Hermes websocket endpoint: %w", err)
			}
			query := u.Query()
			query.Set("ACCESS_TOKEN", c.token)
			u.RawQuery = query.Encode()
			endpoint = u.String()
		}
		conn, response, err := websocket.DefaultDialer.DialContext(ctx, endpoint, nil)
		if err != nil {
			c.mu.Unlock()
			if response != nil {
				return fmt.Errorf("pyth: connect Hermes websocket (%s): %w", response.Status, err)
			}
			return fmt.Errorf("pyth: connect Hermes websocket: %w", err)
		}
		c.conn = conn
	}
	conn := c.conn
	if len(c.subscribed) > 0 {
		ids := make([]string, 0, len(c.subscribed))
		for id := range c.subscribed {
			ids = append(ids, id)
		}
		if err := conn.WriteJSON(map[string]interface{}{
			"type":    "subscribe",
			"ids":     ids,
			"verbose": true,
			"binary":  true,
		}); err != nil {
			c.mu.Unlock()
			return fmt.Errorf("pyth: subscribe Hermes feeds: %w", err)
		}
	}
	c.mu.Unlock()

	stopClose := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-stopClose:
		}
	}()
	defer close(stopClose)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("pyth: read Hermes websocket: %w", err)
		}
		if c.handler != nil {
			if err := c.handler(message); err != nil {
				return fmt.Errorf("pyth: process Hermes message: %w", err)
			}
		}
		price, ok, err := parsePriceUpdate(message)
		if err != nil {
			return err
		}
		if ok {
			c.mu.Lock()
			if !c.closed {
				model.PushToChan(c.newPrice, price)
			}
			c.mu.Unlock()
		}
	}
}

func (c *PythStreamClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true
	var err error
	if c.conn != nil {
		err = c.conn.Close()
	}
	close(c.newPrice)
	close(c.pricePeriod)
	close(c.orderBook)
	return err
}

func parsePriceUpdate(message []byte) (model.PricePoint, bool, error) {
	var update struct {
		Type      string `json:"type"`
		PriceFeed struct {
			Price struct {
				Price       string `json:"price"`
				Expo        int32  `json:"expo"`
				PublishTime int64  `json:"publish_time"`
			} `json:"price"`
		} `json:"price_feed"`
	}
	if err := json.Unmarshal(message, &update); err != nil {
		return model.PricePoint{}, false, fmt.Errorf("pyth: decode Hermes update: %w", err)
	}
	if update.Type != "price_update" {
		return model.PricePoint{}, false, nil
	}
	if update.PriceFeed.Price.Price == "" || update.PriceFeed.Price.PublishTime <= 0 {
		return model.PricePoint{}, false, errNoPrice
	}
	price, err := decimal.NewFromString(update.PriceFeed.Price.Price)
	if err != nil {
		return model.PricePoint{}, false, fmt.Errorf("pyth: invalid stream price: %w", err)
	}
	return model.PricePoint{
		NewPrice:  price.Shift(update.PriceFeed.Price.Expo),
		UpdatedAt: time.Unix(update.PriceFeed.Price.PublishTime, 0),
	}, true, nil
}
