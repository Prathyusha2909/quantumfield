import { Clock3, Radar } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { Badge, EmptyState, LoadingScreen, PageHeader } from '../components/ui'
import api, { errorMessage } from '../lib/api'
import { formatDate } from '../lib/format'
import type { Scan } from '../types'

export default function ScansPage() {
  const [scans, setScans] = useState<Scan[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const load = useCallback(async () => {
    try {
      const { data } = await api.get<Scan[]>('/scans')
      setScans(data)
    } catch (requestError) {
      setError(errorMessage(requestError))
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    void load()
    const timer = window.setInterval(() => void load(), 5000)
    return () => window.clearInterval(timer)
  }, [load])

  if (loading) return <LoadingScreen />

  return (
    <>
      <PageHeader eyebrow="Async scanner" title="Scan jobs" description="Redis-backed jobs processed independently by the Go TLS worker." />
      {error && <div className="mb-5 rounded-xl border border-rose-400/20 bg-rose-400/10 p-4 text-sm text-rose-300">{error}</div>}
      {!scans.length ? <EmptyState icon={Radar} title="No scan jobs" message="Start a scan from the asset inventory to see queue and worker activity here." /> : (
        <div className="table-wrap">
          <table className="data-table">
            <thead><tr><th>Target</th><th>Queued</th><th>Status</th><th>TLS / cipher</th><th>Risk</th><th>Agility</th><th>Runtime</th></tr></thead>
            <tbody>{scans.map((scan) => (
              <tr key={scan.id}>
                <td><Link className="font-medium text-slate-200 hover:text-signal" to={`/assets/${scan.asset_id}`}>{scan.asset?.domain || scan.asset_id.slice(0, 8)}</Link><div className="mt-1 font-mono text-[10px] text-slate-700">{scan.id.slice(0, 13)}</div></td>
                <td><div className="flex items-center gap-2 text-xs"><Clock3 size={13} className="text-slate-700" />{formatDate(scan.created_at, true)}</div></td>
                <td><Badge value={scan.status} />{scan.error_message && <div className="mt-2 max-w-xs text-[10px] text-rose-300">{scan.error_message}</div>}</td>
                <td><div className="text-xs text-slate-300">{scan.tls_version || 'Pending'}</div><div className="mt-1 max-w-xs truncate font-mono text-[10px] text-slate-700">{scan.cipher_suite || '—'}</div></td>
                <td className="font-mono text-amber-300">{scan.status === 'completed' ? scan.risk_score : '—'}</td>
                <td className="font-mono text-cyan-350">{scan.status === 'completed' ? scan.pqc_score : '—'}</td>
                <td>{scan.duration_ms ? `${(scan.duration_ms / 1000).toFixed(2)} s` : '—'}</td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      )}
    </>
  )
}
