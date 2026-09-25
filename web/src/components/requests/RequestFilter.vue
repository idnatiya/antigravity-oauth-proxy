<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { Search, X, Cpu, SlidersHorizontal, RotateCcw } from '@lucide/vue'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import type { RequestFilter } from '@/types'

const props = defineProps<{
  filter: RequestFilter
  availableModels?: string[]
}>()

const emit = defineEmits<{
  (e: 'update:filter', val: RequestFilter): void
  (e: 'search'): void
}>()

const localSearch = ref(props.filter.search || '')
const localModel = ref(props.filter.model || '')
const localStatus = ref(props.filter.status?.toString() || '')

let debounceTimer: ReturnType<typeof setTimeout> | null = null

const activeFilterCount = computed(() => {
  let count = 0
  if (localSearch.value.trim()) count++
  if (localModel.value) count++
  if (localStatus.value) count++
  return count
})

function updateFilter() {
  emit('update:filter', {
    ...props.filter,
    search: localSearch.value.trim(),
    model: localModel.value,
    status: localStatus.value ? parseInt(localStatus.value, 10) : undefined,
    offset: 0, // Reset to page 1 on filter change
  })
  emit('search')
}

function onSearchInput() {
  if (debounceTimer) clearTimeout(debounceTimer)
  debounceTimer = setTimeout(() => {
    updateFilter()
  }, 300)
}

function clearSearch() {
  localSearch.value = ''
  updateFilter()
}

function clearFilters() {
  localSearch.value = ''
  localModel.value = ''
  localStatus.value = ''
  updateFilter()
}

function setQuickStatus(statusVal: string) {
  localStatus.value = statusVal
  updateFilter()
}

watch(
  () => props.filter,
  (newF) => {
    localSearch.value = newF.search || ''
    localModel.value = newF.model || ''
    localStatus.value = newF.status?.toString() || ''
  },
  { deep: true }
)
</script>

<template>
  <Card class="p-4 bg-[#121316] border-zinc-800/80 shadow-sm space-y-3">
    <!-- Filter Controls Row -->
    <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-3">
      <!-- Search Bar -->
      <div class="relative flex-1 min-w-[220px]">
        <Search class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500 pointer-events-none" />
        <Input
          v-model="localSearch"
          @input="onSearchInput"
          placeholder="Filter by keyword, client IP, path..."
          class="bg-[#0d0e11] border-zinc-800/80 pl-9 pr-8 text-xs h-9 focus-visible:ring-blue-500/80 text-zinc-200 placeholder:text-zinc-500"
        />
        <button
          v-if="localSearch"
          type="button"
          @click="clearSearch"
          class="absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300 p-0.5 rounded cursor-pointer"
        >
          <X class="h-3.5 w-3.5" />
        </button>
      </div>

      <!-- Model Filter Dropdown -->
      <div class="relative sm:w-64">
        <select
          v-model="localModel"
          @change="updateFilter"
          class="w-full bg-[#0d0e11] border border-zinc-800/80 focus:border-blue-500/80 rounded-lg px-3 py-1.5 text-xs text-zinc-200 focus:outline-none transition-colors h-9 cursor-pointer appearance-none pr-8 font-mono"
        >
          <option value="">All Models ({{ availableModels?.length || 0 }})</option>
          <option v-for="m in availableModels" :key="m" :value="m">{{ m }}</option>
        </select>
        <Cpu class="h-3.5 w-3.5 absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-500 pointer-events-none" />
      </div>

      <!-- Status Code Filter Dropdown -->
      <div class="relative sm:w-48">
        <select
          v-model="localStatus"
          @change="updateFilter"
          class="w-full bg-[#0d0e11] border border-zinc-800/80 focus:border-blue-500/80 rounded-lg px-3 py-1.5 text-xs text-zinc-200 focus:outline-none transition-colors h-9 cursor-pointer appearance-none pr-8"
        >
          <option value="">All Statuses</option>
          <option value="200">200 OK</option>
          <option value="400">400 Bad Request</option>
          <option value="401">401 Unauthorized</option>
          <option value="404">404 Not Found</option>
          <option value="429">429 Rate Limited</option>
          <option value="500">500 Server Error</option>
        </select>
        <SlidersHorizontal class="h-3.5 w-3.5 absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-500 pointer-events-none" />
      </div>

      <!-- Clear / Reset Button -->
      <Button
        v-if="activeFilterCount > 0"
        variant="ghost"
        size="sm"
        @click="clearFilters"
        class="h-9 px-3 gap-1.5 text-xs text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/60 shrink-0"
      >
        <RotateCcw class="h-3.5 w-3.5 text-zinc-500" />
        <span>Reset</span>
        <Badge variant="outline" class="font-mono text-[10px] px-1 py-0 h-4 bg-zinc-800 border-zinc-700 text-zinc-300">
          {{ activeFilterCount }}
        </Badge>
      </Button>
    </div>

    <!-- Quick Filter Pill Strip -->
    <div class="flex items-center gap-2 pt-1 border-t border-zinc-800/60 text-xs">
      <span class="text-[11px] text-zinc-500 font-medium uppercase tracking-wider">Quick:</span>
      <div class="flex items-center gap-1.5 overflow-x-auto no-scrollbar">
        <button
          type="button"
          @click="setQuickStatus('')"
          class="px-2.5 py-0.5 rounded-md text-[11px] font-medium transition-colors cursor-pointer"
          :class="!localStatus ? 'bg-zinc-800 text-white border border-zinc-700/60' : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900'"
        >
          All
        </button>
        <button
          type="button"
          @click="setQuickStatus('200')"
          class="px-2.5 py-0.5 rounded-md text-[11px] font-medium transition-colors cursor-pointer flex items-center gap-1"
          :class="localStatus === '200' ? 'bg-emerald-500/15 text-emerald-400 border border-emerald-500/30' : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900'"
        >
          <span class="w-1.5 h-1.5 rounded-full bg-emerald-400" />
          <span>200 OK</span>
        </button>
        <button
          type="button"
          @click="setQuickStatus('429')"
          class="px-2.5 py-0.5 rounded-md text-[11px] font-medium transition-colors cursor-pointer flex items-center gap-1"
          :class="localStatus === '429' ? 'bg-amber-500/15 text-amber-400 border border-amber-500/30' : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900'"
        >
          <span class="w-1.5 h-1.5 rounded-full bg-amber-400" />
          <span>429 Rate Limited</span>
        </button>
        <button
          type="button"
          @click="setQuickStatus('500')"
          class="px-2.5 py-0.5 rounded-md text-[11px] font-medium transition-colors cursor-pointer flex items-center gap-1"
          :class="localStatus === '500' ? 'bg-red-500/15 text-red-400 border border-red-500/30' : 'text-zinc-400 hover:text-zinc-200 hover:bg-zinc-900'"
        >
          <span class="w-1.5 h-1.5 rounded-full bg-red-400" />
          <span>500 Error</span>
        </button>
      </div>
    </div>
  </Card>
</template>
