<script setup lang="ts">
import { useRouter } from 'vue-router'
import { AlertCircle, ArrowLeft } from '@lucide/vue'
import { Button } from '@/components/ui/button'

const router = useRouter()
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
      <Button
        @click="router.push('/overview')"
        class="gap-2"
      >
        <ArrowLeft class="h-3.5 w-3.5" />
        <span>Return to Overview</span>
      </Button>
    </div>
  </div>
</template>
