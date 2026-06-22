import { Activity, ArrowUpRight, Boxes, Radar, ShieldAlert, Sparkles } from 'lucide-react'
import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Badge, EmptyState, LoadingScreen, PageHeader, ScoreRing, StatCard } from '../components/ui'
import api, { errorMessage } from '../lib/api'
import { relativeTime } from '../lib/format'
import type { DashboardData } from '../types'

export default function DashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    api.get<DashboardData>('/dashboard')
      .then((response) => setData(response.data))
      .catch((requestError) => setError(errorMessage(requestError)))
  }, [])

  if (!data && !error) return <LoadingScreen />

  return (
    <>
      <PageHeader
        eyebrow="Security overview"
        title="Cryptographic posture"
        description="A live view of TLS exposure, certificate health, and the work required to become crypto-agile."
        action={<Link className="btn-primary" to="/assets"><Radar size={16} /> Add scan target</Link>}
      />
      {error && <div className="mb-6 rounded-xl border border-rose-400/20 bg-rose-400/10 p-4 text-sm text-rose-300">{error}</div>}
      {data && (
        <>
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            <StatCard icon={Boxes} label="Managed assets" value={data.summary.asset_count} detail={`${data.summary.assessed_count} with completed assessments`} />
            <StatCard icon={Activity} label="Average risk" value={data.summary.average_risk_score} detail="Weighted TLS and PKI exposure · lower is safer" tone="amber" />
            <StatCard icon={Sparkles} label="Crypto agility" value={data.summary.average_pqc_score} detail="Migration-readiness score across assessed assets" tone="cyan" />
            <StatCard icon={ShieldAlert} label="Critical findings" value={data.summary.critical_findings} detail="Open findings requiring immediate attention" tone="rose" />
          </div>

          {data.assets.length === 0 ? (
            <div className="mt-6">
              <EmptyState icon={Radar} title="Your attack surface is waiting" message="Add the first internet-facing domain, then queue a TLS scan to build your certificate inventory." action={<Link to="/assets" className="btn-primary">Add first asset</Link>} />
            </div>
          ) : (
            <div className="mt-6 grid gap-6 xl:grid-cols-[1.3fr_.7fr]">
              <section className="panel p-5 md:p-6">
                <div className="mb-5 flex items-center justify-between">
                  <div>
                    <h2 className="font-medium text-white">Exposure by asset</h2>
                    <p className="mt-1 text-xs text-slate-600">Risk and readiness on the same 0–100 scale</p>
                  </div>
                  <Link to="/assets" className="text-xs text-slate-500 hover:text-signal">View inventory</Link>
                </div>
                <div className="space-y-5">
                  {data.assets.slice(0, 6).map((asset) => (
                    <Link key={asset.id} to={`/assets/${asset.id}`} className="group block">
                      <div className="mb-2 flex items-center justify-between gap-4">
                        <div className="min-w-0">
                          <span className="truncate text-sm font-medium text-slate-300 group-hover:text-white">{asset.domain}</span>
                          <span className="ml-2 font-mono text-[10px] text-slate-700">:{asset.port}</span>
                        </div>
                        <div className="flex gap-5 text-[10px] uppercase tracking-wider">
                          <span className="text-slate-600">Risk <b className="ml-1 text-amber-300">{asset.current_risk_score}</b></span>
                          <span className="text-slate-600">Agility <b className="ml-1 text-cyan-350">{asset.current_pqc_score}</b></span>
                        </div>
                      </div>
                      <div className="grid grid-cols-2 gap-2">
                        <div className="h-1.5 overflow-hidden rounded-full bg-slate-800">
                          <div className="h-full rounded-full bg-gradient-to-r from-amber-500 to-rose-400" style={{ width: `${asset.current_risk_score}%` }} />
                        </div>
                        <div className="h-1.5 overflow-hidden rounded-full bg-slate-800">
                          <div className="h-full rounded-full bg-gradient-to-r from-cyan-500 to-signal" style={{ width: `${asset.current_pqc_score}%` }} />
                        </div>
                      </div>
                    </Link>
                  ))}
                </div>
              </section>

              <section className="panel flex items-center justify-around p-6">
                <div className="text-center">
                  <ScoreRing score={data.summary.average_risk_score} label="risk" inverse />
                  <div className="mt-3 text-xs text-slate-500">Portfolio risk</div>
                </div>
                <div className="h-28 w-px bg-white/[0.06]" />
                <div className="text-center">
                  <ScoreRing score={data.summary.average_pqc_score} label="ready" />
                  <div className="mt-3 text-xs text-slate-500">Crypto agility</div>
                </div>
              </section>
            </div>
          )}

          <div className="mt-6 grid gap-6 xl:grid-cols-2">
            <section className="panel overflow-hidden">
              <div className="flex items-center justify-between border-b border-white/[0.06] px-5 py-4">
                <h2 className="text-sm font-medium text-white">Recent scan jobs</h2>
                <Link to="/scans" className="text-xs text-slate-600 hover:text-signal">All scans</Link>
              </div>
              {data.recent_scans.length ? data.recent_scans.map((scan) => (
                <Link to={`/assets/${scan.asset_id}`} key={scan.id} className="flex items-center gap-4 border-b border-white/[0.04] px-5 py-4 last:border-0 hover:bg-white/[0.02]">
                  <span className={`h-2 w-2 rounded-full ${scan.status === 'completed' ? 'bg-signal' : scan.status === 'failed' ? 'bg-rose-400' : 'animate-pulse bg-amber-400'}`} />
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-sm font-medium text-slate-300">{scan.asset?.domain || 'Asset scan'}</div>
                    <div className="mt-1 text-[11px] text-slate-600">{relativeTime(scan.created_at)} · {scan.tls_version || 'TLS pending'}</div>
                  </div>
                  <Badge value={scan.status} />
                </Link>
              )) : <div className="p-8 text-center text-sm text-slate-600">No scan jobs yet.</div>}
            </section>

            <section className="panel overflow-hidden">
              <div className="flex items-center justify-between border-b border-white/[0.06] px-5 py-4">
                <h2 className="text-sm font-medium text-white">Priority findings</h2>
                <Link to="/findings" className="text-xs text-slate-600 hover:text-signal">All findings</Link>
              </div>
              {data.priority_findings.length ? data.priority_findings.map((finding) => (
                <Link to={`/assets/${finding.asset_id}`} key={finding.id} className="group flex items-center gap-4 border-b border-white/[0.04] px-5 py-4 last:border-0 hover:bg-white/[0.02]">
                  <Badge value={finding.severity} />
                  <div className="min-w-0 flex-1">
                    <div className="truncate text-sm text-slate-300">{finding.title}</div>
                    <div className="mt-1 text-[11px] text-slate-600">{finding.evidence}</div>
                  </div>
                  <ArrowUpRight size={15} className="text-slate-700 group-hover:text-signal" />
                </Link>
              )) : <div className="p-8 text-center text-sm text-slate-600">No open findings. Nicely quiet.</div>}
            </section>
          </div>
        </>
      )}
    </>
  )
}
