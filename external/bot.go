package external

import "context"

type Bot interface {
	Send(ctx context.Context, message *Message) error

	Close() error
}
