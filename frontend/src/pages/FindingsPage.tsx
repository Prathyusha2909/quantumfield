import { Search, ShieldCheck } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Badge, EmptyState, LoadingScreen, PageHeader } from '../components/ui'
import api, { errorMessage } from '../lib/api'
import { formatDate, titleCase } from '../lib/format'
import type { Finding } from '../types'

const severities = ['all', 'critical', 'high', 'medium', 'low', 'info']

export default function FindingsPage() {
  const [findings, setFindings] = useState<Finding[]>([])
  const [loading, setLoading] = useState(true)
  const [filter, setFilter] = useState('all')
  const [query, setQuery] = useState('')
  const [error, setError] = useState('')

  useEffect(() => {
    api.get<Finding[]>('/findings')
      .then(({ data }) => setFindings(data))
      .catch((requestError) => setError(errorMessage(requestError)))
      .finally(() => setLoading(false))
  }, [])

  const filtered = useMemo(() => findings.filter((finding) =>
    (filter === 'all' || finding.severity === filter) &&
    `${finding.title} ${finding.type} ${finding.asset?.domain}`.toLowerCase().includes(query.toLowerCase()),
  ), [filter, findings, query])

  if (loading) return <LoadingScreen />

  return (
    <>
      <PageHeader eyebrow="Prioritized remediation" title="Security findings" description="Evidence-backed TLS, certificate, and quantum-dependency observations from every scan." />
      {error && <div className="mb-5 rounded-xl border border-rose-400/20 bg-rose-400/10 p-4 text-sm text-rose-300">{error}</div>}
      <div className="mb-5 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex flex-wrap gap-2">
          {severities.map((severity) => <button key={severity} onClick={() => setFilter(severity)} className={`rounded-full border px-3 py-1.5 text-[10px] font-semibold uppercase tracking-wider transition ${filter === severity ? 'border-signal/30 bg-signal/10 text-signal' : 'border-slate-800 text-slate-600 hover:text-slate-300'}`}>{severity}</button>)}
        </div>
        <div className="flex min-w-72 items-center gap-2 rounded-xl border border-white/[0.06] bg-ink-850 px-3 py-2">
          <Search size={15} className="text-slate-700" /><input className="w-full bg-transparent text-xs text-white outline-none placeholder:text-slate-700" value={query} onChange={(event) => setQuery(event.target.value)} placeholder="Search findings…" />
        </div>
      </div>
      {!filtered.length ? <EmptyState icon={ShieldCheck} title="No matching findings" message="No observations match the current filters." /> : (
        <div className="space-y-3">
          {filtered.map((finding) => (
            <article key={finding.id} className="panel p-5">
              <div className="flex flex-col gap-4 md:flex-row md:items-start">
                <div className="md:w-24"><Badge value={finding.severity} /></div>
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-x-3 gap-y-2">
                    <h2 className="font-medium text-white">{finding.title}</h2>
                    <span className="rounded bg-white/[0.04] px-2 py-1 font-mono text-[9px] uppercase tracking-wider text-slate-600">{titleCase(finding.type)}</span>
                  </div>
                  <p className="mt-2 text-sm leading-6 text-slate-500">{finding.description}</p>
                  <div className="mt-4 grid gap-3 lg:grid-cols-2">
                    <div className="rounded-xl border border-white/[0.05] bg-white/[0.02] p-3"><div className="text-[9px] uppercase tracking-wider text-slate-700">Evidence</div><div className="mt-1.5 text-xs text-slate-500">{finding.evidence}</div></div>
                    <div className="rounded-xl border border-signal/10 bg-signal/[0.025] p-3"><div className="text-[9px] uppercase tracking-wider text-signal/50">Remediation</div><div className="mt-1.5 text-xs text-slate-400">{finding.remediation}</div></div>
                  </div>
                </div>
                <div className="text-right text-[10px] text-slate-700">
                  <Link to={`/assets/${finding.asset_id}`} className="text-cyan-350 hover:text-white">{finding.asset?.domain || 'Open asset'}</Link>
                  <div className="mt-2">{formatDate(finding.created_at)}</div>
                </div>
              </div>
            </article>
          ))}
        </div>
      )}
    </>
  )
}

