<script setup lang="ts">
import { onMounted, ref, computed } from 'vue'
import {
  Users,
  Plus,
  CheckCircle2,
  AlertCircle,
  ExternalLink,
  RefreshCw,
  Search,
  AlertTriangle,
  Activity,
  ArrowRight,
  X,
  Zap,
} from '@lucide/vue'
import { apiClient } from '@/services/apiClient'
import { useUsageStore } from '@/stores/usageStore'
import StatCard from '@/components/overview/StatCard.vue'
import AccountCard from '@/components/accounts/AccountCard.vue'
import DeleteAccountModal from '@/components/accounts/DeleteAccountModal.vue'
import type { AccountItem, AccountTestResult } from '@/types'

const usageStore = useUsageStore()

interface AuthStatus {
  status: string
  authorizationUrl?: string
}

const accounts = ref<AccountItem[]>([])
const loaded = ref(false)
const webLoginEnabled = ref(false)
const authorizationUrl = ref('')
const redirectInput = ref('')
const busy = ref(false)
const successMessage = ref('')
const errorMessage = ref('')

// Filter & Search states
const searchQuery = ref('')
const statusFilter = ref<'all' | 'ready' | 'cooling'>('all')

// Deletion modal state
const accountToDelete = ref<AccountItem | null>(null)
const showDeleteModal = ref(false)

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
  const res = await apiClient.get<{ accounts: AccountItem[]; webLoginEnabled: boolean }>('/api/accounts')
  accounts.value = res.accounts || []
  webLoginEnabled.value = res.webLoginEnabled
  loaded.value = true
  usageStore.fetchStats().catch(() => {}) // keep sidebar account count in sync
}

function startLogin() {
  return run(async () => {
    const res = await apiClient.post<AuthStatus>('/api/accounts/auth/start')
    authorizationUrl.value = res.authorizationUrl || ''
    if (authorizationUrl.value) {
      window.open(authorizationUrl.value, '_blank', 'noopener')
    }
  })
}

function cancelLogin() {
  authorizationUrl.value = ''
  redirectInput.value = ''
}

function completeLogin() {
  return run(async () => {
    await apiClient.post<AuthStatus>('/api/accounts/auth/status', { code: redirectInput.value.trim() })
    authorizationUrl.value = ''
    redirectInput.value = ''
    successMessage.value = 'Account successfully added to the pool!'
    await load()
  })
}

function openDeleteModal(acc: AccountItem) {
  accountToDelete.value = acc
  showDeleteModal.value = true
}

function closeDeleteModal() {
  if (busy.value) return
  showDeleteModal.value = false
  accountToDelete.value = null
}

async function confirmDeleteAccount() {
  if (!accountToDelete.value) return
  const id = accountToDelete.value.id
  await run(async () => {
    await apiClient.delete('/api/accounts', { params: { id } })
    successMessage.value = `Account ${id} removed successfully.`
    closeDeleteModal()
    await load()
  })
}

// Account testing states
const testingAccountIds = ref<Set<string>>(new Set())
const testResults = ref<Record<string, AccountTestResult>>({})
const testingAll = ref(false)

async function testAccount(acc: AccountItem) {
  testingAccountIds.value.add(acc.id)
  try {
    const res = await apiClient.post<{ result: AccountTestResult }>('/api/accounts/test', null, {
      params: { id: acc.id },
    })
    testResults.value[acc.id] = res.result
  } catch (err: unknown) {
    testResults.value[acc.id] = {
      id: acc.id,
      projectId: acc.projectId,
      success: false,
      latencyMs: 0,
      error: err instanceof Error ? err.message : 'Test failed',
    }
  } finally {
    testingAccountIds.value.delete(acc.id)
  }
}

async function testAllAccounts() {
  if (testingAll.value || accounts.value.length === 0) return
  testingAll.value = true
  for (const a of accounts.value) {
    testingAccountIds.value.add(a.id)
  }
  try {
    const res = await apiClient.post<{ results: AccountTestResult[] }>('/api/accounts/test')
    for (const r of res.results || []) {
      testResults.value[r.id] = r
    }
    successMessage.value = `Connection test completed for ${res.results?.length || 0} accounts.`
  } catch (err: unknown) {
    errorMessage.value = err instanceof Error ? err.message : 'Failed to test accounts.'
  } finally {
    testingAccountIds.value.clear()
    testingAll.value = false
  }
}

