// Package notify bridges strategy trigger events to notification bots.
package notify

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/wang900115/quant/external"
	"github.com/wang900115/quant/stoploss"
)

// DefaultSendTimeout bounds how long a callback waits to enqueue a message.
const DefaultSendTimeout = 3 * time.Second

// Env var names used by CredentialsFromEnv.
const (
	EnvTelegramToken  = "TELEGRAM_BOT_TOKEN"
	EnvTelegramChatID = "TELEGRAM_CHAT_ID"
)

// FormatTrigger renders a TriggerEvent as:
//
//	[STOP_LOSS] BTCUSDT — Hit: 29500 — Current: 29480  2026-10-07T15:30:00+08:00
func FormatTrigger(symbol string, e stoploss.TriggerEvent) string {
	return fmt.Sprintf("[%s] %s — Hit: %s — Current: %s  %s",
		strings.ToUpper(e.Category.String()),
		symbol,
		e.HitPrice.String(),
		e.CurrentPrice.String(),
		e.Time().Format(time.RFC3339),
	)
}

// TelegramCallback returns a stoploss.DefaultCallback that immediately
// enqueues a formatted alert on bot. Pass it as the last argument of any
// strategy.NewXXX constructor:
//
//	strategy.NewFixedPercentStop(entry, pct, notify.TelegramCallback(tg, "BTCUSDT"))
func TelegramCallback(bot external.Bot, symbol string) stoploss.DefaultCallback {
	return BotCallback(bot, symbol, DefaultSendTimeout)
}

// BotCallback is TelegramCallback for any external.Bot with a custom timeout.
func BotCallback(bot external.Bot, symbol string, timeout time.Duration) stoploss.DefaultCallback {
	return func(e stoploss.TriggerEvent) error {
		if bot == nil {
			return external.ErrNotRegistered
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		return bot.Send(ctx, &external.Message{
			Raw:   FormatTrigger(symbol, e),
			Title: e.Reason,
			Times: e.Time(),
			Meta: map[string]any{
				"type":          e.Category.String(),
				"symbol":        symbol,
				"hit_price":     e.HitPrice.String(),
				"current_price": e.CurrentPrice.String(),
			},
		})
	}
}

// CredentialsFromEnv reads TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID.
// TELEGRAM_CHAT_ID must be the numeric group id (e.g. -1001234567890).
func CredentialsFromEnv() external.Credentials {
	return external.Credentials{
		Token:   os.Getenv(EnvTelegramToken),
		Channel: os.Getenv(EnvTelegramChatID),
	}
}
