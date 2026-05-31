import axios from 'axios'
import { ElMessage } from 'element-plus'
import type { ApiResponse } from '@/types/api'

const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api/v1',
  timeout: 30000,
  headers: { 'Content-Type': 'application/json' },
})

// 请求拦截器：注入 JWT
request.interceptors.request.use((config) => {
  const token = localStorage.getItem('huihua_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器：统一错误处理
request.interceptors.response.use(
  (response) => response.data,
  (error) => {
    if (error.response) {
      const { status, data } = error.response
      if (status === 401) {
        localStorage.removeItem('huihua_token')
        localStorage.removeItem('huihua_user')
        window.location.href = '/login'
        return Promise.reject(error)
      }
      if (status === 403) {
        ElMessage.error('权限不足，请联系管理员')
        return Promise.reject(error)
      }
      // 业务错误：兼容后端 { error: "..." } 和 { code, message } 格式
      const msg = (data as ApiResponse)?.message || (data as any)?.error || '请求失败'
      ElMessage.error(msg)
    } else {
      ElMessage.error('网络异常，请检查连接')
    }
    return Promise.reject(error)
  },
)

export default request
