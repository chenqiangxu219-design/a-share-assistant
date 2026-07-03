import { useEffect, useRef, useCallback } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useStore } from '../store/store'

export function useWebSocket() {
  const wsRef = useRef<WebSocket | null>(null)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)
  const updateQuote = useStore((s) => s.updateQuote)
  const addAlerts = useStore((s) => s.addAlerts)
  const queryClient = useQueryClient()

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.close()
    }
    if (timerRef.current) clearTimeout(timerRef.current)

    let url: string
    if (typeof window.electronAPI !== 'undefined' && window.electronAPI.isElectron) {
      url = 'ws://localhost:8080/ws'
    } else {
      const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
      url = `${protocol}//${window.location.host}/ws`
    }

    const ws = new WebSocket(url)
    wsRef.current = ws

    ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'quote') {
          const q = msg.data
          updateQuote(q)
          // Update React Query cache so components using useStockQuote stay in sync
          queryClient.setQueryData(['quote', q.code], q)
        } else if (msg.type === 'alert' && Array.isArray(msg.data)) {
          addAlerts(msg.data)
        }
      } catch {
        // Ignore malformed messages (heartbeats, etc.)
      }
    }

    ws.onclose = () => {
      timerRef.current = setTimeout(() => connect(), 3000)
    }

    ws.onerror = () => {
      ws.close()
    }
  }, [updateQuote, addAlerts, queryClient])

  useEffect(() => {
    connect()
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
      if (wsRef.current) wsRef.current.close()
    }
  }, [connect])

  return wsRef
}
