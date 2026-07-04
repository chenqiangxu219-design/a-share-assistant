# A股助手 — 专业级实时交易分析工具

[![CI](https://github.com/chenqiangxu219-design/a-share-assistant/actions/workflows/ci.yml/badge.svg)](https://github.com/chenqiangxu219-design/a-share-assistant/actions/workflows/ci.yml)
[![Release v1.0.0](https://img.shields.io/badge/release-v1.0.0-blue.svg)](https://github.com/chenqiangxu219-design/a-share-assistant/releases/tag/v1.0.0)
[![License](https://img.shields.io/badge/license-MPL--2.0-green.svg)](LICENSE)

> TradingView 风格的 A 股实时行情分析与 AI 辅助决策平台。Go + React + Electron，24,000+ 行代码，1 人 + AI Agent 团队完成。

## 技术架构

```
┌──────────────┐     ┌───────────────┐     ┌──────────────┐
│  React 19    │     │   Go Backend  │     │ Python AI    │
│  Electron    │◄───►│   Gin + WS    │◄───►│  LangChain   │
│  Lightweight │     │   SQLite      │     │  Flask       │
│  Charts      │     │               │     │              │
└──────────────┘     └───────────────┘     └──────────────┘
```

| 层级 | 技术栈 | 代码量 |
|------|--------|--------|
| **前端** | React 19 + TypeScript + Electron + Zustand + TanStack Query + Lightweight Charts v5.2 | ~4,800 行 |
| **后端** | Go 1.26 + Gin + Gorilla WebSocket + SQLite | ~9,700 行 |
| **AI** | Python FastAPI + akshare (K线/新闻/热力图) + oMLX 本地 LLM | ~470 行 |

## 核心功能

### 📊 实时行情
- WebSocket 推送毫秒级价格更新
- TradingView 风格 K 线图（Lightweight Charts）
- 多周期切换：1分钟 / 5分钟 / 日线 / 周线

### 📈 技术分析
- **指标系统**：MA、EMA、MACD、RSI、KDJ、布林带等 10+ 技术指标
- **策略引擎**：可配置的策略回测框架
- **资金流向**：主力资金、超大单追踪

### 🤖 AI 辅助决策
- **智能对话**：基于 LangChain 的 AI 分析助手
- **新闻情绪分析**：实时抓取财经新闻，AI 评估市场情绪
- **板块轮动**：行业板块热度分析

### 💼 投资组合管理
- 持仓追踪与盈亏计算
- 策略回测验证
- 多组合对比分析

## 项目亮点

1. **全栈自研** — 从零构建，前后端分离，WebSocket 实时通信
2. **TradingView 级体验** — Lightweight Charts 实现专业级图表交互
3. **AI 原生集成** — LangChain + LLM，非简单 API 调用
4. **高性能后端** — Go 协程处理并发，SQLite 本地存储零依赖
5. **桌面端交付** — Electron 打包，跨平台运行

## 快速开始

### 📥 下载
| 平台 | 文件 | 大小 |
|------|------|------|
| macOS (Apple Silicon) | [A股智能助手-1.0.0-arm64.dmg](https://github.com/chenqiangxu219-design/a-share-assistant/releases/download/v1.0.0/A.-1.0.0-arm64.dmg) | 440 MB |
| macOS (Intel) | [A股智能助手-1.0.0-x64.dmg](https://github.com/chenqiangxu219-design/a-share-assistant/releases/download/v1.0.0/A.-1.0.0-x64.dmg) | 446 MB |

### 🛠 开发模式
```bash
# 启动后端
cd backend && go run main.go      # HTTP :8080 + WebSocket

# 启动 Python AI 服务
cd backend/python_service && pip install -r requirements.txt && python app.py   # :8081

# 启动前端
cd frontend && npm install && npm run dev   # :5173
```

### ⚙️ 环境变量
```bash
cp .env.example .env
# 编辑 .env 配置 LLM_BASE_URL, IWENCAI_API_KEY 等
```

## 适用场景

- **个人交易者** — 实时看盘 + AI 辅助判断
- **量化研究员** — 策略回测框架
- **独立开发者** — Go + Electron 全栈参考实现

## 路线图

- [x] ~~v1.0: 实时行情 + AI 对话~~ (2026-07)
- [ ] v1.1: WebSocket 增量推送 + Auth 系统
- [ ] v2.0: 策略市场 (DSL + AI 生成)

## 贡献

欢迎 Star、提 Issue、Fork。重大 PR 请先开 Issue 讨论。

## License

[Mozilla Public License 2.0](LICENSE)
