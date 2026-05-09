<template>
  <div class="flex flex-col h-screen">
    <AppHeader />

    <div class="flex-1 overflow-hidden flex flex-col">
      <!-- Messages -->
      <div ref="messagesContainer" class="flex-1 overflow-y-auto px-4 py-4 space-y-4">
        <div v-if="messagesStore.messages.length === 0" class="flex flex-col items-center justify-center h-full text-gray-400">
          <svg class="w-12 h-12 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z"/>
          </svg>
          <p class="text-sm">Start a conversation with OpenCode</p>
        </div>

        <MessageItem
          v-for="msg in messagesStore.messages"
          :key="msg.id"
          :message="msg"
        />
      </div>

      <!-- Input Area -->
      <InputArea @send="sendMessage" :disabled="!wsConnected" />
    </div>

    <BottomNav />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, nextTick, watch } from 'vue'
import { useRoute } from 'vue-router'
import AppHeader from '@/components/layout/AppHeader.vue'
import BottomNav from '@/components/layout/BottomNav.vue'
import MessageItem from '@/components/chat/MessageItem.vue'
import InputArea from '@/components/chat/InputArea.vue'
import { useMessagesStore } from '@/stores/messages'
import { useSessionsStore } from '@/stores/sessions'
import { useWebSocket } from '@/composables/useWebSocket'

const route = useRoute()
const messagesStore = useMessagesStore()
const sessionsStore = useSessionsStore()
const { connected: wsConnected, connect, send, disconnect } = useWebSocket()

const messagesContainer = ref<HTMLElement>()

onMounted(async () => {
  const sessionId = route.params.sessionId as string
  if (sessionId) {
    await sessionsStore.getSession(sessionId)
    await messagesStore.fetchMessages(sessionId)
    connect(sessionId, handleWsMessage)
  }
})

onUnmounted(() => disconnect())

watch(() => messagesStore.messages.length, () => {
  nextTick(() => {
    messagesContainer.value?.scrollTo(0, messagesContainer.value.scrollHeight)
  })
})

function handleWsMessage(msg: any) {
  if (msg.type === 'token' || msg.type === 'assistant') {
    messagesStore.addMessage({
      id: Date.now().toString(),
      sessionId: route.params.sessionId as string,
      role: 'assistant',
      content: msg.payload || msg.content,
      timestamp: new Date().toISOString()
    })
  }
}

function sendMessage(text: string) {
  const sessionId = route.params.sessionId as string
  if (!sessionId) return
  messagesStore.addMessage({
    id: Date.now().toString(),
    sessionId,
    role: 'user',
    content: text,
    timestamp: new Date().toISOString()
  })
  send({ type: 'prompt', payload: text, sessionId })
}
</script>