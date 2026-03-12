package external

import "errors"

// Sentinel errors for the external bot subsystem.
// Callers can match them with errors.Is().
var (
	// ErrQueueFull is returned by Send when the internal message queue
	// has no remaining capacity.
	ErrQueueFull = errors.New("message queue is full")

	// ErrEmptyToken is returned by Register when the Token field is blank.
	ErrEmptyToken = errors.New("token must not be empty")

	// ErrEmptyWebhookURL is returned by Register when the webhook URL is blank.
	ErrEmptyWebhookURL = errors.New("webhook URL must not be empty")

	// ErrInvalidWebhookURL is returned when the webhook URL cannot be parsed
	// into a valid ID/token pair.
	ErrInvalidWebhookURL = errors.New("invalid webhook URL")

	// ErrEmptyChannel is returned by Register when the Channel field is blank.
	ErrEmptyChannel = errors.New("channel must not be empty")

	// ErrInvalidChannel is returned when the channel value cannot be parsed
	// (e.g. a Telegram chat ID that is not a number).
	ErrInvalidChannel = errors.New("invalid channel")

	// ErrNotRegistered is returned by Send / sendDirectly when Register has
	// not been called yet, meaning the bot has no credentials configured.
	ErrNotRegistered = errors.New("bot credentials not registered; call Register first")
)
