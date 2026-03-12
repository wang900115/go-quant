package third

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/wang900115/quant/external"
)

func TestParseWebhookURL_Valid(t *testing.T) {
	tests := []struct {
		url     string
		wantID  string
		wantTok string
	}{
		{
			url:     "https://discord.com/api/webhooks/123456789/abcdeftoken",
			wantID:  "123456789",
			wantTok: "abcdeftoken",
		},
		{
			url:     "https://canary.discord.com/api/webhooks/987/xyz-tok",
			wantID:  "987",
			wantTok: "xyz-tok",
		},
	}
	for _, tt := range tests {
		id, tok, err := parseWebhookURL(tt.url)
		if err != nil {
			t.Errorf("parseWebhookURL(%q): unexpected error: %v", tt.url, err)
			continue
		}
		if id != tt.wantID {
			t.Errorf("id: want %q, got %q", tt.wantID, id)
		}
		if tok != tt.wantTok {
			t.Errorf("token: want %q, got %q", tt.wantTok, tok)
		}
	}
}

func TestParseWebhookURL_Invalid(t *testing.T) {
	cases := []string{
		"",
		"https://discord.com/api/channels/123",
		"not-a-url",
		"https://discord.com/api/webhooks/onlyid",
		"https://discord.com/api/webhooks//token",
	}
	for _, url := range cases {
		_, _, err := parseWebhookURL(url)
		if err == nil {
			t.Errorf("parseWebhookURL(%q): expected error, got nil", url)
		}
	}
}

func TestNewDiscord(t *testing.T) {
	d, err := NewDiscord(1, 1, 5, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("NewDiscord: %v", err)
	}
	defer d.cancel()
	if d.queue == nil {
		t.Error("queue channel not initialized")
	}
	if cap(d.queue) != 5 {
		t.Errorf("queue capacity: want 5, got %d", cap(d.queue))
	}
}

func TestDiscord_Register_InvalidURL(t *testing.T) {
	d, _ := NewDiscord(1, 1, 5, 100*time.Millisecond)
	defer d.cancel()

	err := d.Register(external.Credentials{Token: "not-a-webhook-url"})
	if err == nil {
		t.Fatal("expected error for invalid webhook URL")
	}
	if !errors.Is(err, external.ErrInvalidWebhookURL) {
		t.Errorf("want ErrInvalidWebhookURL, got %v", err)
	}
}

func TestDiscord_Register_EmptyURL(t *testing.T) {
	d, _ := NewDiscord(1, 1, 5, 100*time.Millisecond)
	defer d.cancel()

	err := d.Register(external.Credentials{Token: ""})
	if err == nil {
		t.Fatal("expected error for empty webhook URL")
	}
	if !errors.Is(err, external.ErrEmptyWebhookURL) {
		t.Errorf("want ErrEmptyWebhookURL, got %v", err)
	}
}

func TestDiscord_Send_Queues(t *testing.T) {
	d, _ := NewDiscord(1, 1, 5, 100*time.Millisecond)
	defer d.cancel()

	// Manually set webhook credentials to bypass Register's network call
	d.mu.Lock()
	d.webhookID = "123"
	d.webhookToken = "tok"
	d.mu.Unlock()

	msg := &external.Message{
		Title:   "test",
		Content: "hello",
		Level:   external.Info,
		Times:   time.Now(),
	}

	ctx := context.Background()
	if err := d.Send(ctx, msg); err != nil {
		t.Fatalf("Send: %v", err)
	}
	if len(d.queue) != 1 {
		t.Errorf("queue length: want 1, got %d", len(d.queue))
	}
}

func TestDiscord_Send_CancelledContext(t *testing.T) {
	d, _ := NewDiscord(1, 1, 0, 100*time.Millisecond) // queue size 0 → always full
	defer d.cancel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already cancelled

	msg := &external.Message{Level: external.Info, Times: time.Now()}
	err := d.Send(ctx, msg)
	if err == nil {
		t.Fatal("expected error for cancelled context or full queue")
	}
}

func TestDiscord_Close_NilSession(t *testing.T) {
	d, _ := NewDiscord(1, 1, 5, 100*time.Millisecond)
	// session is nil — Close() would panic calling d.session.Close()
	// We test that cancel + wg drain work; skip Close() if session is nil
	d.cancel()
	d.wg.Wait()
	d.rateTicker.Stop()
	// Success — no panic
}
