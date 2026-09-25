<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter, RouterLink } from 'vue-router'
import {
  ShieldCheck,
  Key,
  Lock,
  CheckCircle2,
  AlertCircle,
  Eye,
  EyeOff,
  Copy,
  Check,
  LogOut,
  ArrowRight,
  KeyRound,
  RotateCw,
} from '@lucide/vue'
import { useAuthStore } from '@/stores/authStore'
import { useUsageStore } from '@/stores/usageStore'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'

const router = useRouter()
const authStore = useAuthStore()
const usageStore = useUsageStore()

const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')

const showCurrentPassword = ref(false)
const showNewPassword = ref(false)
const showConfirmPassword = ref(false)

const isSaving = ref(false)
const isLoggingOut = ref(false)
const copiedProject = ref(false)
const successMessage = ref('')
const errorMessage = ref('')

onMounted(() => {
  if (!usageStore.account) {
    usageStore.fetchStats()
  }
})

const isLengthValid = computed(() => newPassword.value.length >= 4)
const isMatchValid = computed(() => Boolean(newPassword.value && newPassword.value === confirmPassword.value))
const canSubmit = computed(() => {
  return (
    currentPassword.value.length > 0 &&
    isLengthValid.value &&
    isMatchValid.value &&
    !isSaving.value
  )
})

async function handleChangePassword() {
  errorMessage.value = ''
  successMessage.value = ''

  if (!currentPassword.value || !newPassword.value || !confirmPassword.value) {
    errorMessage.value = 'Please fill out all password fields.'
    return
  }

  if (newPassword.value.length < 4) {
    errorMessage.value = 'New password must be at least 4 characters.'
    return
  }

  if (newPassword.value !== confirmPassword.value) {
    errorMessage.value = 'New password and confirmation do not match.'
    return
  }

  isSaving.value = true
  try {
    await authStore.changePassword({
      current_password: currentPassword.value,
      new_password: newPassword.value,
      confirm_password: confirmPassword.value,
    })
    successMessage.value = 'Password changed successfully.'
    currentPassword.value = ''
    newPassword.value = ''
    confirmPassword.value = ''
    showCurrentPassword.value = false
    showNewPassword.value = false
    showConfirmPassword.value = false
  } catch (err: unknown) {
    if (err instanceof Error) {
      errorMessage.value = err.message
    } else {
      errorMessage.value = 'Failed to update password.'
    }
  } finally {
    isSaving.value = false
  }
}

async function handleLogout() {
  isLoggingOut.value = true
  try {
    await authStore.logout()
    router.push({ name: 'login' })
  } finally {
    isLoggingOut.value = false
  }
}

async function copyProjectId() {
  const pid = usageStore.account?.project_id
  if (!pid) return
  try {
    await navigator.clipboard.writeText(pid)
    copiedProject.value = true
    setTimeout(() => {
      copiedProject.value = false
    }, 2000)
  } catch (err) {
    console.error('Failed to copy project ID', err)
  }
}
</script>

