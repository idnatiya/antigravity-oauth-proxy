<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  Copy,
  Check,
  Clock,
  LogOut,
  AlertTriangle,
  Sparkles,
  Bot,
  Shield,
  Lock,
  Zap,
  RefreshCw,
} from '@lucide/vue'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Progress } from '@/components/ui/progress'
import type { AccountItem, QuotaBucket, AccountTestResult } from '@/types'

const props = defineProps<{
  account: AccountItem
  index: number
  disabled?: boolean
  isTesting?: boolean
  testResult?: AccountTestResult | null
}>()

const emit = defineEmits<{
  (e: 'remove', account: AccountItem): void
  (e: 'test', account: AccountItem): void
}>()

const copiedField = ref<'email' | 'project' | null>(null)

function copyToClipboard(text: string, field: 'email' | 'project') {
  navigator.clipboard.writeText(text)
  copiedField.value = field
  setTimeout(() => {
    if (copiedField.value === field) {
      copiedField.value = null
    }
  }, 1800)
}

function groupLabel(name: string) {
  const lower = name.toLowerCase()
  if (lower.startsWith('gemini')) return 'Gemini Models'
  if (lower.includes('claude')) return 'Claude & GPT Models'
  return name
}

function isGeminiGroup(name: string) {
  return name.toLowerCase().startsWith('gemini')
}

const windowOrder: Record<string, number> = { '5h': 0, weekly: 1 }

function sortedBuckets(buckets: QuotaBucket[]) {
  return [...buckets].sort((a, b) => (windowOrder[a.window ?? ''] ?? 9) - (windowOrder[b.window ?? ''] ?? 9))
}

function windowLabel(bucket: QuotaBucket) {
  if (bucket.window === '5h') return '5-Hour Window'
  if (bucket.window === 'weekly') return 'Weekly Limit'
  return bucket.displayName || bucket.bucketId || 'Limit'
}

function percent(bucket: QuotaBucket) {
  return bucket.remainingFraction === undefined ? null : Math.round(bucket.remainingFraction * 100)
}

function barColor(pct: number) {
  if (pct >= 50) return 'bg-emerald-500'
  if (pct >= 20) return 'bg-amber-500'
  return 'bg-rose-500'
}

