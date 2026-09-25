<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Search } from '@lucide/vue'
import { useUsageStore } from '@/stores/usageStore'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'

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

function getTierBadge(id: string): { label: string; variant: 'default' | 'secondary' | 'warning' | 'success' | 'outline' } {
  if (id.includes('high')) return { label: 'High Reasoning', variant: 'default' }
  if (id.includes('medium')) return { label: 'Medium Reasoning', variant: 'default' }
  if (id.includes('low') || id.includes('flash-lite')) return { label: 'Low Reasoning (Fast)', variant: 'success' }
  if (id.includes('thinking')) return { label: 'Extended Thinking', variant: 'warning' }
  if (id.includes('agent')) return { label: 'Autonomous Agent', variant: 'default' }
  return { label: 'Standard', variant: 'secondary' }
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
        <Search class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500 pointer-events-none" />
        <Input
          v-model="search"
          type="text"
          placeholder="Search models..."
          class="bg-[#202227] border-[#2c2e36] pl-9 text-xs h-9"
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
      <Card
        v-for="m in filteredModels"
        :key="m.id"
        class="p-5 bg-[#202227] border-[#2c2e36] hover:border-[#383a45] transition-all flex flex-col justify-between space-y-4"
      >
        <div class="space-y-2">
          <div class="flex items-center justify-between">
            <span class="text-[10px] font-mono uppercase text-zinc-500 tracking-wider">
              {{ getFamily(m.id) }}
            </span>
            <Badge :variant="getTierBadge(m.id).variant" class="font-mono text-[10px]">
              {{ getTierBadge(m.id).label }}
            </Badge>
          </div>

          <h3 class="font-mono text-sm font-semibold text-white tracking-tight break-all">
            {{ m.id }}
          </h3>
        </div>

        <div class="space-y-3">
          <Separator class="bg-[#2a2d34]" />
          <div class="flex items-center justify-between text-xs text-zinc-400">
            <span class="text-zinc-500">Provider: <span class="text-zinc-300 font-mono">{{ m.owned_by }}</span></span>
            <span class="text-[11px] font-mono text-blue-400">OpenAI Compatible</span>
          </div>
        </div>
      </Card>
    </div>
  </div>
</template>
