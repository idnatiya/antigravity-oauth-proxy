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
  ShieldAlert,
  Lock,
  Zap,
  RefreshCw,
  ExternalLink,
  CheckCircle2,
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

// Detect if account is non-Google AI Pro (Free / Standard Tier)
const isNonPro = computed(() => {
  const err = props.account.quotaError?.toLowerCase() || ''
  if (err.includes('no google ai pro') || err.includes('403') || err.includes('permission') || err.includes('subscription')) {
    return true
  }
  if (props.account.quota?.groups?.length) {
    const hasClaude = props.account.quota.groups.some(g => g.displayName.toLowerCase().includes('claude'))
    return !hasClaude
  }
  return !props.account.quota?.groups?.length
})

const hasClaudeGroup = computed(() => {
  return props.account.quota?.groups?.some(g => g.displayName.toLowerCase().includes('claude')) ?? false
})

const isGenuineError = computed(() => {
  if (!props.account.quotaError) return false
  const err = props.account.quotaError.toLowerCase()
  return !err.includes('no google ai pro') && !err.includes('403') && !err.includes('permission') && !err.includes('subscription')
})

const needsVerification = computed(() => {
  if (props.account.needsVerification) return true
  if (props.testResult?.needsVerification) return true
  if (props.account.validationUrl) return true
  const reason = props.account.coolingReason?.toLowerCase() || ''
  if (reason.includes('verification') || reason.includes('verify your account')) return true
  const qErr = props.account.quotaError?.toLowerCase() || ''
  if (qErr.includes('verification') || qErr.includes('verify your account')) return true
  const testErr = props.testResult?.error?.toLowerCase() || ''
  if (testErr.includes('verification') || testErr.includes('verify your account')) return true
  return false
})

