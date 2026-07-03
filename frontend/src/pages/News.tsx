import { useState, useEffect } from 'react'
import { useStore } from '../store/store'
import { apiPath } from '../utils/api'
import { Newspaper, FileText, Search, TrendingUp, TrendingDown, Minus, Sparkles } from 'lucide-react'

interface NewsItem {
  title: string
  digest: string
  info_content: string
  info_url: string
  showtime: string
}

interface SentimentResult {
  title: string
  content: string
  sentiment: string
  score: number
  reason: string
  impact: string
}

interface SentimentSummary {
  code: string
  name: string
  overall_score: number
  positive_count: number
  negative_count: number
  neutral_count: number
  results: SentimentResult[]
  conclusion: string
}

interface ResearchItem {
  title: string
  author: string
  date: string
  digest: string
  url: string
}

interface SearchItem {
  code: string
  name: string
  price: number
  match: string
}

function useStockNews(code: string) {
  const [data, setData] = useState<NewsItem[]>([])
  const [loading, setLoading] = useState(false)

  const fetchData = () => {
    if (!code) return
    setLoading(true)
    fetch(apiPath(`/api/news/${code}`))
      .then((r) => r.json())
      .then((json) => setData(json.news || []))
      .catch(() => setData([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchData() }, [code])
  useEffect(() => {
    const interval = setInterval(fetchData, 60000)
    return () => clearInterval(interval)
  }, [code])

  return { data, loading }
}

function useStockResearch(code: string) {
  const [data, setData] = useState<ResearchItem[]>([])
  const [loading, setLoading] = useState(false)

  const fetchData = () => {
    if (!code) return
    setLoading(true)
    fetch(apiPath(`/api/research/${code}`))
      .then((r) => r.json())
      .then((json) => setData(json.research || []))
      .catch(() => setData([]))
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchData() }, [code])
  return { data, loading }
}

function useStockSentiment(code: string) {
  const [data, setData] = useState<SentimentSummary | null>(null)
  const [loading, setLoading] = useState(false)

  const fetchData = () => {
    if (!code) return
    setLoading(true)
    fetch(apiPath(`/api/news/analyze`), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code }),
    })
      .then((r) => r.json())
      .then((json) => setData(json))
      .catch(() => setData(null))
      .finally(() => setLoading(false))
  }

  useEffect(() => { fetchData() }, [code])
  return { data, loading, refetch: fetchData }
}

function SentimentBadge({ sentiment }: { sentiment: string }) {
  const style = sentiment === 'positive'
    ? { color: '#22c55e', background: 'rgba(34,197,94,0.1)', border: '1px solid rgba(34,197,94,0.3)' }
    : sentiment === 'negative'
    ? { color: '#ef4444', background: 'rgba(239,68,68,0.1)', border: '1px solid rgba(239,68,68,0.3)' }
    : { color: '#94a3b8', background: 'rgba(148,163,184,0.1)', border: '1px solid rgba(148,163,184,0.3)' }

  const label = sentiment === 'positive' ? '看多' : sentiment === 'negative' ? '看空' : '中性'

  return (
    <span style={{
      ...style,
      display: 'inline-flex',
      alignItems: 'center',
      gap: 3,
      padding: '2px 8px',
      borderRadius: 'var(--radius-sm)',
      fontSize: 11,
      fontWeight: 600,
      whiteSpace: 'nowrap',
    }}>
      {sentiment === 'positive' && <TrendingUp size={11} />}
      {sentiment === 'negative' && <TrendingDown size={11} />}
      {sentiment === 'neutral' && <Minus size={11} />}
      {label}
    </span>
  )
}

