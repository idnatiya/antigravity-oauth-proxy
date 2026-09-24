<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { Users, Plus, LogOut, CheckCircle2, AlertCircle, ExternalLink, RefreshCw } from '@lucide/vue'
import { apiClient } from '@/services/apiClient'
import { useUsageStore } from '@/stores/usageStore'

const usageStore = useUsageStore()

interface QuotaBucket {
  bucketId?: string
  displayName?: string
  window?: string
  remainingFraction?: number
  resetTime?: string
}

interface QuotaGroup {
  displayName: string
  buckets: QuotaBucket[]
}

interface Account {
  id: string
  projectId: string
  removable: boolean
  coolingUntil?: string
  coolingReason?: string
  quota?: { groups: QuotaGroup[] }
  quotaError?: string
}

interface AuthStatus {
  status: string
  authorizationUrl?: string
}

const accounts = ref<Account[]>([])
const loaded = ref(false)
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
  loaded.value = true
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

function groupLabel(name: string) {
  const lower = name.toLowerCase()
  if (lower.startsWith('gemini')) return 'Gemini'
  if (lower.includes('claude')) return 'Claude & GPT'
  return name
}

const windowOrder: Record<string, number> = { '5h': 0, weekly: 1 }

function sortedBuckets(buckets: QuotaBucket[]) {
  return [...buckets].sort((a, b) => (windowOrder[a.window ?? ''] ?? 9) - (windowOrder[b.window ?? ''] ?? 9))
}

function windowLabel(bucket: QuotaBucket) {
  if (bucket.window === '5h') return '5-hour'
  if (bucket.window === 'weekly') return 'Weekly'
  return bucket.displayName || bucket.bucketId || 'Limit'
}

function percent(bucket: QuotaBucket) {
  return bucket.remainingFraction === undefined ? null : Math.round(bucket.remainingFraction * 100)
}

function barColor(pct: number) {
  if (pct >= 50) return 'bg-emerald-500'
  if (pct >= 20) return 'bg-amber-500'
  return 'bg-red-500'
}

