<script setup lang="ts">
import { AlertCircle, CheckCircle2, Clock, Cpu, HardDrive } from '@lucide/vue'
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

defineProps<{
  record: RequestRecord | null
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()
</script>

<template>
  <Dialog :open="open && !!record" @update:open="val => { if (!val) emit('close') }">
    <DialogContent class="max-w-2xl bg-[#1a1c22] border-[#2c2e36] p-0 overflow-hidden gap-0">
      <!-- Modal Header -->
      <DialogHeader class="px-6 py-4 border-b border-[#2c2e36] bg-[#141518]">
        <div class="flex items-center gap-3">
          <div
            v-if="record"
            class="h-8 w-8 rounded-lg flex items-center justify-center shrink-0"
            :class="record.status_code < 400 ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'"
          >
            <CheckCircle2 v-if="record.status_code < 400" class="h-4 w-4" />
            <AlertCircle v-else class="h-4 w-4" />
          </div>
          <div class="min-w-0">
            <DialogTitle class="text-sm font-semibold text-white">Request Trace Details</DialogTitle>
            <DialogDescription class="text-xs font-mono text-zinc-500 truncate max-w-md mt-0.5">
              {{ record?.id }}
            </DialogDescription>
          </div>
        </div>
      </DialogHeader>

      <!-- Modal Body -->
      <div v-if="record" class="p-6 overflow-y-auto max-h-[60vh] space-y-5 text-xs font-mono">
        <!-- Error Banner if failed -->
        <div
          v-if="record.status_code >= 400 || record.error_message"
          class="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-300 space-y-1 font-sans"
        >
          <div class="font-semibold text-xs flex items-center gap-1.5">
            <AlertCircle class="h-3.5 w-3.5 text-red-400" />
            <span>Error Status {{ record.status_code }}</span>
          </div>
          <p class="text-xs font-mono text-red-200/90 break-all">
            {{ record.error_message || 'The upstream CloudCode API returned an error response.' }}
          </p>
        </div>

        <!-- Metrics Grid -->
        <div class="grid grid-cols-2 sm:grid-cols-3 gap-3">
          <div class="p-3 rounded-lg bg-[#202227] border border-[#2c2e36]">
            <span class="text-[10px] text-zinc-500 uppercase flex items-center gap-1 font-sans font-medium">
              <Clock class="h-3 w-3" /> Latency
            </span>
            <div class="mt-1 text-sm font-semibold text-white font-mono">{{ record.latency_ms }} ms</div>
          </div>

          <div class="p-3 rounded-lg bg-[#202227] border border-[#2c2e36]">
            <span class="text-[10px] text-zinc-500 uppercase flex items-center gap-1 font-sans font-medium">
              <Cpu class="h-3 w-3" /> Model
            </span>
            <div class="mt-1 text-sm font-semibold text-blue-400 truncate font-mono">{{ record.model }}</div>
          </div>

          <div class="p-3 rounded-lg bg-[#202227] border border-[#2c2e36]">
            <span class="text-[10px] text-zinc-500 uppercase flex items-center gap-1 font-sans font-medium">
              <HardDrive class="h-3 w-3" /> Total Tokens
            </span>
            <div class="mt-1 text-sm font-semibold text-emerald-400 font-mono">{{ record.total_tokens.toLocaleString() }}</div>
          </div>
        </div>

        <!-- Detailed Breakdown -->
        <div class="p-4 rounded-xl bg-[#202227] border border-[#2c2e36] space-y-2.5">
          <div class="flex justify-between py-1">
            <span class="text-zinc-500 font-sans">Timestamp</span>
            <span class="text-zinc-200">{{ record.timestamp }}</span>
          </div>
          <Separator />
          <div class="flex justify-between py-1">
            <span class="text-zinc-500 font-sans">HTTP Endpoint</span>
            <span class="text-zinc-200">{{ record.endpoint }}</span>
          </div>
          <Separator />
          <div class="flex justify-between py-1">
            <span class="text-zinc-500 font-sans">Streaming (SSE)</span>
            <Badge variant="outline" class="font-mono text-[10px]">
              {{ record.is_stream ? 'Yes (text/event-stream)' : 'No (Standard JSON)' }}
            </Badge>
          </div>
          <Separator />
          <div class="flex justify-between py-1">
            <span class="text-zinc-500 font-sans">Prompt Tokens</span>
            <span class="text-zinc-200">{{ record.prompt_tokens.toLocaleString() }}</span>
          </div>
          <Separator />
          <div class="flex justify-between py-1">
            <span class="text-zinc-500 font-sans">Completion Tokens</span>
            <span class="text-zinc-200">{{ record.completion_tokens.toLocaleString() }}</span>
          </div>
          <template v-if="record.client_ip">
            <Separator />
            <div class="flex justify-between py-1">
              <span class="text-zinc-500 font-sans">Client IP</span>
              <span class="text-zinc-200">{{ record.client_ip }}</span>
            </div>
          </template>
        </div>
      </div>

      <!-- Modal Footer -->
      <DialogFooter class="px-6 py-3 border-t border-[#2c2e36] bg-[#141518]">
        <Button
          variant="secondary"
          size="sm"
          @click="emit('close')"
        >
          Close
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
