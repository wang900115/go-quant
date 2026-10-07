package third

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wang900115/quant/external"
)

type Telegram struct {
	mu     sync.RWMutex
	bot    *tgbotapi.BotAPI
	chatID int64

	queue      chan *external.Message
	workers    int
	maxRetry   int
	rateTicker *time.Ticker

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc

	closeOnce sync.Once
	closed    bool // guarded by mu
}

// NewTelegram creates and starts a Telegram bot worker pool.
func NewTelegram(workers, maxSize, maxRetry int, rate time.Duration) (*Telegram, error) {
	ctx, cancel := context.WithCancel(context.Background())
	n := &Telegram{
		queue:      make(chan *external.Message, maxSize),
		workers:    workers,
		maxRetry:   maxRetry,
		rateTicker: time.NewTicker(rate),
		ctx:        ctx,
		cancel:     cancel,
	}
	n.start()
	return n, nil
}

// Register (re)configures the bot credentials at runtime without restarting workers.
// creds.Token   = Telegram bot token.
// creds.Channel = target chat ID (numeric string).
func (t *Telegram) Register(creds external.Credentials) error {
	if creds.Token == "" {
		return external.ErrEmptyToken
	}

	bot, err := tgbotapi.NewBotAPI(creds.Token)
	if err != nil {
		return fmt.Errorf("telegram: invalid token: %w", err)
	}

	if creds.Channel == "" {
		return external.ErrEmptyChannel
	}
	chatID, err := strconv.ParseInt(creds.Channel, 10, 64)
	if err != nil {
		return external.ErrInvalidChannel
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	t.bot = bot
	t.chatID = chatID
	return nil
}

func (t *Telegram) start() {
	for i := 0; i < t.workers; i++ {
		t.wg.Add(1)
		go t.worker()
	}
}

func (t *Telegram) worker() {
	defer t.wg.Done()
	for {
		select {
		case message, ok := <-t.queue:
			if !ok {
				return
			}
			if message == nil {
				continue
			}

			t.rateLimit()
			t.sendWithRetry(message)
		case <-t.ctx.Done():
			return
		}
	}
}

func (t *Telegram) sendWithRetry(message *external.Message) {
	attempts := t.maxRetry
	if attempts < 1 {
		attempts = 1
	}
	var err error
	for i := 0; i < attempts; i++ {
		if err = t.senddirectly(message); err == nil {
			return
		}
		if i == attempts-1 {
			break
		}
		backoff := time.Duration(1<<i) * time.Second
		select {
		case <-time.After(backoff):
		case <-t.ctx.Done():
			log.Printf("telegram: shutdown during retry, message dropped: %v", err)
			return
		}
	}
	log.Printf("telegram: giving up after %d attempts: %v", attempts, err)
}

func (t *Telegram) Send(ctx context.Context, message *external.Message) error {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if t.closed {
		return external.ErrBotClosed
	}
	select {
	case t.queue <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return external.ErrQueueFull
	}
}

func (t *Telegram) senddirectly(message *external.Message) error {
	t.mu.RLock()
	bot := t.bot
	chatID := t.chatID
	t.mu.RUnlock()

	if bot == nil {
		return external.ErrNotRegistered
	}

	var msg tgbotapi.MessageConfig
	if message.Raw != "" {
		// Raw messages go out as plain text so symbols like BTC_USDT are
		// never mis-parsed as Markdown.
		msg = tgbotapi.NewMessage(chatID, message.Raw)
	} else {
		safe := *message
		safe.Title = escapeMarkdown(message.Title)
		safe.Content = escapeMarkdown(message.Content)
		msg = tgbotapi.NewMessage(chatID, safe.Text())
		msg.ParseMode = "Markdown"
	}
	_, err := bot.Send(msg)
	return err
}

func (t *Telegram) rateLimit() {
	<-t.rateTicker.C
}

// Close stops accepting new messages, lets workers drain whatever is
// already queued, then shuts down. Safe to call more than once.
func (t *Telegram) Close() error {
	t.closeOnce.Do(func() {
		t.mu.Lock()
		t.closed = true
		close(t.queue) // workers exit after draining remaining messages
		t.mu.Unlock()

		t.wg.Wait()
		t.cancel()
		t.rateTicker.Stop()
	})
	return nil
}

// markdownEscaper escapes Telegram legacy-Markdown control characters.
var markdownEscaper = strings.NewReplacer("_", "\\_", "*", "\\*", "`", "\\`", "[", "\\[")

func escapeMarkdown(s string) string { return markdownEscaper.Replace(s) }

var _ external.Bot = (*Telegram)(nil)