function shortDuration(iso?: string) {
  if (!iso) return ''
  const ms = new Date(iso).getTime() - Date.now()
  if (Number.isNaN(ms)) return ''
  if (ms <= 0) return 'now'
  const mins = Math.floor(ms / 60000)
  const days = Math.floor(mins / 1440)
  const hours = Math.floor((mins % 1440) / 60)
  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${mins % 60}m`
  return `${mins}m`
}

function localTime(iso?: string) {
  return iso ? new Date(iso).toLocaleString() : ''
}

onMounted(() => run(load))
</script>

<template>
  <div class="space-y-6 max-w-6xl">
    <div class="flex items-start justify-between gap-4">
      <div>
        <h2 class="text-base font-semibold text-white tracking-tight">Google Accounts</h2>
        <p class="text-xs text-zinc-500">Requests use the first ready account; accounts that run out of quota are skipped until they reset</p>
      </div>
      <div class="shrink-0 flex items-center gap-2">
        <button
          :disabled="busy"
          @click="run(load)"
          title="Refresh usage"
          class="p-2 rounded-lg border border-[#2c2e36] text-zinc-400 hover:text-zinc-200 hover:bg-[#1a1c22] disabled:opacity-50 cursor-pointer"
        >
          <RefreshCw class="h-3.5 w-3.5" :class="busy ? 'animate-spin' : ''" />
        </button>
        <button
          v-if="webLoginEnabled"
          :disabled="busy"
          @click="startLogin"
          class="flex items-center gap-1.5 px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-xs font-medium transition-colors cursor-pointer"
        >
          <Plus class="h-3.5 w-3.5" />
          Add account
        </button>
      </div>
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
          class="flex-1 min-w-0 bg-[#18191d] border border-[#2c2e36] focus:border-blue-500 rounded-lg px-3 py-2 text-zinc-200 font-mono placeholder:text-zinc-600 focus:outline-none transition-colors"
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

    <!-- Account cards -->
    <div
      v-if="loaded && !accounts.length"
      class="p-5 rounded-xl bg-[#202227] border border-[#2c2e36] text-xs text-zinc-500 flex items-center gap-2"
    >
      <Users class="h-4 w-4" />
      No accounts. Add one to start serving requests.
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <template v-if="!loaded">
        <div v-for="n in 2" :key="n" class="h-56 rounded-xl bg-[#202227] border border-[#2c2e36] animate-pulse" />
      </template>

      <div
        v-for="(acc, i) in accounts"
        :key="acc.id"
        class="rounded-xl bg-[#202227] border border-[#2c2e36] text-xs overflow-hidden"
      >
        <!-- Header -->
        <div class="px-5 py-4 flex items-start gap-3">
          <span
            class="mt-0.5 h-6 w-6 shrink-0 rounded-full bg-blue-500/10 border border-blue-500/20 text-blue-400 text-[11px] font-mono flex items-center justify-center"
            title="Priority"
          >{{ i + 1 }}</span>
          <div class="flex-1 min-w-0">
            <div class="text-sm text-zinc-100 font-medium truncate" :title="acc.id">{{ acc.id }}</div>
            <div class="text-[11px] text-zinc-500 font-mono truncate" :title="acc.projectId">{{ acc.projectId }}</div>
          </div>
          <div class="shrink-0 flex items-center gap-1.5">
            <span
              v-if="acc.coolingUntil"
              :title="`${acc.coolingReason ?? ''} (until ${localTime(acc.coolingUntil)})`"
              class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/20"
            >
              <span class="h-1.5 w-1.5 rounded-full bg-amber-400" />
              Cooling {{ shortDuration(acc.coolingUntil) }}
            </span>
            <span
              v-else
              class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
            >
              <span class="h-1.5 w-1.5 rounded-full bg-emerald-400" />
              Ready
            </span>
            <button
              v-if="acc.removable"
              :disabled="busy"
              @click="removeAccount(acc.id)"
              title="Log out"
              class="p-1.5 rounded-md text-zinc-500 hover:text-red-400 hover:bg-red-500/10 disabled:opacity-50 cursor-pointer"
            >
              <LogOut class="h-3.5 w-3.5" />
            </button>
          </div>
        </div>

        <!-- Usage -->
        <div
          v-if="acc.quota?.groups?.length"
          class="px-5 py-4 border-t border-[#2a2d34] grid grid-cols-[3.5rem_1fr_2.75rem_5.5rem] items-center gap-x-3 gap-y-2"
        >
          <template v-for="(group, gi) in acc.quota.groups" :key="group.displayName">
            <div
              class="col-span-4 text-[10px] uppercase tracking-wider font-semibold text-zinc-500"
              :class="gi > 0 ? 'mt-2' : ''"
              :title="group.displayName"
            >{{ groupLabel(group.displayName) }}</div>
            <template v-for="bucket in sortedBuckets(group.buckets)" :key="bucket.bucketId || bucket.window">
              <span class="text-zinc-400">{{ windowLabel(bucket) }}</span>
              <div class="h-2 rounded-full bg-[#18191d] overflow-hidden">
                <div
                  v-if="percent(bucket) !== null"
                  class="h-full rounded-full transition-all"
                  :class="barColor(percent(bucket)!)"
                  :style="{ width: `${percent(bucket)}%` }"
                />
              </div>
              <span class="text-right font-mono tabular-nums text-zinc-200">
                {{ percent(bucket) === null ? '–' : `${percent(bucket)}%` }}
              </span>
              <span
                class="text-right font-mono tabular-nums text-zinc-500 whitespace-nowrap"
                :title="bucket.resetTime ? `Resets ${localTime(bucket.resetTime)}` : ''"
              >{{ bucket.resetTime ? `↻ ${shortDuration(bucket.resetTime)}` : '' }}</span>
            </template>
          </template>
        </div>
        <div
          v-else-if="acc.quotaError"
          class="px-5 py-3 border-t border-[#2a2d34] text-red-300 flex items-center gap-2"
        >
          <AlertCircle class="h-3.5 w-3.5 shrink-0 text-red-400" />
          <span class="truncate" :title="acc.quotaError">Couldn't read usage: {{ acc.quotaError }}</span>
        </div>
      </div>
    </div>
  </div>
</template>
