<template>
  <div class="bg-white border-b border-gray-100 px-4 py-2 flex items-center gap-2">
    <span class="w-2 h-2 bg-green-500 rounded-full"></span>
    <span class="text-sm font-medium text-gray-700">{{ deviceName }}</span>
    <button class="ml-auto text-gray-400 hover:text-gray-600">
      <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
        <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
      </svg>
    </button>
  </div>
</template>

<script setup lang="ts">
defineProps<{ deviceName?: string }>()
</script>