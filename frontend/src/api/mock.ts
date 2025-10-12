import type { AxiosInstance, AxiosResponse, InternalAxiosRequestConfig } from 'axios'

/**
 * 开发环境后端不可用时的前端 Mock 适配器。
 * - 默认启用（VITE_ENABLE_MOCK 未设置时等于 'true'）
 * - 当响应报错且状态码为 5xx 或网络错误时，返回本地模拟数据
 * - 仅在调用端的 axios 实例上生效，不影响全局
 */
export function enableMock(api: AxiosInstance): void {
  api.interceptors.response.use(
    (response) => response,
    (error) => {
      const status = error?.response?.status as number | undefined
      const config = error?.config as InternalAxiosRequestConfig
      const message = error?.response?.data?.error || error?.message || '请求失败'

      // 是否开启 Mock（默认开启，生产可通过 VITE_ENABLE_MOCK=false 关闭）
      const enabled = String((import.meta as any)?.env?.VITE_ENABLE_MOCK ?? 'true') === 'true'

      // 非 5xx 或未开启 Mock：走原错误流程
      if (!enabled || (status !== undefined && status < 500)) {
        return Promise.reject(new Error(message))
      }

      // 生成模拟响应
      const mocked = mockAxiosResponse(config)
      if (mocked) {
        return Promise.resolve(mocked)
      }

      // 未匹配到 Mock 路由，仍抛原错误
      return Promise.reject(new Error(message))
    }
  )
}

/**
 * 根据请求 URL/Method 生成对应的模拟响应
 */
