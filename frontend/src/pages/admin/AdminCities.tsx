import { useEffect, useState } from 'react'
import {
  AdminError,
  createCity,
  deleteCity,
  listCities,
  updateCity,
  type CityRow,
} from '../../lib/adminApi'

export default function AdminCities() {
  const [rows, setRows] = useState<CityRow[]>([])
  const [loading, setLoading] = useState(false)
  const [q, setQ] = useState('')
  const [err, setErr] = useState<string | null>(null)

  const [adding, setAdding] = useState(false)
  const [newName, setNewName] = useState('')
  const [newRegion, setNewRegion] = useState('')

  const [editing, setEditing] = useState<number | null>(null)
  const [editName, setEditName] = useState('')
  const [editRegion, setEditRegion] = useState('')

  const load = async () => {
    setLoading(true)
    setErr(null)
    try {
      setRows(await listCities(q))
    } catch (e) {
      setErr(e instanceof AdminError ? e.message : 'load failed')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    const t = setTimeout(load, q ? 250 : 0)
    return () => clearTimeout(t)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [q])

  const onAdd = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newName.trim()) return
    try {
      await createCity(newName.trim(), newRegion.trim())
      setAdding(false)
      setNewName('')
      setNewRegion('')
      load()
    } catch (e) {
      setErr(e instanceof AdminError ? e.message : 'create failed')
    }
  }

  const onSave = async (id: number) => {
    try {
      await updateCity(id, editName.trim(), editRegion.trim())
      setEditing(null)
      load()
    } catch (e) {
      setErr(e instanceof AdminError ? e.message : 'update failed')
    }
  }

  const onDelete = async (row: CityRow) => {
    if (!confirm(`Выдаліць «${row.name}»?`)) return
    try {
      await deleteCity(row.id)
      load()
    } catch (e) {
      setErr(e instanceof AdminError ? e.message : 'delete failed')
    }
  }

  return (
    <div>
      <div className="flex items-center gap-2 mb-3">
        <input
          className="input flex-1"
          placeholder="Пошук гарадоў…"
          value={q}
          onChange={(e) => setQ(e.target.value)}
        />
        <button className="btn btn-primary !py-2" onClick={() => setAdding(true)}>
          + Дадаць
        </button>
      </div>

      {err && <p className="text-danger text-sm mb-2">{err}</p>}

      {adding && (
        <form className="card mb-3 flex flex-col sm:flex-row gap-2 items-stretch" onSubmit={onAdd}>
          <input
            className="input flex-1"
            placeholder="Назва"
            value={newName}
            onChange={(e) => setNewName(e.target.value)}
            autoFocus
          />
          <input
            className="input flex-1"
            placeholder="Рэгіён"
            value={newRegion}
            onChange={(e) => setNewRegion(e.target.value)}
          />
          <div className="flex gap-2">
            <button className="btn btn-primary !py-2" type="submit">
              Захаваць
            </button>
            <button className="btn btn-ghost !py-2" type="button" onClick={() => setAdding(false)}>
              Адмена
            </button>
          </div>
        </form>
      )}

      <div className="overflow-x-auto card !p-0">
        <table className="w-full text-sm">
          <thead className="bg-slate-100">
            <tr>
              <th className="text-left px-3 py-2 w-16">ID</th>
              <th className="text-left px-3 py-2">Назва</th>
              <th className="text-left px-3 py-2">Рэгіён</th>
              <th className="text-left px-3 py-2 w-24">Крыніца</th>
              <th className="text-right px-3 py-2 w-32"></th>
            </tr>
          </thead>
          <tbody>
            {loading && rows.length === 0 && (
              <tr>
                <td colSpan={5} className="text-center py-6 text-slate-400">
                  Загрузка…
                </td>
              </tr>
            )}
            {!loading && rows.length === 0 && (
              <tr>
                <td colSpan={5} className="text-center py-6 text-slate-400">
                  (пуста)
                </td>
              </tr>
            )}
            {rows.map((row) => (
              <tr key={row.id} className="border-t border-slate-100 hover:bg-slate-50">
                <td className="px-3 py-2 text-slate-400 tabular-nums">{row.id}</td>
                <td className="px-3 py-2 font-extrabold">
                  {editing === row.id ? (
                    <input
                      className="input !py-1"
                      value={editName}
                      onChange={(e) => setEditName(e.target.value)}
                    />
                  ) : (
                    row.name
                  )}
                </td>
                <td className="px-3 py-2">
                  {editing === row.id ? (
                    <input
                      className="input !py-1"
                      value={editRegion}
                      onChange={(e) => setEditRegion(e.target.value)}
                    />
                  ) : (
                    row.region
                  )}
                </td>
                <td className="px-3 py-2">
                  <span className={'chip ' + (row.source === 'learned' ? 'chip-bot' : 'chip-you')}>
                    {row.source ?? 'seed'}
                  </span>
                </td>
                <td className="px-3 py-2 text-right">
                  {editing === row.id ? (
                    <div className="flex gap-1 justify-end">
                      <button
                        className="btn btn-success !py-1 !px-2 text-xs"
                        onClick={() => onSave(row.id)}
                      >
                        ✓
                      </button>
                      <button
                        className="btn btn-ghost !py-1 !px-2 text-xs"
                        onClick={() => setEditing(null)}
                      >
                        ✕
                      </button>
                    </div>
                  ) : (
                    <div className="flex gap-1 justify-end">
                      <button
                        className="btn btn-ghost !py-1 !px-2 text-xs"
                        onClick={() => {
                          setEditing(row.id)
                          setEditName(row.name)
                          setEditRegion(row.region)
                        }}
                      >
                        ✎
                      </button>
                      <button
                        className="btn btn-danger !py-1 !px-2 text-xs"
                        onClick={() => onDelete(row)}
                      >
                        🗑
                      </button>
                    </div>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
