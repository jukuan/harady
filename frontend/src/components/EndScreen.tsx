import { useStore } from '../store'
import { S } from '../strings'

interface Props {
  onExit: () => void
  onPlayAgain: () => void
  canPlayAgain: boolean
}

export default function EndScreen({ onExit, onPlayAgain, canPlayAgain }: Props) {
  const scores = useStore((s) => s.finalScores) ?? []
  const ranked = [...scores].sort((a, b) => b.score - a.score)
  const medals = ['🥇', '🥈', '🥉']

  return (
    <div className="min-h-full flex flex-col p-4 gap-4 max-w-md mx-auto w-full">
      <h1 className="text-3xl font-black text-center mt-4">{S.gameOver}</h1>

      <section className="card">
        <h2 className="text-xs uppercase tracking-wider text-slate-500 font-extrabold mb-3">
          {S.scoreboard}
        </h2>
        <ul className="flex flex-col gap-2">
          {ranked.map((p, i) => (
            <li key={p.id} className="flex items-center justify-between bg-slate-50 rounded-2xl px-3 py-3">
              <div className="flex items-center gap-3 min-w-0">
                <span className="text-2xl w-8 text-center">{medals[i] ?? i + 1}</span>
                <span className="text-2xl">{p.is_bot ? '🤖' : '🙂'}</span>
                <span className="font-extrabold truncate">{p.nickname}</span>
                {p.is_bot && <span className="chip chip-bot">{S.bot}</span>}
              </div>
              <span className="font-black tabular-nums text-lg">{p.score}</span>
            </li>
          ))}
        </ul>
      </section>

      <div className="mt-auto flex flex-col gap-3">
        {canPlayAgain && (
          <button className="btn btn-primary w-full" onClick={onPlayAgain}>
            {S.playAgain}
          </button>
        )}
        <button className="btn btn-ghost w-full" onClick={onExit}>
          {S.backHome}
        </button>
      </div>
    </div>
  )
}
