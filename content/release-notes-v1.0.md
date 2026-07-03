# A股智能助手 v1.0

**发布日期:** 2026-07-04

TradingView 风格的 A 股实时行情分析与 AI 辅助决策平台。Go + React + Electron，24,000+ 行代码，1 人 + AI Agent 团队完成。

## 📊 核心功能

### 实时行情
- WebSocket 推送毫秒级价格更新（5 数据源：新浪/东财/腾讯/问财/ITHomer）
- TradingView 风格 K 线图（Lightweight Charts v5.2）
- 多周期切换：1分钟 / 5分钟 / 日线 / 周线

### 技术分析
- MA、EMA、MACD、RSI、KDJ、布林带等 10+ 技术指标
- 4 种交易策略：均线交叉、MACD 金叉/死叉、布林带突破、复合策略
- 资金流向追踪（主力/超大单）

### AI 辅助决策
- 智能对话：基于 oMLX 本地 LLM，9 个金融工具调用
- 新闻情绪分析：实时抓取财经新闻，AI 评估市场情绪
- 板块轮动检测

### 投资组合管理
- 模拟持仓追踪 + 盈亏计算
- 多策略回测对比（胜率/最大回撤/夏普比率）

## 🛠 技术栈
- **后端:** Go 1.26 + Gin + Gorilla WebSocket + SQLite (~9,700 行)
- **前端:** React 19 + TypeScript + Electron + Zustand (~4,800 行)
- **AI:** Python FastAPI microservice

## 📥 下载
| 平台 | 文件 | 大小 |
|------|------|------|
| macOS (Apple Silicon) | A股智能助手-1.0.0-arm64.dmg | 440 MB |
| macOS (Intel) | A股智能助手-1.0.0-x64.dmg | 446 MB |

## 🚀 快速开始
```bash
# 桌面端：下载安装包，解压运行

# 开发模式
cd backend && go run main.go      # 后端 :8080
cd frontend && npm install && npm run dev   # 前端 :5173
```

## 📖 文档
- [首篇文章：我用 AI Agent 搭建了 A 股实时看盘系统](../content/article-01.md)
- [CHANGELOG](CHANGELOG.md)

## ⚠️ 免责声明
本工具仅供学习研究使用，不构成投资建议。市场有风险，投资需谨慎。

---
**License:** MPL-2.0 | **Star us on GitHub →** [链接]
