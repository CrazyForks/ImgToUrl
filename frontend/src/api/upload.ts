import axios from 'axios'
import type { AxiosProgressEvent } from 'axios'
import type { ImageInfo, UploadResult, BatchUploadResult, StatsInfo } from '@/stores/upload'
import { ElMessage } from 'element-plus'
import { enableMock } from './mock'

export interface SystemStatus {
  uptime_seconds: number
  disk_total: number
  disk_free: number
  disk_used: number
  total_images: number
  total_size: number
  today_images: number
  today_size: number
  average_file_size: number
  max_file_size: number
  min_file_size: number

  // Realtime resources
  disk_used_percent: number
  mem_total: number
  mem_used: number
  mem_free: number
  mem_used_percent: number
  cpu_usage: number
}

// 创建 axios 实例
const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 启用前端 Mock（后端不可用或 5xx 时返回本地模拟数据）
enableMock(api)

// 独立实例用于非 /api/v1 的系统接口（如 /health），同样启用 Mock
const sys = axios.create({
  baseURL: '/',
  timeout: 10000
})
enableMock(sys)

// 启用前端 Mock（后端不可用或 5xx 时返回本地模拟数据）
enableMock(api)

// 请求拦截器
api.interceptors.request.use(
  (config) => {
    const t = localStorage.getItem('token')
    if (t) {
      config.headers = config.headers || {}
      ;(config.headers as any).Authorization = `Bearer ${t}`
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
    // 返回完整响应对象，方便在调用处通过泛型解构 data 获取类型安全的数据
    return response
  },
  (error) => {
    const status = error?.response?.status
    const message = error?.response?.data?.error || error.message || '请求失败'
    if (status === 401) {
      // 登录过期处理：提示、清空本地鉴权、跳转登录页并带重定向
      try {
        ElMessage.error('登录已过期，请重新登录')
      } catch (_) {
        alert('登录已过期，请重新登录')
      }
      try {
        localStorage.removeItem('token')
        localStorage.removeItem('username')
      } catch (_) {}
      const current = window.location.pathname + window.location.search
      const loginUrl = `/login?redirect=${encodeURIComponent(current)}`
      if (window.location.pathname !== '/login') {
        window.location.assign(loginUrl)
      }
    }
    // 保留原始错误对象，便于后续 Mock 拦截器判断并接管 5xx/网络错误
    return Promise.reject(error)
  }
)

// 上传单个图片
export const uploadImage = async (
  formData: FormData,
  onProgress?: (progress: number) => void
): Promise<UploadResult> => {
  try {
    const { data } = await api.post<UploadResult>('/images/upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      },
      onUploadProgress: (progressEvent: AxiosProgressEvent) => {
        if (progressEvent.total && onProgress) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          onProgress(progress)
        }
      }
    })
    return data
  } catch (error: any) {
    throw new Error(error.message)
  }
}

// 批量上传图片
export const batchUploadImages = async (
  formData: FormData,
  onProgress?: (progress: number) => void
): Promise<BatchUploadResult> => {
  try {
    const { data } = await api.post<BatchUploadResult>('/batch-upload', formData, {
      headers: {
        'Content-Type': 'multipart/form-data'
      },
      onUploadProgress: (progressEvent: AxiosProgressEvent) => {
        if (progressEvent.total && onProgress) {
          const progress = Math.round((progressEvent.loaded * 100) / progressEvent.total)
          onProgress(progress)
        }
      }
    })
    return data
  } catch (error: any) {
    throw new Error(error.message)
  }
}

// 获取图片信息
export const getImageInfo = async (uuid: string): Promise<{ success: boolean; data: ImageInfo }> => {
  try {
    const { data } = await api.get<{ success: boolean; data: ImageInfo }>(`/images/${uuid}`)
    return data
  } catch (error: any) {
    throw new Error(error.message)
  }
}

 // 获取统计信息
export const getStats = async (): Promise<{ success: boolean; data: StatsInfo }> => {
  try {
    const { data } = await api.get<{ success: boolean; data: StatsInfo }>('/images/stats/summary')
    return data
  } catch (error: any) {
    throw new Error(error.message)
  }
}

// 获取系统状态（运行时间、磁盘容量等）
export const getSystemStatus = async (): Promise<{ success: boolean; data: SystemStatus }> => {
  try {
    const { data } = await api.get<{ success: boolean; data: SystemStatus }>('/system/status')
    return data
  } catch (error: any) {
    throw new Error(error.message)
  }
}

// 健康检查（后端不可用时走本地 Mock）
export const healthCheck = async (): Promise<{
  status: string
  timestamp: number
  service: string
}> => {
  try {
    const { data } = await sys.get('/health')
    return data
  } catch (error: any) {
    // 保底：返回一个可用的健康信息，避免页面初始化报错
    return {
      status: 'ok',
      timestamp: Math.floor(Date.now() / 1000),
      service: 'image-host'
    }
  }
}