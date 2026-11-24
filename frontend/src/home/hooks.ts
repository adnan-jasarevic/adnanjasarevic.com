import { useCallback, useEffect, useState } from 'react'
import { API_URL, getJSON } from '../api'

export type Resource<T> =
  | { status: 'loading' }
  | { status: 'error'; retry: () => void }
  | { status: 'ready'; data: T }

export function useResource<T>(path: string): Resource<T> {
  const [attempt, setAttempt] = useState(0)
  const [state, setState] = useState<Resource<T>>({ status: 'loading' })
  const retry = useCallback(() => {
    setState({ status: 'loading' })
    setAttempt((n) => n + 1)
  }, [])

  useEffect(() => {
    const controller = new AbortController()
    getJSON<T>(path, controller.signal)
      .then((data) => setState({ status: 'ready', data }))
      .catch(() => {
        if (!controller.signal.aborted) setState({ status: 'error', retry })
      })
    return () => controller.abort()
  }, [path, attempt, retry])

  return state
}

export type ApiStatus =
  | { state: 'checking' }
  | { state: 'waking' }
  | { state: 'up'; ms: number; wokeSeconds?: number }
  | { state: 'down' }

// Anything slower than this is a cold start rather than a normal round trip.
const WAKING_AFTER_MS = 1500

// Pinging /health also wakes the API, so it is ready by the time someone sends the form.
export function useApiStatus(): ApiStatus {
  const [status, setStatus] = useState<ApiStatus>({ state: 'checking' })

  useEffect(() => {
    const controller = new AbortController()
    const started = performance.now()
    const wakingTimer = setTimeout(() => setStatus({ state: 'waking' }), WAKING_AFTER_MS)

    const ping = async () => {
      const t0 = performance.now()
      const res = await fetch(API_URL + '/health', {
        cache: 'no-store',
        signal: AbortSignal.any([controller.signal, AbortSignal.timeout(90_000)]),
      })
      if (!res.ok) throw new Error(String(res.status))
      return performance.now() - t0
    }

    ping()
      .then(async (first) => {
        if (first < WAKING_AFTER_MS) {
          setStatus({ state: 'up', ms: Math.round(first) })
          return
        }
        // The first ping included the wake-up; measure a warm round trip too.
        const warm = await ping().catch(() => null)
        const wokeSeconds = Math.round((performance.now() - started) / 1000)
        setStatus(warm === null ? { state: 'up', ms: 0, wokeSeconds } : { state: 'up', ms: Math.round(warm), wokeSeconds })
      })
      .catch(() => {
        if (!controller.signal.aborted) setStatus({ state: 'down' })
      })
      .finally(() => clearTimeout(wakingTimer))

    return () => {
      controller.abort()
      clearTimeout(wakingTimer)
    }
  }, [])

  return status
}
