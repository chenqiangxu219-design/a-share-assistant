import { create } from 'zustand'

interface Quote {
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

interface AlertItem {
  code: string
  name: string
  type: string
  message: string
  severity: string
  time: string
}

interface StoreState {
  watchlist: string[]
  quotes: Record<string, Quote>
  selectedStock: string | null
  alerts: AlertItem[]

  addToWatchlist: (code: string) => void
  removeFromWatchlist: (code: string) => void
  setWatchlist: (codes: string[]) => void
  setQuotes: (quotes: Record<string, Quote>) => void
  updateQuote: (quote: Quote) => void
  setSelectedStock: (code: string | null) => void
  addAlerts: (alerts: AlertItem[]) => void
  setAlerts: (alerts: AlertItem[]) => void
}

const EMPTY_ARRAY: string[] = []

export const useStore = create<StoreState>((set) => ({
  watchlist: EMPTY_ARRAY, // Dynamic - loaded from backend /api/stocks
  quotes: {},
  selectedStock: null,
  alerts: [],

  addToWatchlist: (code) =>
    set((state) => ({
      watchlist: state.watchlist.includes(code) ? state.watchlist : [...state.watchlist, code],
    })),

  removeFromWatchlist: (code) =>
    set((state) => ({
      watchlist: state.watchlist.filter((c) => c !== code),
    })),

  setWatchlist: (codes) => set({ watchlist: codes }),

  setQuotes: (quotes) => set({ quotes }),

  updateQuote: (quote) =>
    set((state) => ({
      quotes: { ...state.quotes, [quote.code]: quote },
    })),

  setSelectedStock: (code) => set({ selectedStock: code }),

  addAlerts: (newAlerts) =>
    set((state) => {
      const combined = [...state.alerts, ...newAlerts]
      return { alerts: combined.slice(-200) } // cap at 200
    }),

  setAlerts: (alerts) => set({ alerts }),
}))