function badgeColor(pct: number) {
  if (pct >= 50) return 'text-emerald-400 bg-emerald-500/10 border-emerald-500/20'
  if (pct >= 20) return 'text-amber-400 bg-amber-500/10 border-amber-500/20'
  return 'text-rose-400 bg-rose-500/10 border-rose-500/20'
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

const isCooling = computed(() => !!props.account.coolingUntil)

// Avatar initial
const initial = computed(() => {
  const clean = props.account.id.trim()
  if (!clean) return 'G'
  return clean.charAt(0).toUpperCase()
})

const isPrimary = computed(() => props.index === 0)
</script>

<template>
  <Card
    class="overflow-hidden flex flex-col justify-between"
    :class="[
      isPrimary && !isCooling
        ? 'border-blue-500/40 shadow-lg shadow-blue-500/5'
        : 'border-[#2c2e36] hover:border-[#3a3d47]',
      isCooling ? 'opacity-90' : ''
    ]"
  >
    <div>
      <!-- Header -->
      <div class="p-5 pb-4 border-b border-[#282a32] flex items-start gap-3.5">
        <!-- Avatar Initial with dynamic ring -->
        <div
          class="h-10 w-10 shrink-0 rounded-xl flex items-center justify-center font-bold text-sm select-none shadow-sm transition-transform"
          :class="[
            isPrimary && !isCooling
              ? 'bg-blue-600/20 text-blue-400 border border-blue-500/30'
              : 'bg-zinc-800 text-zinc-300 border border-zinc-700'
          ]"
        >
          {{ initial }}
        </div>

        <!-- Identity info -->
        <div class="flex-1 min-w-0">
          <div class="flex items-center gap-2">
            <span
              class="text-sm font-semibold text-white truncate font-sans tracking-tight"
              :title="account.id"
            >
              {{ account.id }}
            </span>
            <button
              type="button"
              @click="copyToClipboard(account.id, 'email')"
              :title="copiedField === 'email' ? 'Copied!' : 'Copy account email'"
              class="text-zinc-500 hover:text-zinc-300 p-0.5 rounded transition-colors cursor-pointer"
            >
              <Check v-if="copiedField === 'email'" class="h-3.5 w-3.5 text-emerald-400" />
              <Copy v-else class="h-3.5 w-3.5" />
            </button>
          </div>

          <div class="mt-1 flex items-center gap-2 text-xs">
            <span class="text-zinc-500">Project:</span>
            <span class="text-zinc-300 font-mono text-[11px] truncate max-w-[180px]" :title="account.projectId">
              {{ account.projectId || 'None' }}
            </span>
            <button
              v-if="account.projectId"
              type="button"
              @click="copyToClipboard(account.projectId, 'project')"
              :title="copiedField === 'project' ? 'Copied!' : 'Copy project ID'"
              class="text-zinc-500 hover:text-zinc-300 p-0.5 rounded transition-colors cursor-pointer"
            >
              <Check v-if="copiedField === 'project'" class="h-3.5 w-3.5 text-emerald-400" />
              <Copy v-else class="h-3.5 w-3.5" />
            </button>
          </div>
        </div>

        <!-- Priority & Status Badges -->
        <div class="shrink-0 flex flex-col items-end gap-1.5">
          <div class="flex items-center gap-2">
            <!-- Pool Priority Badge -->
            <Badge
              v-if="isPrimary"
              variant="default"
              class="uppercase tracking-wide font-semibold text-[10px]"
            >
              <Sparkles class="h-2.5 w-2.5" />
              Primary #1
            </Badge>
            <Badge
              v-else
              variant="secondary"
              class="font-mono text-[10px]"
            >
              Fallback #{{ index + 1 }}
            </Badge>

            <!-- Test Connection button -->
            <Button
              type="button"
              variant="ghost"
              size="icon"
              :disabled="disabled || isTesting"
              @click="emit('test', account)"
              :title="isTesting ? 'Testing CloudCode connectivity...' : 'Test connection to CloudCode API'"
              class="h-7 w-7 text-zinc-400 hover:text-blue-400 hover:bg-blue-500/10"
            >
              <RefreshCw v-if="isTesting" class="h-3.5 w-3.5 animate-spin text-blue-400" />
              <Zap v-else class="h-3.5 w-3.5" />
            </Button>

            <!-- Logout button -->
            <Button
              v-if="account.removable"
              type="button"
              variant="ghost"
              size="icon"
              :disabled="disabled || isTesting"
              @click="emit('remove', account)"
              title="Disconnect account"
              class="h-7 w-7 text-zinc-500 hover:text-red-400 hover:bg-red-500/10"
            >
              <LogOut class="h-3.5 w-3.5" />
            </Button>
            <span
              v-else
              title="Configured via environment variable"
              class="p-1.5 text-zinc-600 cursor-not-allowed"
            >
              <Lock class="h-3.5 w-3.5" />
            </span>
          </div>

          <!-- Status Indicator & Test Result -->
          <div class="flex items-center gap-2">
            <!-- Test Result Pill (if tested) -->
            <span
              v-if="testResult"
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-mono font-medium border"
              :class="testResult.success ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/25' : 'bg-rose-500/10 text-rose-400 border-rose-500/25 cursor-help'"
              :title="testResult.success ? `Verification successful in ${testResult.latencyMs}ms` : (testResult.error || 'Connection failed')"
            >
              <Check v-if="testResult.success" class="h-2.5 w-2.5" />
              <AlertTriangle v-else class="h-2.5 w-2.5" />
              {{ testResult.success ? `${testResult.latencyMs}ms OK` : 'Check Failed' }}
            </span>

            <span
              v-if="isCooling"
              :title="`${account.coolingReason ?? ''} (until ${localTime(account.coolingUntil)})`"
              class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-amber-500/10 text-amber-400 border border-amber-500/20 text-[11px] font-medium"
            >
              <span class="h-1.5 w-1.5 rounded-full bg-amber-400 animate-pulse" />
              Cooling ({{ shortDuration(account.coolingUntil) }} left)
            </span>
            <span
              v-else
              class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-[11px] font-medium"
            >
              <span class="relative flex h-2 w-2">
                <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75" />
                <span class="relative inline-flex rounded-full h-2 w-2 bg-emerald-500" />
              </span>
              Ready
            </span>
          </div>
        </div>
      </div>

      <!-- Cooling Alert Banner (if in cooldown) -->
      <div
        v-if="isCooling"
        class="px-5 py-3 bg-amber-500/10 border-b border-amber-500/20 text-amber-200 text-xs flex items-start gap-2.5"
      >
        <AlertTriangle class="h-4 w-4 text-amber-400 shrink-0 mt-0.5" />
        <div class="space-y-0.5 flex-1 min-w-0">
          <div class="font-semibold text-amber-300">Account In Cooldown</div>
          <p class="text-[11px] text-amber-200/90 leading-tight">
            Requests are automatically routing to the next ready account until
            <span class="font-medium text-white">{{ localTime(account.coolingUntil) }}</span>.
          </p>
          <p v-if="account.coolingReason" class="text-[10px] text-amber-300/80 font-mono truncate" :title="account.coolingReason">
            Reason: {{ account.coolingReason }}
          </p>
        </div>
      </div>

      <!-- Quota Section -->
      <div v-if="account.quota?.groups?.length" class="p-5 space-y-4">
        <div
          v-for="(group, gi) in account.quota.groups"
          :key="group.displayName"
          class="space-y-2.5"
          :class="gi > 0 ? 'pt-3 border-t border-[#282a32]' : ''"
        >
          <!-- Group Title Pill -->
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-1.5 text-xs font-semibold text-zinc-300">
              <Sparkles v-if="isGeminiGroup(group.displayName)" class="h-3.5 w-3.5 text-blue-400" />
              <Bot v-else class="h-3.5 w-3.5 text-purple-400" />
              <span>{{ groupLabel(group.displayName) }}</span>
            </div>
            <span class="text-[10px] font-mono uppercase tracking-wider text-zinc-500">
              Quota Limits
            </span>
          </div>

          <!-- Buckets Grid -->
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-2.5">
            <div
              v-for="bucket in sortedBuckets(group.buckets)"
              :key="bucket.bucketId || bucket.window"
              class="p-3 rounded-xl bg-[#18191d] border border-[#282a32] space-y-2"
            >
              <div class="flex items-center justify-between">
                <span class="text-[11px] font-medium text-zinc-400">{{ windowLabel(bucket) }}</span>
                <span
                  v-if="percent(bucket) !== null"
                  class="px-1.5 py-0.5 rounded text-[10px] font-mono font-semibold border"
                  :class="badgeColor(percent(bucket)!)"
                >
                  {{ percent(bucket) }}% left
                </span>
                <span v-else class="text-[11px] font-mono text-zinc-500">–</span>
              </div>

              <!-- Progress bar -->
              <Progress
                :model-value="percent(bucket) ?? 0"
                class="h-2 bg-[#202227]"
                :indicator-class="barColor(percent(bucket) ?? 0)"
              />

              <!-- Reset countdown -->
              <div class="flex items-center justify-between text-[10px] text-zinc-500 font-mono">
                <span class="flex items-center gap-1">
                  <Clock class="h-3 w-3 text-zinc-600" />
                  Reset:
                </span>
                <span
                  v-if="bucket.resetTime"
                  class="text-zinc-300 tabular-nums cursor-help"
                  :title="`Resets at ${localTime(bucket.resetTime)}`"
                >
                  {{ shortDuration(bucket.resetTime) }}
                </span>
                <span v-else class="text-zinc-600">Active</span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Quota Fetch Error -->
      <div
        v-else-if="account.quotaError"
        class="p-4 m-5 rounded-xl bg-red-500/10 border border-red-500/20 text-red-300 flex items-start gap-2.5 text-xs"
      >
        <AlertTriangle class="h-4 w-4 shrink-0 text-red-400 mt-0.5" />
        <div class="space-y-0.5 flex-1 min-w-0">
          <div class="font-semibold text-red-300">Quota Data Unavailable</div>
          <p class="text-[11px] text-red-200/80 break-all font-mono">
            {{ account.quotaError }}
          </p>
        </div>
      </div>

      <!-- No Quota Groups Info -->
      <div
        v-else
        class="p-5 text-center text-xs text-zinc-500 font-mono flex items-center justify-center gap-2"
      >
        <Shield class="h-4 w-4 text-zinc-600" />
        <span>No specific rate limits reported by Google</span>
      </div>
    </div>
  </Card>
</template>
