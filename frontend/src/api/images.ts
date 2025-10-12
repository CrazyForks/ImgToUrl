import axios from 'axios'
import type { ImageInfo } from '@/stores/upload'
import { ElMessage } from 'element-plus'
import { enableMock } from './mock'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

// 启用前端 Mock（后端不可用或 5xx 时返回本地模拟数据）
enableMock(api)

api.interceptors.request.use((config) => {
  const t = localStorage.getItem('token')
  if (t) {
    config.headers = config.headers || {}
    ;(config.headers as any).Authorization = `Bearer ${t}`
  }
  return config
})

// 全局响应 401 处理：提示、清空鉴权、跳转登录并带重定向
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error?.response?.status
    const message = error?.response?.data?.error || error.message || '请求失败'
    if (status === 401) {
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
    // 非 401：保留原始错误对象，留给后续 Mock 拦截器处理（如 5xx）
    return Promise.reject(error)
  }
 )

 // 保留原始错误对象，便于 Mock 拦截器接管 5xx 等错误
 api.interceptors.response.use(
   (response) => response,
   (error) => {
     return Promise.reject(error)
   }
 )

export interface PagedImages {
  items: ImageInfo[]
  total: number
  page: number
  page_size: number
}

export async function listImages(page = 1, pageSize = 20): Promise<{ success: boolean; data?: PagedImages; error?: string }> {
  try {
    const { data } = await api.get<{ success: boolean; data: PagedImages }>(`/images`, {
      params: { page, page_size: pageSize },
    })
    return data
  } catch (e: any) {
    return { success: false, error: e?.message || '获取图片列表失败' }
  }
}

export async function deleteImage(uuid: string): Promise<{ success: boolean; error?: string }> {
  try {
    const { data } = await api.delete<{ success: boolean }>(`/images/${uuid}`)
    return data
  } catch (e: any) {
    return { success: false, error: e?.message || '删除失败' }
  }
}

/** 一键清空图片（仅管理员）
 * scope: 'all' | 'local' | 'r2' | 'oss'（默认 all）
 */
export async function clearImages(scope: 'all' | 'local' | 'r2' | 'oss' = 'all'): Promise<{ success: boolean; data?: any; error?: string }> {
  try {
    const { data } = await api.post<{ success: boolean; data?: any }>(`/images/clear`, { scope })
    return data
  } catch (e: any) {
    return { success: false, error: e?.message || '清空失败' }
  }
}
export async function syncImages(provider: 'r2' | 'oss', opts?: { prefix?: string; limit?: number; overwrite?: boolean }): Promise<{ success: boolean; data?: any; error?: string }> {
  try {
    const payload = {
      provider,
      prefix: opts?.prefix ?? '',
      limit: opts?.limit ?? 1000,
      overwrite: opts?.overwrite ?? false
    }
    const { data } = await api.post<{ success: boolean; data?: any }>(`/images/sync`, payload)
    return data
  } catch (e: any) {
    return { success: false, error: e?.message || '同步失败' }
  }
}