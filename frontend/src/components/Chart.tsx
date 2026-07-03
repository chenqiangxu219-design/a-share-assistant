import { useEffect, useRef, useState } from 'react'
import { createChart, CandlestickSeries, HistogramSeries } from 'lightweight-charts'

interface ChartProps {
  data: Array<{
    date: string
    open: number
    high: number
    low: number
    close: number
    volume: number
  }>
  height?: number
  period?: string
  onPeriodChange?: (period: string) => void
}

const PERIODS = [
  { key: 'd', label: '日K' },
  { key: 'w', label: '周K' },
  { key: 'm', label: '月K' },
]

export function Chart({ data, height = 400, period = 'd', onPeriodChange }: ChartProps) {
  const [activePeriod, setActivePeriod] = useState(period)
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<any>(null)

  useEffect(() => {
    setActivePeriod(period)
  }, [period])

  useEffect(() => {
    if (!chartRef.current || data.length === 0) return

    if (chartInstance.current) {
      chartInstance.current.remove()
      chartInstance.current = null
    }

    const container = chartRef.current
    const w = container.clientWidth
    if (w === 0) return

    const chart = createChart(container, {
      width: w,
      height,
      layout: {
        background: { color: '#111827' },
        textColor: '#94a3b8',
      },
      grid: {
        vertLines: { color: 'rgba(51, 65, 85, 0.3)' },
        horzLines: { color: 'rgba(51, 65, 85, 0.3)' },
      },
      timeScale: {
        borderColor: 'rgba(51, 65, 85, 0.5)',
        rightOffset: 5,
      },
      rightPriceScale: {
        borderColor: 'rgba(51, 65, 85, 0.5)',
      },
    })

    chartInstance.current = chart

    const candleSeries = chart.addSeries(CandlestickSeries, {
      upColor: '#00d099',
      downColor: '#ff4756',
      borderUpColor: '#00d099',
      borderDownColor: '#ff4756',
      wickUpColor: '#00d099',
      wickDownColor: '#ff4756',
    })

    const volumeSeries = chart.addSeries(HistogramSeries, {
      priceScaleId: 'vol',
      priceFormat: { type: 'volume' },
    })

    volumeSeries.priceScale().applyOptions({
      scaleMargins: { top: 0.85, bottom: 0 },
    })

    const chartData = data.map((d) => ({
      time: d.date,
      open: d.open,
      high: d.high,
      low: d.low,
      close: d.close,
    }))

    const volumeData = data.map((d) => ({
      time: d.date,
      value: d.volume,
      color: d.close >= d.open ? 'rgba(0, 208, 153, 0.3)' : 'rgba(255, 71, 86, 0.3)',
    }))

    candleSeries.setData(chartData)
    volumeSeries.setData(volumeData)
    chart.timeScale().fitContent()

    const ro = new ResizeObserver((entries) => {
      for (const entry of entries) {
        if (chartInstance.current && entry.contentRect.width > 0) {
          chartInstance.current.applyOptions({ width: entry.contentRect.width })
        }
      }
    })
    ro.observe(container)

    return () => {
      ro.disconnect()
    }
  }, [data, height])

  const handlePeriodChange = (p: string) => {
    setActivePeriod(p)
    onPeriodChange?.(p)
  }

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: 4, marginBottom: 8 }}>
        {PERIODS.map((p) => (
          <button
            key={p.key}
            onClick={() => handlePeriodChange(p.key)}
            style={{
              padding: '4px 12px',
              border: 'none',
              borderRadius: 4,
              fontSize: 12,
              fontWeight: 600,
              cursor: 'pointer',
              background: activePeriod === p.key ? 'var(--accent-blue-dim)' : 'transparent',
              color: activePeriod === p.key ? 'var(--accent-blue)' : 'var(--text-tertiary)',
              fontFamily: 'var(--font-sans)',
            }}
          >
            {p.label}
          </button>
        ))}
      </div>
      <div ref={chartRef} style={{ width: '100%', height }} />
    </div>
  )
}
