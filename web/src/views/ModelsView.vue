<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import {
  Search,
  X,
  RotateCw,
  LayoutGrid,
  List,
  Sparkles,
  Bot,
  Layers,
  Cpu,
  Copy,
  Check,
  FlaskConical,
  Code2,
  SlidersHorizontal,
} from '@lucide/vue'
import { useUsageStore } from '@/stores/usageStore'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Table,
  TableHeader,
  TableBody,
  TableHead,
  TableRow,
  TableCell,
} from '@/components/ui/table'
import ModelCard from '@/components/models/ModelCard.vue'
import ModelCodeModal from '@/components/models/ModelCodeModal.vue'

const router = useRouter()
const usageStore = useUsageStore()

const search = ref('')
const activeCategory = ref<'all' | 'gemini' | 'claude' | 'oss'>('all')
const activeTier = ref<'all' | 'high' | 'fast' | 'thinking' | 'agent'>('all')
const viewMode = ref<'grid' | 'table'>('grid')

const codeModalOpen = ref(false)
const selectedModelId = ref<string | null>(null)
const copiedBaseUrl = ref(false)
const copiedRowId = ref<string | null>(null)

onMounted(() => {
  if (usageStore.models.length === 0) {
    usageStore.fetchModels()
  }
})

// Counts
const totalCount = computed(() => usageStore.models.length)
const geminiCount = computed(() => usageStore.models.filter(m => m.id.toLowerCase().startsWith('gemini')).length)
const claudeCount = computed(() => usageStore.models.filter(m => m.id.toLowerCase().startsWith('claude')).length)
const ossCount = computed(() =>
  usageStore.models.filter(m => !m.id.toLowerCase().startsWith('gemini') && !m.id.toLowerCase().startsWith('claude')).length
)

// Categories
const categories = computed(() => [
  { key: 'all' as const, label: 'All Models', count: totalCount.value, icon: Cpu },
  { key: 'gemini' as const, label: 'Google Gemini', count: geminiCount.value, icon: Sparkles, color: 'text-blue-400' },
  { key: 'claude' as const, label: 'Anthropic Claude', count: claudeCount.value, icon: Bot, color: 'text-amber-400' },
  ...(ossCount.value > 0
    ? [{ key: 'oss' as const, label: 'Open-Weight & Others', count: ossCount.value, icon: Layers, color: 'text-purple-400' }]
    : []),
])

// Tiers
const tiers = [
  { key: 'all' as const, label: 'All Tiers' },
  { key: 'high' as const, label: 'High Reasoning' },
  { key: 'fast' as const, label: 'Fast & Lite' },
  { key: 'thinking' as const, label: 'Extended Thinking' },
  { key: 'agent' as const, label: 'Autonomous Agent' },
]

// Filtered models
const filteredModels = computed(() => {
  let list = usageStore.models

  // Category filter
  if (activeCategory.value === 'gemini') {
    list = list.filter(m => m.id.toLowerCase().startsWith('gemini'))
  } else if (activeCategory.value === 'claude') {
    list = list.filter(m => m.id.toLowerCase().startsWith('claude'))
  } else if (activeCategory.value === 'oss') {
    list = list.filter(m => !m.id.toLowerCase().startsWith('gemini') && !m.id.toLowerCase().startsWith('claude'))
  }

  // Tier filter
  if (activeTier.value === 'high') {
    list = list.filter(m => {
      const id = m.id.toLowerCase()
      return id.includes('high') || id.includes('pro')
    })
  } else if (activeTier.value === 'fast') {
    list = list.filter(m => {
      const id = m.id.toLowerCase()
      return id.includes('flash-lite') || id.includes('low') || (id.includes('flash') && !id.includes('high'))
    })
  } else if (activeTier.value === 'thinking') {
    list = list.filter(m => m.id.toLowerCase().includes('thinking'))
  } else if (activeTier.value === 'agent') {
    list = list.filter(m => m.id.toLowerCase().includes('agent'))
  }

  // Search filter
  const q = search.value.toLowerCase().trim()
  if (q) {
    list = list.filter(m => m.id.toLowerCase().includes(q) || m.owned_by.toLowerCase().includes(q))
  }

  return list
})

const hasActiveFilters = computed(() => {
  return search.value.trim() !== '' || activeCategory.value !== 'all' || activeTier.value !== 'all'
})

function resetFilters() {
  search.value = ''
  activeCategory.value = 'all'
  activeTier.value = 'all'
}

