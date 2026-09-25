<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { LayoutDashboard, ListFilter, Cpu, ShieldCheck, Radio, Sparkles, Users, FlaskConical } from '@lucide/vue'
import { useAuthStore } from '@/stores/authStore'
import { useUsageStore } from '@/stores/usageStore'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'

const route = useRoute()
const authStore = useAuthStore()
const usageStore = useUsageStore()

const navItems = [
  { name: 'Overview', to: '/overview', icon: LayoutDashboard },
  { name: 'Playground', to: '/playground', icon: FlaskConical },
  { name: 'Requests', to: '/requests', icon: ListFilter },
  { name: 'Models', to: '/models', icon: Cpu },
  { name: 'Accounts', to: '/accounts', icon: Users },
  { name: 'Security', to: '/security', icon: ShieldCheck },
]

const accountsText = computed(() => {
  const acc = usageStore.account
  if (!acc) return '…'
  if (!acc.accounts_total) return 'None'
  return `${acc.accounts_ready}/${acc.accounts_total} ready`
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
          <Badge variant="outline" class="text-[10px] uppercase font-mono px-1.5 py-0.5 bg-blue-500/10 text-blue-400 border-blue-500/20">
            Proxy
          </Badge>
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
          <Radio
            class="h-3 w-3"
            :class="usageStore.account?.accounts_ready ? 'text-emerald-400 animate-pulse' : 'text-amber-400'"
          />
          <span>Google Accounts</span>
        </span>
        <RouterLink to="/accounts" class="font-mono text-zinc-300 font-medium hover:text-white">{{ accountsText }}</RouterLink>
      </div>

      <Separator class="bg-[#1e2027]" />

      <div class="flex items-center justify-between text-xs">
        <div class="truncate text-zinc-400">
          <span class="text-zinc-500">User: </span>
          <span class="font-medium text-zinc-200">{{ authStore.user?.username || 'admin' }}</span>
        </div>
      </div>
    </div>
  </aside>
</template>
