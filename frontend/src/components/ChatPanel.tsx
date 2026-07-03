import { useState, useRef, useEffect } from 'react'
import { apiPath } from '../utils/api'
import { MessageCircle, X, Send, Bot, Loader2, Zap } from 'lucide-react'
import { useStore } from '../store/store'

interface Message {
  role: 'user' | 'assistant'
  content: string
  streaming?: boolean // true while streaming text
}

const TOOL_NAMES: Record<string, string> = {
  get_quote: '获取报价',
  get_kline: '获取K线',
  get_signals: '获取技术信号',
  get_news: '获取资讯',
  search: '搜索股票',
  get_hot: '获取热点',
}

function toolLabel(name: string, args: string): string {
  const base = TOOL_NAMES[name] || name
  if (args) return `${base}(${args})`
  return base
}

export function ChatPanel() {
  const [open, setOpen] = useState(false)
  const [input, setInput] = useState('')
  const [messages, setMessages] = useState<Message[]>([])
  const [loading, setLoading] = useState(false)
  const [toolStatus, setToolStatus] = useState<{ name: string; args: string } | null>(null)
  const messagesRef = useRef<HTMLDivElement>(null)
  const selectedStock = useStore((s) => s.selectedStock)

  useEffect(() => {
    if (messagesRef.current) {
      messagesRef.current.scrollTop = messagesRef.current.scrollHeight
    }
  }, [messages, toolStatus])

  const sendMessage = async () => {
    if (!input.trim() || loading) return

    const userMsg: Message = { role: 'user', content: input }
    setMessages(prev => [...prev, userMsg])
    setInput('')
    setLoading(true)
    setToolStatus(null)

    // Placeholder for streaming text
    const assistantMsg: Message = { role: 'assistant', content: '', streaming: true }
    setMessages(prev => [...prev, assistantMsg])

    let accumulated = ''

    try {
      const res = await fetch(apiPath('/api/chat'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          message: userMsg.content,
          code: selectedStock || undefined,
          stream: true,
        }),
      })

      if (!res.ok) {
        const err = await res.json()
        setMessages(prev => {
          const filtered = prev.filter(m => !m.streaming)
          return [...filtered, { role: 'assistant', content: `错误: ${err.error}` }]
        })
        return
      }

      if (!res.body) return
      const reader = res.body.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

      while (true) {
        const { done, value } = await reader.read()
        if (done) break

        buffer += decoder.decode(value, { stream: true })

        const lines = buffer.split('\n\n')
        buffer = lines.pop() || ''

        for (const event of lines) {
          if (!event.trim()) continue

          const data = event.replace(/^data: /, '').trim()

          if (data === '[DONE]') {
            // Finalize
            setMessages(prev =>
              prev.map(m => m.streaming ? { ...m, streaming: false } : m)
            )
            continue
          }

          if (data.startsWith('tool_start:')) {
            const parts = data.split(':', 3)
            const name = parts[1] || ''
            const args = parts[2] || ''
            setToolStatus({ name, args })
            continue
          }

          if (data.startsWith('tool_done:')) {
            const name = data.replace('tool_done:', '')
            setToolStatus(null)
            continue
          }

          if (data === 'final_start') {
            setToolStatus(null)
            accumulated = ''
            continue
          }

          if (data.startsWith('[ERROR]') || data.startsWith('[超时]') || data.startsWith('[未知工具')) {
            setMessages(prev =>
              prev.map(m => m.streaming ? { ...m, content: m.content + data, streaming: true } : m)
            )
            continue
          }

          // Text chunk
          accumulated += data
          setMessages(prev =>
            prev.map(m => m.streaming ? { ...m, content: accumulated, streaming: true } : m)
          )
        }
      }

      // Remaining buffer may contain events without trailing \n\n
      if (buffer.trim()) {
        const data = buffer.replace(/^data: /, '').trim()
        if (data && data !== '[DONE]') {
          accumulated += data
          setMessages(prev =>
            prev.map(m => m.streaming ? { ...m, content: accumulated, streaming: false } : m)
          )
        }
      }
    } catch (err: any) {
      setMessages(prev => {
        const filtered = prev.filter(m => !m.streaming)
        return [...filtered, { role: 'assistant', content: `网络错误: ${err.message}` }]
      })
    } finally {
      setLoading(false)
      setToolStatus(null)
    }
  }

  const toggle = () => setOpen(!open)

  return (
    <>
      {/* Floating Button */}
      <button
        onClick={toggle}
        style={{
          position: 'fixed',
          bottom: 24,
          right: 24,
          width: 56,
          height: 56,
          borderRadius: '50%',
          border: 'none',
          background: 'linear-gradient(135deg, var(--accent-blue), var(--accent-purple))',
          color: 'white',
          cursor: 'pointer',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          boxShadow: '0 4px 20px rgba(59, 130, 246, 0.4)',
          transition: 'transform 0.2s',
          zIndex: 1000,
        }}
        onMouseEnter={(e) => { e.currentTarget.style.transform = 'scale(1.1)' }}
        onMouseLeave={(e) => { e.currentTarget.style.transform = 'scale(1)' }}
      >
        {open ? <X size={24} /> : <MessageCircle size={24} />}
      </button>

      {/* Panel */}
      {open && (
        <div style={{
          position: 'fixed',
          bottom: 92,
          right: 24,
          width: 380,
          height: 520,
          background: 'var(--bg-secondary)',
          border: '1px solid var(--border-subtle)',
          borderRadius: 'var(--radius-lg)',
          display: 'flex',
          flexDirection: 'column',
          overflow: 'hidden',
          boxShadow: '0 8px 40px rgba(0,0,0,0.5)',
          zIndex: 999,
          animation: 'slideUp 0.3s ease',
        }}>
          {/* Header */}
          <div style={{
            padding: '14px 16px',
            borderBottom: '1px solid var(--border-subtle)',
            display: 'flex',
            alignItems: 'center',
            gap: 10,
            background: 'var(--bg-glass)',
          }}>
            <Bot size={20} style={{ color: 'var(--accent-blue)' }} />
            <span style={{ fontWeight: 700, fontSize: 14 }}>AI 智能助手</span>
            {selectedStock && (
              <span style={{
                fontSize: 11,
                color: 'var(--text-tertiary)',
                fontFamily: 'var(--font-mono)',
                marginLeft: 'auto'
              }}>
                {selectedStock}
              </span>
            )}
          </div>

          {/* Messages */}
          <div
            ref={messagesRef}
            style={{
              flex: 1,
              overflowY: 'auto',
              padding: '12px 16px',
              display: 'flex',
              flexDirection: 'column',
              gap: 12,
            }}
          >
            {messages.length === 0 && (
              <div style={{
                textAlign: 'center',
                color: 'var(--text-tertiary)',
                fontSize: 13,
                marginTop: 40,
                lineHeight: 1.8,
              }}>
                <Bot size={32} style={{ marginBottom: 12, opacity: 0.3 }} />
                <div>你好！我是 A 股智能助手</div>
                <div style={{ fontSize: 11, marginTop: 8 }}>
                  试试问我：
                </div>
                <div style={{
                  display: 'flex',
                  flexDirection: 'column',
                  gap: 6,
                  marginTop: 8,
                  alignItems: 'center'
                }}>
                  {['帮我分析贵州茅台', '低估值白酒股有哪些', '当前市场情绪如何'].map((q) => (
                    <button
                      key={q}
                      onClick={() => setInput(q)}
                      style={{
                        padding: '4px 12px',
                        background: 'var(--bg-tertiary)',
                        border: '1px solid var(--border-subtle)',
                        borderRadius: 12,
                        fontSize: 11,
                        color: 'var(--text-secondary)',
                        cursor: 'pointer',
                        fontFamily: 'var(--font-sans)',
                      }}
                    >
                      {q}
                    </button>
                  ))}
                </div>
              </div>
            )}

            {messages.map((msg, i) => (
              <div
                key={i}
                style={{
                  display: 'flex',
                  justifyContent: msg.role === 'user' ? 'flex-end' : 'flex-start',
                }}
              >
                <div style={{
                  maxWidth: '85%',
                  padding: '10px 14px',
                  borderRadius: msg.role === 'user'
                    ? '14px 14px 4px 14px'
                    : '14px 14px 14px 4px',
                  background: msg.role === 'user'
                    ? 'linear-gradient(135deg, var(--accent-blue), var(--accent-purple))'
                    : 'var(--bg-tertiary)',
                  color: msg.role === 'user' ? 'white' : 'var(--text-primary)',
                  fontSize: 13,
                  lineHeight: 1.6,
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                }}>
                  {msg.content}
                  {msg.streaming && (
                    <span style={{ display: 'inline-block', marginLeft: 2, animation: 'blink 1s infinite' }}>|</span>
                  )}
                </div>
              </div>
            ))}

            {/* Tool status indicator */}
            {toolStatus && (
              <div style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                padding: '6px 12px',
                margin: '0 auto',
                background: 'var(--bg-tertiary)',
                border: '1px solid var(--border-subtle)',
                borderRadius: 16,
                fontSize: 11,
                color: 'var(--text-secondary)',
              }}>
                <Zap size={10} style={{ color: 'var(--accent-blue)' }} />
                <span>正在{toolLabel(toolStatus.name, toolStatus.args)}...</span>
              </div>
            )}

            {loading && !toolStatus && (
              <div style={{ display: 'flex', gap: 4, padding: '8px 0', justifyContent: 'center' }}>
                <Loader2 size={14} style={{ color: 'var(--accent-blue)', animation: 'spin 1s linear infinite' }} />
                <span style={{ fontSize: 11, color: 'var(--text-tertiary)' }}>思考中...</span>
              </div>
            )}
          </div>

          {/* Input */}
          <div style={{
            padding: '12px 16px',
            borderTop: '1px solid var(--border-subtle)',
            display: 'flex',
            gap: 8,
          }}>
            <input
              type="text"
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && sendMessage()}
              placeholder="输入问题..."
              style={{
                flex: 1,
                background: 'var(--bg-tertiary)',
                border: '1px solid var(--border-subtle)',
                borderRadius: 20,
                padding: '8px 14px',
                fontSize: 13,
                color: 'var(--text-primary)',
                outline: 'none',
                fontFamily: 'var(--font-sans)',
              }}
            />
            <button
              onClick={sendMessage}
              disabled={!input.trim() || loading}
              style={{
                width: 36,
                height: 36,
                borderRadius: '50%',
                border: 'none',
                background: input.trim() && !loading
                  ? 'linear-gradient(135deg, var(--accent-blue), var(--accent-purple))'
                  : 'var(--bg-tertiary)',
                color: input.trim() && !loading ? 'white' : 'var(--text-tertiary)',
                cursor: input.trim() && !loading ? 'pointer' : 'default',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                transition: 'all 0.2s',
              }}
            >
              <Send size={16} />
            </button>
          </div>
        </div>
      )}
    </>
  )
}
