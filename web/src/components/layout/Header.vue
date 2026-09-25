<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { LogOut, RefreshCw, Layers } from '@lucide/vue'
import { useAuthStore } from '@/stores/authStore'
import { useUsageStore } from '@/stores/usageStore'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const usageStore = useUsageStore()

const pageTitle = computed(() => {
  if (route.path.startsWith('/overview')) return 'Overview & Analytics'
  if (route.path.startsWith('/playground')) return 'API Playground & Tester'
  if (route.path.startsWith('/requests')) return 'Request Telemetry Logs'
  if (route.path.startsWith('/models')) return 'AI Model Catalog'
  if (route.path.startsWith('/accounts')) return 'Google Accounts'
  if (route.path.startsWith('/security')) return 'Security & Access Control'
  return 'Dashboard'
})

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
  <header class="h-16 border-b border-[#24262e] bg-[#141518]/90 backdrop-blur px-8 flex items-center justify-between shrink-0 sticky top-0 z-20">
    <div class="flex items-center gap-4">
      <h1 class="text-lg font-semibold text-white tracking-tight">{{ pageTitle }}</h1>

      <Badge
        v-if="usageStore.account?.project_id"
        variant="secondary"
        class="hidden sm:inline-flex items-center gap-1.5 font-mono text-zinc-300"
      >
        <Layers class="h-3 w-3 text-blue-400" />
        <span class="text-zinc-500">Project:</span>
        <span class="text-blue-300 font-medium">{{ usageStore.account.project_id }}</span>
      </Badge>
    </div>

    <div class="flex items-center gap-3">
      <Button
        variant="outline"
        size="icon"
        @click="refreshData"
        title="Refresh data"
      >
        <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': usageStore.loadingStats || usageStore.loadingRequests }" />
      </Button>

      <Separator orientation="vertical" class="h-4" />

      <Button
        variant="destructive"
        size="sm"
        @click="handleLogout"
        class="bg-red-500/10 hover:bg-red-500/20 text-red-400 border border-red-500/20 shadow-none gap-2"
      >
        <LogOut class="h-3.5 w-3.5" />
        <span>Sign Out</span>
      </Button>
    </div>
  </header>
</template>
