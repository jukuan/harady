import { FormEvent, useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { API_BASE, STORAGE_NICKNAME } from '../config'
import { useStore } from '../store'
import { S } from '../strings'

export default function Landing() {
  const nav = useNavigate()
  const setNickname = useStore((s) => s.setNickname)
  const [name, setName] = useState('')
  const [code, setCode] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr]   = useState<string | null>(null)

  useEffect(() => {
    const saved = localStorage.getItem(STORAGE_NICKNAME)
    if (saved) setName(saved)
  }, [])

  const remember = (n: string) => {
    localStorage.setItem(STORAGE_NICKNAME, n)
    setNickname(n)
  }

  async function createRoom(e: FormEvent) {
    e.preventDefault()
    setErr(null)
    const n = name.trim()
    if (!n) { setErr(S.nameRequired); return }
    remember(n)
    setBusy(true)
    try {
      const res = await fetch(`${API_BASE}/api/rooms`, { method: 'POST' })
      if (!res.ok) throw new Error('create failed')
      const body = (await res.json()) as { code: string }
      nav(`/room/${body.code}`)
    } catch {
      setErr(S.genericError)
    } finally {
      setBusy(false)
    }
  }

  function joinRoom(e: FormEvent) {
    e.preventDefault()
    setErr(null)
    const n = name.trim()
    const c = code.trim().toUpperCase()
    if (!n) { setErr(S.nameRequired); return }
    if (!/^[A-Z0-9]{4,12}$/.test(c)) { setErr(S.invalidCode); return }
    remember(n)
    nav(`/room/${c}`)
  }

  return (
    <div className="min-h-full flex flex-col items-center justify-center p-5">
      <div className="w-full max-w-md flex flex-col gap-5 animate-pop">
        <header className="text-center">
          <div className="mx-auto w-20 h-20 rounded-3xl bg-primary text-white grid place-items-center
                          text-4xl font-black shadow-[0_6px_0_0_theme(colors.primary.dark)]">
            H
          </div>
          <h1 className="mt-4 text-4xl font-black tracking-tight">{S.appName}</h1>
          <p className="text-slate-500 mt-1">{S.tagline}</p>
        </header>

        <div className="card flex flex-col gap-4">
          <label className="flex flex-col gap-2">
            <span className="text-xs font-extrabold uppercase tracking-wider text-slate-500">
              {S.jukuan}
            </span>
            <input
              className="input"
              value={name}
              maxLength={20}
              placeholder={S.namePlaceholder}
              onChange={(e) => setName(e.target.value)}
              autoComplete="nickname"
            />
          </label>

          <form onSubmit={createRoom}>
            <button className="btn btn-primary w-full" disabled={busy}>
              {busy ? S.connecting : S.createRoom}
            </button>
          </form>
        </div>

        <div className="flex items-center gap-3 text-slate-400 text-xs font-extrabold uppercase">
          <div className="h-px flex-1 bg-slate-200" />
          {S.or}
          <div className="h-px flex-1 bg-slate-200" />
        </div>

        <form onSubmit={joinRoom} className="card flex flex-col gap-3">
          <label className="flex flex-col gap-2">
            <span className="text-xs font-extrabold uppercase tracking-wider text-slate-500">
              {S.roomCode}
            </span>
            <input
              className="input tracking-widest text-center text-lg font-black"
              value={code}
              maxLength={12}
              placeholder={S.roomCodePlaceholder}
              onChange={(e) => setCode(e.target.value.toUpperCase())}
              autoCapitalize="characters"
              autoComplete="off"
            />
          </label>
          <button className="btn btn-ghost w-full" type="submit">
            {S.joinRoom}
          </button>
        </form>

        {err && (
          <div className="text-danger text-center text-sm font-bold animate-shake" role="alert">
            {err}
          </div>
        )}
      </div>
    </div>
  )
}
