import { useState, useEffect } from 'react'
import { apiPath } from '../utils/api'
import { useStockQuote } from '../hooks/useStockQuote'
import { Wallet, TrendingUp, TrendingDown, ArrowUpRight, ArrowDownRight } from 'lucide-react'

interface Position {
  code: string
  name: string
  shares: number
  cost_price: number
  current_price: number
  pnL: number
  pnL_pct: number
}

interface PortfolioData {
  cash: number
  positions: Position[]
  trades: Array<{
    id: number
    code: string
    name: string
    direction: string
    price: number
    shares: number
    amount: number
    time: string
  }>
  total_value: number
  pnL: number
  pnL_pct: number
}

const StatCard = ({ title, value, icon: Icon, positive }: { title: string; value: string; icon: any; positive?: boolean }) => (
  <div className="card" style={{ padding: 14 }}>
    <div style={{ display: 'flex', alignItems: 'center', gap: 6, marginBottom: 8 }}>
      <Icon size={14} style={{ color: 'var(--text-tertiary)' }} />
      <div style={{
        fontSize: 10,
        color: 'var(--text-tertiary)',
        textTransform: 'uppercase',
        letterSpacing: '0.08em',
        fontWeight: 600
      }}>
        {title}
      </div>
    </div>
    <div style={{
      fontSize: 22,
      fontWeight: 700,
      fontFamily: 'var(--font-mono)',
      letterSpacing: '-0.02em'
    }} className={positive === true ? 'positive' : positive === false ? 'negative' : ''}>
      {value}
    </div>
  </div>
)

