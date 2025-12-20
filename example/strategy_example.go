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
	"fmt"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/model"
	"github.com/wang900115/quant/stoploss/engine"
	"github.com/wang900115/quant/stoploss/strategy"
)

var prices = make([]float64, 100)

// generate simulated price data
func init() {
	(rand.New(rand.NewSource(time.Now().UnixNano())))
	// Start price
	startPrice := 100.0
	// Slice for 100,000 prices
	prices[0] = startPrice

	for i := 1; i < len(prices); i++ {
		// Small random change: -1.0 to +1.0
		change := (rand.Float64()*2 - 1) // [-1,1)
		prices[i] = prices[i-1] + change

		// Optional: round to 2 decimal places
		prices[i] = math.Round(prices[i]*100) / 100
	}
}

// demonstrates usage with various strategies
func StrategyUsage() {
	fmt.Println("=== Strategy Usage Example ===")
	// Create  configuration
	config := engine.Config{
		BufferSize:    1000,
		ReadTimeout:   time.Second * 5,
		CheckInterval: time.Second * 3,
	}

	
	// Create the manager instance
	manager := engine.New(config)

	// Entry price for all strategies
	entryPrice := decimal.NewFromFloat(100.0)

	// Create callback function for all strategies
	callback := func(reason string) error {
		fmt.Printf("Strategy triggered: %s\n", reason)
		return nil
	}

	fmt.Println("Registering strategies...")

	// Register Fixed Take Profit strategies
	percentProfitStrategy, _ := strategy.NewFixedPercentProfit(
		entryPrice,
		decimal.NewFromFloat(0.08), // 8% take profit
		callback,
	)
	manager.RegisterStrategy("Fixed-Percent-Profit-8%", percentProfitStrategy)

	// Register Fixed Stop Loss strategies
	percentStopStrategy, _ := strategy.NewFixedPercentStop(
		entryPrice,
		decimal.NewFromFloat(0.05), // 5% stop loss
		callback,
	)
	manager.RegisterStrategy("Fixed-Percent-Stop-5%", percentStopStrategy)

	//  Register Risk/Reward Hybrid Strategy
	hybridStrategy, _ := strategy.NewRiskRewardRatio(
		entryPrice,
		decimal.NewFromFloat(0.03), // 3% risk
		decimal.NewFromFloat(0.09), // 9% reward
		callback,
	)
	manager.RegisterStrategy("Hybrid-Fixed-Risk-Reward-3-9%", hybridStrategy)

	// Start the manager
	log.Println("Starting manager with goroutines...")
	manager.Start()

	// Simulate price movements
	log.Println("\nSimulating price movements...")
	for i, price := range prices {
		log.Printf("\n--- Price Update #%d: $%.2f ---\n", i+1, price)
		currentPrice := decimal.NewFromFloat(price)

		// Create price point
		pricePoint := model.PricePoint{
			NewPrice:  currentPrice,
			UpdatedAt: time.Now(),
		}

		// Send price update to all goroutines with error callback
		manager.Collect(pricePoint, func() {
			log.Printf("Warning: Channel full for price $%.2f\n", price)
		})

		// Small delay to see the processing
		time.Sleep(200 * time.Millisecond)
	}

	// Let the system process for a bit more
	log.Println("\nLetting system process for 2 more seconds...")
	time.Sleep(2 * time.Second)

	// Stop the manager
	log.Println("Stopping manager...")
	manager.Stop()

	log.Println("Strategy Manager example completed!")
}
