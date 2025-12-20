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
package binance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/common/sys"
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/currency"
	"github.com/wang900115/quant/model/trade"
)

type BinanceTradeClient struct {
	client *http.Client
	ws     *websocket.Conn

	engine    *sys.Engine
	eventChan chan model.OrderEvent

	apiKey    string
	secretKey string
}

func NewTradeClient(cfg BinanceConfig) (*BinanceTradeClient, error) {
	if cfg.PrivateTimeout == 0 {
		cfg.PrivateTimeout = defaultTradeTimeout
	}
	if cfg.BufferSize == 0 {
		cfg.BufferSize = defaultBufferSize
	}
	b := &BinanceTradeClient{
		engine:    sys.NewEngine(cfg.RetryInterval, cfg.HealthCheckInterval),
		client:    &http.Client{Timeout: cfg.PrivateTimeout},
		apiKey:    cfg.APIKey,
		secretKey: cfg.SecretKey,
		eventChan: make(chan model.OrderEvent, cfg.BufferSize),
	}
	err := b.connect()
	if err != nil {
		return nil, errInitFailed
	}
	return b, nil
}

func (btc *BinanceTradeClient) createListenKey() (string, error) {
	url := fmt.Sprintf("%s/api/v3/userDataStream", spotEndpoint)
	httpReq, _ := http.NewRequest(http.MethodPost, url, nil)
	httpReq.Header.Set("X-MBX-APIKEY", btc.apiKey)
	resp, err := btc.client.Do(httpReq)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var result struct {
		ListenKey string `json:"listenKey"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	return result.ListenKey, nil
}

func (btc *BinanceTradeClient) connect() error {
	listenKey, err := btc.createListenKey()
	if err != nil {
		return err
	}
	wsEndpoint := fmt.Sprintf("%s/ws/%s", spotWsEndpoint, listenKey)
	ws, _, err := websocket.DefaultDialer.Dial(wsEndpoint, nil)
	if err != nil {
		return err
	}
	btc.ws = ws
	go btc.listen()
	return nil
}

func (btc *BinanceTradeClient) PlaceOrder(ctx context.Context, req model.OrderRequest) (*model.OrderResult, error) {
	url := fmt.Sprintf("%s/api/v3/order", spotEndpoint)

	data := map[string]string{
		"symbol":      strings.ToUpper(req.Symbol),
		"side":        string(req.Side),
		"type":        string(req.Type),
		"quantity":    req.Quantity.String(),
		"timeInForce": string(req.TimeInForce),
	}
	if req.Type == trade.LIMIT {
		data["price"] = req.Price.String()
	}

	body := strings.NewReader(buildQuery(data))
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, body)
	httpReq.Header.Set("X-MBX-APIKEY", btc.apiKey)
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := btc.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		OrderID       int64  `json:"orderId"`
		ClientOrderID string `json:"clientOrderId"`
		Status        string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &model.OrderResult{
		OrderID:       fmt.Sprint(result.OrderID),
		ClientOrderID: result.ClientOrderID,
		Symbol:        req.Symbol,
		Status:        trade.Status(result.Status),
		ExecutedQty:   decimal.Zero, // Initial executed quantity is zero
	}, nil
}

func (btc *BinanceTradeClient) GetOrder(ctx context.Context, symbol string, orderID string) (*model.OrderDetail, error) {
	url := fmt.Sprintf("%s/api/v3/order?symbol=%s&orderId=%s", spotEndpoint, strings.ToUpper(symbol), orderID)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	httpReq.Header.Set("X-MBX-APIKEY", btc.apiKey)

	resp, err := btc.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var od struct {
		OrderID       int64  `json:"orderId"`
		Symbol        string `json:"symbol"`
		Price         string `json:"price"`
		OrigQty       string `json:"origQty"`
		ExecutedQty   string `json:"executedQty"`
		Status        string `json:"status"`
		Side          string `json:"side"`
		Type          string `json:"type"`
		UpdateTime    int64  `json:"updateTime"`
		ClientOrderID string `json:"clientOrderId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&od); err != nil {
		return nil, err
	}
	price, _ := decimal.NewFromString(od.Price)
	origQty, _ := decimal.NewFromString(od.OrigQty)
	executedQty, _ := decimal.NewFromString(od.ExecutedQty)
	return &model.OrderDetail{
		OrderID:     fmt.Sprint(od.OrderID),
		Symbol:      od.Symbol,
		Price:       price,
		OrigQty:     origQty,
		ExecutedQty: executedQty,
		Status:      trade.Status(od.Status),
		Side:        trade.Signal(od.Side),
		Type:        trade.Type(od.Type),
		UpdateTime:  od.UpdateTime,
	}, nil
}

func (btc *BinanceTradeClient) CancelOrder(ctx context.Context, symbol string, orderID string) error {
	url := fmt.Sprintf("%s/api/v3/order?symbol=%s&orderId=%s", spotEndpoint, strings.ToUpper(symbol), orderID)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	httpReq.Header.Set("X-MBX-APIKEY", btc.apiKey)

	resp, err := btc.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cancel order failed: %v", resp.Status)
	}
	return nil
}

func (btc *BinanceTradeClient) GetAssetBalance(ctx context.Context, asset string) (*model.AssetBalance, error) {
	url := fmt.Sprintf("%s/api/v3/account", spotEndpoint)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	httpReq.Header.Set("X-MBX-APIKEY", btc.apiKey)

	resp, err := btc.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var account struct {
		Balances []struct {
			Asset  string `json:"asset"`
			Free   string `json:"free"`
			Locked string `json:"locked"`
		} `json:"balances"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, err
	}

	for _, b := range account.Balances {
		if strings.EqualFold(b.Asset, asset) {
			free, _ := decimal.NewFromString(b.Free)
			locked, _ := decimal.NewFromString(b.Locked)
			return &model.AssetBalance{
				Asset:  currency.CurrencySymbol(b.Asset),
				Free:   free,
				Locked: locked,
			}, nil
		}
	}

	return nil, errNonAssetFound
}

func (btc *BinanceTradeClient) listen() {
	for {
		select {
		case <-btc.engine.Done():
			return
		default:
			_, message, err := btc.ws.ReadMessage()
			if err != nil {
				continue
			}
			btc.handleMessage(message)
		}
	}
}

func (b *BinanceTradeClient) handleMessage(msg []byte) {
	var raw struct {
		EventType  string `json:"e"`
		Symbol     string `json:"s"`
		OrderID    int64  `json:"i"`
		Side       string `json:"S"`
		Type       string `json:"o"`
		Status     string `json:"X"`
		LastQty    string `json:"l"`
		FilledQty  string `json:"z"`
		UpdateTime int64  `json:"T"`
	}

	if err := json.Unmarshal(msg, &raw); err != nil {
		return
	}

	if raw.EventType != "executionReport" {
		return
	}

	last, _ := decimal.NewFromString(raw.LastQty)
	filled, _ := decimal.NewFromString(raw.FilledQty)

	b.eventChan <- model.OrderEvent{
		OrderID:    fmt.Sprint(raw.OrderID),
		Symbol:     raw.Symbol,
		Status:     trade.Status(raw.Status),
		LastQty:    last,
		FilledQty:  filled,
		Side:       trade.Signal(raw.Side),
		Type:       trade.Type(raw.Type),
		UpdateTime: raw.UpdateTime,
	}
}

func buildQuery(params map[string]string) string {
	var parts []string
	for k, v := range params {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(parts, "&")
}
