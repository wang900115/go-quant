# Stop Loss Strategy Test Documentation

## Test Architecture Overview

This project's stop loss strategy tests use unified mock test data, with all test data defined in `testdata.go`.

## File Structure

```
stoploss/strategy/
├── testdata.go                    # Unified test data definitions
├── atr.go                         # ATR stop loss strategy implementation
├── atr_test.go                    # ATR stop loss tests
├── fixed-percentcile.go           # Fixed percentage stop loss implementation
├── fixed-percentcile_test.go      # Fixed percentage stop loss tests
├── trailing.go                    # Trailing stop loss implementation
├── trailing_test.go               # Trailing stop loss tests
├── trailing-Debounced.go          # Time-based trailing stop loss implementation
└── trailing-Debounced_test.go     # Time-based trailing stop loss tests
```

## Unified Test Data (testdata.go)

### Data Structure

```go
type PriceData struct {
    High   decimal.Decimal  // Highest price
    Low    decimal.Decimal  // Lowest price
    Close  decimal.Decimal  // Closing price
    ATR    decimal.Decimal  // Average True Range
    Period int              // Period number
}
```

### Test Scenarios

#### 1. `GetMockHistoricalData()` - Downtrend
- **Purpose**: Test stop loss triggering
- **Description**: Simulates BTC rising from 49500 to 51800, then sharply declining to 47500
- **Drop**: Approximately -4%
- **Use Case**: Verify stop loss triggers correctly during downtrend

#### 2. `GetMockConsolidationData()` - Consolidation Market
- **Purpose**: Test false trigger prevention
- **Description**: Price oscillates in 98-102 range
- **Volatility**: ±2%
- **Use Case**: Verify stop loss doesn't trigger during normal oscillation

#### 3. `GetMockTrendingData()` - Uptrend
- **Purpose**: Test uptrend behavior
- **Description**: Price continuously rises from 99 to 114
- **Gain**: Approximately +15%
- **Use Case**: Verify stop loss remains active during uptrend

#### 4. `GetMockVolatileData()` - High Volatility Market
- **Purpose**: Test extreme volatility
- **Description**: Price fluctuates dramatically, ATR increases from 3.5 to 5.5
- **Use Case**: Test ATR strategy response to volatility

#### 5. `GetMockGradualDeclineData()` - Slow Decline
- **Purpose**: Test gradual stop loss breach
- **Description**: Price gradually declines from 99.5 to 95
- **Drop**: Approximately -0.5% per period
- **Use Case**: Test fixed percentage stop loss sensitivity

#### 6. `GetMockSharpDropData()` - Sharp Drop
- **Purpose**: Test rapid drop response
- **Description**: Price suddenly plunges from 100.5 to 91
- **Drop**: Single period -9%
- **Use Case**: Test stop loss execution during flash crashes

#### 7. `GetMockRecoveryData()` - Post-Drop Recovery
- **Purpose**: Test stop loss reset functionality
- **Description**: Price drops to 95 then recovers to 102
- **Use Case**: Test ReSet functionality

## Strategy Comparison

### ATR Stop Loss (atr_test.go)

**Characteristics**:
- Dynamically adjusts stop loss position
- Based on market volatility (ATR)
- Suitable for volatile markets

**Test Coverage**:
- ✅ Creation validation (valid/invalid parameters)
- ✅ Stop loss calculation
- ✅ ATR updates
- ✅ Trigger detection
- ✅ Reset functionality
- ✅ Deactivation functionality
- ✅ Downtrend scenario
- ✅ Consolidation scenario
- ✅ Uptrend scenario
- ✅ Performance benchmarks

**Key Formula**:
```
Stop Loss Price = Entry Price - (ATR × Multiplier)
```

### Fixed Percentage Stop Loss (fixed-percentcile_test.go)

**Characteristics**:
- Fixed percentage
- Simple and intuitive
- Suitable for stable markets

**Test Coverage**:
- ✅ Creation validation (valid/invalid percentages)
- ✅ Stop loss calculation
- ✅ Trigger detection
- ✅ Reset functionality
- ✅ Deactivation functionality
- ✅ Downtrend scenario
- ✅ Consolidation scenario
- ✅ Uptrend scenario
- ✅ Gradual decline scenario
- ✅ Sharp drop scenario
- ✅ Percentage comparison
- ✅ Performance benchmarks

**Key Formula**:
```
Stop Loss Price = Entry Price × (1 - Stop Loss Percentage)
```

### Trailing Stop Loss (trailing_test.go)

**Characteristics**:
- Stop loss trails up as price increases
- Never decreases, only increases
- Locks in profits during uptrends

**Test Coverage**:
- ✅ Creation validation (valid/invalid rates)
- ✅ Stop loss calculation and trailing behavior
- ✅ Trigger detection
- ✅ Reset functionality
- ✅ Deactivation functionality
- ✅ Uptrend scenario (trails without triggering)
- ✅ Downtrend scenario (triggers correctly)
- ✅ Consolidation scenario
- ✅ Volatile market scenario
- ✅ Percentage comparison (3%, 5%, 10%)
- ✅ Performance benchmarks

