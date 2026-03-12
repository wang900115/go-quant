package external

import "context"

// Credentials holds the authentication information for a bot platform.
// For Telegram: Token = bot token, Channel = chat ID (numeric string).
// For Discord:  Token = webhook URL, Channel can be left empty.
type Credentials struct {
	Token   string            // Bot token, API key, or webhook URL
	Secret  string            // API secret when required by the platform
	Channel string            // Target channel / chat ID
	Extra   map[string]string // Additional platform-specific configuration
}

// Bot defines the interface for notification bots.
type Bot interface {
	// Register configures (or reconfigures) the bot with the given credentials.
	Register(creds Credentials) error

	// Send enqueues a message for delivery.
	Send(ctx context.Context, message *Message) error

	// Close gracefully shuts down the bot.
	Close() error
}
