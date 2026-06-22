import { Atom, CheckCircle2, CircleOff, KeyRound, Network, RefreshCw, Waves } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { EmptyState, LoadingScreen, PageHeader, ScoreRing, StatCard } from '../components/ui'
import api, { errorMessage } from '../lib/api'
import type { PQCAssessment } from '../types'

export default function PQCPage() {
  const [assessments, setAssessments] = useState<PQCAssessment[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    api.get<PQCAssessment[]>('/pqc-assessments')
      .then(({ data }) => setAssessments(data))
      .catch((requestError) => setError(errorMessage(requestError)))
      .finally(() => setLoading(false))
  }, [])

  const latest = useMemo(() => {
    const seen = new Set<string>()
    return assessments.filter((assessment) => {
      if (seen.has(assessment.asset_id)) return false
      seen.add(assessment.asset_id)
      return true
    })
  }, [assessments])
  const average = latest.length ? Math.round(latest.reduce((sum, item) => sum + item.score, 0) / latest.length) : 0
  const rsa = latest.filter((item) => item.rsa_dependency).length
  const ecc = latest.filter((item) => item.ecc_dependency).length
  const tls13 = latest.filter((item) => item.tls13_enabled).length

  if (loading) return <LoadingScreen />

  return (
    <>
      <PageHeader eyebrow="Crypto agility" title="Post-quantum readiness" description="A migration-oriented assessment of classical key dependencies, protocol posture, and certificate rotation agility." />
      {error && <div className="mb-5 rounded-xl border border-rose-400/20 bg-rose-400/10 p-4 text-sm text-rose-300">{error}</div>}
      <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard icon={Atom} label="Portfolio readiness" value={average} detail="Average score across latest asset assessments" tone="cyan" />
        <StatCard icon={KeyRound} label="RSA dependency" value={rsa} detail={`${latest.length ? Math.round((rsa / latest.length) * 100) : 0}% of assessed endpoints`} tone="rose" />
        <StatCard icon={Waves} label="ECC dependency" value={ecc} detail={`${latest.length ? Math.round((ecc / latest.length) * 100) : 0}% of assessed endpoints`} tone="amber" />
        <StatCard icon={Network} label="TLS 1.3" value={`${tls13}/${latest.length}`} detail="Endpoints negotiating the agility baseline" />
      </div>

      <div className="panel mt-6 p-5 md:p-6">
        <div className="grid gap-6 lg:grid-cols-[180px_1fr] lg:items-center">
          <div className="mx-auto text-center"><ScoreRing score={average} label="portfolio" size={132} /><div className="mt-3 text-xs text-slate-600">Readiness baseline</div></div>
          <div>
            <h2 className="font-medium text-white">What this score means</h2>
            <p className="mt-2 text-sm leading-6 text-slate-500">QuantumField does not claim the endpoint already uses post-quantum TLS. It measures how much classical cryptography remains, whether modern TLS is in use, and whether certificate operations are agile enough to support a future migration.</p>
            <div className="mt-5 grid gap-3 sm:grid-cols-3">
              {[['80–100', 'Migration leading'], ['50–79', 'Preparation needed'], ['0–49', 'Classical dependency']].map(([range, label]) => <div key={range} className="rounded-xl border border-white/[0.05] bg-white/[0.02] p-3"><div className="font-mono text-xs text-cyan-350">{range}</div><div className="mt-1 text-[11px] text-slate-600">{label}</div></div>)}
            </div>
          </div>
        </div>
      </div>

      {!latest.length ? <div className="mt-6"><EmptyState icon={Atom} title="No readiness assessments" message="Run a TLS scan to calculate cryptographic dependencies and migration readiness." /></div> : (
        <div className="mt-6 grid gap-4 xl:grid-cols-2">
          {latest.map((assessment) => (
            <Link to={`/assets/${assessment.asset_id}`} key={assessment.id} className="panel group p-5 transition hover:border-signal/15">
              <div className="flex items-start gap-5">
                <ScoreRing score={assessment.score} label={`grade ${assessment.grade}`} size={88} />
                <div className="min-w-0 flex-1">
                  <div className="flex items-center justify-between gap-3">
                    <h2 className="truncate font-medium text-white group-hover:text-signal">{assessment.asset?.domain || assessment.asset_id}</h2>
                    {assessment.quantum_vulnerable ? <span className="text-[10px] uppercase tracking-wider text-rose-300">Classical</span> : <span className="text-[10px] uppercase tracking-wider text-signal">Candidate</span>}
                  </div>
                  <div className="mt-4 grid grid-cols-2 gap-2 text-[11px]">
                    <Signal label="RSA dependency" healthy={!assessment.rsa_dependency} />
                    <Signal label="ECC dependency" healthy={!assessment.ecc_dependency} />
                    <Signal label="TLS 1.3" healthy={assessment.tls13_enabled} />
                    <Signal label="Rotation agility" healthy={assessment.certificate_rotation_ready} />
                  </div>
                </div>
              </div>
              <div className="mt-4 flex items-start gap-2 border-t border-white/[0.05] pt-4 text-[11px] leading-5 text-slate-600"><RefreshCw size={13} className="mt-1 shrink-0 text-slate-700" />{assessment.rationale[0]}</div>
            </Link>
          ))}
        </div>
      )}
    </>
  )
}

function Signal({ label, healthy }: { label: string; healthy: boolean }) {
  return <div className="flex items-center gap-2 text-slate-500">{healthy ? <CheckCircle2 size={13} className="text-signal" /> : <CircleOff size={13} className="text-rose-300" />}{label}</div>
}

