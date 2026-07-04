# A股助手 v1.0 发布：Go + Electron 全栈实战

> 从 0 到 v1.0，一个人加上 AI Agent 团队，到底需要多久？

上周，[A股助手](https://github.com/chenqiangxu219-design/a-share-assistant) 正式发布了 v1.0。后端 Go 约 9,700 行，前端 React 约 4,800 行，打包成 macOS DMG 和 Windows NSIS 安装包。这篇文章不讲功能有多好用——讲的是一个独立开发者在架构选型、跨平台打包、以及 AI 协作过程中踩过的坑。

如果你也在做全栈桌面应用，或者对 Go + Electron 的组合感兴趣，希望这篇文章能帮你少走弯路。

## 一、架构决策：为什么是 Go + Electron？

### Go 后端：协程天生适合实时数据推送

A股助手需要同时对接 5 个数据源（东方财富、同花顺、新浪、腾讯、akshare），维护 WebSocket 长连接，实时推送行情数据。Go 的 goroutine + channel 模型在这里是天然契合的。

项目里有一个 WebSocket Hub，负责管理所有客户端连接和广播行情更新：

```go
type Hub struct {
    clients    map[*Client]bool
    mu         sync.RWMutex
    broadcast  chan []byte
    register   chan *Client
    unregister chan *Client
}
```

Hub 的 `Run` 方法是一个无限循环，通过 `select` 监听三个 channel：新连接注册、断开注销、广播消息。每个客户端也有独立的 goroutine 负责发送，避免一个慢客户端阻塞整个广播通道：

```go
func (h *Hub) Run() {
    for {
        select {
        case client := <-h.register:
            h.mu.Lock()
            h.clients[client] = true
            h.mu.Unlock()
        case client := <-h.unregister:
            // ...清理逻辑
        case message := <-h.broadcast:
            h.mu.RLock()
            for client := range h.clients {
                select {
                case client.send <- message:
                default:
                    close(client.send)
                }
            }
            h.mu.RUnlock()
        }
    }
}
```

注意 `default` 分支——如果某个客户端的发送 channel 满了，直接断开连接而不是阻塞广播。这是 WebSocket Hub 的标准模式，gorilla/websocket 的官方示例就是这么写的。

每个数据源一个 goroutine，Hub 统一广播。如果换成 Node.js 单线程模型，处理大量并发 WebSocket 连接时需要额外引入 worker thread 或 cluster 模式，复杂度陡增。Go 的 `CGO_ENABLED=0` 编译出的静态二进制文件，也让后续打包变得非常简单——一个文件，没有依赖。

Go 后端还有一个亮点：AI 流式响应。`ChatStream` 函数通过 SSE（Server-Sent Events）将 oMLX 本地模型的输出实时推送到前端：

```go
w.Header().Set("Content-Type", "text/event-stream")
w.Header().Set("Cache-Control", "no-cache")

scanner := bufio.NewScanner(resp.Body)
scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

for scanner.Scan() {
    line := scanner.Text()
    if !strings.HasPrefix(line, "data: ") { continue }
    // 解析 SSE event，转发 text_delta
}
```

这里有一个细节：`scanner.Buffer` 被设为 1MB。因为 oMLX 返回的工具调用（tool use）可能很长，默认的 64KB buffer 会截断。踩了这个坑才加上的。

### Electron 桌面端：交付体验优先

为什么不直接用 Tauri？坦率地说，初期调研过。但 Electron 的 `extraResources` 机制可以原生嵌入任意二进制文件，配合 `spawn` 启动子进程，实现"一个安装包包含后端 + AI 服务 + 前端"的方案。Tauri 虽然体积小，但在嵌入 Python 打包产物时的兼容性还没有经过充分验证。

最终架构是这样的：Electron 主进程启动后，通过 `child_process.spawn` 分别拉起 Go 后端和 Python AI 服务，等待两个服务的健康检查通过后再创建窗口。

```typescript
app.whenReady().then(async () => {
  if (!isDev) {
    startGoBackend()
    startPythonService()
  }

  try {
    await Promise.all([waitForHealth(), waitForPythonHealth()])
  } catch (err) {
    if (!isDev) throw err
  }

  createWindow()
})
```

`waitForHealth` 是一个轮询函数，最多重试 30 次、每次间隔 1 秒，检查 `localhost:8080/health`。Python 服务对应的是 `localhost:8081/health`，最多重试 20 次。这里用 `Promise.all` 并发等待两个服务就绪——Go 后端通常 1-2 秒就起来了，Python 服务则需要 57 秒左右。如果串行等待，总启动时间会更长，但实际上 `Promise.all` 让两个健康检查并行进行，整体等待时间等于最慢的那个。

开发模式下（`VITE_DEV_MODE=true`），前端直接连 Vite dev server (`localhost:5173`)，后端和 Python 服务也独立运行。这样修改前端代码时可以热更新，不需要重新打包 Electron。

## 二、打包实战：三个进程，一套安装包

### Go 交叉编译

Go 后端编译最简单。`CGO_ENABLED=0` 确保生成纯静态二进制，不需要任何 C 库依赖：

```bash
CGO_ENABLED=0 go build -o a-share-backend .
```

但 Windows 交叉编译有个坑——`syscall.Stat_t`。Go 标准库在 macOS/Linux 上用的 `syscall.Stat_t` 结构体和 Windows 上的定义不同。如果你的代码直接引用了这个类型（比如在文件操作时），交叉编译到 Windows 会失败。解决方案是用 `build tags` 把平台相关的代码拆开：

```go
// +build !windows

package yourpkg
import "syscall"

func getFileSize(fi os.FileInfo) int64 {
    s := fi.Sys().(*syscall.Stat_t)
    return s.Size
}
```

Windows 版本用 `fi.Size()` 替代即可。交叉编译到 Windows 时：

```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o a-share-backend.exe .
```

注意 Go 1.21+ 已经大幅改善了 Windows 交叉编译的兼容性，但如果你用了 `syscall` 包里的平台相关类型，还是需要用 build tags 隔离。

### Python PyInstaller：akshare + torch = 314MB

这是整个打包过程中最痛苦的部分。Python AI 服务依赖 `akshare`（A股数据接口库）和 `torch`（本地模型推理），打包后直接飙到 314MB。

```bash
pyinstaller --onefile --noconsole app.py
cp dist/app ../builds/a-share-python-service
```

`--onefile` 把所有依赖打包成单个可执行文件，`--noconsole` 隐藏控制台窗口（Windows 上尤其重要，不然每次启动弹一个黑框）。但 akshare 内部引用了大量 `.csv`、`.json` 数据文件，PyInstaller 默认不会自动收集这些。需要手动写一个 hook：

```python
# hook-akshare.py
from PyInstaller.utils.hooks import collect_data_files

datas = collect_data_files('akshare')
```

把 hook 文件放在 `hooks/` 目录，打包时通过 `--additional-hooks-dir=./hooks` 指定。akshare 的数据文件大约有 20MB，主要是 A股板块分类、行业概念映射等静态数据。如果不加 hook，运行时 akshare 会报 `FileNotFoundError`——这个错误非常难排查，因为 akshare 的调用链很深， traceback 指向的是内部某个 `pd.read_csv`。

另一个经验：PyInstaller 打包前，先用 `pip install -r requirements.txt` 在一个干净的虚拟环境中安装依赖，然后用 `pyinstaller --onefile app.py`。不要在开发环境直接打包——你本地装的几百个包会被全部打进 exe，体积翻倍都不止。

还有一个隐蔽的问题：`tdxpy` 库连接通达信服务器时，需要保持 TCP 长连接。PyInstaller 打包后的单文件模式会在临时目录解压运行，如果临时目录权限不对，连接会失败。最终通过 `--noconsole` 配合显式指定服务器地址解决：

```python
TDX_HOST = "202.108.253.139"
TDX_PORT = 7709

def get_tdx_api():
    global _tdx_api, _tdx_connected
    if not _tdx_connected:
        try:
            _tdx_api.connect(TDX_HOST, TDX_PORT)
        except Exception:
            _tdx_api.connect()  # fallback to default server
```

### Electron extraResources：把两个进程塞进安装包

electron-builder 的配置很直接，通过 `extraResources` 指定要嵌入的额外文件：

```typescript
const config: Configuration = {
  appId: 'com.astock.assistant',
  productName: 'A股智能助手',
  mac: {
    category: 'public.app-category.finance',
    target: [{ target: 'dmg', arch: ['arm64', 'x64'] }],
  },
  win: {
    target: [{ target: 'nsis', arch: ['x64'] }],
  },
  extraResources: [
    { from: '../builds/a-share-backend', to: 'backend/a-share-backend' },
    { from: '../builds/a-share-python-service', to: 'python_service/app' },
  ],
}
```

打包后，Go 后端和 Python AI 服务会被放在 `process.resourcesPath` 目录下。Electron 主进程通过路径定位并启动它们：

```typescript
// Go backend
const binaryName = isWin ? 'a-share-backend.exe' : 'a-share-backend'
const target = path.join(process.resourcesPath, 'backend', binaryName)

// Python AI service
const pyBinary = isWin ? 'app.exe' : 'app'
const target = path.join(process.resourcesPath, 'python_service', pyBinary)
```

最终产物：macOS arm64 DMG 约 440MB，x64 约 446MB。体积主要来自 Python 服务（314MB）和 Electron 运行时（约 100MB），Go 后端只有几 MB。

完整的构建流程被封装在一个 shell 脚本里：

```bash
#!/bin/bash
# Full build pipeline: Go backend + Python service + Frontend + Electron packaging
set -e

echo "=== Step 1: Building Go backend ==="
bash build-backend.sh

echo "=== Step 2: Building Python service ==="
cd backend/python_service
pyinstaller --onefile --noconsole app.py
cp dist/app ../builds/a-share-python-service

echo "=== Step 3: Building Frontend ==="
cd ../frontend
npm ci && npm run build

echo "=== Step 4: Packaging Electron ==="
npx electron-builder --config electron-builder.config.ts --mac --publish never

echo "=== Build Complete ==="
```

四个步骤串行执行，总耗时约 15-20 分钟（取决于网络速度，`npm ci` 和 `pyinstaller` 最慢）。

## 三、踩坑记录

### 1. Python 启动慢：57 秒的等待

PyInstaller 打包的单文件模式有一个固有延迟——每次启动都需要先解压到临时目录，再执行。加上 `akshare` 和 `torch` 的初始化开销，Python 服务从 `spawn` 到健康检查通过需要约 57 秒。

拆解一下这 57 秒花在哪：
- **PyInstaller 解压**：约 30 秒。314MB 的单文件需要解压到临时目录，SSD 上也要这么久
- **torch 导入**：约 15 秒。`import torch` 本身就慢，更别说加载模型了
- **akshare + pandas**：约 5 秒。pandas 的 C 扩展初始化
- **FastAPI 启动**：约 1 秒。这个可以忽略

Python AI 服务基于 FastAPI，提供了行情查询、指标计算等 REST API。它同时依赖 `tdxpy` 连接通达信服务器获取实时行情，用 Redis 做缓存。Python 服务是整个架构中体积最大、启动最慢的部分——也是最大的优化目标。

这是目前最大的痛点。解决方案有两个方向：
- **短期**：在 Electron 窗口中显示加载动画，告知用户"正在初始化 AI 服务..."
- **长期**：考虑用 `--onedir` 模式打包，避免每次解压；或者把 torch 模型加载延迟到首次请求时再做

### 2. akshare 的数据文件路径问题

akshare 内部有一些本地缓存文件，PyInstaller `--onefile` 模式下工作目录是临时路径，导致二次启动时缓存丢失。通过在 `app.py` 启动时显式设置缓存目录到 `user_data_dir` 解决。

### 3. Windows 信号处理差异

Linux/macOS 上可以用 `SIGTERM` 优雅关闭进程，Windows 不支持。Electron 主进程的退出逻辑需要区分平台：

```typescript
function stopGoBackend() {
  if (!goProcess) return
  try {
    if (process.platform === 'win32') {
      goProcess.kill('SIGKILL')  // Windows 直接 SIGKILL
    } else {
      goProcess.kill('SIGTERM')   // macOS/Linux 优雅退出
      setTimeout(() => {
        if (!goProcess?.killed) goProcess?.kill('SIGKILL')
      }, 3000)
    }
  } catch { /* ignore */ }
}
```

### 4. oMLX 本地模型的 thinking process 剥离

Claude 模型通过 oMLX 本地运行时返回的响应包含一个 "thinking process"——模型在给出最终答案前会先输出一段推理过程。前端不需要看到这段内容，需要在 Go 层做剥离：

```go
func stripThinking(text string) string {
    for _, prefix := range []string{"Here's a thinking process", "Let me think through"} {
        idx := strings.Index(text, prefix)
        if idx < 0 { continue }
        // 找最后一个 [TOOL，那是真正的工具调用
        lastTool := strings.LastIndex(text[idx:], "[TOOL")
        if lastTool >= 0 {
            return strings.TrimSpace(text[idx+lastTool:])
        }
        // 没有工具调用，找输出部分的分隔符
        for _, marker := range []string{"\n## ", "\n##\n", "Output Generation:"} {
            fullIdx := strings.Index(text[idx:], marker)
            if fullIdx >= 0 {
                return strings.TrimSpace(text[idx+fullIdx+len(marker):])
            }
        }
        return strings.TrimSpace(text[idx:])
    }
    return strings.TrimSpace(text)
}
```

这段代码虽然 ugly，但实际运行稳定。更好的方案是等 oMLX 支持原生 `thinking` block 分离，但目前只能这样处理。

## 四、AI Agent 协作：Coder + Reviewer + QA 并行工作

这个项目最大的实验是——用 AI Agent 团队完成绝大部分编码工作。

### 分工模式

- **Coder**：负责具体功能实现。给它一个任务描述，它会读相关文件、写代码、跑测试。
- **Reviewer**：对 Coder 的产出做结构化审查，标注 severity（critical/high/medium/low），检查安全问题。
- **QA**：验证功能是否按预期工作，编写测试用例。

### 并行策略

关键 insight 是：Coder、Reviewer、QA 不是串行的，而是可以并行的。比如 Coder 在实现"技术指标计算"模块时，Reviewer 可以同时审查上一轮 Coder 完成的"WebSocket Hub"代码。QA 则在 Reviewer 通过后介入验证。

```
Coder → [实现功能 A] ──→ Reviewer → [审查 A] ──→ QA → [验证 A]
         ↓
    [实现功能 B] ──→ Reviewer → [审查 B] ──→ QA → [验证 B]
```

实际协作中，最有效的模式是"滚动交付"：Coder 每完成一个模块就交给 Reviewer，Reviewer 审查通过后 QA 验证，同时 Coder 已经开始下一个模块。整个流水线保持流动，没有等待时间。

### Prompt 工程的重要性

AI Agent 的效果高度依赖 prompt 质量。初期我给 Coder 的指令是"实现 MACD 指标计算"，结果它写了一堆冗余代码。后来改成结构化 prompt：

```
任务：实现 MACD 指标计算
输入：K线数据（open, high, low, close, volume）
输出：DIF、DEA、MACD 三个数组
要求：
1. 使用 EMA（指数移动平均），周期分别为 12 和 26
2. DEA 是 DIF 的 9 周期 EMA
3. MACD = (DIF - DEA) * 2
4. 返回 JSON 格式，包含 code、timestamp、values
```

效果立竿见影——代码量减少 40%，而且一次通过 Reviewer。经验是：**越具体的 prompt，Agent 产出质量越高**。不要说"实现一个功能"，要说清楚输入、输出、约束条件。

### 质量控制

AI Agent 不是万能的——它写的代码需要人工 review。我的做法是：
1. 每个模块完成后都过一遍 `go vet`、`gofmt`，确保基本规范
2. 前端跑 lint 检查，杜绝低级语法错误
3. 关键路径（WebSocket 连接管理、AI 流式响应）人工逐行 review
4. Reviewer Agent 标注的 critical/high 问题必须修复，medium/low 根据优先级决定

还有一个容易被忽视的点：**Agent 之间的上下文传递**。Coder 修改了一个函数签名，Reviewer 需要知道这个变更影响了哪些调用方。目前的做法是让 Coder 在完成任务后输出一份"变更摘要"，Reviewer 据此做影响面分析。这比让 Reviewer 自己去 diff 整个文件高效得多。

## 五、项目现状与后续计划

A股助手 v1.0 目前的功能矩阵：
- **5 个数据源**：东方财富、同花顺、新浪、腾讯、akshare
- **6 个技术指标**：MA、MACD、KDJ、RSI、BOLL、Volume
- **4 个交易策略**：均线交叉、MACD 金叉死叉、布林带突破、量价背离
- **AI 对话**：支持流式输出，9 个工具函数（查行情、算指标、分析策略等）

GitHub: https://github.com/chenqiangxu219-design/a-share-assistant
下载: https://github.com/chenqiangxu219-design/a-share-assistant/releases/tag/v1.0.0

### 下一步计划
- Python 服务启动优化（目标：从 57s 降到 10s 以内）
- Windows 打包完善（目前 NSIS 安装包还在调试中）
- Linux AppImage 支持
- 更多技术指标和策略

如果你对这个项目感兴趣，欢迎 Star、提 Issue、或者直接 Fork。一个人做全栈不容易，但有了 AI Agent 团队，效率确实提升了不少。

---

*本文首发于掘金，同步发布于 GitHub。项目采用 MPL-2.0 协议开源。*

**觉得有用的话点个 Star 支持一下吧！** ⭐ https://github.com/chenqiangxu219-design/a-share-assistant
