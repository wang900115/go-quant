// Copyright 2025 Quantive. All rights reserved.

// Licensed under the MIT License (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at

// https://opensource.org/licenses/MIT
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wang900115/quant/config"
	"github.com/wang900115/quant/stoploss/engine"
)

func main() {
	cfg, err := config.NewFromEnv()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	if err := cfg.Validate(); err != nil {
		slog.Error("invalid config", "error", err)
		os.Exit(1)
	}

	engineCfg := cfg.Engine
	if engineCfg.BufferSize == 0 {
		engineCfg = engine.DefaultConfig()
	}

	eng := engine.New(engineCfg)
	slog.Info("strategy engine initialized")

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := eng.Start(); err != nil {
			slog.Error("engine start failed", "error", err)
			stop()
		}
	}()

	slog.Info("platform started — press Ctrl+C to stop")
	<-ctx.Done()

	slog.Info("shutdown signal received, stopping engine...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	done := make(chan struct{})
	go func() {
		eng.Stop()
		close(done)
	}()

	select {
	case <-done:
		slog.Info("engine stopped cleanly")
	case <-shutdownCtx.Done():
		slog.Warn("shutdown timed out, forcing exit")
	}
}
