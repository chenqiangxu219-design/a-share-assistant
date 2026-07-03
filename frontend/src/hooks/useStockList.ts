import { useQuery } from '@tanstack/react-query'
import { apiPath } from '../utils/api'
import { STOCK_DATABASE } from '../utils/stockList'

interface StockInfo {
  code: string
  name: string
}

// Build a merged stock list: API results + hardcoded fallback (deduped by code)
function mergeStocks(apiStocks: StockInfo[]): StockInfo[] {
  const map = new Map<string, string>() // code → name

  // Hardcoded database first (always available)
  for (const [code, name] of STOCK_DATABASE) {
    if (!map.has(code)) {
      map.set(code, name)
    }
  }

  // API results override/add
  for (const s of apiStocks) {
    if (!map.has(s.code)) {
      map.set(s.code, s.name)
    }
  }

  return Array.from(map.entries()).map(([code, name]) => ({ code, name }))
}

export function useStockList() {
  const { data: apiStocks } = useQuery({
    queryKey: ['stock-list'],
    queryFn: async () => {
      const res = await fetch(apiPath('/api/stocks'))
      if (!res.ok) throw new Error('Failed to fetch stock list')
      const data = await res.json()
      return data.stocks as StockInfo[]
    },
    staleTime: 3600000, // 1 hour — stock list doesn't change often
    refetchOnWindowFocus: false,
  })

  const allStocks = apiStocks ? mergeStocks(apiStocks) : STOCK_DATABASE.map(([code, name]) => ({ code, name }))

  return {
    stocks: allStocks,
    isLoading: !apiStocks, // technically loading until first API fetch, but hardcoded data is available immediately
  }
}

// Search stocks by code/name/pinyin
export function searchStocks(stocks: StockInfo[], query: string): StockInfo[] {
  const q = query.toLowerCase().trim()
  if (!q) return []

  return stocks
    .filter((s) => s.code.includes(q) || s.name.toLowerCase().includes(q))
    .slice(0, 10)
}