export function Portfolio() {
  const [data, setData] = useState<PortfolioData | null>(null)
  const [code, setCode] = useState('')
  const [shares, setShares] = useState('')
  const { data: liveQuote } = useStockQuote(code)

  const fetchPortfolio = async () => {
    const res = await fetch(apiPath('/api/portfolio'))
    if (res.ok) setData(await res.json())
  }

  useEffect(() => {
    fetchPortfolio()
    const interval = setInterval(fetchPortfolio, 5000)
    return () => clearInterval(interval)
  }, [])

  const handleTrade = async (direction: 'buy' | 'sell') => {
    const posPrice = data?.positions.find((p) => p.code === code)?.current_price
    const price = liveQuote?.price || posPrice || 0
    if (price === 0) return

    const res = await fetch(apiPath(`/api/portfolio/${direction}`), {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ code, price, shares: parseInt(shares) }),
    })
    if (res.ok) {
      setCode('')
      setShares('')
      fetchPortfolio()
    }
  }

  if (!data) return (
    <div className="card">
      <div className="card-title">模拟盘</div>
      <div className="skeleton" style={{ width: '100%', height: 100 }} />
    </div>
  )

  return (
    <div className="fade-in">
      <div className="grid-4" style={{ marginBottom: 16 }}>
        <StatCard title="总资产" value={data.total_value.toFixed(2)} icon={Wallet} />
        <StatCard title="可用资金" value={data.cash.toFixed(2)} icon={Wallet} />
        <StatCard title="总盈亏" value={`${data.pnL >= 0 ? '+' : ''}${data.pnL.toFixed(2)}`} icon={data.pnL >= 0 ? ArrowUpRight : ArrowDownRight} positive={data.pnL >= 0} />
        <StatCard title="收益率" value={`${data.pnL_pct >= 0 ? '+' : ''}${data.pnL_pct.toFixed(2)}%`} icon={data.pnL_pct >= 0 ? TrendingUp : TrendingDown} positive={data.pnL_pct >= 0} />
      </div>

      <div className="grid-2">
        <div className="card slide-up">
          <div className="card-title">持仓</div>
          {(!data.positions || data.positions.length === 0) ? (
            <div style={{ color: 'var(--text-tertiary)', fontSize: 13, textAlign: 'center', padding: '20px 0' }}>
              暂无持仓
            </div>
          ) : (
            <table className="data-table">
              <thead>
                <tr>
                  <th>代码</th><th>名称</th>
                  <th style={{ textAlign: 'right' }}>持仓</th>
                  <th style={{ textAlign: 'right' }}>成本</th>
                  <th style={{ textAlign: 'right' }}>现价</th>
                  <th style={{ textAlign: 'right' }}>盈亏</th>
                </tr>
              </thead>
              <tbody>
                {data.positions.map((p) => (
                  <tr key={p.code}>
                    <td>{p.code}</td>
                    <td style={{ fontFamily: 'var(--font-sans)', fontWeight: 500 }}>{p.name}</td>
                    <td style={{ textAlign: 'right' }}>{p.shares}</td>
                    <td style={{ textAlign: 'right' }}>{p.cost_price.toFixed(2)}</td>
                    <td style={{ textAlign: 'right' }}>{p.current_price.toFixed(2)}</td>
                    <td style={{
                      textAlign: 'right',
                      color: p.pnL >= 0 ? 'var(--up)' : 'var(--down)',
                      fontWeight: 600
                    }}>
                      {p.pnL >= 0 ? '+' : ''}{p.pnL.toFixed(2)} ({p.pnL_pct.toFixed(2)}%)
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          )}
        </div>

        <div className="card slide-up" style={{ animationDelay: '100ms' }}>
          <div className="card-title">交易</div>
          <div style={{ display: 'flex', gap: 8, marginBottom: liveQuote ? 4 : 16, flexWrap: 'wrap' }}>
            <input
              type="text"
              placeholder="股票代码"
              value={code}
              onChange={(e) => setCode(e.target.value.toUpperCase())}
              style={{ flex: 1, fontFamily: 'var(--font-mono)', minWidth: 100 }}
            />
            <input
              type="number"
              placeholder="数量"
              value={shares}
              onChange={(e) => setShares(e.target.value)}
              style={{ width: 80, fontFamily: 'var(--font-mono)' }}
            />
            <button className="btn btn-buy" onClick={() => handleTrade('buy')}>买入</button>
            <button className="btn btn-sell" onClick={() => handleTrade('sell')}>卖出</button>
          </div>
          {liveQuote && (
            <div style={{
              padding: '10px 14px',
              background: 'var(--bg-tertiary)',
              borderRadius: 'var(--radius-sm)',
              marginBottom: 16,
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center'
            }}>
              <div>
                <span style={{ fontSize: 13, fontWeight: 600, color: 'var(--text-primary)' }}>
                  {liveQuote.name}
                </span>
                <span style={{ fontSize: 11, color: 'var(--text-tertiary)', marginLeft: 8, fontFamily: 'var(--font-mono)' }}>
                  {liveQuote.code}
                </span>
              </div>
              <div style={{ textAlign: 'right' }}>
                <span style={{
                  fontSize: 16,
                  fontWeight: 700,
                  fontFamily: 'var(--font-mono)',
                  color: liveQuote.change_pct >= 0 ? 'var(--up)' : 'var(--down)'
                }}>
                  {liveQuote.price.toFixed(2)}
                </span>
                <span style={{
                  fontSize: 12,
                  marginLeft: 8,
                  color: liveQuote.change_pct >= 0 ? 'var(--up)' : 'var(--down)'
                }}>
                  {liveQuote.change_pct >= 0 ? '+' : ''}{liveQuote.change_pct.toFixed(2)}%
                </span>
              </div>
            </div>
          )}
          <div style={{
            maxHeight: 280,
            overflowY: 'auto',
            scrollbarWidth: 'thin',
            scrollbarColor: 'var(--border-subtle) transparent'
          }}>
            {((data.trades || [])).slice(-10).reverse().map((t) => (
              <div
                key={t.id}
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '8px 0',
                  fontSize: 12,
                  borderBottom: '1px solid rgba(51, 65, 85, 0.25)',
                }}
              >
                <span>
                  <span className={`tag tag-${t.direction}`} style={{ marginRight: 8 }}>
                    {t.direction === 'buy' ? '买' : '卖'}
                  </span>
                  <span style={{ fontFamily: 'var(--font-sans)', color: 'var(--text-secondary)' }}>
                    {t.name}
                  </span>
                  <span style={{ fontFamily: 'var(--font-mono)', color: 'var(--text-tertiary)', marginLeft: 4 }}>
                    {t.code}
                  </span>
                </span>
                <span style={{ color: 'var(--text-tertiary)', fontFamily: 'var(--font-mono)' }}>
                  {t.shares}股 @ {t.price.toFixed(2)}
                </span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  )
}
