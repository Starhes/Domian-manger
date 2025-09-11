import { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '../stores/authStore'

interface ProtectedRouteProps {
  children: ReactNode
  adminOnly?: boolean
}

const ProtectedRoute = ({ children, adminOnly = false }: ProtectedRouteProps) => {
  const { user, setRedirectPath } = useAuthStore()
  const location = useLocation()

  // 如果用户未登录
  if (!user) {
    // 保存当前路径，登录后重定向回来
    setRedirectPath(location.pathname + location.search)
    return <Navigate to="/login" replace />
  }

  // 如果需要管理员权限但用户不是管理员
  if (adminOnly && !user.is_admin) {
    return <Navigate to="/" replace />
  }

  // 如果用户未激活（但管理员账号可以跳过激活）
  if (!user.is_active && !user.is_admin) {
    return (
      <div style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        padding: '0 16px'
      }}>
        <div style={{
          background: 'white',
          borderRadius: '8px',
          padding: '40px',
          maxWidth: '500px',
          textAlign: 'center',
          boxShadow: '0 4px 12px rgba(0,0,0,0.15)'
        }}>
          <div style={{ fontSize: '48px', marginBottom: '20px' }}>📧</div>
          <h2 style={{ color: '#ff4d4f', marginBottom: '16px' }}>账户未激活</h2>
          <p style={{ color: '#666', marginBottom: '24px', lineHeight: 1.6 }}>
            您的账户尚未激活，请检查您的邮箱并点击激活链接。
          </p>
          <p style={{ color: '#999', fontSize: '14px', marginBottom: '24px' }}>
            没有收到邮件？请检查垃圾邮件文件夹，或联系管理员重新发送激活邮件。
          </p>
          <button 
            onClick={() => {
              useAuthStore.getState().logout()
            }}
            style={{
              background: '#1890ff',
              color: 'white',
              border: 'none',
              padding: '12px 24px',
              borderRadius: '6px',
              cursor: 'pointer',
              fontSize: '16px'
            }}
          >
            重新登录
          </button>
        </div>
      </div>
    )
  }

  return <>{children}</>
}

export default ProtectedRoute
