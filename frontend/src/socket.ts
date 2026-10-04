import { API_BASE, WS_BASE } from './config'
import { S } from './strings'
import type { ServerMessage } from './types'

type MsgHandler = (m: ServerMessage) => void
type StatusHandler = (
  s: 'connecting' | 'connected' | 'reconnecting' | 'error',
  err?: string,
) => void

// After this many consecutive failed reconnects we stop and surface an error.
const MAX_RECONNECTS = 4

// Player identity is persisted per room in sessionStorage. This survives a
// page reload but is cleared when the tab is closed, and is isolated per
// tab so two tabs on the same room don't hijack each other.
const playerIdKey = (code: string) => `harady:player_id:${code}`

function getOrCreatePlayerId(code: string): string {
  try {
    const existing = sessionStorage.getItem(playerIdKey(code))
    if (existing) return existing
    const id = (crypto as { randomUUID?: () => string }).randomUUID?.() ??
      `p-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
    sessionStorage.setItem(playerIdKey(code), id)
    return id
  } catch {
    // Private mode / quota — fall back to no persistence.
    return `p-${Date.now()}-${Math.random().toString(36).slice(2, 10)}`
  }
}

let ws: WebSocket | null = null
let onMsg: MsgHandler | null = null
let onStatus: StatusHandler | null = null
let currentCode: string | null = null
let currentNick: string | null = null
let retry = 0
let retryTimer: number | null = null
let manuallyClosed = false

export function setHandlers(msg: MsgHandler, status: StatusHandler) {
  onMsg = msg
  onStatus = status
}

export function connect(code: string, nickname: string) {
  currentCode = code
  currentNick = nickname
  manuallyClosed = false
  retry = 0
  void openWithPrecheck()
}

export function disconnect() {
  manuallyClosed = true
  if (retryTimer !== null) {
    window.clearTimeout(retryTimer)
    retryTimer = null
  }
  if (ws) {
    try { ws.close() } catch { /* ignore */ }
    ws = null
  }
}

export function send(type: string, data?: unknown) {
  if (!ws || ws.readyState !== WebSocket.OPEN) return
  ws.send(JSON.stringify(data === undefined ? { type } : { type, data }))
}

// ---- pre-check ------------------------------------------------------------

// Ask the REST API whether the room exists before opening the WS.
// A WS handshake against a missing room yields only an opaque 1006 close
// event, which is indistinguishable from a transient network blip; the REST
// call gives us a definitive answer up front.
async function verifyRoom(code: string): Promise<'ok' | 'not_found' | 'unreachable'> {
  try {
    const res = await fetch(`${API_BASE}/api/rooms/${encodeURIComponent(code)}`, {
      method: 'GET',
      headers: { Accept: 'application/json' },
    })
    if (res.status === 404) return 'not_found'
    if (res.ok) return 'ok'
    return 'unreachable'
  } catch {
    return 'unreachable'
  }
}

async function openWithPrecheck() {
  if (!currentCode || !currentNick) return
  onStatus?.('connecting')

  const v = await verifyRoom(currentCode)
  if (manuallyClosed) return

  if (v === 'not_found') {
    onStatus?.('error', S.roomNotFound)
    return
  }
  // ok or unreachable — try the WS anyway; if the server is truly down the
  // reconnect cap below will eventually surface an error.
  open()
}

// ---- WS open + reconnect --------------------------------------------------

function open() {
  if (!currentCode || !currentNick) return
  onStatus?.(retry === 0 ? 'connecting' : 'reconnecting')

  const url = `${WS_BASE}/ws/rooms/${encodeURIComponent(currentCode)}`
  try {
    ws = new WebSocket(url)
  } catch {
    scheduleReconnect()
    return
  }

  ws.onopen = () => {
    retry = 0
    onStatus?.('connected')
    {
      const pid = getOrCreatePlayerId(currentCode!)
      ws?.send(JSON.stringify({ type: 'join', data: { nickname: currentNick, player_id: pid } }))
    }
  }

  ws.onmessage = (ev) => {
    try {
      const m = JSON.parse(ev.data as string) as ServerMessage
      // Cache the server's canonical player_id — it may differ from the one
      // we sent if the server regenerated it.
      if (m.type === 'joined' && currentCode && m.data?.player_id) {
        try { sessionStorage.setItem(playerIdKey(currentCode), m.data.player_id) } catch {}
      }
      onMsg?.(m)
    } catch { /* ignore malformed */ }
  }

  ws.onerror = () => { /* close handler will follow */ }

  ws.onclose = () => {
    ws = null
    if (manuallyClosed) return
    scheduleReconnect()
  }
}

function scheduleReconnect() {
  if (manuallyClosed) return
  retry++
  if (retry > MAX_RECONNECTS) {
    onStatus?.('error', S.connectionFailed)
    return
  }
  onStatus?.('reconnecting')
  const delay = Math.min(500 * 2 ** (retry - 1), 8000)
  retryTimer = window.setTimeout(() => {
    retryTimer = null
    open()
  }, delay)
}
