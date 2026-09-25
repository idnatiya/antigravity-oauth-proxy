<script setup lang="ts">
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { Sparkles, Lock, User, AlertCircle, ArrowRight, CheckCircle2 } from '@lucide/vue'
import { useAuthStore } from '@/stores/authStore'
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
  <div class="min-h-screen w-screen grid grid-cols-1 lg:grid-cols-[1.15fr_1fr] bg-[#0b0c0e] select-none overflow-x-hidden">
    <!-- Left Hero & Telemetry Preview Section (Desktop only) -->
    <div class="hidden lg:flex flex-col justify-between p-12 lg:p-16 bg-[#08090b] border-r border-zinc-800/80 relative overflow-hidden">
      <!-- Ambient Glows -->
      <div class="absolute -top-28 -left-28 w-96 h-96 rounded-full bg-blue-500/10 blur-[100px] pointer-events-none" />
      <div class="absolute -bottom-28 -right-28 w-96 h-96 rounded-full bg-emerald-500/10 blur-[100px] pointer-events-none" />

      <!-- Top Brand -->
      <div class="flex items-center gap-3 relative z-10">
        <div class="h-10 w-10 rounded-xl bg-gradient-to-br from-blue-500/20 to-indigo-500/10 border border-blue-500/30 flex items-center justify-center text-blue-400 shadow-[0_0_20px_-3px_rgba(59,130,246,0.3)]">
          <Sparkles class="h-5 w-5" />
        </div>
        <div>
          <div class="font-bold text-base tracking-tight text-white flex items-center gap-2">
            <span>Antigravity Proxy</span>
            <span class="inline-flex items-center px-1.5 py-0.5 rounded text-[10px] font-mono font-medium bg-blue-500/15 text-blue-400 border border-blue-500/30">
              PROXY
            </span>
          </div>
          <div class="text-[11px] text-zinc-500 font-mono mt-0.5">v1beta • SQLite WAL</div>
        </div>
      </div>

      <!-- Center Value Proposition & Mockup Card -->
      <div class="my-auto py-8 relative z-10 space-y-6">
        <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-blue-500/25 bg-blue-500/10 text-blue-400 font-mono text-xs shadow-sm">
          <span class="relative flex h-2 w-2">
            <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-blue-400 opacity-75"></span>
            <span class="relative inline-flex rounded-full h-2 w-2 bg-blue-500"></span>
          </span>
          <span>AI Gateway & Usage Telemetry</span>
        </div>

        <h1 class="text-3xl lg:text-4xl font-bold tracking-tight text-white leading-tight">
          Real-time LLM Usage Telemetry & Cost Control.
        </h1>

        <p class="text-sm text-zinc-400 max-w-lg leading-relaxed">
          Monitor token consumption, track model latency, and view historical request logs for Antigravity-powered Gemini & OpenAI clients.
        </p>

        <!-- Simulated Telemetry Window Card -->
        <div class="max-w-md rounded-xl bg-[#121316]/95 border border-zinc-800/80 p-4 font-mono text-xs backdrop-blur-md shadow-2xl space-y-3">
          <!-- Window Title Bar -->
          <div class="flex items-center justify-between border-b border-zinc-800/80 pb-2.5">
            <div class="flex items-center gap-1.5">
              <span class="h-2.5 w-2.5 rounded-full bg-red-500/80"></span>
              <span class="h-2.5 w-2.5 rounded-full bg-amber-500/80"></span>
              <span class="h-2.5 w-2.5 rounded-full bg-emerald-500/80"></span>
              <span class="text-[11px] text-zinc-500 ml-2">antigravity-proxy :8080</span>
            </div>
            <span class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-[10px] bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-semibold">
              <span class="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse"></span>
              CONNECTED
            </span>
          </div>

          <!-- Telemetry Table Content -->
          <div class="space-y-1.5">
            <div class="flex justify-between text-[11px] text-zinc-500 px-1">
              <span>MODEL</span>
              <span>STATUS</span>
              <span>TOKENS</span>
            </div>

            <div class="flex justify-between items-center text-xs p-1.5 rounded bg-zinc-900/60 border border-zinc-800/50">
              <span class="text-zinc-200">gemini-2.5-flash</span>
              <span class="text-emerald-400 font-semibold">200 OK</span>
              <span class="text-zinc-400">1.4K tok</span>
            </div>

            <div class="flex justify-between items-center text-xs p-1.5 rounded bg-zinc-900/60 border border-zinc-800/50">
              <span class="text-zinc-200">claude-3-7-sonnet</span>
              <span class="text-emerald-400 font-semibold">200 OK</span>
              <span class="text-zinc-400">850 tok</span>
            </div>

            <div class="flex justify-between items-center text-xs p-1.5 rounded bg-zinc-900/60 border border-zinc-800/50">
              <span class="text-zinc-200">gemini-2.5-pro</span>
              <span class="text-emerald-400 font-semibold">200 OK</span>
              <span class="text-zinc-400">4.2K tok</span>
            </div>
          </div>
        </div>

        <!-- Feature Bullet Checkmarks -->
        <div class="grid grid-cols-2 gap-3 pt-2 text-xs text-zinc-400">
          <div class="flex items-center gap-2">
            <CheckCircle2 class="h-3.5 w-3.5 text-blue-400 shrink-0" />
            <span>SQLite Request History</span>
          </div>
          <div class="flex items-center gap-2">
            <CheckCircle2 class="h-3.5 w-3.5 text-blue-400 shrink-0" />
            <span>Multi-Account OAuth Pool</span>
          </div>
          <div class="flex items-center gap-2">
            <CheckCircle2 class="h-3.5 w-3.5 text-blue-400 shrink-0" />
            <span>OpenAI & Gemini API Parity</span>
          </div>
          <div class="flex items-center gap-2">
            <CheckCircle2 class="h-3.5 w-3.5 text-blue-400 shrink-0" />
            <span>Zero-Latency Interceptor</span>
          </div>
        </div>
      </div>

      <!-- Quote Footer -->
      <div class="border-t border-zinc-800/60 pt-4 relative z-10">
        <blockquote class="text-xs text-zinc-400 italic">
          "Simplicity is prerequisite for reliability."
        </blockquote>
        <div class="text-[11px] text-zinc-500 font-mono mt-1">
          — Edsger W. Dijkstra
        </div>
      </div>
    </div>

    <!-- Right Block: Sign In Form -->
    <div class="flex items-center justify-center p-6 sm:p-12 relative z-10 bg-[#0b0c0e]">
      <div class="w-full max-w-sm space-y-6">
        <!-- Mobile Brand Header (Visible on < lg) -->
        <div class="lg:hidden text-center space-y-2 mb-6">
          <div class="inline-flex h-12 w-12 rounded-xl bg-gradient-to-br from-blue-500/20 to-indigo-500/10 border border-blue-500/30 items-center justify-center text-blue-400 mx-auto mb-1 shadow-lg shadow-blue-500/10">
            <Sparkles class="h-6 w-6" />
          </div>
          <h2 class="text-xl font-bold tracking-tight text-white">Antigravity Proxy</h2>
          <p class="text-xs text-zinc-400 font-mono">Control Plane & Telemetry</p>
        </div>

        <!-- Form Heading -->
        <div class="space-y-1">
          <h2 class="text-2xl font-bold tracking-tight text-white">Sign In</h2>
          <p class="text-xs text-zinc-400">
            Enter your dashboard credentials to access the control plane.
          </p>
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
          <div class="space-y-1.5">
            <Label for="login-username" class="text-xs font-medium text-zinc-300">Username</Label>
            <div class="relative">
              <User class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500 pointer-events-none" />
              <Input
                id="login-username"
                v-model="username"
                type="text"
                required
                autocomplete="username"
                placeholder="admin"
                class="bg-[#121316] border-zinc-800 pl-9 text-sm h-10 text-zinc-100 focus-visible:ring-blue-500"
              />
            </div>
          </div>

          <div class="space-y-1.5">
            <div class="flex items-center justify-between">
              <Label for="login-password" class="text-xs font-medium text-zinc-300">Password</Label>
            </div>
            <div class="relative">
              <Lock class="h-4 w-4 absolute left-3 top-1/2 -translate-y-1/2 text-zinc-500 pointer-events-none" />
              <Input
                id="login-password"
                v-model="password"
                type="password"
                required
                autocomplete="current-password"
                placeholder="••••••••"
                class="bg-[#121316] border-zinc-800 pl-9 text-sm h-10 text-zinc-100 focus-visible:ring-blue-500"
              />
            </div>
          </div>

          <Button
            type="submit"
            size="lg"
            :disabled="isSubmitting"
            class="w-full mt-2 bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-600/20 font-semibold"
          >
            <span v-if="isSubmitting">Signing in...</span>
            <template v-else>
              <span>Sign In</span>
              <ArrowRight class="h-4 w-4 ml-1.5" />
            </template>
          </Button>
        </form>

        <Separator class="bg-zinc-800/80 my-4" />

        <!-- Default Credentials Hint -->
        <div class="text-center text-[11px] text-zinc-500">
          Default credentials:
          <code class="text-zinc-300 font-mono bg-zinc-800/80 px-1.5 py-0.5 rounded mx-1">admin</code> /
          <code class="text-zinc-300 font-mono bg-zinc-800/80 px-1.5 py-0.5 rounded mx-1">admin</code>
        </div>
      </div>
    </div>
  </div>
</template>
