import { ShieldCheck } from 'lucide-react'

export function Logo({ compact = false }: { compact?: boolean }) {
  return (
    <div className="flex items-center gap-3">
      <span className="grid h-10 w-10 place-items-center rounded-xl border border-signal/30 bg-signal/10 text-signal shadow-glow">
        <ShieldCheck size={22} strokeWidth={1.8} />
      </span>
      {!compact && (
        <div>
          <div className="text-[15px] font-semibold tracking-[0.12em] text-white">QUANTUMFIELD</div>
          <div className="text-[9px] uppercase tracking-[0.24em] text-slate-500">TLS & PKI intelligence</div>
        </div>
      )}
    </div>
  )
}
