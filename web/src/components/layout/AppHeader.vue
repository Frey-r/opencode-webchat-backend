<template>
  <header class="bg-white border-b border-gray-200 px-4 py-3">
    <div class="flex items-center justify-between">
      <!-- Left: Menu -->
      <button @click="uiStore.toggleSidebar()" class="p-2 -ml-2 text-gray-600 hover:text-gray-900">
        <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h16"/>
        </svg>
      </button>

      <!-- Center: Logo -->
      <div class="flex items-center gap-2">
        <svg class="w-5 h-5 text-gray-900" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <rect x="3" y="5" width="18" height="14" rx="1" stroke-width="2"/>
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 10l3 3-3 3M12 13h4"/>
        </svg>
        <span class="font-bold text-gray-900">OpenCode</span>
      </div>

      <!-- Right: Settings -->
      <button class="p-2 -mr-2 text-gray-600 hover:text-gray-900">
        <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z"/>
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
        </svg>
      </button>
    </div>
  </header>
</template>

<script setup lang="ts">
import { useUiStore } from '@/stores/ui'
const uiStore = useUiStore()
</script>