<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()
</script>

<template>
  <tr :class="cn('border-b border-[#2c2e36] transition-colors hover:bg-[#252830] data-[state=selected]:bg-[#282a32]', props.class)">
    <slot />
  </tr>
</template>
