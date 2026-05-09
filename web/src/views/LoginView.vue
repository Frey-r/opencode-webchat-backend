<template>
  <div class="min-h-screen bg-[var(--color-bg-canvas)] flex items-center justify-center p-4">
    <div class="w-full max-w-sm">
      <!-- Logo -->
      <div class="flex flex-col items-center mb-8">
        <div class="w-16 h-16 border-2 border-gray-300 rounded-lg flex items-center justify-center mb-4">
          <svg class="w-10 h-10" viewBox="0 0 40 40" fill="none">
            <rect x="4" y="6" width="32" height="22" rx="2" stroke="currentColor" stroke-width="2"/>
            <text x="8" y="20" font-family="monospace" font-size="10" fill="currentColor">&gt;_</text>
          </svg>
        </div>
        <h1 class="text-2xl font-bold text-gray-900">OpenCode</h1>
        <p class="text-sm text-gray-500 mt-1">WebChat Authentication</p>
      </div>

      <!-- Form Card -->
      <div class="bg-white rounded-xl border border-gray-200 p-6">
        <form @submit.prevent="handleSubmit" class="space-y-4">
          <!-- Node ID -->
          <div>
            <label class="block text-xs font-bold text-gray-900 mb-1.5">Node ID / Username</label>
            <div class="relative">
              <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"/>
                </svg>
              </span>
              <input
                v-model="nodeId"
                type="text"
                placeholder="Enter your identifier"
                class="w-full pl-10 pr-4 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
            </div>
          </div>

          <!-- Access Token -->
          <div>
            <div class="flex justify-between items-center mb-1.5">
              <label class="block text-xs font-bold text-gray-900">Access Token</label>
              <button type="button" class="text-xs text-gray-500 hover:text-gray-700">Reset Token?</button>
            </div>
            <div class="relative">
              <span class="absolute left-3 top-1/2 -translate-y-1/2 text-gray-400">
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 7a2 2 0 012 2m4 0a6 6 0 01-7.743 5.743L11 17H9v2H7v2H4a1 1 0 01-1-1v-2.586a1 1 0 01.293-.707l5.964-5.964A6 6 0 1121 9z"/>
                </svg>
              </span>
              <input
                v-model="accessToken"
                :type="showToken ? 'text' : 'password'"
                placeholder="••••••••••••••••"
                class="w-full pl-10 pr-10 py-2.5 border border-gray-200 rounded-lg text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
              <button
                type="button"
                @click="showToken = !showToken"
                class="absolute right-3 top-1/2 -translate-y-1/2 text-gray-400 hover:text-gray-600"
              >
                <svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path v-if="!showToken" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"/>
                  <path v-if="!showToken" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"/>
                  <path v-if="showToken" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.875 18.825A10.05 10.05 0 0112 19c-4.478 0-8.268-2.943-9.543-7a9.97 9.97 0 011.563-3.029m5.858.908a3 3 0 114.243 4.243M9.878 9.878l4.242 4.242M9.88 9.88l-3.29-3.29m7.532 7.532l3.29 3.29M3 3l3.59 3.59m0 0A9.953 9.953 0 0112 5c4.478 0 8.268 2.943 9.543 7a10.025 10.025 0 01-4.132 5.411m0 0L21 21"/>
                </svg>
              </button>
            </div>
          </div>

          <!-- Error -->
          <p v-if="error" class="text-xs text-red-500">{{ error }}</p>

          <!-- Submit -->
          <button
            type="submit"
            :disabled="loading"
            class="w-full py-3 bg-gray-900 text-white text-sm font-bold rounded-lg hover:bg-gray-800 disabled:opacity-50 transition-colors"
          >
            {{ loading ? 'AUTHENTICATING...' : 'AUTHENTICATE' }}
          </button>
        </form>
      </div>

      <!-- Security indicator -->
      <div class="flex items-center justify-center mt-6 text-xs text-gray-400">
        <svg class="w-3 h-3 mr-1.5" fill="currentColor" viewBox="0 0 20 20">
          <path fill-rule="evenodd" d="M2.166 4.999A11.042 11.042 0 0010 1.944 11.042 11.042 0 0017.834 5c.11.65.166 1.32.166 2.001 0 5.225-3.34 9.67-8 11.317C5.34 16.67 2 12.225 2 7c0-.682.057-1.35.166-2.001zm11.541 3.708a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd"/>
        </svg>
        Secured with end-to-end encryption
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const authStore = useAuthStore()

const nodeId = ref('')
const accessToken = ref('')
const showToken = ref(false)
const loading = ref(false)
const error = ref('')

async function handleSubmit() {
  error.value = ''
  loading.value = true
  try {
    await authStore.login(nodeId.value, accessToken.value)
    router.push('/chat')
  } catch (e: any) {
    error.value = e.response?.data?.error || 'Authentication failed'
  } finally {
    loading.value = false
  }
}
</script>