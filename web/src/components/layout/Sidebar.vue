<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink, useRoute } from 'vue-router'
import { LayoutDashboard, ListFilter, Cpu, ShieldCheck, Sparkles, Users, FlaskConical, CircleDot, KeyRound, Palette } from '@lucide/vue'
import { useAuthStore } from '@/stores/authStore'
import { useUsageStore } from '@/stores/usageStore'

const route = useRoute()
const authStore = useAuthStore()
const usageStore = useUsageStore()

const navItems = [
  { name: 'Overview', to: '/overview', icon: LayoutDashboard },
  { name: 'Playground', to: '/playground', icon: FlaskConical },
  { name: 'Image Studio', to: '/images', icon: Palette },
  { name: 'Requests', to: '/requests', icon: ListFilter },
  { name: 'Models', to: '/models', icon: Cpu },
  { name: 'Accounts', to: '/accounts', icon: Users },
  { name: 'API Keys', to: '/api-keys', icon: KeyRound },
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
  <aside class="w-64 border-r border-zinc-800/60 bg-[#0e0f12] flex flex-col shrink-0 select-none z-30">
    <!-- Brand Header -->
    <div class="h-16 px-5 border-b border-zinc-800/60 flex items-center gap-3">
      <div class="h-9 w-9 rounded-lg bg-gradient-to-br from-blue-500/20 to-indigo-500/10 border border-blue-500/30 flex items-center justify-center text-blue-400 shadow-[0_0_15px_-3px_rgba(59,130,246,0.3)]">
        <Sparkles class="h-4.5 w-4.5" />
      </div>
      <div>
        <div class="font-semibold text-sm tracking-tight text-white flex items-center gap-1.5">
          <span>Antigravity</span>
          <span class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono font-medium bg-blue-500/15 text-blue-400 border border-blue-500/30">
            PROXY
          </span>
        </div>
        <div class="text-xs text-zinc-500 font-mono flex items-center gap-1.5 mt-0.5">
          <span class="inline-block h-1.5 w-1.5 rounded-full bg-emerald-400"></span>
          <span>Control Plane</span>
        </div>
      </div>
    </div>

    <!-- Navigation -->
    <nav class="flex-1 px-3 py-4 space-y-1 overflow-y-auto">
      <RouterLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="group relative flex items-center gap-3 px-3.5 py-2.5 rounded-lg text-sm font-medium transition-all duration-150"
        :class="[
          route.path.startsWith(item.to)
            ? 'bg-blue-500/15 text-blue-300 font-semibold border border-blue-500/25 shadow-[inset_0_1px_0_0_rgba(255,255,255,0.05)]'
            : 'text-zinc-400 hover:text-zinc-100 hover:bg-zinc-800/40 border border-transparent'
        ]"
      >
        <span
          v-if="route.path.startsWith(item.to)"
          class="absolute left-0 top-1.5 bottom-1.5 w-1 rounded-r bg-blue-500 shadow-[0_0_8px_rgba(59,130,246,0.8)]"
        />
        <component
          :is="item.icon"
          class="h-4.5 w-4.5 shrink-0 transition-colors"
          :class="route.path.startsWith(item.to) ? 'text-blue-400' : 'text-zinc-500 group-hover:text-zinc-300'"
        />
        <span class="truncate">{{ item.name }}</span>
      </RouterLink>
    </nav>

    <!-- System Status Footer -->
    <div class="p-3 border-t border-zinc-800/60 bg-[#090a0c] space-y-2.5">
      <RouterLink
        to="/accounts"
        class="flex items-center justify-between p-2.5 rounded-lg bg-zinc-900/60 border border-zinc-800/60 hover:border-zinc-700/80 transition-all text-xs group"
      >
        <div class="flex items-center gap-2">
          <span class="relative flex h-2 w-2">
            <span
              v-if="usageStore.account?.accounts_ready"
              class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"
            />
            <span
              class="relative inline-flex rounded-full h-2 w-2"
              :class="usageStore.account?.accounts_ready ? 'bg-emerald-500' : 'bg-amber-500'"
            />
          </span>
          <span class="text-zinc-400 group-hover:text-zinc-200 font-medium">OAuth Pool</span>
        </div>
        <span class="font-mono text-zinc-300 font-semibold group-hover:text-white">{{ accountsText }}</span>
      </RouterLink>

      <div class="flex items-center justify-between px-2.5 py-1 text-xs text-zinc-500">
        <span class="flex items-center gap-1.5 truncate">
          <CircleDot class="h-3.5 w-3.5 text-zinc-600" />
          <span>{{ authStore.user?.username || 'admin' }}</span>
        </span>
        <span class="font-mono text-[11px] text-zinc-600 uppercase">v1beta</span>
      </div>
    </div>
  </aside>
</template>
