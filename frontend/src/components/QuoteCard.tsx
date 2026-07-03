import { useState, useEffect, useRef, useCallback } from 'react'
import { useStore } from '../store/store'
import { useStockQuote } from '../hooks/useStockQuote'

interface QuoteCardProps {
  code: string
}

export function QuoteCard({ code }: QuoteCardProps) {
  const { data: quote } = useStockQuote(code)
  const setSelectedStock = useStore((s) => s.setSelectedStock)
  const storeQuotes = useStore((s) => s.quotes)
  const prevPriceRef = useRef<number | null>(null)
  const [flashDir, setFlashDir] = useState<'up' | 'down' | null>(null)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const storeQuote = storeQuotes[code]
  const currentPrice = quote?.price ?? storeQuote?.price ?? null

  useEffect(() => {
    if (currentPrice === null || prevPriceRef.current === null) {
      if (currentPrice !== null) prevPriceRef.current = currentPrice
      return
    }
    if (currentPrice !== prevPriceRef.current) {
      setFlashDir(currentPrice > prevPriceRef.current ? 'up' : 'down')
      if (timerRef.current) clearTimeout(timerRef.current)
      timerRef.current = setTimeout(() => setFlashDir(null), 600)
      prevPriceRef.current = currentPrice
    }
  }, [currentPrice])

  if (!quote) {
    return (
      <div className="quote-card">
        <div className="quote-card-header">
          <div>
            <div className="skeleton" style={{ width: 80, height: 14, marginBottom: 6 }} />
            <div className="skeleton" style={{ width: 50, height: 10 }} />
          </div>
          <div style={{ textAlign: 'right' }}>
            <div className="skeleton" style={{ width: 70, height: 20, marginBottom: 4 }} />
            <div className="skeleton" style={{ width: 45, height: 12, marginLeft: 'auto' }} />
          </div>
        </div>
        <div className="quote-card-footer">
          <div className="skeleton" style={{ width: 40, height: 10 }} />
          <div className="skeleton" style={{ width: 40, height: 10 }} />
          <div className="skeleton" style={{ width: 50, height: 10 }} />
        </div>
      </div>
    )
  }

  const isUp = quote.change_pct >= 0
  const changeSign = isUp ? '+' : ''
  const flashClass = flashDir === 'up' ? 'glow-up' : flashDir === 'down' ? 'glow-down' : ''

  return (
    <div
      className={`quote-card ${flashClass}`}
      onClick={() => setSelectedStock(code)}
    >
      <div className="quote-card-header">
        <div>
          <div className="quote-card-name">{quote.name}</div>
          <div className="quote-card-code">{quote.code}</div>
        </div>
        <div style={{ textAlign: 'right' }}>
          <div className={`quote-card-price ${isUp ? 'positive' : 'negative'}`}>
            {quote.price.toFixed(2)}
          </div>
          <span className={`quote-card-change ${isUp ? 'up' : 'down'}`}>
            {changeSign}{quote.change_pct.toFixed(2)}%
          </span>
        </div>
      </div>
      <div className="quote-card-footer">
        <span>高 {quote.high.toFixed(2)}</span>
        <span>低 {quote.low.toFixed(2)}</span>
        <span>量 {formatVolume(quote.volume)}</span>
      </div>
    </div>
  )
}

function formatVolume(v: number): string {
  if (v >= 100000000) return (v / 100000000).toFixed(1) + '亿'
  if (v >= 10000) return (v / 10000).toFixed(0) + '万'
  return v.toFixed(0)
}
