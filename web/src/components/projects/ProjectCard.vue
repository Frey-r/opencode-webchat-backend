<template>
  <div class="bg-white rounded-xl border border-gray-200 p-4 flex items-center gap-3">
    <!-- Status stripe -->
    <div :class="['w-1 h-12 rounded-full', project.status === 'running' ? 'bg-green-500' : 'bg-gray-300']"></div>

    <!-- Content -->
    <div class="flex-1 min-w-0">
      <div class="flex items-center gap-2">
        <h3 class="font-semibold text-gray-900 truncate">{{ project.name }}</h3>
        <span v-if="project.status === 'running'" class="px-2 py-0.5 bg-green-100 text-green-700 text-xs font-medium rounded-full">Running</span>
      </div>
      <p class="text-xs text-gray-500 mt-0.5">{{ project.sessionCount }} sessions · {{ project.device }}</p>
      <p class="text-xs text-gray-400 mt-0.5">{{ timeAgo(project.updatedAt) }}</p>
    </div>

    <!-- Arrow -->
    <svg class="w-5 h-5 text-gray-400 flex-shrink-0" fill="none" viewBox="0 0 24 24" stroke="currentColor">
      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
    </svg>
  </div>
</template>

<script setup lang="ts">
import type { Project } from '@/stores/projects'

defineProps<{ project: Project }>()

function timeAgo(ts: string) {
  const diff = Date.now() - new Date(ts).getTime()
  const hours = Math.floor(diff / 3600000)
  return hours > 0 ? `${hours}h ago` : 'just now'
}
</script>