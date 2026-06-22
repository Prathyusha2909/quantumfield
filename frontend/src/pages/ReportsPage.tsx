import { BarChart3, Download, FileBarChart, ShieldAlert, TimerReset } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { EmptyState, LoadingScreen, PageHeader, StatCard } from '../components/ui'
import api, { errorMessage } from '../lib/api'
import { formatDate, titleCase } from '../lib/format'
import type { Asset, ReportSummary } from '../types'

const severityColors: Record<string, string> = {
  critical: 'bg-rose-400',
  high: 'bg-orange-400',
  medium: 'bg-amber-400',
  low: 'bg-cyan-350',
  info: 'bg-slate-500',
}

export default function ReportsPage() {
  const [report, setReport] = useState<ReportSummary | null>(null)
  const [assets, setAssets] = useState<Asset[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    Promise.all([api.get<ReportSummary>('/reports/summary'), api.get<Asset[]>('/assets')])
      .then(([reportResponse, assetResponse]) => {
        setReport(reportResponse.data)
        setAssets(assetResponse.data)
      })
      .catch((requestError) => setError(errorMessage(requestError)))
      .finally(() => setLoading(false))
  }, [])

  const totalFindings = useMemo(() => report?.findings_by_severity.reduce((sum, item) => sum + item.count, 0) || 0, [report])
  const highRisk = assets.filter((asset) => asset.current_risk_score >= 70).length

  function downloadJSON() {
    const payload = JSON.stringify({ report, assets }, null, 2)
    const url = URL.createObjectURL(new Blob([payload], { type: 'application/json' }))
    const link = document.createElement('a')
    link.href = url
    link.download = `quantumfield-report-${new Date().toISOString().slice(0, 10)}.json`
    link.click()
    URL.revokeObjectURL(url)
  }

  if (loading) return <LoadingScreen />

  return (
    <>
      <PageHeader eyebrow="Portfolio evidence" title="Security reports" description="A concise export of TLS exposure, certificate algorithms, and near-term renewal pressure." action={<button className="btn-secondary" onClick={downloadJSON} disabled={!report}><Download size={15} /> Export JSON</button>} />
      {error && <div className="mb-5 rounded-xl border border-rose-400/20 bg-rose-400/10 p-4 text-sm text-rose-300">{error}</div>}
      {!report ? <EmptyState icon={FileBarChart} title="Report unavailable" message="Add and scan assets before generating portfolio evidence." /> : (
        <>
          <div className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
            <StatCard icon={FileBarChart} label="Reported assets" value={assets.length} detail={`Generated ${formatDate(report.generated_at, true)}`} />
            <StatCard icon={ShieldAlert} label="Total findings" value={totalFindings} detail="Across all retained scan evidence" tone="amber" />
            <StatCard icon={BarChart3} label="High-risk assets" value={highRisk} detail="Assets at or above a 70 risk score" tone="rose" />
            <StatCard icon={TimerReset} label="Expiring ≤90 days" value={report.certificates_expiring_90_days} detail="Certificate records entering renewal windows" tone="cyan" />
          </div>

          <div className="mt-6 grid gap-6 xl:grid-cols-2">
            <section className="panel p-5 md:p-6">
              <h2 className="font-medium text-white">Findings by severity</h2>
              <p className="mt-1 text-xs text-slate-600">All retained observations</p>
              <div className="mt-7 space-y-5">
                {report.findings_by_severity.length ? report.findings_by_severity
                  .sort((a, b) => b.count - a.count)
                  .map((item) => (
                    <div key={item.severity}>
                      <div className="mb-2 flex justify-between text-xs"><span className="text-slate-500">{titleCase(item.severity)}</span><span className="font-mono text-slate-300">{item.count}</span></div>
                      <div className="h-2 overflow-hidden rounded-full bg-slate-800"><div className={`h-full rounded-full ${severityColors[item.severity] || 'bg-slate-500'}`} style={{ width: `${Math.max(4, (item.count / totalFindings) * 100)}%` }} /></div>
                    </div>
                  )) : <div className="py-10 text-center text-sm text-slate-600">No findings in the report window.</div>}
              </div>
            </section>

            <section className="panel p-5 md:p-6">
              <h2 className="font-medium text-white">Certificate key algorithms</h2>
              <p className="mt-1 text-xs text-slate-600">Classical cryptography observed across scan history</p>
              <div className="mt-7 space-y-4">
                {report.certificates_by_algorithm.length ? report.certificates_by_algorithm.map((item) => {
                  const total = report.certificates_by_algorithm.reduce((sum, row) => sum + row.count, 0)
                  return (
                    <div key={item.algorithm} className="flex items-center gap-4 rounded-xl border border-white/[0.05] bg-white/[0.02] p-4">
                      <div className="grid h-10 w-10 place-items-center rounded-xl bg-cyan-350/10 font-mono text-xs text-cyan-350">{item.algorithm.slice(0, 3)}</div>
                      <div className="flex-1"><div className="text-sm text-slate-300">{item.algorithm}</div><div className="mt-2 h-1.5 overflow-hidden rounded-full bg-slate-800"><div className="h-full bg-cyan-350" style={{ width: `${(item.count / total) * 100}%` }} /></div></div>
                      <div className="font-mono text-sm text-white">{item.count}</div>
                    </div>
                  )
                }) : <div className="py-10 text-center text-sm text-slate-600">No certificate algorithms observed yet.</div>}
              </div>
            </section>
          </div>

          <section className="panel mt-6 overflow-hidden">
            <div className="border-b border-white/[0.06] px-5 py-4"><h2 className="text-sm font-medium text-white">Asset posture appendix</h2></div>
            <div className="overflow-x-auto"><table className="data-table"><thead><tr><th>Domain</th><th>Status</th><th>Risk score</th><th>Crypto agility</th><th>Last assessed</th></tr></thead><tbody>
              {assets.map((asset) => <tr key={asset.id}><td className="font-medium text-slate-200">{asset.domain}:{asset.port}</td><td>{titleCase(asset.status)}</td><td className="font-mono text-amber-300">{asset.current_risk_score}</td><td className="font-mono text-cyan-350">{asset.current_pqc_score}</td><td>{formatDate(asset.last_scanned_at, true)}</td></tr>)}
            </tbody></table></div>
          </section>
        </>
      )}
    </>
  )
}
