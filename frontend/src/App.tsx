import { Navigate, Route, Routes } from 'react-router-dom'
import type { ReactNode } from 'react'
import { AppShell } from './components/AppShell'
import { LoadingScreen } from './components/ui'
import { useAuth } from './context/AuthContext'
import AssetDetailsPage from './pages/AssetDetailsPage'
import AssetsPage from './pages/AssetsPage'
import CertificatesPage from './pages/CertificatesPage'
import DashboardPage from './pages/DashboardPage'
import FindingsPage from './pages/FindingsPage'
import LoginPage from './pages/LoginPage'
import PQCPage from './pages/PQCPage'
import RegisterPage from './pages/RegisterPage'
import ReportsPage from './pages/ReportsPage'
import ScansPage from './pages/ScansPage'

function ProtectedLayout() {
  const { user, loading } = useAuth()
  if (loading) return <LoadingScreen />
  if (!user) return <Navigate to="/login" replace />
  return <AppShell />
}

function GuestRoute({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth()
  if (loading) return <LoadingScreen />
  if (user) return <Navigate to="/" replace />
  return children
}

export default function App() {
  return (
    <Routes>
      <Route path="/login" element={<GuestRoute><LoginPage /></GuestRoute>} />
      <Route path="/register" element={<GuestRoute><RegisterPage /></GuestRoute>} />
      <Route element={<ProtectedLayout />}>
        <Route index element={<DashboardPage />} />
        <Route path="assets" element={<AssetsPage />} />
        <Route path="assets/:id" element={<AssetDetailsPage />} />
        <Route path="scans" element={<ScansPage />} />
        <Route path="findings" element={<FindingsPage />} />
        <Route path="certificates" element={<CertificatesPage />} />
        <Route path="pqc" element={<PQCPage />} />
        <Route path="reports" element={<ReportsPage />} />
      </Route>
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}
