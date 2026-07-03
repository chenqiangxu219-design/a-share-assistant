# A股助手 — 专业级实时交易分析工具

> TradingView 风格的 A 股实时行情分析与 AI 辅助决策平台。Go + React + Electron，14,000+ 行代码。

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
| **前端** | React 19 + TypeScript + Electron + Zustand + TanStack Query + Lightweight Charts | ~4,800 行 |
| **后端** | Go 1.26 + Gin + Gorilla WebSocket + SQLite | ~9,700 行 |
| **AI** | Python + LangChain + Flask (新闻分析/智能对话) | ~90 行 |

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

```bash
# 启动后端
cd backend && go run main.go

# 启动前端
cd frontend && npm install && npm run dev
```

## 适用场景

- **个人交易者** — 实时看盘 + AI 辅助判断
- **量化研究员** — 策略回测框架
- **外包交付物** — 可直接作为金融类 App 的基础架构