// KPI Stats computation
const totalAccountsCount = computed(() => accounts.value.length)

const readyAccountsCount = computed(() => {
  return accounts.value.filter(a => !a.coolingUntil).length
})

const coolingAccountsCount = computed(() => {
  return accounts.value.filter(a => !!a.coolingUntil).length
})

const lowestQuotaRemaining = computed(() => {
  let minPct: number | null = null
  for (const acc of accounts.value) {
    if (acc.coolingUntil) continue
    for (const group of acc.quota?.groups || []) {
      for (const bucket of group.buckets || []) {
        if (bucket.remainingFraction !== undefined) {
          const pct = Math.round(bucket.remainingFraction * 100)
          if (minPct === null || pct < minPct) {
            minPct = pct
          }
        }
      }
    }
  }
  if (minPct === null) return '–'
  return `${minPct}%`
})

// Filtered Accounts
const filteredAccounts = computed(() => {
  const query = searchQuery.value.toLowerCase().trim()
  return accounts.value.filter((acc) => {
    // Status filter
    if (statusFilter.value === 'ready' && acc.coolingUntil) return false
    if (statusFilter.value === 'cooling' && !acc.coolingUntil) return false

    // Search query
    if (query) {
      const matchEmail = acc.id.toLowerCase().includes(query)
      const matchProject = acc.projectId?.toLowerCase().includes(query)
      if (!matchEmail && !matchProject) return false
    }

    return true
  })
})