function testModel(modelId: string) {
  router.push({ path: '/playground', query: { model: modelId } })
}

function openCodeSnippet(modelId: string) {
  selectedModelId.value = modelId
  codeModalOpen.value = true
}

async function copyBaseUrl() {
  const url = typeof window !== 'undefined' ? `${window.location.origin}/v1` : 'http://localhost:8080/v1'
  try {
    await navigator.clipboard.writeText(url)
    copiedBaseUrl.value = true
    setTimeout(() => {
      copiedBaseUrl.value = false
    }, 2000)
  } catch (err) {
    console.error('Failed to copy base URL', err)
  }
}

async function copyModelId(id: string) {
  try {
    await navigator.clipboard.writeText(id)
    copiedRowId.value = id
    setTimeout(() => {
      if (copiedRowId.value === id) copiedRowId.value = null
    }, 2000)
  } catch (err) {
    console.error('Failed to copy model ID', err)
  }
}

function getFamilyBadge(id: string) {
  const lower = id.toLowerCase()
  if (lower.startsWith('claude')) {
    return { name: 'Claude', class: 'bg-amber-500/10 text-amber-400 border-amber-500/20', icon: Bot }
  }
  if (lower.startsWith('gemini')) {
    return { name: 'Gemini', class: 'bg-blue-500/10 text-blue-400 border-blue-500/20', icon: Sparkles }
  }
  return { name: 'OSS', class: 'bg-purple-500/10 text-purple-400 border-purple-500/20', icon: Layers }
}

function getTierBadge(id: string) {
  const lower = id.toLowerCase()
  if (lower.includes('high') || lower.includes('pro')) {
    return { label: 'High Reasoning', class: 'bg-indigo-500/10 text-indigo-300 border-indigo-500/20' }
  }
  if (lower.includes('low') || lower.includes('flash-lite')) {
    return { label: 'Fast / Lite', class: 'bg-emerald-500/10 text-emerald-300 border-emerald-500/20' }
  }
  if (lower.includes('thinking')) {
    return { label: 'Thinking', class: 'bg-amber-500/10 text-amber-300 border-amber-500/20' }
  }
  if (lower.includes('agent')) {
    return { label: 'Agent', class: 'bg-sky-500/10 text-sky-300 border-sky-500/20' }
  }
  return { label: 'Standard', class: 'bg-zinc-800 text-zinc-400 border-zinc-700/60' }
}
</script>

