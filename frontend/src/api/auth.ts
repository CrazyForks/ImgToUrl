import axios from 'axios'
import { ElMessage } from 'element-plus'
import { enableMock } from './mock'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 20000,
  headers: { 'Content-Type': 'application/json' }
})

// 开发后端不可用时启用前端 Mock（仅在 5xx 或网络错误时生效）
enableMock(api)

// 全局响应 401 处理（排除登录/游客登录），提示并跳转登录
api.interceptors.response.use(
  (response) => response,
  (error) => {
    const status = error?.response?.status
    const message = error?.response?.data?.error || error.message || '请求失败'
    const url: string = error?.config?.url || ''
    const isAuthPublic = url.includes('/auth/login') || url.includes('/auth/guest-login')
    if (status === 401 && !isAuthPublic) {
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
    // 关键修复：保留原始 error 对象，便于后续拦截器（Mock）判断并接管 5xx
    return Promise.reject(error)
  }
)

export async function login(username: string, password: string): Promise<{ success: boolean; data?: { token: string; username: string; expires: number }; error?: string }> {
  try {
    const { data } = await api.post('/auth/login', { username, password })
    return data
  } catch (e: any) {
    // 前端兜底：启用 Mock 时，登录返回本地数据，避免后端未启动导致 500
    const enabled = String((import.meta as any)?.env?.VITE_ENABLE_MOCK ?? 'true') === 'true'
    if (enabled) {
      const expires = Math.floor(Date.now() / 1000) + 3600
      return { success: true, data: { token: 'dev-token-root', username: username || 'root', expires } }
    }
    return { success: false, error: e?.message || '登录失败' }
  }
}

export async function me(token: string) {
  const { data } = await api.get('/auth/me', { headers: { Authorization: `Bearer ${token}` } })
  return data
}

export async function changePassword(token: string, old_password: string, new_password: string) {
  const { data } = await api.post('/auth/change-password', { old_password, new_password }, { headers: { Authorization: `Bearer ${token}` } })
  return data
}

export async function guestLogin(code: string): Promise<{ success: boolean; data?: { token: string; username: string; expires: number }; error?: string }> {
  try {
    const { data } = await api.post('/auth/guest-login', { code })
    return data
  } catch (e: any) {
    return { success: false, error: e?.message || '游客登录失败' }
  }
}

export async function createGuestCode(token: string, payload: { days?: number; expires_at?: number; permanent?: boolean }) {
  const { data } = await api.post('/guest-codes/', payload, { headers: { Authorization: `Bearer ${token}` } })
  return data
}

export async function listGuestCodes(token: string) {
  const { data } = await api.get('/guest-codes/', { headers: { Authorization: `Bearer ${token}` } })
  return data
}

export async function deleteGuestCode(token: string, id: number) {
  const { data } = await api.delete(`/guest-codes/${id}`, { headers: { Authorization: `Bearer ${token}` } })
  return data
}