import { ArrowLeft, CalendarClock, CheckCircle2, Copy, Fingerprint, Radar, ShieldAlert, XCircle } from 'lucide-react'
import { useCallback, useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Badge, BooleanSignal, LoadingScreen, PageHeader, ScoreRing } from '../components/ui'
import api, { errorMessage } from '../lib/api'
import { formatDate, relativeTime } from '../lib/format'
import type { Asset, Finding } from '../types'

export default function AssetDetailsPage() {
  const { id } = useParams()
  const [asset, setAsset] = useState<Asset | null>(null)
  const [findings, setFindings] = useState<Finding[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [scanning, setScanning] = useState(false)

  const load = useCallback(async () => {
    try {
      const [assetResponse, findingResponse] = await Promise.all([
        api.get<Asset>(`/assets/${id}`),
        api.get<Finding[]>(`/assets/${id}/findings`),
      ])
      setAsset(assetResponse.data)
      setFindings(findingResponse.data)
    } catch (requestError) {
      setError(errorMessage(requestError))
    } finally {
      setLoading(false)
    }
  }, [id])

  useEffect(() => { void load() }, [load])

  useEffect(() => {
    if (!asset || !['queued', 'running'].includes(asset.status)) return
    const timer = window.setInterval(() => void load(), 4000)
    return () => window.clearInterval(timer)
  }, [asset, load])

  async function startScan() {
    setScanning(true)
    setError('')
    try {
      await api.post(`/assets/${id}/scan`)
      await load()
    } catch (requestError) {
      setError(errorMessage(requestError))
    } finally {
      setScanning(false)
    }
  }

  if (loading) return <LoadingScreen />
  if (!asset) return <div className="panel p-8 text-rose-300">{error || 'Asset not found'}</div>

  const latestScan = asset.scans?.[0]
  const certificate = latestScan?.certificate
  const assessment = latestScan?.pqc_assessment

  return (
    <>
      <Link to="/assets" className="mb-5 inline-flex items-center gap-2 text-xs text-slate-600 hover:text-signal"><ArrowLeft size={14} /> Back to inventory</Link>
      <PageHeader
        eyebrow={asset.label || 'Managed endpoint'}
        title={asset.domain}
        description={`TLS endpoint ${asset.domain}:${asset.port} · Added ${formatDate(asset.created_at)}`}
        action={<button className="btn-primary" onClick={() => void startScan()} disabled={scanning || ['queued', 'running'].includes(asset.status)}><Radar size={16} /> {scanning ? 'Queueing…' : ['queued', 'running'].includes(asset.status) ? 'Scan in progress' : 'Run new scan'}</button>}
      />
      {error && <div className="mb-5 rounded-xl border border-rose-400/20 bg-rose-400/10 px-4 py-3 text-sm text-rose-300">{error}</div>}

      <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-4">
        <div className="panel flex items-center gap-5 p-5"><ScoreRing score={asset.current_risk_score} label="risk" inverse size={88} /><div><div className="text-xs uppercase tracking-wider text-slate-600">Risk score</div><div className="mt-2 text-sm text-slate-300">{asset.current_risk_score >= 70 ? 'Immediate action' : asset.current_risk_score >= 40 ? 'Elevated exposure' : 'Controlled exposure'}</div></div></div>
        <div className="panel flex items-center gap-5 p-5"><ScoreRing score={asset.current_pqc_score} label="agility" size={88} /><div><div className="text-xs uppercase tracking-wider text-slate-600">Crypto agility</div><div className="mt-2 text-sm text-slate-300">Grade {assessment?.grade || '—'}</div></div></div>
        <div className="panel p-5"><div className="text-xs uppercase tracking-wider text-slate-600">Current state</div><div className="mt-4"><Badge value={asset.status} /></div><div className="mt-3 text-xs text-slate-600">{relativeTime(asset.last_scanned_at)}</div></div>
        <div className="panel p-5"><div className="text-xs uppercase tracking-wider text-slate-600">Negotiated transport</div><div className="mt-3 text-xl font-medium text-white">{latestScan?.tls_version || 'Not scanned'}</div><div className="mt-2 truncate font-mono text-[10px] text-slate-600">{latestScan?.cipher_suite || 'Cipher suite pending'}</div></div>
      </div>

      {!latestScan ? (
        <div className="panel mt-6 grid min-h-56 place-items-center p-8 text-center">
          <div><Radar className="mx-auto text-slate-600" /><h3 className="mt-4 text-white">No scan evidence yet</h3><p className="mt-2 text-sm text-slate-600">Run a scan to populate the certificate inventory and findings.</p></div>
        </div>
      ) : (
        <>
          <div className="mt-6 grid gap-6 xl:grid-cols-[1.2fr_.8fr]">
            <section className="panel overflow-hidden">
              <div className="flex items-center gap-3 border-b border-white/[0.06] px-5 py-4"><Fingerprint size={17} className="text-cyan-350" /><h2 className="text-sm font-medium text-white">Leaf certificate</h2></div>
              {certificate ? (
                <div className="grid gap-x-8 gap-y-5 p-5 md:grid-cols-2">
                  {[
                    ['Common name', certificate.common_name],
                    ['Issuer', certificate.issuer],
                    ['Valid from', formatDate(certificate.not_before, true)],
                    ['Expires', formatDate(certificate.not_after, true)],
                    ['Public key', `${certificate.public_key_algorithm} · ${certificate.key_size} bit`],
                    ['Signature', certificate.signature_algorithm],
                  ].map(([label, value]) => <div key={label}><div className="text-[10px] uppercase tracking-wider text-slate-600">{label}</div><div className="mt-1.5 break-words text-sm text-slate-300">{value}</div></div>)}
                  <div><div className="text-[10px] uppercase tracking-wider text-slate-600">Chain validation</div><div className="mt-2"><BooleanSignal value={certificate.chain_valid} trueLabel="Trusted chain" falseLabel="Invalid chain" /></div></div>
                  <div><div className="text-[10px] uppercase tracking-wider text-slate-600">Hostname validation</div><div className="mt-2"><BooleanSignal value={certificate.hostname_valid} trueLabel="Hostname covered" falseLabel="Name mismatch" /></div></div>
                  <div className="md:col-span-2"><div className="text-[10px] uppercase tracking-wider text-slate-600">SHA-256 fingerprint</div><button onClick={() => navigator.clipboard.writeText(certificate.fingerprint_sha256)} className="mt-2 flex max-w-full items-center gap-2 break-all text-left font-mono text-[10px] leading-5 text-slate-500 hover:text-cyan-350">{certificate.fingerprint_sha256}<Copy size={12} className="shrink-0" /></button></div>
                </div>
              ) : <div className="p-8 text-sm text-slate-600">Certificate retrieval failed for this scan.</div>}
            </section>

            <section className="panel overflow-hidden">
              <div className="border-b border-white/[0.06] px-5 py-4"><h2 className="text-sm font-medium text-white">Migration signals</h2></div>
              {assessment ? (
                <div className="space-y-4 p-5">
                  {[
                    ['RSA dependency', assessment.rsa_dependency, false],
                    ['ECC dependency', assessment.ecc_dependency, false],
                    ['TLS 1.3 negotiated', assessment.tls13_enabled, true],
                    ['Legacy TLS accepted', assessment.legacy_tls_supported, false],
                    ['Rotation-ready lifecycle', assessment.certificate_rotation_ready, true],
                  ].map(([label, value, positive]) => (
                    <div key={String(label)} className="flex items-center justify-between rounded-xl border border-white/[0.05] bg-white/[0.02] px-4 py-3">
                      <span className="text-xs text-slate-400">{label as string}</span>
                      {(positive ? value : !value) ? <CheckCircle2 size={16} className="text-signal" /> : <XCircle size={16} className="text-rose-300" />}
                    </div>
                  ))}
                  <div className="pt-1 text-[11px] leading-5 text-slate-600">{assessment.rationale[0]}</div>
                </div>
              ) : <div className="p-8 text-sm text-slate-600">No crypto-agility assessment available.</div>}
            </section>
          </div>

          <section className="panel mt-6 overflow-hidden">
            <div className="flex items-center gap-3 border-b border-white/[0.06] px-5 py-4"><ShieldAlert size={17} className="text-amber-300" /><h2 className="text-sm font-medium text-white">Detected findings</h2><span className="ml-auto text-xs text-slate-600">{findings.length} observations</span></div>
            {findings.length ? findings.slice(0, 12).map((finding) => (
              <div key={finding.id} className="grid gap-4 border-b border-white/[0.04] px-5 py-5 last:border-0 md:grid-cols-[100px_1fr]">
                <div><Badge value={finding.severity} /></div>
                <div>
                  <h3 className="text-sm font-medium text-slate-200">{finding.title}</h3>
                  <p className="mt-2 text-xs leading-5 text-slate-500">{finding.description}</p>
                  <div className="mt-3 rounded-lg border border-white/[0.04] bg-black/10 px-3 py-2 text-[11px] text-slate-600"><span className="text-slate-500">Remediation:</span> {finding.remediation}</div>
                </div>
              </div>
            )) : <div className="p-8 text-center text-sm text-slate-600">No findings recorded for this asset.</div>}
          </section>

          <section className="panel mt-6 overflow-hidden">
            <div className="flex items-center gap-3 border-b border-white/[0.06] px-5 py-4"><CalendarClock size={17} className="text-slate-500" /><h2 className="text-sm font-medium text-white">Scan history</h2></div>
            <div className="overflow-x-auto">
              <table className="data-table"><thead><tr><th>Started</th><th>Status</th><th>TLS</th><th>Risk</th><th>Agility</th><th>Duration</th></tr></thead><tbody>
                {asset.scans?.map((scan) => <tr key={scan.id}><td>{formatDate(scan.created_at, true)}</td><td><Badge value={scan.status} /></td><td>{scan.tls_version || '—'}</td><td className="font-mono text-amber-300">{scan.risk_score}</td><td className="font-mono text-cyan-350">{scan.pqc_score}</td><td>{scan.duration_ms ? `${scan.duration_ms} ms` : '—'}</td></tr>)}
              </tbody></table>
            </div>
          </section>
        </>
      )}
    </>
  )
}
