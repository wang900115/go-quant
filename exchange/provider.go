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

package exchange

import (
	"context"
	"errors"
	"sync"

	"github.com/wang900115/quant/model"
)

var (
	errMissingProvider = errors.New("provider not found")
)

type Provider interface {
	GetPrice(ctx context.Context, pair model.QuotesPair) (*model.PricePoint, error)
	GetKlines(ctx context.Context, pair model.QuotesPair, interval string, limit int) ([]model.PriceInterval, error)
	GetOrderBook(ctx context.Context, pair model.QuotesPair, limit int) (*model.OrderBook, error)
	SubscribeStream(pair model.QuotesPair, channel []string) error
	Dispatch(ctx context.Context) error
	ReceiveStream() (<-chan model.PricePoint, <-chan model.PriceInterval, <-chan model.OrderBook)
	PlaceOrder(ctx context.Context, req model.OrderRequest) (*model.OrderResult, error)
	GetOrder(ctx context.Context, symbol string, orderID string) (*model.OrderDetail, error)
	CancelOrder(ctx context.Context, symbol string, orderID string) error
	GetAssetBalance(ctx context.Context, asset string) (*model.AssetBalance, error)
	Close() error
}

type Providers struct {
	mu       sync.RWMutex
	registry map[model.ExchangeId]Provider
}

func New() Providers {
	return Providers{
		registry: make(map[model.ExchangeId]Provider),
	}
}

func (p *Providers) Register(exchangeID model.ExchangeId, provider Provider) {
	p.mu.Lock()
	p.registry[exchangeID] = provider
	p.mu.Unlock()
}

func (p *Providers) Unregister(exchangeID model.ExchangeId) {
	p.mu.Lock()
	delete(p.registry, exchangeID)
	p.mu.Unlock()
}

func (p *Providers) GetPrice(ctx context.Context, pair model.QuotesPair) (*model.PricePoint, error) {
	p.mu.RLock()
	provider, ok := p.registry[pair.ExchangeID]
	p.mu.RUnlock()
	if !ok {
		return nil, errMissingProvider
	}
	return provider.GetPrice(ctx, pair)
}

func (p *Providers) GetKlines(ctx context.Context, pair model.QuotesPair, interval string, limit int) ([]model.PriceInterval, error) {
	p.mu.RLock()
	provider, ok := p.registry[pair.ExchangeID]
	p.mu.RUnlock()
	if !ok {
		return nil, errMissingProvider
	}
	return provider.GetKlines(ctx, pair, interval, limit)
}

func (p *Providers) GetOrderBook(ctx context.Context, pair model.QuotesPair, limit int) (*model.OrderBook, error) {
	p.mu.RLock()
	provider, ok := p.registry[pair.ExchangeID]
	p.mu.RUnlock()
	if !ok {
		return nil, errMissingProvider
	}
	return provider.GetOrderBook(ctx, pair, limit)
}

func (p *Providers) SubscribeStream(pair model.QuotesPair, channel []string) error {
	p.mu.RLock()
	provider, ok := p.registry[pair.ExchangeID]
	p.mu.RUnlock()
	if !ok {
		return errMissingProvider
	}
	return provider.SubscribeStream(pair, channel)
}

func (p *Providers) StartStream(ctx context.Context) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	for _, provider := range p.registry {
		go provider.Dispatch(ctx)
	}
}

func (p *Providers) ReceiveStream(pair model.QuotesPair) (<-chan model.PricePoint, <-chan model.PriceInterval, <-chan model.OrderBook, error) {
	p.mu.RLock()
	provider, ok := p.registry[pair.ExchangeID]
	p.mu.RUnlock()
	if !ok {
		return nil, nil, nil, errMissingProvider
	}
	ch1, ch2, ch3 := provider.ReceiveStream()
	return ch1, ch2, ch3, nil
}

func (p *Providers) CloseProvider(exchangeID model.ExchangeId) error {
	p.mu.RLock()
	provider, ok := p.registry[exchangeID]
	p.mu.RUnlock()
	if !ok {
		return errMissingProvider
	}
	return provider.Close()
}

func (p *Providers) ListProviders() []model.Exchange {
	p.mu.RLock()
	defer p.mu.RUnlock()
	exchanges := make([]model.Exchange, 0, len(p.registry))
	for id := range p.registry {
		ex := model.GetExchange(id)
		exchanges = append(exchanges, ex)
	}
	return exchanges
}
