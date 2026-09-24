<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { Activity, CheckCircle2, HardDrive, Zap } from '@lucide/vue'
import { useUsageStore } from '@/stores/usageStore'
import StatCard from '@/components/overview/StatCard.vue'
import UsageChart from '@/components/overview/UsageChart.vue'
import TopModelsTable from '@/components/overview/TopModelsTable.vue'

const usageStore = useUsageStore()

const ranges = [
  { label: 'Last 24 Hours', value: '24h' },
  { label: 'Last 7 Days', value: '7d' },
  { label: 'Last 30 Days', value: '30d' },
]

function selectRange(r: string) {
  usageStore.fetchStats(r)
}

onMounted(() => {
  usageStore.fetchStats()
})

const successRate = computed(() => {
  const total = usageStore.stats?.total_requests || 0
  if (total === 0) return '100%'
  const success = usageStore.stats?.success_requests || 0
  return ((success / total) * 100).toFixed(1) + '%'
})

const formattedTokens = computed(() => {
  const tokens = usageStore.stats?.total_tokens || 0
  if (tokens >= 1_000_000) return (tokens / 1_000_000).toFixed(2) + 'M'
  if (tokens >= 1_000) return (tokens / 1_000).toFixed(1) + 'k'
  return tokens.toLocaleString()
})
</script>

<template>
  <div class="space-y-6">
    <!-- Controls Header -->
    <div class="flex items-center justify-between">
      <div>
        <h2 class="text-base font-semibold text-white tracking-tight">System Performance</h2>
        <p class="text-xs text-zinc-500">Real-time aggregate usage across all LLM client endpoints</p>
      </div>

      <!-- Time Range Selector -->
      <div class="flex items-center bg-[#202227] border border-[#2c2e36] rounded-lg p-1 text-xs">
        <button
          v-for="r in ranges"
          :key="r.value"
          @click="selectRange(r.value)"
          class="px-3 py-1 rounded-md transition-all font-medium cursor-pointer"
          :class="[
            usageStore.timeRange === r.value
              ? 'bg-blue-600 text-white shadow-sm'
              : 'text-zinc-400 hover:text-zinc-200'
          ]"
        >
          {{ r.label }}
        </button>
      </div>
    </div>

    <!-- Stat Cards Grid -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <StatCard
        title="Total Requests"
        :value="usageStore.stats?.total_requests?.toLocaleString() || '0'"
        subtitle="Prompt & completion calls"
        :icon="Activity"
        color="blue"
      />

      <StatCard
        title="Success Rate"
        :value="successRate"
        :subtitle="`${usageStore.stats?.error_requests || 0} failed requests`"
        :icon="CheckCircle2"
        color="green"
      />

      <StatCard
        title="Token Throughput"
        :value="formattedTokens"
        :subtitle="`${(usageStore.stats?.prompt_tokens || 0).toLocaleString()} in / ${(usageStore.stats?.completion_tokens || 0).toLocaleString()} out`"
        :icon="HardDrive"
        color="purple"
      />

      <StatCard
        title="Average Latency"
        :value="`${Math.round(usageStore.stats?.avg_latency_ms || 0)} ms`"
        subtitle="End-to-end response time"
        :icon="Zap"
        color="amber"
      />
    </div>

    <!-- Usage Trends Chart -->
    <UsageChart
      :points="usageStore.stats?.time_series"
      :loading="usageStore.loadingStats"
    />

    <!-- Models Breakdown -->
    <TopModelsTable
      :models="usageStore.stats?.model_breakdown"
    />
  </div>
</template>
