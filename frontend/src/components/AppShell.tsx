import {
  BarChart3,
  Boxes,
  ClipboardList,
  FileBarChart,
  Fingerprint,
  Gauge,
  LogOut,
  Menu,
  Radar,
  ShieldAlert,
  X,
} from 'lucide-react'
import { useState } from 'react'
import { NavLink, Outlet, useLocation } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { Logo } from './Logo'

const navigation = [
  { label: 'Dashboard', to: '/', icon: Gauge },
  { label: 'Assets', to: '/assets', icon: Boxes },
  { label: 'Scan jobs', to: '/scans', icon: Radar },
  { label: 'Findings', to: '/findings', icon: ShieldAlert },
  { label: 'Certificates', to: '/certificates', icon: Fingerprint },
  { label: 'Crypto agility', to: '/pqc', icon: BarChart3 },
  { label: 'Reports', to: '/reports', icon: FileBarChart },
]

function Navigation({ close }: { close?: () => void }) {
  return (
    <nav className="mt-9 space-y-1">
      {navigation.map(({ label, to, icon: Icon }) => (
        <NavLink
          key={to}
          to={to}
          onClick={close}
          className={({ isActive }) => `group flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm transition ${
            isActive
              ? 'border border-signal/15 bg-signal/10 text-white'
              : 'border border-transparent text-slate-500 hover:bg-white/[0.03] hover:text-slate-200'
          }`}
        >
          {({ isActive }) => (
            <>
              <Icon size={17} className={isActive ? 'text-signal' : 'text-slate-600 group-hover:text-slate-400'} />
              {label}
              {isActive && <span className="ml-auto h-1.5 w-1.5 rounded-full bg-signal shadow-glow" />}
            </>
          )}
        </NavLink>
      ))}
    </nav>
  )
}

export function AppShell() {
  const [mobileOpen, setMobileOpen] = useState(false)
  const { user, logout } = useAuth()
  const location = useLocation()

  return (
    <div className="min-h-screen bg-ink-950 text-slate-200">
      <aside className="fixed inset-y-0 left-0 z-30 hidden w-64 border-r border-white/[0.06] bg-ink-900/95 px-5 py-6 backdrop-blur-xl lg:block">
        <Logo />
        <Navigation />
        <div className="absolute bottom-5 left-5 right-5">
          <div className="mb-3 rounded-xl border border-white/[0.06] bg-white/[0.025] p-3">
            <div className="flex items-center gap-3">
              <div className="grid h-8 w-8 place-items-center rounded-lg bg-cyan-350/10 text-xs font-semibold text-cyan-350">
                {user?.name?.slice(0, 2).toUpperCase()}
              </div>
              <div className="min-w-0">
                <div className="truncate text-xs font-medium text-slate-200">{user?.name}</div>
                <div className="truncate text-[10px] text-slate-600">{user?.email}</div>
              </div>
            </div>
          </div>
          <button onClick={logout} className="flex w-full items-center gap-3 rounded-xl px-3 py-2 text-xs text-slate-600 transition hover:bg-rose-400/5 hover:text-rose-300">
            <LogOut size={15} /> Sign out
          </button>
        </div>
      </aside>

      {mobileOpen && (
        <div className="fixed inset-0 z-50 bg-black/70 backdrop-blur-sm lg:hidden" onClick={() => setMobileOpen(false)}>
          <aside className="h-full w-72 border-r border-white/10 bg-ink-900 p-5" onClick={(event) => event.stopPropagation()}>
            <div className="flex items-center justify-between">
              <Logo />
              <button onClick={() => setMobileOpen(false)} className="text-slate-500"><X size={20} /></button>
            </div>
            <Navigation close={() => setMobileOpen(false)} />
          </aside>
        </div>
      )}

      <main className="min-h-screen lg:pl-64">
        <header className="sticky top-0 z-20 flex h-16 items-center justify-between border-b border-white/[0.05] bg-ink-950/80 px-4 backdrop-blur-xl md:px-8">
          <button onClick={() => setMobileOpen(true)} className="text-slate-400 lg:hidden"><Menu size={21} /></button>
          <div className="hidden items-center gap-2 text-[10px] uppercase tracking-[0.18em] text-slate-600 sm:flex">
            <ClipboardList size={13} />
            Security workspace
            <span className="text-slate-800">/</span>
            <span className="text-slate-400">{navigation.find((item) => item.to === location.pathname)?.label || 'Asset intelligence'}</span>
          </div>
          <div className="ml-auto flex items-center gap-2 rounded-full border border-signal/15 bg-signal/5 px-3 py-1.5 text-[10px] font-medium uppercase tracking-wider text-signal">
            <span className="h-1.5 w-1.5 animate-pulse rounded-full bg-signal" />
            Monitoring
          </div>
        </header>
        <div className="mx-auto max-w-[1500px] px-4 py-7 md:px-8 md:py-9">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