const verificationLink = computed(() => {
  return props.account.validationUrl || props.testResult?.validationUrl || 'https://accounts.google.com/'
})
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
            <!-- Plan Tier Badge -->
            <Badge
              v-if="!isNonPro"
              variant="outline"
              class="border-purple-500/30 bg-purple-500/10 text-purple-300 font-sans text-[10px] flex items-center gap-1"
            >
              <Sparkles class="h-2.5 w-2.5 text-purple-400" />
              AI Pro
            </Badge>
            <Badge
              v-else
              variant="outline"
              class="border-zinc-700 bg-zinc-800/80 text-zinc-400 font-sans text-[10px]"
            >
              Standard Free
            </Badge>

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
              v-if="needsVerification"
              class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full bg-amber-500/15 text-amber-300 border border-amber-500/30 text-[11px] font-semibold"
            >
              <ShieldAlert class="h-3 w-3 text-amber-400" />
              Verification Needed
            </span>
            <span
              v-else-if="isCooling"
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

      <!-- Verification Required Alert Banner -->
      <div
        v-if="needsVerification"
        class="px-5 py-3.5 bg-gradient-to-r from-amber-500/20 via-orange-500/15 to-amber-500/10 border-b border-amber-500/30 text-amber-200 text-xs space-y-2.5 shadow-inner"
      >
        <div class="flex items-start gap-2.5">
          <ShieldAlert class="h-4.5 w-4.5 text-amber-400 shrink-0 mt-0.5" />
          <div class="space-y-1 flex-1 min-w-0">
            <div class="font-semibold text-amber-300 flex items-center gap-2">
              <span>Google Account Verification Required (HTTP 403)</span>
              <span class="px-1.5 py-0.2 rounded text-[10px] bg-amber-400/20 text-amber-300 border border-amber-400/30 font-mono font-medium">ACTION REQUIRED</span>
            </div>
            <p class="text-[11px] text-amber-200/90 leading-relaxed">
              Google has flagged this account for manual verification (captcha/security confirmation). Complete the check in your browser to re-enable API access.
            </p>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="flex flex-wrap items-center gap-2 pt-1 pl-7">
          <a
            :href="verificationLink"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-amber-500 hover:bg-amber-400 text-zinc-950 font-semibold text-xs transition-colors shadow-sm cursor-pointer"
            title="Open Google verification in a new browser tab"
          >
            <ExternalLink class="h-3.5 w-3.5" />
            <span>Verify Account</span>
          </a>

          <Button
            type="button"
            variant="outline"
            size="sm"
            class="h-7 text-xs border-amber-500/40 bg-zinc-900/80 text-amber-300 hover:text-white hover:bg-amber-500/20"
            :disabled="disabled || isTesting"
            @click="emit('test', account)"
            title="Test account connectivity and clear warning if verification is complete"
          >
            <RefreshCw v-if="isTesting" class="h-3 w-3 animate-spin mr-1 text-amber-400" />
            <CheckCircle2 v-else class="h-3 w-3 mr-1 text-emerald-400" />
            <span>Re-check Status</span>
          </Button>
        </div>
      </div>

      <!-- Cooling Alert Banner (if in cooldown and not verification) -->
      <div
        v-else-if="isCooling"
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

        <!-- Claude & GPT locked row if not in Pro account -->
        <div v-if="!hasClaudeGroup" class="pt-3 border-t border-[#282a32] space-y-2">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-1.5 text-xs font-semibold text-zinc-400">
              <Bot class="h-3.5 w-3.5 text-zinc-500" />
              <span>Claude & GPT Models</span>
            </div>
            <span class="text-[10px] font-mono uppercase tracking-wider text-zinc-600">
              Exclusive
            </span>
          </div>
          <div class="p-3 rounded-xl bg-[#141518] border border-[#282a32] flex items-center justify-between">
            <span class="text-[11px] text-zinc-400">Claude 3.5/3.7 Sonnet & Opus</span>
            <span class="px-2 py-0.5 rounded text-[10px] font-mono text-amber-300/90 bg-amber-500/10 border border-amber-500/20 flex items-center gap-1">
              <Lock class="h-3 w-3 text-amber-400" />
              Google AI Pro Required
            </span>
          </div>
        </div>
      </div>

      <!-- Clean Non-Pro Account Info State (Free Tier / No AI Pro Subscription) -->
      <div v-else-if="isNonPro" class="p-5 space-y-3">
        <div class="p-4 rounded-xl bg-[#16181d] border border-zinc-800/80 space-y-3">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <div class="h-8 w-8 rounded-lg bg-zinc-800/90 border border-zinc-700/60 flex items-center justify-center text-zinc-400">
                <Sparkles class="h-4 w-4 text-blue-400" />
              </div>
              <div>
                <div class="text-xs font-semibold text-zinc-200">Standard Google Account</div>
                <div class="text-[11px] text-zinc-400">Free Tier (No Google AI Pro plan)</div>
              </div>
            </div>
            <Badge variant="outline" class="text-[10px] text-zinc-400 border-zinc-700 bg-zinc-800/50">
              Free Quota
            </Badge>
          </div>

          <p class="text-[11px] text-zinc-400 leading-relaxed">
            Detailed rate limit windows (5-hour and weekly limits) are only provided by Google for AI Pro subscribers. This account can still proxy Gemini models under Google's standard free tier.
          </p>

          <div class="grid grid-cols-1 sm:grid-cols-2 gap-2 pt-1 text-[11px] font-mono">
            <div class="p-2.5 rounded-lg bg-[#0d0e11] border border-zinc-800/70 flex items-center gap-2 text-zinc-300">
              <Check class="h-3.5 w-3.5 text-emerald-400 shrink-0" />
              <span>Gemini Flash & Pro (Free)</span>
            </div>
            <div class="p-2.5 rounded-lg bg-[#0d0e11] border border-zinc-800/70 flex items-center gap-2 text-zinc-500">
              <Lock class="h-3.5 w-3.5 text-amber-500/80 shrink-0" />
              <span>Claude & GPT (Requires Pro)</span>
            </div>
          </div>
        </div>
      </div>

      <!-- Quota Fetch Genuine Error -->
      <div
        v-else-if="isGenuineError"
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
        <span>Standard rate limits managed by Google CloudCode</span>
      </div>
    </div>
  </Card>
</template>
