import { ref } from 'vue'

export interface WsMessage {
  type: 'prompt' | 'token' | 'done' | 'error' | 'tool_call' | 'tool_result' | 'cancel' | 'ping' | 'pong'
  payload?: string
  sessionId?: string
}

export const useWebSocket = () => {
  const ws = ref<WebSocket | null>(null)
  const connected = ref(false)
  const reconnectAttempts = ref(0)
  const maxReconnectAttempts = 5

  let onMessageCallback: ((msg: WsMessage) => void) | null = null
  let reconnectTimeout: ReturnType<typeof setTimeout> | null = null

  function connect(sessionId: string, onMessage: (msg: WsMessage) => void) {
    onMessageCallback = onMessage
    const wsUrl = `${import.meta.env.VITE_WS_URL || 'ws://localhost:8080'}/ws?session=${sessionId}`
    ws.value = new WebSocket(wsUrl)

    ws.value.onopen = () => {
      connected.value = true
      reconnectAttempts.value = 0
    }

    ws.value.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        onMessageCallback?.(data)
      } catch {
        onMessageCallback?.({ type: 'token', payload: event.data })
      }
    }

    ws.value.onclose = () => {
      connected.value = false
      attemptReconnect(sessionId)
    }

    ws.value.onerror = () => {
      connected.value = false
    }
  }

  function attemptReconnect(sessionId: string) {
    if (reconnectAttempts.value < maxReconnectAttempts) {
      const delay = Math.min(1000 * Math.pow(2, reconnectAttempts.value), 30000)
      reconnectTimeout = setTimeout(() => {
        reconnectAttempts.value++
        connect(sessionId, onMessageCallback!)
      }, delay)
    }
  }

  function send(data: WsMessage) {
    if (ws.value && connected.value) {
      ws.value.send(JSON.stringify(data))
    }
  }

  function sendPrompt(text: string, sessionId: string) {
    send({ type: 'prompt', payload: text, sessionId })
  }

  function disconnect() {
    if (reconnectTimeout) clearTimeout(reconnectTimeout)
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
    connected.value = false
  }

  return { connected, connect, send, sendPrompt, disconnect }
}