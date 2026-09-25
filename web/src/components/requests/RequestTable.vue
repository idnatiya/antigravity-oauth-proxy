<script setup lang="ts">
import { Radio, Eye, ChevronLeft, ChevronRight } from '@lucide/vue'
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

function getStatusBadgeVariant(code: number): 'success' | 'warning' | 'destructive' {
  if (code >= 200 && code < 300) return 'success'
  if (code >= 400 && code < 500) return 'warning'
  return 'destructive'
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
  <Card class="bg-[#202227] border-[#2c2e36] overflow-hidden flex flex-col">
    <Table>
      <TableHeader class="bg-[#1a1c22]/50">
        <TableRow class="hover:bg-transparent">
          <TableHead>Time</TableHead>
          <TableHead>Endpoint</TableHead>
          <TableHead>Model</TableHead>
          <TableHead class="text-center">Status</TableHead>
          <TableHead class="text-right">Latency</TableHead>
          <TableHead class="text-right">Tokens</TableHead>
          <TableHead class="text-center">Action</TableHead>
        </TableRow>
      </TableHeader>

      <TableBody v-if="loading">
        <TableEmpty :colspan="7" class="animate-pulse">
          Fetching request records...
        </TableEmpty>
      </TableBody>

      <TableBody v-else-if="requests.length === 0">
        <TableEmpty :colspan="7">
          No matching request telemetry records found.
        </TableEmpty>
      </TableBody>

      <TableBody v-else>
        <TableRow
          v-for="r in requests"
          :key="r.id"
          class="cursor-pointer group"
          @click="emit('select', r)"
        >
          <TableCell class="text-zinc-400 whitespace-nowrap">
            <span>{{ formatTime(r.timestamp) }}</span>
            <span class="text-[10px] text-zinc-600 ml-1.5">{{ formatDate(r.timestamp) }}</span>
          </TableCell>

          <TableCell class="text-zinc-300">
            <div class="flex items-center gap-1.5">
              <span class="truncate max-w-[140px]">{{ r.endpoint }}</span>
              <Radio v-if="r.is_stream" class="h-3 w-3 text-blue-400 shrink-0" title="Streaming SSE" />
            </div>
          </TableCell>

          <TableCell>
            <Badge variant="outline" class="font-mono bg-blue-500/10 text-blue-400 border-blue-500/20">
              {{ r.model }}
            </Badge>
          </TableCell>

          <TableCell class="text-center">
            <Badge :variant="getStatusBadgeVariant(r.status_code)" class="font-mono">
              {{ r.status_code }}
            </Badge>
          </TableCell>

          <TableCell class="text-right text-zinc-300">
            {{ r.latency_ms }}ms
          </TableCell>

          <TableCell class="text-right text-zinc-300 whitespace-nowrap">
            <span class="text-zinc-200">{{ r.total_tokens.toLocaleString() }}</span>
            <span class="text-[10px] text-zinc-500 ml-1">({{ r.prompt_tokens }}/{{ r.completion_tokens }})</span>
          </TableCell>

          <TableCell class="text-center">
            <Button
              variant="secondary"
              size="icon"
              class="h-7 w-7"
              @click.stop="emit('select', r)"
              title="View request details"
            >
              <Eye class="h-3.5 w-3.5" />
            </Button>
          </TableCell>
        </TableRow>
      </TableBody>
    </Table>

    <!-- Pagination Footer -->
    <div class="px-4 py-3 border-t border-[#2c2e36] bg-[#1a1c22]/30 flex items-center justify-between text-xs text-zinc-400 font-mono">
      <div>
        Showing <span class="text-zinc-200">{{ total > 0 ? offset + 1 : 0 }}</span> -
        <span class="text-zinc-200">{{ Math.min(offset + limit, total) }}</span> of
        <span class="text-zinc-200">{{ total.toLocaleString() }}</span> requests
      </div>

      <div class="flex items-center gap-2">
        <span class="text-zinc-500 text-[11px]">Page {{ currentPage() }} of {{ totalPages() }}</span>
        <Button
          variant="secondary"
          size="icon"
          class="h-7 w-7"
          @click="prevPage"
          :disabled="offset === 0"
        >
          <ChevronLeft class="h-3.5 w-3.5" />
        </Button>
        <Button
          variant="secondary"
          size="icon"
          class="h-7 w-7"
          @click="nextPage"
          :disabled="offset + limit >= total"
        >
          <ChevronRight class="h-3.5 w-3.5" />
        </Button>
      </div>
    </div>
  </Card>
</template>
