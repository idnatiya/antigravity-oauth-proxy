<script setup lang="ts">
import { ref } from 'vue'
import {
  AlertCircle,
  CheckCircle2,
  Clock,
  Cpu,
  HardDrive,
  Copy,
  Check,
  Globe,
  Radio,
} from '@lucide/vue'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import type { RequestRecord } from '@/types'

const props = defineProps<{
  record: RequestRecord | null
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const copiedId = ref(false)
const copiedJson = ref(false)

function copyId() {
  if (!props.record) return
  navigator.clipboard.writeText(props.record.id)
  copiedId.value = true
  setTimeout(() => (copiedId.value = false), 2000)
}

function copyJson() {
  if (!props.record) return
  navigator.clipboard.writeText(JSON.stringify(props.record, null, 2))
  copiedJson.value = true
  setTimeout(() => (copiedJson.value = false), 2000)
}
</script>

<template>
  <Dialog :open="open && !!record" @update:open="val => { if (!val) emit('close') }">
    <DialogContent class="max-w-2xl bg-[#121316] border-zinc-800 p-0 overflow-hidden gap-0 shadow-2xl">
      <!-- Modal Header -->
      <DialogHeader class="px-6 py-4 border-b border-zinc-800/80 bg-zinc-900/40">
        <div class="flex items-center justify-between">
          <div class="flex items-center gap-3 min-w-0">
            <div
              v-if="record"
              class="h-9 w-9 rounded-xl flex items-center justify-center shrink-0 border"
              :class="record.status_code < 400
                ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20'
                : 'bg-red-500/10 text-red-400 border-red-500/20'"
            >
              <CheckCircle2 v-if="record.status_code < 400" class="h-4 w-4" />
              <AlertCircle v-else class="h-4 w-4" />
            </div>

            <div class="min-w-0 flex-1">
              <DialogTitle class="text-sm font-semibold text-white flex items-center gap-2">
                <span>Request Trace Details</span>
                <Badge
                  v-if="record"
                  variant="outline"
                  class="font-mono text-[10px] px-1.5 py-0 h-4"
                  :class="record.status_code < 400 ? 'text-emerald-400 border-emerald-500/30' : 'text-red-400 border-red-500/30'"
                >
                  {{ record.status_code }} {{ record.status_code < 400 ? 'OK' : 'Error' }}
                </Badge>
              </DialogTitle>

              <!-- Trace ID with Copy Button -->
              <DialogDescription class="text-xs font-mono text-zinc-400 truncate flex items-center gap-1.5 mt-1">
                <span>{{ record?.id }}</span>
                <button
                  type="button"
                  @click="copyId"
                  class="text-zinc-500 hover:text-zinc-300 transition-colors p-0.5 rounded cursor-pointer"
                  title="Copy trace ID"
                >
                  <Check v-if="copiedId" class="h-3 w-3 text-emerald-400" />
                  <Copy v-else class="h-3 w-3" />
                </button>
              </DialogDescription>
            </div>
          </div>
        </div>
      </DialogHeader>

      <!-- Modal Body -->
      <div v-if="record" class="p-6 overflow-y-auto max-h-[65vh] space-y-4 text-xs font-mono">
        <!-- Error Alert Banner -->
        <div
          v-if="record.status_code >= 400 || record.error_message"
          class="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-300 space-y-1 font-sans"
        >
          <div class="font-semibold text-xs flex items-center gap-1.5">
            <AlertCircle class="h-3.5 w-3.5 text-red-400" />
            <span>Error Status {{ record.status_code }}</span>
          </div>
          <p class="text-xs font-mono text-red-200/90 break-all leading-relaxed">
            {{ record.error_message || 'The upstream CloudCode API returned an error response.' }}
          </p>
        </div>

        <!-- 3 KPI Metric Cards -->
        <div class="grid grid-cols-3 gap-3">
          <div class="p-3.5 rounded-xl bg-[#0d0e11] border border-zinc-800/80 space-y-1">
            <span class="text-[10px] text-zinc-500 uppercase flex items-center gap-1.5 font-sans font-medium">
              <Clock class="h-3 w-3 text-zinc-400" /> Latency
            </span>
            <div class="text-base font-semibold text-white font-mono">{{ ((record.latency_ms ?? record.duration_ms) ?? 0).toLocaleString() }} ms</div>
          </div>

          <div class="p-3.5 rounded-xl bg-[#0d0e11] border border-zinc-800/80 space-y-1">
            <span class="text-[10px] text-zinc-500 uppercase flex items-center gap-1.5 font-sans font-medium">
              <Cpu class="h-3 w-3 text-blue-400" /> Model
            </span>
            <div class="text-xs font-semibold text-blue-400 truncate font-mono pt-0.5">{{ record.model }}</div>
          </div>

          <div class="p-3.5 rounded-xl bg-[#0d0e11] border border-zinc-800/80 space-y-1">
            <span class="text-[10px] text-zinc-500 uppercase flex items-center gap-1.5 font-sans font-medium">
              <HardDrive class="h-3 w-3 text-emerald-400" /> Tokens
            </span>
            <div class="text-base font-semibold text-emerald-400 font-mono">{{ (record.total_tokens ?? 0).toLocaleString() }}</div>
          </div>
        </div>

        <!-- Detailed Breakdown Card -->
        <div class="p-4 rounded-xl bg-[#0d0e11] border border-zinc-800/80 space-y-2.5">
          <div class="flex items-center justify-between py-1">
            <span class="text-zinc-500 font-sans">Timestamp</span>
            <span class="text-zinc-200">{{ record.timestamp }}</span>
          </div>
          <Separator class="bg-zinc-800/60" />

          <div class="flex items-center justify-between py-1">
            <span class="text-zinc-500 font-sans">HTTP Endpoint</span>
            <div class="flex items-center gap-1.5">
              <span class="px-1.5 py-0.2 rounded text-[10px] font-mono font-semibold bg-blue-500/10 text-blue-400 border border-blue-500/20">POST</span>
              <span class="text-zinc-200">{{ record.endpoint }}</span>
            </div>
          </div>
          <Separator class="bg-zinc-800/60" />

          <div class="flex items-center justify-between py-1">
            <span class="text-zinc-500 font-sans">Streaming Protocol</span>
            <Badge
              variant="outline"
              class="font-mono text-[10px] gap-1"
              :class="(record.is_stream || record.stream) ? 'text-blue-400 border-blue-500/30 bg-blue-500/10' : 'text-zinc-400 border-zinc-800'"
            >
              <Radio v-if="record.is_stream || record.stream" class="h-2.5 w-2.5 text-blue-400 animate-pulse" />
              {{ (record.is_stream || record.stream) ? 'Server-Sent Events (SSE)' : 'Standard JSON' }}
            </Badge>
          </div>
          <Separator class="bg-zinc-800/60" />

          <div class="flex items-center justify-between py-1">
            <span class="text-zinc-500 font-sans">Prompt Tokens</span>
            <span class="text-zinc-200">{{ (record.prompt_tokens ?? 0).toLocaleString() }}</span>
          </div>
          <Separator class="bg-zinc-800/60" />

          <div class="flex items-center justify-between py-1">
            <span class="text-zinc-500 font-sans">Completion Tokens</span>
            <span class="text-zinc-200">{{ (record.completion_tokens ?? 0).toLocaleString() }}</span>
          </div>

          <template v-if="record.client_ip">
            <Separator class="bg-zinc-800/60" />
            <div class="flex items-center justify-between py-1">
              <span class="text-zinc-500 font-sans flex items-center gap-1">
                <Globe class="h-3 w-3 text-zinc-500" /> Client IP
              </span>
              <span class="text-zinc-200">{{ record.client_ip }}</span>
            </div>
          </template>
        </div>
      </div>

      <!-- Modal Footer -->
      <DialogFooter class="px-6 py-3 border-t border-zinc-800/80 bg-zinc-900/40 flex items-center justify-between sm:justify-between">
        <Button
          variant="outline"
          size="sm"
          @click="copyJson"
          class="h-8 text-xs gap-1.5 bg-zinc-900 border-zinc-800 hover:bg-zinc-800 text-zinc-300"
        >
          <Check v-if="copiedJson" class="h-3 w-3 text-emerald-400" />
          <Copy v-else class="h-3 w-3" />
          <span>{{ copiedJson ? 'JSON Copied' : 'Copy JSON' }}</span>
        </Button>

        <Button
          variant="secondary"
          size="sm"
          @click="emit('close')"
          class="h-8 text-xs bg-zinc-800 hover:bg-zinc-700 text-white"
        >
          Close
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
