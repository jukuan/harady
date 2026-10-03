import { useStore } from '../store'
import { S } from '../strings'

interface Props { onExit: () => void }

import { useState } from 'react'

export default function EndScreen({ onExit }: Props) {
  const [copied, setCopied] = useState(false)
  const finalState = useStore((s) => s.finalState)
  const winner = useStore((s) => s.winner)

  if (!finalState) return null

  const copyResult = async () => {
    const lines = [
      `Harady · ${chain.length} гарадоў`,
      chain.map((e) => e.city).join(' → '),
      winner ? `🏆 ${winner.nickname}` : '🤝 Нічыя',
    ]
    const text = lines.join('\n')
    try { await navigator.clipboard.writeText(text) } catch { /* ignore */ }
    setCopied(true)
    window.setTimeout(() => setCopied(false), 1500)
  }


  const chain = finalState.chain ?? []
  const players = finalState.players ?? []
  const named = new Map<string, number>()
  for (const e of chain) named.set(e.player_id, (named.get(e.player_id) ?? 0) + 1)

  const ranked = [...players].sort((a, b) => {
    if (a.id === winner?.id) return -1
    if (b.id === winner?.id) return 1
    if (a.out !== b.out) return a.out ? 1 : -1
    return (named.get(b.id) ?? 0) - (named.get(a.id) ?? 0)
  })

  return (
    <div className="min-h-full flex flex-col p-4 gap-4 max-w-md mx-auto w-full">
      <div className="text-center mt-4">
        <div className="text-6xl">🏆</div>
        <h1 className="text-3xl font-black mt-2">{S.gameOver}</h1>
        {winner ? (
          <p className="text-lg font-extrabold text-primary-dark mt-1">
            {S.winner}: {winner.nickname}
          </p>
        ) : (
          <p className="text-lg font-extrabold text-slate-500 mt-1">
            {S.noWinner} · {S.noWinnerHint}
          </p>
        )}
        <p className="text-sm text-slate-500 mt-1">
          {S.chainLength}: <span className="font-black text-slate-700">{chain.length}</span>
        </p>
      </div>

      <section className="card">
        <h2 className="text-xs uppercase tracking-wider text-slate-500 font-extrabold mb-3">
          {S.scoreboard}
        </h2>
        <ul className="flex flex-col gap-2">
          {ranked.map((p) => (
            <li key={p.id} className="flex items-center justify-between bg-slate-50 rounded-2xl px-3 py-3">
              <div className="flex items-center gap-2 min-w-0">
                <span className="text-2xl">{p.is_bot ? '🤖' : '🙂'}</span>
                <span className="font-extrabold truncate">{p.nickname}</span>
                {p.is_bot && <span className="chip chip-bot">{S.bot}</span>}
                {p.out && <span className="chip" style={{ background: '#fdecea', color: '#b71c1c' }}>{S.out}</span>}
              </div>
              <span className="font-black tabular-nums text-lg">{named.get(p.id) ?? 0}</span>
            </li>
          ))}
        </ul>
      </section>

      <div className="mt-auto flex flex-col gap-3">
        <button className="btn btn-secondary w-full" onClick={copyResult}>
          {copied ? S.copied : S.shareResult}
        </button>
        <button className="btn btn-primary w-full" onClick={onExit}>{S.backHome}</button>
      </div>
    </div>
  )
}
