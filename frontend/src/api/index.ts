import axios from 'axios'
import type { AxiosInstance, InternalAxiosRequestConfig, AxiosError } from 'axios'
import { useAuthStore } from '@/stores/auth'
import { useMessage } from 'naive-ui'

// 创建 axios 实例
const apiClient: AxiosInstance = axios.create({
  baseURL: 'http://localhost:8080',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器：自动注入 JWT
apiClient.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error: AxiosError) => {
    return Promise.reject(error)
  }
)

// 响应拦截器：处理 401
apiClient.interceptors.response.use(
  (response) => response,
  (error: AxiosError) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    
    // 错误提示
    const message = error.response?.data 
      ? (error.response.data as any).message || '请求失败' 
      : error.message || '网络错误'
    
    // 尝试显示 naive-ui message
    try {
      const messageHook = useMessage()
      messageHook.error(message)
    } catch {
      console.error('Message API error:', message)
    }
    
    return Promise.reject(error)
  }
)

export default apiClient