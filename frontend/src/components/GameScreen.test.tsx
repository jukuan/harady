import { describe, it, expect, beforeEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import GameScreen from './GameScreen'
import { useStore } from '../store'
import type { RoomState } from '../types'

function room(overrides: Partial<RoomState> = {}): RoomState {
  return {
    code: 'abcd12',
    host_id: 'me',
    phase: 'playing',
    players: [
      {
        id: 'me',
        nickname: 'Me',
        is_bot: false,
        is_host: true,
        missed: 0,
        out: false,
        online: true,
      },
      {
        id: 'bot1',
        nickname: 'Bot',
        is_bot: true,
        is_host: false,
        missed: 0,
        out: false,
        online: true,
      },
    ],
    chain: [{ player_id: 'me', nickname: 'Me', city: 'Мінск', is_bot: false }],
    required_letter: 'К',
    current_turn_id: 'me',
    current_turn_nickname: 'Me',
    min_players: 2,
    max_players: 10,
    max_misses: 2,
    ...overrides,
  }
}

const noop = () => {}

describe('GameScreen', () => {
  beforeEach(() => {
    useStore.getState().reset()
    useStore.getState().setJoined('me', true)
  })

  it('enables the input and shows the guess placeholder when it is my turn', () => {
    render(
      <GameScreen
        room={room({ current_turn_id: 'me' })}
        onSubmitCity={noop}
        onPass={noop}
        onReact={noop}
        onEnd={noop}
        onExit={noop}
      />,
    )
    const input = screen.getByPlaceholderText(/Назва горада/i) as HTMLInputElement
    expect(input.disabled).toBe(false)
  })

  it('disables the input when it is not my turn', () => {
    render(
      <GameScreen
        room={room({ current_turn_id: 'bot1', current_turn_nickname: 'Bot' })}
        onSubmitCity={noop}
        onPass={noop}
        onReact={noop}
        onEnd={noop}
        onExit={noop}
      />,
    )
    const input = screen.getByPlaceholderText(/Назва горада/i) as HTMLInputElement
    expect(input.disabled).toBe(true)
  })

  it('renders the required letter in the header', () => {
    render(
      <GameScreen
        room={room({ required_letter: 'Л' })}
        onSubmitCity={noop}
        onPass={noop}
        onReact={noop}
        onEnd={noop}
        onExit={noop}
      />,
    )
    // The letter appears inside «Л» in the header card
    expect(screen.getByText(/«Л»/)).toBeInTheDocument()
  })

  it('renders the reaction bar only when the chain is non-empty', () => {
    const { rerender } = render(
      <GameScreen
        room={room({ chain: [] })}
        onSubmitCity={noop}
        onPass={noop}
        onReact={noop}
        onEnd={noop}
        onExit={noop}
      />,
    )
    expect(screen.queryByLabelText('Рэакцыя 👍')).not.toBeInTheDocument()

    rerender(
      <GameScreen
        room={room()}
        onSubmitCity={noop}
        onPass={noop}
        onReact={noop}
        onEnd={noop}
        onExit={noop}
      />,
    )
    expect(screen.getByLabelText('Рэакцыя 👍')).toBeInTheDocument()
  })

  it('shows the spectator message when the player is out', () => {
    render(
      <GameScreen
        room={room({
          players: [
            {
              id: 'me',
              nickname: 'Me',
              is_bot: false,
              is_host: true,
              missed: 2,
              out: true,
              online: true,
            },
            {
              id: 'bot1',
              nickname: 'Bot',
              is_bot: true,
              is_host: false,
              missed: 0,
              out: false,
              online: true,
            },
          ],
        })}
        onSubmitCity={noop}
        onPass={noop}
        onReact={noop}
        onEnd={noop}
        onExit={noop}
      />,
    )
    expect(screen.getByText(/Ты выбыў/i)).toBeInTheDocument()
  })
})
