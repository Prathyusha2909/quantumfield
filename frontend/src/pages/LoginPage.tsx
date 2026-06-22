import { ArrowRight, Eye, EyeOff } from 'lucide-react'
import { FormEvent, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { AuthLayout } from '../components/AuthLayout'
import { useAuth } from '../context/AuthContext'
import { errorMessage } from '../lib/api'

export default function LoginPage() {
  const [email, setEmail] = useState('demo@quantumfield.dev')
  const [password, setPassword] = useState('QuantumField123!')
  const [showPassword, setShowPassword] = useState(false)
  const [error, setError] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const { login } = useAuth()
  const navigate = useNavigate()

  async function submit(event: FormEvent) {
    event.preventDefault()
    setSubmitting(true)
    setError('')
    try {
      await login(email, password)
      navigate('/')
    } catch (requestError) {
      setError(errorMessage(requestError))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <AuthLayout title="Welcome back" subtitle="Sign in to inspect your cryptographic attack surface.">
      <form onSubmit={submit} className="space-y-5">
        {error && <div className="rounded-xl border border-rose-400/20 bg-rose-400/10 px-4 py-3 text-sm text-rose-300">{error}</div>}
        <div>
          <label className="label" htmlFor="email">Work email</label>
          <input id="email" className="input" type="email" value={email} onChange={(event) => setEmail(event.target.value)} required />
        </div>
        <div>
          <label className="label" htmlFor="password">Password</label>
          <div className="relative">
            <input id="password" className="input pr-11" type={showPassword ? 'text' : 'password'} value={password} onChange={(event) => setPassword(event.target.value)} required />
            <button type="button" onClick={() => setShowPassword((value) => !value)} className="absolute right-3 top-1/2 -translate-y-1/2 text-slate-600 hover:text-slate-300">
              {showPassword ? <EyeOff size={17} /> : <Eye size={17} />}
            </button>
          </div>
        </div>
        <button className="btn-primary w-full" disabled={submitting}>
          {submitting ? 'Authenticating…' : 'Enter workspace'} <ArrowRight size={16} />
        </button>
      </form>
      <div className="mt-6 rounded-xl border border-cyan-350/10 bg-cyan-350/5 px-4 py-3 text-xs leading-5 text-slate-500">
        Demo credentials are prefilled when <span className="font-mono text-cyan-350">SEED_DEMO=true</span>.
      </div>
      <p className="mt-7 text-center text-sm text-slate-600">
        New to QuantumField? <Link to="/register" className="font-medium text-signal hover:text-lime-300">Create an account</Link>
      </p>
    </AuthLayout>
  )
}

