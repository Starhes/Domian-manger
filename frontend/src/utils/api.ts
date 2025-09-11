import axios from 'axios'
import { message } from 'antd'

// 根据环境自动确定API基础URL
const getBaseURL = (): string => {
  // 如果是开发环境
  if (import.meta.env.DEV) {
    return import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080'
  }
  
  // 生产环境：使用当前域名
  const protocol = window.location.protocol
  const host = window.location.host
  
  // 如果有自定义的API URL环境变量，使用它
  if (import.meta.env.VITE_API_BASE_URL) {
    return import.meta.env.VITE_API_BASE_URL
  }
  
  // 否则使用当前域名
  return `${protocol}//${host}`
}

const api = axios.create({
  baseURL: getBaseURL(),
  timeout: 10000,
})

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    if (error.response) {
      const { status, data } = error.response
      
      switch (status) {
        case 401:
          message.error('未授权，请重新登录')
          // 清除本地存储的认证信息
          localStorage.removeItem('auth-storage')
          window.location.href = '/login'
          break
        case 403:
          message.error('权限不足')
          break
        case 404:
          message.error('请求的资源不存在')
          break
        case 500:
          message.error('服务器内部错误')
          break
        default:
          message.error(data?.error || '请求失败')
      }
    } else if (error.request) {
      message.error('网络错误，请检查网络连接')
    } else {
      message.error('请求失败')
    }
    
    return Promise.reject(error)
  }
)

export default api
