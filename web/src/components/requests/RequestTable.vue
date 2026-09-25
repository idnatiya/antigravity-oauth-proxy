<script setup lang="ts">
import { Radio, Eye, ChevronLeft, ChevronRight, ListFilter, Clock } from '@lucide/vue'
import { Card } from '@/components/ui/card'
import {
  Table,
  TableHeader,
  TableRow,
  TableHead,
  TableBody,
  TableCell,
  TableEmpty,
} from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
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
  (e: 'limit-change', newLimit: number): void
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

function getStatusBadge(code: number): { label: string; class: string; dot: string } {
  if (code >= 200 && code < 300) {
    return {
      label: `${code} OK`,
      class: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
      dot: 'bg-emerald-400',
    }
  }
  if (code >= 400 && code < 500) {
    return {
      label: code === 429 ? '429 Rate Limit' : `${code} Client Err`,
      class: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
      dot: 'bg-amber-400',
    }
  }
  return {
    label: `${code} Error`,
    class: 'bg-red-500/10 text-red-400 border-red-500/20',
    dot: 'bg-red-400',
  }
}

function formatLatency(ms?: number | null): string {
  if (ms === undefined || ms === null) return '–'
  if (ms >= 1000) return `${(ms / 1000).toFixed(2)}s`
  return `${ms.toLocaleString()}ms`
}

