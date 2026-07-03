interface SignalPanelProps {
  data: {
    composite?: {
      direction: string
      score: number
      strength: number
      message: string
      indicators: string[]
    }
    signals?: Array<{
      direction: string
      strength: number
      message: string
      indicators: string[]
    }>
  }
}

export function SignalPanel({ data }: SignalPanelProps) {
  const { composite, signals } = data

  if (!composite && !signals) {
    return (
      <div className="card">
        <div className="card-title">信号分析</div>
        <div className="skeleton" style={{ width: '100%', height: 80 }} />
      </div>
    )
  }

  const scoreColor = (score: number) => {
    if (score > 5) return 'var(--up)'
    if (score > 0) return 'rgba(0, 208, 153, 0.7)'
    if (score < -5) return 'var(--down)'
    if (score < 0) return 'rgba(255, 71, 86, 0.7)'
    return 'var(--text-secondary)'
  }

  const scoreBarWidth = (score: number) => {
    return `${Math.min(Math.abs(score) / 10 * 100, 100)}%`
  }

  return (
    <div className="card slide-up">
      <div className="card-title">信号分析</div>

      {composite && (
        <div style={{ marginBottom: 16 }}>
          <div style={{
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center',
            marginBottom: 8
          }}>
            <span className={`tag tag-${composite.direction}`}>
              {composite.direction === 'buy' ? '● 买入' : composite.direction === 'sell' ? '● 卖出' : '● 中性'}
            </span>
            <span style={{
              fontSize: 28,
              fontWeight: 700,
              fontFamily: 'var(--font-mono)',
              color: scoreColor(composite.score),
              letterSpacing: '-0.02em'
            }}>
              {composite.score > 0 ? '+' : ''}{composite.score}
            </span>
          </div>
          <div className="score-bar">
            <div
              className="score-fill"
              style={{
                width: scoreBarWidth(composite.score),
                background: scoreColor(composite.score),
                marginLeft: composite.score < 0 ? 'auto' : 0,
              }}
            />
          </div>
          <div style={{
            marginTop: 10,
            fontSize: 12,
            color: 'var(--text-secondary)',
            lineHeight: 1.5
          }}>
            {composite.message}
          </div>
        </div>
      )}

      {signals && signals.length > 0 && (
        <div>
          <div style={{
            fontSize: 11,
            color: 'var(--text-tertiary)',
            marginBottom: 8,
            textTransform: 'uppercase',
            letterSpacing: '0.05em'
          }}>
            详细信号
          </div>
          {signals.slice(0, 8).map((sig, i) => (
            <div
              key={i}
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: '8px 0',
                borderBottom: '1px solid rgba(51, 65, 85, 0.25)',
              }}
            >
              <div style={{ flex: 1, minWidth: 0 }}>
                <div style={{
                  fontSize: 12,
                  color: 'var(--text-secondary)',
                  overflow: 'hidden',
                  textOverflow: 'ellipsis',
                  whiteSpace: 'nowrap'
                }}>
                  {sig.message}
                </div>
              </div>
              <span className={`tag tag-${sig.direction}`} style={{ flexShrink: 0, marginLeft: 8 }}>
                {sig.direction === 'buy' ? '买' : '卖'} {sig.strength}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
