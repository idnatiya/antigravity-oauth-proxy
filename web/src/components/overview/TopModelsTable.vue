<script setup lang="ts">
import { computed } from 'vue'
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
import type { ModelStat } from '@/types'

const props = defineProps<{
  models?: ModelStat[]
}>()

const totalReqs = computed(() => {
  if (!props.models) return 0
  return props.models.reduce((sum, m) => sum + m.request_count, 0)
})

function formatTokens(count: number): string {
  if (count >= 1_000_000) return (count / 1_000_000).toFixed(1) + 'M'
  if (count >= 1_000) return (count / 1_000).toFixed(1) + 'k'
  return count.toString()
}
</script>

<template>
  <Card class="p-6 bg-[#202227] border-[#2c2e36] flex flex-col gap-4">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="text-sm font-semibold text-white tracking-tight">Top Active Models</h3>
        <p class="text-xs text-zinc-500">Distribution by request count and token throughput</p>
      </div>
    </div>

    <div v-if="!models || models.length === 0" class="py-8 text-center text-xs text-zinc-500 font-mono">
      No model usage logged yet.
    </div>

    <div v-else>
      <Table>
        <TableHeader>
          <TableRow class="hover:bg-transparent">
            <TableHead>Model</TableHead>
            <TableHead class="text-right">Requests</TableHead>
            <TableHead class="text-right">Tokens</TableHead>
            <TableHead class="w-36 text-right">Share</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          <TableRow v-for="m in models" :key="m.model">
            <TableCell class="font-medium text-zinc-200">
              <Badge variant="outline" class="font-mono bg-blue-500/10 text-blue-400 border-blue-500/20">
                {{ m.model }}
              </Badge>
            </TableCell>
            <TableCell class="text-right text-zinc-300">
              {{ m.request_count.toLocaleString() }}
            </TableCell>
            <TableCell class="text-right text-zinc-300">
              {{ formatTokens(m.total_tokens) }}
            </TableCell>
            <TableCell class="text-right">
              <div class="flex items-center justify-end gap-2.5">
                <Progress
                  :model-value="totalReqs > 0 ? (m.request_count / totalReqs) * 100 : 0"
                  class="w-20 h-1.5 bg-zinc-800"
                />
                <span class="font-mono text-[10px] text-zinc-400 w-8 text-right">
                  {{ totalReqs > 0 ? Math.round((m.request_count / totalReqs) * 100) : 0 }}%
                </span>
              </div>
            </TableCell>
          </TableRow>
        </TableBody>
      </Table>
    </div>
  </Card>
</template>
