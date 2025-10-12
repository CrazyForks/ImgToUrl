import axios from 'axios'
import { enableMock } from './mock'

export type StorageProvider = 'local' | 'r2' | 'oss'

export interface R2Config {
  account_id: string
  access_key_id: string
  secret_access_key: string
  bucket_name: string
  endpoint: string
  region: string
  public_url: string
}

export interface OssConfig {
  access_key_id: string
  secret_access_key: string
  bucket_name: string
  endpoint: string
  region: string
  public_url: string
}

export interface StorageConfigData {
  provider: StorageProvider
  r2: R2Config
  oss: OssConfig
}

export interface ApiResult<T = any> {
  success: boolean
  data?: T
  error?: string
}

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 20000,
  headers: { 'Content-Type': 'application/json' }
})

// 携带 token
api.interceptors.request.use((config) => {
  const t = localStorage.getItem('token')
  if (t) {
    config.headers = config.headers || {}
    ;(config.headers as any).Authorization = `Bearer ${t}`
  }
  return config
})

// 401 处理交给全局各处已存在的拦截器，这里仅保留 error 供 Mock 接手
api.interceptors.response.use(
  (res) => res,
  (err) => Promise.reject(err)
)

// 启用 Mock
enableMock(api)

// 获取存储配置（仅管理员可用）
export async function getStorageConfig(): Promise<ApiResult<StorageConfigData>> {
  try {
    const { data } = await api.get<ApiResult<StorageConfigData>>('/system/storage-config')
    return data
  } catch (e: any) {
    const enabled = String((import.meta as any)?.env?.VITE_ENABLE_MOCK ?? 'true') === 'true'
    if (enabled) {
      return {
        success: true,
        data: {
          provider: 'local',
          r2: {
            account_id: '',
            access_key_id: '',
            secret_access_key: '',
            bucket_name: '',
            endpoint: '',
            region: 'auto',
            public_url: ''
          },
          oss: {
            access_key_id: '',
            secret_access_key: '',
            bucket_name: '',
            endpoint: '',
            region: '',
            public_url: ''
          }
        }
      }
    }
    return { success: false, error: e?.message || '获取存储配置失败' }
  }
}

// 更新存储配置（仅管理员可用）
export async function updateStorageConfig(payload: StorageConfigData): Promise<ApiResult> {
  try {
    const { data } = await api.post<ApiResult>('/system/storage-config', payload)
    return data
  } catch (e: any) {
    const enabled = String((import.meta as any)?.env?.VITE_ENABLE_MOCK ?? 'true') === 'true'
    if (enabled) {
      return { success: true }
    }
    return { success: false, error: e?.message || '保存失败' }
  }
}