function SentimentSummaryCard({ summary, onRefresh }: { summary: SentimentSummary; onRefresh: () => void }) {
  const scoreColor = summary.overall_score > 0.3 ? '#22c55e' : summary.overall_score < -0.3 ? '#ef4444' : '#94a3b8'
  const scoreLabel = summary.overall_score > 0.3 ? '偏多' : summary.overall_score < -0.3 ? '偏空' : '中性'

  return (
    <div className="card" style={{ padding: 16, marginBottom: 16 }}>
      <div style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        marginBottom: 12,
      }}>
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: 6,
          fontSize: 14,
          fontWeight: 600,
          color: 'var(--text-primary)',
        }}>
          <Sparkles size={14} style={{ color: 'var(--accent-blue)' }} />
          新闻情感分析
        </div>
        <button
          onClick={onRefresh}
          style={{
            padding: '4px 10px',
            fontSize: 11,
            background: 'var(--bg-tertiary)',
            border: '1px solid var(--border-subtle)',
            borderRadius: 'var(--radius-sm)',
            color: 'var(--text-secondary)',
            cursor: 'pointer',
          }}
        >
          刷新
        </button>
      </div>

      <div style={{
        display: 'flex',
        alignItems: 'center',
        gap: 16,
        marginBottom: 12,
      }}>
        <div style={{
          fontSize: 28,
          fontWeight: 700,
          color: scoreColor,
          fontFamily: 'var(--font-mono)',
        }}>
          {summary.overall_score.toFixed(2)}
        </div>
        <div style={{ fontSize: 12, color: 'var(--text-secondary)' }}>
          <span style={{ color: scoreColor, fontWeight: 600 }}>{scoreLabel}</span>
          {' · '}
          {summary.name || summary.code}
        </div>
      </div>

      <div style={{
        display: 'flex',
        gap: 12,
        marginBottom: 10,
      }}>
        <span style={{ fontSize: 12, color: '#22c55e' }}>
          <TrendingUp size={12} style={{ display: 'inline', verticalAlign: 'middle' }} />
          {' '}{summary.positive_count}
        </span>
        <span style={{ fontSize: 12, color: '#94a3b8' }}>
          <Minus size={12} style={{ display: 'inline', verticalAlign: 'middle' }} />
          {' '}{summary.neutral_count}
        </span>
        <span style={{ fontSize: 12, color: '#ef4444' }}>
          <TrendingDown size={12} style={{ display: 'inline', verticalAlign: 'middle' }} />
          {' '}{summary.negative_count}
        </span>
      </div>

      <div style={{
        fontSize: 12,
        color: 'var(--text-secondary)',
        lineHeight: 1.6,
        padding: '8px 10px',
        background: 'var(--bg-tertiary)',
        borderRadius: 'var(--radius-sm)',
      }}>
        {summary.conclusion}
      </div>
    </div>
  )
}

