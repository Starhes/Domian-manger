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
  timeout: 30000, // 增加超时时间到30秒
  
  // 增加重试相关配置
  headers: {
    'Content-Type': 'application/json',
  },
})

// 请求重试次数
const MAX_RETRIES = 3
const RETRY_DELAY = 1000 // 1秒

// 指数退避重试函数
const exponentialBackoff = (retryCount: number): Promise<void> => {
  const delay = RETRY_DELAY * Math.pow(2, retryCount)
  return new Promise(resolve => setTimeout(resolve, delay))
}

// 判断是否应该重试
const shouldRetry = (error: any): boolean => {
  if (!error.response) {
    // 网络错误，应该重试
    return true
  }
  
  const status = error.response.status
  // 5xx 错误或 429 (Too Many Requests) 应该重试
  return status >= 500 || status === 429
}

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    // 添加时间戳避免缓存
    config.params = {
      ...config.params,
      _t: Date.now()
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器
api.interceptors.response.use(
  (response) => {
    // 如果响应中有 data 字段，直接返回 data 部分，简化前端调用
    if (response.data && response.data.success && response.data.data) {
      return { ...response, data: response.data.data }
    }
    return response
  },
  async (error) => {
    const originalRequest = error.config
    
    // 如果请求没有重试次数标记，初始化为0
    if (!originalRequest._retryCount) {
      originalRequest._retryCount = 0
    }
    
    // 如果可以重试且未超过最大重试次数
    if (shouldRetry(error) && originalRequest._retryCount < MAX_RETRIES) {
      originalRequest._retryCount++
      
      // 等待指数退避时间
      await exponentialBackoff(originalRequest._retryCount - 1)
      
      // 重试请求
      return api(originalRequest)
    }
    
    // 处理各种错误状态
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
        case 429:
          message.error('请求过于频繁，请稍后重试')
          break
        case 500:
          message.error('服务器内部错误')
          break
        case 502:
          message.error('网关错误')
          break
        case 503:
          message.error('服务暂时不可用')
          break
        case 504:
          message.error('网关超时')
          break
        default:
          message.error(data?.error || `请求失败 (${status})`)
      }
    } else if (error.request) {
      // 网络错误
      if (error.code === 'ECONNABORTED') {
        message.error('请求超时，请检查网络连接')
      } else {
        message.error('网络错误，请检查网络连接')
      }
    } else {
      message.error('请求失败')
    }
    
    return Promise.reject(error)
  }
)

export default api
