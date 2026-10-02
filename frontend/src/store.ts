import { create } from 'zustand'
import type {
  ChatMessage, CorrectGuess, PlayerView, RoomState, RoundEnded,
  RoundStarted, TickData, YourWord,
} from './types'

export type ConnStatus = 'idle' | 'connecting' | 'connected' | 'reconnecting' | 'error'

interface State {
  status: ConnStatus
  errorMessage: string | null

  nickname: string
  playerId: string | null
  isHost: boolean

  room: RoomState | null
  round: RoundStarted | null
  myWord: YourWord | null
  secondsLeft: number | null

  messages: ChatMessage[]
  lastCorrect: CorrectGuess | null
  lastRoundEnded: RoundEnded | null
  finalScores: PlayerView[] | null

  setStatus: (s: ConnStatus, err?: string | null) => void
  setNickname: (n: string) => void
  setJoined: (id: string, isHost: boolean) => void
  setRoom: (r: RoomState) => void
  startRound: (r: RoundStarted) => void
  setTick: (t: TickData) => void
  setWord: (w: YourWord) => void
  addChat: (m: ChatMessage) => void
  setCorrect: (c: CorrectGuess) => void
  setRoundEnded: (r: RoundEnded) => void
  setGameEnded: (scores: PlayerView[]) => void
  reset: () => void
}

export const useStore = create<State>((set) => ({
  status: 'idle',
  errorMessage: null,
  nickname: '',
  playerId: null,
  isHost: false,
  room: null,
  round: null,
  myWord: null,
  secondsLeft: null,
  messages: [],
  lastCorrect: null,
  lastRoundEnded: null,
  finalScores: null,

  setStatus: (s, err = null) => set({ status: s, errorMessage: err }),
  setNickname: (n) => set({ nickname: n }),
  setJoined: (id, isHost) => set({ playerId: id, isHost }),
  setRoom: (r) => set({ room: r }),
  startRound: (r) => set({
    round: r, secondsLeft: r.duration, myWord: null,
    lastCorrect: null, lastRoundEnded: null, messages: [],
  }),
  setTick: (t) => set({ secondsLeft: t.seconds_left }),
  setWord: (w) => set({ myWord: w }),
  addChat: (m) => set((s) => ({ messages: [...s.messages.slice(-200), m] })),
  setCorrect: (c) => set({ lastCorrect: c }),
  setRoundEnded: (r) => set({ lastRoundEnded: r }),
  setGameEnded: (scores) => set({ finalScores: scores }),
  reset: () => set({
    status: 'idle', errorMessage: null,
    playerId: null, isHost: false,
    room: null, round: null, myWord: null, secondsLeft: null,
    messages: [], lastCorrect: null, lastRoundEnded: null, finalScores: null,
  }),
}))
