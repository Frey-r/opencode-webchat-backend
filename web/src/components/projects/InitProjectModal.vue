<template>
  <div class="fixed inset-0 bg-black/50 flex items-end justify-center z-50" @click.self="$emit('close')">
    <div class="bg-white w-full max-w-lg rounded-t-2xl p-6">
      <div class="flex items-center justify-between mb-6">
        <h2 class="text-lg font-bold text-gray-900">Initialize Project</h2>
        <button @click="$emit('close')" class="p-2 text-gray-400 hover:text-gray-600">
          <svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
          </svg>
        </button>
      </div>

      <form @submit.prevent="handleSubmit" class="space-y-4">
        <div>
          <label class="block text-xs font-bold text-gray-900 mb-1.5">Project Name</label>
          <input v-model="name" type="text" placeholder="my-project" class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" required />
        </div>
        <div>
          <label class="block text-xs font-bold text-gray-900 mb-1.5">Project Path</label>
          <input v-model="path" type="text" placeholder="~/dev/my-project" class="w-full px-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500" required />
        </div>
        <p v-if="error" class="text-xs text-red-500">{{ error }}</p>
        <button type="submit" class="w-full py-3 bg-gray-900 text-white text-sm font-bold rounded-lg">CREATE PROJECT</button>
      </form>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useProjectsStore } from '@/stores/projects'
import { useRouter } from 'vue-router'

const emit = defineEmits<{ close: [] }>()
const projectsStore = useProjectsStore()
const router = useRouter()

const name = ref('')
const path = ref('')
const error = ref('')

async function handleSubmit() {
  error.value = ''
  try {
    const project = await projectsStore.createProject(name.value, path.value)
    emit('close')
    router.push(`/chat?project=${project.id}`)
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Failed to create project'
  }
}
</script>