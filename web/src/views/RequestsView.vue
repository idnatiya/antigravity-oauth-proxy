<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { RefreshCw, Radio } from '@lucide/vue'
import { useUsageStore } from '@/stores/usageStore'
import RequestFilter from '@/components/requests/RequestFilter.vue'
import RequestTable from '@/components/requests/RequestTable.vue'
import RequestDetailModal from '@/components/requests/RequestDetailModal.vue'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import type { RequestRecord, RequestFilter as IRequestFilter } from '@/types'

const route = useRoute()
const router = useRouter()
const usageStore = useUsageStore()

const selectedRecord = ref<RequestRecord | null>(null)
const isModalOpen = ref(false)

const availableModelNames = computed(() => {
  return usageStore.models.map(m => m.id)
})

function syncFromQuery() {
  const query = route.query
  const limit = query.limit ? parseInt(query.limit as string, 10) : 25
  const page = query.page ? parseInt(query.page as string, 10) : 1
  const offset = (page - 1) * limit
  const search = (query.search as string) || ''
  const model = (query.model as string) || ''
  const status = query.status ? parseInt(query.status as string, 10) : undefined

  usageStore.fetchRequests({
    limit,
    offset,
    search,
    model,
    status,
  })
}

function syncToQuery(filter: IRequestFilter) {
  const page = Math.floor((filter.offset ?? 0) / (filter.limit ?? 25)) + 1
  const query: Record<string, string> = {}

  if (page > 1) query.page = String(page)
  if (filter.limit && filter.limit !== 25) query.limit = String(filter.limit)
  if (filter.search) query.search = filter.search
  if (filter.model) query.model = filter.model
  if (filter.status) query.status = String(filter.status)

  router.replace({ query })
}

function handleFilterUpdate(newFilter: IRequestFilter) {
  syncToQuery(newFilter)
  usageStore.fetchRequests(newFilter)
}

function handlePageChange(newOffset: number) {
  const updated = { ...usageStore.currentFilter, offset: newOffset }
  syncToQuery(updated)
  usageStore.fetchRequests(updated)
}

function handleLimitChange(newLimit: number) {
  const updated = { ...usageStore.currentFilter, limit: newLimit, offset: 0 }
  syncToQuery(updated)
  usageStore.fetchRequests(updated)
}

function openDetail(record: RequestRecord) {
  selectedRecord.value = record
  isModalOpen.value = true
}

function closeModal() {
  isModalOpen.value = false
  selectedRecord.value = null
}

function handleRefresh() {
  syncFromQuery()
}

onMounted(async () => {
  if (usageStore.models.length === 0) {
    usageStore.fetchModels().catch(() => {})
  }
  syncFromQuery()
})

// Listen to browser Back/Forward buttons
watch(
  () => route.query,
  () => {
    syncFromQuery()
  }
)
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2.5">
          <h2 class="text-base font-semibold text-white tracking-tight">Request Telemetry Logs</h2>
          <Badge variant="outline" class="font-mono text-emerald-400 bg-emerald-500/10 border-emerald-500/20 text-[11px] gap-1">
            <Radio class="h-2.5 w-2.5 text-emerald-400 animate-pulse" />
            Live Tracing
          </Badge>
        </div>
        <p class="text-xs text-zinc-400 mt-0.5">
          Queryable request lifecycle logs, latency benchmarks, and token telemetry from SQLite storage
        </p>
      </div>

      <div class="flex items-center gap-2.5">
        <Badge variant="outline" class="font-mono text-xs text-zinc-400 border-zinc-800 bg-zinc-900/60 h-8 px-2.5">
          <strong class="text-white mr-1">{{ (usageStore.totalRequests ?? 0).toLocaleString() }}</strong> Total Traces
        </Badge>

        <Button
          variant="outline"
          size="sm"
          @click="handleRefresh"
          :disabled="usageStore.loadingRequests"
          class="h-8 text-xs bg-zinc-900/60 border-zinc-800 hover:bg-zinc-800 text-zinc-300 gap-1.5"
        >
          <RefreshCw
            class="h-3.5 w-3.5 text-zinc-400"
            :class="{ 'animate-spin': usageStore.loadingRequests }"
          />
          <span>Refresh</span>
        </Button>
      </div>
    </div>

    <!-- Filter Component -->
    <RequestFilter
      :filter="usageStore.currentFilter"
      :available-models="availableModelNames"
      @update:filter="handleFilterUpdate"
    />

    <!-- Telemetry Table -->
    <RequestTable
      :requests="usageStore.requests"
      :total="usageStore.totalRequests"
      :limit="usageStore.currentFilter.limit ?? 25"
      :offset="usageStore.currentFilter.offset ?? 0"
      :loading="usageStore.loadingRequests"
      @select="openDetail"
      @page-change="handlePageChange"
      @limit-change="handleLimitChange"
    />

    <!-- Detail Modal -->
    <RequestDetailModal
      :record="selectedRecord"
      :open="isModalOpen"
      @close="closeModal"
    />
  </div>
</template>
