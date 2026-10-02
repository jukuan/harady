import { useEffect, useRef, useState } from 'react'

const EMOJI = [
  '😀','😄','😁','😂','🤣','😊','😍','🥳','😎','🤔',
  '😴','😱','🤯','🥺','😭','😡','👍','👎','👏','🙌',
  '🎉','🔥','💯','❤️','🎯','🏆','⭐️','🌟','🧠','🗺️',
]

export default function EmojiPicker({ onPick }: { onPick: (e: string) => void }) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!open) return
    const onDoc = (e: MouseEvent) => {
      if (ref.current && !ref.current.contains(e.target as Node)) setOpen(false)
    }
    document.addEventListener('mousedown', onDoc)
    return () => document.removeEventListener('mousedown', onDoc)
  }, [open])

  return (
    <div className="relative" ref={ref}>
      <button
        type="button"
        className="btn btn-ghost !px-3 !py-2.5 text-xl"
        onClick={() => setOpen((v) => !v)}
        aria-label="Emoji"
      >
        😊
      </button>

      {open && (
        <div className="absolute bottom-full left-0 mb-2 w-64 max-w-[85vw] bg-white border-2 border-slate-200
                        rounded-2xl p-2 grid grid-cols-8 gap-1 shadow-lg animate-pop z-40">
          {EMOJI.map((e) => (
            <button
              key={e}
              type="button"
              className="text-xl rounded-lg hover:bg-slate-100 active:bg-slate-200 p-1"
              onClick={() => { onPick(e); setOpen(false) }}
            >
              {e}
            </button>
          ))}
        </div>
      )}
    </div>
  )
}
