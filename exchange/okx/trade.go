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
package okx

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/common/sys"
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/currency"
	"github.com/wang900115/quant/model/trade"
)

type OkxTradeClient struct {
	client *http.Client
	ws     *websocket.Conn

	engine    *sys.Engine
	eventChan chan model.OrderEvent

	apiKey     string
	secretKey  string
	passphrase string
}

func NewTradeClient(cfg OkxConfig) (*OkxTradeClient, error) {
	if cfg.PrivateTimeout == 0 {
		cfg.PrivateTimeout = 10 * time.Second
	}
	if cfg.BufferSize == 0 {
		cfg.BufferSize = defaultBufferSize
	}

	o := &OkxTradeClient{
		client:     &http.Client{Timeout: cfg.PrivateTimeout},
		apiKey:     cfg.APIKey,
		secretKey:  cfg.SecretKey,
		passphrase: cfg.Passphrase,
		engine:     sys.NewEngine(cfg.RetryInterval, cfg.HealthCheckInterval),
		eventChan:  make(chan model.OrderEvent, cfg.BufferSize),
	}
	err := o.connect()
	if err != nil {
		return nil, errInitFailed
	}
	return o, nil
}
func (ok *OkxTradeClient) connect() error {
	c, _, err := websocket.DefaultDialer.Dial(wsEndPointPrivate, nil)
	if err != nil {
		return err
	}
	ok.ws = c
	go ok.listen()
	return nil
}

func (ok *OkxTradeClient) PlaceOrder(ctx context.Context, req model.OrderRequest) (*model.OrderResult, error) {
	url := fmt.Sprintf("%s/api/v5/trade/order", endPoint)

	data := map[string]interface{}{
		"instId":  strings.ToUpper(req.Symbol),
		"tdMode":  "cash",
		"side":    strings.ToLower(string(req.Side)),
		"ordType": strings.ToLower(string(req.Type)),
		"sz":      req.Quantity.String(),
	}
	if req.Type == trade.LIMIT {
		data["px"] = req.Price.String()
	}

	bodyBytes, _ := json.Marshal(data)
	httpReq, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(string(bodyBytes)))
	httpReq.Header.Set("OK-ACCESS-KEY", ok.apiKey)
	httpReq.Header.Set("OK-ACCESS-SIGN", ok.secretKey) // Note: In practice, the signature should be generated properly
	httpReq.Header.Set("OK-ACCESS-PASSPHRASE", ok.passphrase)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := ok.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Code string `json:"code"`
		Data []struct {
			OrdId string `json:"ordId"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Data) == 0 {
		return nil, fmt.Errorf("order failed: %s", result.Code)
	}

	return &model.OrderResult{
		OrderID:       result.Data[0].OrdId,
		ClientOrderID: req.ClientOrderID,
		Symbol:        req.Symbol,
		Status:        trade.NEW,
		ExecutedQty:   decimal.Zero,
	}, nil
}

func (ok *OkxTradeClient) GetOrder(ctx context.Context, symbol string, orderID string) (*model.OrderDetail, error) {
	url := fmt.Sprintf("%s/api/v5/trade/order?ordId=%s&instId=%s", endPoint, orderID, strings.ToUpper(symbol))
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	req.Header.Set("OK-ACCESS-KEY", ok.apiKey)
	req.Header.Set("OK-ACCESS-SIGN", ok.secretKey) // Note: In practice, the signature should be generated properly
	req.Header.Set("OK-ACCESS-PASSPHRASE", ok.passphrase)

	resp, err := ok.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw struct {
		Code string `json:"code"`
		Data []struct {
			OrdId     string `json:"ordId"`
			InstId    string `json:"instId"`
			Px        string `json:"px"`
			Sz        string `json:"sz"`
			AccFillSz string `json:"accFillSz"`
			State     string `json:"state"`
			Side      string `json:"side"`
			TdMode    string `json:"tdMode"`
			OrdType   string `json:"ordType"`
			UTime     string `json:"uTime"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	if len(raw.Data) == 0 {
		return nil, fmt.Errorf("order not found")
	}

	d := raw.Data[0]
	price, _ := decimal.NewFromString(d.Px)
	origQty, _ := decimal.NewFromString(d.Sz)
	executedQty, _ := decimal.NewFromString(d.AccFillSz)
	updateTime := time.Now().UnixMilli()

	return &model.OrderDetail{
		OrderID:     d.OrdId,
		Symbol:      d.InstId,
		Price:       price,
		OrigQty:     origQty,
		ExecutedQty: executedQty,
		Status:      trade.Status(d.State),
		Side:        trade.Signal(d.Side),
		Type:        trade.Type(d.OrdType),
		UpdateTime:  updateTime,
	}, nil
}

