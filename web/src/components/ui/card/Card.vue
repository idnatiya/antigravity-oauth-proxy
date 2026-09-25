<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()
</script>

<template>
  <div :class="cn('rounded-2xl border border-[#2c2e36] bg-[#202227] text-zinc-100 shadow-sm transition-all', props.class)">
    <slot />
  </div>
</template>
