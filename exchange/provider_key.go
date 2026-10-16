package exchange

import (
	"fmt"

	"github.com/wang900115/quant/model"
)

// Exchange cache keys use the format:
// exchange:<operation>:<exchange>:<category>:<symbol>[:<operation parameters>].
// Values are JSON-encoded model.PricePoint, []model.PriceInterval, or
// model.OrderBook records, respectively. These keys cache the latest result
// for each query; they do not represent historical price storage.
const exchangeCacheKeyPrefix = "exchange:"

// PriceCacheKey identifies the latest price returned for a trading pair.
func PriceCacheKey(pair model.QuotesPair) []byte {
	return []byte(fmt.Sprintf("%sprice:%s:%s:%s", exchangeCacheKeyPrefix, exchangeName(pair.ExchangeID), pair.Category, pair.Symbol()))
}

// KlinesCacheKey identifies a candlestick query, including its interval and limit.
func KlinesCacheKey(pair model.QuotesPair, interval string, limit int) []byte {
	return []byte(fmt.Sprintf("%sklines:%s:%s:%s:%s:%d", exchangeCacheKeyPrefix, exchangeName(pair.ExchangeID), pair.Category, pair.Symbol(), interval, limit))
}

// OrderBookCacheKey identifies an order book query, including its depth limit.
func OrderBookCacheKey(pair model.QuotesPair, limit int) []byte {
	return []byte(fmt.Sprintf("%sorderbook:%s:%s:%s:%d", exchangeCacheKeyPrefix, exchangeName(pair.ExchangeID), pair.Category, pair.Symbol(), limit))
}

func exchangeName(id model.ExchangeId) string {
	if exchange := model.GetExchange(id); exchange.Name != "" {
		return string(exchange.Name)
	}
	return fmt.Sprintf("exchange-%d", id)
}
