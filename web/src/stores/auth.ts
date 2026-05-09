import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import api from '@/services/api'

export interface User {
  id: string
  nodeId: string
}

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('token'))

  const isAuthenticated = computed(() => !!token.value)

  async function login(nodeId: string, accessToken: string) {
    const response = await api.post('/auth/login', { node_id: nodeId, access_token: accessToken })
    token.value = response.data.token
    user.value = response.data.user
    localStorage.setItem('token', response.data.token)
    return response.data
  }

  async function register(nodeId: string, accessToken: string) {
    const response = await api.post('/auth/register', { node_id: nodeId, access_token: accessToken })
    token.value = response.data.token
    user.value = response.data.user
    localStorage.setItem('token', response.data.token)
    return response.data
  }

  function logout() {
    token.value = null
    user.value = null
    localStorage.removeItem('token')
  }

  return { user, token, isAuthenticated, login, register, logout }
})