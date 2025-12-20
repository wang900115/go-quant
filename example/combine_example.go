package example

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/exchange"
	"github.com/wang900115/quant/exchange/coinbase"
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/currency"
	"github.com/wang900115/quant/model/trade"
	"github.com/wang900115/quant/stoploss/engine"
	"github.com/wang900115/quant/stoploss/strategy"
)

func CombineExample() {
	// ============ Context ============
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ============ Setup Trading Pair ============
	QuotesPair := model.QuotesPair{
		ExchangeID: model.COINBASE,
		Base:       currency.BTCSymbol,
		Quote:      currency.USDTSymbol,
		Category:   trade.SPOT,
	}

	// ============ Register Provider ============
	providers := exchange.New()
	providers.Register(model.COINBASE, coinbase.New(coinbase.CoinbaseConfig{}))

	// ============ Initial Price ============
	pricePoint, err := providers.GetPrice(ctx, QuotesPair)
	if err != nil {
		panic(err)
	}

	// ============ StopLoss Engine ============
	manger := engine.New(engine.DefaultConfig())
	trailingStopStrategy, _ := strategy.NewFixedTrailingStop(
		pricePoint.NewPrice,
		decimal.NewFromFloat(0.03),
		nil,
	)
	manger.RegisterStrategy("Fixed-Trailing-Stop-3%", trailingStopStrategy)
	manger.Start()

	// ============ Stream Subscribe ============
	err = providers.SubscribeStream(QuotesPair, []string{"ticker"})
	if err != nil {
		panic(err)
	}

	// ============ Start Dispatch ============
	providers.StartStream(ctx)

	// ============ Channels ============
	ch1, ch2, ch3, err := providers.ReceiveStream(QuotesPair)
	if err != nil {
		panic(err)
	}

	// ============ Start Workers ============
	go func() {
		for p := range ch1 {
			log.Printf("Stream PricePoint: %+v\n", p)
			manger.Collect(p, func() {
				log.Printf("Warning: Channel full")
			})
		}
	}()

	go func() {
		for k := range ch2 {
			log.Printf("Stream PriceInterval: %+v\n", k)
		}
	}()

	go func() {
		for ob := range ch3 {
			log.Printf("Stream OrderBook: %+v\n", ob)
		}
	}()

	// ============ Handle Shutdown Signals ============
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Println("🛑 Received interrupt, shutting down...")

	providers.CloseProvider(QuotesPair.ExchangeID)
	manger.Stop()
	cancel()

	time.Sleep(2 * time.Second)
}
