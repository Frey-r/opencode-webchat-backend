import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUiStore = defineStore('ui', () => {
  const activeTab = ref<'sessions' | 'projects' | 'docs'>('sessions')
  const sidebarOpen = ref(false)
  const modalOpen = ref(false)

  function setActiveTab(tab: 'sessions' | 'projects' | 'docs') {
    activeTab.value = tab
  }

  function toggleSidebar() {
    sidebarOpen.value = !sidebarOpen.value
  }

  function openModal() {
    modalOpen.value = true
  }

  function closeModal() {
    modalOpen.value = false
  }

  return { activeTab, sidebarOpen, modalOpen, setActiveTab, toggleSidebar, openModal, closeModal }
})