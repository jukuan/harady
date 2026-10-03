import { create } from 'zustand'
import type { ChainEntry, ChainRejected, PlayerOut, PlayerPassed, RoomState, TurnStarted } from './types'

export type ConnStatus = 'idle' | 'connecting' | 'connected' | 'reconnecting' | 'error'

interface State {
  status: ConnStatus
  errorMessage: string | null

  nickname: string
  playerId: string | null
  isHost: boolean

  room: RoomState | null
  turn: TurnStarted | null
  lastRejection: { reason: ChainRejected['reason']; text: string; at: number } | null
  lastOut: PlayerOut | null
  lastPassed: PlayerPassed | null
  finalState: RoomState | null  // frozen snapshot on game end
  winner: { id: string; nickname: string } | null

  setStatus: (s: ConnStatus, err?: string | null) => void
  setNickname: (n: string) => void
  setJoined: (id: string, isHost: boolean) => void
  setRoom: (r: RoomState) => void
  setTurn: (t: TurnStarted) => void
  appendChain: (e: ChainEntry, nextLetter: string) => void
  reject: (r: ChainRejected) => void
  playerOut: (o: PlayerOut) => void
  playerPassed: (p: PlayerPassed) => void
  gameEnded: (r: RoomState, winnerID?: string, winnerNick?: string) => void
  reset: () => void
}

export const useStore = create<State>((set) => ({
  status: 'idle',
  errorMessage: null,
  nickname: '',
  playerId: null,
  isHost: false,
  room: null,
  turn: null,
  lastRejection: null,
  lastOut: null,
  lastPassed: null,
  finalState: null,
  winner: null,

  setStatus: (s, err = null) => set({ status: s, errorMessage: err }),
  setNickname: (n) => set({ nickname: n }),
  setJoined: (id, isHost) => set({ playerId: id, isHost }),
  setRoom: (r) => set((s) => {
    // Don't clobber finalState after the game has ended.
    if (s.finalState && r.phase !== 'playing') return { room: r }
    return { room: r }
  }),
  setTurn: (t) => set({ turn: t }),
  appendChain: (e, nextLetter) => set((s) => {
    if (!s.room) return {}
    const chain = [...(s.room.chain ?? []), e]
    return { room: { ...s.room, chain, required_letter: nextLetter } }
  }),
  reject: (r) => set({ lastRejection: { reason: r.reason, text: r.text, at: Date.now() } }),
  playerOut: (o) => set((s) => {
    if (!s.room) return { lastOut: o }
    const players = s.room.players.map((p) =>
      p.id === o.player_id ? { ...p, out: true, missed: o.missed } : p
    )
    return { room: { ...s.room, players }, lastOut: o }
  }),
  gameEnded: (r, id, nick) => set({
    room: r,
    finalState: r,
    winner: id ? { id, nickname: nick ?? '' } : null,
  }),
  playerPassed: (p) => set({ lastPassed: p }),
  reset: () => set({
    status: 'idle', errorMessage: null,
    playerId: null, isHost: false,
    room: null, turn: null,
    lastRejection: null, lastOut: null, lastPassed: null,
    finalState: null, winner: null,
  }),
}))
