import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'
import { STORAGE_NICKNAME } from '../config'
import { useRoomSocket } from '../hooks/useRoomSocket'
import { useStore } from '../store'
import { S } from '../strings'
import Lobby from '../components/Lobby'
import GameScreen from '../components/GameScreen'
import EndScreen from '../components/EndScreen'

export default function Room() {
  const { code } = useParams<{ code: string }>()
  const nav = useNavigate()
  const nickname = useStore((s) => s.nickname)
  const status   = useStore((s) => s.status)
  const errorMessage = useStore((s) => s.errorMessage)
  const room = useStore((s) => s.room)
  const finalState = useStore((s) => s.finalState)

  const [pendingName, setPendingName] = useState('')
  const [needName, setNeedName] = useState(false)

  useEffect(() => {
    const saved = localStorage.getItem(STORAGE_NICKNAME)
    if (saved && !nickname) useStore.getState().setNickname(saved)
    else if (!saved && !nickname) setNeedName(true)
  }, [nickname])

  const confirmName = () => {
    const n = pendingName.trim()
    if (!n) return
    localStorage.setItem(STORAGE_NICKNAME, n)
    useStore.getState().setNickname(n)
    setNeedName(false)
  }

  const api = useRoomSocket(code, nickname)

  if (needName) return (
    <Centered>
      <h2 className="text-xl font-black mb-3">{S.yourName}</h2>
      <input className="input mb-3" autoFocus value={pendingName} maxLength={20}
        placeholder={S.namePlaceholder}
        onChange={(e) => setPendingName(e.target.value)}
        onKeyDown={(e) => e.key === 'Enter' && confirmName()} />
      <button className="btn btn-primary w-full" onClick={confirmName}>{S.joinRoom}</button>
    </Centered>
  )

  if (status === 'error') return (
    <Centered>
      <div className="text-danger font-extrabold mb-4 text-center">
        {errorMessage ?? S.genericError}
      </div>
      <button className="btn btn-primary w-full" onClick={() => nav('/')}>
        {S.backHome}
      </button>
    </Centered>
  )

  if (!room) return (
    <Centered>
      <div className="animate-pulse text-slate-500 font-extrabold uppercase tracking-wider text-center">
        {status === 'reconnecting' ? S.reconnecting : S.connecting}
      </div>
      <button className="btn btn-ghost w-full mt-4" onClick={() => nav('/')}>
        {S.backHome}
      </button>
    </Centered>
  )

  return (
    <>
      {status === 'reconnecting' && (
        <div className="fixed top-0 inset-x-0 bg-warn text-white text-center text-xs font-extrabold
                        uppercase tracking-wider py-1.5 z-50">
          {S.reconnecting}
        </div>
      )}

      {finalState ? (
        <EndScreen onExit={() => nav('/')} />
      ) : room.phase === 'lobby' ? (
        <Lobby room={room} onAddBot={api.addBot} onStart={api.startGame} onExit={() => nav('/')} />
      ) : (
        <GameScreen room={room} onSubmitCity={api.submitCity} onPass={api.pass}
                    onEnd={api.endGame} onExit={() => nav('/')} />
      )}
    </>
  )
}

function Centered({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-full flex items-center justify-center p-5">
      <div className="w-full max-w-md card flex flex-col">{children}</div>
    </div>
  )
}
