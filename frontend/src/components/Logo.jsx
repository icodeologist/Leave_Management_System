export default function Logo() {
  return (
    <div className="flex items-center gap-3">
      <span className="grid h-10 w-10 place-items-center rounded-lg bg-blue-600 text-white">
        <svg viewBox="0 0 24 24" className="h-5 w-5" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M7 3v3M17 3v3M4 9h16M5 5h14a1 1 0 0 1 1 1v13a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1Z" />
          <path d="m9 14 2 2 4-4" />
        </svg>
      </span>
      <div>
        <p className="font-semibold text-slate-900">Leave Management</p>
        <p className="text-xs text-slate-500">Employee leave portal</p>
      </div>
    </div>
  )
}