function mockAxiosResponse(config: InternalAxiosRequestConfig): AxiosResponse | null {
  // 使用 baseURL + url 组合完整路径，避免仅有相对路径时匹配失败（例如 '/auth/login'）
  const url = ((config.baseURL || '') + (config.url || '')).toLowerCase()
  const method = (config.method || 'get').toUpperCase()

  // 统一时间与示例数据
  const now = Date.now()
  const expires = Math.floor(now / 1000) + 3600
  const sampleImage = (i = 1) => ({
    id: i,
    uuid: `mock-uuid-${i}`,
    original_name: `sample-${i}.jpg`,
    file_size: 123456 * i,
    mime_type: 'image/jpeg',
    width: 1024,
    height: 768,
    public_url: `/uploads/images/2025/01/01/sample-${i}.jpg`,
    created_at: new Date(now - i * 86400000).toISOString()
  })

  // Auth
  if (method === 'POST' && url.includes('/api/v1/auth/login')) {
    const data = {
      success: true,
      data: { token: 'dev-token-root', username: 'root', expires }
    }
    return buildResponse(config, data)
  }
  if (method === 'POST' && url.includes('/api/v1/auth/guest-login')) {
    const data = {
      success: true,
      data: { token: 'dev-token-guest', username: 'guest', expires }
    }
    return buildResponse(config, data)
  }
  if (method === 'GET' && url.includes('/api/v1/auth/me')) {
    const data = { success: true, data: { username: 'root' } }
    return buildResponse(config, data)
  }
  if (method === 'POST' && url.includes('/api/v1/auth/change-password')) {
    const data = { success: true }
    return buildResponse(config, data)
  }

  // Images list
  if (method === 'GET' && matchPath(url, '/api/v1/images')) {
    const { page, pageSize } = parseQuery(url)
    const items = [sampleImage(1), sampleImage(2), sampleImage(3)].slice(0, pageSize)
    const data = {
      success: true,
      data: {
        items,
        total: 3,
        page,
        page_size: pageSize
      }
    }
    return buildResponse(config, data)
  }

  // Image info
  if (method === 'GET' && url.match(/\/api\/v1\/images\/[a-z0-9\-]+/)) {
    const data = { success: true, data: sampleImage(1) }
    return buildResponse(config, data)
  }

  // Delete image
  if (method === 'DELETE' && url.match(/\/api\/v1\/images\/[a-z0-9\-]+/)) {
    const data = { success: true }
    return buildResponse(config, data)
  }

  // Upload single
  if (method === 'POST' && url.includes('/api/v1/images/upload')) {
    const data = { success: true, data: sampleImage(1) }
    return buildResponse(config, data)
  }

  // Batch upload
  if (method === 'POST' && url.includes('/api/v1/batch-upload')) {
    const results = [sampleImage(1), sampleImage(2)]
    const data = {
      success: true,
      data: {
        successful: results.length,
        failed: 0,
        results,
        errors: []
      }
    }
    return buildResponse(config, data)
  }

  // Stats summary
  if (method === 'GET' && url.includes('/api/v1/images/stats/summary')) {
    const data = {
      success: true,
      data: {
        total_images: 42,
        total_size: 987654321,
        today_images: 3
      }
    }
    return buildResponse(config, data)
  }

  // System status
  if (method === 'GET' && url.includes('/api/v1/system/status')) {
    const data = {
      success: true,
      data: {
        uptime_seconds: 12345,
        disk_total: 500 * 1024 * 1024 * 1024,
        disk_free: 200 * 1024 * 1024 * 1024,
        disk_used: 300 * 1024 * 1024 * 1024,
        total_images: 42,
        total_size: 987654321,
        today_images: 3,
        today_size: 1234567,
        average_file_size: 234567,
        max_file_size: 3456789,
        min_file_size: 1234,
        disk_used_percent: 60.0,
        mem_total: 16 * 1024 * 1024 * 1024,
        mem_used: 8 * 1024 * 1024 * 1024,
        mem_free: 8 * 1024 * 1024 * 1024,
        mem_used_percent: 50.0,
        cpu_usage: 12.3
      }
    }
    return buildResponse(config, data)
  }

  // System storage-config (GET)
  if (method === 'GET' && url.includes('/api/v1/system/storage-config')) {
    const data = {
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
    return buildResponse(config, data)
  }

  // System storage-config (POST)
  if (method === 'POST' && url.includes('/api/v1/system/storage-config')) {
    const data = { success: true }
    return buildResponse(config, data)
  }

  // Guest codes
  if (method === 'GET' && url.includes('/api/v1/guest-codes')) {
    const data = { success: true, data: [] }
    return buildResponse(config, data)
  }
  if (method === 'POST' && url.includes('/api/v1/guest-codes')) {
    const data = {
      success: true,
      data: { id: 1, code: 'MOCKCODE123', expires_at: null, created_by: 'root' }
    }
    return buildResponse(config, data)
  }
  if (method === 'DELETE' && url.match(/\/api\/v1\/guest-codes\/\d+/)) {
    const data = { success: true }
    return buildResponse(config, data)
  }

  // Health
  if (method === 'GET' && url.endsWith('/health')) {
    const data = { status: 'ok', timestamp: Math.floor(now / 1000), service: 'image-host' }
    return buildResponse(config, data)
  }

  return null
}

/** 构造 AxiosResponse */
function buildResponse(config: InternalAxiosRequestConfig, data: any): AxiosResponse<any> {
  return {
    data,
    status: 200,
    statusText: 'OK',
    headers: {},
    // 规避 headers 可空类型不兼容：强制断言
    config: config as any
  }
}

/** 简易匹配路径（忽略查询参数） */
function matchPath(url: string, pathPrefix: string): boolean {
  const u = url.split('?')[0]
  return u.startsWith(pathPrefix)
}

/** 解析查询参数（分页） */
function parseQuery(url: string): { page: number; pageSize: number } {
  try {
    const qs = url.split('?')[1] || ''
    const sp = new URLSearchParams(qs)
    const page = Math.max(1, parseInt(sp.get('page') || '1', 10))
    const pageSize = Math.min(100, Math.max(1, parseInt(sp.get('page_size') || '20', 10)))
    return { page, pageSize }
  } catch {
    return { page: 1, pageSize: 20 }
  }
}