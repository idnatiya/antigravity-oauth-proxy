<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import { Search, FlaskConical } from '@lucide/vue'
import { useUsageStore } from '@/stores/usageStore'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import { Skeleton } from '@/components/ui/skeleton'

const router = useRouter()
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

function testModel(modelId: string) {
  router.push({ path: '/playground', query: { model: modelId } })
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h2 class="text-base font-semibold text-white tracking-tight">AI Model Catalog</h2>
        <p class="text-xs text-zinc-400">
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
          class="bg-[#121316] border-zinc-800/80 pl-9 text-xs h-9 focus-visible:ring-blue-500"
        />
      </div>
    </div>

    <!-- Models Grid Skeleton -->
    <div v-if="usageStore.loadingModels" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <Skeleton v-for="i in 6" :key="i" class="h-44 rounded-xl bg-zinc-900/60 border border-zinc-800/60" />
    </div>

    <div v-else-if="filteredModels.length === 0" class="py-16 text-center text-zinc-500 text-xs font-mono">
      No models matching "{{ search }}".
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <Card
        v-for="m in filteredModels"
        :key="m.id"
        class="p-5 bg-[#121316] border-zinc-800/80 hover:border-zinc-700/80 transition-all flex flex-col justify-between space-y-4 group shadow-sm hover:shadow-lg hover:shadow-black/30"
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

          <h3 class="font-mono text-sm font-semibold text-white tracking-tight break-all group-hover:text-blue-300 transition-colors">
            {{ m.id }}
          </h3>
        </div>

        <div class="space-y-3">
          <Separator class="bg-zinc-800/60" />
          <div class="flex items-center justify-between text-xs text-zinc-400">
            <span class="text-zinc-500 text-[11px]">Provider: <span class="text-zinc-300 font-mono">{{ m.owned_by }}</span></span>
            <Button
              variant="outline"
              size="sm"
              @click="testModel(m.id)"
              class="h-7 px-2.5 text-xs bg-zinc-900/80 border-zinc-800 hover:bg-blue-600 hover:text-white hover:border-blue-500 gap-1.5 transition-all text-zinc-300"
            >
              <FlaskConical class="h-3 w-3" />
              <span>Test</span>
            </Button>
          </div>
        </div>
      </Card>
    </div>
  </div>
</template>
