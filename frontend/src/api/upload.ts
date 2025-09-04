import axios from 'axios'
import type { AxiosProgressEvent } from 'axios'
import type { ImageInfo, UploadResult, BatchUploadResult, StatsInfo } from '@/stores/upload'

// 创建 axios 实例
const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
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
    return response.data
  },
  (error) => {
    const message = error.response?.data?.error || error.message || '请求失败'
    return Promise.reject(new Error(message))
  }
)

// 上传单个图片
export const uploadImage = async (
  formData: FormData,
  onProgress?: (progress: number) => void
): Promise<UploadResult> => {
  try {
    const response = await api.post('/images/upload', formData, {
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
    return response
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
    const response = await api.post('/batch-upload', formData, {
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
    return response
  } catch (error: any) {
    throw new Error(error.message)
  }
}

// 获取图片信息
export const getImageInfo = async (uuid: string): Promise<{ success: boolean; data: ImageInfo }> => {
  try {
    const response = await api.get(`/images/${uuid}`)
    return response
  } catch (error: any) {
    throw new Error(error.message)
  }
}

// 获取统计信息
export const getStats = async (): Promise<{ success: boolean; data: StatsInfo }> => {
  try {
    const response = await api.get('/images/stats/summary')
    return response
  } catch (error: any) {
    throw new Error(error.message)
  }
}

// 健康检查
export const healthCheck = async (): Promise<{ status: string; timestamp: number; service: string }> => {
  try {
    const response = await axios.get('/health')
    return response.data
  } catch (error: any) {
    throw new Error(error.message)
  }
}