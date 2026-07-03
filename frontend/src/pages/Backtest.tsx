import { useState } from 'react'
import { useStore } from '../store/store'
import { createChart, LineSeries } from 'lightweight-charts'
import { useEffect, useRef } from 'react'
import { apiPath } from '../utils/api'
import { LineChart, Loader2, GitCompare } from 'lucide-react'

const STRATEGIES = [
  { key: 'composite', label: '综合评分' },
  { key: 'ma_cross', label: '均线交叉' },
  { key: 'macd', label: 'MACD' },
  { key: 'boll', label: '布林带' },
]

export function Backtest() {
  const watchlist = useStore((s) => s.watchlist)
  const [code, setCode] = useState(watchlist[0] || '')
  const [startDate, setStartDate] = useState('')
  const [endDate, setEndDate] = useState('')
  const [selectedStrategies, setSelectedStrategies] = useState<string[]>(['composite'])
  const [loading, setLoading] = useState(false)
  const [results, setResults] = useState<any[]>([])
  const chartRef = useRef<HTMLDivElement>(null)

  const toggleStrategy = (key: string) => {
    setSelectedStrategies(prev =>
      prev.includes(key) ? prev.filter(s => s !== key) : [...prev, key]
    )
  }

  const handleBacktest = async () => {
    setLoading(true)
    try {
      const res = await fetch(apiPath('/api/backtest'), {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          code,
          start: startDate,
          end: endDate,
          strategies: selectedStrategies,
        }),
      })
      if (res.ok) {
        const data = await res.json()
        if (data.results) {
          setResults(data.results)
        } else {
          setResults([data])
        }
      }
    } finally {
      setLoading(false)
    }
  }

  // Render comparison chart
  useEffect(() => {
    if (results.length === 0 || !chartRef.current) return

    const chart = createChart(chartRef.current, {
      width: chartRef.current.clientWidth,
      height: 300,
      layout: {
        background: { color: '#111827' },
        textColor: '#94a3b8',
      },
      grid: {
        vertLines: { color: 'rgba(51, 65, 85, 0.2)' },
        horzLines: { color: 'rgba(51, 65, 85, 0.2)' },
      },
      timeScale: {
        borderColor: 'rgba(51, 65, 85, 0.5)',
        timeVisible: false,
      },
      rightPriceScale: {
        borderColor: 'rgba(51, 65, 85, 0.5)',
      },
    }) as any

    const colors = ['#3b82f6', '#00d099', '#f59e0b', '#8b5cf6']

    results.forEach((result, i) => {
      if (!result.net_value_curve) return
      const series = chart.addSeries(LineSeries, {
        color: colors[i % colors.length],
        lineWidth: 2,
        priceLineVisible: false,
        lastValueVisible: true,
      })
      const curveData = result.net_value_curve.map((nv: any) => ({
        time: nv.date,
        value: nv.net_value,
      }))
      series.setData(curveData)
    })

    chart.timeScale().fitContent()

    const resizeObserver = new ResizeObserver((entries) => {
      for (const entry of entries) {
        chart.applyOptions({ width: entry.contentRect.width })
      }
    })
    resizeObserver.observe(chartRef.current)

    return () => {
      resizeObserver.disconnect()
      chart.remove()
    }
  }, [results])

  const StatCard = ({ title, value, className }: { title: string; value: string; className?: string }) => (
    <div className="card" style={{ padding: 14 }}>
      <div style={{
        fontSize: 10,
        color: 'var(--text-tertiary)',
        marginBottom: 6,
        textTransform: 'uppercase',
        letterSpacing: '0.08em',
        fontWeight: 600
      }}>
        {title}
      </div>
      <div style={{
        fontSize: 22,
        fontWeight: 700,
        fontFamily: 'var(--font-mono)',
        letterSpacing: '-0.02em'
      }} className={className}>
        {value}
      </div>
    </div>
  )

  return (
    <div className="fade-in">
      <div className="card">
        <div className="card-title">回测参数</div>
        <div style={{ display: 'flex', gap: 12, alignItems: 'end', flexWrap: 'wrap', marginBottom: 16 }}>
          <div>
            <label style={{
              fontSize: 11, color: 'var(--text-tertiary)', display: 'block',
              marginBottom: 6, textTransform: 'uppercase', fontWeight: 600, letterSpacing: '0.05em'
            }}>股票</label>
            <select value={code} onChange={(e) => setCode(e.target.value)}
              style={{ width: 160, fontFamily: 'var(--font-mono)' }}>
              {watchlist.map((c) => (
                <option key={c} value={c}>{c}</option>
              ))}
            </select>
          </div>
          <div>
            <label style={{
              fontSize: 11, color: 'var(--text-tertiary)', display: 'block',
              marginBottom: 6, textTransform: 'uppercase', fontWeight: 600, letterSpacing: '0.05em'
            }}>开始日期</label>
            <input type="date" value={startDate} onChange={(e) => setStartDate(e.target.value)} />
          </div>
          <div>
            <label style={{
              fontSize: 11, color: 'var(--text-tertiary)', display: 'block',
              marginBottom: 6, textTransform: 'uppercase', fontWeight: 600, letterSpacing: '0.05em'
            }}>结束日期</label>
            <input type="date" value={endDate} onChange={(e) => setEndDate(e.target.value)} />
          </div>
          <button className="btn btn-primary" onClick={handleBacktest} disabled={loading || !code || selectedStrategies.length === 0}>
            {loading ? <><Loader2 size={14} className="animate-spin" /> 回测中...</> : <><LineChart size={14} /> 开始回测</>}
          </button>
        </div>

        {/* Strategy Selection */}
        <div>
          <div style={{
            fontSize: 11, color: 'var(--text-tertiary)', marginBottom: 8,
            textTransform: 'uppercase', fontWeight: 600, letterSpacing: '0.05em'
          }}>
            <GitCompare size={12} style={{ display: 'inline', marginRight: 4, verticalAlign: 'middle' }} />
            策略选择 (可多选对比)
          </div>
          <div style={{ display: 'flex', gap: 8, flexWrap: 'wrap' }}>
            {STRATEGIES.map((s) => (
              <button
                key={s.key}
                onClick={() => toggleStrategy(s.key)}
                className="btn"
                style={{
                  background: selectedStrategies.includes(s.key) ? 'var(--accent-blue-dim)' : undefined,
                  borderColor: selectedStrategies.includes(s.key) ? 'var(--accent-blue)' : undefined,
                  color: selectedStrategies.includes(s.key) ? 'var(--accent-blue)' : undefined,
                }}
              >
                {s.label}
              </button>
            ))}
          </div>
        </div>

        <div style={{
          fontSize: 11, color: 'var(--text-tertiary)', marginTop: 12,
          fontFamily: 'var(--font-mono)'
        }}>
          初始资金: 100 万 | 手续费: 0.1% | 每次使用 90% 资金
        </div>
      </div>

      {/* Multi-strategy comparison */}
      {results.length > 0 && (
        <>
          {/* Stats comparison table */}
          <div className="card slide-up">
            <div className="card-title">策略对比</div>
            <table className="data-table">
              <thead>
                <tr>
                  <th>策略</th>
                  <th style={{ textAlign: 'right' }}>收益率</th>
                  <th style={{ textAlign: 'right' }}>最大回撤</th>
                  <th style={{ textAlign: 'right' }}>夏普比率</th>
                  <th style={{ textAlign: 'right' }}>胜率</th>
                  <th style={{ textAlign: 'right' }}>交易次数</th>
                </tr>
              </thead>
              <tbody>
                {results.map((r, i) => {
                  const colors = ['#3b82f6', '#00d099', '#f59e0b', '#8b5cf6']
                  const stratNames: Record<string, string> = {
                    composite: '综合评分', ma_cross: '均线交叉', macd: 'MACD', boll: '布林带'
                  }
                  const name = stratNames[r.strategy] || r.strategy || '综合评分'
                  return (
                    <tr key={i}>
                      <td>
                        <span style={{
                          display: 'inline-flex',
                          alignItems: 'center',
                          gap: 6,
                          fontFamily: 'var(--font-sans)',
                          fontWeight: 600
                        }}>
                          <span style={{
                            width: 8, height: 8, borderRadius: '50%',
                            background: colors[i % colors.length],
                            display: 'inline-block'
                          }} />
                          {name}
                        </span>
                      </td>
                      <td style={{
                        textAlign: 'right',
                        color: r.total_return >= 0 ? 'var(--up)' : 'var(--down)',
                        fontWeight: 700
                      }}>
                        {r.total_return >= 0 ? '+' : ''}{r.total_return.toFixed(2)}%
                      </td>
                      <td style={{ textAlign: 'right', color: 'var(--down)' }}>
                        -{r.max_drawdown.toFixed(2)}%
                      </td>
                      <td style={{ textAlign: 'right' }}>{r.sharpe_ratio.toFixed(2)}</td>
                      <td style={{ textAlign: 'right' }}>{r.win_rate.toFixed(1)}%</td>
                      <td style={{ textAlign: 'right' }}>{r.trade_count}</td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>

          {/* Net value comparison chart */}
          <div className="card slide-up" style={{ animationDelay: '100ms' }}>
            <div className="card-title">净值曲线对比</div>
            <div ref={chartRef} style={{ width: '100%' }} />
          </div>
        </>
      )}

      {results.length === 0 && !loading && (
        <div className="card" style={{ textAlign: 'center', color: 'var(--text-tertiary)', padding: 40 }}>
          选择股票和策略后点击"开始回测"
        </div>
      )}
    </div>
  )
}
