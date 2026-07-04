# 我用 AI Agent 搭建了 A 股实时看盘系统

> 一个人，一台 Mac，6 个 AI Agent。24,000+ 行代码。零外包。

## 为什么做这个

每天看东方财富、同花顺，切来切去，信息碎片化。我想做一个 **TradingView 风格的 A 股看盘工具**——实时行情、技术分析、AI 辅助，全在一个界面里。

但真正让我兴奋的不是产品本身，而是 **如何用 AI Agent 团队一个人完成全栈开发**。

## 架构：Go + React + Electron

```
┌──────────────┐     ┌───────────────┐     ┌──────────────┐
│  React 19    │     │   Go Backend  │     │ Python AI    │
│  Electron    │◄───►│   Gin + WS    │◌───►│  LangChain   │
│  Lightweight │     │   SQLite      │     │  Flask       │
│  Charts      │     │               │     │              │
└──────────────┘     └───────────────┘     └──────────────┘
```

**为什么选 Go 做后端？**
- 协程天然适合 WebSocket 推送——5 个数据源并发轮询，Go 只需几个 goroutine
- SQLite 本地存储，零依赖部署（不需要 Redis、PostgreSQL）
- 单文件编译，Electron 打包时直接嵌入

**前端为什么选 React + Electron？**
- TradingView 级别的图表交互，Lightweight Charts 完美支持
- Zustand 状态管理——轻量、无 boilerplate
- TanStack Query 处理 API 缓存和自动重试

**代码量分布：** Go 后端 ~9,700 行，React 前端 ~4,800 行，Python AI 服务 ~90 行。

## 核心功能

### 实时行情 — WebSocket 推送

5 个数据源（新浪、东方财富、腾讯、问财、ITHomer）并发轮询，通过 WebSocket 推送毫秒级价格更新。

```go
// Go 协程轮询 — 每个数据源一个 goroutine
for _, source := range sources {
    go func(s DataSource) {
        for {
            quotes, err := s.Fetch()
            if err == nil {
                hub.Broadcast(quotes)  // WebSocket 广播
            }
            time.Sleep(time.Second * 3)
        }
    }(source)
}
```

前端用 TanStack Query + WebSocket hook，实现了 **断线自动重连** 和 **增量更新**（只变的价格会刷新）。

### AI 对话 — 9 个金融工具

这是最有意思的部分。AI 不只是聊天——它能 **调用工具**：

```
用户: "帮我看看贵州茅台的技术面"
    ↓
AI Agent → 调用 getQuote("600519")     ← 获取实时行情
         → 调用 calculateIndicator("MACD") ← 计算技术指标
         → 调用 analyzeSentiment()        ← 分析新闻情绪
    ↓
AI: "茅台当前价格 ¥1,850，MACD 金叉信号，
     新闻情绪偏正面（3/5），短期看多"
```

AI 对话集成的是 oMLX 本地 LLM（Claude API 格式），支持流式输出。思维链处理通过 `stripThinking()` 函数解析——oMLX 会把思考过程和最终回答混在一起，需要提取。

### 策略回测 — 4 种内置策略

- **均线交叉** (MA Cross)
- **MACD 金叉/死叉**
- **布林带突破** (Bollinger Band Break)
- **复合策略** (多信号加权)

回测引擎支持历史 K 线数据，输出胜率、最大回撤、夏普比率。

## AI Agent 团队 — 6 角色协作

这不是 "用 Copilot 写代码"，而是 **完整的 Agent 团队**：

| Agent | 模型 | 职责 |
|-------|------|------|
| **CEO (Elon Musk)** | Opus 4.7 | 脑暴 PRD、调度下游、汇总交付 |
| **Architect** | Opus 4.7 | PRD → plan.md（细到文件名） |
| **Coder** | Sonnet 4.6 | plan.md → 代码 + changes.log |
| **Reviewer** | Codex | 静态审查（方法化验证） |
| **QA/Verifier** | Sonnet 4.6 | 跑起来 → 截图 → 走用户流程 |
| **Content** | Sonnet 4.6 | 研究 → 初稿 → CEO 审方向 → 我审调性 |

**真实工作流：**
```
我说: "CEO，启动 自然语言选股"
    ↓
CEO (Opus) → 出 PRD → 我批准
Architect (Opus) → plan.md（细到文件名）→ 我健康检查
Coder (Sonnet) → 写代码（可并行多个 subagent）
Reviewer (Codex) → 逐行审查 → 循环收敛
QA (Sonnet) → 跑起来 + 截图
CEO → 汇总交付 → 我一键确认
```

**我的角色从 "一人全包" 变成了 "CEO + 质检员"**——把精力放在上游定义和下游验收。

## 踩坑记录

### Electron 打包 — Python 微服务怎么办？

Python AI 服务（FastAPI）需要作为独立进程启动。最终方案：
- **PyInstaller** 打包成单文件可执行
- Electron `extraResources` 嵌入
- main.ts 用 `spawn()` 启动 Go + Python 两个进程
- Health check 确认都起来后再打开窗口

### oMLX 思维链处理

本地 LLM 的思维链输出格式不稳定——有时有 `Here's a thinking process`，有时没有。写了 `stripThinking()` 函数做容错解析：先找 thinking prefix → 再找最后一个 `[TOOL`（真正的工具调用）→ fallback 到原始文本。

### Go + Electron 跨平台

Go 后端需要交叉编译（`GOOS=darwin/windows/linux`），CGO_ENABLED=0 确保静态链接。Windows 版需要 `.exe` 后缀，路径用 `path.Join` 处理。

## 下一步

- **开源** — MPL-2.0 协议，GitHub 已初始化
- **商业化** — 免费基础版 + 付费高级功能（AI选股、策略编辑器）
- **策略市场** — DSL 定义 + AI 自动生成变体

## 完整代码

👉 [GitHub: a-share-assistant](https://github.com/chenqiangxu219-design/a-share-assistant)
📥 [v1.0.0 下载](https://github.com/chenqiangxu219-design/a-share-assistant/releases/tag/v1.0.0)

**如果你也在用 AI Agent 做开发，欢迎交流。**