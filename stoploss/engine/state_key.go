package engine

import "time"

const RuntimeStateKey = "engine:state"

// TODO: Add an engine instance ID to the key when supporting multiple engines.
const (
	RuntimeStateRunning = "running"
	RuntimeStateStopped = "stopped"
)

// RuntimeState is an observability snapshot; strategy internals are not
// serialized and cannot be restored from this record.
type RuntimeState struct {
	Status        string                 `json:"status"`
	UpdatedAt     time.Time              `json:"updated_at"`
	StrategyCount int                    `json:"strategy_count"`
	Results       map[string]int64       `json:"results"`
	Metrics       map[string]interface{} `json:"metrics"`
}
