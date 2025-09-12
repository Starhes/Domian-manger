import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/auth-store'
import Layout from '@/components/layout'
import ProtectedRoute from '@/components/protected-route'
import Login from '@/pages/login'
import Register from '@/pages/register'
import ResetPassword from '@/pages/reset-password'
import ForgotPassword from '@/pages/forgot-password'
import Dashboard from '@/pages/dashboard'
import DNSRecords from '@/pages/dns-records'
import Profile from '@/pages/profile'
import AdminLayout from '@/components/admin-layout'
import AdminDashboard from '@/pages/admin/dashboard'
import AdminUsers from '@/pages/admin/users'
import AdminDomains from '@/pages/admin/domains'
import AdminProviders from '@/pages/admin/providers'
import AdminSMTPConfigs from '@/pages/admin/smtp-configs'
import { useEffect } from 'react'

function App() {
  const { user, initAuth } = useAuthStore()

  useEffect(() => {
    initAuth()
  }, [initAuth])

  return (
    <Routes>
      {/* 公共路由 */}
      <Route path="/login" element={user ? <Navigate to="/" replace /> : <Login />} />
      <Route path="/register" element={user ? <Navigate to="/" replace /> : <Register />} />
      <Route path="/forgot-password" element={<ForgotPassword />} />
      <Route path="/reset-password" element={<ResetPassword />} />
      
      {/* 用户路由 */}
      <Route path="/" element={
        <ProtectedRoute>
          <Layout />
        </ProtectedRoute>
      }>
        <Route index element={<Dashboard />} />
        <Route path="dns-records" element={<DNSRecords />} />
        <Route path="profile" element={<Profile />} />
      </Route>

      {/* 管理员路由 */}
      <Route path="/admin" element={
        <ProtectedRoute adminOnly>
          <AdminLayout />
        </ProtectedRoute>
      }>
        <Route index element={<AdminDashboard />} />
        <Route path="users" element={<AdminUsers />} />
        <Route path="domains" element={<AdminDomains />} />
        <Route path="providers" element={<AdminProviders />} />
        <Route path="smtp-configs" element={<AdminSMTPConfigs />} />
      </Route>

      {/* 404 */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  )
}

export default App