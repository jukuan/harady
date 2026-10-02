export type Phase = 'lobby' | 'round_active' | 'round_revealed' | 'ended'

export interface PlayerView {
  id: string
  nickname: string
  is_bot: boolean
  is_host: boolean
  score: number
  online: boolean
}

export interface RoomState {
  code: string
  host_id: string
  phase: Phase
  round: number
  actor_id: string
  actor_nickname: string
  masked_city: string
  players: PlayerView[]
  min_players: number
  max_players: number
  duration: number
}

export interface RoundStarted {
  actor_id: string
  actor_nickname: string
  masked_city: string
  round: number
  duration: number
}

export interface TickData {
  seconds_left: number
  round: number
  duration: number
}

export interface YourWord {
  city: string
  region?: string
}

export interface ChatMessage {
  player_id: string
  nickname: string
  text: string
  kind: 'clue' | 'guess'
  is_bot: boolean
  ts: number
}

export interface CorrectGuess {
  player_id: string
  nickname: string
  city: string
}

export interface RoundEnded {
  winner_id?: string
  city: string
  scores: PlayerView[]
}

export type ServerMessage =
  | { type: 'joined';           data: { player_id: string; room: string; is_host: boolean } }
  | { type: 'room_state';       data: RoomState }
  | { type: 'round_started';    data: RoundStarted }
  | { type: 'tick';             data: TickData }
  | { type: 'your_word';        data: YourWord }
  | { type: 'chat';             data: ChatMessage }
  | { type: 'correct_guess';    data: CorrectGuess }
  | { type: 'round_ended';      data: RoundEnded }
  | { type: 'game_ended';       data: RoomState }
  | { type: 'error';            data: { message: string } }
