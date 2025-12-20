# Strategy Manager Result Processing Flow

This document explains the complete flow from strategy execution to result processing in the concurrent strategy manager.

## Architecture Overview

The manager uses **6 goroutines** (not 4) to handle different strategy types:

1. **Goroutine 1**: Fixed Stop Loss strategies
2. **Goroutine 2**: Debounced Stop Loss strategies  
3. **Goroutine 3**: Fixed Take Profit strategies
4. **Goroutine 4**: Debounced Take Profit strategies
5. **Goroutine 5**: Fixed Hybrid strategies (without time)
6. **Goroutine 6**: Debounced Hybrid strategies (with time)

## Data Flow

```
Price Update → Manager.Collect() → Individual Channels → Processing Goroutines → Result Channels → Consumers
```

### Step 1: Price Collection
```go
// Send price to all 6 channels
pricePoint := model.PricePoint{
    NewPrice:  decimal.NewFromFloat(100.0),
    UpdatedAt: time.Now(),
}

manager.Collect(pricePoint, func() {
    fmt.Println("Channel full warning")
})
```

### Step 2: Strategy Processing
Each goroutine processes its assigned strategies:

```go
// Example: Fixed Stop Loss processing
func (csm *Manager) processFixedStopStrategies(update model.PricePoint) {
    strategies := csm.portfolio.GetFixedStoplossStrategies()
    for name, strategy := range strategies {
        newThreshold, err := strategy.CalculateStopLoss(update.NewPrice)
        if err == nil {
            result := result.NewGeneral(name, "Fixed", "StopLoss", 
                update.NewPrice, newThreshold, update.UpdatedAt, time.Duration(0))
            
            shouldTrigger, err := strategy.ShouldTriggerStopLoss(update.NewPrice)
            if err == nil {
                result.SetTriggered(shouldTrigger)
            } else {
                result.SetError(err)
            }
            
            // Send to result channel
            csm.execution.generalResults <- *result
        }
    }
}
```

### Step 3: Result Consumption
```go
// Get result channels
generalResults, hybridResults := manager.GetResultChannels()

// Consume general results
go func() {
    for result := range generalResults {
        if result.Triggered {
            fmt.Printf("🚨 TRIGGERED: %s at %s\n", 
                result.StrategyName, result.LastPrice.String())
        } else {
            fmt.Printf("✅ UPDATED: %s threshold: %s\n", 
                result.StrategyName, result.Stat.PriceThreshold.String())
        }
    }
}()

// Consume hybrid results  
go func() {
    for result := range hybridResults {
        if result.Triggered {
            fmt.Printf("🚨 HYBRID TRIGGERED: %s\n", result.StrategyName)
        }
    }
}()
```

## Result Types

### General Result
Used for single-purpose strategies (stop loss OR take profit):

```go
type StrategyGeneralResult struct {
    StrategyName  string          // "Fixed-Percent-5%"
    StrategyType  string          // "Fixed" or "Debounced"  
    Triggered     bool            // true if threshold hit
    TriggerType   string          // "StopLoss" or "TakeProfit"
    LastPrice     decimal.Decimal // Current price
    Stat          StrategyStat    // Contains PriceThreshold
    LastTime      time.Time       // Update timestamp
    TimeThreshold time.Duration   // For Debounced strategies
    Error         error           // Any processing error
}
```

### Hybrid Result
Used for combined stop loss AND take profit strategies:

```go
type StrategyHybridResult struct {
    StrategyName  string          // "Risk-Reward-1:2"
    StrategyType  string          // "Fixed" or "Debounced"
    Triggered     bool            // true if either SL or TP hit
    TriggerType   string          // "Hybrid"
    LastTime      time.Time       // Update timestamp
    LastPrice     decimal.Decimal // Current price
    TimeThreshold time.Duration   // For Debounced strategies
    stopStat      StrategyStat    // Stop loss threshold
    profitStat    StrategyStat    // Take profit threshold
    Error         error           // Any processing error
}
```

## Complete Usage Example

```go
func CompleteExample() {
    // 1. Create and configure manager
    config := ManagerConfig{
        BufferSize:    1000,
        ReadTimeout:   time.Second * 5,
        CheckInterval: time.Millisecond * 100,
    }
    manager := New(config)
    
    // 2. Register strategies
    entryPrice := decimal.NewFromFloat(100.0)
    callback := func(reason string) error { return nil }
    
    stopStrategy, _ := strategy.NewFixedPercentStop(entryPrice, decimal.NewFromFloat(0.05), callback)
    manager.RegisterStrategy("Stop-5%", stopStrategy)
    
    profitStrategy, _ := strategy.NewFixedPercentProfit(entryPrice, decimal.NewFromFloat(0.08), callback)
    manager.RegisterStrategy("Profit-8%", profitStrategy)
    
    hybridStrategy, _ := strategy.NewRiskRewardRatio(entryPrice, decimal.NewFromFloat(0.03), decimal.NewFromFloat(0.06), callback)
    manager.RegisterStrategy("RiskReward-1:2", hybridStrategy)
    
    // 3. Start manager (launches 6 goroutines)
    manager.Start()
    
    
    // 4. Send price updates
    for _, price := range []float64{100, 102, 98, 105, 95} {
        pricePoint := model.PricePoint{
            NewPrice:  decimal.NewFromFloat(price),
            UpdatedAt: time.Now(),
        }
        
        manager.Collect(pricePoint, func() {
            fmt.Println("Channel full")
        })
        
        time.Sleep(100 * time.Millisecond)
    }
    
    // 5. Clean shutdown
    manager.Stop()
}
```

## Performance Characteristics

### Concurrent Processing
- **6 goroutines** process strategies in parallel
- **Individual channels** prevent blocking between strategy types
- **Buffered channels** handle burst traffic

### Result Processing
- **2 result channels** (general + hybrid) consolidate outputs
- **Non-blocking sends** prevent strategy processing delays
- **Consumer flexibility** allows custom result handling

### Error Handling
- **Per-strategy error tracking** in results
- **Channel overflow callbacks** for monitoring
- **Graceful degradation** when channels full

## Best Practices

### 1. Buffer Sizing
```go
// For high-frequency trading
config := ManagerConfig{
    BufferSize: 10000,
    CheckInterval: time.Millisecond * 10,
}

// For lower frequency
config := ManagerConfig{
    BufferSize: 100,
    CheckInterval: time.Millisecond * 100,
}
```

### 2. Result Processing 
```go
// Always consume results to prevent blocking
manager.Start()
```

### 3. Error Monitoring
```go
// Track errors and performance
type Report struct {
	generalCount int64
	hybridCount  int64
	triggerCount int64
	errorCount   int64
}

func (rp *Report) Stats() map[string]int64 {
	return map[string]int64{
		"general_results": rp.generalCount,
		"hybrid_results":  rp.hybridCount,
		"triggers":        rp.triggerCount,
		"errors":          rp.errorCount,
	}
}

```

### 4. Graceful Shutdown
```go
// Always stop gracefully
defer manager.Stop()

// Or with timeout
go func() {
    time.Sleep(5 * time.Second)
    manager.Stop()
}()
```

## Result Processing Patterns

### 1. Real-time Processing
Process results immediately as they arrive.

### 2. Batch Processing
Collect results and process in batches for efficiency.

### 3. Filtering and Routing
Route different result types to different handlers.

### 4. Aggregation and Analytics
Combine results for portfolio-level analytics.

This architecture provides scalable, concurrent strategy processing with comprehensive result handling for trading systems.