<script setup lang="ts">
import { ref, watch } from 'vue'
import { Search, X } from '@lucide/vue'
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

function clearFilters() {
  localSearch.value = ''
  localModel.value = ''
  localStatus.value = ''
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
  <div class="flex flex-wrap items-center gap-3 bg-[#202227] p-4 rounded-xl border border-[#2c2e36]">
    <!-- Search Bar -->
    <div class="relative flex-1 min-w-[240px]">
      <Search class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
      <input
        v-model="localSearch"
        @input="onSearchInput"
        placeholder="Filter by keyword, IP, endpoint..."
        class="w-full bg-[#18191d] border border-[#2c2e36] focus:border-blue-500 rounded-lg pl-9 pr-3 py-1.5 text-xs text-zinc-200 placeholder:text-zinc-600 focus:outline-none transition-colors"
      />
    </div>

    <!-- Model Filter Dropdown -->
    <select
      v-model="localModel"
      @change="updateFilter"
      class="bg-[#18191d] border border-[#2c2e36] focus:border-blue-500 rounded-lg px-3 py-1.5 text-xs text-zinc-200 focus:outline-none transition-colors"
    >
      <option value="">All Models</option>
      <option v-for="m in availableModels" :key="m" :value="m">{{ m }}</option>
    </select>

    <!-- Status Code Filter Dropdown -->
    <select
      v-model="localStatus"
      @change="updateFilter"
      class="bg-[#18191d] border border-[#2c2e36] focus:border-blue-500 rounded-lg px-3 py-1.5 text-xs text-zinc-200 focus:outline-none transition-colors"
    >
      <option value="">All Statuses</option>
      <option value="200">200 OK</option>
      <option value="400">400 Bad Request</option>
      <option value="401">401 Unauthorized</option>
      <option value="404">404 Not Found</option>
      <option value="429">429 Rate Limited</option>
      <option value="500">500 Server Error</option>
    </select>

    <!-- Clear Filters Button -->
    <button
      v-if="localSearch || localModel || localStatus"
      @click="clearFilters"
      class="flex items-center gap-1 px-2.5 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-400 hover:text-zinc-200 text-xs transition-colors cursor-pointer"
    >
      <X class="h-3.5 w-3.5" />
      <span>Reset</span>
    </button>
  </div>
</template>
