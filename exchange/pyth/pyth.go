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
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/exchange"
	"github.com/wang900115/quant/model"
)

const (
	defaultHermesEndpoint   = "https://hermes.pyth.network"
	defaultHermesWSEndpoint = "wss://hermes.pyth.network/ws"
	defaultTimeout          = 10 * time.Second
	defaultBufferSize       = 100
)

var (
	ErrUnsupported  = errors.New("pyth: operation not supported")
	errFeedNotFound = errors.New("pyth: no feed configured for pair")
	errNoPrice      = errors.New("pyth: no price returned")
)

type PythConfig struct {
	// FeedIDs maps pair symbols such as "BTC/USD" to Pyth price feed IDs.
	FeedIDs       map[string]string
	AccessToken   string
	PublicTimeout time.Duration
	BufferSize    int
	Callback      func(message []byte) error

	// HermesEndpoint and HermesWSEndpoint can be overridden for testing or self-hosted Hermes.
	HermesEndpoint   string
	HermesWSEndpoint string
}

type PythSingleClient struct {
	client      *http.Client
	endpoint    string
	feedIDs     map[string]string
	accessToken string
}

func NewSingleClient(cfg PythConfig) *PythSingleClient {
	timeout := cfg.PublicTimeout
	if timeout == 0 {
		timeout = defaultTimeout
	}
	endpoint := strings.TrimRight(cfg.HermesEndpoint, "/")
	if endpoint == "" {
		endpoint = defaultHermesEndpoint
	}
	feedIDs := make(map[string]string, len(cfg.FeedIDs))
	for symbol, feedID := range cfg.FeedIDs {
		feedIDs[strings.ToUpper(strings.TrimSpace(symbol))] = strings.TrimSpace(feedID)
	}
	return &PythSingleClient{
		client:      &http.Client{Timeout: timeout},
		endpoint:    endpoint,
		feedIDs:     feedIDs,
		accessToken: strings.TrimSpace(cfg.AccessToken),
	}
}

func (c *PythSingleClient) feedID(pair model.QuotesPair) (string, error) {
	feedID, ok := c.feedIDs[strings.ToUpper(pair.Symbol())]
	if !ok || feedID == "" {
		return "", fmt.Errorf("%w: %s", errFeedNotFound, pair.Symbol())
	}
	return feedID, nil
}

func (c *PythSingleClient) GetPrice(ctx context.Context, pair model.QuotesPair) (*model.PricePoint, error) {
	feedID, err := c.feedID(pair)
	if err != nil {
		return nil, err
	}
	endpoint, err := url.Parse(c.endpoint + "/v2/updates/price/latest")
	if err != nil {
		return nil, fmt.Errorf("pyth: invalid Hermes endpoint: %w", err)
	}
	query := endpoint.Query()
	query.Add("ids[]", feedID)
	query.Set("parsed", "true")
	query.Set("encoding", "base64")
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, err
	}
	if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("pyth: Hermes request failed: %s", resp.Status)
	}

	var result struct {
		Parsed []struct {
			ID    string `json:"id"`
			Price struct {
				Price       string `json:"price"`
				Expo        int32  `json:"expo"`
				PublishTime int64  `json:"publish_time"`
			} `json:"price"`
		} `json:"parsed"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("pyth: decode Hermes response: %w", err)
	}
	for _, parsed := range result.Parsed {
		if !strings.EqualFold(parsed.ID, feedID) {
			continue
		}
		if parsed.Price.Price == "" || parsed.Price.PublishTime <= 0 {
			return nil, errNoPrice
		}
		price, err := decimal.NewFromString(parsed.Price.Price)
		if err != nil {
			return nil, fmt.Errorf("pyth: invalid price: %w", err)
		}
		return &model.PricePoint{
			NewPrice:  price.Shift(parsed.Price.Expo),
			UpdatedAt: time.Unix(parsed.Price.PublishTime, 0),
		}, nil
	}
	return nil, errNoPrice
}

func (c *PythSingleClient) GetKlines(context.Context, model.QuotesPair, string, int) ([]model.PriceInterval, error) {
	return nil, ErrUnsupported
}

func (c *PythSingleClient) GetOrderBook(context.Context, model.QuotesPair, int) (*model.OrderBook, error) {
	return nil, ErrUnsupported
}

type PythClient struct {
	*PythSingleClient
	*PythStreamClient
}

func New(cfg PythConfig) *PythClient {
	return &PythClient{
		PythSingleClient: NewSingleClient(cfg),
		PythStreamClient: NewStreamClient(cfg),
	}
}

func (c *PythClient) PlaceOrder(context.Context, model.OrderRequest) (*model.OrderResult, error) {
	return nil, ErrUnsupported
}

func (c *PythClient) GetOrder(context.Context, string, string) (*model.OrderDetail, error) {
	return nil, ErrUnsupported
}

func (c *PythClient) CancelOrder(context.Context, string, string) error {
	return ErrUnsupported
}

func (c *PythClient) GetAssetBalance(context.Context, string) (*model.AssetBalance, error) {
	return nil, ErrUnsupported
}

func (c *PythClient) Close() error {
	c.client.CloseIdleConnections()
	return c.PythStreamClient.Close()
}

var _ exchange.Provider = (*PythClient)(nil)
