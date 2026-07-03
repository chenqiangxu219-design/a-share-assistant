import { useState } from 'react'
import { useStore } from '../store/store'
import { useMultiQuote } from '../hooks/useStockQuote'
import { apiPath } from '../utils/api'
import { Filter, Plus } from 'lucide-react'

export function Screener() {
  const watchlist = useStore((s) => s.watchlist)
  const addToWatchlist = useStore((s) => s.addToWatchlist)

  const [minScore, setMinScore] = useState('')
  const [maxScore, setMaxScore] = useState('')
  const [maxRSI6, setMaxRSI6] = useState('')
  const [minVolRatio, setMinVolRatio] = useState('')
  const [results, setResults] = useState<any[]>([])
  const [loading, setLoading] = useState(false)

  const handleScreen = async () => {
    setLoading(true)
    try {
      const res = await fetch(apiPath('/api/screener'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          codes: watchlist,
          min_score: minScore ? parseInt(minScore) : 0,
          max_score: maxScore ? parseInt(maxScore) : 0,
          max_rsi6: maxRSI6 ? parseFloat(maxRSI6) : 0,
          min_volume_ratio: minVolRatio ? parseFloat(minVolRatio) : 0,
        }),
      })
      if (res.ok) {
        const data = await res.json()
        setResults(data)
      }
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="fade-in">
      <div className="card">
        <div className="card-title">选股条件</div>
        <div style={{
          display: 'grid',
          gridTemplateColumns: 'repeat(4, 1fr)',
          gap: 16,
          marginBottom: 20
        }}>
          <div>
            <label style={{
              fontSize: 11,
              color: 'var(--text-tertiary)',
              display: 'block',
              marginBottom: 6,
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
              fontWeight: 600
            }}>最低评分</label>
            <input
              type="number"
              placeholder="-10 ~ 10"
              value={minScore}
              onChange={(e) => setMinScore(e.target.value)}
              style={{ width: '100%', fontFamily: 'var(--font-mono)' }}
            />
          </div>
          <div>
            <label style={{
              fontSize: 11,
              color: 'var(--text-tertiary)',
              display: 'block',
              marginBottom: 6,
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
              fontWeight: 600
            }}>最高评分</label>
            <input
              type="number"
              placeholder="-10 ~ 10"
              value={maxScore}
              onChange={(e) => setMaxScore(e.target.value)}
              style={{ width: '100%', fontFamily: 'var(--font-mono)' }}
            />
          </div>
          <div>
            <label style={{
              fontSize: 11,
              color: 'var(--text-tertiary)',
              display: 'block',
              marginBottom: 6,
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
              fontWeight: 600
            }}>RSI6 小于</label>
            <input
              type="number"
              placeholder="0 = 不限制"
              value={maxRSI6}
              onChange={(e) => setMaxRSI6(e.target.value)}
              style={{ width: '100%', fontFamily: 'var(--font-mono)' }}
            />
          </div>
          <div>
            <label style={{
              fontSize: 11,
              color: 'var(--text-tertiary)',
              display: 'block',
              marginBottom: 6,
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
              fontWeight: 600
            }}>量比 大于</label>
            <input
              type="number"
              placeholder="0 = 不限制"
              value={minVolRatio}
              onChange={(e) => setMinVolRatio(e.target.value)}
              style={{ width: '100%', fontFamily: 'var(--font-mono)' }}
            />
          </div>
        </div>
        <div style={{ display: 'flex', gap: 10, alignItems: 'center' }}>
          <button className="btn btn-primary" onClick={handleScreen} disabled={loading}>
            <Filter size={14} />
            {loading ? '筛选中...' : '开始筛选'}
          </button>
          <span style={{ fontSize: 12, color: 'var(--text-tertiary)' }}>
            当前自选股 <span style={{ color: 'var(--accent-blue)', fontWeight: 600 }}>{watchlist.length}</span> 只
          </span>
        </div>
      </div>

      {results.length > 0 && (
        <div className="card slide-up">
          <div className="card-title">
            筛选结果{' '}
            <span style={{
              fontSize: 12,
              color: 'var(--accent-blue)',
              fontWeight: 600,
              marginLeft: 8
            }}>
              {results.length} 只
            </span>
          </div>
          <table className="data-table">
            <thead>
              <tr>
                <th>代码</th>
                <th>名称</th>
                <th style={{ textAlign: 'right' }}>价格</th>
                <th style={{ textAlign: 'right' }}>评分</th>
                <th style={{ textAlign: 'center' }}>信号</th>
                <th style={{ textAlign: 'right' }}>RSI6</th>
                <th style={{ textAlign: 'right' }}>量比</th>
                <th style={{ textAlign: 'center' }}>操作</th>
              </tr>
            </thead>
            <tbody>
              {results.map((r) => (
                <tr key={r.code}>
                  <td style={{ fontFamily: 'var(--font-mono)' }}>{r.code}</td>
                  <td style={{ fontFamily: 'var(--font-sans)', fontWeight: 500 }}>{r.name}</td>
                  <td style={{ textAlign: 'right' }}>{r.price.toFixed(2)}</td>
                  <td style={{
                    textAlign: 'right',
                    color: r.score > 0 ? 'var(--up)' : r.score < 0 ? 'var(--down)' : 'var(--text-secondary)',
                    fontWeight: 700
                  }}>
                    {r.score > 0 ? '+' : ''}{r.score}
                  </td>
                  <td style={{ textAlign: 'center' }}>
                    <span className={`tag tag-${r.signal}`}>
                      {r.signal === 'buy' ? '买入' : r.signal === 'sell' ? '卖出' : '中性'}
                    </span>
                  </td>
                  <td style={{ textAlign: 'right' }}>{r.rsi6.toFixed(2)}</td>
                  <td style={{ textAlign: 'right' }}>{r.volume_ratio.toFixed(2)}</td>
                  <td style={{ textAlign: 'center' }}>
                    <button
                      className="btn"
                      style={{ padding: '4px 10px', fontSize: 11 }}
                      onClick={() => addToWatchlist(r.code)}
                    >
                      <Plus size={12} />
                      加入自选
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      {results.length === 0 && !loading && (
        <div className="card" style={{ textAlign: 'center', color: 'var(--text-tertiary)', padding: 40 }}>
          设置筛选条件后点击"开始筛选"
        </div>
      )}
    </div>
  )
}
