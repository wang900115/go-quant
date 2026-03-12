# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
# Build
go build -o build/app ./cmd

# Run all tests with race detection
go test -v -race ./...

# Run a single package's tests
go test -v -race ./stoploss/strategy/...

# Run a specific test
go test -v -race -run TestFixedTrailingStop ./stoploss/strategy/...

# Lint
go vet ./...

# Download dependencies
go mod download

# VS Code uses build tag: -tags=integration
```

CI runs on push/PR to `main`, `dev`, `stage`, `release/*`. Pipeline: `go vet` → `go test -race` with coverage → `go build`.

## Architecture

This is a Go quantitative trading platform (`github.com/wang900115/quant`). Data flows through three layers:

```
Exchange Providers (exchange/)  →  Strategy Engine (stoploss/engine/)  →  Notifications (external/)
```

### Exchange Layer (`exchange/`)
- `Provider` interface in `provider.go` defines REST (GetPrice, GetKlines, GetOrderBook, PlaceOrder) and WebSocket (SubscribeStream, Dispatch, ReceiveStream) operations.
- `Providers` is a thread-safe registry (`ExchangeId` → `Provider`).
- Each exchange (Binance, Coinbase, OKX) follows the same composition: `{name}.go` (REST), `trade.go` (orders), `ws_stream.go` (WebSocket).

### Strategy Layer (`stoploss/`)
- Core interfaces in `stoploss.go` (StopLoss family) and `takeprofit.go` (TakeProfit family). `Hybrid` combines both.
- `BaseResolver` provides shared Active/Callback logic via embedding.
- Six strategy families in `strategy/`: trailing, percentile, ATR, moving average, risk-reward, structure-swing. Each has Fixed and Debounced (time-delayed) variants.
- `Composite` in `composite.go` aggregates multiple strategies with `TriggerAny`/`TriggerAll` modes.
- Shared test data lives in `strategy/testdata.go`.

### Engine (`stoploss/engine/`)
- `StrategyEngine` orchestrates: `Portfolio` (strategy registry), `Execution` (buffered channels per strategy type for price fan-out), `Report` (result processing + callbacks), `Metrics` (observability).
- Uses `common/sys.Engine` for goroutine lifecycle with `SafeGo` (panic recovery + auto-restart).

### Notification Layer (`external/`)
- `Bot` interface (Register, Send, Close) in `bot.go`.
- Discord (webhook) and Telegram (bot API) in `third/`. Both use worker pools with rate limiting and exponential backoff retry.

### Storage (`storage/`)
- Hot/cold data separation: `KVStore` (random-access) and `AncientStore` (append-only historical).
- Pebble-backed implementation in `storage/pebble/`.
- `HookBatch` decorator adds callbacks on Put/Delete operations.

### Supporting Modules
- `model/` — Domain types: `PricePoint`, `OrderRequest`, `Currency`, `StrategyType`, trade enums, result types with JSON marshalling and `BuildMessage` for notifications.
- `metric/` — Thread-safe `CounterInt64` (atomic), `CounterPrecision` (CAS-based decimal), `GaugeInfo` (string k/v).
- `config/` — Centralized `Config` struct using functional options pattern.
- `common/sys/` — `Engine` goroutine manager. `common/parse/` — interval string parsing.

## Key Design Patterns

- **Interface-first**: Every major module exposes interfaces; implementations are swappable.
- **Composition via embedding**: Exchange clients embed Single+Stream+Trade clients. Debounced strategies embed their Fixed counterparts.
- **Channel-based concurrency**: Buffered channels per strategy type with non-blocking sends (drop + metric on overflow).
- **Generics**: Used for `PriceTick` constraint, `ParseOrderEntries`, `PushToChan`.
- **Decimal arithmetic**: All financial math uses `shopspring/decimal`, not float.

## Branch Strategy

- `main` — production
- `dev` — active development (default branch)
- `stage`, `release/*` — staging and releases

## Role-Specific Instructions

Specialized guidance for each role lives in `.claude/actors/`:

| File | Role | Scope |
|------|------|-------|
| `pm.md` | 專案經理 | 需求分析、任務拆解、Issue/PR 管理、跨角色協調 |
| `frontend.md` | 前端工程師 | ReactJS + TypeScript、型別對齊、金融數據顯示、WebSocket |
| `backend.md` | 後端工程師 | Go（延續現有架構模式）+ Rust（高效能模組） |
| `dba.md` | 資料庫管理員 | PostgreSQL + Pebble（延續 storage/ 層）+ Redis |
| `devops.md` | DevOps 工程師 | Ubuntu/Debian、Docker、CI/CD、監控、安全 |
