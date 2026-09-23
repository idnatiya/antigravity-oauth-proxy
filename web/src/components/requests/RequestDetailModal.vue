<script setup lang="ts">
import { X, AlertCircle, CheckCircle2, Clock, Cpu, HardDrive } from '@lucide/vue'
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
  <div
    v-if="open && record"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/70 backdrop-blur-sm"
    @click.self="emit('close')"
  >
    <div class="w-full max-w-2xl bg-[#1a1c22] border border-[#2c2e36] rounded-2xl shadow-2xl overflow-hidden flex flex-col max-h-[85vh]">
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-[#2c2e36] flex items-center justify-between bg-[#141518]">
        <div class="flex items-center gap-3">
          <div
            class="h-8 w-8 rounded-lg flex items-center justify-center"
            :class="record.status_code < 400 ? 'bg-emerald-500/10 text-emerald-400' : 'bg-red-500/10 text-red-400'"
          >
            <CheckCircle2 v-if="record.status_code < 400" class="h-4 w-4" />
            <AlertCircle v-else class="h-4 w-4" />
          </div>
          <div>
            <h3 class="text-sm font-semibold text-white">Request Trace Details</h3>
            <p class="text-xs font-mono text-zinc-500 truncate max-w-md">{{ record.id }}</p>
          </div>
        </div>

        <button
          @click="emit('close')"
          class="p-1.5 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800 transition-colors cursor-pointer"
        >
          <X class="h-4 w-4" />
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-6 overflow-y-auto space-y-6 text-xs font-mono">
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
            <span class="text-[10px] text-zinc-500 uppercase flex items-center gap-1">
              <Clock class="h-3 w-3" /> Latency
            </span>
            <div class="mt-1 text-sm font-semibold text-white">{{ record.latency_ms }} ms</div>
          </div>

          <div class="p-3 rounded-lg bg-[#202227] border border-[#2c2e36]">
            <span class="text-[10px] text-zinc-500 uppercase flex items-center gap-1">
              <Cpu class="h-3 w-3" /> Model
            </span>
            <div class="mt-1 text-sm font-semibold text-blue-400 truncate">{{ record.model }}</div>
          </div>

          <div class="p-3 rounded-lg bg-[#202227] border border-[#2c2e36]">
            <span class="text-[10px] text-zinc-500 uppercase flex items-center gap-1">
              <HardDrive class="h-3 w-3" /> Total Tokens
            </span>
            <div class="mt-1 text-sm font-semibold text-emerald-400">{{ record.total_tokens.toLocaleString() }}</div>
          </div>
        </div>

        <!-- Detailed Breakdown -->
        <div class="p-4 rounded-xl bg-[#202227] border border-[#2c2e36] space-y-2.5">
          <div class="flex justify-between py-1 border-b border-[#2c2e36]">
            <span class="text-zinc-500">Timestamp</span>
            <span class="text-zinc-200">{{ record.timestamp }}</span>
          </div>
          <div class="flex justify-between py-1 border-b border-[#2c2e36]">
            <span class="text-zinc-500">HTTP Endpoint</span>
            <span class="text-zinc-200">{{ record.endpoint }}</span>
          </div>
          <div class="flex justify-between py-1 border-b border-[#2c2e36]">
            <span class="text-zinc-500">Streaming (SSE)</span>
            <span class="text-zinc-200">{{ record.is_stream ? 'Yes (text/event-stream)' : 'No (Standard JSON)' }}</span>
          </div>
          <div class="flex justify-between py-1 border-b border-[#2c2e36]">
            <span class="text-zinc-500">Prompt Tokens</span>
            <span class="text-zinc-200">{{ record.prompt_tokens.toLocaleString() }}</span>
          </div>
          <div class="flex justify-between py-1 border-b border-[#2c2e36]">
            <span class="text-zinc-500">Completion Tokens</span>
            <span class="text-zinc-200">{{ record.completion_tokens.toLocaleString() }}</span>
          </div>
          <div v-if="record.client_ip" class="flex justify-between py-1">
            <span class="text-zinc-500">Client IP</span>
            <span class="text-zinc-200">{{ record.client_ip }}</span>
          </div>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="px-6 py-3 border-t border-[#2c2e36] bg-[#141518] flex justify-end">
        <button
          @click="emit('close')"
          class="px-4 py-1.5 rounded-lg bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs font-medium transition-colors cursor-pointer"
        >
          Close
        </button>
      </div>
    </div>
  </div>
</template>
