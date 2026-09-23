<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Sparkles, Lock, User, AlertCircle, ArrowRight } from '@lucide/vue'
import { useAuthStore } from '@/stores/authStore'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()

const username = ref('admin')
const password = ref('admin')
const errorMessage = ref('')
const isSubmitting = ref(false)

async function handleLogin() {
  if (!username.value || !password.value) {
    errorMessage.value = 'Please provide both username and password.'
    return
  }

  isSubmitting.value = true
  errorMessage.value = ''

  try {
    await authStore.login(username.value, password.value)
    const redirectPath = (route.query.redirect as string) || '/overview'
    router.push(redirectPath)
  } catch (err: unknown) {
    if (err instanceof Error) {
      errorMessage.value = err.message
    } else {
      errorMessage.value = 'Invalid username or password'
    }
  } finally {
    isSubmitting.value = false
  }
}
</script>

<template>
  <div class="min-h-screen w-screen flex items-center justify-center bg-[#101114] p-4 select-none">
    <div class="w-full max-w-md bg-[#1a1c22] border border-[#2c2e36] rounded-2xl shadow-2xl p-8 space-y-6">
      <!-- Brand Header -->
      <div class="text-center space-y-2">
        <div class="inline-flex h-12 w-12 rounded-xl bg-blue-600/20 border border-blue-500/30 items-center justify-center text-blue-400 mb-2">
          <Sparkles class="h-6 w-6" />
        </div>
        <h2 class="text-xl font-bold tracking-tight text-white">Antigravity Proxy</h2>
        <p class="text-xs text-zinc-400">Usage Telemetry & Control Plane</p>
      </div>

      <!-- Error Alert -->
      <div
        v-if="errorMessage"
        class="p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-red-300 text-xs flex items-center gap-2"
      >
        <AlertCircle class="h-4 w-4 shrink-0 text-red-400" />
        <span>{{ errorMessage }}</span>
      </div>

      <!-- Login Form -->
      <form @submit.prevent="handleLogin" class="space-y-4">
        <div>
          <label class="block text-xs font-medium text-zinc-400 mb-1.5">Username</label>
          <div class="relative">
            <User class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
            <input
              v-model="username"
              type="text"
              required
              autocomplete="username"
              placeholder="admin"
              class="w-full bg-[#141518] border border-[#2c2e36] focus:border-blue-500 rounded-lg pl-9 pr-3 py-2 text-sm text-zinc-100 placeholder:text-zinc-600 focus:outline-none transition-colors"
            />
          </div>
        </div>

        <div>
          <label class="block text-xs font-medium text-zinc-400 mb-1.5">Password</label>
          <div class="relative">
            <Lock class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500" />
            <input
              v-model="password"
              type="password"
              required
              autocomplete="current-password"
              placeholder="••••••••"
              class="w-full bg-[#141518] border border-[#2c2e36] focus:border-blue-500 rounded-lg pl-9 pr-3 py-2 text-sm text-zinc-100 placeholder:text-zinc-600 focus:outline-none transition-colors"
            />
          </div>
        </div>

        <button
          type="submit"
          :disabled="isSubmitting"
          class="w-full mt-2 py-2.5 px-4 rounded-lg bg-blue-600 hover:bg-blue-500 active:bg-blue-700 disabled:opacity-50 text-white font-medium text-sm flex items-center justify-center gap-2 transition-colors cursor-pointer shadow-lg shadow-blue-600/20"
        >
          <span v-if="isSubmitting">Signing in...</span>
          <template v-else>
            <span>Sign In</span>
            <ArrowRight class="h-4 w-4" />
          </template>
        </button>
      </form>

      <!-- Default Credentials Hint -->
      <div class="pt-4 border-t border-[#24262e] text-center text-[11px] text-zinc-500">
        Default credentials: <code class="text-zinc-300 font-mono bg-zinc-800/80 px-1 py-0.5 rounded">admin</code> / <code class="text-zinc-300 font-mono bg-zinc-800/80 px-1 py-0.5 rounded">admin</code>
      </div>
    </div>
  </div>
</template>
