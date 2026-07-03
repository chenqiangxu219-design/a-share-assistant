import { useEffect, useRef } from 'react'
import { createChart, LineSeries } from 'lightweight-charts'

interface IntradayChartProps {
  data: Array<{
    time: string
    price: number
    volume: number
  }>
  yesterdayClose: number
  height?: number
}

export function IntradayChart({ data, yesterdayClose, height = 280 }: IntradayChartProps) {
  const chartRef = useRef<HTMLDivElement>(null)
  const chartInstance = useRef<any>(null)

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
        rightOffset: 2,
      },
      rightPriceScale: {
        borderColor: 'rgba(51, 65, 85, 0.5)',
      },
    })

    chartInstance.current = chart

    const lineS = chart.addSeries(LineSeries, {
      color: '#58a6ff',
      lineWidth: 2,
      priceLineVisible: false,
      lastValueVisible: false,
    })

    const refS = chart.addSeries(LineSeries, {
      color: 'rgba(148, 163, 184, 0.3)',
      lineWidth: 1,
      lineStyle: 2,
      priceLineVisible: false,
      lastValueVisible: false,
    })

    const priceData = data.map((d) => ({ time: d.time, value: d.price }))
    const refData = data.map((d) => ({ time: d.time, value: yesterdayClose }))

    lineS.setData(priceData)
    refS.setData(refData)
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
  }, [data, yesterdayClose, height])

  return <div ref={chartRef} style={{ width: '100%', height }} />
}
