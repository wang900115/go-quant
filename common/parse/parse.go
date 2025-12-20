package parse

import (
	"strings"
	"time"
)

func ParseInterval(interval string) time.Duration {
	switch strings.ToLower(interval) {
	case "1m":
		return 1 * time.Minute
	case "3m":
		return 3 * time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "30m":
		return 30 * time.Minute
	case "1h":
		return 60 * time.Minute
	case "2h":
		return 120 * time.Minute
	case "4h":
		return 240 * time.Minute
	case "6h":
		return 360 * time.Minute
	case "12h":
		return 720 * time.Minute
	case "1d":
		return 1440 * time.Minute
	case "1w":
		return 7 * 24 * 60 * time.Minute
	case "1mth":
		return 30 * 24 * 60 * time.Minute
	case "3mth":
		return 90 * 24 * 60 * time.Minute
	default:
		return 1 * time.Minute
	}
}
