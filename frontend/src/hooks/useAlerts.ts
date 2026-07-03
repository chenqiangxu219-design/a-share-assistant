import { useEffect, useRef } from 'react'
import { apiPath } from '../utils/api'
import { useStore } from '../store/store'

export type AlertItem = {
  code: string
  name: string
  type: string
  message: string
  severity: string
  time: string
}

let fetchOnce = false

export function useAlerts() {
  const alerts = useStore((s) => s.alerts)
  const setAlerts = useStore((s) => s.setAlerts)
  const fetched = useRef(false)

  // Fetch initial alerts once on mount
  useEffect(() => {
    if (fetched.current) return
    fetched.current = true

    fetch(apiPath('/api/alerts'))
      .then((r) => r.json())
      .then((data) => {
        if (Array.isArray(data)) {
          setAlerts(data)
        }
      })
      .catch(() => {})

    return () => {
      fetched.current = false
    }
  }, [setAlerts])

  return { alerts }
}
