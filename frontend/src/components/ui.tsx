import type { LucideIcon } from 'lucide-react'
import type { ReactNode } from 'react'
import { scoreTone, titleCase } from '../lib/format'

export function PageHeader({ eyebrow, title, description, action }: {
  eyebrow?: string
  title: string
  description?: string
  action?: ReactNode
}) {
  return (
    <div className="mb-7 flex flex-col justify-between gap-4 md:flex-row md:items-end">
      <div>
        {eyebrow && <div className="mb-2 text-[10px] font-semibold uppercase tracking-[0.24em] text-signal">{eyebrow}</div>}
        <h1 className="text-2xl font-semibold tracking-tight text-white md:text-3xl">{title}</h1>
        {description && <p className="mt-2 max-w-2xl text-sm leading-6 text-slate-400">{description}</p>}
      </div>
      {action}
    </div>
  )
}

export function StatCard({ label, value, detail, icon: Icon, tone = 'signal' }: {
  label: string
  value: ReactNode
  detail: string
  icon: LucideIcon
  tone?: 'signal' | 'cyan' | 'amber' | 'rose'
}) {
  const tones = {
    signal: 'border-signal/20 bg-signal/10 text-signal',
    cyan: 'border-cyan-350/20 bg-cyan-350/10 text-cyan-350',
    amber: 'border-amber-400/20 bg-amber-400/10 text-amber-300',
    rose: 'border-rose-400/20 bg-rose-400/10 text-rose-300',
  }
  return (
    <div className="panel p-5">
      <div className="flex items-start justify-between">
        <div>
          <div className="text-xs font-medium uppercase tracking-[0.12em] text-slate-500">{label}</div>
          <div className="mt-3 text-3xl font-semibold tracking-tight text-white">{value}</div>
        </div>
        <div className={`grid h-10 w-10 place-items-center rounded-xl border ${tones[tone]}`}>
          <Icon size={18} />
        </div>
      </div>
      <div className="mt-3 text-xs text-slate-500">{detail}</div>
    </div>
  )
}

export function Badge({ value }: { value: string }) {
  const normalized = value.toLowerCase()
  const tone =
    normalized === 'critical' || normalized === 'failed' || normalized === 'expired'
      ? 'border-rose-400/25 bg-rose-400/10 text-rose-300'
      : normalized === 'high' || normalized === 'running'
        ? 'border-orange-400/25 bg-orange-400/10 text-orange-300'
        : normalized === 'medium' || normalized === 'queued' || normalized === 'pending'
          ? 'border-amber-400/25 bg-amber-400/10 text-amber-300'
          : normalized === 'completed' || normalized === 'valid' || normalized === 'a'
            ? 'border-signal/25 bg-signal/10 text-signal'
            : normalized === 'low' || normalized === 'info'
              ? 'border-cyan-350/25 bg-cyan-350/10 text-cyan-350'
              : 'border-slate-600 bg-slate-700/30 text-slate-300'
  return <span className={`inline-flex rounded-full border px-2.5 py-1 text-[10px] font-semibold uppercase tracking-wider ${tone}`}>{titleCase(value)}</span>
}

export function ScoreRing({ score, label, inverse = false, size = 104 }: {
  score: number
  label?: string
  inverse?: boolean
  size?: number
}) {
  const color = scoreTone(score, inverse)
  const displayPercent = inverse ? 100 - score : score
  return (
    <div className="relative grid shrink-0 place-items-center rounded-full" style={{
      width: size,
      height: size,
      background: `conic-gradient(${color} ${displayPercent * 3.6}deg, rgba(100,116,139,.16) 0deg)`,
    }}>
      <div className="absolute grid place-items-center rounded-full bg-ink-850" style={{ width: size - 12, height: size - 12 }}>
        <span className="text-2xl font-semibold text-white">{score}</span>
        {label && <span className="-mt-1 text-[9px] uppercase tracking-widest text-slate-500">{label}</span>}
      </div>
    </div>
  )
}

export function EmptyState({ icon: Icon, title, message, action }: {
  icon: LucideIcon
  title: string
  message: string
  action?: ReactNode
}) {
  return (
    <div className="panel grid min-h-64 place-items-center p-8 text-center">
      <div>
        <div className="mx-auto grid h-12 w-12 place-items-center rounded-2xl border border-slate-700 bg-ink-800 text-slate-400">
          <Icon size={21} />
        </div>
        <h3 className="mt-4 font-medium text-white">{title}</h3>
        <p className="mx-auto mt-2 max-w-sm text-sm leading-6 text-slate-500">{message}</p>
        {action && <div className="mt-5">{action}</div>}
      </div>
    </div>
  )
}

export function LoadingScreen() {
  return (
    <div className="grid min-h-[50vh] place-items-center">
      <div className="flex items-center gap-3 text-sm text-slate-400">
        <span className="h-2.5 w-2.5 animate-pulse rounded-full bg-signal shadow-glow" />
        Resolving cryptographic inventory…
      </div>
    </div>
  )
}

export function BooleanSignal({ value, positive = true, trueLabel = 'Yes', falseLabel = 'No' }: {
  value: boolean
  positive?: boolean
  trueLabel?: string
  falseLabel?: string
}) {
  const healthy = positive ? value : !value
  return (
    <span className={`inline-flex items-center gap-2 text-xs font-medium ${healthy ? 'text-signal' : 'text-rose-300'}`}>
      <span className={`h-1.5 w-1.5 rounded-full ${healthy ? 'bg-signal' : 'bg-rose-400'}`} />
      {value ? trueLabel : falseLabel}
    </span>
  )
}

