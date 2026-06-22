import { ExternalLink, Plus, Radar, Search, Trash2, X } from 'lucide-react'
import { FormEvent, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { Badge, EmptyState, LoadingScreen, PageHeader } from '../components/ui'
import api, { errorMessage } from '../lib/api'
import { relativeTime } from '../lib/format'
import type { Asset } from '../types'

export default function AssetsPage() {
  const [assets, setAssets] = useState<Asset[]>([])
  const [loading, setLoading] = useState(true)
  const [showForm, setShowForm] = useState(false)
  const [query, setQuery] = useState('')
  const [form, setForm] = useState({ domain: '', port: 443, label: '', scan: true })
  const [submitting, setSubmitting] = useState(false)
  const [error, setError] = useState('')

  async function loadAssets() {
    try {
      const { data } = await api.get<Asset[]>('/assets')
      setAssets(data)
    } catch (requestError) {
      setError(errorMessage(requestError))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void loadAssets() }, [])

  const filtered = useMemo(() => assets.filter((asset) =>
    `${asset.domain} ${asset.label}`.toLowerCase().includes(query.toLowerCase()),
  ), [assets, query])

  async function createAsset(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    setError('')
    try {
      const { data } = await api.post<Asset>('/assets', {
        domain: form.domain,
        port: Number(form.port),
        label: form.label,
      })
      if (form.scan) await api.post(`/assets/${data.id}/scan`)
      setForm({ domain: '', port: 443, label: '', scan: true })
      setShowForm(false)
      await loadAssets()
    } catch (requestError) {
      setError(errorMessage(requestError))
    } finally {
      setSubmitting(false)
    }
  }

  async function scan(asset: Asset) {
    setError('')
    try {
      await api.post(`/assets/${asset.id}/scan`)
      await loadAssets()
    } catch (requestError) {
      setError(errorMessage(requestError))
    }
  }

  async function remove(asset: Asset) {
    if (!window.confirm(`Remove ${asset.domain} and all of its scan history?`)) return
    try {
      await api.delete(`/assets/${asset.id}`)
      setAssets((items) => items.filter((item) => item.id !== asset.id))
    } catch (requestError) {
      setError(errorMessage(requestError))
    }
  }

  if (loading) return <LoadingScreen />

  return (
    <>
      <PageHeader
        eyebrow="Attack surface"
        title="Domain assets"
        description="Internet-facing endpoints monitored for certificate, TLS, and post-quantum migration risk."
        action={<button className="btn-primary" onClick={() => setShowForm(true)}><Plus size={16} /> Add asset</button>}
      />
      {error && <div className="mb-5 rounded-xl border border-rose-400/20 bg-rose-400/10 px-4 py-3 text-sm text-rose-300">{error}</div>}

      {showForm && (
        <form onSubmit={createAsset} className="panel mb-6 p-5 md:p-6">
          <div className="mb-5 flex items-start justify-between">
            <div>
              <h2 className="font-medium text-white">Add a TLS endpoint</h2>
              <p className="mt-1 text-xs text-slate-600">Use a fully qualified domain. HTTPS URLs are normalized automatically.</p>
            </div>
            <button type="button" onClick={() => setShowForm(false)} className="text-slate-600 hover:text-white"><X size={18} /></button>
          </div>
          <div className="grid gap-4 md:grid-cols-[1fr_120px_1fr]">
            <div>
              <label className="label">Domain</label>
              <input className="input" placeholder="api.example.com" value={form.domain} onChange={(event) => setForm({ ...form, domain: event.target.value })} required />
            </div>
            <div>
              <label className="label">TLS port</label>
              <input className="input" type="number" min={1} max={65535} value={form.port} onChange={(event) => setForm({ ...form, port: Number(event.target.value) })} required />
            </div>
            <div>
              <label className="label">Label <span className="font-normal normal-case tracking-normal">(optional)</span></label>
              <input className="input" placeholder="Production API" value={form.label} onChange={(event) => setForm({ ...form, label: event.target.value })} />
            </div>
          </div>
          <div className="mt-5 flex flex-wrap items-center justify-between gap-4">
            <label className="flex cursor-pointer items-center gap-3 text-xs text-slate-400">
              <input type="checkbox" className="h-4 w-4 accent-lime-400" checked={form.scan} onChange={(event) => setForm({ ...form, scan: event.target.checked })} />
              Queue the first scan immediately
            </label>
            <button className="btn-primary" disabled={submitting}>{submitting ? 'Adding…' : 'Add to inventory'}</button>
          </div>
        </form>
      )}

      {assets.length === 0 ? (
        <EmptyState icon={Radar} title="No managed assets" message="Add a domain to begin collecting live certificate and TLS posture evidence." action={<button className="btn-primary" onClick={() => setShowForm(true)}>Add first asset</button>} />
      ) : (
        <>
          <div className="mb-4 flex items-center gap-3 rounded-xl border border-white/[0.06] bg-ink-850/60 px-4 py-2.5">
            <Search size={16} className="text-slate-600" />
            <input className="w-full bg-transparent text-sm text-white outline-none placeholder:text-slate-700" placeholder="Filter domains or labels…" value={query} onChange={(event) => setQuery(event.target.value)} />
            <span className="text-xs text-slate-700">{filtered.length} assets</span>
          </div>
          <div className="table-wrap">
            <table className="data-table">
              <thead><tr><th>Asset</th><th>Status</th><th>Risk</th><th>PQC readiness</th><th>Last scan</th><th /></tr></thead>
              <tbody>
                {filtered.map((asset) => (
                  <tr key={asset.id}>
                    <td>
                      <Link to={`/assets/${asset.id}`} className="group">
                        <div className="font-medium text-slate-200 group-hover:text-signal">{asset.label || asset.domain}</div>
                        <div className="mt-1 font-mono text-[11px] text-slate-600">{asset.domain}:{asset.port}</div>
                      </Link>
                    </td>
                    <td><Badge value={asset.status} /></td>
                    <td>
                      <div className="flex items-center gap-3">
                        <span className="w-6 font-mono text-sm text-amber-300">{asset.current_risk_score}</span>
                        <div className="h-1.5 w-20 overflow-hidden rounded-full bg-slate-800"><div className="h-full bg-amber-400" style={{ width: `${asset.current_risk_score}%` }} /></div>
                      </div>
                    </td>
                    <td>
                      <div className="flex items-center gap-3">
                        <span className="w-6 font-mono text-sm text-cyan-350">{asset.current_pqc_score}</span>
                        <div className="h-1.5 w-20 overflow-hidden rounded-full bg-slate-800"><div className="h-full bg-cyan-350" style={{ width: `${asset.current_pqc_score}%` }} /></div>
                      </div>
                    </td>
                    <td className="text-xs">{relativeTime(asset.last_scanned_at)}</td>
                    <td>
                      <div className="flex justify-end gap-1">
                        <button title="Start scan" onClick={() => void scan(asset)} disabled={asset.status === 'queued' || asset.status === 'running'} className="rounded-lg p-2 text-slate-600 hover:bg-signal/10 hover:text-signal disabled:opacity-30"><Radar size={16} /></button>
                        <Link title="Open asset" to={`/assets/${asset.id}`} className="rounded-lg p-2 text-slate-600 hover:bg-white/5 hover:text-white"><ExternalLink size={16} /></Link>
                        <button title="Delete asset" onClick={() => void remove(asset)} className="rounded-lg p-2 text-slate-700 hover:bg-rose-400/10 hover:text-rose-300"><Trash2 size={16} /></button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </>
      )}
    </>
  )
}

