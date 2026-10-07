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
	"github.com/wang900115/quant/exchange/pyth"
	"github.com/wang900115/quant/external/notify"
	"github.com/wang900115/quant/external/third"
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/model/currency"
	"github.com/wang900115/quant/model/trade"
	"github.com/wang900115/quant/stoploss/engine"
	"github.com/wang900115/quant/stoploss/strategy"
)

func CombineExample() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ============ Trading Pair Setting ============
	tradingPair := model.QuotesPair{
		ExchangeID: model.PYTH,
		Base:       currency.BTCSymbol,
		Quote:      currency.USDSymbol,
		Category:   trade.SPOT,
	}

	// ============ Provider Setting ============
	accessToken := os.Getenv("PYTH_PRO_ACCESS_TOKEN")
	if accessToken == "" {
		log.Fatal("PYTH_PRO_ACCESS_TOKEN must be set to use Pyth Hermes")
	}

	providers := exchange.New()
	providers.Register(model.PYTH, pyth.New(pyth.PythConfig{
		AccessToken: accessToken,
		FeedIDs: map[string]string{
			"BTC/USD": pythBTCUSDFeedID,
		},
	}))

	if err := providers.SubscribeStream(tradingPair, []string{"ticker"}); err != nil {
		panic(err)
	}
	priceCh, intervalCh, orderBookCh, err := providers.ReceiveStream(tradingPair)
	if err != nil {
		panic(err)
	}

	// ============ Entry Price ============
	pricePoint, err := providers.GetPrice(ctx, tradingPair)
	if err != nil {
		panic(err)
	}

	// ============ External Setting ============
	tg, err := third.NewTelegram(1, 100, 3, time.Second)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := tg.Close(); err != nil {
			log.Printf("Failed to close Telegram bot: %v", err)
		}
	}()
	if err := tg.Register(notify.CredentialsFromEnv()); err != nil {
		panic(err)
	}
	callback := notify.TelegramCallback(tg, "BTC/USD")

	// ============ SLTP Engine ============
	manager := engine.New(engine.DefaultConfig())

	// Fixed percentage stop loss.
	percentStopStrategy, err := strategy.NewFixedPercentStop(
		pricePoint.NewPrice,
		decimal.NewFromFloat(0.00001),
		callback,
	)
	if err != nil {
		panic(err)
	}
	manager.RegisterStrategy("Percent-Stop-0.001%", percentStopStrategy)

	// Uncomment any strategy below to enable it. Debounced thresholds are in milliseconds.

	// Fixed percentage take profit:
	// percentProfitStrategy, err := strategy.NewFixedPercentProfit(
	// 	pricePoint.NewPrice,
	// 	decimal.NewFromFloat(0.03),
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Percent-Profit-3%", percentProfitStrategy)

	// Debounced percentage stop loss / take profit:
	// debouncedPercentStopStrategy, err := strategy.NewDebouncedPercentStop(
	// 	pricePoint.NewPrice,
	// 	decimal.NewFromFloat(0.02),
	// 	5_000,
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Debounced-Percent-Stop-2%", debouncedPercentStopStrategy)
	//
	// debouncedPercentProfitStrategy, err := strategy.NewDebouncedPercentProfit(
	// 	pricePoint.NewPrice,
	// 	decimal.NewFromFloat(0.03),
	// 	5_000,
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Debounced-Percent-Profit-3%", debouncedPercentProfitStrategy)

	// Fixed trailing stop loss / take profit:
	// trailingStopStrategy, err := strategy.NewFixedTrailingStop(
	// 	pricePoint.NewPrice,
	// 	decimal.NewFromFloat(0.03),
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Fixed-Trailing-Stop-3%", trailingStopStrategy)
	//
	// trailingProfitStrategy, err := strategy.NewFixedTrailingProfit(
	// 	pricePoint.NewPrice,
	// 	decimal.NewFromFloat(0.03),
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Fixed-Trailing-Profit-3%", trailingProfitStrategy)

	// Debounced trailing stop loss / take profit:
	// debouncedTrailingStopStrategy, err := strategy.NewTrailingDebouncedStop(
	// 	pricePoint.NewPrice,
	// 	decimal.NewFromFloat(0.03),
	// 	5_000,
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Debounced-Trailing-Stop-3%", debouncedTrailingStopStrategy)
	//
	// debouncedTrailingProfitStrategy, err := strategy.NewTrailingDebouncedProfit(
	// 	pricePoint.NewPrice,
	// 	decimal.NewFromFloat(0.03),
	// 	5_000,
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Debounced-Trailing-Profit-3%", debouncedTrailingProfitStrategy)

	// ATR strategies use an initial ATR value; call UpdateATR with fresh ATR values
	// from your market-data/indicator pipeline to keep the thresholds current.
	// atrValue := decimal.NewFromFloat(500)
	// fixedATRStopStrategy, err := strategy.NewFixedATRStop(
	// 	pricePoint.NewPrice,
	// 	atrValue,
	// 	decimal.NewFromInt(2),
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Fixed-ATR-Stop-2x", fixedATRStopStrategy)
	//
	// fixedATRProfitStrategy, err := strategy.NewFixedATRProfit(
	// 	pricePoint.NewPrice,
	// 	atrValue,
	// 	decimal.NewFromInt(2),
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Fixed-ATR-Profit-2x", fixedATRProfitStrategy)
	//
	// debouncedATRStopStrategy, err := strategy.NewDebouncedATRStop(
	// 	pricePoint.NewPrice,
	// 	atrValue,
	// 	decimal.NewFromInt(2),
	// 	5_000,
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Debounced-ATR-Stop-2x", debouncedATRStopStrategy)
	//
	// debouncedATRProfitStrategy, err := strategy.NewDebouncedATRProfit(
	// 	pricePoint.NewPrice,
	// 	atrValue,
	// 	decimal.NewFromInt(2),
	// 	5_000,
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Debounced-ATR-Profit-2x", debouncedATRProfitStrategy)

	// Moving-average strategies use an initial MA value; call SetMA with updated
	// values from your moving-average/indicator pipeline to keep thresholds current.
	// initialMA := pricePoint.NewPrice
	// fixedMAStopStrategy, err := strategy.NewFixedMovingAverageStop(
	// 	pricePoint.NewPrice,
	// 	initialMA,
	// 	decimal.NewFromFloat(0.02),
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Fixed-MA-Stop-2%", fixedMAStopStrategy)
	//
	// fixedMAProfitStrategy, err := strategy.NewFixedMovingAverageProfit(
	// 	pricePoint.NewPrice,
	// 	initialMA,
	// 	decimal.NewFromFloat(0.02),
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Fixed-MA-Profit-2%", fixedMAProfitStrategy)
	//
	// debouncedMAStopStrategy, err := strategy.NewDebouncedMovingAverageStop(
	// 	pricePoint.NewPrice,
	// 	initialMA,
	// 	decimal.NewFromFloat(0.02),
	// 	5_000,
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Debounced-MA-Stop-2%", debouncedMAStopStrategy)
	//
	// debouncedMAProfitStrategy, err := strategy.NewDebouncedMovingAverageProfit(
	// 	pricePoint.NewPrice,
	// 	initialMA,
	// 	decimal.NewFromFloat(0.02),
	// 	5_000,
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Debounced-MA-Profit-2%", debouncedMAProfitStrategy)

	// Debounced risk/reward hybrid:
	// debouncedRiskRewardStrategy, err := strategy.NewRiskRewardRatioDebounced(
	// 	pricePoint.NewPrice,
	// 	decimal.NewFromFloat(0.02),
	// 	decimal.NewFromFloat(0.03),
	// 	5_000,
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Debounced-Risk-Reward-(2:3)", debouncedRiskRewardStrategy)

	// Structure-swing hybrid (long position): stop/profit multipliers are relative
	// to entry price; the strategy updates its swing history as prices arrive.
	// structureSwingStrategy, err := strategy.NewStructureSwingStop(
	// 	14,
	// 	pricePoint.NewPrice,
	// 	decimal.NewFromFloat(0.01),
	// 	decimal.NewFromFloat(0.98),
	// 	decimal.NewFromFloat(1.03),
	// 	true,
	// 	callback,
	// )
	// if err != nil {
	// 	panic(err)
	// }
	// manager.RegisterStrategy("Structure-Swing-Long", structureSwingStrategy)

	// Fixed risk/reward hybrid:
	riskRewardStrategy, err := strategy.NewRiskRewardRatio(
		pricePoint.NewPrice,
		decimal.NewFromFloat(0.02),
		decimal.NewFromFloat(0.03),
		callback,
	)
	if err != nil {
		panic(err)
	}
	manager.RegisterStrategy("Risk-Reward-(2:3)", riskRewardStrategy)
	manager.Start()

	providers.StartStream(ctx)

	go func() {
		for p := range priceCh {
			log.Printf("Stream PricePoint: %+v\n", p)
			manager.Collect(p, func() {
				log.Printf("Warning: Channel full")
			})
		}
	}()

	go func() {
		for k := range intervalCh {
			log.Printf("Stream PriceInterval: %+v\n", k)
		}
	}()

	go func() {
		for ob := range orderBookCh {
			log.Printf("Stream OrderBook: %+v\n", ob)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigChan)

	<-sigChan
	log.Println("🛑 Received interrupt, shutting down...")

	providers.CloseProvider(tradingPair.ExchangeID)
	manager.Stop()
	cancel()
}
