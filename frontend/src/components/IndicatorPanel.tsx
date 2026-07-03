interface IndicatorPanelProps {
  data: {
    ma?: { ma5: number[]; ma10: number[]; ma20: number[]; ma60: number[] }
    macd?: { dif: number[]; dea: number[]; hist: number[] }
    rsi?: { rsi6: number[]; rsi12: number[]; rsi24: number[] }
    kdj?: { k: number[]; d: number[]; j: number[] }
    boll?: { mid: number[]; upper: number[]; lower: number[] }
  }
}

function getLast(arr: number[] | undefined): number | null {
  if (!arr || arr.length === 0) return null
  for (let i = arr.length - 1; i >= 0; i--) {
    if (arr[i] !== 0) return arr[i]
  }
  return arr[arr.length - 1]
}

const Row = ({ label, value, color, mono = true }: { label: string; value: string; color?: string; mono?: boolean }) => (
  <div style={{
    display: 'flex',
    justifyContent: 'space-between',
    padding: '5px 0',
    fontSize: 12,
    borderBottom: '1px solid rgba(51, 65, 85, 0.15)'
  }}>
    <span style={{ color: 'var(--text-tertiary)', fontWeight: 500 }}>{label}</span>
    <span style={{
      color: color || 'var(--text-primary)',
      fontFamily: mono ? 'var(--font-mono)' : 'var(--font-sans)',
      fontWeight: 600
    }}>
      {value}
    </span>
  </div>
)

const Section = ({ title, children }: { title: string; children: React.ReactNode }) => (
  <div>
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
    {children}
  </div>
)

export function IndicatorPanel({ data }: IndicatorPanelProps) {
  const ma5 = getLast(data.ma?.ma5)
  const ma10 = getLast(data.ma?.ma10)
  const ma20 = getLast(data.ma?.ma20)
  const ma60 = getLast(data.ma?.ma60)
  const dif = getLast(data.macd?.dif)
  const dea = getLast(data.macd?.dea)
  const hist = getLast(data.macd?.hist)
  const rsi6 = getLast(data.rsi?.rsi6)
  const rsi12 = getLast(data.rsi?.rsi12)
  const rsi24 = getLast(data.rsi?.rsi24)
  const k = getLast(data.kdj?.k)
  const d = getLast(data.kdj?.d)
  const j = getLast(data.kdj?.j)
  const bollMid = getLast(data.boll?.mid)
  const bollUp = getLast(data.boll?.upper)
  const bollLow = getLast(data.boll?.lower)

  return (
    <div className="card slide-up" style={{ animationDelay: '100ms' }}>
      <div className="card-title">技术指标</div>
      <div style={{
        display: 'grid',
        gridTemplateColumns: '1fr 1fr',
        gap: '0 24px'
      }}>
        <Section title="均线 MA">
          <Row label="MA5" value={ma5?.toFixed(2) || '-'} color="#3b82f6" />
          <Row label="MA10" value={ma10?.toFixed(2) || '-'} color="#f59e0b" />
          <Row label="MA20" value={ma20?.toFixed(2) || '-'} color="#8b5cf6" />
          <Row label="MA60" value={ma60?.toFixed(2) || '-'} color="#00d099" />
        </Section>

        <Section title="MACD">
          <Row label="DIF" value={dif?.toFixed(4) || '-'} color={hist && hist > 0 ? 'var(--up)' : 'var(--down)'} />
          <Row label="DEA" value={dea?.toFixed(4) || '-'} />
          <Row label="HIST" value={hist?.toFixed(4) || '-'} color={hist && hist > 0 ? 'var(--up)' : 'var(--down)'} />
        </Section>

        <Section title="RSI">
          <Row label="RSI6" value={rsi6?.toFixed(2) || '-'}
            color={rsi6 !== null && rsi6 < 30 ? 'var(--up)' : rsi6 !== null && rsi6 > 70 ? 'var(--down)' : undefined} />
          <Row label="RSI12" value={rsi12?.toFixed(2) || '-'} />
          <Row label="RSI24" value={rsi24?.toFixed(2) || '-'} />
        </Section>

        <Section title="KDJ">
          <Row label="K" value={k?.toFixed(2) || '-'} />
          <Row label="D" value={d?.toFixed(2) || '-'} />
          <Row label="J" value={j?.toFixed(2) || '-'}
            color={j !== null && j < 0 ? 'var(--up)' : j !== null && j > 100 ? 'var(--down)' : undefined} />
        </Section>

        <Section title="布林带 BOLL">
          <Row label="上轨" value={bollUp?.toFixed(2) || '-'} color="#f59e0b" />
          <Row label="中轨" value={bollMid?.toFixed(2) || '-'} color="#3b82f6" />
          <Row label="下轨" value={bollLow?.toFixed(2) || '-'} color="#00d099" />
        </Section>
      </div>
    </div>
  )
}
