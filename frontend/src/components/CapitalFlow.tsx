import { useEffect, useState, useRef } from 'react'
import { apiPath } from '../utils/api'
import { createChart, LineSeries } from 'lightweight-charts'
import { TrendingUp, TrendingDown } from 'lucide-react'

interface CapitalFlowData {
  date: string
  main_net: number
  main_net_pct: number
  super_large_net: number
  large_net: number
  medium_net: number
  small_net: number
}

export function CapitalFlow({ code }: { code: string }) {
  const [data, setData] = useState<CapitalFlowData[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!code) return
    setLoading(true)
    fetch(apiPath(`/api/capital/${code}?days=10`))
      .then(r => r.json())
      .then(d => {
        setData(d.reverse())
        setLoading(false)
      })
      .catch(() => setLoading(false))
  }, [code])

  if (loading || data.length === 0) {
    return (
      <div className="card">
        <div className="card-title">资金流向</div>
        <div className="skeleton" style={{ width: '100%', height: 100 }} />
      </div>
    )
  }

  const latest = data[data.length - 1]
  const isPositive = latest.main_net > 0
  const flowText = (val: number) => {
    const abs = Math.abs(val)
    if (abs >= 100000000) return (val / 100000000).toFixed(2) + '亿'
    if (abs >= 10000) return (val / 10000).toFixed(1) + '万'
    return val.toFixed(0)
  }

  return (
    <div className="card">
      <div className="card-title">资金流向</div>

      {/* Summary Cards */}
      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(2, 1fr)',
        gap: 10,
        marginBottom: 16,
      }}>
        <div style={{
          padding: 12,
          background: isPositive ? 'var(--up-bg)' : 'var(--down-bg)',
          borderRadius: 'var(--radius-sm)',
          border: `1px solid ${isPositive ? 'var(--up)' : 'var(--down)'}`,
        }}>
          <div style={{
            fontSize: 10,
            color: 'var(--text-tertiary)',
            marginBottom: 4,
            textTransform: 'uppercase',
          }}>
            主力净流入
          </div>
          <div style={{
            fontSize: 18,
            fontWeight: 700,
            fontFamily: 'var(--font-mono)',
            color: isPositive ? 'var(--up)' : 'var(--down)',
            display: 'flex',
            alignItems: 'center',
            gap: 4,
          }}>
            {isPositive ? <TrendingUp size={16} /> : <TrendingDown size={16} />}
            {flowText(latest.main_net)}
          </div>
          <div style={{
            fontSize: 11,
            color: isPositive ? 'var(--up)' : 'var(--down)',
            opacity: 0.7,
          }}>
            {latest.main_net_pct >= 0 ? '+' : ''}{latest.main_net_pct.toFixed(2)}%
          </div>
        </div>

        <div style={{
          padding: 12,
          background: 'var(--bg-tertiary)',
          borderRadius: 'var(--radius-sm)',
        }}>
          <div style={{ fontSize: 10, color: 'var(--text-tertiary)', marginBottom: 8 }}>
            资金分布
          </div>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 4, fontSize: 11 }}>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span style={{ color: 'var(--text-tertiary)' }}>超大单</span>
              <span style={{ color: latest.super_large_net > 0 ? 'var(--up)' : 'var(--down)' }}>
                {flowText(latest.super_large_net)}
              </span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span style={{ color: 'var(--text-tertiary)' }}>大单</span>
              <span style={{ color: latest.large_net > 0 ? 'var(--up)' : 'var(--down)' }}>
                {flowText(latest.large_net)}
              </span>
            </div>
            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
              <span style={{ color: 'var(--text-tertiary)' }}>中小单</span>
              <span style={{ color: latest.medium_net > 0 ? 'var(--up)' : 'var(--down)' }}>
                {flowText(latest.medium_net)}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* Chart */}
      <CapitalFlowChart data={data} />
    </div>
  )
}

function CapitalFlowChart({ data }: { data: CapitalFlowData[] }) {
  const chartRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!chartRef.current || data.length === 0) return

    const chart = createChart(chartRef.current, {
      width: chartRef.current.clientWidth,
      height: 160,
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
        rightOffset: 2,
      },
      rightPriceScale: {
        borderColor: 'rgba(51, 65, 85, 0.5)',
      },
    })

    const series = chart.addSeries(LineSeries, {
      color: '#3b82f6',
      lineWidth: 2,
      priceLineVisible: false,
      lastValueVisible: false,
    })

    const chartData = data.map(d => ({
      time: d.date,
      value: d.main_net,
    }))

    series.setData(chartData)
    chart.timeScale().fitContent()

    const ro = new ResizeObserver((entries) => {
      for (const entry of entries) {
        chart.applyOptions({ width: entry.contentRect.width })
      }
    })
    ro.observe(chartRef.current)

    return () => {
      ro.disconnect()
      chart.remove()
    }
  }, [data])

  return <div ref={chartRef} style={{ width: '100%' }} />
}