export function News() {
  const selectedStock = useStore((s) => s.selectedStock)
  const defaultCode = selectedStock || '600519'

  const [searchCode, setSearchCode] = useState(defaultCode)
  const [activeTab, setActiveTab] = useState<'news' | 'research'>('news')
  const [searchQuery, setSearchQuery] = useState('')
  const [searchResults, setSearchResults] = useState<SearchItem[]>([])
  const [searching, setSearching] = useState(false)

  const { data: news, loading: newsLoading } = useStockNews(searchCode)
  const { data: research, loading: researchLoading } = useStockResearch(searchCode)
  const { data: sentiment, loading: sentimentLoading, refetch: refetchSentiment } = useStockSentiment(searchCode)

  const handleSearch = async () => {
    if (!searchQuery.trim()) return
    setSearching(true)
    try {
      const resp = await fetch(apiPath('/api/search'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ query: searchQuery }),
      })
      const json = await resp.json()
      setSearchResults(json.results || [])
    } catch {
      setSearchResults([])
    } finally {
      setSearching(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') handleSearch()
  }

  return (
    <div className="fade-in">
      <div style={{ marginBottom: 20 }}>
        <div style={{ fontSize: 18, fontWeight: 700, marginBottom: 4 }}>
          <Newspaper size={18} style={{ display: 'inline', verticalAlign: 'middle', marginRight: 6 }} />
          研报资讯
        </div>
        <div style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>
          个股新闻 · 研报摘要 · 语义选股
        </div>
      </div>

      {/* Stock Selector */}
      <div style={{
        display: 'flex',
        gap: 8,
        marginBottom: 16,
        alignItems: 'center',
      }}>
        <input
          type="text"
          value={searchCode}
          onChange={(e) => setSearchCode(e.target.value)}
          placeholder="输入股票代码"
          style={{
            width: 120,
            padding: '6px 10px',
            background: 'var(--bg-tertiary)',
            border: '1px solid var(--border-subtle)',
            borderRadius: 'var(--radius-sm)',
            color: 'var(--text-primary)',
            fontFamily: 'var(--font-mono)',
            fontSize: 13,
          }}
        />
        <div style={{
          flex: 1,
          display: 'flex',
          gap: 4,
        }}>
          {([
            ['news', <Newspaper key="n" size={14} />],
            ['research', <FileText key="r" size={14} />],
          ] as const).map(([key, icon]) => (
            <button
              key={key}
              onClick={() => setActiveTab(key)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                padding: '6px 14px',
                background: activeTab === key ? 'var(--accent-blue-dim)' : 'var(--bg-tertiary)',
                border: '1px solid',
                borderColor: activeTab === key ? 'var(--accent-blue)' : 'var(--border-subtle)',
                borderRadius: 'var(--radius-sm)',
                color: activeTab === key ? 'var(--accent-blue)' : 'var(--text-secondary)',
                fontSize: 12,
                fontWeight: activeTab === key ? 600 : 400,
                cursor: 'pointer',
              }}
            >
              {icon}
              {key === 'news' ? '新闻' : '研报'}
            </button>
          ))}
        </div>
      </div>

      {/* News List */}
      {activeTab === 'news' && (
        <div className="card" style={{ padding: 16 }}>
          {/* Sentiment Summary */}
          {sentimentLoading ? (
            <div style={{ textAlign: 'center', padding: 12, color: 'var(--text-tertiary)', fontSize: 12 }}>
              情感分析计算中...
            </div>
          ) : sentiment && (
            <SentimentSummaryCard summary={sentiment} onRefresh={refetchSentiment} />
          )}

          {newsLoading ? (
            <div style={{ textAlign: 'center', padding: 20, color: 'var(--text-tertiary)' }}>
              加载中...
            </div>
          ) : news.length === 0 ? (
            <div style={{ textAlign: 'center', padding: 20, color: 'var(--text-tertiary)' }}>
              暂无新闻数据
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              {news.map((item, i) => (
                <div
                  key={i}
                  style={{
                    padding: '12px 0',
                    borderBottom: i < news.length - 1 ? '1px solid var(--border-subtle)' : 'none',
                  }}
                >
                  <div style={{
                    display: 'flex',
                    alignItems: 'flex-start',
                    gap: 8,
                  }}>
                    <div style={{ flex: 1 }}>
                      <div style={{
                        fontSize: 13,
                        fontWeight: 600,
                        color: 'var(--text-primary)',
                        marginBottom: 4,
                        lineHeight: 1.5,
                      }}>
                        {item.title || '无标题'}
                      </div>
                      {item.digest && (
                        <div style={{
                          fontSize: 12,
                          color: 'var(--text-secondary)',
                          lineHeight: 1.6,
                          marginBottom: 4,
                        }}>
                          {item.digest}
                        </div>
                      )}
                      <div style={{
                        fontSize: 11,
                        color: 'var(--text-tertiary)',
                        display: 'flex',
                        gap: 12,
                      }}>
                        <span>{item.showtime}</span>
                        {item.info_url && (
                          <a
                            href={item.info_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            style={{ color: 'var(--accent-blue)', textDecoration: 'none' }}
                          >
                            原文链接
                          </a>
                        )}
                      </div>
                    </div>
                    {sentiment?.results[i] && (
                      <SentimentBadge sentiment={sentiment.results[i].sentiment} />
                    )}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Research List */}
      {activeTab === 'research' && (
        <div className="card" style={{ padding: 16 }}>
          {researchLoading ? (
            <div style={{ textAlign: 'center', padding: 20, color: 'var(--text-tertiary)' }}>
              加载中...
            </div>
          ) : research.length === 0 ? (
            <div style={{ textAlign: 'center', padding: 20, color: 'var(--text-tertiary)' }}>
              暂无研报数据
            </div>
          ) : (
            <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
              {research.map((item, i) => (
                <div
                  key={i}
                  style={{
                    padding: '12px 0',
                    borderBottom: i < research.length - 1 ? '1px solid var(--border-subtle)' : 'none',
                  }}
                >
                  <div style={{
                    fontSize: 13,
                    fontWeight: 600,
                    color: 'var(--text-primary)',
                    marginBottom: 4,
                    lineHeight: 1.5,
                  }}>
                    {item.title}
                  </div>
                  {item.author && (
                    <div style={{
                      fontSize: 11,
                      color: 'var(--text-tertiary)',
                      marginBottom: 2,
                    }}>
                      {item.author} {item.date && `· ${item.date}`}
                    </div>
                  )}
                  {item.digest && (
                    <div style={{
                      fontSize: 12,
                      color: 'var(--text-secondary)',
                      lineHeight: 1.6,
                    }}>
                      {item.digest}
                    </div>
                  )}
                  {item.url && (
                    <a
                      href={item.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      style={{
                        fontSize: 11,
                        color: 'var(--accent-blue)',
                        textDecoration: 'none',
                      }}
                    >
                      查看原文 →
                    </a>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      )}

      {/* NL Search Section */}
      <div className="card" style={{ padding: 16, marginTop: 16 }}>
        <div style={{
          fontSize: 14,
          fontWeight: 600,
          marginBottom: 12,
          color: 'var(--text-primary)',
        }}>
          <Search size={14} style={{ display: 'inline', verticalAlign: 'middle', marginRight: 4 }} />
          语义选股
        </div>
        <div style={{
          display: 'flex',
          gap: 8,
          marginBottom: 12,
        }}>
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="输入选股条件，如：低估值白酒股、PE低于20的银行股"
            style={{
              flex: 1,
              padding: '8px 12px',
              background: 'var(--bg-tertiary)',
              border: '1px solid var(--border-subtle)',
              borderRadius: 'var(--radius-sm)',
              color: 'var(--text-primary)',
              fontSize: 13,
            }}
          />
          <button
            className="btn btn-primary"
            onClick={handleSearch}
            disabled={searching || !searchQuery.trim()}
            style={{ padding: '8px 16px', whiteSpace: 'nowrap' }}
          >
            {searching ? '搜索中...' : '搜索'}
          </button>
        </div>

        {searchResults.length > 0 && (
          <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
            {searchResults.map((item, i) => (
              <div
                key={i}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 12,
                  padding: '8px 12px',
                  background: 'var(--bg-tertiary)',
                  borderRadius: 'var(--radius-sm)',
                }}
              >
                <span style={{
                  fontFamily: 'var(--font-mono)',
                  fontSize: 12,
                  color: 'var(--text-secondary)',
                  minWidth: 70,
                }}>
                  {item.code}
                </span>
                <span style={{
                  fontWeight: 600,
                  fontSize: 13,
                  color: 'var(--text-primary)',
                  flex: 1,
                }}>
                  {item.name}
                </span>
                <span style={{
                  fontFamily: 'var(--font-mono)',
                  fontSize: 12,
                  color: 'var(--text-secondary)',
                }}>
                  {item.price > 0 ? item.price.toFixed(2) : '--'}
                </span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
