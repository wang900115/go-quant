package third

import (
	"context"
	"errors"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/wang900115/quant/external"
)

type Telegram struct {
	bot    *tgbotapi.BotAPI
	chatID int64

	queue      chan *external.Message
	workers    int
	maxRetry   int
	rateTicker *time.Ticker

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

func NewTelegram(token string, chatID int64, workers int, rate time.Duration) (*Telegram, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	n := &Telegram{
		bot:        bot,
		chatID:     chatID,
		queue:      make(chan *external.Message, 100),
		workers:    workers,
		maxRetry:   3,
		rateTicker: time.NewTicker(rate),
		ctx:        ctx,
		cancel:     cancel,
	}
	n.start()
	return n, nil
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
	for i := 0; i < t.maxRetry; i++ {
		err := t.senddirectly(message)
		if err == nil {
			return
		}
		backoff := time.Duration(1<<i) * time.Second
		time.Sleep(backoff)
	}
}

func (t *Telegram) Send(ctx context.Context, message *external.Message) error {
	select {
	case t.queue <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return errors.New("message queue is full")
	}
}

func (t *Telegram) senddirectly(message *external.Message) error {
	msg := tgbotapi.NewMessage(t.chatID, message.Text())
	msg.ParseMode = "Markdown"
	_, err := t.bot.Send(msg)
	return err
}

func (t *Telegram) rateLimit() {
	<-t.rateTicker.C
}

func (t *Telegram) Close() error {
	t.cancel()
	close(t.queue)
	t.wg.Wait()
	t.rateTicker.Stop()
	return nil
}

var _ external.Bot = (*Telegram)(nil)
