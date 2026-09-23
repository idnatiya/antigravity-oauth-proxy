<script setup lang="ts">
import { computed } from 'vue'
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
  <div class="p-6 rounded-xl bg-[#202227] border border-[#2c2e36] flex flex-col gap-4">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="text-sm font-semibold text-white tracking-tight">Top Active Models</h3>
        <p class="text-xs text-zinc-500">Distribution by request count and token throughput</p>
      </div>
    </div>

    <div v-if="!models || models.length === 0" class="py-8 text-center text-xs text-zinc-500 font-mono">
      No model usage logged yet.
    </div>

    <div v-else class="overflow-x-auto">
      <table class="w-full text-left text-xs">
        <thead class="text-zinc-500 border-b border-[#2c2e36] uppercase font-mono tracking-wider">
          <tr>
            <th class="pb-3 font-medium">Model</th>
            <th class="pb-3 font-medium text-right">Requests</th>
            <th class="pb-3 font-medium text-right">Tokens</th>
            <th class="pb-3 font-medium w-36 text-right">Share</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-[#2c2e36]">
          <tr v-for="m in models" :key="m.model" class="hover:bg-[#252830] transition-colors">
            <td class="py-3 font-mono font-medium text-zinc-200">
              <span class="px-2 py-0.5 rounded bg-zinc-800 border border-zinc-700 text-blue-400">
                {{ m.model }}
              </span>
            </td>
            <td class="py-3 text-right font-mono text-zinc-300">
              {{ m.request_count.toLocaleString() }}
            </td>
            <td class="py-3 text-right font-mono text-zinc-300">
              {{ formatTokens(m.total_tokens) }}
            </td>
            <td class="py-3 text-right">
              <div class="flex items-center justify-end gap-2">
                <div class="w-20 bg-zinc-800 rounded-full h-1.5 overflow-hidden">
                  <div
                    class="bg-blue-500 h-full rounded-full"
                    :style="{ width: `${totalReqs > 0 ? (m.request_count / totalReqs) * 100 : 0}%` }"
                  ></div>
                </div>
                <span class="font-mono text-[10px] text-zinc-400 w-8 text-right">
                  {{ totalReqs > 0 ? Math.round((m.request_count / totalReqs) * 100) : 0 }}%
                </span>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
