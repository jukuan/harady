import { useEffect, useState } from 'react'
import {
  AdminError,
  dumpTable,
  listTables,
  type TableDump,
  type TableInfo,
} from '../../lib/adminApi'

export default function AdminTables() {
  const [tables, setTables] = useState<TableInfo[]>([])
  const [active, setActive] = useState<string | null>(null)
  const [dump, setDump] = useState<TableDump | null>(null)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    listTables()
      .then((t) => {
        setTables(t)
        if (t.length && !active) setActive(t[0].name)
      })
      .catch((e) => setErr(e instanceof AdminError ? e.message : 'load failed'))
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  useEffect(() => {
    if (!active) return
    setDump(null)
    dumpTable(active, 100, 0)
      .then(setDump)
      .catch((e) => setErr(e instanceof AdminError ? e.message : 'dump failed'))
  }, [active])

  return (
    <div>
      <div className="flex flex-wrap gap-2 mb-3">
        {tables.map((t) => (
          <button
            key={t.name}
            className={
              'px-3 py-1.5 rounded-xl text-xs font-extrabold uppercase tracking-wider border-2 ' +
              (active === t.name
                ? 'bg-primary text-white border-primary'
                : 'bg-white text-slate-600 border-slate-200 hover:border-slate-400')
            }
            onClick={() => setActive(t.name)}
          >
            {t.name}
            <span className="ml-2 opacity-60">({t.rows})</span>
          </button>
        ))}
      </div>

      {err && <p className="text-danger text-sm mb-2">{err}</p>}

      {dump && (
        <>
          <details className="mb-3 text-xs">
            <summary className="cursor-pointer text-slate-500 hover:text-slate-700">
              Схема табліцы
            </summary>
            <pre className="bg-slate-100 rounded-xl p-2 mt-2 overflow-x-auto">{dump.schema}</pre>
          </details>
          <div className="overflow-x-auto card !p-0">
            <table className="w-full text-xs">
              <thead className="bg-slate-100">
                <tr>
                  {dump.columns.map((c) => (
                    <th key={c} className="text-left px-2 py-2 font-extrabold">
                      {c}
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {dump.rows.length === 0 && (
                  <tr>
                    <td colSpan={dump.columns.length} className="text-center py-6 text-slate-400">
                      (пуста)
                    </td>
                  </tr>
                )}
                {dump.rows.map((row, i) => (
                  <tr key={i} className="border-t border-slate-100 hover:bg-slate-50">
                    {row.map((cell, j) => (
                      <td key={j} className="px-2 py-1.5 font-mono whitespace-nowrap">
                        {cell === null ? (
                          <span className="text-slate-400 italic">null</span>
                        ) : (
                          String(cell)
                        )}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          <p className="text-xs text-slate-400 mt-2">
            Паказаныя першыя 100 радкоў. Поўная колькасць:{' '}
            {tables.find((t) => t.name === active)?.rows ?? '?'}
          </p>
        </>
      )}
    </div>
  )
}
