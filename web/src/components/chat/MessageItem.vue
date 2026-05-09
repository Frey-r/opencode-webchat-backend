<template>
  <div :class="['flex', message.role === 'user' ? 'justify-end' : 'justify-start']">
    <div :class="[
      'max-w-[80%] rounded-2xl px-4 py-2.5 text-sm',
      message.role === 'user'
        ? 'bg-gray-900 text-white'
        : 'bg-white border border-gray-200'
    ]">
      <p class="whitespace-pre-wrap">{{ message.content }}</p>
      <p v-if="message.role === 'assistant'" class="text-xs text-gray-400 mt-1">
        {{ formatTime(message.timestamp) }}
      </p>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { Message } from '@/stores/messages'

defineProps<{ message: Message }>()

function formatTime(ts: string) {
  return new Date(ts).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
</script>