func (ok *OkxTradeClient) CancelOrder(ctx context.Context, symbol string, orderID string) error {
	url := fmt.Sprintf("%s/api/v5/trade/cancel-order?instId=%s&ordId=%s", endPoint, strings.ToUpper(symbol), orderID)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)

	req.Header.Set("OK-ACCESS-KEY", ok.apiKey)
	req.Header.Set("OK-ACCESS-SIGN", ok.secretKey) // Note: In practice, the signature should be generated properly
	req.Header.Set("OK-ACCESS-PASSPHRASE", ok.passphrase)

	resp, err := ok.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("cancel order failed: %v", resp.Status)
	}
	return nil
}

func (ok *OkxTradeClient) GetAssetBalance(ctx context.Context, asset string) (*model.AssetBalance, error) {
	url := fmt.Sprintf("%s/api/v5/account/balance", endPoint)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)

	req.Header.Set("OK-ACCESS-KEY", ok.apiKey)
	req.Header.Set("OK-ACCESS-SIGN", ok.secretKey) // Note: In practice, the signature should be generated properly
	req.Header.Set("OK-ACCESS-PASSPHRASE", ok.passphrase)

	resp, err := ok.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw struct {
		Code string `json:"code"`
		Data []struct {
			Eq      string `json:"eq"`
			Details []struct {
				Ccy    string `json:"ccy"`
				Avail  string `json:"avail"`
				Locked string `json:"locked"`
			} `json:"details"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	for _, account := range raw.Data {
		for _, b := range account.Details {
			if strings.EqualFold(b.Ccy, asset) {
				free, _ := decimal.NewFromString(b.Avail)
				locked, _ := decimal.NewFromString(b.Locked)
				return &model.AssetBalance{
					Asset:  currency.CurrencySymbol(b.Ccy),
					Free:   free,
					Locked: locked,
				}, nil
			}
		}
	}
	return nil, errNonAssetFound
}

func (ok *OkxTradeClient) listen() {
	orderPrevFilled := make(map[string]decimal.Decimal) // orderID -> previously filled quantity

	for {
		select {
		case <-ok.engine.Done():
			return
		default:
			_, message, err := ok.ws.ReadMessage()
			if err != nil {
				continue
			}
			ok.handleMessage(message, orderPrevFilled)
		}
	}
}

func (ok *OkxTradeClient) handleMessage(msg []byte, orderPrevFilled map[string]decimal.Decimal) {
	var raw struct {
		Arg  map[string]interface{} `json:"arg"`
		Data []struct {
			InstId    string `json:"instId"`
			OrdId     string `json:"ordId"`
			Side      string `json:"side"`
			OrdType   string `json:"ordType"`
			State     string `json:"state"`
			AccFillSz string `json:"accFillSz"`
			Ts        string `json:"ts"`
		} `json:"data"`
	}

	if err := json.Unmarshal(msg, &raw); err != nil {
		return
	}

	for _, d := range raw.Data {
		filledQty, _ := decimal.NewFromString(d.AccFillSz)
		prevFilled := orderPrevFilled[d.OrdId]
		lastQty := filledQty.Sub(prevFilled)
		orderPrevFilled[d.OrdId] = filledQty

		ts, _ := strconv.ParseInt(d.Ts, 10, 64)

		ok.eventChan <- model.OrderEvent{
			OrderID:    d.OrdId,
			Symbol:     d.InstId,
			Status:     okxStateToTradeStatus(d.State),
			FilledQty:  filledQty,
			LastQty:    lastQty,
			Side:       trade.Signal(d.Side),
			Type:       trade.Type(d.OrdType),
			UpdateTime: ts,
		}
	}
}

func okxStateToTradeStatus(state string) trade.Status {
	switch state {
	case "live":
		return trade.NEW
	case "partially_filled":
		return trade.PARTIALLY_FILLED
	case "filled":
		return trade.FILLED
	case "canceled":
		return trade.CANCELED
	default:
		return trade.UNKNOWN
	}
}
