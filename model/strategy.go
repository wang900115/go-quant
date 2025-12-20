package model

type StrategyType string

const (
	FIXED           StrategyType = "Fixed"
	DEBUNCED        StrategyType = "Debounced"
	HYBRID_DEBUNCED StrategyType = "Hybrid-Debounced"
	HYBRID_FIXED    StrategyType = "Hybrid-Fixed"
)

func (st StrategyType) String() string {
	return string(st)
}

type StrategyCategory string

const (
	STOP_LOSS   StrategyCategory = "stop_loss"
	TAKE_PROFIT StrategyCategory = "take_profit"
)

func (sc StrategyCategory) String() string {
	return string(sc)
}