<template>
  <div class="space-y-6 w-full">
    <!-- Top Header & Status -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2.5">
          <h2 class="text-base font-semibold text-white tracking-tight">Security & Credentials</h2>
          <Badge variant="outline" class="font-mono text-xs bg-emerald-500/10 text-emerald-400 border-emerald-500/20 px-2 py-0.5 gap-1.5 flex items-center">
            <span class="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
            <span>Active Session</span>
          </Badge>
        </div>
        <p class="text-xs text-zinc-400 mt-1">
          Manage administrative authentication, token status, and upstream Google CloudCode connectivity.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          @click="handleLogout"
          :disabled="isLoggingOut"
          class="h-8 px-3 text-xs bg-[#121316] border-zinc-800/80 hover:bg-red-500/10 hover:text-red-400 hover:border-red-500/30 text-zinc-300 gap-1.5 transition-all shadow-xs"
        >
          <RotateCw v-if="isLoggingOut" class="h-3.5 w-3.5 animate-spin" />
          <LogOut v-else class="h-3.5 w-3.5" />
          <span>{{ isLoggingOut ? 'Logging out...' : 'Sign Out' }}</span>
        </Button>
      </div>
    </div>

    <!-- 3 Security KPI Metric Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-3.5">
      <!-- Session Status -->
      <Card class="p-4 bg-[#121316] border-zinc-800/80 flex items-center justify-between shadow-xs">
        <div class="space-y-1">
          <span class="text-[11px] font-medium text-zinc-400 uppercase tracking-wider">Session</span>
          <div class="text-base font-bold font-mono text-emerald-400">Authenticated</div>
          <span class="text-[10px] text-zinc-500 block">Superadmin Role</span>
        </div>
        <div class="w-9 h-9 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400">
          <ShieldCheck class="h-4.5 w-4.5" />
        </div>
      </Card>

      <!-- Upstream GCP Project -->
      <Card class="p-4 bg-[#121316] border-zinc-800/80 flex items-center justify-between shadow-xs">
        <div class="space-y-1 min-w-0 pr-2">
          <span class="text-[11px] font-medium text-zinc-400 uppercase tracking-wider">CloudCode Project</span>
          <div class="text-xs font-bold font-mono text-blue-400 truncate pt-0.5">
            {{ usageStore.account?.project_id || 'Auto-Discovered' }}
          </div>
          <span class="text-[10px] text-zinc-500 block">Google Cloud Assist</span>
        </div>
        <div class="w-9 h-9 rounded-xl bg-blue-500/10 border border-blue-500/20 flex items-center justify-center text-blue-400 shrink-0">
          <Key class="h-4.5 w-4.5" />
        </div>
      </Card>

      <!-- API Key Protection -->
      <RouterLink to="/api-keys" class="block">
        <Card class="p-4 bg-[#121316] border-zinc-800/80 hover:border-zinc-700/80 transition-all flex items-center justify-between shadow-xs cursor-pointer group">
          <div class="space-y-1">
            <span class="text-[11px] font-medium text-zinc-400 uppercase tracking-wider group-hover:text-blue-400 transition-colors">API Keys Guard</span>
            <div class="text-sm font-bold font-mono text-emerald-400">Multiple Keys</div>
            <span class="text-[10px] text-zinc-500 block">Manage in API Keys &rarr;</span>
          </div>
          <div class="w-9 h-9 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400 group-hover:scale-105 transition-transform">
            <KeyRound class="h-4.5 w-4.5" />
          </div>
        </Card>
      </RouterLink>
    </div>

    <!-- Cards Grid: Administrator Profile & Upstream CloudCode -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <!-- Admin Account Card -->
      <Card class="p-5 bg-[#121316] border-zinc-800/80 space-y-4 shadow-xs">
        <div class="flex items-center gap-3">
          <div class="h-9 w-9 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-400 flex items-center justify-center shrink-0">
            <ShieldCheck class="h-4.5 w-4.5" />
          </div>
          <div class="min-w-0 flex-1">
            <h3 class="text-sm font-semibold text-white">Administrator Profile</h3>
            <p class="text-xs text-zinc-400">Active authenticated dashboard session</p>
          </div>
          <Badge variant="outline" class="font-mono text-[10px] bg-emerald-500/10 text-emerald-400 border-emerald-500/20 px-2 py-0.5">
            Superadmin
          </Badge>
        </div>

        <div class="space-y-2.5 pt-1 text-xs">
          <Separator class="bg-zinc-800/60" />
          <div class="flex justify-between py-1 items-center">
            <span class="text-zinc-500 font-sans">Username</span>
            <span class="text-zinc-200 font-mono font-medium">{{ authStore.user?.username || 'admin' }}</span>
          </div>
          <Separator class="bg-zinc-800/60" />
          <div class="flex justify-between py-1 items-center">
            <span class="text-zinc-500 font-sans">Session Mechanism</span>
            <span class="text-zinc-300 font-mono">HttpOnly Secure Cookie</span>
          </div>
          <Separator class="bg-zinc-800/60" />
          <div class="flex justify-between py-1 items-center">
            <span class="text-zinc-500 font-sans">Password Hashing</span>
            <span class="text-zinc-300 font-mono">Bcrypt (Cost 10)</span>
          </div>
        </div>
      </Card>

      <!-- CloudCode Token Card -->
      <Card class="p-5 bg-[#121316] border-zinc-800/80 space-y-4 shadow-xs">
        <div class="flex items-center justify-between gap-3">
          <div class="flex items-center gap-3">
            <div class="h-9 w-9 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-400 flex items-center justify-center shrink-0">
              <Key class="h-4.5 w-4.5" />
            </div>
            <div>
              <h3 class="text-sm font-semibold text-white">Google CloudCode Upstream</h3>
              <p class="text-xs text-zinc-400">Upstream Gemini Code Assist credentials</p>
            </div>
          </div>
          <router-link
            to="/accounts"
            class="text-xs text-blue-400 hover:text-blue-300 flex items-center gap-1 transition-colors font-medium cursor-pointer"
          >
            <span>Manage</span>
            <ArrowRight class="h-3 w-3" />
          </router-link>
        </div>

        <div class="space-y-2.5 pt-1 text-xs">
          <Separator class="bg-zinc-800/60" />
          <div class="flex justify-between py-1 items-center">
            <span class="text-zinc-500 font-sans">Project ID</span>
            <div class="flex items-center gap-1.5 font-mono">
              <span class="text-blue-400">{{ usageStore.account?.project_id || '—' }}</span>
              <button
                v-if="usageStore.account?.project_id"
                type="button"
                @click="copyProjectId"
                class="text-zinc-500 hover:text-zinc-200 transition-colors p-0.5 rounded cursor-pointer"
                title="Copy Project ID"
              >
                <Check v-if="copiedProject" class="h-3 w-3 text-emerald-400" />
                <Copy v-else class="h-3 w-3" />
              </button>
            </div>
          </div>
          <Separator class="bg-zinc-800/60" />
          <div class="flex justify-between py-1 items-center">
            <span class="text-zinc-500 font-sans">Accounts Ready</span>
            <span class="text-zinc-200 font-mono font-medium">
              {{ usageStore.account?.accounts_ready ?? 1 }} / {{ usageStore.account?.accounts_total ?? 1 }} Ready
            </span>
          </div>
          <Separator class="bg-zinc-800/60" />
          <div class="flex justify-between py-1 items-center">
            <span class="text-zinc-500 font-sans">Token Refresh</span>
            <span class="text-emerald-400 font-mono">Auto 5m Buffer</span>
          </div>
        </div>
      </Card>
    </div>

    <!-- Change Password Form Card -->
    <Card class="p-6 bg-[#121316] border-zinc-800/80 space-y-5 shadow-xs">
      <div class="flex items-center gap-3">
        <div class="h-9 w-9 rounded-xl bg-amber-500/10 border border-amber-500/20 text-amber-400 flex items-center justify-center shrink-0">
          <Lock class="h-4.5 w-4.5" />
        </div>
        <div>
          <h3 class="text-sm font-semibold text-white">Update Password</h3>
          <p class="text-xs text-zinc-400">Change your local administrative credentials</p>
        </div>
      </div>

      <!-- Feedback messages -->
      <div
        v-if="successMessage"
        class="p-3.5 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-300 text-xs flex items-center gap-2.5 leading-relaxed"
      >
        <CheckCircle2 class="h-4 w-4 text-emerald-400 shrink-0" />
        <span>{{ successMessage }}</span>
      </div>

      <div
        v-if="errorMessage"
        class="p-3.5 rounded-xl bg-red-500/10 border border-red-500/20 text-red-300 text-xs flex items-center gap-2.5 leading-relaxed"
      >
        <AlertCircle class="h-4 w-4 text-red-400 shrink-0" />
        <span>{{ errorMessage }}</span>
      </div>

      <form @submit.prevent="handleChangePassword" class="space-y-4 max-w-lg text-xs">
        <!-- Current Password -->
        <div class="space-y-1.5">
          <Label for="current-password" class="text-zinc-300 font-medium text-xs">Current Password</Label>
          <div class="relative">
            <Input
              id="current-password"
              v-model="currentPassword"
              :type="showCurrentPassword ? 'text' : 'password'"
              required
              placeholder="Enter current password"
              class="bg-[#0d0e11] border-zinc-800/80 pr-9 text-xs h-9 focus-visible:ring-blue-500 text-zinc-100 placeholder:text-zinc-600 rounded-lg"
            />
            <button
              type="button"
              @click="showCurrentPassword = !showCurrentPassword"
              class="absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300 p-0.5 cursor-pointer"
              :title="showCurrentPassword ? 'Hide password' : 'Show password'"
            >
              <EyeOff v-if="showCurrentPassword" class="h-3.5 w-3.5" />
              <Eye v-else class="h-3.5 w-3.5" />
            </button>
          </div>
        </div>

        <!-- New Password -->
        <div class="space-y-1.5">
          <Label for="new-password" class="text-zinc-300 font-medium text-xs">New Password</Label>
          <div class="relative">
            <Input
              id="new-password"
              v-model="newPassword"
              :type="showNewPassword ? 'text' : 'password'"
              required
              minlength="4"
              placeholder="At least 4 characters"
              class="bg-[#0d0e11] border-zinc-800/80 pr-9 text-xs h-9 focus-visible:ring-blue-500 text-zinc-100 placeholder:text-zinc-600 rounded-lg"
            />
            <button
              type="button"
              @click="showNewPassword = !showNewPassword"
              class="absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300 p-0.5 cursor-pointer"
              :title="showNewPassword ? 'Hide password' : 'Show password'"
            >
              <EyeOff v-if="showNewPassword" class="h-3.5 w-3.5" />
              <Eye v-else class="h-3.5 w-3.5" />
            </button>
          </div>
        </div>

        <!-- Confirm New Password -->
        <div class="space-y-1.5">
          <Label for="confirm-password" class="text-zinc-300 font-medium text-xs">Confirm New Password</Label>
          <div class="relative">
            <Input
              id="confirm-password"
              v-model="confirmPassword"
              :type="showConfirmPassword ? 'text' : 'password'"
              required
              minlength="4"
              placeholder="Repeat new password"
              class="bg-[#0d0e11] border-zinc-800/80 pr-9 text-xs h-9 focus-visible:ring-blue-500 text-zinc-100 placeholder:text-zinc-600 rounded-lg"
            />
            <button
              type="button"
              @click="showConfirmPassword = !showConfirmPassword"
              class="absolute right-2.5 top-1/2 -translate-y-1/2 text-zinc-500 hover:text-zinc-300 p-0.5 cursor-pointer"
              :title="showConfirmPassword ? 'Hide password' : 'Show password'"
            >
              <EyeOff v-if="showConfirmPassword" class="h-3.5 w-3.5" />
              <Eye v-else class="h-3.5 w-3.5" />
            </button>
          </div>
        </div>

        <!-- Live Validation Checklist -->
        <div v-if="newPassword" class="p-3 rounded-lg bg-[#0d0e11] border border-zinc-800/70 space-y-1.5 pt-2">
          <div class="text-[11px] font-medium text-zinc-400">Password Requirements:</div>
          <div class="flex items-center gap-2 text-[11px]">
            <Check v-if="isLengthValid" class="h-3 w-3 text-emerald-400" />
            <span v-else class="w-1.5 h-1.5 rounded-full bg-zinc-600 mx-0.75" />
            <span :class="isLengthValid ? 'text-emerald-400 font-medium' : 'text-zinc-400'">
              Minimum 4 characters
            </span>
          </div>
          <div class="flex items-center gap-2 text-[11px]">
            <Check v-if="isMatchValid" class="h-3 w-3 text-emerald-400" />
            <span v-else class="w-1.5 h-1.5 rounded-full bg-zinc-600 mx-0.75" />
            <span :class="isMatchValid ? 'text-emerald-400 font-medium' : 'text-zinc-400'">
              Passwords match
            </span>
          </div>
        </div>

        <Button
          type="submit"
          :disabled="!canSubmit"
          class="h-9 px-4 text-xs font-medium bg-blue-600 hover:bg-blue-500 text-white rounded-lg transition-all cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed gap-2"
        >
          <RotateCw v-if="isSaving" class="h-3.5 w-3.5 animate-spin" />
          <span>{{ isSaving ? 'Saving...' : 'Update Password' }}</span>
        </Button>
      </form>
    </Card>

    <!-- API Security & Protection Architecture Card -->
    <Card class="p-6 bg-[#121316] border-zinc-800/80 space-y-4 shadow-xs">
      <div class="flex items-center gap-3">
        <div class="h-9 w-9 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-400 flex items-center justify-center shrink-0">
          <KeyRound class="h-4.5 w-4.5" />
        </div>
        <div>
          <h3 class="text-sm font-semibold text-white">API Security & Endpoint Protection</h3>
          <p class="text-xs text-zinc-400">How external AI tools and proxy clients authenticate</p>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4 pt-1">
        <!-- Protected Endpoints -->
        <div class="p-4 rounded-xl bg-[#0d0e11] border border-zinc-800/80 space-y-2 text-xs">
          <div class="font-medium text-zinc-200 flex items-center gap-1.5">
            <ShieldCheck class="h-3.5 w-3.5 text-blue-400" />
            <span>Protected Endpoints</span>
          </div>
          <p class="text-[11px] text-zinc-400 leading-relaxed">
            Requests to proxy model execution endpoints require an authenticated API Key (managed via Dashboard):
          </p>
          <div class="rounded-lg bg-[#08090b] border border-zinc-800 p-2.5 font-mono text-[11px] text-blue-300">
            Authorization: Bearer &lt;API_KEY&gt;
          </div>
          <ul class="text-[11px] text-zinc-400 space-y-1 list-disc pl-4 pt-1">
            <li><code class="font-mono text-zinc-300">/v1/chat/completions</code> (OpenAI compatible)</li>
            <li><code class="font-mono text-zinc-300">/v1beta/models/*</code> (Gemini native)</li>
            <li><code class="font-mono text-zinc-300">/mcp</code> (Model Context Protocol streamable)</li>
          </ul>
          <div class="pt-1">
            <RouterLink to="/api-keys" class="inline-flex items-center text-[11px] text-blue-400 hover:text-blue-300 font-sans gap-1">
              <span>Manage API Keys</span>
              <ArrowRight class="h-3 w-3" />
            </RouterLink>
          </div>
        </div>

        <!-- Security Best Practices -->
        <div class="p-4 rounded-xl bg-[#0d0e11] border border-zinc-800/80 space-y-2 text-xs">
          <div class="font-medium text-zinc-200 flex items-center gap-1.5">
            <CheckCircle2 class="h-3.5 w-3.5 text-emerald-400" />
            <span>Proxy Hardening & Privacy</span>
          </div>
          <p class="text-[11px] text-zinc-400 leading-relaxed">
            Built with modern security and privacy practices for local and edge deployments:
          </p>
          <ul class="text-[11px] text-zinc-400 space-y-1.5 list-disc pl-4 pt-1">
            <li>Bcrypt-hashed administrator passwords stored locally in SQLite</li>
            <li>HttpOnly session cookies preventing cross-site script (XSS) leaks</li>
            <li>Automatic periodic OAuth token refreshing with 5-minute expiry buffer</li>
            <li>Zero prompt logging to third parties; requests logged only locally</li>
          </ul>
        </div>
      </div>
    </Card>
  </div>
</template>
