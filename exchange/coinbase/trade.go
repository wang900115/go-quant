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

// todo subscribe user order channel
package coinbase

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/common/sys"
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/currency"
	"github.com/wang900115/quant/model/trade"
)

type CoinbaseTradeClient struct {
	client *http.Client
	ws     *websocket.Conn

	engine    *sys.Engine
	eventChan chan model.OrderEvent

	apiKey     string
	secretKey  string
	passphrase string
}

func NewTradeClient(cfg CoinbaseConfig) (*CoinbaseTradeClient, error) {
	if cfg.PrivateTimeout == 0 {
		cfg.PrivateTimeout = 10 * time.Second
	}
	if cfg.BufferSize == 0 {
		cfg.BufferSize = defaultBufferSize
	}
	c := &CoinbaseTradeClient{
		client:     &http.Client{Timeout: cfg.PrivateTimeout},
		apiKey:     cfg.APIKey,
		secretKey:  cfg.SecretKey,
		passphrase: cfg.Passphrase,
		engine:     sys.NewEngine(cfg.RetryInterval, cfg.HealthCheckInterval),
		eventChan:  make(chan model.OrderEvent, cfg.BufferSize),
	}
	err := c.connect()
	if err != nil {
		return nil, errInitFailed
	}
	return c, nil
}

func (cb *CoinbaseTradeClient) connect() error {
	c, _, err := websocket.DefaultDialer.Dial(spotWsPEndpoint, nil)
	if err != nil {
		return err
	}
	cb.ws = c
	go cb.listen()
	return nil
}

func (cb *CoinbaseTradeClient) PlaceOrder(ctx context.Context, req model.OrderRequest) (*model.OrderResult, error) {
	url := fmt.Sprintf("%s/orders", spotEndpoint)

	data := map[string]string{
		"product_id": strings.ToUpper(req.Symbol),
		"side":       strings.ToLower(string(req.Side)),
		"type":       strings.ToLower(string(req.Type)),
		"size":       req.Quantity.String(),
	}
	if req.Type == trade.LIMIT {
		data["price"] = req.Price.String()
	}

	bodyBytes, _ := json.Marshal(data)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(bodyBytes)))
	httpReq.Header.Set("CB-ACCESS-KEY", cb.apiKey)
	httpReq.Header.Set("CB-ACCESS-SIGN", cb.secretKey) // Note: In practice, the signature should be generated properly
	httpReq.Header.Set("CB-ACCESS-PASSPHRASE", cb.passphrase)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := cb.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &model.OrderResult{
		OrderID:       result.ID,
		ClientOrderID: req.ClientOrderID,
		Symbol:        req.Symbol,
		Status:        trade.NEW,
		ExecutedQty:   decimal.Zero,
	}, nil
}

func (cb *CoinbaseTradeClient) GetOrder(ctx context.Context, symbol string, orderID string) (*model.OrderDetail, error) {
	url := fmt.Sprintf("%s/orders/%s", spotEndpoint, orderID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	req.Header.Set("CB-ACCESS-KEY", cb.apiKey)
	req.Header.Set("CB-ACCESS-SIGN", cb.secretKey) // in Note: In practice, the signature should be generated properly
	req.Header.Set("CB-ACCESS-PASSPHRASE", cb.passphrase)

	resp, err := cb.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var od struct {
		ID         string `json:"id"`
		ProductID  string `json:"product_id"`
		Price      string `json:"price"`
		Size       string `json:"size"`
		FilledSize string `json:"filled_size"`
		Status     string `json:"status"`
		Side       string `json:"side"`
		Type       string `json:"type"`
		CreatedAt  string `json:"created_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&od); err != nil {
		return nil, err
	}

	price, _ := decimal.NewFromString(od.Price)
	origQty, _ := decimal.NewFromString(od.Size)
	executedQty, _ := decimal.NewFromString(od.FilledSize)
	updateTime := time.Now().UnixMilli()

	return &model.OrderDetail{
		OrderID:     od.ID,
		Symbol:      od.ProductID,
		Price:       price,
		OrigQty:     origQty,
		ExecutedQty: executedQty,
		Status:      trade.Status(od.Status),
		Side:        trade.Signal(od.Side),
		Type:        trade.Type(od.Type),
		UpdateTime:  updateTime,
	}, nil
}

func (cb *CoinbaseTradeClient) CancelOrder(ctx context.Context, symbol string, orderID string) error {
	url := fmt.Sprintf("%s/orders/%s", spotEndpoint, orderID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)

	req.Header.Set("CB-ACCESS-KEY", cb.apiKey)
	req.Header.Set("CB-ACCESS-SIGN", cb.secretKey) // in Note: In practice, the signature should be generated properly
	req.Header.Set("CB-ACCESS-PASSPHRASE", cb.passphrase)

	resp, err := cb.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cancel order failed: %v", resp.Status)
	}
	return nil
}

func (cb *CoinbaseTradeClient) GetAssetBalance(ctx context.Context, asset string) (*model.AssetBalance, error) {
	url := fmt.Sprintf("%s/accounts", spotEndpoint)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	req.Header.Set("CB-ACCESS-KEY", cb.apiKey)
	req.Header.Set("CB-ACCESS-SIGN", cb.secretKey) // Note: In practice, the signature should be generated properly
	req.Header.Set("CB-ACCESS-PASSPHRASE", cb.passphrase)

	resp, err := cb.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var accounts []struct {
		Currency string `json:"currency"`
		Balance  string `json:"balance"`
		Hold     string `json:"hold"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&accounts); err != nil {
		return nil, err
	}

	for _, a := range accounts {
		if strings.EqualFold(a.Currency, asset) {
			free, _ := decimal.NewFromString(a.Balance)
			locked, _ := decimal.NewFromString(a.Hold)
			return &model.AssetBalance{
				Asset:  currency.CurrencySymbol(a.Currency),
				Free:   free,
				Locked: locked,
			}, nil
		}
	}

	return nil, errNonAssetFound
}

func (cb *CoinbaseTradeClient) listen() {
	for {
		select {
		case <-cb.engine.Done():
			return
		default:
			_, msg, err := cb.ws.ReadMessage()
			if err != nil {
				continue
			}
			cb.handleMessage(msg)
		}
	}
}

func (cb *CoinbaseTradeClient) handleMessage(msg []byte) {
	var raw struct {
		Type    string `json:"type"`
		Event   string `json:"event"`
		OrderID string `json:"order_id"`
		Product string `json:"product_id"`
		Size    string `json:"size"`
		Filled  string `json:"filled_size"`
		Status  string `json:"status"`
		Side    string `json:"side"`
		TimeMS  int64  `json:"timestamp"`
	}

	if err := json.Unmarshal(msg, &raw); err != nil {
		return
	}

	// only process real order events
	if raw.Event == "" || !strings.HasPrefix(raw.Event, "user.order_") {
		return
	}

	filled, _ := decimal.NewFromString(raw.Filled)
	qty, _ := decimal.NewFromString(raw.Size)

	evt := model.OrderEvent{
		OrderID:    raw.OrderID,
		Symbol:     raw.Product,
		Status:     trade.Status(raw.Status),
		LastQty:    qty.Sub(filled),
		FilledQty:  filled,
		Side:       trade.Signal(raw.Side),
		Type:       trade.Type(raw.Type),
		UpdateTime: raw.TimeMS,
	}

	select {
	case cb.eventChan <- evt:
	default:
	}
}
