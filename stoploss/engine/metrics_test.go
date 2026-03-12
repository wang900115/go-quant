package engine

import (
	"testing"

	"github.com/wang900115/quant/model"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics()
	if m.StartTime.IsZero() {
		t.Error("StartTime should be set")
	}
	if m.TotalReceived == nil || m.TotalDropped == nil {
		t.Error("TotalReceived/TotalDropped counters not initialized")
	}
	if m.TotalReceived.Snapshot().Count() != 0 {
		t.Error("TotalReceived should start at 0")
	}
}

func TestRecordReceived_FixedStop(t *testing.T) {
	m := NewMetrics()
	m.RecordReceived()
	m.RecordChannelSend(model.FIXED, model.STOP_LOSS)

	if m.TotalReceived.Snapshot().Count() != 1 {
		t.Errorf("TotalReceived: want 1, got %d", m.TotalReceived.Snapshot().Count())
	}
	if m.FixedStopReceived.Snapshot().Count() != 1 {
		t.Errorf("FixedStopReceived: want 1, got %d", m.FixedStopReceived.Snapshot().Count())
	}
}

func TestRecordDropped_DebouncedStop(t *testing.T) {
	m := NewMetrics()
	m.RecordDropped()
	m.RecordChannelDrop(model.DEBUNCED, model.STOP_LOSS)

	if m.TotalDropped.Snapshot().Count() != 1 {
		t.Errorf("TotalDropped: want 1, got %d", m.TotalDropped.Snapshot().Count())
	}
	if m.DebouncedStopDropped.Snapshot().Count() != 1 {
		t.Errorf("DebouncedStopDropped: want 1, got %d", m.DebouncedStopDropped.Snapshot().Count())
	}
}

func TestRecordTimeout_HybridFixed(t *testing.T) {
	m := NewMetrics()
	m.RecordChannelTimeout(model.HYBRID_FIXED, "")
	if m.HybridFixedTimeout.Snapshot().Count() != 1 {
		t.Errorf("HybridFixedTimeout: want 1, got %d", m.HybridFixedTimeout.Snapshot().Count())
	}
}

func TestGetDropRate_ZeroReceived(t *testing.T) {
	m := NewMetrics()
	rate := m.GetDropRate()
	if rate != 0.0 {
		t.Errorf("GetDropRate with zero received should be 0.0, got %f", rate)
	}
}

func TestGetDropRate_Normal(t *testing.T) {
	m := NewMetrics()
	// 10 received, 2 dropped → 20%
	for i := 0; i < 10; i++ {
		m.RecordReceived()
	}
	for i := 0; i < 2; i++ {
		m.RecordDropped()
	}
	rate := m.GetDropRate()
	if rate != 20.0 {
		t.Errorf("GetDropRate: want 20.0, got %f", rate)
	}
}

func TestMetricsStats(t *testing.T) {
	m := NewMetrics()
	stats := m.Stats()
	expectedKeys := []string{
		"uptime_seconds", "total_received", "total_dropped", "drop_rate_percent", "channels",
	}
	for _, k := range expectedKeys {
		if _, ok := stats[k]; !ok {
			t.Errorf("Stats() missing key %q", k)
		}
	}
}
