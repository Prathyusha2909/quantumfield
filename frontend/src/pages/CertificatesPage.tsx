import { Fingerprint, Search } from 'lucide-react'
import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Badge, BooleanSignal, EmptyState, LoadingScreen, PageHeader } from '../components/ui'
import api, { errorMessage } from '../lib/api'
import { formatDate } from '../lib/format'
import type { Certificate } from '../types'

function certificateState(certificate: Certificate) {
  const days = (new Date(certificate.not_after).getTime() - Date.now()) / 86_400_000
  if (days < 0) return 'expired'
  if (days <= 30) return 'critical'
  if (days <= 90) return 'medium'
  return 'valid'
}

export default function CertificatesPage() {
  const [certificates, setCertificates] = useState<Certificate[]>([])
  const [query, setQuery] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    api.get<Certificate[]>('/certificates')
      .then(({ data }) => setCertificates(data))
      .catch((requestError) => setError(errorMessage(requestError)))
      .finally(() => setLoading(false))
  }, [])

  const latest = useMemo(() => {
    const seen = new Set<string>()
    return certificates.filter((certificate) => {
      if (seen.has(certificate.asset_id)) return false
      seen.add(certificate.asset_id)
      return `${certificate.common_name} ${certificate.issuer} ${certificate.asset?.domain}`.toLowerCase().includes(query.toLowerCase())
    })
  }, [certificates, query])

  if (loading) return <LoadingScreen />

  return (
    <>
      <PageHeader eyebrow="PKI inventory" title="Certificate inventory" description="The latest observed leaf certificate for every managed endpoint." />
      {error && <div className="mb-5 rounded-xl border border-rose-400/20 bg-rose-400/10 p-4 text-sm text-rose-300">{error}</div>}
      <div className="mb-4 flex items-center gap-3 rounded-xl border border-white/[0.06] bg-ink-850/60 px-4 py-2.5">
        <Search size={16} className="text-slate-600" /><input className="w-full bg-transparent text-sm text-white outline-none placeholder:text-slate-700" placeholder="Search subject, issuer, or asset…" value={query} onChange={(event) => setQuery(event.target.value)} />
      </div>
      {!latest.length ? <EmptyState icon={Fingerprint} title="No certificates inventoried" message="Completed scans will add leaf certificates here." /> : (
        <div className="table-wrap">
          <table className="data-table">
            <thead><tr><th>Certificate</th><th>Issuer</th><th>Key</th><th>Validation</th><th>Expiry</th><th>Status</th></tr></thead>
            <tbody>{latest.map((certificate) => (
              <tr key={certificate.id}>
                <td><Link to={`/assets/${certificate.asset_id}`} className="font-medium text-slate-200 hover:text-signal">{certificate.common_name || certificate.asset?.domain}</Link><div className="mt-1 max-w-56 truncate font-mono text-[10px] text-slate-700">{certificate.serial_number}</div></td>
                <td><div className="max-w-64 truncate text-xs text-slate-400" title={certificate.issuer}>{certificate.issuer}</div></td>
                <td><div className="text-xs text-slate-300">{certificate.public_key_algorithm} · {certificate.key_size}</div><div className="mt-1 text-[10px] text-slate-700">{certificate.signature_algorithm}</div></td>
                <td><div className="space-y-2"><BooleanSignal value={certificate.chain_valid} trueLabel="Trusted" falseLabel="Untrusted" /><br /><BooleanSignal value={certificate.hostname_valid} trueLabel="Name valid" falseLabel="Mismatch" /></div></td>
                <td><div className="text-xs text-slate-300">{formatDate(certificate.not_after)}</div><div className="mt-1 text-[10px] text-slate-700">Issued {formatDate(certificate.not_before)}</div></td>
                <td><Badge value={certificateState(certificate)} /></td>
              </tr>
            ))}</tbody>
          </table>
        </div>
      )}
    </>
  )
}

