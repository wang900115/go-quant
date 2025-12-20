// Copyright 2025 Quantive. All rights reserved.

// Licensed under the MIT License (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// https://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package example

import (
	"context"
	"log"

	"github.com/wang900115/quant/exchange"
	"github.com/wang900115/quant/exchange/binance"
	"github.com/wang900115/quant/exchange/coinbase"
	"github.com/wang900115/quant/exchange/okx"
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/currency"
	"github.com/wang900115/quant/model/trade"
)

func exchangeExample1() {
	ps := exchange.New()
	ps.Register(model.BINANCE, binance.New(binance.BinanceConfig{}))
	ps.Register(model.COINBASE, coinbase.New(coinbase.CoinbaseConfig{}))
	ps.Register(model.OKX, okx.New(okx.OkxConfig{}))

	log.Printf("exchanges registered: %+v \n", ps.ListProviders())

	QuotesPair := model.QuotesPair{
		ExchangeID: model.BINANCE,
		Base:       currency.BTCSymbol,
		Quote:      currency.USDTSymbol,
		Category:   trade.SPOT,
	}
	pricePoint, err := ps.GetPrice(context.Background(), QuotesPair)
	if err != nil {
		log.Fatalf("Failed to get price: %v \n", err)
	}
	log.Printf("Price for %s: %+v", QuotesPair.Symbol(), *pricePoint)

	pricePoint, err = ps.GetPrice(context.Background(), QuotesPair)
	if err != nil {
		log.Fatalf("Failed to get price: %v \n", err)
	}
	log.Printf("Price for %s: %+v \n", QuotesPair.Symbol(), *pricePoint)

	klines, err := ps.GetKlines(context.Background(), QuotesPair, "1h", 10)
	if err != nil {
		log.Fatalf("Failed to get klines: %v \n", err)
	}
	log.Printf("Klines for %s: %+v \n", QuotesPair.Symbol(), klines)

	orderBook, err := ps.GetOrderBook(context.Background(), QuotesPair, 5)
	if err != nil {
		log.Fatalf("Failed to get order book: %v \n", err)
	}
	log.Printf("Order book for %s: %+v \n", QuotesPair.Symbol(), orderBook)
}

func exchangeExample2() {
	ps := exchange.New()
	ps.Register(model.BINANCE, binance.New(binance.BinanceConfig{}))
	ps.Register(model.COINBASE, coinbase.New(coinbase.CoinbaseConfig{}))
	ps.Register(model.OKX, okx.New(okx.OkxConfig{}))
	log.Printf("exchanges registered: %+v \n", ps.ListProviders())
	QuotesPair := model.QuotesPair{
		ExchangeID: model.COINBASE,
		Base:       currency.BTCSymbol,
		Quote:      currency.USDTSymbol,
		Category:   trade.SPOT,
	}
	if err := ps.SubscribeStream(QuotesPair, []string{"ticker", "order_book"}); err != nil {
		log.Fatalf("Failed to subscribe to stream: %v \n", err)
	}
	ps.StartStream(context.Background())

	ch1, ch2, ch3, err := ps.ReceiveStream(QuotesPair)
	if err != nil {
		log.Fatalf("Failed to receive stream: %v \n", err)
	}

	go func() {
		for pricePoint := range ch1 {
			log.Printf("Stream PricePoint: %+v \n", pricePoint)
		}
	}()

	go func() {
		for priceInterval := range ch2 {
			log.Printf("Stream PriceInterval: %+v \n", priceInterval)
		}
	}()

	go func() {
		for orderBook := range ch3 {
			log.Printf("Stream OrderBook: %+v \n", orderBook)
		}
	}()
	select {}
}
