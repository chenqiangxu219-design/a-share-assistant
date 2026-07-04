# V2EX 发帖内容

**标题:** 仿制 A 股实时看盘系统 — Go + Electron, AI Agent 辅助开发

**分类:** 程序员 / 创造者

**标签:** Go、Electron、React、独立开发

---

**正文:**

大家好，分享一个我最近用 AI Agent 团队辅助完成的开源项目 —— **A股助手**。

这是一个 TradingView 风格的 A 股实时行情分析与 AI 辅助决策平台。技术栈：

- **后端**: Go 1.26 + Gin + Gorilla WebSocket + SQLite（约 9,700 行）
- **前端**: React 19 + TypeScript + Electron + Lightweight Charts（约 4,800 行）
- **AI**: Python FastAPI + LangChain + oMLX 本地大模型

核心功能：
- WebSocket 推送毫秒级实时行情（接入新浪、东方财富等 5 个数据源）
- TradingView 风格 K 线图，支持多周期切换与 10+ 技术指标
- AI 智能对话助手：新闻情绪分析、板块轮动分析、策略建议
- 投资组合管理与策略回测引擎

项目亮点在于 **1 人 + AI Agent 团队** 的协作模式 —— 从架构设计到代码实现，由 6 个 AI Agent 角色（CEO、架构师、程序员、Reviewer、QA、内容）协同完成，总代码量 24,000+ 行。

目前已发布 v1.0.0，支持 macOS、Windows、Linux。欢迎 Star 和提 Issue：

🔗 https://github.com/chenqiangxu219-design/a-share-assistant

---
*首次发帖，欢迎拍砖。项目使用 MPL-2.0 协议开源。*
