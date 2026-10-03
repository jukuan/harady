import { useEffect, useRef, useState } from 'react'
import { useStore } from '../store'
import type { RoomState } from '../types'
import { S } from '../strings'

interface Props {
  room: RoomState
  onSubmitCity: (city: string) => void
  onPass: () => void
  onEnd: () => void
  onExit: () => void
}

export default function GameScreen({ room, onSubmitCity, onPass, onEnd, onExit }: Props) {
  const me = useStore((s) => s.playerId)
  const rejection = useStore((s) => s.lastRejection)
  const lastOut = useStore((s) => s.lastOut)
  const lastPassed = useStore((s) => s.lastPassed)

  // Defensive: JSON from Go may deliver `null` where we expect an array.
  const players = room.players ?? []
  const chain = room.chain ?? []

  const isHost = room.host_id === me
  const isMyTurn = room.current_turn_id === me
  const myView = players.find((p) => p.id === me)
  const iAmOut = myView?.out ?? false

  const [text, setText] = useState('')
  const [rejToast, setRejToast] = useState<string | null>(null)
  const [outToast, setOutToast] = useState<string | null>(null)
  const [passToast, setPassToast] = useState<string | null>(null)
  const scrollRef = useRef<HTMLDivElement>(null)

  // Turn change clears the input.
  useEffect(() => { setText('') }, [room.current_turn_id])

  // Auto-scroll chain to the bottom on new entries.
  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight, behavior: 'smooth' })
  }, [chain.length])

  // Show rejection toast for 2s.
  useEffect(() => {
    if (!rejection) return
    const label =
      rejection.reason === 'wrong_letter' ? S.r_wrong_letter(room.required_letter) :
      rejection.reason === 'already_used' ? S.r_already_used :
      rejection.reason === 'not_in_db'    ? S.r_not_in_db :
      S.r_empty
    setRejToast(label)
    const t = window.setTimeout(() => setRejToast(null), 2000)
    return () => window.clearTimeout(t)
  }, [rejection?.at, rejection?.reason, room.required_letter])

  // Show "X is out" toast for 3s.
  useEffect(() => {
    if (!lastOut) return
    setOutToast(`${lastOut.nickname} ${S.out}`)
    const t = window.setTimeout(() => setOutToast(null), 3000)
    return () => window.clearTimeout(t)
  }, [lastOut?.player_id, lastOut?.missed])

  // Show "X passed" toast for 2.5s (only if they are not eliminated yet).
  useEffect(() => {
    if (!lastPassed) return
    if (lastPassed.missed >= room.max_misses) return // player_out handles it
    setPassToast(S.passedTurn(lastPassed.nickname, lastPassed.missed, room.max_misses))
    const t = window.setTimeout(() => setPassToast(null), 2500)
    return () => window.clearTimeout(t)
  }, [lastPassed?.player_id, lastPassed?.missed, room.max_misses])

  const canSubmit = isMyTurn && !iAmOut && text.trim().length > 0

  const submit = () => {
    if (!canSubmit) return
    onSubmitCity(text.trim())
    setText('')
  }

  const liveHint = computeLiveHint(text, room.required_letter)

  return (
    <div className="min-h-full flex flex-col max-w-md mx-auto w-full">
      {/* Top bar */}
      <header className="sticky top-0 z-30 bg-slate-50/95 backdrop-blur border-b border-slate-200">
        <div className="px-4 pt-3 pb-2 flex items-center justify-between">
          <div className="text-xs uppercase tracking-wider text-slate-500 font-extrabold truncate">
            {S.appName} · {room.code}
          </div>
          <button className="btn btn-ghost !px-3 !py-1.5 text-xs" onClick={onExit}>
            {S.leaveRoom}
          </button>
        </div>

        <div className="px-4 pb-3">
          <div className="card !py-3 !px-4 bg-primary/5 border-primary/30">
            <div className="text-xs uppercase tracking-wider text-primary-dark font-extrabold">
              {isMyTurn && !iAmOut ? S.yourTurn : S.turnOf(room.current_turn_nickname)}
            </div>
            <div className="text-2xl font-black tracking-wide mt-0.5">
              {room.required_letter
                ? <>{S.nextLetter} <span className="text-primary-dark">«{room.required_letter}»</span></>
                : S.firstMove}
            </div>
          </div>
        </div>
      </header>

      {/* Chain */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto px-4 py-3 space-y-2">
        {chain.length === 0 && (
          <div className="text-center text-slate-400 font-extrabold uppercase tracking-wider py-8">
            {S.firstMove}
          </div>
        )}
        {chain.map((e, i) => (
          <ChainRow key={i} index={i + 1} entry={e} isMine={e.player_id === me} />
        ))}
      </div>

      {/* Toasts */}
      <div className="pointer-events-none fixed inset-x-0 bottom-32 flex flex-col items-center gap-2 z-40 px-4">
        {rejToast && (
          <div className="bg-danger text-white rounded-2xl px-4 py-2 font-extrabold shadow-lg animate-pop">
            {rejToast}
          </div>
        )}
        {passToast && (
          <div className="bg-slate-700 text-white rounded-2xl px-4 py-2 font-extrabold shadow-lg animate-pop">
            {passToast}
          </div>
        )}
        {outToast && (
          <div className="bg-warn text-white rounded-2xl px-4 py-2 font-extrabold shadow-lg animate-pop">
            {outToast}
          </div>
        )}
      </div>

      {/* Input area */}
      <div className="sticky bottom-0 bg-slate-50/95 backdrop-blur border-t border-slate-200 p-3"
           style={{ paddingBottom: 'max(0.75rem, env(safe-area-inset-bottom))' }}>
        {iAmOut ? (
          <div className="text-center text-slate-500 font-extrabold uppercase tracking-wider py-2">
            {S.youAreOut}
          </div>
        ) : (
          <>
            <div className="flex items-stretch gap-2">
              <input
                className={'input flex-1 ' + (liveHint ? '!border-danger' : '')}
                value={text}
                maxLength={80}
                placeholder={S.placeholder}
                onChange={(e) => setText(e.target.value)}
                onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); submit() } }}
                disabled={!isMyTurn}
                autoComplete="off"
                autoCapitalize="words"
              />
              <button className="btn btn-primary !px-4" onClick={submit} disabled={!canSubmit}>
                {S.send}
              </button>
            </div>
            {liveHint && (
              <div className="text-danger text-xs font-extrabold mt-1">{liveHint}</div>
            )}
            <button className="btn btn-ghost w-full mt-2 !py-2 text-xs"
                    onClick={onPass} disabled={!isMyTurn || iAmOut}>
              {S.pass} ({myView?.missed ?? 0}/{room.max_misses})
            </button>
            {isHost && (
              <button className="btn btn-danger w-full mt-2 !py-2 text-xs"
                      onClick={() => { if (confirm(S.endGameConfirm)) onEnd() }}>
                {S.endGame}
              </button>
            )}
          </>
        )}
      </div>
    </div>
  )
}

function ChainRow({ index, entry, isMine }: { index: number; entry: import('../types').ChainEntry; isMine: boolean }) {
  return (
    <div className={'flex ' + (isMine ? 'justify-end' : 'justify-start')}>
      <div className={
        'max-w-[85%] rounded-2xl border-2 px-3 py-2 animate-pop ' +
        (isMine ? 'bg-primary/10 border-primary/30' : 'bg-white border-slate-200')
      }>
        <div className="flex items-center gap-1.5 text-xs font-extrabold text-slate-500">
          <span className="text-slate-400">{index}.</span>
          <span className="text-slate-700">{entry.nickname}</span>
          {entry.is_bot && <span className="chip chip-bot">бот</span>}
        </div>
        <div className="text-lg font-black text-slate-900">{entry.city}</div>
      </div>
    </div>
  )
}

function computeLiveHint(text: string, requiredLetter: string): string | null {
  const t = text.trim()
  if (!t || !requiredLetter) return null
  const first = Array.from(t)[0]
  if (!first) return null
  if (first.toUpperCase() !== requiredLetter.toUpperCase()) {
    return `Патрэбна літара «${requiredLetter}»`
  }
  return null
}
