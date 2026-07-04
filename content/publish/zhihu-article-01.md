# 知乎回答: "一个人能用 AI 做出什么产品？"

> 刚发布了 A股助手 v1.0。Go + React + Electron，24,000+ 行代码，1 人 + AI Agent 团队，零外包。

## 产品长什么样

TradingView 风格的 A 股实时看盘系统：
- WebSocket 推送毫秒级行情（5 数据源）
- K 线图 + 10 种技术指标
- AI 对话（自然语言选股）
- 策略回测（4 种内置策略对比）

## 架构

```
React 19 + Electron ◄─── Go 1.26 (WebSocket) ◄─── Python AI
```

Go 协程轮询 5 个数据源，通过 WebSocket 推送前端。Python 微服务提供 K 线、新闻、热力图数据。AI 对话集成 oMLX 本地 LLM，支持 9 个金融工具调用。

## AI Agent 团队

这不是 "用 Copilot 写代码"，而是完整的 Agent 协作：

| Agent | 模型 | 职责 |
|-------|------|------|
| CEO | Opus 4.7 | 脑暴 PRD、调度下游 |
| Architect | Opus 4.7 | plan.md（细到文件名） |
| Coder | Sonnet 4.6 | 写代码（可并行） |
| Reviewer | Codex | 逐行审查 |
| QA | Sonnet 4.6 | 跑起来 + 截图 |
| Content | Sonnet 4.6 | 文章/文档 |

**我的角色从 "一人全包" 变成了 "CEO + 质检员"。**

## 打包实战

Go 交叉编译 (CGO_ENABLED=0) → macOS/Windows/Linux 三平台。
Python PyInstaller 打包 (akshare + torch = 314MB)。
Electron extraResources 嵌入双进程，auto-start + health check。

最终 DMG: macOS arm64 (440MB), x64 (446MB)。Windows NSIS 安装包: 104MB。

## 开源了

完整代码 + 5 平台安装包，MPL-2.0 协议。

👉 https://github.com/chenqiangxu219-design/a-share-assistant

---

*如果你也在用 AI 做独立开发，欢迎交流。*
