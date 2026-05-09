<template>
  <div class="border-t border-gray-200 bg-white px-4 py-3">
    <div class="flex items-center gap-2">
      <input
        v-model="inputText"
        @keydown.enter="handleSend"
        type="text"
        placeholder="Send a message to OpenCode..."
        :disabled="disabled"
        class="flex-1 px-4 py-2.5 bg-gray-50 border border-gray-200 rounded-full text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
      />
      <button
        @click="handleSend"
        :disabled="!inputText.trim() || disabled"
        class="p-2.5 bg-gray-900 text-white rounded-full disabled:opacity-50 hover:bg-gray-800 transition-colors"
      >
        <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 19l9 2-9-18-9 18 9-2zm0 0v-8"/>
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const emit = defineEmits<{ send: [text: string] }>()
defineProps<{ disabled?: boolean }>()

const inputText = ref('')

function handleSend() {
  if (inputText.value.trim()) {
    emit('send', inputText.value)
    inputText.value = ''
  }
}
</script>