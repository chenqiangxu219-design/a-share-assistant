import { useQuery } from '@tanstack/react-query'
import { apiPath } from '../utils/api'

interface QuoteResponse {
  code: string
  name: string
  price: number
  change: number
  change_pct: number
  open: number
  high: number
  low: number
  yesterday_close: number
  volume: number
  turnover: number
  turnover_rate: number
  amount: number
  time: string
}

export function useStockQuote(code: string) {
  return useQuery<QuoteResponse>({
    queryKey: ['quote', code],
    queryFn: async () => {
      const res = await fetch(apiPath(`/api/quote/${code}`))
      if (!res.ok) throw new Error('Quote not found')
      return res.json()
    },
    staleTime: 30000, // 30s — WebSocket pushes real-time updates
    enabled: !!code,
  })
}

export function useMultiQuote(codes: string[]) {
  return useQuery<QuoteResponse[]>({
    queryKey: ['multi-quote', codes.join(',')],
    queryFn: async () => {
      const res = await fetch(apiPath(`/api/quote?codes=${codes.join(',')}`))
      if (!res.ok) throw new Error('Quote not found')
      return res.json()
    },
    staleTime: 30000, // 30s — WebSocket pushes real-time updates
    enabled: codes.length > 0,
  })
}

export function useKLines(code: string, period: string = 'd', count: number = 100) {
  return useQuery({
    queryKey: ['kline', code, period, count],
    queryFn: async () => {
      const res = await fetch(apiPath(`/api/kline/${code}?period=${period}&count=${count}`))
      if (!res.ok) throw new Error('KLine not found')
      return res.json()
    },
    refetchInterval: false,
    staleTime: 60000,
    enabled: !!code,
  })
}

export function useSignals(code: string) {
  return useQuery({
    queryKey: ['signals', code],
    queryFn: async () => {
      const res = await fetch(apiPath(`/api/signals/${code}`))
      if (!res.ok) throw new Error('Signals not found')
      return res.json()
    },
    staleTime: 60000, // signals change with K-line candles
    enabled: !!code,
  })
}

export function useIndicators(code: string) {
  return useQuery({
    queryKey: ['indicators', code],
    queryFn: async () => {
      const res = await fetch(apiPath(`/api/indicators/${code}`))
      if (!res.ok) throw new Error('Indicators not found')
      return res.json()
    },
    refetchInterval: false,
    staleTime: 60000,
    enabled: !!code,
  })
}
