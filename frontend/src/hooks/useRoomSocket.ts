import { useEffect } from 'react'
import { useStore } from '../store'
import * as sock from '../socket'
import type { ServerMessage } from '../types'

export function useRoomSocket(code: string | undefined, nickname: string) {
  useEffect(() => {
    if (!code || !nickname) return

    sock.setHandlers(
      (m: ServerMessage) => {
        switch (m.type) {
          case 'joined':
            useStore.getState().setJoined(m.data.player_id, m.data.is_host)
            break
          case 'room_state':
            useStore.getState().setRoom(m.data)
            break
          case 'round_started':
            useStore.getState().startRound(m.data)
            break
          case 'tick':
            useStore.getState().setTick(m.data)
            break
          case 'your_word':
            useStore.getState().setWord(m.data)
            break
          case 'chat':
            useStore.getState().addChat(m.data)
            break
          case 'correct_guess':
            useStore.getState().setCorrect(m.data)
            break
          case 'round_ended':
            useStore.getState().setRoundEnded(m.data)
            break
          case 'game_ended':
            useStore.getState().setGameEnded(m.data.players)
            break
          case 'error':
            useStore.getState().setStatus('error', m.data.message)
            break
        }
      },
      (s, err) => {
        if (s === 'error') {
          useStore.getState().setStatus('error', err ?? null)
        } else {
          useStore.getState().setStatus(s)
        }
      },
    )

    sock.connect(code, nickname)

    return () => {
      sock.disconnect()
      useStore.getState().reset()
    }
  }, [code, nickname])

  return {
    sendClue:  (text: string) => sock.send('clue',  { text }),
    sendGuess: (text: string) => sock.send('guess', { text }),
    addBot:    ()             => sock.send('add_bot'),
    startGame: ()             => sock.send('start'),
    endGame:   ()             => sock.send('end_game'),
  }
}
