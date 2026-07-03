# Changelog

## [1.0.0] - 2026-07-04
### Added
- Real-time A-share quotes via WebSocket (5 data sources: Sina, EastMoney, Tencent, iwencai, ithomer)
- Technical analysis: MA, EMA, MACD, RSI, KDJ, Bollinger Band indicators
- AI chat assistant with tool-use capability (9 tools, oMLX local LLM)
- Sector rotation analysis + heatmap visualization
- Portfolio simulation with P&L tracking
- Stock screener with multi-condition filtering
- Strategy backtesting engine (4 strategies: MA Cross, MACD, BOLL Break, Composite)
- News sentiment analysis via Python microservice
- Capital flow tracking (main fund / large order)
- TradingView-style charts with Lightweight Charts v5.2
- Electron desktop app (macOS / Windows)

### Tech Stack
- Backend: Go 1.26 + Gin + Gorilla WebSocket + SQLite (~9,700 lines)
- Frontend: React 19 + TypeScript + Electron + Zustand (~4,800 lines)
- AI: Python FastAPI microservice (~90 lines)

### Notes
- License: MPL-2.0
- Local LLM: oMLX at localhost:8000 (Claude API format)
