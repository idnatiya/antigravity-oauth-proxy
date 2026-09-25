<script setup lang="ts">
import { computed } from 'vue'
import { Trash2, X, RefreshCw, ShieldAlert } from '@lucide/vue'
import type { AccountItem } from '@/types'

const props = defineProps<{
  account: AccountItem | null
  open: boolean
  busy: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'confirm'): void
}>()

const isRemovable = computed(() => props.account?.removable ?? false)
</script>

<template>
  <div
    v-if="open && account"
    class="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-xs animate-in fade-in duration-150"
    @click.self="emit('close')"
  >
    <div
      class="w-full max-w-md bg-[#1a1c22] border border-[#2c2e36] rounded-2xl shadow-2xl overflow-hidden flex flex-col"
    >
      <!-- Modal Header -->
      <div class="px-6 py-4 border-b border-[#2c2e36] flex items-center justify-between bg-[#141518]">
        <div class="flex items-center gap-3">
          <div class="h-9 w-9 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 flex items-center justify-center">
            <Trash2 class="h-4.5 w-4.5" />
          </div>
          <div>
            <h3 class="text-sm font-semibold text-white">Remove Account</h3>
            <p class="text-xs text-zinc-400">Disconnect account from proxy pool</p>
          </div>
        </div>

        <button
          @click="emit('close')"
          :disabled="busy"
          class="p-1.5 rounded-lg text-zinc-400 hover:text-white hover:bg-zinc-800 disabled:opacity-50 transition-colors cursor-pointer"
        >
          <X class="h-4 w-4" />
        </button>
      </div>

      <!-- Modal Body -->
      <div class="p-6 space-y-4 text-xs">
        <p class="text-zinc-300 leading-relaxed">
          Are you sure you want to remove <span class="font-medium text-white font-mono break-all">{{ account.id }}</span>?
        </p>

        <!-- Account summary card -->
        <div class="p-3.5 rounded-xl bg-[#202227] border border-[#2c2e36] space-y-1.5 font-mono">
          <div class="flex justify-between items-center text-zinc-400">
            <span class="text-[11px] text-zinc-500 uppercase tracking-wider font-sans">Account ID</span>
            <span class="text-zinc-200 truncate max-w-[240px]">{{ account.id }}</span>
          </div>
          <div class="flex justify-between items-center text-zinc-400">
            <span class="text-[11px] text-zinc-500 uppercase tracking-wider font-sans">Project ID</span>
            <span class="text-blue-400 truncate max-w-[240px]">{{ account.projectId || 'None' }}</span>
          </div>
        </div>

        <!-- Warning info note -->
        <div class="p-3.5 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-300 flex items-start gap-2.5">
          <ShieldAlert class="h-4 w-4 text-amber-400 shrink-0 mt-0.5" />
          <p class="text-[11px] leading-relaxed text-amber-200/90 font-sans">
            Its stored credentials will be removed locally from this server. This will <strong class="text-amber-100 font-semibold">not</strong> revoke your OAuth session on Google or sign you out of your Antigravity IDE.
          </p>
        </div>
      </div>

      <!-- Modal Footer -->
      <div class="px-6 py-3.5 border-t border-[#2c2e36] bg-[#141518] flex items-center justify-end gap-2.5">
        <button
          type="button"
          :disabled="busy"
          @click="emit('close')"
          class="px-4 py-2 rounded-lg bg-zinc-800 hover:bg-zinc-700 disabled:opacity-50 text-zinc-300 text-xs font-medium transition-colors cursor-pointer"
        >
          Cancel
        </button>
        <button
          type="button"
          :disabled="busy || !isRemovable"
          @click="emit('confirm')"
          class="flex items-center gap-1.5 px-4 py-2 rounded-lg bg-red-600 hover:bg-red-500 disabled:opacity-50 text-white text-xs font-medium transition-colors cursor-pointer shadow-sm"
        >
          <RefreshCw v-if="busy" class="h-3.5 w-3.5 animate-spin" />
          <Trash2 v-else class="h-3.5 w-3.5" />
          {{ busy ? 'Removing...' : 'Remove Account' }}
        </button>
      </div>
    </div>
  </div>
</template>
