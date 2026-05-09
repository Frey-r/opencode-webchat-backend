<template>
  <div class="flex flex-col h-screen">
    <AppHeader />

    <div class="flex-1 overflow-y-auto px-4 py-4">
      <!-- Page Header -->
      <div class="mb-6">
        <h1 class="text-2xl font-bold text-gray-900">Projects</h1>
        <p class="text-sm text-gray-500 mt-1">Manage and monitor active workspaces.</p>
      </div>

      <!-- Init Button -->
      <button
        @click="uiStore.openModal()"
        class="w-full py-3 bg-gray-900 text-white text-sm font-bold rounded-lg mb-6"
      >
        + INIT_PROJECT
      </button>

      <!-- Projects List -->
      <div v-if="projectsStore.loading" class="flex justify-center py-8">
        <div class="w-6 h-6 border-2 border-gray-300 border-t-gray-900 rounded-full animate-spin"></div>
      </div>

      <div v-else class="space-y-3">
        <ProjectCard
          v-for="project in projectsStore.projects"
          :key="project.id"
          :project="project"
        />
      </div>

      <!-- Empty State -->
      <div v-if="!projectsStore.loading && projectsStore.projects.length === 0" class="flex flex-col items-center justify-center py-12 text-gray-400">
        <svg class="w-12 h-12 mb-3" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"/>
        </svg>
        <p class="text-sm">No projects yet. Initialize one to get started.</p>
      </div>
    </div>

    <!-- Init Project Modal -->
    <InitProjectModal v-if="uiStore.modalOpen" @close="uiStore.closeModal()" />

    <BottomNav />
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import BottomNav from '@/components/layout/BottomNav.vue'
import ProjectCard from '@/components/projects/ProjectCard.vue'
import InitProjectModal from '@/components/projects/InitProjectModal.vue'
import { useProjectsStore } from '@/stores/projects'
import { useUiStore } from '@/stores/ui'

const projectsStore = useProjectsStore()
const uiStore = useUiStore()

onMounted(() => projectsStore.fetchProjects())
</script>