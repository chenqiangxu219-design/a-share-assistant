import { useParams, useNavigate } from 'react-router-dom'
import { Chart } from '../components/Chart'
import { SignalPanel } from '../components/SignalPanel'
import { IndicatorPanel } from '../components/IndicatorPanel'
import { CapitalFlow } from '../components/CapitalFlow'
import { DepthOfBook } from '../components/DepthOfBook'
import { useKLines, useSignals, useIndicators, useStockQuote } from '../hooks/useStockQuote'
import { ArrowLeft } from 'lucide-react'
import { useState } from 'react'

export function StockDetail() {
  const { code } = useParams<{ code: string }>()
  const navigate = useNavigate()
  const [period, setPeriod] = useState('d')
  const { data: klines } = useKLines(code || '', period, 120)
  const { data: signals } = useSignals(code || '')
  const { data: indicators } = useIndicators(code || '')
  const { data: quote } = useStockQuote(code || '')

  if (!code) return null

  return (
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
        <button className="btn" onClick={() => navigate(-1)} style={{ padding: '6px 12px', fontSize: 12 }}>
          <ArrowLeft size={13} />
          返回
        </button>
        <div style={{ flex: 1 }}>
          <div style={{ fontSize: 16, fontWeight: 700 }}>{quote?.name || code}</div>
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

      <div className="card" style={{ marginBottom: 16, padding: 12 }}>
        {klines && klines.length > 0 ? (
          <Chart data={klines} height={450} period={period} onPeriodChange={setPeriod} />
        ) : (
          <div style={{
            height: 450,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: 'var(--text-tertiary)',
            fontFamily: 'var(--font-mono)'
          }}>
            <span className="skeleton" style={{ width: 120, height: 14 }} />
          </div>
        )}
      </div>

      <div className="grid-2">
        <SignalPanel data={signals || {}} />
        <IndicatorPanel data={indicators || {}} />
      </div>
      <DepthOfBook code={code} />
      <CapitalFlow code={code} />
    </div>
  )
}
