import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/services/api'

export interface Project {
  id: string
  name: string
  path: string
  status: 'running' | 'stopped'
  sessionCount: number
  device: string
  createdAt: string
  updatedAt: string
}

export const useProjectsStore = defineStore('projects', () => {
  const projects = ref<Project[]>([])
  const loading = ref(false)

  async function fetchProjects() {
    loading.value = true
    try {
      const response = await api.get('/projects')
      projects.value = response.data
    } finally {
      loading.value = false
    }
  }

  async function createProject(name: string, path: string) {
    const response = await api.post('/projects', { name, path })
    projects.value.push(response.data)
    return response.data
  }

  async function deleteProject(id: string) {
    await api.delete(`/projects/${id}`)
    projects.value = projects.value.filter(p => p.id !== id)
  }

  return { projects, loading, fetchProjects, createProject, deleteProject }
})