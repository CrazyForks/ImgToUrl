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