<template>
  <div class="space-y-6">
    <!-- Top Header & Actions -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2.5">
          <h2 class="text-base font-semibold text-white tracking-tight">AI Model Catalog</h2>
          <Badge variant="outline" class="font-mono text-xs bg-zinc-800/80 text-zinc-300 border-zinc-700 px-2 py-0.5">
            {{ totalCount }} Available
          </Badge>
        </div>
        <p class="text-xs text-zinc-400 mt-1">
          Explore and query models accessible via OpenAI-compatible endpoints with auto OAuth token refresh.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <!-- Base URL Copy Pill -->
        <button
          type="button"
          @click="copyBaseUrl"
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-[#121316] border border-zinc-800/80 text-xs font-mono text-zinc-300 hover:border-zinc-700 hover:text-white transition-all cursor-pointer shadow-xs"
          title="Click to copy OpenAI-compatible Base URL"
        >
          <span class="text-zinc-500 font-sans text-[11px]">Base:</span>
          <span>/v1</span>
          <Check v-if="copiedBaseUrl" class="h-3 w-3 text-emerald-400 ml-0.5" />
          <Copy v-else class="h-3 w-3 text-zinc-500 ml-0.5" />
        </button>

        <!-- Refresh Button -->
        <Button
          variant="outline"
          size="sm"
          @click="usageStore.fetchModels()"
          :disabled="usageStore.loadingModels"
          class="h-8 px-3 text-xs bg-[#121316] border-zinc-800/80 hover:bg-zinc-800 hover:text-white text-zinc-300 gap-1.5 transition-all shadow-xs"
        >
          <RotateCw class="h-3.5 w-3.5" :class="{ 'animate-spin': usageStore.loadingModels }" />
          <span class="hidden sm:inline">Refresh</span>
        </Button>
      </div>
    </div>

    <!-- 4 KPI Summary Cards -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-3.5">
      <!-- Total Models -->
      <Card class="p-4 bg-[#121316] border-zinc-800/80 flex items-center justify-between shadow-xs">
        <div class="space-y-1">
          <span class="text-[11px] font-medium text-zinc-400 uppercase tracking-wider">Total Models</span>
          <div class="text-xl font-bold font-mono text-white">
            <span v-if="usageStore.loadingModels" class="inline-block w-8 h-6 bg-zinc-800/60 rounded animate-pulse" />
            <span v-else>{{ totalCount }}</span>
          </div>
          <span class="text-[10px] text-zinc-500 block">OpenAI Compatible</span>
        </div>
        <div class="w-9 h-9 rounded-xl bg-zinc-800/60 border border-zinc-700/40 flex items-center justify-center text-zinc-300">
          <Cpu class="h-4.5 w-4.5" />
        </div>
      </Card>

      <!-- Google Gemini -->
      <Card class="p-4 bg-[#121316] border-zinc-800/80 flex items-center justify-between shadow-xs">
        <div class="space-y-1">
          <span class="text-[11px] font-medium text-zinc-400 uppercase tracking-wider">Google Gemini</span>
          <div class="text-xl font-bold font-mono text-blue-400">
            <span v-if="usageStore.loadingModels" class="inline-block w-8 h-6 bg-zinc-800/60 rounded animate-pulse" />
            <span v-else>{{ geminiCount }}</span>
          </div>
          <span class="text-[10px] text-zinc-500 block">1M+ Token Window</span>
        </div>
        <div class="w-9 h-9 rounded-xl bg-blue-500/10 border border-blue-500/20 flex items-center justify-center text-blue-400">
          <Sparkles class="h-4.5 w-4.5" />
        </div>
      </Card>

      <!-- Anthropic Claude -->
      <Card class="p-4 bg-[#121316] border-zinc-800/80 flex items-center justify-between shadow-xs">
        <div class="space-y-1">
          <span class="text-[11px] font-medium text-zinc-400 uppercase tracking-wider">Anthropic Claude</span>
          <div class="text-xl font-bold font-mono text-amber-400">
            <span v-if="usageStore.loadingModels" class="inline-block w-8 h-6 bg-zinc-800/60 rounded animate-pulse" />
            <span v-else>{{ claudeCount }}</span>
          </div>
          <span class="text-[10px] text-zinc-500 block">Extended Thinking</span>
        </div>
        <div class="w-9 h-9 rounded-xl bg-amber-500/10 border border-amber-500/20 flex items-center justify-center text-amber-400">
          <Bot class="h-4.5 w-4.5" />
        </div>
      </Card>

      <!-- Open-Weight / OSS -->
      <Card class="p-4 bg-[#121316] border-zinc-800/80 flex items-center justify-between shadow-xs">
        <div class="space-y-1">
          <span class="text-[11px] font-medium text-zinc-400 uppercase tracking-wider">Open-Weight & OSS</span>
          <div class="text-xl font-bold font-mono text-purple-400">
            <span v-if="usageStore.loadingModels" class="inline-block w-8 h-6 bg-zinc-800/60 rounded animate-pulse" />
            <span v-else>{{ ossCount }}</span>
          </div>
          <span class="text-[10px] text-zinc-500 block">Open Source / Misc</span>
        </div>
        <div class="w-9 h-9 rounded-xl bg-purple-500/10 border border-purple-500/20 flex items-center justify-center text-purple-400">
          <Layers class="h-4.5 w-4.5" />
        </div>
      </Card>
    </div>

    <!-- Filter & View Controls Card -->
    <Card class="p-4 bg-[#121316] border-zinc-800/80 shadow-xs space-y-3.5">
      <div class="flex flex-col lg:flex-row lg:items-center justify-between gap-3">
        <!-- Search Input -->
        <div class="relative flex-1 max-w-md">
          <Search class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500 pointer-events-none" />
          <Input
            v-model="search"
            type="text"
            placeholder="Search model ID or provider..."
            class="bg-[#0d0e11] border-zinc-800/80 pl-9 pr-8 text-xs h-9 focus-visible:ring-blue-500 text-zinc-100 placeholder:text-zinc-600 rounded-lg"
          />
          <button
            v-if="search"
            type="button"
            @click="search = ''"
            class="absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300 p-0.5 rounded cursor-pointer"
          >
            <X class="h-3.5 w-3.5" />
          </button>
        </div>

        <div class="flex flex-wrap items-center gap-2">
          <!-- Tier Filter Selector -->
          <div class="flex items-center gap-1.5">
            <SlidersHorizontal class="h-3.5 w-3.5 text-zinc-500" />
            <select
              v-model="activeTier"
              class="h-9 rounded-lg bg-[#0d0e11] border border-zinc-800/80 px-2.5 py-1 text-xs text-zinc-200 focus:outline-none focus:border-zinc-700 cursor-pointer font-sans"
            >
              <option v-for="t in tiers" :key="t.key" :value="t.key" class="bg-[#121316] text-zinc-200">
                {{ t.label }}
              </option>
            </select>
          </div>

          <!-- View Mode Toggle -->
          <div class="flex items-center bg-[#0d0e11] border border-zinc-800/80 rounded-lg p-0.5">
            <button
              type="button"
              @click="viewMode = 'grid'"
              class="p-1.5 rounded-md transition-all cursor-pointer"
              :class="viewMode === 'grid' ? 'bg-zinc-800 text-white shadow-xs' : 'text-zinc-500 hover:text-zinc-300'"
              title="Card Grid View"
            >
              <LayoutGrid class="h-3.5 w-3.5" />
            </button>
            <button
              type="button"
              @click="viewMode = 'table'"
              class="p-1.5 rounded-md transition-all cursor-pointer"
              :class="viewMode === 'table' ? 'bg-zinc-800 text-white shadow-xs' : 'text-zinc-500 hover:text-zinc-300'"
              title="Dense Table View"
            >
              <List class="h-3.5 w-3.5" />
            </button>
          </div>
        </div>
      </div>

      <!-- Category Filter Pills -->
      <div class="flex flex-wrap items-center justify-between gap-2 pt-1 border-t border-zinc-800/60">
        <div class="flex flex-wrap items-center gap-1.5">
          <button
            v-for="cat in categories"
            :key="cat.key"
            type="button"
            @click="activeCategory = cat.key"
            class="px-2.5 py-1 rounded-md text-xs font-medium transition-all flex items-center gap-1.5 cursor-pointer"
            :class="
              activeCategory === cat.key
                ? 'bg-zinc-800 text-white border border-zinc-700/80'
                : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/40 border border-transparent'
            "
          >
            <component :is="cat.icon" class="h-3 w-3" :class="cat.color || 'text-zinc-400'" />
            <span>{{ cat.label }}</span>
            <span
              class="text-[10px] font-mono px-1.5 py-0.2 rounded-full"
              :class="activeCategory === cat.key ? 'bg-zinc-700 text-white' : 'bg-zinc-900 text-zinc-500'"
            >
              {{ cat.count }}
            </span>
          </button>
        </div>

        <!-- Showing status & reset button -->
        <div class="flex items-center gap-2 text-xs text-zinc-500">
          <span>Showing <strong class="text-zinc-300 font-mono">{{ filteredModels.length }}</strong> of {{ totalCount }}</span>
          <button
            v-if="hasActiveFilters"
            type="button"
            @click="resetFilters"
            class="text-blue-400 hover:underline cursor-pointer flex items-center gap-1 font-medium"
          >
            <X class="h-3 w-3" />
            <span>Reset</span>
          </button>
        </div>
      </div>
    </Card>

    <!-- Loading Skeletons -->
    <div v-if="usageStore.loadingModels">
      <!-- Grid Skeleton -->
      <div v-if="viewMode === 'grid'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        <Skeleton v-for="i in 6" :key="i" class="h-48 rounded-xl bg-zinc-900/60 border border-zinc-800/60" />
      </div>

      <!-- Table Skeleton -->
      <Card v-else class="bg-[#121316] border-zinc-800/80 p-4 space-y-3">
        <Skeleton v-for="i in 6" :key="i" class="h-10 w-full rounded bg-zinc-900/60 border border-zinc-800/60" />
      </Card>
    </div>

    <!-- Empty State -->
    <Card
      v-else-if="filteredModels.length === 0"
      class="py-16 text-center bg-[#121316] border-zinc-800/80 space-y-3"
    >
      <div class="w-12 h-12 rounded-2xl bg-zinc-900/90 border border-zinc-800 flex items-center justify-center mx-auto text-zinc-500">
        <Search class="h-5 w-5" />
      </div>
      <div class="space-y-1">
        <h3 class="text-sm font-semibold text-white">No models found</h3>
        <p class="text-xs text-zinc-500 max-w-sm mx-auto">
          No available AI models matched your current filter criteria.
        </p>
      </div>
      <Button
        variant="outline"
        size="sm"
        @click="resetFilters"
        class="h-8 px-3 text-xs bg-zinc-900 border-zinc-800 hover:bg-zinc-800 text-zinc-300 gap-1.5"
      >
        <X class="h-3.5 w-3.5" />
        <span>Reset Filters</span>
      </Button>
    </Card>

    <!-- Grid View Mode -->
    <div v-else-if="viewMode === 'grid'" class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <ModelCard
        v-for="m in filteredModels"
        :key="m.id"
        :model="m"
        @test="testModel"
        @code="openCodeSnippet"
      />
    </div>

    <!-- Table View Mode -->
    <Card v-else class="bg-[#121316] border-zinc-800/80 shadow-xs overflow-hidden">
      <div class="overflow-x-auto">
        <Table>
          <TableHeader>
            <TableRow class="border-zinc-800/80 hover:bg-transparent">
              <TableHead class="text-zinc-400 font-sans text-xs">Model Identifier</TableHead>
              <TableHead class="text-zinc-400 font-sans text-xs">Provider</TableHead>
              <TableHead class="text-zinc-400 font-sans text-xs">Reasoning Tier</TableHead>
              <TableHead class="text-zinc-400 font-sans text-xs">Context Limit</TableHead>
              <TableHead class="text-right text-zinc-400 font-sans text-xs pr-6">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <TableRow
              v-for="m in filteredModels"
              :key="m.id"
              class="border-zinc-800/60 hover:bg-zinc-800/30 transition-colors group"
            >
              <!-- Model ID -->
              <TableCell class="py-3 font-mono text-xs text-white">
                <div class="flex items-center gap-2">
                  <component :is="getFamilyBadge(m.id).icon" class="h-3.5 w-3.5 shrink-0" :class="getFamilyBadge(m.id).class.split(' ')[1]" />
                  <span class="font-semibold select-all">{{ m.id }}</span>
                  <button
                    type="button"
                    @click="copyModelId(m.id)"
                    class="p-1 rounded text-zinc-500 hover:text-white transition-colors cursor-pointer"
                    :title="copiedRowId === m.id ? 'Copied!' : 'Copy Model ID'"
                  >
                    <Check v-if="copiedRowId === m.id" class="h-3 w-3 text-emerald-400" />
                    <Copy v-else class="h-3 w-3 opacity-0 group-hover:opacity-100 transition-opacity" />
                  </button>
                </div>
              </TableCell>

              <!-- Provider -->
              <TableCell class="py-3">
                <Badge variant="outline" class="font-mono text-[11px]" :class="getFamilyBadge(m.id).class">
                  {{ m.owned_by }}
                </Badge>
              </TableCell>

              <!-- Tier -->
              <TableCell class="py-3">
                <Badge variant="outline" class="font-mono text-[10px]" :class="getTierBadge(m.id).class">
                  {{ getTierBadge(m.id).label }}
                </Badge>
              </TableCell>

              <!-- Context Limit -->
              <TableCell class="py-3 text-xs text-zinc-400 font-mono">
                <span v-if="m.id.toLowerCase().startsWith('gemini')" class="text-blue-300">1,000,000+</span>
                <span v-else-if="m.id.toLowerCase().startsWith('claude')" class="text-amber-300">200,000</span>
                <span v-else class="text-zinc-400">128,000</span>
                <span class="text-zinc-500 text-[10px] ml-1">tokens</span>
              </TableCell>

              <!-- Actions -->
              <TableCell class="py-3 text-right pr-6">
                <div class="flex items-center justify-end gap-1.5">
                  <Button
                    variant="outline"
                    size="sm"
                    @click="openCodeSnippet(m.id)"
                    class="h-7 px-2 text-xs bg-zinc-900 border-zinc-800 hover:bg-zinc-800 text-zinc-300 gap-1"
                    title="View integration code snippet"
                  >
                    <Code2 class="h-3 w-3" />
                    <span>Code</span>
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    @click="testModel(m.id)"
                    class="h-7 px-2.5 text-xs bg-zinc-900 border-zinc-800 hover:bg-blue-600 hover:text-white hover:border-blue-500 gap-1 text-zinc-300 transition-all"
                  >
                    <FlaskConical class="h-3 w-3" />
                    <span>Test</span>
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </Card>

    <!-- Code Snippet Modal -->
    <ModelCodeModal
      v-model:open="codeModalOpen"
      :model-id="selectedModelId"
    />
  </div>
</template>
