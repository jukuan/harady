import { describe, it, expect, beforeEach } from 'vitest'
import { useStore } from './store'
import type { ChainEntry, RoomState } from './types'

function room(overrides: Partial<RoomState> = {}): RoomState {
  return {
    code: 'abcd12',
    host_id: 'p1',
    phase: 'playing',
    players: [
      {
        id: 'p1',
        nickname: 'A',
        is_bot: false,
        is_host: true,
        missed: 0,
        out: false,
        online: true,
      },
      {
        id: 'p2',
        nickname: 'B',
        is_bot: false,
        is_host: false,
        missed: 0,
        out: false,
        online: true,
      },
    ],
    chain: [],
    required_letter: 'М',
    current_turn_id: 'p1',
    current_turn_nickname: 'A',
    min_players: 2,
    max_players: 10,
    max_misses: 2,
    ...overrides,
  }
}

function entry(city: string, playerID = 'p1'): ChainEntry {
  return { player_id: playerID, nickname: 'A', city, is_bot: false }
}

describe('store', () => {
  beforeEach(() => {
    useStore.getState().reset()
  })

  it('setRoom replaces the room snapshot', () => {
    useStore.getState().setRoom(room({ code: 'x1' }))
    expect(useStore.getState().room?.code).toBe('x1')
  })

  it('appendChain adds to the chain and updates required_letter', () => {
    useStore.getState().setRoom(room())
    useStore.getState().appendChain(entry('Мінск'), 'К')

    const s = useStore.getState()
    expect(s.room?.chain).toHaveLength(1)
    expect(s.room?.chain[0].city).toBe('Мінск')
    expect(s.room?.required_letter).toBe('К')
  })

  it('appendChain tolerates a null chain (defensive against Go null slices)', () => {
    useStore.getState().setRoom(room({ chain: null as unknown as ChainEntry[] }))
    useStore.getState().appendChain(entry('Мінск'), 'К')
    expect(useStore.getState().room?.chain).toHaveLength(1)
  })

  it('appendChain is a no-op when there is no room', () => {
    useStore.getState().appendChain(entry('X'), 'Y')
    expect(useStore.getState().room).toBeNull()
  })

  it('playerOut marks the player out with the missed count', () => {
    useStore.getState().setRoom(room())
    useStore.getState().playerOut({ player_id: 'p2', nickname: 'B', missed: 2 })

    const p = useStore.getState().room?.players.find((x) => x.id === 'p2')
    expect(p?.out).toBe(true)
    expect(p?.missed).toBe(2)
  })

  it('gameEnded freezes the final state and records the winner', () => {
    const r = room({ phase: 'ended' })
    useStore.getState().gameEnded(r, 'p1', 'A')

    const s = useStore.getState()
    expect(s.finalState).toBe(r)
    expect(s.winner).toEqual({ id: 'p1', nickname: 'A' })
  })

  it('reset clears everything', () => {
    useStore.getState().setRoom(room())
    useStore.getState().gameEnded(room(), 'p1', 'A')
    useStore.getState().reset()

    const s = useStore.getState()
    expect(s.room).toBeNull()
    expect(s.finalState).toBeNull()
    expect(s.winner).toBeNull()
    expect(s.playerId).toBeNull()
  })

  it('addReaction appends to the reactions list', () => {
    useStore.getState().addReaction({
      player_id: 'p1',
      nickname: 'A',
      emoji: '🔥',
      chain_index: 0,
      ts: Date.now(),
    })
    expect(useStore.getState().reactions).toHaveLength(1)
  })
})
