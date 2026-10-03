import { colorFor } from '../colors'
import type { ChainEntry } from '../types'

interface Props {
  index: number
  entry: ChainEntry
  isMine: boolean
}

export default function ChainRow({ index, entry, isMine }: Props) {
  const c = colorFor(entry.player_id)
  return (
    <div className={'flex ' + (isMine ? 'justify-end' : 'justify-start')}>
      <div className={
        'max-w-[85%] rounded-2xl border-2 px-3 py-2 animate-pop ' +
        c.bg + ' ' + c.border
      }>
        <div className={'flex items-center gap-1.5 text-xs font-extrabold ' + c.text}>
          <span className="opacity-60">{index}.</span>
          <span>{entry.nickname}</span>
          {entry.is_bot && <span className="chip chip-bot">бот</span>}
        </div>
        <div className={'text-lg font-black ' + c.text}>{entry.city}</div>
      </div>
    </div>
  )
}
