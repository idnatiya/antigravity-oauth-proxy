<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Search } from '@lucide/vue'
import { useUsageStore } from '@/stores/usageStore'

const usageStore = useUsageStore()
const search = ref('')

onMounted(() => {
  if (usageStore.models.length === 0) {
    usageStore.fetchModels()
  }
})

function getFamily(id: string): string {
  if (id.startsWith('claude')) return 'Anthropic Claude'
  if (id.startsWith('gemini')) return 'Google Gemini'
  if (id.startsWith('gpt-oss')) return 'Open-Weight OSS'
  return 'General'
}

function getTierBadge(id: string): { label: string; color: string } {
  if (id.includes('high')) return { label: 'High Reasoning', color: 'bg-purple-500/10 text-purple-400 border-purple-500/20' }
  if (id.includes('medium')) return { label: 'Medium Reasoning', color: 'bg-indigo-500/10 text-indigo-400 border-indigo-500/20' }
  if (id.includes('low') || id.includes('flash-lite')) return { label: 'Low Reasoning (Fast)', color: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' }
  if (id.includes('thinking')) return { label: 'Extended Thinking', color: 'bg-amber-500/10 text-amber-400 border-amber-500/20' }
  if (id.includes('agent')) return { label: 'Autonomous Agent', color: 'bg-blue-500/10 text-blue-400 border-blue-500/20' }
  return { label: 'Standard', color: 'bg-zinc-800 text-zinc-300 border-zinc-700' }
}

const filteredModels = computed(() => {
  const q = search.value.toLowerCase().trim()
  if (!q) return usageStore.models
  return usageStore.models.filter(m => m.id.toLowerCase().includes(q) || m.owned_by.toLowerCase().includes(q))
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h2 class="text-base font-semibold text-white tracking-tight">AI Model Catalog</h2>
        <p class="text-xs text-zinc-500">
          {{ usageStore.models.length }} models accessible via OpenAI-compatible API endpoint
        </p>
      </div>

      <!-- Search Input -->
      <div class="relative w-full sm:w-72">
        <Search class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
        <input
          v-model="search"
          type="text"
          placeholder="Search models..."
          class="w-full bg-[#202227] border border-[#2c2e36] focus:border-blue-500 rounded-lg pl-9 pr-3 py-1.5 text-xs text-zinc-200 placeholder:text-zinc-600 focus:outline-none transition-colors"
        />
      </div>
    </div>

    <!-- Models Grid -->
    <div v-if="usageStore.loadingModels" class="py-16 text-center text-zinc-500 text-xs font-mono animate-pulse">
      Loading models catalog...
    </div>

    <div v-else-if="filteredModels.length === 0" class="py-16 text-center text-zinc-500 text-xs font-mono">
      No models matching "{{ search }}".
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div
        v-for="m in filteredModels"
        :key="m.id"
        class="p-5 rounded-xl bg-[#202227] border border-[#2c2e36] hover:border-[#383a45] transition-all flex flex-col justify-between space-y-4"
      >
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-[10px] font-mono uppercase text-zinc-500 tracking-wider">
              {{ getFamily(m.id) }}
            </span>
            <span
              class="px-2 py-0.5 rounded-full border text-[10px] font-medium font-mono"
              :class="getTierBadge(m.id).color"
            >
              {{ getTierBadge(m.id).label }}
            </span>
          </div>

          <h3 class="font-mono text-sm font-semibold text-white tracking-tight break-all">
            {{ m.id }}
          </h3>
        </div>

        <div class="pt-3 border-t border-[#2a2d34] flex items-center justify-between text-xs text-zinc-400">
          <span class="text-zinc-500">Provider: <span class="text-zinc-300 font-mono">{{ m.owned_by }}</span></span>
          <span class="text-[11px] font-mono text-blue-400">OpenAI Compatible</span>
        </div>
      </div>
    </div>
  </div>
</template>
