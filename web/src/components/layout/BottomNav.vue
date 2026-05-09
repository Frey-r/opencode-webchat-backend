<template>
  <nav class="bg-white border-t border-gray-200 px-4 py-2">
    <div class="flex justify-around">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        @click="setTab(tab.id)"
        :class="[
          'flex flex-col items-center gap-1 px-4 py-1 rounded-lg text-xs font-medium transition-colors',
          uiStore.activeTab === tab.id
            ? 'text-blue-600 bg-blue-50'
            : 'text-gray-500 hover:text-gray-700'
        ]"
      >
        <component :is="tab.icon" class="w-5 h-5" />
        {{ tab.label }}
      </button>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { h } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUiStore } from '@/stores/ui'

const router = useRouter()
const route = useRoute()
const uiStore = useUiStore()

const SessionsIcon = () => h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', class: 'w-5 h-5' }, [
  h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z' })
])

const ProjectsIcon = () => h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', class: 'w-5 h-5' }, [
  h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z' })
])

const DocsIcon = () => h('svg', { fill: 'none', viewBox: '0 0 24 24', stroke: 'currentColor', class: 'w-5 h-5' }, [
  h('path', { 'stroke-linecap': 'round', 'stroke-linejoin': 'round', 'stroke-width': '2', d: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z' })
])

const tabs = [
  { id: 'sessions' as const, label: 'Sessions', icon: SessionsIcon },
  { id: 'projects' as const, label: 'Projects', icon: ProjectsIcon },
  { id: 'docs' as const, label: 'Docs', icon: DocsIcon }
]

function setTab(tab: 'sessions' | 'projects' | 'docs') {
  uiStore.setActiveTab(tab)
  if (tab === 'sessions') router.push('/chat')
  else if (tab === 'projects') router.push('/projects')
  else router.push('/docs')
}
</script>