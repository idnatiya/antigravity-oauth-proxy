<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  defaultValue?: string | number
  modelValue?: string | number
  class?: HTMLAttributes['class']
}>()

const emits = defineEmits<{
  (e: 'update:modelValue', payload: string | number): void
}>()
</script>

<template>
  <input
    :value="modelValue"
    @input="emits('update:modelValue', ($event.target as HTMLInputElement).value)"
    :class="
      cn(
        'flex h-9 w-full rounded-xl border border-[#2c2e36] bg-[#18191d] px-3 py-1 text-xs text-zinc-100 shadow-inner transition-colors placeholder:text-zinc-600 focus-visible:outline-none focus-visible:border-blue-500 focus-visible:ring-1 focus-visible:ring-blue-500 disabled:cursor-not-allowed disabled:opacity-50 font-sans',
        props.class
      )
    "
  />
</template>
