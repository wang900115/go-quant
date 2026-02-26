package external

import (
	"fmt"
	"time"
)

type Message struct {
	Title   string
	Content string
	Level   Level
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

type Level int

const (
	Info Level = iota
	Warn
	Error
	Critical
)

var levelEmojis = map[Level]string{
	Info:     "ℹ️",
	Warn:     "⚠️",
	Error:    "❌",
	Critical: "🔥",
}

type platform string

const (
	Telegram platform = "telegram"
	Discord  platform = "discord"
)
