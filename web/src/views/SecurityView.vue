<script setup lang="ts">
import { ref } from 'vue'
import { ShieldCheck, Key, Lock, CheckCircle2, AlertCircle } from '@lucide/vue'
import { useAuthStore } from '@/stores/authStore'
import { useUsageStore } from '@/stores/usageStore'
import { Card } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'

const authStore = useAuthStore()
const usageStore = useUsageStore()

const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')

const isSaving = ref(false)
const successMessage = ref('')
const errorMessage = ref('')

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
</script>

<template>
  <div class="space-y-6 max-w-4xl">
    <div>
      <h2 class="text-base font-semibold text-white tracking-tight">Security & Credentials</h2>
      <p class="text-xs text-zinc-500">Manage dashboard administrative authentication and token status</p>
    </div>

    <!-- Cards Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <!-- Admin Account Card -->
      <Card class="p-5 bg-[#202227] border-[#2c2e36] space-y-4">
        <div class="flex items-center gap-3">
          <div class="h-9 w-9 rounded-lg bg-blue-500/10 text-blue-400 flex items-center justify-center">
            <ShieldCheck class="h-5 w-5" />
          </div>
          <div>
            <h3 class="text-sm font-semibold text-white">Administrator Profile</h3>
            <p class="text-xs text-zinc-400">Authenticated dashboard session</p>
          </div>
        </div>

        <div class="space-y-2.5 pt-1 text-xs">
          <Separator class="bg-[#2a2d34]" />
          <div class="flex justify-between py-1">
            <span class="text-zinc-500">Username</span>
            <span class="text-zinc-200 font-mono font-medium">{{ authStore.user?.username || 'admin' }}</span>
          </div>
          <div class="flex justify-between py-1 items-center">
            <span class="text-zinc-500">Role</span>
            <Badge variant="success" class="font-mono text-[10px]">Superadmin</Badge>
          </div>
        </div>
      </Card>

      <!-- CloudCode Token Card -->
      <Card class="p-5 bg-[#202227] border-[#2c2e36] space-y-4">
        <div class="flex items-center gap-3">
          <div class="h-9 w-9 rounded-lg bg-emerald-500/10 text-emerald-400 flex items-center justify-center">
            <Key class="h-5 w-5" />
          </div>
          <div>
            <h3 class="text-sm font-semibold text-white">Google OAuth Token</h3>
            <p class="text-xs text-zinc-400">Upstream CloudCode authorization</p>
          </div>
        </div>

        <div class="space-y-2.5 pt-1 text-xs">
          <Separator class="bg-[#2a2d34]" />
          <div class="flex justify-between py-1">
            <span class="text-zinc-500">Project ID</span>
            <span class="text-blue-400 font-mono">{{ usageStore.account?.project_id || '—' }}</span>
          </div>
          <div class="flex justify-between py-1">
            <span class="text-zinc-500">Credential Store</span>
            <span class="text-zinc-200 font-mono">{{ usageStore.account?.provider || 'FileProvider' }}</span>
          </div>
        </div>
      </Card>
    </div>

    <!-- Change Password Form Card -->
    <Card class="p-6 bg-[#202227] border-[#2c2e36] space-y-5">
      <div class="flex items-center gap-3">
        <div class="h-8 w-8 rounded-lg bg-amber-500/10 text-amber-400 flex items-center justify-center">
          <Lock class="h-4 w-4" />
        </div>
        <div>
          <h3 class="text-sm font-semibold text-white">Update Password</h3>
          <p class="text-xs text-zinc-400">Change your local dashboard login credentials</p>
        </div>
      </div>

      <!-- Feedback messages -->
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

      <form @submit.prevent="handleChangePassword" class="space-y-4 max-w-md text-xs">
        <div class="space-y-1.5">
          <Label for="current-password">Current Password</Label>
          <Input
            id="current-password"
            v-model="currentPassword"
            type="password"
            required
            class="bg-[#18191d] border-[#2c2e36] h-9"
          />
        </div>

        <div class="space-y-1.5">
          <Label for="new-password">New Password</Label>
          <Input
            id="new-password"
            v-model="newPassword"
            type="password"
            required
            minlength="4"
            placeholder="At least 4 characters"
            class="bg-[#18191d] border-[#2c2e36] h-9"
          />
        </div>

        <div class="space-y-1.5">
          <Label for="confirm-password">Confirm New Password</Label>
          <Input
            id="confirm-password"
            v-model="confirmPassword"
            type="password"
            required
            minlength="4"
            placeholder="Repeat new password"
            class="bg-[#18191d] border-[#2c2e36] h-9"
          />
        </div>

        <Button
          type="submit"
          :disabled="isSaving"
        >
          {{ isSaving ? 'Saving...' : 'Update Password' }}
        </Button>
      </form>
    </Card>
  </div>
</template>
