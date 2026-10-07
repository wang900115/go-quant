package notify

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/wang900115/quant/external"
	"github.com/wang900115/quant/stoploss"
	"github.com/wang900115/quant/stoploss/strategy"
)

type mockBot struct {
	mu   sync.Mutex
	msgs []*external.Message
	err  error
}

func (m *mockBot) Register(external.Credentials) error { return nil }
func (m *mockBot) Close() error                        { return nil }
func (m *mockBot) Send(_ context.Context, msg *external.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.msgs = append(m.msgs, msg)
	return nil
}

func TestFormatTrigger(t *testing.T) {
	ts := time.Date(2026, 10, 7, 7, 30, 0, 0, time.UTC).Unix()
	e := stoploss.NewStopLossEvent("r", decimal.RequireFromString("29500"), decimal.RequireFromString("29480.5"), ts)
	got := FormatTrigger("BTC_USDT", e)
	want := "[STOP_LOSS] BTC_USDT — Hit: 29500 — Current: 29480.5  " + time.Unix(ts, 0).Format(time.RFC3339)
	if got != want {
		t.Fatalf("got  %q\nwant %q", got, want)
	}
}

func TestTelegramCallback_SendsRaw(t *testing.T) {
	bot := &mockBot{}
	cb := TelegramCallback(bot, "ETHUSDT")
	e := stoploss.NewTakeProfitEvent("tp", decimal.NewFromInt(1800), decimal.NewFromInt(1805), 0)
	if err := cb(e); err != nil {
		t.Fatal(err)
	}
	if len(bot.msgs) != 1 {
		t.Fatalf("want 1 msg, got %d", len(bot.msgs))
	}
	m := bot.msgs[0]
	if !strings.HasPrefix(m.Raw, "[TAKE_PROFIT] ETHUSDT — Hit: 1800 — Current: 1805") {
		t.Errorf("unexpected raw: %q", m.Raw)
	}
	if m.Text() != m.Raw {
		t.Errorf("Text() should return Raw verbatim")
	}
}

func TestTelegramCallback_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	cb := TelegramCallback(&mockBot{err: want}, "X")
	if err := cb(stoploss.NewStopLossEvent("r", decimal.Zero, decimal.Zero, 0)); !errors.Is(err, want) {
		t.Fatalf("want %v, got %v", want, err)
	}
	if err := TelegramCallback(nil, "X")(stoploss.TriggerEvent{}); !errors.Is(err, external.ErrNotRegistered) {
		t.Fatalf("nil bot: got %v", err)
	}
}

// End-to-end: a real strategy fires the callback with hit/current prices.
func TestStrategyIntegration_FixedPercentStop(t *testing.T) {
	bot := &mockBot{}
	sl, err := strategy.NewFixedPercentStop(decimal.NewFromInt(100), decimal.NewFromFloat(0.05), TelegramCallback(bot, "BTCUSDT"))
	if err != nil {
		t.Fatal(err)
	}
	if hit, _ := sl.ShouldTriggerStopLoss(decimal.NewFromInt(99)); hit {
		t.Fatal("should not trigger at 99")
	}
	hit, err := sl.ShouldTriggerStopLoss(decimal.NewFromInt(94))
	if err != nil || !hit {
		t.Fatalf("expected trigger, hit=%v err=%v", hit, err)
	}
	if len(bot.msgs) != 1 {
		t.Fatalf("want 1 msg, got %d", len(bot.msgs))
	}
	if !strings.HasPrefix(bot.msgs[0].Raw, "[STOP_LOSS] BTCUSDT — Hit: 95 — Current: 94") {
		t.Errorf("unexpected raw: %q", bot.msgs[0].Raw)
	}
}
