<template>
  <div class="flex flex-col h-screen">
    <AppHeader />

    <div class="flex-1 overflow-y-auto px-4 py-4">
      <!-- Breadcrumb -->
      <div class="flex items-center gap-2 text-sm text-gray-500 mb-4">
        <span>Repository</span>
        <span>/</span>
        <span class="text-gray-900 font-medium">{{ projectName }}</span>
        <span>/</span>
        <span class="text-gray-900">{{ fileName }}</span>
      </div>

      <!-- Doc Card -->
      <div class="bg-white rounded-xl border border-gray-200 p-6">
        <MarkdownRenderer :content="content" />
      </div>
    </div>

    <BottomNav />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import AppHeader from '@/components/layout/AppHeader.vue'
import BottomNav from '@/components/layout/BottomNav.vue'
import MarkdownRenderer from '@/components/docs/MarkdownRenderer.vue'

const projectName = ref('project-name')
const fileName = ref('README.md')
const content = ref(`# Project Name

## Overview
This is a sample README file displayed in the OpenCode WebChat application.

## Installation
\`\`\`bash
npm install
\`\`\`

## Usage
\`\`\`javascript
const opencode = require('opencode-webchat')
console.log('Hello, OpenCode!')
\`\`\`

## Configuration
\`\`\`json
{
  "name": "opencode-webchat",
  "version": "1.0.0"
}
\`\`\`
`)
</script>