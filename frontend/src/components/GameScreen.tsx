import { useEffect, useMemo, useRef, useState } from 'react'
import { useStore } from '../store'
import type { RoomState } from '../types'
import { S } from '../strings'
import EmojiPicker from './EmojiPicker'

interface Props {
  room: RoomState
  onClue: (text: string) => void
  onGuess: (text: string) => void
  onEnd: () => void
  onExit: () => void
}

export default function GameScreen({ room, onClue, onGuess, onEnd, onExit }: Props) {
  const me = useStore((s) => s.playerId)
  const round = useStore((s) => s.round)
  const word = useStore((s) => s.myWord)
  const secondsLeft = useStore((s) => s.secondsLeft)
  const messages = useStore((s) => s.messages)
  const lastCorrect = useStore((s) => s.lastCorrect)
  const lastRoundEnded = useStore((s) => s.lastRoundEnded)
  const isHost = room.host_id === me
  const isActor = room.actor_id === me

  const [text, setText] = useState('')
  const scrollRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight })
  }, [messages.length])

  const duration = round?.duration ?? room.duration ?? 90
  const seconds = secondsLeft ?? duration
  const pct = Math.max(0, Math.min(100, (seconds / duration) * 100))
  const low = seconds <= 10

  const placeholder = isActor ? S.cluePlaceholder : S.guessPlaceholder
  const submitLabel = isActor ? S.send : S.guess

  const submit = () => {
    const t = text.trim()
    if (!t) return
    if (isActor) onClue(t)
    else onGuess(t)
    setText('')
  }

  const insertEmoji = (e: string) => setText((t) => t + e)

  const banner = useMemo(() => {
    if (!lastCorrect && !lastRoundEnded) return null
    if (lastCorrect) {
      return (
        <div className="card bg-success/10 border-success/30 animate-pop">
          <div className="text-success-dark font-black text-lg">
            {S.correct} — {lastCorrect.nickname}
          </div>
          <div className="text-slate-700">{S.wasCity(lastCorrect.city)}</div>
        </div>
      )
    }
    if (lastRoundEnded) {
      return (
        <div className="card bg-slate-100 border-slate-200 animate-pop">
          <div className="text-slate-700 font-black">{S.nobodyGuessed}</div>
          <div className="text-slate-600">{S.wasCity(lastRoundEnded.city)}</div>
        </div>
      )
    }
    return null
  }, [lastCorrect, lastRoundEnded])

  return (
    <div className="min-h-full flex flex-col max-w-md mx-auto w-full">
      {/* Top bar */}
      <header className="sticky top-0 z-30 bg-slate-50/90 backdrop-blur border-b border-slate-200">
        <div className="px-4 py-3 flex items-center justify-between">
          <div className="flex items-center gap-2 min-w-0">
            <span className="text-xs uppercase tracking-wider text-slate-500 font-extrabold">
              {S.round} {round?.round ?? room.round}
            </span>
            <span className="text-slate-300">·</span>
            <span className="text-xs uppercase tracking-wider text-slate-500 font-extrabold truncate">
              {S.actor}: {room.actor_nickname}
            </span>
          </div>
          <button className="btn btn-ghost !px-3 !py-1.5 text-xs" onClick={onExit}>
            {S.leaveRoom}
          </button>
        </div>

        <div className="px-4 pb-3 flex items-center gap-3">
          <div className="flex-1 h-3 rounded-full bg-slate-200 overflow-hidden">
            <div
              className={
                'h-full rounded-full transition-[width] duration-500 ease-linear ' +
                (low ? 'bg-danger' : 'bg-success')
              }
              style={{ width: `${pct}%` }}
            />
          </div>
          <div className={'font-black tabular-nums ' + (low ? 'text-danger' : 'text-slate-700')}>
            {seconds}{S.seconds}
          </div>
        </div>
      </header>

      {/* Word row */}
      <div className="px-4 pt-3">
        {isActor && word ? (
          <div className="card bg-primary/5 border-primary/30">
            <div className="text-xs uppercase tracking-wider text-primary-dark font-extrabold">
              {S.youAreActor}
            </div>
            <div className="text-2xl font-black tracking-wide">{word.city}</div>
            {word.region && (
              <div className="text-sm text-slate-500">{word.region}</div>
            )}
          </div>
        ) : (
          <div className="card bg-slate-100 border-slate-200">
            <div className="text-xs uppercase tracking-wider text-slate-500 font-extrabold">
              {S.youAreGuesser}
            </div>
            <div className="text-3xl font-black tracking-[0.4em] tabular-nums">
              {room.masked_city || '—'}
            </div>
          </div>
        )}
      </div>

      {/* Banner */}
      {banner && <div className="px-4 pt-3">{banner}</div>}

      {/* Chat stream */}
      <div ref={scrollRef} className="flex-1 overflow-y-auto px-4 py-3 space-y-2">
        {messages.map((m, i) => (
          <MessageBubble key={i} msg={m} isMe={m.player_id === me} />
        ))}
        {messages.length === 0 && (
          <div className="text-center text-slate-400 font-extrabold uppercase tracking-wider py-8">
            {isActor ? S.giveClue : S.waitingActor}
          </div>
        )}
      </div>

      {/* Input */}
      <div className="sticky bottom-0 bg-slate-50/95 backdrop-blur border-t border-slate-200 p-3"
           style={{ paddingBottom: 'max(0.75rem, env(safe-area-inset-bottom))' }}>
        <div className="flex items-end gap-2">
          <EmojiPicker onPick={insertEmoji} />
          <textarea
            className="input flex-1 resize-none !py-2.5"
            rows={1}
            value={text}
            maxLength={120}
            placeholder={placeholder}
            onChange={(e) => setText(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                submit()
              }
            }}
          />
          <button className="btn btn-primary !px-4" onClick={submit} disabled={!text.trim()}>
            {submitLabel}
          </button>
        </div>
        {isHost && (
          <button className="btn btn-ghost w-full mt-2 !py-2 text-xs" onClick={onEnd}>
            {S.endGame}
          </button>
        )}
      </div>
    </div>
  )
}

function MessageBubble({ msg, isMe }: { msg: import('../types').ChatMessage; isMe: boolean }) {
  const mine = isMe
  const isClue = msg.kind === 'clue'
  const bg = isClue ? 'bg-warn/10 border-warn/30' : mine ? 'bg-primary/10 border-primary/30' : 'bg-white border-slate-200'
  return (
    <div className={'flex ' + (mine ? 'justify-end' : 'justify-start')}>
      <div className={'max-w-[85%] rounded-2xl border-2 px-3 py-2 animate-pop ' + bg}>
        <div className="flex items-center gap-1.5 text-xs font-extrabold">
          <span className="text-slate-700">{msg.nickname}</span>
          {msg.is_bot && <span className="chip chip-bot">{S.bot}</span>}
          {isClue && <span className="text-warn-dark uppercase tracking-wider">clue</span>}
        </div>
        <div className="text-slate-900 whitespace-pre-wrap break-words">{msg.text}</div>
      </div>
    </div>
  )
}
