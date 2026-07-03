import { useState, useEffect } from 'react'
import { useStore } from '../store/store'
import { apiPath } from '../utils/api'
import { TrendingUp, TrendingDown, Flame } from 'lucide-react'

interface SectorData {
  name: string
  change_pct: number
  lead_stock: string
  lead_stock_pct: number
}

interface HotStock {
  code: string
  name: string
  change_pct: number
  volume: number
  concept: string
}

interface ConceptData {
  name: string
  change_pct: number
  stocks: string[]
  reason: string
}

function useHeatmapData() {
  const [data, setData] = useState<SectorData[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const fetchHeatmap = async () => {
      try {
        const resp = await fetch(apiPath('/api/heatmap'))
        const json = await resp.json()
        setData(json.sectors || [])
      } catch (e) {
        setError(e instanceof Error ? e.message : 'failed')
      } finally {
        setLoading(false)
      }
    }
    fetchHeatmap()
    const interval = setInterval(fetchHeatmap, 30000)
    return () => clearInterval(interval)
  }, [])

  return { data, loading, error }
}

function useHotStocks() {
  const [data, setData] = useState<HotStock[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const resp = await fetch(apiPath('/api/hot-stocks'))
        const json = await resp.json()
        setData(json.hot_stocks || [])
      } catch { /* ignore */ } finally {
        setLoading(false)
      }
    }
    fetchData()
    const interval = setInterval(fetchData, 30000)
    return () => clearInterval(interval)
  }, [])

  return { data, loading }
}

function useConcepts() {
  const [data, setData] = useState<ConceptData[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const fetchData = async () => {
      try {
        const resp = await fetch(apiPath('/api/concepts'))
        const json = await resp.json()
        setData(json.concepts || [])
      } catch { /* ignore */ } finally {
        setLoading(false)
      }
    }
    fetchData()
    const interval = setInterval(fetchData, 60000)
    return () => clearInterval(interval)
  }, [])

  return { data, loading }
}

// TradingView-like treemap
function HeatmapChart({ sectors }: { sectors: SectorData[] }) {
  const addToWatchlist = useStore((s) => s.addToWatchlist)

  // Simple grid layout: sort by absolute change, render as colored blocks
  const sorted = [...sectors].sort((a, b) => Math.abs(b.change_pct) - Math.abs(a.change_pct)).slice(0, 30)

  // Size based on absolute change percentage (larger change = bigger block)
  const maxAbs = Math.max(...sorted.map(s => Math.abs(s.change_pct)), 1)

  return (
    <div style={{
      display: 'flex',
      flexWrap: 'wrap',
      gap: 4,
      padding: 12,
    }}>
      {sorted.map((sector) => {
        const abs = Math.abs(sector.change_pct)
        const size = Math.max(8, Math.min(20, (abs / maxAbs) * 100))
        const isUp = sector.change_pct >= 0
        const bg = isUp
          ? `rgba(0, 208, 153, ${Math.min(0.3, abs / 15)})`
          : `rgba(255, 71, 86, ${Math.min(0.3, abs / 15)})`

        return (
          <div
            key={sector.name}
            onClick={() => addToWatchlist(sector.lead_stock)}
            style={{
              width: `${size}%`,
              minWidth: 100,
              height: Math.max(60, size * 0.6),
              background: bg,
              border: '1px solid rgba(255,255,255,0.06)',
              borderRadius: 'var(--radius-sm)',
              padding: '8px 10px',
              cursor: 'pointer',
              display: 'flex',
              flexDirection: 'column',
              justifyContent: 'center',
              transition: 'all var(--transition-fast)',
            }}
            title={sector.lead_stock ? `点击添加 ${sector.lead_stock} 到自选` : ''}
          >
            <div style={{
              fontSize: 12,
              fontWeight: 600,
              color: 'var(--text-primary)',
              marginBottom: 4,
              overflow: 'hidden',
              textOverflow: 'ellipsis',
              whiteSpace: 'nowrap',
            }}>
              {sector.name}
            </div>
            <div style={{
              fontSize: 14,
              fontWeight: 700,
              fontFamily: 'var(--font-mono)',
              color: isUp ? 'var(--up)' : 'var(--down)',
            }}>
              {sector.change_pct >= 0 ? '+' : ''}{sector.change_pct.toFixed(2)}%
            </div>
          </div>
        )
      })}
    </div>
  )
}

