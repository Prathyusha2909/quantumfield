import { Atom, Binary, KeyRound, LockKeyhole } from 'lucide-react'
import type { ReactNode } from 'react'
import { Logo } from './Logo'

export function AuthLayout({ children, title, subtitle }: {
  children: ReactNode
  title: string
  subtitle: string
}) {
  return (
    <main className="grid min-h-screen bg-ink-950 lg:grid-cols-[1.05fr_.95fr]">
      <section className="relative hidden overflow-hidden border-r border-white/[0.06] lg:flex lg:flex-col lg:justify-between lg:p-12">
        <div className="absolute -left-32 top-1/3 h-96 w-96 rounded-full bg-signal/5 blur-3xl" />
        <div className="absolute right-0 top-0 h-80 w-80 rounded-full bg-cyan-350/5 blur-3xl" />
        <Logo />
        <div className="relative max-w-xl">
          <div className="mb-7 flex items-center gap-3 text-[10px] font-semibold uppercase tracking-[0.24em] text-signal">
            <span className="h-px w-10 bg-signal/40" />
            Cryptographic attack surface
          </div>
          <h1 className="text-5xl font-medium leading-[1.12] tracking-[-0.04em] text-white">
            See the risk hiding in every TLS handshake.
          </h1>
          <p className="mt-6 max-w-lg text-base leading-7 text-slate-500">
            Continuous certificate intelligence, transport security findings, and a practical path from classical PKI to post-quantum readiness.
          </p>
          <div className="mt-10 grid grid-cols-3 gap-3">
            {[
              { icon: KeyRound, text: 'X.509 inventory' },
              { icon: Binary, text: 'PQC dependency' },
              { icon: LockKeyhole, text: 'TLS posture' },
            ].map(({ icon: Icon, text }) => (
              <div key={text} className="rounded-xl border border-white/[0.06] bg-white/[0.025] p-4">
                <Icon size={17} className="text-cyan-350" />
                <div className="mt-3 text-xs text-slate-400">{text}</div>
              </div>
            ))}
          </div>
        </div>
        <div className="flex items-center gap-2 text-[10px] uppercase tracking-[0.18em] text-slate-700">
          <Atom size={14} /> Built for crypto-agile security teams
        </div>
      </section>
      <section className="flex items-center justify-center px-5 py-12">
        <div className="w-full max-w-md">
          <div className="mb-10 lg:hidden"><Logo /></div>
          <div className="mb-8">
            <h2 className="text-3xl font-semibold tracking-tight text-white">{title}</h2>
            <p className="mt-2 text-sm leading-6 text-slate-500">{subtitle}</p>
          </div>
          {children}
        </div>
      </section>
    </main>
  )
}