onMounted(() => run(load))
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2.5">
          <h2 class="text-base font-semibold text-white tracking-tight">Google Accounts</h2>
          <span
            v-if="loaded"
            class="px-2 py-0.5 rounded-full text-[11px] font-mono font-medium bg-[#202227] border border-[#2c2e36] text-zinc-400"
          >
            {{ accounts.length }} total
          </span>
        </div>
        <p class="text-xs text-zinc-500 mt-0.5">
          Requests use the first ready account; accounts exceeding quota fail over seamlessly until reset.
        </p>
      </div>

      <div class="shrink-0 flex items-center gap-2.5">
        <button
          v-if="accounts.length > 0"
          :disabled="busy || testingAll"
          @click="testAllAccounts"
          title="Test connectivity for all accounts"
          class="flex items-center gap-1.5 px-3 py-2 rounded-lg border border-[#2c2e36] bg-[#202227] text-zinc-300 hover:text-white hover:bg-[#282a32] disabled:opacity-50 transition-colors text-xs font-medium cursor-pointer shadow-xs"
        >
          <RefreshCw v-if="testingAll" class="h-3.5 w-3.5 animate-spin text-amber-400" />
          <Zap v-else class="h-3.5 w-3.5 text-amber-400" />
          <span>{{ testingAll ? 'Testing...' : 'Test All' }}</span>
        </button>

        <button
          :disabled="busy || testingAll"
          @click="run(load)"
          title="Refresh account pool & quotas"
          class="flex items-center gap-1.5 px-3 py-2 rounded-lg border border-[#2c2e36] bg-[#202227] text-zinc-300 hover:text-white hover:bg-[#282a32] disabled:opacity-50 transition-colors text-xs font-medium cursor-pointer shadow-xs"
        >
          <RefreshCw class="h-3.5 w-3.5" :class="busy ? 'animate-spin' : ''" />
          <span>Refresh</span>
        </button>

        <button
          v-if="webLoginEnabled"
          :disabled="busy || testingAll"
          @click="startLogin"
          class="flex items-center gap-1.5 px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-xs font-medium transition-all shadow-sm shadow-blue-500/10 cursor-pointer"
        >
          <Plus class="h-3.5 w-3.5" />
          <span>Add Google Account</span>
        </button>
      </div>
    </div>

    <!-- Alert notifications -->
    <div
      v-if="successMessage"
      class="p-3.5 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-300 text-xs flex items-center justify-between"
    >
      <div class="flex items-center gap-2.5">
        <CheckCircle2 class="h-4 w-4 text-emerald-400 shrink-0" />
        <span>{{ successMessage }}</span>
      </div>
      <button @click="successMessage = ''" class="text-emerald-400/80 hover:text-emerald-200 cursor-pointer">
        <X class="h-3.5 w-3.5" />
      </button>
    </div>

    <div
      v-if="errorMessage"
      class="p-3.5 rounded-xl bg-red-500/10 border border-red-500/20 text-red-300 text-xs flex items-center justify-between"
    >
      <div class="flex items-center gap-2.5">
        <AlertCircle class="h-4 w-4 text-red-400 shrink-0" />
        <span>{{ errorMessage }}</span>
      </div>
      <button @click="errorMessage = ''" class="text-red-400/80 hover:text-red-200 cursor-pointer">
        <X class="h-3.5 w-3.5" />
      </button>
    </div>

    <!-- KPI Summary Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <StatCard
        title="Total Accounts"
        :value="totalAccountsCount"
        subtitle="Configured in failover pool"
        :icon="Users"
        color="blue"
      />

      <StatCard
        title="Ready to Serve"
        :value="readyAccountsCount"
        :subtitle="`${readyAccountsCount} active / ready`"
        :icon="CheckCircle2"
        color="green"
      />

      <StatCard
        title="In Cooldown"
        :value="coolingAccountsCount"
        :subtitle="coolingAccountsCount > 0 ? 'Recovering from rate limit' : 'All accounts operational'"
        :icon="AlertTriangle"
        :color="coolingAccountsCount > 0 ? 'amber' : 'green'"
      />

      <StatCard
        title="Lowest Quota Left"
        :value="lowestQuotaRemaining"
        subtitle="Lowest bucket in active pool"
        :icon="Activity"
        color="purple"
      />
    </div>

    <!-- Interactive Pending Login Wizard -->
    <div
      v-if="authorizationUrl"
      class="p-6 rounded-2xl bg-[#1b1d24] border border-blue-500/30 shadow-xl space-y-5 animate-in fade-in slide-in-from-top-2 duration-200"
    >
      <div class="flex items-start justify-between">
        <div class="space-y-1">
          <div class="flex items-center gap-2">
            <span class="flex h-2 w-2 rounded-full bg-blue-500 animate-pulse" />
            <h3 class="text-sm font-semibold text-white">Connect New Google Account</h3>
          </div>
          <p class="text-xs text-zinc-400">
            Complete the standard OAuth sign-in flow to authorize this proxy to query Gemini Code Assist.
          </p>
        </div>

        <button
          @click="cancelLogin"
          class="p-1 rounded-lg text-zinc-500 hover:text-zinc-300 hover:bg-zinc-800 transition-colors cursor-pointer"
          title="Cancel"
        >
          <X class="h-4 w-4" />
        </button>
      </div>

      <!-- Step Cards -->
      <div class="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
        <div class="p-3.5 rounded-xl bg-[#202227] border border-[#2c2e36] space-y-2">
          <div class="flex items-center gap-2">
            <span class="h-5 w-5 rounded-full bg-blue-500/20 text-blue-400 font-mono text-[11px] font-bold flex items-center justify-center">1</span>
            <span class="font-medium text-white">Open Sign-In</span>
          </div>
          <p class="text-zinc-400 text-[11px]">
            Sign in with the Google account in the newly opened browser tab.
          </p>
          <a
            :href="authorizationUrl"
            target="_blank"
            rel="noopener"
            class="inline-flex items-center gap-1.5 text-xs font-medium text-blue-400 hover:text-blue-300 hover:underline pt-1"
          >
            <span>Re-open Auth Tab</span>
            <ExternalLink class="h-3 w-3" />
          </a>
        </div>

        <div class="p-3.5 rounded-xl bg-[#202227] border border-[#2c2e36] space-y-2">
          <div class="flex items-center gap-2">
            <span class="h-5 w-5 rounded-full bg-blue-500/20 text-blue-400 font-mono text-[11px] font-bold flex items-center justify-center">2</span>
            <span class="font-medium text-white">Approve Access</span>
          </div>
          <p class="text-zinc-400 text-[11px] leading-relaxed">
            Google redirects to <span class="font-mono text-zinc-300 bg-zinc-800/80 px-1 py-0.5 rounded">localhost:51121/…</span>, which fails to load. That is normal.
          </p>
        </div>

        <div class="p-3.5 rounded-xl bg-[#202227] border border-[#2c2e36] space-y-2">
          <div class="flex items-center gap-2">
            <span class="h-5 w-5 rounded-full bg-blue-500/20 text-blue-400 font-mono text-[11px] font-bold flex items-center justify-center">3</span>
            <span class="font-medium text-white">Paste Callback URL</span>
          </div>
          <p class="text-zinc-400 text-[11px]">
            Copy the full URL from the address bar of that failed tab and submit it below.
          </p>
        </div>
      </div>

      <!-- Input Form -->
      <form @submit.prevent="completeLogin" class="flex flex-col sm:flex-row gap-2.5 pt-1">
        <div class="relative flex-1">
          <input
            v-model="redirectInput"
            required
            placeholder="http://localhost:51121/oauth-callback?state=…&code=…"
            class="w-full bg-[#141518] border border-[#2c2e36] focus:border-blue-500 rounded-xl px-3.5 py-2.5 text-xs text-zinc-200 font-mono placeholder:text-zinc-600 focus:outline-none transition-colors shadow-inner"
          />
        </div>
        <div class="flex items-center gap-2">
          <button
            type="submit"
            :disabled="busy || !redirectInput.trim()"
            class="flex-1 sm:flex-initial flex items-center justify-center gap-2 px-5 py-2.5 rounded-xl bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-xs font-medium transition-colors cursor-pointer shadow-sm"
          >
            <RefreshCw v-if="busy" class="h-3.5 w-3.5 animate-spin" />
            <ArrowRight v-else class="h-3.5 w-3.5" />
            <span>{{ busy ? 'Verifying...' : 'Finish Authorization' }}</span>
          </button>
          <button
            type="button"
            @click="cancelLogin"
            class="px-4 py-2.5 rounded-xl bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs font-medium transition-colors cursor-pointer"
          >
            Cancel
          </button>
        </div>
      </form>
    </div>

    <!-- Filter & Search Toolbar -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pt-2">
      <!-- Search Input -->
      <div class="relative flex-1 max-w-sm">
        <Search class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
        <input
          v-model="searchQuery"
          type="text"
          placeholder="Filter by email or project ID..."
          class="w-full bg-[#202227] border border-[#2c2e36] focus:border-blue-500 rounded-xl pl-9 pr-8 py-2 text-xs text-zinc-200 placeholder:text-zinc-500 focus:outline-none transition-colors"
        />
        <button
          v-if="searchQuery"
          @click="searchQuery = ''"
          class="absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300 p-0.5 cursor-pointer"
        >
          <X class="h-3.5 w-3.5" />
        </button>
      </div>

      <!-- Status Filter Tabs -->
      <div class="flex items-center bg-[#202227] border border-[#2c2e36] rounded-xl p-1 text-xs self-start sm:self-auto">
        <button
          @click="statusFilter = 'all'"
          class="px-3 py-1.5 rounded-lg transition-all font-medium cursor-pointer"
          :class="[
            statusFilter === 'all'
              ? 'bg-[#2e313b] text-white shadow-xs'
              : 'text-zinc-400 hover:text-zinc-200'
          ]"
        >
          All ({{ totalAccountsCount }})
        </button>
        <button
          @click="statusFilter = 'ready'"
          class="px-3 py-1.5 rounded-lg transition-all font-medium cursor-pointer flex items-center gap-1.5"
          :class="[
            statusFilter === 'ready'
              ? 'bg-emerald-500/20 text-emerald-300 border border-emerald-500/30 shadow-xs'
              : 'text-zinc-400 hover:text-zinc-200'
          ]"
        >
          <span class="h-1.5 w-1.5 rounded-full bg-emerald-400" />
          Ready ({{ readyAccountsCount }})
        </button>
        <button
          @click="statusFilter = 'cooling'"
          class="px-3 py-1.5 rounded-lg transition-all font-medium cursor-pointer flex items-center gap-1.5"
          :class="[
            statusFilter === 'cooling'
              ? 'bg-amber-500/20 text-amber-300 border border-amber-500/30 shadow-xs'
              : 'text-zinc-400 hover:text-zinc-200'
          ]"
        >
          <span class="h-1.5 w-1.5 rounded-full bg-amber-400" />
          Cooling ({{ coolingAccountsCount }})
        </button>
      </div>
    </div>

    <!-- Loading Skeleton Cards -->
    <div v-if="!loaded" class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <div
        v-for="n in 2"
        :key="n"
        class="h-64 rounded-2xl bg-[#202227] border border-[#2c2e36] p-5 space-y-4 animate-pulse"
      >
        <div class="flex items-center gap-3">
          <div class="h-10 w-10 rounded-xl bg-zinc-800" />
          <div class="space-y-2 flex-1">
            <div class="h-4 w-48 bg-zinc-800 rounded" />
            <div class="h-3 w-32 bg-zinc-800/60 rounded" />
          </div>
        </div>
        <div class="h-px bg-[#282a32]" />
        <div class="grid grid-cols-2 gap-3 pt-2">
          <div class="h-16 bg-zinc-800/40 rounded-xl" />
          <div class="h-16 bg-zinc-800/40 rounded-xl" />
        </div>
      </div>
    </div>

    <!-- Empty State: Zero accounts configured -->
    <div
      v-else-if="accounts.length === 0"
      class="p-10 rounded-2xl bg-[#202227] border border-[#2c2e36] text-center space-y-4 max-w-lg mx-auto my-8"
    >
      <div class="h-12 w-12 rounded-2xl bg-blue-500/10 border border-blue-500/20 text-blue-400 mx-auto flex items-center justify-center">
        <Users class="h-6 w-6" />
      </div>
      <div class="space-y-1.5">
        <h3 class="text-sm font-semibold text-white">No Accounts Connected</h3>
        <p class="text-xs text-zinc-400 max-w-sm mx-auto leading-relaxed">
          Connect one or more Google accounts to serve Gemini requests. The proxy will seamlessly load-balance and fail over when limits are reached.
        </p>
      </div>
      <div v-if="webLoginEnabled" class="pt-2">
        <button
          @click="startLogin"
          :disabled="busy"
          class="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl bg-blue-600 hover:bg-blue-500 text-white text-xs font-medium transition-all shadow-md shadow-blue-500/10 cursor-pointer"
        >
          <Plus class="h-4 w-4" />
          <span>Connect Google Account</span>
        </button>
      </div>
    </div>

    <!-- Filter Empty State: Search or status filter matches nothing -->
    <div
      v-else-if="filteredAccounts.length === 0"
      class="p-8 rounded-2xl bg-[#202227] border border-[#2c2e36] text-center space-y-3"
    >
      <div class="text-xs text-zinc-400">
        No accounts match <span v-if="searchQuery">"{{ searchQuery }}"</span>
        <span v-if="statusFilter !== 'all'"> with status <strong>{{ statusFilter }}</strong></span>.
      </div>
      <button
        @click="searchQuery = ''; statusFilter = 'all'"
        class="text-xs text-blue-400 hover:underline font-medium cursor-pointer"
      >
        Clear filters
      </button>
    </div>

    <!-- Account Cards Grid -->
    <div v-else class="grid grid-cols-1 lg:grid-cols-2 gap-4">
      <AccountCard
        v-for="(acc, i) in filteredAccounts"
        :key="acc.id"
        :account="acc"
        :index="i"
        :disabled="busy || testingAll"
        :is-testing="testingAccountIds.has(acc.id)"
        :test-result="testResults[acc.id]"
        @remove="openDeleteModal"
        @test="testAccount"
      />
    </div>

    <!-- Delete Confirmation Modal -->
    <DeleteAccountModal
      :open="showDeleteModal"
      :account="accountToDelete"
      :busy="busy"
      @close="closeDeleteModal"
      @confirm="confirmDeleteAccount"
    />
  </div>
</template>
