import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import api from '../utils/api'

interface User {
  id: number
  email: string
  nickname?: string
  is_active: boolean
  is_admin: boolean
  created_at: string
}

interface AuthState {
  user: User | null
  token: string | null
  isLoading: boolean
  redirectPath: string | null
  login: (email: string, password: string) => Promise<void>
  register: (email: string, password: string, confirmPassword: string, nickname?: string) => Promise<void>
  logout: () => void
  initAuth: () => void
  setRedirectPath: (path: string | null) => void
}

export const useAuthStore = create<AuthState>()(
  persist(
    (set, get) => ({
      user: null,
      token: null,
      isLoading: false,
      redirectPath: null,

      login: async (email: string, password: string) => {
        set({ isLoading: true })
        try {
          const response = await api.post('/api/login', { email, password })
          const { token, user } = response.data
          
          const { redirectPath } = get()
          set({ user, token, isLoading: false })
          
          // 设置默认的Authorization header
          api.defaults.headers.common['Authorization'] = `Bearer ${token}`
          
          // 登录成功后重定向到之前想访问的页面
          if (redirectPath) {
            set({ redirectPath: null })
            window.location.href = redirectPath
          } else {
            // 默认重定向到首页
            window.location.href = '/'
          }
        } catch (error) {
          set({ isLoading: false })
          throw error
        }
      },

      register: async (email: string, password: string, confirmPassword: string, nickname?: string) => {
        set({ isLoading: true })
        try {
          const registerData: any = { 
            email, 
            password, 
            confirm_password: confirmPassword 
          }
          
          if (nickname) {
            registerData.nickname = nickname
          }
          
          await api.post('/api/register', registerData)
          set({ isLoading: false })
        } catch (error) {
          set({ isLoading: false })
          throw error
        }
      },

      logout: () => {
        set({ user: null, token: null, redirectPath: null })
        delete api.defaults.headers.common['Authorization']
        // 重定向到登录页面
        window.location.href = '/login'
      },

      setRedirectPath: (path: string | null) => {
        set({ redirectPath: path })
      },

      initAuth: () => {
        const { token } = get()
        if (token) {
          api.defaults.headers.common['Authorization'] = `Bearer ${token}`
        }
      },
    }),
    {
      name: 'auth-storage',
      partialize: (state) => ({ 
        user: state.user, 
        token: state.token,
        redirectPath: state.redirectPath
      }),
    }
  )
)