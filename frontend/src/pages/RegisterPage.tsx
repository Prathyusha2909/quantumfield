import { ArrowRight } from 'lucide-react'
import { FormEvent, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { AuthLayout } from '../components/AuthLayout'
import { useAuth } from '../context/AuthContext'
import { errorMessage } from '../lib/api'

export default function RegisterPage() {
  const [form, setForm] = useState({ name: '', email: '', password: '' })
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const { register } = useAuth()
  const navigate = useNavigate()

  async function submit(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    setError('')
    try {
      await register(form.name, form.email, form.password)
      navigate('/')
    } catch (requestError) {
      setError(errorMessage(requestError))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthLayout title="Create your workspace" subtitle="Start mapping certificate risk and post-quantum dependencies.">
      <form onSubmit={submit} className="space-y-5">
        {error && <div className="rounded-xl border border-rose-400/20 bg-rose-400/10 px-4 py-3 text-sm text-rose-300">{error}</div>}
        <div>
          <label className="label" htmlFor="name">Full name</label>
          <input id="name" className="input" value={form.name} onChange={(event) => setForm({ ...form, name: event.target.value })} required minLength={2} />
        </div>
        <div>
          <label className="label" htmlFor="email">Work email</label>
          <input id="email" className="input" type="email" value={form.email} onChange={(event) => setForm({ ...form, email: event.target.value })} required />
        </div>
        <div>
          <label className="label" htmlFor="password">Password</label>
          <input id="password" className="input" type="password" value={form.password} onChange={(event) => setForm({ ...form, password: event.target.value })} required minLength={8} maxLength={72} />
          <p className="mt-2 text-[11px] text-slate-700">Use at least 8 characters. Passwords are bcrypt-hashed.</p>
        </div>
        <button className="btn-primary w-full" disabled={submitting}>
          {submitting ? 'Creating workspace…' : 'Create account'} <ArrowRight size={16} />
        </button>
      </form>
      <p className="mt-7 text-center text-sm text-slate-600">
        Already have an account? <Link to="/login" className="font-medium text-signal hover:text-lime-300">Sign in</Link>
      </p>
    </AuthLayout>
  )
}

