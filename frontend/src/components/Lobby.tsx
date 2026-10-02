import { useState } from 'react'
import { useStore } from '../store'
import type { RoomState } from '../types'
import { S } from '../strings'

interface Props {
  room: RoomState
  onAddBot: () => void
  onStart: () => void
  onExit: () => void
}

export default function Lobby({ room, onAddBot, onStart, onExit }: Props) {
  const me = useStore((s) => s.playerId)
  const isHost = room.host_id === me
  const [copied, setCopied] = useState(false)

  const shareUrl = `${window.location.origin}/room/${room.code}`

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(shareUrl)
    } catch {
      const el = document.createElement('textarea')
      el.value = shareUrl
      document.body.appendChild(el)
      el.select()
      document.execCommand('copy')
      el.remove()
    }
    setCopied(true)
    window.setTimeout(() => setCopied(false), 1500)
  }

  const canStart = isHost && room.players.length >= room.min_players

  return (
    <div className="min-h-full flex flex-col p-4 gap-4 max-w-md mx-auto w-full">
      <header className="flex items-center justify-between">
        <div>
          <div className="text-xs uppercase tracking-wider text-slate-500 font-extrabold">
            {S.roomCode}
          </div>
          <div className="text-3xl font-black tracking-widest">{room.code}</div>
        </div>
        <button className="btn btn-ghost !px-3 !py-2 text-sm" onClick={onExit}>
          {S.leaveRoom}
        </button>
      </header>

      <button className="btn btn-ghost w-full" onClick={copy}>
        {copied ? S.copied : S.copyLink}
      </button>
      <p className="text-center text-slate-500 text-sm -mt-2">{S.shareHint}</p>

      <section className="card">
        <h2 className="text-xs uppercase tracking-wider text-slate-500 font-extrabold mb-3">
          {S.players} · {room.players.length}/{room.max_players}
        </h2>
        <ul className="flex flex-col gap-2">
          {room.players.map((p) => (
            <li key={p.id} className="flex items-center justify-between bg-slate-50 rounded-2xl px-3 py-2">
              <div className="flex items-center gap-2 min-w-0">
                <span className="text-2xl">{p.is_bot ? '🤖' : '🙂'}</span>
                <span className="font-extrabold truncate">{p.nickname}</span>
                {p.id === me && <span className="chip chip-you">{S.you}</span>}
                {p.is_host && <span className="chip chip-host">{S.host}</span>}
                {p.is_bot && <span className="chip chip-bot">{S.bot}</span>}
              </div>
              {!p.online && <span className="text-xs text-slate-400">offline</span>}
            </li>
          ))}
        </ul>
      </section>

      {isHost && (
        <div className="flex flex-col gap-3 mt-auto">
          <button
            className="btn btn-secondary w-full"
            onClick={onAddBot}
            disabled={room.players.length >= room.max_players}
          >
            + {S.addBot}
          </button>
          <button
            className="btn btn-primary w-full"
            onClick={onStart}
            disabled={!canStart}
          >
            {S.start}
          </button>
          {!canStart && room.players.length < room.min_players && (
            <p className="text-center text-slate-500 text-sm">
              {S.needPlayers(room.min_players)}
            </p>
          )}
        </div>
      )}

      {!isHost && (
        <div className="mt-auto text-center text-slate-500 font-extrabold uppercase tracking-wider animate-pulse">
          {S.waiting}
        </div>
      )}
    </div>
  )
}
