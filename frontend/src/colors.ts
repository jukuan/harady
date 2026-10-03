// Deterministic colour per player id. Same id → same colour, forever.
// Palette is Duolingo-flavoured, high-contrast against white/slate.

const PALETTE = [
  { bg: 'bg-rose-100',   border: 'border-rose-300',   text: 'text-rose-900',   solid: '#e11d48' },
  { bg: 'bg-amber-100',  border: 'border-amber-300',  text: 'text-amber-900',  solid: '#d97706' },
  { bg: 'bg-emerald-100',border: 'border-emerald-300',text: 'text-emerald-900',solid: '#059669' },
  { bg: 'bg-sky-100',    border: 'border-sky-300',    text: 'text-sky-900',    solid: '#0284c7' },
  { bg: 'bg-violet-100', border: 'border-violet-300', text: 'text-violet-900', solid: '#7c3aed' },
  { bg: 'bg-orange-100', border: 'border-orange-300', text: 'text-orange-900', solid: '#ea580c' },
  { bg: 'bg-teal-100',   border: 'border-teal-300',   text: 'text-teal-900',   solid: '#0d9488' },
  { bg: 'bg-fuchsia-100',border: 'border-fuchsia-300',text: 'text-fuchsia-900',solid: '#c026d3' },
] as const

export type Palette = typeof PALETTE[number]

export function colorFor(id: string): Palette {
  let h = 0
  for (let i = 0; i < id.length; i++) {
    h = (h * 31 + id.charCodeAt(i)) | 0
  }
  return PALETTE[Math.abs(h) % PALETTE.length]
}
