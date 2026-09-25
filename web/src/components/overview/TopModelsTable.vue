<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { Card } from '@/components/ui/card'
import {
  Table,
  TableHeader,
  TableRow,
  TableHead,
  TableBody,
  TableCell,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Progress } from '@/components/ui/progress'
import { FlaskConical } from '@lucide/vue'
import type { ModelStat } from '@/types'

const router = useRouter()

const props = defineProps<{
  models?: ModelStat[]
}>()

const totalReqs = computed(() => {
  if (!props.models) return 0
  return props.models.reduce((sum, m) => sum + m.request_count, 0)
})

function formatTokens(count: number): string {
  if (count >= 1_000_000) return (count / 1_000_000).toFixed(2) + 'M'
  if (count >= 1_000) return (count / 1_000).toFixed(1) + 'k'
  return count.toLocaleString()
}

function getProviderInfo(modelName: string) {
  if (modelName.startsWith('claude')) {
    return { name: 'Anthropic', badgeClass: 'bg-amber-500/10 text-amber-400 border-amber-500/25' }
  }
  if (modelName.startsWith('gemini')) {
    return { name: 'Google', badgeClass: 'bg-blue-500/10 text-blue-400 border-blue-500/25' }
  }
  return { name: 'OpenOSS', badgeClass: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/25' }
}

function testInPlayground(model: string) {
  router.push({ path: '/playground', query: { model } })
}
</script>

<template>
  <Card class="p-6 bg-[#121316] border-zinc-800/80 shadow-sm flex flex-col gap-4">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="text-sm font-semibold text-white tracking-tight">Active Models Distribution</h3>
        <p class="text-xs text-zinc-400 mt-0.5">Top models by cumulative request volume and token throughput</p>
      </div>
      <div v-if="models && models.length > 0" class="text-xs font-mono text-zinc-500">
        {{ models.length }} active models
      </div>
    </div>

    <div v-if="!models || models.length === 0" class="py-10 text-center text-xs text-zinc-500 font-mono">
      No model usage logged yet. Send requests through the proxy to see analytics.
    </div>

    <div v-else class="overflow-x-auto">
      <Table>
        <TableHeader>
          <TableRow class="hover:bg-transparent border-zinc-800/60">
            <TableHead class="text-zinc-400">Model</TableHead>
            <TableHead class="text-zinc-400 text-right">Requests</TableHead>
            <TableHead class="text-zinc-400 text-right">Throughput</TableHead>
            <TableHead class="text-zinc-400 w-44 text-right">Traffic Share</TableHead>
            <TableHead class="text-zinc-400 w-24 text-right">Action</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow
            v-for="m in models"
            :key="m.model"
            class="border-zinc-800/40 hover:bg-zinc-850/50 transition-colors group"
          >
            <!-- Model Name & Provider Badge -->
            <TableCell class="font-medium text-zinc-200">
              <div class="flex items-center gap-2">
                <Badge
                  variant="outline"
                  class="font-mono text-xs font-medium"
                  :class="getProviderInfo(m.model).badgeClass"
                >
                  {{ m.model }}
                </Badge>
                <span class="text-[10px] text-zinc-500 hidden sm:inline font-mono">
                  {{ getProviderInfo(m.model).name }}
                </span>
              </div>
            </TableCell>

            <!-- Requests -->
            <TableCell class="text-right font-mono text-zinc-300 text-xs">
              {{ m.request_count.toLocaleString() }}
            </TableCell>

            <!-- Tokens -->
            <TableCell class="text-right font-mono text-zinc-300 text-xs">
              {{ formatTokens(m.total_tokens) }}
            </TableCell>

            <!-- Traffic Share Bar -->
            <TableCell class="text-right">
              <div class="flex items-center justify-end gap-2.5">
                <Progress
                  :model-value="totalReqs > 0 ? (m.request_count / totalReqs) * 100 : 0"
                  class="w-24 h-1.5 bg-zinc-800"
                />
                <span class="font-mono text-[11px] text-zinc-400 w-9 text-right font-medium">
                  {{ totalReqs > 0 ? Math.round((m.request_count / totalReqs) * 100) : 0 }}%
                </span>
              </div>
            </TableCell>

            <!-- Test Action -->
            <TableCell class="text-right">
              <Button
                variant="ghost"
                size="sm"
                @click="testInPlayground(m.model)"
                class="h-7 px-2 text-xs text-zinc-400 hover:text-white hover:bg-blue-500/15 gap-1.5 transition-all opacity-80 group-hover:opacity-100"
                title="Test this model in Playground"
              >
                <FlaskConical class="h-3.5 w-3.5 text-blue-400" />
                <span class="text-[11px]">Test</span>
              </Button>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </Card>
</template>
