<script setup lang="ts">
import { RouterLink } from 'vue-router'
import { AlertCircle, ArrowLeft } from '@lucide/vue'
</script>

<template>
  <div class="py-24 text-center space-y-4">
    <div class="inline-flex h-16 w-16 rounded-2xl bg-zinc-800/80 border border-zinc-700 items-center justify-center text-zinc-400">
      <AlertCircle class="h-8 w-8 text-amber-400" />
    </div>
    <h2 class="text-xl font-bold text-white tracking-tight">404 — Page Not Found</h2>
    <p class="text-xs text-zinc-500 max-w-sm mx-auto">
      The requested dashboard view or route does not exist.
    </p>
    <div class="pt-2">
      <RouterLink
        to="/overview"
        class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 text-white text-xs font-medium transition-colors"
      >
        <ArrowLeft class="h-3.5 w-3.5" />
        <span>Return to Overview</span>
      </RouterLink>
    </div>
  </div>
</template>
