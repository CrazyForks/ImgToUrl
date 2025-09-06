import { defineStore } from 'pinia'
import { ref, computed } from 'vue'

export interface UserInfo {
  username: string
}

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const user = ref<UserInfo | null>(token.value ? { username: localStorage.getItem('username') || '' } : null)

  const isAuthenticated = computed(() => !!token.value)

  function setAuth(t: string, u: string) {
    token.value = t
    user.value = { username: u }
    localStorage.setItem('token', t)
    localStorage.setItem('username', u)
  }

  function clearAuth() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('username')
  }

  return { token, user, isAuthenticated, setAuth, clearAuth }
})