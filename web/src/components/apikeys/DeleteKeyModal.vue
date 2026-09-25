<script setup lang="ts">
import { Trash2, Loader2, ShieldAlert } from '@lucide/vue'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import type { APIKey } from '@/types'

defineProps<{
  keyItem: APIKey | null
  open: boolean
  busy: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'confirm'): void
}>()
</script>

<template>
  <Dialog :open="open && !!keyItem" @update:open="(val) => { if (!val) emit('close') }">
    <DialogContent v-if="keyItem" class="max-w-md p-0 overflow-hidden border-[#2c2e36] bg-[#1a1c22]">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-[#2c2e36] flex items-center gap-3 bg-[#141518]">
        <div class="h-9 w-9 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 flex items-center justify-center shrink-0">
          <Trash2 class="h-4.5 w-4.5" />
        </div>
        <div>
          <DialogTitle class="text-sm font-semibold text-white">Revoke API Key</DialogTitle>
          <DialogDescription class="text-xs text-zinc-400 mt-0.5">Permanently delete and invalidate this key</DialogDescription>
        </div>
      </div>

      <!-- Body -->
      <div class="p-6 space-y-4 text-xs">
        <p class="text-zinc-300 leading-relaxed">
          Are you sure you want to revoke and delete <span class="font-medium text-white">{{ keyItem.name }}</span>?
        </p>

        <!-- Summary -->
        <div class="p-3.5 rounded-xl bg-[#202227] border border-[#2c2e36] space-y-1.5 font-mono">
          <div class="flex justify-between items-center text-zinc-400">
            <span class="text-[11px] text-zinc-500 uppercase tracking-wider font-sans">Key Name</span>
            <span class="text-zinc-200 truncate max-w-[240px] font-sans font-medium">{{ keyItem.name }}</span>
          </div>
          <div class="flex justify-between items-center text-zinc-400">
            <span class="text-[11px] text-zinc-500 uppercase tracking-wider font-sans">Token Prefix</span>
            <span class="text-blue-400 truncate max-w-[240px]">{{ keyItem.key_prefix }}</span>
          </div>
        </div>

        <!-- Warning info note -->
        <div class="p-3.5 rounded-xl bg-red-500/10 border border-red-500/20 text-red-300 flex items-start gap-2.5">
          <ShieldAlert class="h-4 w-4 text-red-400 shrink-0 mt-0.5" />
          <p class="text-[11px] leading-relaxed text-red-200/90 font-sans">
            Any client, IDE, or tool using this API key will immediately lose access and receive a <strong class="text-red-100 font-semibold">401 Unauthorized</strong> error. This action cannot be undone.
          </p>
        </div>
      </div>

      <!-- Footer -->
      <DialogFooter class="px-6 py-3.5 border-t border-[#2c2e36] bg-[#141518] flex items-center justify-end gap-2.5">
        <Button
          type="button"
          variant="outline"
          :disabled="busy"
          @click="emit('close')"
        >
          Cancel
        </Button>
        <Button
          type="button"
          variant="destructive"
          :disabled="busy"
          @click="emit('confirm')"
          class="gap-1.5"
        >
          <Loader2 v-if="busy" class="h-3.5 w-3.5 animate-spin" />
          <Trash2 v-else class="h-3.5 w-3.5" />
          <span>{{ busy ? 'Revoking...' : 'Revoke Key' }}</span>
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