export function Heatmap() {
  const { data: sectors, loading: sectorsLoading, error: sectorsError } = useHeatmapData()
  const { data: hotStocks, loading: hotLoading } = useHotStocks()
  const { data: concepts, loading: conceptsLoading } = useConcepts()
  const [activeTab, setActiveTab] = useState<'heatmap' | 'hot' | 'concepts'>('heatmap')

  return (
    <div className="fade-in">
      <div style={{ marginBottom: 20 }}>
        <div style={{ fontSize: 18, fontWeight: 700, marginBottom: 4 }}>
          <Flame size={18} style={{ display: 'inline', verticalAlign: 'middle', marginRight: 6 }} />
          市场热点
        </div>
        <div style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>
          板块热力图 · 强势股排行 · 题材归因
        </div>
      </div>

      {/* Tabs */}
      <div style={{
        display: 'flex',
        gap: 4,
        marginBottom: 16,
        borderBottom: '1px solid var(--border-subtle)',
        paddingBottom: 0,
      }}>
        {([
          ['heatmap', '板块热力'],
          ['hot', '强势股'],
          ['concepts', '题材归因'],
        ] as const).map(([key, label]) => (
          <button
            key={key}
            onClick={() => setActiveTab(key)}
            style={{
              padding: '8px 16px',
              background: 'none',
              border: 'none',
              borderBottom: activeTab === key ? '2px solid var(--accent-blue)' : '2px solid transparent',
              color: activeTab === key ? 'var(--accent-blue)' : 'var(--text-secondary)',
              fontSize: 13,
              fontWeight: activeTab === key ? 600 : 400,
              cursor: 'pointer',
              transition: 'all var(--transition-fast)',
            }}
          >
            {label}
          </button>
        ))}
      </div>

      {/* Heatmap Tab */}
      {activeTab === 'heatmap' && (
        sectorsLoading ? (
          <div style={{ textAlign: 'center', padding: 40, color: 'var(--text-tertiary)' }}>加载中...</div>
        ) : sectorsError ? (
          <div style={{ textAlign: 'center', padding: 40, color: 'var(--down)' }}>
            数据加载失败: {sectorsError}
          </div>
        ) : (
          <div className="card">
            <HeatmapChart sectors={sectors} />
          </div>
        )
      )}

      {/* Hot Stocks Tab */}
      {activeTab === 'hot' && (
        <div className="card">
          <table className="data-table">
            <thead>
              <tr>
                <th style={{ width: 40 }}>#</th>
                <th>代码</th>
                <th>名称</th>
                <th style={{ textAlign: 'right' }}>涨跌幅</th>
                <th style={{ textAlign: 'right' }}>成交量</th>
              </tr>
            </thead>
            <tbody>
              {hotLoading ? (
                <tr><td colSpan={5} style={{ textAlign: 'center', padding: 20, color: 'var(--text-tertiary)' }}>加载中...</td></tr>
              ) : (
                hotStocks.map((stock, i) => {
                  const isUp = stock.change_pct >= 0
                  return (
                    <tr key={stock.code}>
                      <td style={{ color: 'var(--text-tertiary)', fontFamily: 'var(--font-mono)' }}>
                        {i + 1}
                      </td>
                      <td style={{ fontFamily: 'var(--font-mono)' }}>{stock.code}</td>
                      <td style={{ fontFamily: 'var(--font-sans)', fontWeight: 500 }}>
                        {stock.name}
                      </td>
                      <td style={{
                        textAlign: 'right',
                        color: isUp ? 'var(--up)' : 'var(--down)',
                        fontWeight: 600,
                        fontFamily: 'var(--font-mono)',
                      }}>
                        {isUp ? '+' : ''}{stock.change_pct.toFixed(2)}%
                      </td>
                      <td style={{
                        textAlign: 'right',
                        fontFamily: 'var(--font-mono)',
                        color: 'var(--text-secondary)',
                      }}>
                        {stock.volume > 0 ? (stock.volume / 10000).toFixed(0) + '万' : '--'}
                      </td>
                    </tr>
                  )
                })
              )}
            </tbody>
          </table>
        </div>
      )}

      {/* Concepts Tab */}
      {activeTab === 'concepts' && (
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
          gap: 12,
        }}>
          {conceptsLoading ? (
            <div style={{ color: 'var(--text-tertiary)' }}>加载中...</div>
          ) : (
            concepts.map((concept) => {
              const isUp = concept.change_pct >= 0
              return (
                <div key={concept.name} className="card" style={{ padding: 16 }}>
                  <div style={{
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center',
                    marginBottom: 8,
                  }}>
                    <span style={{ fontWeight: 600, fontSize: 14 }}>{concept.name}</span>
                    {isUp ? <TrendingUp size={16} className="positive" /> : <TrendingDown size={16} className="negative" />}
                  </div>
                  <div style={{
                    fontSize: 20,
                    fontWeight: 700,
                    fontFamily: 'var(--font-mono)',
                    color: isUp ? 'var(--up)' : 'var(--down)',
                    marginBottom: 6,
                  }}>
                    {isUp ? '+' : ''}{concept.change_pct.toFixed(2)}%
                  </div>
                  {concept.reason && (
                    <div style={{
                      fontSize: 11,
                      color: 'var(--text-tertiary)',
                    }}>
                      {concept.reason}
                    </div>
                  )}
                </div>
              )
            })
          )}
        </div>
      )}
    </div>
  )
}
