package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/wang900115/quant/exchange/binance"
	"github.com/wang900115/quant/exchange/coinbase"
	"github.com/wang900115/quant/exchange/okx"
	"github.com/wang900115/quant/external"
	"github.com/wang900115/quant/stoploss/engine"
)

type Config struct {
	// Strategy configurations
	Engine engine.Config

	// Exchange configurations
	Binance  binance.BinanceConfig
	Coinbase coinbase.CoinbaseConfig
	Okx      okx.OkxConfig

	// Notification bot configurations
	Discord  external.Credentials
	Telegram external.Credentials
}

type configOpts func(c *Config)

func (c *Config) apply(opts ...configOpts) {
	for _, opt := range opts {
		opt(c)
	}
}

// New creates a Config with functional options applied.
func New(opts ...configOpts) *Config {
	c := &Config{}
	c.apply(opts...)
	return c
}

// NewFromEnv builds a Config from environment variables.
// Returns an error only on parse failure; missing vars are silently skipped.
func NewFromEnv() (*Config, error) {
	c := &Config{}

	if v := os.Getenv("ENGINE_BUFFER_SIZE"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("ENGINE_BUFFER_SIZE: %w", err)
		}
		c.Engine.BufferSize = n
	}
	if v := os.Getenv("ENGINE_CHECK_INTERVAL"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("ENGINE_CHECK_INTERVAL: %w", err)
		}
		c.Engine.CheckInterval = d
	}

	c.Binance.APIKey = os.Getenv("BINANCE_API_KEY")
	c.Binance.SecretKey = os.Getenv("BINANCE_SECRET")
	c.Coinbase.APIKey = os.Getenv("COINBASE_API_KEY")
	c.Coinbase.SecretKey = os.Getenv("COINBASE_SECRET")
	c.Okx.APIKey = os.Getenv("OKX_API_KEY")
	c.Okx.SecretKey = os.Getenv("OKX_SECRET")
	c.Discord.Token = os.Getenv("DISCORD_WEBHOOK_URL")
	c.Telegram.Token = os.Getenv("TELEGRAM_BOT_TOKEN")
	c.Telegram.Channel = os.Getenv("TELEGRAM_CHAT_ID")

	return c, nil
}

// Validate returns an error if the config contains inconsistent credential pairs.
func (c *Config) Validate() error {
	if (c.Binance.APIKey != "") != (c.Binance.SecretKey != "") {
		return fmt.Errorf("binance: APIKey and SecretKey must both be set or both be empty")
	}
	if (c.Coinbase.APIKey != "") != (c.Coinbase.SecretKey != "") {
		return fmt.Errorf("coinbase: APIKey and SecretKey must both be set or both be empty")
	}
	if (c.Okx.APIKey != "") != (c.Okx.SecretKey != "") {
		return fmt.Errorf("okx: APIKey and SecretKey must both be set or both be empty")
	}
	return nil
}

func (c *Config) WithEngine(opt engine.Config) configOpts {
	return func(c *Config) {
		c.Engine = opt
	}
}

func (c *Config) WithBinance(opt binance.BinanceConfig) configOpts {
	return func(c *Config) {
		c.Binance = opt
	}
}

func (c *Config) WithCoinbase(opt coinbase.CoinbaseConfig) configOpts {
	return func(c *Config) {
		c.Coinbase = opt
	}
}

func (c *Config) WithOkx(opt okx.OkxConfig) configOpts {
	return func(c *Config) {
		c.Okx = opt
	}
}

func (c *Config) XDiscord(opt external.Credentials) configOpts {
	return func(c *Config) {
		c.Discord = opt
	}
}

func (c *Config) XTelegram(opt external.Credentials) configOpts {
	return func(c *Config) {
		c.Telegram = opt
	}
}
