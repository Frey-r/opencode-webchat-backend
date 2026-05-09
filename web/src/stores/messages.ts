import { defineStore } from 'pinia'
import { ref } from 'vue'
import api from '@/services/api'

export interface Message {
  id: string
  sessionId: string
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: string
}

export const useMessagesStore = defineStore('messages', () => {
  const messages = ref<Message[]>([])
  const loading = ref(false)

  async function fetchMessages(sessionId: string) {
    loading.value = true
    try {
      const response = await api.get(`/sessions/${sessionId}/messages`)
      messages.value = response.data
    } finally {
      loading.value = false
    }
  }

  function addMessage(message: Message) {
    messages.value.push(message)
  }

  function clearMessages() {
    messages.value = []
  }

  return { messages, loading, fetchMessages, addMessage, clearMessages }
})