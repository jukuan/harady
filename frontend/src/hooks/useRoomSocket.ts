import { useEffect } from 'react'
import { useStore } from '../store'
import * as sock from '../socket'
import type { ServerMessage } from '../types'

export function useRoomSocket(code: string | undefined, nickname: string) {
  useEffect(() => {
    if (!code || !nickname) return

    sock.setHandlers(
      (m: ServerMessage) => {
        const s = useStore.getState()
        switch (m.type) {
          case 'joined':         s.setJoined(m.data.player_id, m.data.is_host); break
          case 'room_state':     s.setRoom(m.data); break
          case 'turn_started':   s.clearRejection(); s.setTurn(m.data); break
          case 'chain_added':    s.clearRejection(); s.appendChain(m.data.entry, m.data.next_required_letter); break
          case 'chain_rejected': s.reject(m.data); break
          case 'player_passed':  s.playerPassed(m.data); break
          case 'player_out':     s.playerOut(m.data); break
          case 'game_ended':     s.gameEnded(s.room!, m.data.winner_id, m.data.winner_nickname); break
          case 'error':          s.setStatus('error', m.data.message); break
        }
      },
      (status, err) => {
        if (status === 'error') useStore.getState().setStatus('error', err ?? null)
        else useStore.getState().setStatus(status)
      },
    )

    sock.connect(code, nickname)

    return () => {
      sock.disconnect()
      useStore.getState().reset()
    }
  }, [code, nickname])

  return {
    submitCity: (city: string) => sock.send('submit_city', { city }),
    pass:       ()             => sock.send('pass'),
    addBot:     ()             => sock.send('add_bot'),
    startGame:  ()             => sock.send('start'),
    endGame:    ()             => sock.send('end_game'),
  }
}
