import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from 'react'
import api from '../lib/api'
import type { User } from '../types'

interface AuthContextValue {
  user: User | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (name: string, email: string, password: string) => Promise<void>
  logout: () => void
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem('quantumfield_token')
    if (!token) {
      setLoading(false)
      return
    }
    api.get<User>('/auth/me')
      .then(({ data }) => setUser(data))
      .catch(() => localStorage.removeItem('quantumfield_token'))
      .finally(() => setLoading(false))
  }, [])

  const value = useMemo<AuthContextValue>(() => ({
    user,
    loading,
    login: async (email, password) => {
      const { data } = await api.post<{ token: string; user: User }>('/auth/login', { email, password })
      localStorage.setItem('quantumfield_token', data.token)
      setUser(data.user)
    },
    register: async (name, email, password) => {
      const { data } = await api.post<{ token: string; user: User }>('/auth/register', { name, email, password })
      localStorage.setItem('quantumfield_token', data.token)
      setUser(data.user)
    },
    logout: () => {
      localStorage.removeItem('quantumfield_token')
      setUser(null)
    },
  }), [loading, user])

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}

export function useAuth() {
  const context = useContext(AuthContext)
  if (!context) throw new Error('useAuth must be used inside AuthProvider')
  return context
}

