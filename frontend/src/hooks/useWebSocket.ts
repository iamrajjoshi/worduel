import { useCallback, useEffect, useRef, useState } from 'react'
import type { ServerMessage } from '../types/messages'

export type ConnectionStatus = 'connecting' | 'connected' | 'disconnected' | 'reconnecting' | 'failed'

export type SendFn = (type: string, data: Record<string, unknown>) => void

interface UseWebSocketOptions {
  onMessage: (message: ServerMessage) => void
  onConnected?: (send: SendFn) => void
  enabled?: boolean
}

export function useWebSocket({ onMessage, onConnected, enabled = true }: UseWebSocketOptions) {
  const [status, setStatus] = useState<ConnectionStatus>('disconnected')
  const [clientId, setClientId] = useState<string | null>(null)
  const wsRef = useRef<WebSocket | null>(null)
  const onMessageRef = useRef(onMessage)
  const onConnectedRef = useRef(onConnected)
  const reconnectAttempts = useRef(0)
  const reconnectTimer = useRef<ReturnType<typeof setTimeout>>(undefined)
  const intentionalClose = useRef(false)
  const pendingConnectedCallback = useRef(false)

  onMessageRef.current = onMessage
  onConnectedRef.current = onConnected

  const getWsUrl = useCallback(() => {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    return `${protocol}//${window.location.host}/ws`
  }, [])

  const sendViaRef: SendFn = useCallback((type, data) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify({ type, data, timestamp: new Date().toISOString() }))
    }
  }, [])

  const connect = useCallback(() => {
    if (wsRef.current?.readyState === WebSocket.OPEN) return

    intentionalClose.current = false
    setStatus(reconnectAttempts.current > 0 ? 'reconnecting' : 'connecting')

    const ws = new WebSocket(getWsUrl())
    wsRef.current = ws

    ws.onopen = () => {
      setStatus('connected')
      reconnectAttempts.current = 0
    }

    ws.onmessage = (event) => {
      try {
        const message: ServerMessage = JSON.parse(event.data)
        if (message.type === 'connection_ack') {
          const data = message.data as { client_id: string }
          setClientId(data.client_id)
          pendingConnectedCallback.current = true
        }
        onMessageRef.current(message)
        if (pendingConnectedCallback.current) {
          pendingConnectedCallback.current = false
          onConnectedRef.current?.(sendViaRef)
        }
      } catch {
        // ignore malformed messages
      }
    }

    ws.onclose = () => {
      wsRef.current = null
      if (intentionalClose.current) {
        setStatus('disconnected')
        return
      }
      if (reconnectAttempts.current < 10) {
        const delay = Math.min(1000 * 2 ** reconnectAttempts.current, 30000)
        reconnectAttempts.current++
        setStatus('reconnecting')
        reconnectTimer.current = setTimeout(connect, delay)
      } else {
        setStatus('failed')
      }
    }

    ws.onerror = () => {
      // onclose will fire after this
    }
  }, [getWsUrl, sendViaRef])

  const disconnect = useCallback(() => {
    intentionalClose.current = true
    clearTimeout(reconnectTimer.current)
    reconnectAttempts.current = 0
    wsRef.current?.close()
    wsRef.current = null
    setStatus('disconnected')
    setClientId(null)
  }, [])

  // This single useEffect manages the WebSocket lifecycle (external system sync - the intended use)
  useEffect(() => {
    if (enabled) {
      connect()
    }
    return () => {
      intentionalClose.current = true
      clearTimeout(reconnectTimer.current)
      wsRef.current?.close()
    }
  }, [enabled, connect])

  return { status, clientId, send: sendViaRef, disconnect }
}
