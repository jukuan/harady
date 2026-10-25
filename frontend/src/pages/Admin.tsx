import { useEffect, useState } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { AdminError, fetchMe, logout, requestMagicLink } from '../lib/adminApi'
import AdminCities from './admin/AdminCities'
import AdminTables from './admin/AdminTables'
import AdminStats from './admin/AdminStats'

type Status = 'loading' | 'logged_out' | 'logged_in'

export default function Admin() {
  const [status, setStatus] = useState<Status>('loading')
  const [email, setEmail] = useState<string | null>(null)
  const [params] = useSearchParams()

  useEffect(() => {
    fetchMe()
      .then((res) => {
        setEmail(res.email)
        setStatus('logged_in')
      })
      .catch(() => setStatus('logged_out'))
  }, [])

  const doLogout = async () => {
    await logout()
    setEmail(null)
    setStatus('logged_out')
  }

  if (status === 'loading') {
    return (
      <Centered>
        <p className="text-slate-500 font-extrabold uppercase tracking-wider animate-pulse">
          Загрузка…
        </p>
      </Centered>
    )
  }

  if (status === 'logged_out') {
    return (
      <AdminLogin initialError={params.get('error')} onSuccess={() => window.location.reload()} />
    )
  }

  return (
    <div className="min-h-full p-4 max-w-6xl mx-auto">
      <header className="flex items-center justify-between mb-4">
        <div>
          <h1 className="text-2xl font-black">Harady · Admin</h1>
          <p className="text-xs text-slate-500">{email}</p>
        </div>
        <div className="flex gap-2">
          <Link to="/" className="btn btn-ghost !py-2 text-xs">
            ← Да гульні
          </Link>
          <button className="btn btn-ghost !py-2 text-xs" onClick={doLogout}>
            Выйсці
          </button>
        </div>
      </header>
      <AdminTabs />
    </div>
  )
}

function AdminTabs() {
  const [tab, setTab] = useState<'cities' | 'tables' | 'stats'>('cities')
  return (
    <>
      <div className="flex gap-1 mb-4 border-b border-slate-200">
        {(
          [
            ['cities', 'Гарады'],
            ['tables', 'Табліцы'],
            ['stats', 'Статыстыка'],
          ] as const
        ).map(([id, label]) => (
          <button
            key={id}
            className={
              'px-4 py-2 text-sm font-extrabold uppercase tracking-wider ' +
              (tab === id
                ? 'text-primary-dark border-b-2 border-primary-dark -mb-px'
                : 'text-slate-500 hover:text-slate-700')
            }
            onClick={() => setTab(id)}
          >
            {label}
          </button>
        ))}
      </div>
      {tab === 'cities' && <AdminCities />}
      {tab === 'tables' && <AdminTables />}
      {tab === 'stats' && <AdminStats />}
    </>
  )
}

function AdminLogin({
  initialError,
  onSuccess,
}: {
  initialError: string | null
  onSuccess: () => void
}) {
  const [email, setEmail] = useState('')
  const [sent, setSent] = useState(false)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(initialError)

  const submit = async (e: React.FormEvent) => {
    e.preventDefault()
    setErr(null)
    setBusy(true)
    try {
      await requestMagicLink(email.trim())
      setSent(true)
    } catch (e) {
      setErr(e instanceof AdminError ? e.message : 'Нешта пайшло не так')
    } finally {
      setBusy(false)
    }
  }

  if (sent) {
    return (
      <Centered>
        <div className="card text-center">
          <div className="text-4xl mb-2">✉️</div>
          <h2 className="text-xl font-black mb-2">Праверце пошту</h2>
          <p className="text-sm text-slate-600">
            Мы даслалі спасылку для ўваходу на <strong>{email}</strong>. Яна дзейнічае 15 хвілін.
          </p>
          <p className="text-xs text-slate-400 mt-3">
            (Калі SMTP не наладжаны, спасылка запісана ў harady.log.)
          </p>
          <button
            className="btn btn-ghost w-full mt-4"
            onClick={() => {
              setSent(false)
              onSuccess()
            }}
          >
            Назад
          </button>
        </div>
      </Centered>
    )
  }

  return (
    <Centered>
      <div className="card">
        <h2 className="text-xl font-black mb-3">Уваход у адмінку</h2>
        <p className="text-sm text-slate-600 mb-4">
          Увядзіце email адміністратара. Мы дашлем спасылку для ўваходу.
        </p>
        <form onSubmit={submit} className="flex flex-col gap-3">
          <input
            className="input"
            type="email"
            required
            placeholder="admin@example.com"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            autoFocus
          />
          <button className="btn btn-primary w-full" disabled={busy || !email}>
            {busy ? 'Даслаем…' : 'Даслаць спасылку'}
          </button>
          {err && <p className="text-danger text-sm text-center">{err}</p>}
          <Link to="/" className="text-xs text-center text-slate-400 hover:text-slate-600">
            ← Да гульні
          </Link>
        </form>
      </div>
    </Centered>
  )
}

function Centered({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-full flex items-center justify-center p-5">
      <div className="w-full max-w-md">{children}</div>
    </div>
  )
}
