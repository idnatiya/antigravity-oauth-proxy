<script setup lang="ts">
import { onMounted, computed } from 'vue'
import { RouterLink } from 'vue-router'
import {
  Activity,
  CheckCircle2,
  HardDrive,
  Zap,
  FlaskConical,
  ListFilter,
  Users,
} from '@lucide/vue'
import { useUsageStore } from '@/stores/usageStore'
import StatCard from '@/components/overview/StatCard.vue'
import UsageChart from '@/components/overview/UsageChart.vue'
import TopModelsTable from '@/components/overview/TopModelsTable.vue'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'

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
    <!-- Gateway Health Hero Banner -->
    <div class="relative overflow-hidden rounded-xl bg-gradient-to-r from-blue-950/30 via-zinc-900/60 to-purple-950/20 border border-zinc-800/80 p-5 md:p-6 shadow-sm">
      <div class="absolute -right-10 -bottom-10 w-60 h-60 bg-blue-500/5 rounded-full blur-3xl pointer-events-none" />

      <div class="flex flex-col md:flex-row md:items-center justify-between gap-5 relative z-10">
        <div class="space-y-1.5">
          <div class="flex items-center gap-2">
            <span class="relative flex h-2.5 w-2.5">
              <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
              <span class="relative inline-flex rounded-full h-2.5 w-2.5 bg-emerald-500" />
            </span>
            <span class="text-xs font-semibold uppercase tracking-wider text-emerald-400 font-mono">
              Gateway Operational
            </span>
          </div>

          <h2 class="text-xl font-bold tracking-tight text-white flex items-center gap-2">
            <span>Antigravity LLM Proxy Control Plane</span>
          </h2>

          <p class="text-xs text-zinc-400 max-w-2xl leading-relaxed">
            High-performance bridge converting OpenAI and Gemini API protocols to Google Cloud Code Assist. Zero cold starts, automatic token management, and pooled multi-account rotation.
          </p>
        </div>

        <!-- Quick Action Shortcuts -->
        <div class="flex items-center flex-wrap gap-2.5 shrink-0">
          <RouterLink to="/playground">
            <Button size="sm" class="bg-blue-600 hover:bg-blue-500 text-white gap-1.5 text-xs h-9 shadow-md shadow-blue-500/20">
              <FlaskConical class="h-3.5 w-3.5" />
              <span>Test Playground</span>
            </Button>
          </RouterLink>

          <RouterLink to="/requests">
            <Button variant="outline" size="sm" class="bg-zinc-900/80 border-zinc-800/80 hover:bg-zinc-800 text-zinc-300 gap-1.5 text-xs h-9">
              <ListFilter class="h-3.5 w-3.5 text-zinc-400" />
              <span>Telemetry Logs</span>
            </Button>
          </RouterLink>

          <RouterLink to="/accounts">
            <Button variant="outline" size="sm" class="bg-zinc-900/80 border-zinc-800/80 hover:bg-zinc-800 text-zinc-300 gap-1.5 text-xs h-9">
              <Users class="h-3.5 w-3.5 text-zinc-400" />
              <span>OAuth Pool</span>
            </Button>
          </RouterLink>
        </div>
      </div>
    </div>

    <!-- Controls Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pt-2">
      <div>
        <h3 class="text-base font-semibold text-white tracking-tight">System Performance</h3>
        <p class="text-xs text-zinc-400">Aggregate telemetry across all proxy completions and streaming channels</p>
      </div>

      <!-- Time Range Selector -->
      <Tabs :model-value="usageStore.timeRange" @update:model-value="selectRange($event as string)">
        <TabsList class="bg-zinc-900/90 border border-zinc-800/80 p-1">
          <TabsTrigger
            v-for="r in ranges"
            :key="r.value"
            :value="r.value"
            :data-state="usageStore.timeRange === r.value ? 'active' : 'inactive'"
            class="text-xs px-3 py-1 data-[state=active]:bg-zinc-800 data-[state=active]:text-white font-medium"
          >
            {{ r.label }}
          </TabsTrigger>
        </TabsList>
      </Tabs>
    </div>

    <!-- Stat Cards Grid -->
    <div v-if="usageStore.loadingStats && !usageStore.stats" class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <Skeleton v-for="i in 4" :key="i" class="h-32 rounded-xl bg-zinc-900/80 border border-zinc-800/60" />
    </div>

    <div v-else class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
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
      :points="usageStore.stats?.time_series || (usageStore.stats as any)?.timeline"
      :loading="usageStore.loadingStats"
    />

    <!-- Models Breakdown -->
    <TopModelsTable
      :models="usageStore.stats?.model_breakdown"
    />
  </div>
</template>
