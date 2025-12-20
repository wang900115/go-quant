
![Quant Logo](./assets/assets.png)

## Overview

**Quant** is a quantitative analysis platform designed for financial data tracking, strategy modeling, and performance evaluation. It provides modular tools for algorithmic trading, data analytics, and strategy backtesting.

## Features

- 📊 **Data Integration** – Connect to multiple exchanges and data providers.  
- ⚙️ **Strategy Engine** – Design and test your own trading strategies.  
- 🧩 **Modular Design** – Clean architecture suitable for research and production.



## External Source 
Quant integrates multiple external market data sources for real-time and historical analysis:

- Binance
 – Spot, Futures, and Options trading data.

- Coinbase
 – Cryptocurrency spot market data and account info.

- OKX
 – Spot, Futures, and Perpetual contracts.


## Architecture 
```kotlin
┌─────────────────────┐
│   Data Collector    │  ← Market data feeds
└────────┬────────────┘
         ↓
┌─────────────────────┐
│  Strategy Engine    │  ← Backtesting / Live trading
└────────┬────────────┘
         ↓
┌─────────────────────┐
│  Analytics Module   │  ← Reports & visualization
└─────────────────────┘
```

## Examples

For detailed usage examples and code samples, see the [Example Documentation](./example/README.md).

The examples include:
- Strategy engine setup and configuration
- Registering multiple trading strategies
- Processing real-time price updates
- Implementing custom callbacks
- Complete working demos

## License

This project is dual-licensed under:

- **MIT License** – for individuals, researchers, and open-source developers.  
- **GNU AGPLv3 License** – for organizations deploying this software as a network service (e.g., SaaS).

You may choose which license to apply.

See [LICENSE-MIT](./LICENSE-MIT) and [LICENSE-AGPLv3](./LICENSE-AGPLv3) for details.

