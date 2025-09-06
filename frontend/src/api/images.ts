import axios from 'axios'
import type { ImageInfo } from '@/stores/upload'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
})

api.interceptors.request.use((config) => {
  const t = localStorage.getItem('token')
  if (t) {
    config.headers = config.headers || {}
    ;(config.headers as any).Authorization = `Bearer ${t}`
  }
  return config
})

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