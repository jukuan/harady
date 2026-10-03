export type Phase = 'lobby' | 'playing' | 'ended'

export interface PlayerView {
  id: string
  nickname: string
  is_bot: boolean
  is_host: boolean
  missed: number
  out: boolean
  online: boolean
}

export interface ChainEntry {
  player_id: string
  nickname: string
  city: string
  is_bot: boolean
}

export interface RoomState {
  code: string
  host_id: string
  phase: Phase
  players: PlayerView[]
  chain: ChainEntry[]
  required_letter: string
  current_turn_id: string
  current_turn_nickname: string
  min_players: number
  max_players: number
  max_misses: number
}

export interface TurnStarted {
  player_id: string
  nickname: string
  required_letter: string
  round: number
}

export interface ChainAdded {
  entry: ChainEntry
  next_required_letter: string
}

export interface ChainRejected {
  reason: 'wrong_letter' | 'already_used' | 'not_in_db' | 'empty'
  text: string
}

export interface PlayerOut {
  player_id: string
  nickname: string
  missed: number
}

export interface GameEnded {
  winner_id?: string
  winner_nickname?: string
  players: PlayerView[]
  chain: ChainEntry[]
}

export type ServerMessage =
  | { type: 'joined';         data: { player_id: string; room: string; is_host: boolean } }
  | { type: 'room_state';     data: RoomState }
  | { type: 'turn_started';   data: TurnStarted }
  | { type: 'chain_added';    data: ChainAdded }
  | { type: 'chain_rejected'; data: ChainRejected }
  | { type: 'player_out';     data: PlayerOut }
  | { type: 'game_ended';     data: GameEnded }
  | { type: 'error';          data: { message: string } }
