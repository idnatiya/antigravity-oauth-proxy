import { defineStore } from 'pinia'
import { ref } from 'vue'
import { usageService } from '@/services/usageService'
import { modelsService } from '@/services/modelsService'
import type { UsageStats, AccountInfo, RequestRecord, RequestFilter, OpenAIModel } from '@/types'

export const useUsageStore = defineStore('usage', () => {
  const stats = ref<UsageStats | null>(null)
  const account = ref<AccountInfo | null>(null)
  const timeRange = ref<string>('24h')
  const loadingStats = ref(false)

  const requests = ref<RequestRecord[]>([])
  const totalRequests = ref(0)
  const currentFilter = ref<RequestFilter>({
    limit: 25,
    offset: 0,
    search: '',
    model: '',
    status: undefined,
  })
  const loadingRequests = ref(false)

  const models = ref<OpenAIModel[]>([])
  const loadingModels = ref(false)

  async function fetchStats(range?: string): Promise<void> {
    if (range) timeRange.value = range
    loadingStats.value = true
    try {
      const data = await usageService.getStats(timeRange.value)
      stats.value = data.stats
      account.value = data.account
    } finally {
      loadingStats.value = false
    }
  }

  async function fetchRequests(filterUpdates?: Partial<RequestFilter>): Promise<void> {
    if (filterUpdates) {
      currentFilter.value = { ...currentFilter.value, ...filterUpdates }
    }
    loadingRequests.value = true
    try {
      const data = await usageService.getRequests(currentFilter.value)
      requests.value = data.requests || []
      totalRequests.value = data.total || 0
    } finally {
      loadingRequests.value = false
    }
  }

  async function fetchModels(): Promise<void> {
    loadingModels.value = true
    try {
      models.value = await modelsService.getModels()
    } finally {
      loadingModels.value = false
    }
  }

  return {
    stats,
    account,
    timeRange,
    loadingStats,
    requests,
    totalRequests,
    currentFilter,
    loadingRequests,
    models,
    loadingModels,
    fetchStats,
    fetchRequests,
    fetchModels,
  }
})
