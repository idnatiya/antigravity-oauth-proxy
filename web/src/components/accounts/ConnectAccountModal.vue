<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import {
  ExternalLink,
  RefreshCw,
  CheckCircle2,
  AlertCircle,
  ChevronDown,
  ChevronUp,
  Sparkles,
  ArrowRight,
} from '@lucide/vue'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { apiClient } from '@/services/apiClient'

interface AuthStatusResponse {
  status: 'idle' | 'pending' | 'authenticated' | 'expired'
  authorizationUrl?: string
  expiresAt?: string
}

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'success'): void
}>()

const state = ref<'idle' | 'pending' | 'authenticated'>('idle')
const authorizationUrl = ref('')
const manualInput = ref('')
const showManualFallback = ref(false)
const loading = ref(false)
const errorMessage = ref('')
let pollInterval: ReturnType<typeof setInterval> | null = null

function stopPolling() {
  if (pollInterval) {
    clearInterval(pollInterval)
    pollInterval = null
  }
}

onUnmounted(() => {
  stopPolling()
})

watch(() => props.open, (isOpen) => {
  if (!isOpen) {
    stopPolling()
    reset()
  }
})

function reset() {
  state.value = 'idle'
  authorizationUrl.value = ''
  manualInput.value = ''
  showManualFallback.value = false
  loading.value = false
  errorMessage.value = ''
}

async function startAuth() {
  loading.value = true
  errorMessage.value = ''
  try {
    const res = await apiClient.post<AuthStatusResponse>('/api/accounts/auth/start')
    authorizationUrl.value = res.authorizationUrl || ''
    state.value = 'pending'
    
    if (authorizationUrl.value) {
      window.open(authorizationUrl.value, '_blank', 'noopener,noreferrer')
    }

    // Start polling status
    startPolling()
  } catch (err: unknown) {
    errorMessage.value = err instanceof Error ? err.message : 'Failed to start authorization.'
    state.value = 'idle'
  } finally {
    loading.value = false
  }
}

function startPolling() {
  stopPolling()
  pollInterval = setInterval(async () => {
    try {
      const res = await apiClient.get<AuthStatusResponse>('/api/accounts/auth/status')
      if (res.status === 'authenticated') {
        stopPolling()
        state.value = 'authenticated'
        setTimeout(() => {
          emit('success')
          emit('close')
        }, 1200)
      } else if (res.status === 'expired') {
        stopPolling()
        errorMessage.value = 'Authorization session expired. Please try again.'
        state.value = 'idle'
      }
    } catch {
      // Ignore polling errors
    }
  }, 1500)
}

