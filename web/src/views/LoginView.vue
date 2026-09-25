<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Sparkles, Lock, User, AlertCircle, ArrowRight } from '@lucide/vue'
import { useAuthStore } from '@/stores/authStore'
import { Card, CardHeader, CardTitle, CardDescription, CardContent, CardFooter } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'

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
    <Card class="w-full max-w-md bg-[#1a1c22] border-[#2c2e36] shadow-2xl p-6 sm:p-8 space-y-6">
      <!-- Brand Header -->
      <CardHeader class="p-0 text-center space-y-2">
        <div class="inline-flex h-12 w-12 rounded-xl bg-blue-600/20 border border-blue-500/30 items-center justify-center text-blue-400 mx-auto mb-1">
          <Sparkles class="h-6 w-6" />
        </div>
        <CardTitle class="text-xl font-bold tracking-tight text-white">Antigravity Proxy</CardTitle>
        <CardDescription class="text-xs text-zinc-400">Usage Telemetry & Control Plane</CardDescription>
      </CardHeader>

      <!-- Error Alert -->
      <div
        v-if="errorMessage"
        class="p-3 rounded-lg bg-red-500/10 border border-red-500/20 text-red-300 text-xs flex items-center gap-2"
      >
        <AlertCircle class="h-4 w-4 shrink-0 text-red-400" />
        <span>{{ errorMessage }}</span>
      </div>

      <!-- Login Form -->
      <CardContent class="p-0">
        <form @submit.prevent="handleLogin" class="space-y-4">
          <div class="space-y-1.5">
            <Label for="login-username" class="text-xs text-zinc-400">Username</Label>
            <div class="relative">
              <User class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500 pointer-events-none" />
              <Input
                id="login-username"
                v-model="username"
                type="text"
                required
                autocomplete="username"
                placeholder="admin"
                class="bg-[#141518] border-[#2c2e36] pl-9 text-sm h-10 text-zinc-100"
              />
            </div>
          </div>

          <div class="space-y-1.5">
            <Label for="login-password" class="text-xs text-zinc-400">Password</Label>
            <div class="relative">
              <Lock class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500 pointer-events-none" />
              <Input
                id="login-password"
                v-model="password"
                type="password"
                required
                autocomplete="current-password"
                placeholder="••••••••"
                class="bg-[#141518] border-[#2c2e36] pl-9 text-sm h-10 text-zinc-100"
              />
            </div>
          </div>

          <Button
            type="submit"
            size="lg"
            :disabled="isSubmitting"
            class="w-full mt-2"
          >
            <span v-if="isSubmitting">Signing in...</span>
            <template v-else>
              <span>Sign In</span>
              <ArrowRight class="h-4 w-4 ml-1" />
            </template>
          </Button>
        </form>
      </CardContent>

      <Separator class="bg-[#24262e]" />

      <!-- Default Credentials Hint -->
      <CardFooter class="p-0 justify-center text-[11px] text-zinc-500">
        Default credentials: <code class="text-zinc-300 font-mono bg-zinc-800/80 px-1 py-0.5 rounded mx-1">admin</code> / <code class="text-zinc-300 font-mono bg-zinc-800/80 px-1 py-0.5 rounded mx-1">admin</code>
      </CardFooter>
    </Card>
  </div>
</template>
