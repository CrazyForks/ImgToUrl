import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 20000,
  headers: { 'Content-Type': 'application/json' }
})

export async function login(username: string, password: string): Promise<{ success: boolean; data?: { token: string; username: string; expires: number }; error?: string }> {
  try {
    const { data } = await api.post('/auth/login', { username, password })
    return data
  } catch (e: any) {
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