function getLatencyClass(ms?: number | null): string {
  if (!ms || ms < 1000) return 'text-zinc-200'
  if (ms < 3000) return 'text-amber-400 font-semibold'
  return 'text-red-400 font-semibold'
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

function onLimitChange(e: Event) {
  const target = e.target as HTMLSelectElement
  emit('limit-change', parseInt(target.value, 10))
}
</script>

<template>
  <Card class="bg-[#121316] border-zinc-800/80 shadow-sm overflow-hidden flex flex-col">
    <Table>
      <TableHeader class="bg-zinc-900/50 border-b border-zinc-800/80">
        <TableRow class="hover:bg-transparent border-zinc-800/80">
          <TableHead class="text-zinc-400 font-semibold text-xs py-3 w-[150px]">Time</TableHead>
          <TableHead class="text-zinc-400 font-semibold text-xs py-3">Endpoint</TableHead>
          <TableHead class="text-zinc-400 font-semibold text-xs py-3">Model</TableHead>
          <TableHead class="text-zinc-400 font-semibold text-xs py-3 text-center w-[130px]">Status</TableHead>
          <TableHead class="text-zinc-400 font-semibold text-xs py-3 text-right w-[110px]">Latency</TableHead>
          <TableHead class="text-zinc-400 font-semibold text-xs py-3 text-right w-[160px]">Tokens</TableHead>
          <TableHead class="text-zinc-400 font-semibold text-xs py-3 text-center w-[70px]">View</TableHead>
        </TableRow>
      </TableHeader>

      <!-- Skeleton Loading State -->
      <TableBody v-if="loading">
        <TableRow v-for="i in 6" :key="i" class="border-zinc-800/40">
          <TableCell class="py-3.5"><Skeleton class="h-4 w-28 bg-zinc-800/50" /></TableCell>
          <TableCell class="py-3.5"><Skeleton class="h-4 w-48 bg-zinc-800/50" /></TableCell>
          <TableCell class="py-3.5"><Skeleton class="h-4 w-32 bg-zinc-800/50" /></TableCell>
          <TableCell class="py-3.5 text-center"><Skeleton class="h-5 w-20 mx-auto rounded-full bg-zinc-800/50" /></TableCell>
          <TableCell class="py-3.5 text-right"><Skeleton class="h-4 w-16 ml-auto bg-zinc-800/50" /></TableCell>
          <TableCell class="py-3.5 text-right"><Skeleton class="h-4 w-24 ml-auto bg-zinc-800/50" /></TableCell>
          <TableCell class="py-3.5 text-center"><Skeleton class="h-7 w-7 mx-auto rounded-lg bg-zinc-800/50" /></TableCell>
        </TableRow>
      </TableBody>

      <!-- Empty State -->
      <TableBody v-else-if="requests.length === 0">
        <TableEmpty :colspan="7" class="py-12 text-center">
          <div class="flex flex-col items-center justify-center space-y-2">
            <div class="h-10 w-10 rounded-xl bg-zinc-900 border border-zinc-800 text-zinc-500 flex items-center justify-center">
              <ListFilter class="h-5 w-5" />
            </div>
            <div class="text-xs font-semibold text-zinc-300">No Request Records Found</div>
            <p class="text-[11px] text-zinc-500 max-w-sm">
              No telemetry traces match your active filter criteria. Try adjusting keywords or resetting filters.
            </p>
          </div>
        </TableEmpty>
      </TableBody>

      <!-- Table Rows -->
      <TableBody v-else>
        <TableRow
          v-for="r in requests"
          :key="r.id"
          class="cursor-pointer group hover:bg-zinc-800/40 border-zinc-800/40 transition-colors"
          @click="emit('select', r)"
        >
          <!-- Time -->
          <TableCell class="text-zinc-400 whitespace-nowrap py-3 font-mono text-xs">
            <div class="flex items-center gap-1.5">
              <Clock class="h-3 w-3 text-zinc-600 shrink-0" />
              <span>{{ formatTime(r.timestamp) }}</span>
              <span class="text-[10px] text-zinc-500">{{ formatDate(r.timestamp) }}</span>
            </div>
          </TableCell>

          <!-- Endpoint with Method & Streaming -->
          <TableCell class="text-zinc-300 py-3">
            <div class="flex items-center gap-2 max-w-xs md:max-w-md">
              <span class="px-1.5 py-0.5 rounded text-[10px] font-mono font-semibold bg-blue-500/10 text-blue-400 border border-blue-500/20 shrink-0">
                POST
              </span>
              <span class="font-mono text-xs text-zinc-200 truncate group-hover:text-blue-300 transition-colors" :title="r.endpoint">
                {{ r.endpoint }}
              </span>
              <Radio
                v-if="r.is_stream || r.stream"
                class="h-3.5 w-3.5 text-blue-400 shrink-0"
                title="Streaming SSE"
              />
            </div>
          </TableCell>

          <!-- Model -->
          <TableCell class="py-3">
            <Badge variant="outline" class="font-mono text-[11px] bg-blue-500/10 text-blue-400 border-blue-500/20">
              {{ r.model }}
            </Badge>
          </TableCell>

          <!-- Status Code -->
          <TableCell class="text-center py-3">
            <Badge
              variant="outline"
              class="font-mono text-[11px] gap-1.5 px-2 py-0.5"
              :class="getStatusBadge(r.status_code).class"
            >
              <span class="w-1.5 h-1.5 rounded-full" :class="getStatusBadge(r.status_code).dot" />
              <span>{{ getStatusBadge(r.status_code).label }}</span>
            </Badge>
          </TableCell>

          <!-- Latency -->
          <TableCell class="text-right py-3 font-mono text-xs" :class="getLatencyClass(r.latency_ms ?? r.duration_ms)">
            {{ formatLatency(r.latency_ms ?? r.duration_ms) }}
          </TableCell>

          <!-- Tokens -->
          <TableCell class="text-right text-zinc-300 whitespace-nowrap py-3 font-mono text-xs">
            <span class="text-zinc-100 font-semibold">{{ (r.total_tokens ?? 0).toLocaleString() }}</span>
            <span class="text-[10px] text-zinc-500 ml-1">({{ r.prompt_tokens ?? 0 }}/{{ r.completion_tokens ?? 0 }})</span>
          </TableCell>

          <!-- Action -->
          <TableCell class="text-center py-3">
            <Button
              variant="secondary"
              size="icon"
              class="h-7 w-7 bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 text-zinc-300 hover:text-white"
              @click.stop="emit('select', r)"
              title="Inspect request trace"
            >
              <Eye class="h-3.5 w-3.5" />
            </Button>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <!-- Pagination Footer -->
    <div class="px-5 py-3 border-t border-zinc-800/80 bg-zinc-900/40 flex flex-col sm:flex-row items-center justify-between text-xs text-zinc-400 font-mono gap-3">
      <div class="flex items-center gap-3">
        <div>
          Showing <strong class="text-zinc-200">{{ total > 0 ? offset + 1 : 0 }}</strong> -
          <strong class="text-zinc-200">{{ Math.min(offset + limit, total ?? 0) }}</strong> of
          <strong class="text-zinc-200">{{ (total ?? 0).toLocaleString() }}</strong> requests
        </div>

        <!-- Page Size Selector -->
        <div class="flex items-center gap-1.5 text-zinc-500 text-[11px]">
          <span>Rows:</span>
          <select
            :value="limit"
            @change="onLimitChange"
            class="bg-[#0d0e11] border border-zinc-800 rounded px-1.5 py-0.5 text-zinc-300 text-xs focus:outline-none cursor-pointer"
          >
            <option :value="10">10</option>
            <option :value="25">25</option>
            <option :value="50">50</option>
            <option :value="100">100</option>
          </select>
        </div>
      </div>

      <div class="flex items-center gap-2">
        <span class="text-zinc-500 text-[11px]">Page {{ currentPage() }} of {{ totalPages() }}</span>
        <Button
          variant="secondary"
          size="icon"
          class="h-7 w-7 bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 text-zinc-300 disabled:opacity-40"
          @click="prevPage"
          :disabled="offset === 0"
        >
          <ChevronLeft class="h-3.5 w-3.5" />
        </Button>
        <Button
          variant="secondary"
          size="icon"
          class="h-7 w-7 bg-zinc-900 border border-zinc-800 hover:bg-zinc-800 text-zinc-300 disabled:opacity-40"
          @click="nextPage"
          :disabled="offset + limit >= total"
        >
          <ChevronRight class="h-3.5 w-3.5" />
        </Button>
      </div>
    </div>
  </Card>
</template>
