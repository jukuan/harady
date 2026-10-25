import { useEffect, useState } from 'react'
import { AdminError, fetchSummary, type StatsSummary } from '../../lib/adminApi'

export default function AdminStats() {
  const [data, setData] = useState<StatsSummary | null>(null)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    fetchSummary()
      .then(setData)
      .catch((e) => setErr(e instanceof AdminError ? e.message : 'load failed'))
  }, [])

  if (err) return <p className="text-danger text-sm">{err}</p>
  if (!data) return <p className="text-slate-400 text-sm">Загрузка…</p>

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
      <div className="card">
        <h3 className="text-xs font-extrabold uppercase tracking-wider text-slate-500 mb-2">
          Гарады
        </h3>
        <dl className="space-y-1 text-sm">
          <Row k="Усяго ў базе" v={data.cities.total_cities} />
          <Row k="Адпраўлена (усяго)" v={data.cities.total_submissions} />
          <Row k="Унікальных выкарыстана" v={data.cities.unique_cities_used} />
          <Row k="Ніколі не называлі" v={data.cities.never_used} />
          <Row k="Дададзена гульцамі" v={data.cities.learned_cities} />
        </dl>
      </div>
      <div className="card">
        <h3 className="text-xs font-extrabold uppercase tracking-wider text-slate-500 mb-2">
          Гульні
        </h3>
        <dl className="space-y-1 text-sm">
          <Row k="Згуляна" v={data.games.total_games} />
          <Row k="Хадоў усяго" v={data.games.total_moves} />
          <Row k="Найдаўжэйшы ланцуг" v={data.games.max_chain_length} />
          <Row k="…без ботаў" v={data.games.max_chain_human_only} />
          <Row k="Сярэдняя даўжыня" v={data.games.avg_chain_length.toFixed(1)} />
          <Row k="З ботамі" v={data.games.games_with_bots} />
        </dl>
      </div>
    </div>
  )
}

function Row({ k, v }: { k: string; v: string | number }) {
  return (
    <div className="flex justify-between border-b border-slate-100 pb-1">
      <dt className="text-slate-600">{k}</dt>
      <dd className="font-black tabular-nums">{v}</dd>
    </div>
  )
}
