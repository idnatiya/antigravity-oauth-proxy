<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { LogOut, RefreshCw, Layers, Copy, Check, ChevronRight } from '@lucide/vue'
import { useAuthStore } from '@/stores/authStore'
import { useUsageStore } from '@/stores/usageStore'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const usageStore = useUsageStore()

const copiedEndpoint = ref(false)

const pageTitle = computed(() => {
  if (route.path.startsWith('/overview')) return 'Overview & Analytics'
  if (route.path.startsWith('/playground')) return 'API Playground & Tester'
  if (route.path.startsWith('/requests')) return 'Request Telemetry Logs'
  if (route.path.startsWith('/models')) return 'AI Model Catalog'
  if (route.path.startsWith('/accounts')) return 'Google Accounts'
  if (route.path.startsWith('/security')) return 'Security & Access Control'
  return 'Dashboard'
})

const proxyEndpointUrl = computed(() => {
  return `${window.location.origin}/v1`
})

async function copyEndpoint() {
  try {
    await navigator.clipboard.writeText(proxyEndpointUrl.value)
    copiedEndpoint.value = true
    setTimeout(() => {
      copiedEndpoint.value = false
    }, 2000)
  } catch (e) {
    console.error('Failed to copy', e)
  }
}

async function handleLogout() {
  await authStore.logout()
  router.push('/login')
}

async function refreshData() {
  if (route.path.startsWith('/overview')) {
    await usageStore.fetchStats()
  } else if (route.path.startsWith('/requests')) {
    await usageStore.fetchRequests()
  } else if (route.path.startsWith('/models')) {
    await usageStore.fetchModels()
  }
}
</script>

<template>
  <header class="h-16 border-b border-zinc-800/60 bg-[#0e0f12]/80 backdrop-blur-md px-6 md:px-8 flex items-center justify-between shrink-0 sticky top-0 z-20">
    <!-- Breadcrumb & Page Info -->
    <div class="flex items-center gap-3">
      <div class="flex items-center text-xs text-zinc-500 font-medium">
        <span>Control Plane</span>
        <ChevronRight class="h-3.5 w-3.5 mx-1.5 text-zinc-600" />
        <span class="text-zinc-200 font-semibold text-sm">{{ pageTitle }}</span>
      </div>

      <div
        v-if="usageStore.account?.project_id"
        class="hidden lg:flex items-center gap-1.5 px-2.5 py-1 rounded-md bg-zinc-900/80 border border-zinc-800/80 text-xs font-mono text-zinc-400"
      >
        <Layers class="h-3 w-3 text-blue-400" />
        <span class="text-zinc-500">Project:</span>
        <span class="text-blue-300 font-medium">{{ usageStore.account.project_id }}</span>
      </div>
    </div>

    <!-- Actions & Status Bar -->
    <div class="flex items-center gap-3">
      <!-- 1-Click Copy Proxy Endpoint Pill -->
      <button
        @click="copyEndpoint"
        class="hidden sm:inline-flex items-center gap-2 px-2.5 py-1.5 rounded-lg bg-zinc-900/90 border border-zinc-800/80 hover:border-zinc-700 text-xs font-mono transition-all text-zinc-300 hover:text-white group cursor-pointer shadow-sm"
        title="Click to copy OpenAI-compatible endpoint URL"
      >
        <span class="relative flex h-2 w-2">
          <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
          <span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
        </span>
        <span class="text-zinc-500 font-sans text-[11px]">Proxy:</span>
        <span class="text-zinc-300 group-hover:text-blue-400">{{ proxyEndpointUrl }}</span>
        <span class="pl-1 border-l border-zinc-800 text-zinc-500 group-hover:text-zinc-300">
          <Check v-if="copiedEndpoint" class="h-3.5 w-3.5 text-emerald-400" />
          <Copy v-else class="h-3.5 w-3.5" />
        </span>
      </button>

      <!-- Refresh Data Button -->
      <Button
        variant="outline"
        size="icon"
        @click="refreshData"
        class="h-8 w-8 bg-zinc-900/80 border-zinc-800/80 hover:bg-zinc-800 text-zinc-400 hover:text-zinc-100"
        title="Refresh data"
      >
        <RefreshCw class="h-3.5 w-3.5" :class="{ 'animate-spin': usageStore.loadingStats || usageStore.loadingRequests }" />
      </Button>

      <Separator orientation="vertical" class="h-4 bg-zinc-800/80" />

      <!-- Sign Out Button -->
      <Button
        variant="destructive"
        size="sm"
        @click="handleLogout"
        class="h-8 px-3 text-xs bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/20 shadow-none gap-1.5"
      >
        <LogOut class="h-3.5 w-3.5" />
        <span class="hidden md:inline">Sign Out</span>
      </Button>
    </div>
  </header>
</template>
