<script setup lang="ts">
import { Radio, Eye, ChevronLeft, ChevronRight } from '@lucide/vue'
import type { RequestRecord } from '@/types'

const props = defineProps<{
  requests: RequestRecord[]
  total: number
  limit: number
  offset: number
  loading?: boolean
}>()

const emit = defineEmits<{
  (e: 'select', record: RequestRecord): void
  (e: 'page-change', newOffset: number): void
}>()

function formatTime(iso: string): string {
  try {
    const d = new Date(iso)
    return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' })
  } catch {
    return iso
  }
}

function formatDate(iso: string): string {
  try {
    const d = new Date(iso)
    return d.toLocaleDateString([], { month: 'short', day: 'numeric' })
  } catch {
    return ''
  }
}

function statusColor(code: number): string {
  if (code >= 200 && code < 300) return 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
  if (code >= 400 && code < 500) return 'bg-amber-500/10 text-amber-400 border-amber-500/20'
  return 'bg-red-500/10 text-red-400 border-red-500/20'
}

const totalPages = () => Math.ceil(props.total / props.limit) || 1
const currentPage = () => Math.floor(props.offset / props.limit) + 1

function prevPage() {
  if (props.offset >= props.limit) {
    emit('page-change', props.offset - props.limit)
  }
}

function nextPage() {
  if (props.offset + props.limit < props.total) {
    emit('page-change', props.offset + props.limit)
  }
}
</script>

<template>
  <div class="bg-[#202227] rounded-xl border border-[#2c2e36] overflow-hidden flex flex-col">
    <div class="overflow-x-auto">
      <table class="w-full text-left text-xs">
        <thead class="text-zinc-500 border-b border-[#2c2e36] bg-[#1a1c22]/50 uppercase font-mono tracking-wider">
          <tr>
            <th class="py-3 px-4 font-medium">Time</th>
            <th class="py-3 px-4 font-medium">Endpoint</th>
            <th class="py-3 px-4 font-medium">Model</th>
            <th class="py-3 px-4 font-medium text-center">Status</th>
            <th class="py-3 px-4 font-medium text-right">Latency</th>
            <th class="py-3 px-4 font-medium text-right">Tokens</th>
            <th class="py-3 px-4 font-medium text-center">Action</th>
          </tr>
        </thead>

        <tbody v-if="loading" class="divide-y divide-[#2c2e36]">
          <tr>
            <td colspan="7" class="py-12 text-center text-zinc-500 font-mono animate-pulse">
              Fetching request records...
            </td>
          </tr>
        </tbody>

        <tbody v-else-if="requests.length === 0" class="divide-y divide-[#2c2e36]">
          <tr>
            <td colspan="7" class="py-12 text-center text-zinc-500 font-mono">
              No matching request telemetry records found.
            </td>
          </tr>
        </tbody>

        <tbody v-else class="divide-y divide-[#2c2e36]">
          <tr
            v-for="r in requests"
            :key="r.id"
            class="hover:bg-[#252830] transition-colors group cursor-pointer"
            @click="emit('select', r)"
          >
            <td class="py-3 px-4 font-mono text-zinc-400 whitespace-nowrap">
              <span>{{ formatTime(r.timestamp) }}</span>
              <span class="text-[10px] text-zinc-600 ml-1.5">{{ formatDate(r.timestamp) }}</span>
            </td>

            <td class="py-3 px-4 font-mono text-zinc-300 flex items-center gap-1.5">
              <span class="truncate max-w-[140px]">{{ r.endpoint }}</span>
              <Radio v-if="r.is_stream" class="h-3 w-3 text-blue-400 shrink-0" title="Streaming SSE" />
            </td>

            <td class="py-3 px-4 font-mono">
              <span class="px-2 py-0.5 rounded bg-zinc-800 text-blue-400 border border-zinc-700/60 font-medium">
                {{ r.model }}
              </span>
            </td>

            <td class="py-3 px-4 text-center font-mono">
              <span class="px-2 py-0.5 rounded border text-[11px] font-semibold" :class="statusColor(r.status_code)">
                {{ r.status_code }}
              </span>
            </td>

            <td class="py-3 px-4 text-right font-mono text-zinc-300">
              {{ r.latency_ms }}ms
            </td>

            <td class="py-3 px-4 text-right font-mono text-zinc-300 whitespace-nowrap">
              <span class="text-zinc-200">{{ r.total_tokens.toLocaleString() }}</span>
              <span class="text-[10px] text-zinc-500 ml-1">({{ r.prompt_tokens }}/{{ r.completion_tokens }})</span>
            </td>

            <td class="py-3 px-4 text-center">
              <button
                @click.stop="emit('select', r)"
                class="p-1.5 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-400 hover:text-white transition-colors cursor-pointer"
                title="View request details"
              >
                <Eye class="h-3.5 w-3.5" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Pagination Footer -->
    <div class="px-4 py-3 border-t border-[#2c2e36] bg-[#1a1c22]/30 flex items-center justify-between text-xs text-zinc-400 font-mono">
      <div>
        Showing <span class="text-zinc-200">{{ total > 0 ? offset + 1 : 0 }}</span> -
        <span class="text-zinc-200">{{ Math.min(offset + limit, total) }}</span> of
        <span class="text-zinc-200">{{ total.toLocaleString() }}</span> requests
      </div>

      <div class="flex items-center gap-2">
        <span class="text-zinc-500 text-[11px]">Page {{ currentPage() }} of {{ totalPages() }}</span>
        <button
          @click="prevPage"
          :disabled="offset === 0"
          class="p-1.5 rounded bg-zinc-800 hover:bg-zinc-700 disabled:opacity-30 disabled:cursor-not-allowed text-zinc-300 cursor-pointer"
        >
          <ChevronLeft class="h-3.5 w-3.5" />
        </button>
        <button
          @click="nextPage"
          :disabled="offset + limit >= total"
          class="p-1.5 rounded bg-zinc-800 hover:bg-zinc-700 disabled:opacity-30 disabled:cursor-not-allowed text-zinc-300 cursor-pointer"
        >
          <ChevronRight class="h-3.5 w-3.5" />
        </button>
      </div>
    </div>
  </div>
</template>