**Key Formula**:
```
Stop Loss Price = max(Previous Stop Loss, Current Price × (1 - Stop Loss Rate))
```

### Time-Based Trailing Stop Loss (trailing-Debounced_test.go)

**Characteristics**:
- Combines trailing stop with time confirmation
- Requires price to stay below stop for time threshold before triggering
- Prevents false triggers from short-term wicks
- Timer resets when price recovers above stop loss

**Test Coverage**:
- ✅ Creation validation (valid/invalid parameters and time thresholds)
- ✅ Stop loss calculation (inherits trailing behavior)
- ✅ Time threshold behavior (needs continuous breach)
- ✅ Price recovery resets timer
- ✅ Trigger detection with timestamp simulation
- ✅ Reset functionality
- ✅ Deactivation functionality
- ✅ Uptrend scenario with time simulation
- ✅ Gradual decline scenario (5 min periods)
- ✅ Sharp drop scenario (3 min periods)
- ✅ Recovery scenario (timer reset testing)
- ✅ Time threshold comparison (1min, 5min, 10min)
- ✅ Performance benchmarks

**Key Formulas**:
```
Stop Loss Price = max(Previous Stop Loss, Current Price × (1 - Stop Loss Rate))
Trigger Condition = (Price ≤ Stop Loss) AND (Duration ≥ Time Threshold)
Timer Reset = Price > Stop Loss
```

## Usage Examples

### 1. Using Unified Test Data

```go
func TestMyStrategy(t *testing.T) {
    // Use downtrend data
    data := GetMockHistoricalData()
    
    // Use first period as entry point
    entryData := data[0]
    entryPrice := entryData.Close
    
    // Create strategy
    s, _ := NewMyStrategy(entryPrice)
    
    // Simulate each period
    for i := 1; i < len(data); i++ {
        period := data[i]
        // Test logic...
    }
}
```

### 2. Time Simulation for Time-Based Tests

```go
func TestDebouncedStrategy(t *testing.T) {
    data := GetMockHistoricalData()
    entryPrice := data[0].Close
    timeThreshold := int64(300) // 5 minutes
    
    s, _ := NewTrailingDebounced(entryPrice, d(0.05), timeThreshold, nil)
    
    baseTime := int64(1000000) // Starting timestamp
    
    for i := 1; i < len(data); i++ {
        period := data[i]
        timestamp := baseTime + int64(i*300) // 5 minutes per period
        
        stopLoss, _ := s.CalculateStopLoss(period.Close)
        triggered, _ := s.ShouldTriggerStopLoss(period.Low, timestamp)
        
        if triggered {
            t.Logf("Triggered at period %d (T+%dm)", period.Period, (timestamp-baseTime)/60)
            break
        }
    }
}
```

### 3. Adding New Test Scenarios

Add new function in `testdata.go`:

```go
func GetMockMyScenarioData() []PriceData {
    return []PriceData{
        {High: d(100), Low: d(98), Close: d(99), ATR: d(2), Period: 1},
        // ... more data
    }
}
```

### 4. Quick Decimal Creation

Use helper function `d()`:

```go
price := d(100.5)  // Equivalent to decimal.NewFromFloat(100.5)
```

## Running Tests

```bash
# Run all stop loss strategy tests
go test ./stoploss/strategy -v

# Run specific test
go test ./stoploss/strategy -v -run TestATRStopLoss_WithHistoricalData

# Run benchmark tests
go test ./stoploss/strategy -bench=. -benchmem

# View test coverage
go test ./stoploss/strategy -cover

# Run trailing stop tests
go test ./stoploss/strategy -v -run TestTrailing

# Run time-based tests
go test ./stoploss/strategy -v -run TestTrailingDebounced
```

## Test Best Practices

1. **Use Unified Data**: All tests use data defined in `testdata.go`
2. **Descriptive Naming**: Test function names clearly describe scenarios
3. **Detailed Logging**: Use `t.Logf()` to output key data
4. **Independent Tests**: Each test case recreates strategy instances
5. **Boundary Testing**: Cover valid and invalid inputs
6. **Performance Testing**: Include benchmarks to ensure performance
7. **Time Simulation**: For time-based strategies, simulate timestamps incrementally
8. **Timer Reset Testing**: Verify timer resets when conditions change

## Adding New Strategy Tests

Checklist when creating new strategy tests:

- [ ] Risk-Reward Ratio
- [ ] Drawdown-based
- [ ] Moving Average 
- [ ] Structe / Swing
- [ ] Event-based
- [ ] Composite Weighted


## Test Data Maintenance

When updating test data:

1. Only modify `testdata.go`
2. Ensure all related tests still pass
3. Update this documentation with new data purposes
4. Consider backward compatibility

## Notes

- All price data uses `decimal.Decimal` for precise calculation
- Test data simulates real markets but is simplified
- Period numbering starts from 1
- Using Low price for stop loss trigger checks is more realistic
- Callback functions in tests typically return nil
- Time-based tests use Unix timestamps (seconds)
- Time intervals in tests: typically 300s (5 min), 600s (10 min), or 180s (3 min) per period
