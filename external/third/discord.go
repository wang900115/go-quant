package third

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/wang900115/quant/external"
)

// Discord embed colours per message level.
var discordLevelColors = map[external.Level]int{
	external.Info:     0x3498DB, // blue
	external.Warn:     0xF39C12, // orange
	external.Error:    0xE74C3C, // red
	external.Critical: 0x8E44AD, // purple
}

// Discord sends notifications to a Discord channel via an Incoming Webhook
// using github.com/bwmarrin/discordgo.
//
// Webhook URL format: https://discord.com/api/webhooks/{id}/{token}
type Discord struct {
	mu           sync.RWMutex
	session      *discordgo.Session
	webhookID    string
	webhookToken string

	queue      chan *external.Message
	workers    int
	maxRetry   int
	rateTicker *time.Ticker

	wg     sync.WaitGroup
	ctx    context.Context
	cancel context.CancelFunc
}

// NewDiscord creates and starts a Discord webhook worker pool.
// webhookURL must be a full Discord Incoming Webhook URL:
// https://discord.com/api/webhooks/{id}/{token}
func NewDiscord(workers, maxRetry, maxSize int, rate time.Duration) (*Discord, error) {
	ctx, cancel := context.WithCancel(context.Background())
	d := &Discord{
		queue:      make(chan *external.Message, maxSize),
		workers:    workers,
		maxRetry:   maxRetry,
		rateTicker: time.NewTicker(rate),
		ctx:        ctx,
		cancel:     cancel,
	}
	d.start()
	return d, nil
}

// Register (re)configures the webhook credentials at runtime without restarting workers.
// creds.Token = full Discord Incoming Webhook URL.
func (d *Discord) Register(creds external.Credentials) error {
	id, token, err := parseWebhookURL(creds.Token)
	if err != nil {
		return err
	}
	session, err := discordgo.New("")
	if err != nil {
		return fmt.Errorf("discord: create session: %w", err)
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.session = session
	d.webhookID = id
	d.webhookToken = token
	return nil
}

func (d *Discord) start() {
	for i := 0; i < d.workers; i++ {
		d.wg.Add(1)
		go d.worker()
	}
}

func (d *Discord) worker() {
	defer d.wg.Done()
	for {
		select {
		case message, ok := <-d.queue:
			if !ok {
				return
			}
			if message == nil {
				continue
			}
			d.rateLimit()
			d.sendWithRetry(message)
		case <-d.ctx.Done():
			return
		}
	}
}

func (d *Discord) sendWithRetry(message *external.Message) {
	for i := 0; i < d.maxRetry; i++ {
		err := d.sendDirectly(message)
		if err == nil {
			return
		}
		time.Sleep(time.Duration(1<<i) * time.Second)
	}
}

func (d *Discord) Send(ctx context.Context, message *external.Message) error {
	select {
	case d.queue <- message:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	default:
		return external.ErrQueueFull
	}
}

func (d *Discord) sendDirectly(message *external.Message) error {
	d.mu.RLock()
	session := d.session
	id := d.webhookID
	token := d.webhookToken
	d.mu.RUnlock()

	if session == nil || id == "" {
		return external.ErrNotRegistered
	}

	color := discordLevelColors[message.Level]

	embed := &discordgo.MessageEmbed{
		Title:       message.Title,
		Description: message.Content,
		Color:       color,
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Time: " + message.Times.Format("2006-01-02 15:04:05 MST"),
		},
	}

	params := &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	}

	_, err := session.WebhookExecute(id, token, false, params)
	if err != nil {
		return fmt.Errorf("discord: webhook execute: %w", err)
	}
	return nil
}

func (d *Discord) rateLimit() {
	<-d.rateTicker.C
}

func (d *Discord) Close() error {
	d.cancel()
	close(d.queue)
	d.wg.Wait()
	d.rateTicker.Stop()
	return d.session.Close()
}

// parseWebhookURL extracts the webhook ID and token from a Discord webhook URL.
// Expected format: https://discord.com/api/webhooks/{id}/{token}
func parseWebhookURL(webhookURL string) (id, token string, err error) {
	if webhookURL == "" {
		return "", "", external.ErrEmptyWebhookURL
	}
	const marker = "/webhooks/"
	idx := strings.Index(webhookURL, marker)
	if idx == -1 {
		return "", "", external.ErrInvalidWebhookURL
	}
	rest := webhookURL[idx+len(marker):]
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", external.ErrInvalidWebhookURL
	}
	return parts[0], parts[1], nil
}

var _ external.Bot = (*Discord)(nil)
