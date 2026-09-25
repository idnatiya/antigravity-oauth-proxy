<script setup lang="ts">
import { ref, watch } from 'vue'
import { KeyRound, Sparkles, Copy, Check, Eye, EyeOff, Loader2 } from '@lucide/vue'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { apiKeysService } from '@/services/apiKeysService'
import type { APIKey } from '@/types'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'created', key: APIKey): void
}>()

const name = ref('')
const customKey = ref('')
const isSubmitting = ref(false)
const error = ref('')
const createdKey = ref<APIKey | null>(null)
const showPlainKey = ref(false)
const copied = ref(false)

watch(
  () => props.open,
  (val) => {
    if (val) {
      name.value = ''
      customKey.value = ''
      error.value = ''
      createdKey.value = null
      showPlainKey.value = true
      copied.value = false
    }
  }
)

async function handleSubmit() {
  if (!name.value.trim()) {
    error.value = 'Key name is required'
    return
  }

  isSubmitting.value = true
  error.value = ''

  try {
    const key = await apiKeysService.createKey({
      name: name.value.trim(),
      key: customKey.value.trim() || undefined,
    })
    createdKey.value = key
    emit('created', key)
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : 'Failed to create API key'
  } finally {
    isSubmitting.value = false
  }
}

async function copyToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch {
    // Fallback
    const textarea = document.createElement('textarea')
    textarea.value = text
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="(val) => { if (!val) emit('close') }">
    <DialogContent class="max-w-md p-0 overflow-hidden border-[#2c2e36] bg-[#1a1c22]">
      <!-- Header -->
      <div class="px-6 py-4 border-b border-[#2c2e36] flex items-center gap-3 bg-[#141518]">
        <div class="h-9 w-9 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-400 flex items-center justify-center shrink-0">
          <KeyRound v-if="!createdKey" class="h-4.5 w-4.5" />
          <Sparkles v-else class="h-4.5 w-4.5 text-emerald-400" />
        </div>
        <div>
          <DialogTitle class="text-sm font-semibold text-white">
            {{ createdKey ? 'API Key Created' : 'Create New API Key' }}
          </DialogTitle>
          <DialogDescription class="text-xs text-zinc-400 mt-0.5">
            {{ createdKey ? 'Your key is ready to use with external AI clients' : 'Generate an API key to authenticate AI clients and tools' }}
          </DialogDescription>
        </div>
      </div>

      <!-- State 1: Creation Form -->
      <form v-if="!createdKey" @submit.prevent="handleSubmit" class="p-6 space-y-4 text-xs">
        <div v-if="error" class="p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-red-400">
          {{ error }}
        </div>

        <div class="space-y-1.5">
          <Label for="key-name" class="text-zinc-300 font-medium">Key Name / Description</Label>
          <Input
            id="key-name"
            v-model="name"
            placeholder="e.g. Cursor IDE, Claude Code, Cline, Work Laptop"
            required
            class="bg-[#0f1013] border-zinc-800 text-xs h-9 focus-visible:ring-blue-500 text-zinc-100 placeholder:text-zinc-600 rounded-lg"
          />
          <span class="text-[11px] text-zinc-500">A recognizable label for where this key is configured.</span>
        </div>

        <div class="space-y-1.5 pt-1">
          <Label for="custom-key" class="text-zinc-300 font-medium flex items-center justify-between">
            <span>Custom Secret Key (Optional)</span>
            <span class="text-[10px] text-zinc-500 font-normal">Leave blank to auto-generate</span>
          </Label>
          <Input
            id="custom-key"
            v-model="customKey"
            placeholder="Auto-generated (sk-agy-...)"
            class="bg-[#0f1013] border-zinc-800 font-mono text-xs h-9 focus-visible:ring-blue-500 text-zinc-100 placeholder:text-zinc-600 rounded-lg"
          />
          <span class="text-[11px] text-zinc-500">
            If left blank, a cryptographically secure <code class="font-mono text-blue-400">sk-agy-...</code> token will be created.
          </span>
        </div>

        <DialogFooter class="px-0 pt-4 border-t border-[#2c2e36] flex items-center justify-end gap-2.5">
          <Button
            type="button"
            variant="outline"
            :disabled="isSubmitting"
            @click="emit('close')"
          >
            Cancel
          </Button>
          <Button
            type="submit"
            :disabled="isSubmitting || !name.trim()"
            class="bg-blue-600 hover:bg-blue-500 text-white gap-2"
          >
            <Loader2 v-if="isSubmitting" class="h-3.5 w-3.5 animate-spin" />
            <KeyRound v-else class="h-3.5 w-3.5" />
            <span>{{ isSubmitting ? 'Creating...' : 'Create Key' }}</span>
          </Button>
        </DialogFooter>
      </form>

      <!-- State 2: Key Generated Success -->
      <div v-else class="p-6 space-y-4 text-xs">
        <div class="p-3.5 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-300 space-y-1">
          <div class="font-semibold text-emerald-400 flex items-center gap-1.5">
            <Check class="h-4 w-4" />
            <span>Key successfully generated!</span>
          </div>
          <p class="text-[11px] text-emerald-200/90 leading-relaxed font-sans">
            Copy this key now to configure in your AI extensions, IDEs, or client scripts.
          </p>
        </div>

        <div class="space-y-1.5">
          <div class="flex justify-between items-center text-zinc-400">
            <span class="font-medium text-zinc-300">API Key</span>
            <span class="text-[11px] font-mono text-blue-400">{{ createdKey.name }}</span>
          </div>

          <div class="relative flex items-center">
            <Input
              readonly
              :type="showPlainKey ? 'text' : 'password'"
              :value="createdKey.key"
              class="bg-[#0d0e11] border-zinc-800 font-mono text-xs h-10 pr-20 text-zinc-200 select-all rounded-lg"
            />
            <div class="absolute right-1.5 flex items-center gap-1">
              <button
                type="button"
                @click="showPlainKey = !showPlainKey"
                class="p-1.5 text-zinc-400 hover:text-zinc-200 rounded cursor-pointer transition-colors"
                :title="showPlainKey ? 'Hide key' : 'Show key'"
              >
                <EyeOff v-if="showPlainKey" class="h-3.5 w-3.5" />
                <Eye v-else class="h-3.5 w-3.5" />
              </button>
              <Button
                type="button"
                size="sm"
                @click="copyToClipboard(createdKey.key)"
                class="h-7 px-2.5 text-[11px] bg-blue-600 hover:bg-blue-500 text-white gap-1 rounded"
              >
                <Check v-if="copied" class="h-3 w-3 text-emerald-300" />
                <Copy v-else class="h-3 w-3" />
                <span>{{ copied ? 'Copied' : 'Copy' }}</span>
              </Button>
            </div>
          </div>
        </div>

        <!-- Quick Integration Example -->
        <div class="p-3 rounded-lg bg-[#0d0e11] border border-zinc-800 space-y-1.5 font-mono text-[11px]">
          <div class="text-zinc-400 font-sans text-[11px] font-medium">Header Example:</div>
          <div class="text-blue-300 select-all">
            Authorization: Bearer {{ createdKey.key }}
          </div>
        </div>

        <DialogFooter class="px-0 pt-2 border-t border-[#2c2e36] flex items-center justify-end">
          <Button
            type="button"
            class="bg-zinc-800 hover:bg-zinc-700 text-white text-xs px-5"
            @click="emit('close')"
          >
            Done
          </Button>
        </DialogFooter>
      </div>
    </DialogContent>
  </Dialog>
</template>
