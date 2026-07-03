import { useState } from 'react'
import { useStore } from '../store/store'
import { useMultiQuote } from '../hooks/useStockQuote'
import { apiPath } from '../utils/api'
import { SECTOR_LISTS } from '../utils/stockList'
import { TrendingUp, TrendingDown } from 'lucide-react'

export function Sectors() {
  const addToWatchlist = useStore((s) => s.addToWatchlist)
  const [selectedSector, setSelectedSector] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  // Calculate sector performance
  const sectorData = Object.entries(SECTOR_LISTS).map(([name, stocks]) => {
    const codes = stocks.map(s => s.code)
    const { data: quotes } = useMultiQuote(codes)

    if (!quotes || quotes.length === 0) {
      return { name, stocks, change: 0, avgPrice: 0, upCount: 0, total: stocks.length }
    }

    const totalChange = quotes.reduce((sum, q) => sum + q.change_pct, 0)
    const avgChange = totalChange / quotes.length
    const upCount = quotes.filter(q => q.change_pct > 0).length

    return {
      name,
      stocks,
      change: avgChange,
      upCount,
      total: stocks.length,
    }
  })

  const handleAddSector = (name: string) => {
    const stocks = SECTOR_LISTS[name]
    if (stocks) {
      stocks.forEach(s => {
        if (!useStore.getState().watchlist.includes(s.code)) {
          addToWatchlist(s.code)
        }
      })
    }
  }

  const selectedStocks = selectedSector ? SECTOR_LISTS[selectedSector] : []
  const selectedCodes = selectedStocks.map(s => s.code)
  const { data: selectedQuotes } = useMultiQuote(selectedCodes)

  return (
    <div className="fade-in">
      <div style={{ marginBottom: 20 }}>
        <div style={{
          fontSize: 18,
          fontWeight: 700,
          marginBottom: 4,
        }}>
          板块监控
        </div>
        <div style={{
          fontSize: 12,
          color: 'var(--text-tertiary)',
        }}>
          实时追踪热门板块表现，点击板块查看详情
        </div>
      </div>

      {!selectedSector ? (
        <>
          {/* Sector Cards */}
          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(200px, 1fr))',
            gap: 12,
            marginBottom: 20,
          }}>
            {sectorData.map((sector) => (
              <div
                key={sector.name}
                className="card"
                style={{ cursor: 'pointer', padding: 16 }}
                onClick={() => setSelectedSector(sector.name)}
              >
                <div style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  marginBottom: 10,
                }}>
                  <span style={{ fontWeight: 600, fontSize: 14 }}>{sector.name}</span>
                  {sector.change >= 0 ? (
                    <TrendingUp size={16} className="positive" />
                  ) : (
                    <TrendingDown size={16} className="negative" />
                  )}
                </div>
                <div style={{
                  fontSize: 22,
                  fontWeight: 700,
                  fontFamily: 'var(--font-mono)',
                  color: sector.change >= 0 ? 'var(--up)' : 'var(--down)',
                  marginBottom: 8,
                }}>
                  {sector.change >= 0 ? '+' : ''}{sector.change.toFixed(2)}%
                </div>
                <div style={{
                  fontSize: 11,
                  color: 'var(--text-tertiary)',
                }}>
                  {sector.upCount}涨 {sector.total - sector.upCount}跌 / {sector.total}只
                </div>
                <button
                  className="btn"
                  style={{
                    marginTop: 10,
                    width: '100%',
                    justifyContent: 'center',
                    padding: '6px 0',
                    fontSize: 11,
                  }}
                  onClick={(e) => {
                    e.stopPropagation()
                    handleAddSector(sector.name)
                  }}
                >
                  一键加入自选
                </button>
              </div>
            ))}
          </div>
        </>
      ) : (
        <>
          {/* Sector Detail */}
          <div style={{
            display: 'flex',
            alignItems: 'center',
            gap: 12,
            marginBottom: 16,
          }}>
            <button className="btn" onClick={() => setSelectedSector(null)}>
              ← 返回板块
            </button>
            <span style={{ fontWeight: 700, fontSize: 16 }}>{selectedSector}</span>
            <button
              className="btn btn-primary"
              onClick={() => handleAddSector(selectedSector)}
            >
              一键加入自选
            </button>
          </div>

          {/* Stock Table */}
          <div className="card">
            <table className="data-table">
              <thead>
                <tr>
                  <th>代码</th>
                  <th>名称</th>
                  <th style={{ textAlign: 'right' }}>价格</th>
                  <th style={{ textAlign: 'right' }}>涨跌</th>
                  <th style={{ textAlign: 'right' }}>量比</th>
                  <th style={{ textAlign: 'center' }}>操作</th>
                </tr>
              </thead>
              <tbody>
                {selectedStocks.map((stock) => {
                  const quote = selectedQuotes?.find(q => q.code === stock.code)
                  if (!quote) return null
                  const isUp = quote.change_pct >= 0
                  return (
                    <tr key={stock.code}>
                      <td>{stock.code}</td>
                      <td style={{ fontFamily: 'var(--font-sans)', fontWeight: 500 }}>
                        {stock.name}
                      </td>
                      <td style={{ textAlign: 'right' }}>{quote.price.toFixed(2)}</td>
                      <td style={{
                        textAlign: 'right',
                        color: isUp ? 'var(--up)' : 'var(--down)',
                        fontWeight: 600,
                      }}>
                        {isUp ? '+' : ''}{quote.change_pct.toFixed(2)}%
                      </td>
                      <td style={{ textAlign: 'right' }}>
                        {(quote.volume / 10000).toFixed(0)}万
                      </td>
                      <td style={{ textAlign: 'center' }}>
                        <button
                          className="btn"
                          style={{ padding: '2px 8px', fontSize: 11 }}
                          onClick={() => addToWatchlist(stock.code)}
                        >
                          加入自选
                        </button>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  )
}
