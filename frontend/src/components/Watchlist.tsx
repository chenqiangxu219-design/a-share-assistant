import { useState, useRef, useEffect } from 'react'
import { useStore } from '../store/store'
import { useMultiQuote } from '../hooks/useStockQuote'
import { useStockList, searchStocks } from '../hooks/useStockList'
import { Plus, Search, Layers, X } from 'lucide-react'
import { SECTOR_LISTS } from '../utils/stockList'

export function Watchlist() {
  const watchlist = useStore((s) => s.watchlist)
  const addToWatchlist = useStore((s) => s.addToWatchlist)
  const setSelectedStock = useStore((s) => s.setSelectedStock)
  // Sort watchlist for stable React Query key — prevents unnecessary refetches
  const { data: quotes } = useMultiQuote([...watchlist].sort())
  const { stocks } = useStockList()

  const [searchQuery, setSearchQuery] = useState('')
  const [showSearch, setShowSearch] = useState(false)
  const [showSectors, setShowSectors] = useState(false)
  const [selectedSector, setSelectedSector] = useState<string | null>(null)
  const searchRef = useRef<HTMLDivElement>(null)

  // Search results from merged stock list
  const searchResults = searchStocks(stocks, searchQuery)

  // Click outside to close dropdowns
  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (searchRef.current && !searchRef.current.contains(e.target as Node)) {
        setShowSearch(false)
        setShowSectors(false)
      }
    }
    document.addEventListener('mousedown', handler)
    return () => document.removeEventListener('mousedown', handler)
  }, [])

  const handleAddStock = (code: string) => {
    if (!watchlist.includes(code)) {
      addToWatchlist(code)
    }
    setSearchQuery('')
    setShowSearch(false)
  }

  const handleAddSector = (sector: string) => {
    const stocks = SECTOR_LISTS[sector]
    if (stocks) {
      stocks.forEach((s) => {
        if (!watchlist.includes(s.code)) {
          addToWatchlist(s.code)
        }
      })
    }
    setSelectedSector(null)
    setShowSectors(false)
  }

  const handleRemoveStock = (code: string, e: React.MouseEvent) => {
    e.stopPropagation()
    useStore.getState().removeFromWatchlist(code)
  }

  return (
    <div className="fade-in">
      <div className="card" style={{ marginBottom: 20 }}>
        <div className="card-title">自选股监控</div>
        <div style={{ display: 'flex', gap: 10, alignItems: 'center' }}>
          <div ref={searchRef} style={{ flex: 1, position: 'relative' }}>
            <input
              type="text"
              placeholder="搜索股票 (代码/名称/拼音)"
              value={searchQuery}
              onChange={(e) => {
                setSearchQuery(e.target.value)
                setShowSearch(true)
              }}
              onFocus={() => searchQuery && setShowSearch(true)}
              style={{ flex: 1, fontFamily: 'var(--font-mono)', paddingRight: 32 }}
            />
            {searchQuery && (
              <button
                onClick={() => { setSearchQuery(''); setShowSearch(false) }}
                style={{
                  position: 'absolute',
                  right: 8,
                  top: '50%',
                  transform: 'translateY(-50%)',
                  background: 'none',
                  border: 'none',
                  color: 'var(--text-tertiary)',
                  cursor: 'pointer',
                  padding: 4
                }}
              >
                <X size={14} />
              </button>
            )}

            {showSearch && searchResults.length > 0 && (
              <div style={{
                position: 'absolute',
                top: '100%',
                left: 0,
                right: 0,
                zIndex: 100,
                background: 'var(--bg-secondary)',
                border: '1px solid var(--border-subtle)',
                borderRadius: 'var(--radius-sm)',
                maxHeight: 240,
                overflowY: 'auto',
                marginTop: 4,
                boxShadow: '0 8px 24px rgba(0,0,0,0.4)'
              }}>
                {searchResults.map((stock) => {
                  const isAdded = watchlist.includes(stock.code)
                  return (
                    <div
                      key={stock.code}
                      onClick={() => handleAddStock(stock.code)}
                      style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        padding: '8px 12px',
                        cursor: 'pointer',
                        background: isAdded ? 'var(--accent-blue-dim)' : 'transparent',
                        borderBottom: '1px solid rgba(51,65,85,0.25)'
                      }}
                    >
                      <div>
                        <span style={{ fontWeight: 600, fontSize: 13 }}>{stock.name}</span>
                        <span style={{
                          fontSize: 11,
                          color: 'var(--text-tertiary)',
                          marginLeft: 8,
                          fontFamily: 'var(--font-mono)'
                        }}>
                          {stock.code}
                        </span>
                      </div>
                      {isAdded && (
                        <span style={{
                          fontSize: 10,
                          color: 'var(--accent-blue)',
                          background: 'var(--accent-blue-dim)',
                          padding: '2px 6px',
                          borderRadius: 4
                        }}>
                          已添加
                        </span>
                      )}
                    </div>
                  )
                })}
              </div>
            )}
          </div>

          <button
            className="btn"
            onClick={() => { setShowSectors(!showSectors); setShowSearch(false) }}
            style={{ position: 'relative' }}
          >
            <Layers size={14} />
            板块
          </button>

          <button className="btn btn-primary" onClick={() => {}}>
            <Plus size={14} />
            添加
          </button>
        </div>

        {/* Sector Dropdown */}
        {showSectors && (
          <div style={{
            marginTop: 16,
            padding: 16,
            background: 'var(--bg-tertiary)',
            borderRadius: 'var(--radius-sm)'
          }}>
            <div style={{
              fontSize: 11,
              color: 'var(--text-tertiary)',
              marginBottom: 12,
              textTransform: 'uppercase',
              letterSpacing: '0.05em',
              fontWeight: 600
            }}>
              批量导入板块成分股
            </div>
            <div style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(120px, 1fr))',
              gap: 8
            }}>
              {Object.keys(SECTOR_LISTS).map((sector) => (
                <button
                  key={sector}
                  className="btn"
                  onClick={() => handleAddSector(sector)}
                  style={{
                    justifyContent: 'center',
                    padding: '8px 12px',
                    fontSize: 13,
                    fontWeight: 600
                  }}
                >
                  {sector}
                  <span style={{
                    fontSize: 10,
                    color: 'var(--text-tertiary)',
                    marginLeft: 4
                  }}>
                    ({SECTOR_LISTS[sector].length}只)
                  </span>
                </button>
              ))}
            </div>
          </div>
        )}
      </div>

      <div style={{
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 1fr))',
        gap: 12
      }}>
        {watchlist.map((code, i) => {
          const quote = quotes?.find((q) => q.code === code)
          return (
            <WatchlistCard
              key={code}
              code={code}
              quote={quote || undefined}
              onClick={() => setSelectedStock(code)}
              delay={i * 50}
            />
          )
        })}
      </div>
    </div>
  )
}

function WatchlistCard({ code, quote, onClick, delay }: {
  code: string
  quote?: { code: string; name: string; price: number; change_pct: number; high: number; low: number; volume: number }
  onClick: () => void
  delay: number
}) {
  if (!quote) {
    return (
      <div className="quote-card slide-up" style={{ animationDelay: `${delay}ms` }}>
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

  return (
    <div
      className="quote-card slide-up"
      style={{ animationDelay: `${delay}ms` }}
      onClick={onClick}
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