async function submitManualCode() {
  const code = manualInput.value.trim()
  if (!code) return

  loading.value = true
  errorMessage.value = ''
  try {
    await apiClient.post<AuthStatusResponse>('/api/accounts/auth/status', { code })
    stopPolling()
    state.value = 'authenticated'
    setTimeout(() => {
      emit('success')
      emit('close')
    }, 1200)
  } catch (err: unknown) {
    errorMessage.value = err instanceof Error ? err.message : 'Invalid code or failed to verify.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="(val) => { if (!val) emit('close') }">
    <DialogContent class="max-w-md p-0 overflow-hidden border-[#2c2e36] bg-[#121316] text-zinc-100 shadow-2xl">
      <!-- Header -->
      <div class="px-6 py-5 border-b border-zinc-800/80 bg-[#16181d] flex items-center gap-3.5">
        <div class="h-10 w-10 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-400 flex items-center justify-center shrink-0 shadow-xs">
          <!-- Google 'G' icon -->
          <svg class="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
            <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
            <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z" fill="#FBBC05"/>
            <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z" fill="#EA4335"/>
          </svg>
        </div>
        <div>
          <DialogTitle class="text-sm font-semibold text-white">Connect Google Account</DialogTitle>
          <DialogDescription class="text-xs text-zinc-400 mt-0.5">
            Add account to Antigravity failover quota pool
          </DialogDescription>
        </div>
      </div>

      <!-- Body -->
      <div class="p-6 space-y-4 text-xs">
        <!-- Error Banner -->
        <div
          v-if="errorMessage"
          class="p-3.5 rounded-xl bg-red-500/10 border border-red-500/20 text-red-300 flex items-start gap-2.5"
        >
          <AlertCircle class="h-4 w-4 text-red-400 shrink-0 mt-0.5" />
          <span class="flex-1 leading-relaxed">{{ errorMessage }}</span>
        </div>

        <!-- STATE: Authenticated -->
        <div v-if="state === 'authenticated'" class="py-6 text-center space-y-3">
          <div class="h-12 w-12 rounded-full bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 mx-auto flex items-center justify-center animate-bounce">
            <CheckCircle2 class="h-6 w-6" />
          </div>
          <h4 class="text-sm font-semibold text-white">Account Connected Successfully!</h4>
          <p class="text-zinc-400 text-xs">The account has been added to your failover pool and is ready to serve requests.</p>
        </div>

        <!-- STATE: Pending (Waiting for Google sign-in) -->
        <div v-else-if="state === 'pending'" class="space-y-4">
          <div class="p-4 rounded-xl bg-[#16181d] border border-zinc-800/80 flex items-center gap-3.5">
            <RefreshCw class="h-5 w-5 text-blue-400 animate-spin shrink-0" />
            <div class="space-y-1">
              <div class="text-xs font-medium text-white">Waiting for Google authorization...</div>
              <div class="text-[11px] text-zinc-400 leading-relaxed">
                Approve access in the Google tab. This modal will automatically detect completion.
              </div>
            </div>
          </div>

          <div class="flex items-center justify-between text-[11px] text-zinc-400 px-1">
            <span>Tab did not open?</span>
            <a
              v-if="authorizationUrl"
              :href="authorizationUrl"
              target="_blank"
              rel="noopener noreferrer"
              class="text-blue-400 hover:text-blue-300 font-medium inline-flex items-center gap-1"
            >
              Open Google Sign-in <ExternalLink class="h-3 w-3" />
            </a>
          </div>

          <!-- Fallback accordion for manual code paste -->
          <div class="pt-2 border-t border-zinc-800/80">
            <button
              type="button"
              @click="showManualFallback = !showManualFallback"
              class="flex items-center justify-between w-full text-zinc-400 hover:text-zinc-200 text-[11px] py-1 transition-colors cursor-pointer"
            >
              <span>Connecting remotely or callback blocked?</span>
              <component :is="showManualFallback ? ChevronUp : ChevronDown" class="h-3.5 w-3.5 text-zinc-500" />
            </button>

            <div v-if="showManualFallback" class="mt-3 space-y-2.5">
              <p class="text-[11px] text-zinc-400 leading-relaxed">
                Copy the final URL or code from the browser address bar after signing in:
              </p>
              <div class="flex gap-2">
                <input
                  v-model="manualInput"
                  type="text"
                  placeholder="Paste URL or code (http://localhost:51121/...)"
                  class="flex-1 bg-[#0d0e11] border border-zinc-800 focus:border-blue-500 rounded-lg px-3 py-1.5 text-xs text-zinc-200 placeholder:text-zinc-600 focus:outline-none"
                  @keydown.enter.prevent="submitManualCode"
                />
                <Button
                  size="sm"
                  :disabled="loading || !manualInput.trim()"
                  @click="submitManualCode"
                  class="shrink-0 gap-1 bg-blue-600 hover:bg-blue-500 text-white"
                >
                  <RefreshCw v-if="loading" class="h-3 w-3 animate-spin" />
                  <ArrowRight v-else class="h-3 w-3" />
                  <span>Submit</span>
                </Button>
              </div>
            </div>
          </div>
        </div>

        <!-- STATE: Idle (Ready to start) -->
        <div v-else class="space-y-4">
          <div class="p-3.5 rounded-xl bg-blue-500/5 border border-blue-500/15 space-y-2">
            <div class="flex items-center gap-2 text-blue-400 font-medium text-xs">
              <Sparkles class="h-3.5 w-3.5" />
              <span>Multi-Account Quota Failover</span>
            </div>
            <p class="text-[11px] text-zinc-400 leading-relaxed">
              When an account exhausts its hourly or daily rate limit, Antigravity Proxy automatically fails over to the next available account without dropping requests.
            </p>
          </div>

          <div class="text-[11px] text-zinc-400 space-y-1.5">
            <p>Signing in will grant Cloud Code / Gemini Code Assist access to proxy requests on your behalf.</p>
          </div>

          <Button
            type="button"
            class="w-full gap-2.5 bg-blue-600 hover:bg-blue-500 text-white py-2.5 font-medium shadow-md transition-all cursor-pointer"
            :disabled="loading"
            @click="startAuth"
          >
            <RefreshCw v-if="loading" class="h-4 w-4 animate-spin" />
            <svg v-else class="h-4 w-4 shrink-0" viewBox="0 0 24 24" fill="currentColor">
              <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#fff"/>
              <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#fff"/>
              <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z" fill="#fff"/>
              <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z" fill="#fff"/>
            </svg>
            <span>Authorize with Google</span>
          </Button>
        </div>
      </div>

      <!-- Footer -->
      <DialogFooter class="px-6 py-3.5 border-t border-zinc-800/80 bg-[#16181d] flex items-center justify-between">
        <span class="text-[11px] text-zinc-500 font-mono">
          {{ state === 'pending' ? 'Callback: localhost:51121' : 'Secure OAuth 2.0 (PKCE)' }}
        </span>
        <Button
          type="button"
          variant="outline"
          size="sm"
          @click="emit('close')"
        >
          {{ state === 'authenticated' ? 'Close' : 'Cancel' }}
        </Button>
      </DialogFooter>
    </DialogContent>
  </Dialog>
</template>
