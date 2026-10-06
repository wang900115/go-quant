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
	"os"
	"time"

	"github.com/wang900115/quant/exchange"
	"github.com/wang900115/quant/exchange/binance"
	"github.com/wang900115/quant/exchange/coinbase"
	"github.com/wang900115/quant/exchange/okx"
	"github.com/wang900115/quant/exchange/pyth"
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/currency"
	"github.com/wang900115/quant/model/trade"
)

const pythBTCUSDFeedID = "e62df6c8b4a85fe1a67db44dc12de5db330f7ac66b72dc658afedf0f4a415b43"

func exchangeExamplePythPriceStream() {
	accessToken := os.Getenv("PYTH_PRO_ACCESS_TOKEN")
	if accessToken == "" {
		log.Printf("Set PYTH_PRO_ACCESS_TOKEN to authenticate with Pyth Hermes")
		return
	}

	provider := pyth.New(pyth.PythConfig{
		AccessToken: accessToken,
		FeedIDs: map[string]string{
			"BTC/USD": pythBTCUSDFeedID,
		},
	})
	defer provider.Close()

	providers := exchange.New()
	providers.Register(model.PYTH, provider)

	pair := model.QuotesPair{
		ExchangeID: model.PYTH,
		Base:       currency.BTCSymbol,
		Quote:      currency.USDSymbol,
		Category:   trade.SPOT,
	}
	if err := providers.SubscribeStream(pair, []string{"ticker"}); err != nil {
		log.Printf("Failed to subscribe to Pyth price stream: %v", err)
		return
	}
	priceStream, _, _, err := providers.ReceiveStream(pair)
	if err != nil {
		log.Printf("Failed to receive Pyth price stream: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	dispatchDone := make(chan error, 1)
	go func() {
		dispatchDone <- provider.Dispatch(ctx)
	}()

	select {
	case point, ok := <-priceStream:
		if !ok {
			log.Printf("Pyth price stream closed before receiving a price")
			return
		}
		log.Printf("Pyth BTC/USD price: %s (published at %s)", point.NewPrice, point.UpdatedAt.Format(time.RFC3339))
		cancel()
		if err := <-dispatchDone; err != nil {
			log.Printf("Pyth price stream stopped with error: %v", err)
		}
	case err := <-dispatchDone:
		if err != nil {
			log.Printf("Pyth price stream failed: %v", err)
		} else {
			log.Printf("Pyth price stream stopped before receiving a price")
		}
	case <-ctx.Done():
		log.Printf("Timed out waiting for a Pyth BTC/USD price update: %v", ctx.Err())
		<-dispatchDone
	}
}

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
