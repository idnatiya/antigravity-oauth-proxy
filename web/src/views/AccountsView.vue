<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Users, Plus, LogOut, CheckCircle2, AlertCircle, ExternalLink } from '@lucide/vue'
import { apiClient } from '@/services/apiClient'
import { useUsageStore } from '@/stores/usageStore'

const usageStore = useUsageStore()

interface Account {
  id: string
  projectId: string
  removable: boolean
  coolingUntil?: string
  coolingReason?: string
}

interface AuthStatus {
  status: string
  authorizationUrl?: string
}

const accounts = ref<Account[]>([])
const webLoginEnabled = ref(false)
const authorizationUrl = ref('')
const redirectInput = ref('')
const busy = ref(false)
const successMessage = ref('')
const errorMessage = ref('')

async function run(action: () => Promise<void>) {
  errorMessage.value = ''
  successMessage.value = ''
  busy.value = true
  try {
    await action()
  } catch (err: unknown) {
    errorMessage.value = err instanceof Error ? err.message : 'Request failed.'
  } finally {
    busy.value = false
  }
}

async function load() {
  const res = await apiClient.get<{ accounts: Account[]; webLoginEnabled: boolean }>('/api/accounts')
  accounts.value = res.accounts || []
  webLoginEnabled.value = res.webLoginEnabled
  usageStore.fetchStats().catch(() => {}) // keep sidebar account count in sync
}

function startLogin() {
  return run(async () => {
    const res = await apiClient.post<AuthStatus>('/api/accounts/auth/start')
    authorizationUrl.value = res.authorizationUrl || ''
    if (authorizationUrl.value) window.open(authorizationUrl.value, '_blank', 'noopener')
  })
}

function completeLogin() {
  return run(async () => {
    await apiClient.post<AuthStatus>('/api/accounts/auth/status', { code: redirectInput.value.trim() })
    authorizationUrl.value = ''
    redirectInput.value = ''
    successMessage.value = 'Account added.'
    await load()
  })
}

function removeAccount(id: string) {
  if (!window.confirm(`Log out ${id}? Its stored tokens will be deleted from this server.`)) return
  return run(async () => {
    await apiClient.delete('/api/accounts', { params: { id } })
    await load()
  })
}

function formatTime(iso: string) {
  return new Date(iso).toLocaleString()
}

onMounted(() => run(load))
</script>

<template>
  <div class="space-y-6 max-w-4xl">
    <div class="flex items-start justify-between gap-4">
      <div>
        <h2 class="text-base font-semibold text-white tracking-tight">Google Accounts</h2>
        <p class="text-xs text-zinc-500">Requests use the first ready account; accounts that run out of quota are skipped until they reset</p>
      </div>
      <button
        v-if="webLoginEnabled"
        :disabled="busy"
        @click="startLogin"
        class="shrink-0 flex items-center gap-1.5 px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-xs font-medium transition-colors cursor-pointer"
      >
        <Plus class="h-3.5 w-3.5" />
        Add account
      </button>
    </div>

    <div
      v-if="successMessage"
      class="p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-300 text-xs flex items-center gap-2"
    >
      <CheckCircle2 class="h-4 w-4 text-emerald-400" />
      <span>{{ successMessage }}</span>
    </div>
    <div
      v-if="errorMessage"
      class="p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-red-300 text-xs flex items-center gap-2"
    >
      <AlertCircle class="h-4 w-4 text-red-400" />
      <span>{{ errorMessage }}</span>
    </div>

    <!-- Pending login -->
    <div v-if="authorizationUrl" class="p-5 rounded-xl bg-[#202227] border border-[#2c2e36] space-y-3 text-xs">
      <ol class="list-decimal list-inside space-y-1 text-zinc-300">
        <li>
          Sign in with Google in the
          <a :href="authorizationUrl" target="_blank" rel="noopener" class="text-blue-400 hover:underline inline-flex items-center gap-1">
            opened tab <ExternalLink class="h-3 w-3" />
          </a>.
        </li>
        <li>Google redirects to <span class="font-mono">localhost:51121/…</span>, which will fail to load. That's expected.</li>
        <li>Copy the full URL from that tab's address bar and paste it below.</li>
      </ol>
      <form @submit.prevent="completeLogin" class="flex gap-2">
        <input
          v-model="redirectInput"
          required
          placeholder="http://localhost:51121/oauth-callback?state=…&code=…"
          class="flex-1 bg-[#18191d] border border-[#2c2e36] focus:border-blue-500 rounded-lg px-3 py-2 text-zinc-200 font-mono placeholder:text-zinc-600 focus:outline-none transition-colors"
        />
        <button
          type="submit"
          :disabled="busy"
          class="px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white font-medium transition-colors cursor-pointer"
        >
          {{ busy ? 'Verifying…' : 'Finish' }}
        </button>
      </form>
    </div>

    <!-- Account list -->
    <div class="rounded-xl bg-[#202227] border border-[#2c2e36] divide-y divide-[#2a2d34]">
      <div v-if="!accounts.length" class="p-5 text-xs text-zinc-500 flex items-center gap-2">
        <Users class="h-4 w-4" />
        No accounts. Add one to start serving requests.
      </div>
      <div v-for="(acc, i) in accounts" :key="acc.id" class="p-4 flex items-center gap-4 text-xs">
        <span class="w-5 text-zinc-500 font-mono">{{ i + 1 }}</span>
        <div class="flex-1 min-w-0">
          <div class="text-zinc-200 font-medium truncate">{{ acc.id }}</div>
          <div class="text-zinc-500 font-mono truncate">{{ acc.projectId }}</div>
        </div>
        <span
          v-if="acc.coolingUntil"
          :title="acc.coolingReason"
          class="px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/20 font-mono"
        >
          Cooling until {{ formatTime(acc.coolingUntil) }}
        </span>
        <span v-else class="px-2 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-mono">Ready</span>
        <button
          v-if="acc.removable"
          :disabled="busy"
          @click="removeAccount(acc.id)"
          title="Log out"
          class="p-1.5 rounded-md text-zinc-500 hover:text-red-400 hover:bg-red-500/10 disabled:opacity-50 cursor-pointer"
        >
          <LogOut class="h-4 w-4" />
        </button>
      </div>
    </div>
  </div>
</template>
