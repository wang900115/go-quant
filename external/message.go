package external

import (
	"fmt"
	"time"
)

type Message struct {
	Title   string
	Content string
	Level   level
	Times   time.Time
	Meta    map[string]any
}

func (m *Message) Text() string {
	emoji, ok := levelEmojis[m.Level]
	if !ok {
		return ""
	}
	return fmt.Sprintf(
		"%s *%s*\n\n%s\n\n_Time: %s_",
		emoji,
		m.Title,
		m.Content,
		m.Times.Format(time.RFC3339),
	)
}

type level int

const (
	Info level = iota
	Warn
	Error
	Critical
)

var levelEmojis = map[level]string{
	Info:     "ℹ️",
	Warn:     "⚠️",
	Error:    "❌",
	Critical: "🔥",
}

type platform string

const (
	Telegram  platform = "telegram"
	Discord   platform = "discord"
	Messenger platform = "messenger"
	Slack     platform = "slack"
)
