import { create } from 'zustand'

interface User {
  id: string
  email: string
  name: string
  role: 'user' | 'admin'
}

interface AuthState {
  user: User | null
  isLoading: boolean
  login: (email: string, password: string) => Promise<void>
  logout: () => void
  initAuth: () => void
}

export const useAuthStore = create<AuthState>((set: any) => ({
  user: null,
  isLoading: false,
  
  login: async (email: string, password: string) => {
    set({ isLoading: true })
    try {
      // TODO: 实现实际的登录逻辑
      console.log('Login attempt:', { email, password })
      // 临时模拟用户
      set({ 
        user: { 
          id: '1', 
          email, 
          name: 'Test User', 
          role: 'user' 
        }, 
        isLoading: false 
      })
    } catch (error) {
      console.error('Login failed:', error)
      set({ isLoading: false })
      throw error
    }
  },
  
  logout: () => {
    set({ user: null })
  },
  
  initAuth: () => {
    // TODO: 从 localStorage 或 API 恢复用户状态
    console.log('Initializing auth')
  }
}))