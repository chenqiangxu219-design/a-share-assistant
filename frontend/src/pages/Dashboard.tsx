import { Watchlist } from '../components/Watchlist'
import { useStore } from '../store/store'
import { Chart } from '../components/Chart'
import { IntradayChart } from '../components/IntradayChart'
import { SignalPanel } from '../components/SignalPanel'
import { IndicatorPanel } from '../components/IndicatorPanel'
import { useKLines, useSignals, useIndicators, useStockQuote } from '../hooks/useStockQuote'
import { ArrowLeft, BarChart3, Activity } from 'lucide-react'
import { useState } from 'react'

export function Dashboard() {
  const selectedStock = useStore((s) => s.selectedStock)
  const setSelectedStock = useStore((s) => s.setSelectedStock)
  const [period, setPeriod] = useState('d')
  const [chartView, setChartView] = useState<'kline' | 'intraday'>('kline')
  const { data: klines } = useKLines(selectedStock || '', period, 120)
  const { data: intradayData } = useKLines(selectedStock || '', '5m', 78)
  const { data: quote } = useStockQuote(selectedStock || '')

  return (
    <div className="fade-in">
      {!selectedStock ? (
        <Watchlist />
      ) : (
        <div className="slide-up">
          <div style={{
            display: 'flex',
            gap: 16,
            marginBottom: 16,
            alignItems: 'center',
            padding: '12px 16px',
            background: 'var(--bg-card)',
            borderRadius: 'var(--radius-md)',
            border: '1px solid var(--border-subtle)'
          }}>
            <button
              className="btn"
              onClick={() => setSelectedStock(null)}
              style={{ padding: '6px 12px', fontSize: 12 }}
            >
              <ArrowLeft size={13} />
              返回
            </button>
            <div style={{ flex: 1 }}>
              <div style={{ fontSize: 16, fontWeight: 700 }}>{quote?.name || selectedStock}</div>
              <div style={{ fontSize: 11, color: 'var(--text-tertiary)', fontFamily: 'var(--font-mono)' }}>
                {quote?.code}
              </div>
            </div>
            {quote && (
              <div style={{ textAlign: 'right' }}>
                <div style={{
                  fontSize: 22,
                  fontWeight: 700,
                  fontFamily: 'var(--font-mono)',
                  color: quote.change_pct >= 0 ? 'var(--up)' : 'var(--down)'
                }}>
                  {quote.price.toFixed(2)}
                </div>
                <div style={{
                  fontSize: 13,
                  fontWeight: 600,
                  fontFamily: 'var(--font-mono)',
                  color: quote.change_pct >= 0 ? 'var(--up)' : 'var(--down)'
                }}>
                  {quote.change_pct >= 0 ? '+' : ''}{quote.change_pct.toFixed(2)}%
                </div>
              </div>
            )}
          </div>

          {/* Chart View Toggle */}
          <div style={{
            display: 'flex',
            gap: 4,
            marginBottom: 12,
            padding: 4,
            background: 'var(--bg-tertiary)',
            borderRadius: 'var(--radius-sm)',
            width: 'fit-content'
          }}>
            <button
              onClick={() => setChartView('kline')}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                padding: '6px 14px',
                border: 'none',
                borderRadius: 4,
                fontSize: 12,
                fontWeight: 600,
                cursor: 'pointer',
                background: chartView === 'kline' ? 'var(--accent-blue-dim)' : 'transparent',
                color: chartView === 'kline' ? 'var(--accent-blue)' : 'var(--text-tertiary)',
                transition: 'all 150ms',
                fontFamily: 'var(--font-sans)'
              }}
            >
              <BarChart3 size={13} />
              K 线
            </button>
            <button
              onClick={() => setChartView('intraday')}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                padding: '6px 14px',
                border: 'none',
                borderRadius: 4,
                fontSize: 12,
                fontWeight: 600,
                cursor: 'pointer',
                background: chartView === 'intraday' ? 'var(--accent-blue-dim)' : 'transparent',
                color: chartView === 'intraday' ? 'var(--accent-blue)' : 'var(--text-tertiary)',
                transition: 'all 150ms',
                fontFamily: 'var(--font-sans)'
              }}
            >
              <Activity size={13} />
              分时
            </button>
          </div>

          <div className="card" style={{ marginBottom: 16, padding: 12 }}>
            {chartView === 'kline' ? (
              klines && klines.length > 0 ? (
                <Chart data={klines} height={420} period={period} onPeriodChange={setPeriod} />
              ) : (
                <ChartLoading />
              )
            ) : (
              intradayData && intradayData.length > 2 ? (
                <IntradayChart
                  data={intradayData.map((d: any) => ({
                    time: d.date,
                    price: d.close,
                    volume: d.volume
                  }))}
                  yesterdayClose={quote?.yesterday_close || 0}
                  height={360}
                />
              ) : (
                <ChartLoading />
              )
            )}
          </div>

          <div className="grid-2">
            <SignalPlaceholder code={selectedStock} />
            <IndicatorPlaceholder code={selectedStock} />
          </div>
        </div>
      )}
    </div>
  )
}

function ChartLoading() {
  return (
    <div style={{
      height: 420,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      color: 'var(--text-tertiary)',
      fontFamily: 'var(--font-mono)'
    }}>
      <span className="skeleton" style={{ width: 120, height: 14 }} />
    </div>
  )
}

function SignalPlaceholder({ code }: { code: string }) {
  const { data } = useSignals(code)
  if (!data) return (
    <div className="card">
      <div className="card-title">信号分析</div>
      <div className="skeleton" style={{ width: '100%', height: 80 }} />
    </div>
  )
  return <SignalPanel data={data} />
}

function IndicatorPlaceholder({ code }: { code: string }) {
  const { data } = useIndicators(code)
  if (!data) return (
    <div className="card">
      <div className="card-title">技术指标</div>
      <div className="skeleton" style={{ width: '100%', height: 120 }} />
    </div>
  )
  return <IndicatorPanel data={data} />
}
