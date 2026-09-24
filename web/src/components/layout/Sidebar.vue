<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { LayoutDashboard, ListFilter, Cpu, ShieldCheck, Radio, Sparkles, Users } from '@lucide/vue'
import { useAuthStore } from '@/stores/authStore'
import { useUsageStore } from '@/stores/usageStore'

const route = useRoute()
const authStore = useAuthStore()
const usageStore = useUsageStore()

const navItems = [
  { name: 'Overview', to: '/overview', icon: LayoutDashboard },
  { name: 'Requests', to: '/requests', icon: ListFilter },
  { name: 'Models', to: '/models', icon: Cpu },
  { name: 'Accounts', to: '/accounts', icon: Users },
  { name: 'Security', to: '/security', icon: ShieldCheck },
]

const tokenValidText = computed(() => {
  const sec = usageStore.account?.token_valid_seconds
  if (sec === undefined) return 'Active'
  if (sec <= 0) return 'Expired'
  const hours = Math.floor(sec / 3600)
  const mins = Math.floor((sec % 3600) / 60)
  return `${hours}h ${mins}m`
})
</script>

<template>
  <aside class="w-64 border-r border-[#24262e] bg-[#101114] flex flex-col shrink-0 select-none">
    <!-- Brand -->
    <div class="h-16 px-6 border-b border-[#24262e] flex items-center gap-3">
      <div class="h-8 w-8 rounded-lg bg-blue-600/20 border border-blue-500/30 flex items-center justify-center text-blue-400">
        <Sparkles class="h-4 w-4" />
      </div>
      <div>
        <div class="font-semibold text-sm tracking-tight text-white flex items-center gap-1.5">
          Antigravity
          <span class="text-[10px] uppercase font-mono px-1.5 py-0.5 rounded bg-blue-500/10 text-blue-400 border border-blue-500/20">Proxy</span>
        </div>
        <div class="text-[11px] text-zinc-500 font-mono">Control Plane</div>
      </div>
    </div>

    <!-- Navigation -->
    <nav class="flex-1 p-3 space-y-1">
      <RouterLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="flex items-center gap-3 px-3 py-2 rounded-md text-sm font-medium transition-colors"
        :class="[
          route.path.startsWith(item.to)
            ? 'bg-blue-600 text-white shadow-sm'
            : 'text-zinc-400 hover:text-zinc-200 hover:bg-[#1a1c22]'
        ]"
      >
        <component :is="item.icon" class="h-4 w-4 shrink-0" />
        <span>{{ item.name }}</span>
      </RouterLink>
    </nav>

    <!-- System Status Footer -->
    <div class="p-4 border-t border-[#24262e] bg-[#0d0e11] space-y-3">
      <div class="flex items-center justify-between text-xs text-zinc-400">
        <span class="flex items-center gap-1.5">
          <Radio class="h-3 w-3 text-emerald-400 animate-pulse" />
          <span>OAuth Session</span>
        </span>
        <span class="font-mono text-zinc-300 font-medium">{{ tokenValidText }}</span>
      </div>

      <div class="pt-2 border-t border-[#1e2027] flex items-center justify-between text-xs">
        <div class="truncate text-zinc-400">
          <span class="text-zinc-500">User: </span>
          <span class="font-medium text-zinc-200">{{ authStore.user?.username || 'admin' }}</span>
        </div>
      </div>
    </div>
  </aside>
</template>
