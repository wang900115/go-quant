package config

import (
	"github.com/wang900115/quant/exchange/binance"
	"github.com/wang900115/quant/exchange/coinbase"
	"github.com/wang900115/quant/exchange/okx"
	"github.com/wang900115/quant/stoploss/engine"
)

type Config struct {
	// Strategy configurations
	Engine engine.Config
	// Exchange configurations
	Binance  binance.BinanceConfig
	Coinbase coinbase.CoinbaseConfig
	Okx      okx.OkxConfig
}

type configOpts func(c *Config)

func (c *Config) apply(opts ...configOpts) {
	for _, opt := range opts {
		opt(c)
	}
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
