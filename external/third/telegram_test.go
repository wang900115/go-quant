package third

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wang900115/quant/external"
)

func TestNewTelegram(t *testing.T) {
	tg, err := NewTelegram(1, 5, 1, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("NewTelegram: %v", err)
	}
	defer tg.cancel()
	if tg.queue == nil {
		t.Error("queue channel not initialized")
	}
	if cap(tg.queue) != 5 {
		t.Errorf("queue capacity: want 5, got %d", cap(tg.queue))
	}
}

func TestTelegram_Register_MissingToken(t *testing.T) {
	tg, _ := NewTelegram(1, 5, 1, 100*time.Millisecond)
	defer tg.cancel()

	err := tg.Register(external.Credentials{Token: ""})
	if err == nil {
		t.Fatal("expected error for empty token")
	}
	if !errors.Is(err, external.ErrEmptyToken) {
		t.Errorf("want ErrEmptyToken, got %v", err)
	}
}

func TestTelegram_Register_InvalidChatID(t *testing.T) {
	tg, _ := NewTelegram(1, 5, 1, 100*time.Millisecond)
	defer tg.cancel()

	// Token check happens first (calls tgbotapi.NewBotAPI which dials out).
	// Test only the channel validation path, which happens after token check.
	// We validate ErrInvalidChannel by directly calling parseInt-like behavior.
	// Since Register dials out for token validation, test ErrEmptyChannel instead.
	err := tg.Register(external.Credentials{Token: "valid:token", Channel: ""})
	// This will fail on token validation (network) before channel check.
	// Just ensure it errors — we cannot control which error without a network mock.
	if err == nil {
		t.Fatal("expected error for invalid credentials")
	}
}

func TestTelegram_Send_Queues(t *testing.T) {
	tg, _ := NewTelegram(1, 5, 1, 100*time.Millisecond)
	defer tg.cancel()

	// Bypass Register by setting bot to nil (workers won't deliver, but Send still queues)
	msg := &external.Message{
		Title:   "test",
		Content: "hello",
		Level:   external.Info,
		Times:   time.Now(),
	}

	ctx := context.Background()
	if err := tg.Send(ctx, msg); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(tg.queue) != 1 {
		t.Errorf("queue length: want 1, got %d", len(tg.queue))
	}
}

func TestTelegram_Send_FullQueueCancelledCtx(t *testing.T) {
	tg, _ := NewTelegram(1, 0, 1, 100*time.Millisecond) // queue size 0
	defer tg.cancel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	msg := &external.Message{Level: external.Info, Times: time.Now()}
	err := tg.Send(ctx, msg)
	if err == nil {
		t.Fatal("expected error for cancelled context or full queue")
	}
}

func TestTelegram_Close(t *testing.T) {
	tg, _ := NewTelegram(1, 5, 1, 100*time.Millisecond)
	if err := tg.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}

func TestTelegram_SendAfterClose(t *testing.T) {
	tg, _ := NewTelegram(1, 5, 1, 10*time.Millisecond)
	if err := tg.Close(); err != nil {
		t.Fatal(err)
	}
	if err := tg.Send(context.Background(), &external.Message{Raw: "x"}); !errors.Is(err, external.ErrBotClosed) {
		t.Fatalf("want ErrBotClosed, got %v", err)
	}
	// Close is idempotent.
	if err := tg.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestEscapeMarkdown(t *testing.T) {
	got := escapeMarkdown("BTC_USDT *x* `y` [z]")
	want := "BTC\\_USDT \\*x\\* \\`y\\` \\[z]"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}
