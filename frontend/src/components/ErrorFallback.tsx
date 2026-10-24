interface Props {
  error: unknown
  resetError: () => void
}

export default function ErrorFallback({ error, resetError }: Props) {
  const msg = error instanceof Error ? error.message : String(error)
  return (
    <div className="min-h-full flex items-center justify-center p-5">
      <div className="w-full max-w-md card flex flex-col gap-3 text-center">
        <div className="text-5xl">😵</div>
        <h1 className="text-2xl font-black">Нешта пайшло не так</h1>
        <p className="text-sm text-slate-500 break-words">{msg}</p>
        <button className="btn btn-primary w-full" onClick={resetError}>
          Паспрабаваць зноў
        </button>
        <button
          className="btn btn-ghost w-full"
          onClick={() => {
            window.location.href = '/'
          }}
        >
          На галоўную
        </button>
      </div>
    </div>
  )
}
