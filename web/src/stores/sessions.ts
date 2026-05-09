import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/services/api'

export interface Session {
  id: string
  projectId: string
  projectName: string
  status: 'active' | 'inactive'
  messageCount: number
  device: string
  createdAt: string
  updatedAt: string
}

export const useSessionsStore = defineStore('sessions', () => {
  const sessions = ref<Session[]>([])
  const activeSession = ref<Session | null>(null)
  const loading = ref(false)

  async function fetchSessions() {
    loading.value = true
    try {
      const response = await api.get('/sessions')
      sessions.value = response.data
    } finally {
      loading.value = false
    }
  }

  async function createSession(projectId: string) {
    const response = await api.post('/sessions', { project_id: projectId })
    sessions.value.unshift(response.data)
    activeSession.value = response.data
    return response.data
  }

  async function getSession(id: string) {
    const response = await api.get(`/sessions/${id}`)
    activeSession.value = response.data
    return response.data
  }

  return { sessions, activeSession, loading, fetchSessions, createSession, getSession }
})