<script setup lang="ts">
import type { Component } from 'vue'

defineProps<{
  title: string
  value: string | number
  subtitle?: string
  icon?: Component
  trend?: string
  color?: 'blue' | 'green' | 'amber' | 'purple' | 'red'
}>()
</script>

<template>
  <div class="p-5 rounded-xl bg-[#202227] border border-[#2c2e36] flex flex-col justify-between transition-all hover:border-[#383a45]">
    <div class="flex items-center justify-between">
      <span class="text-xs font-medium text-zinc-400 uppercase tracking-wider">{{ title }}</span>
      <div
        v-if="icon"
        class="h-8 w-8 rounded-lg flex items-center justify-center"
        :class="{
          'bg-blue-500/10 text-blue-400': color === 'blue' || !color,
          'bg-emerald-500/10 text-emerald-400': color === 'green',
          'bg-amber-500/10 text-amber-400': color === 'amber',
          'bg-purple-500/10 text-purple-400': color === 'purple',
          'bg-red-500/10 text-red-400': color === 'red',
        }"
      >
        <component :is="icon" class="h-4 w-4" />
      </div>
    </div>

    <div class="mt-4 flex items-baseline justify-between">
      <div class="text-2xl font-bold tracking-tight text-white font-mono">{{ value }}</div>
      <div v-if="trend" class="text-xs font-mono text-emerald-400 font-medium">{{ trend }}</div>
    </div>

    <div v-if="subtitle" class="mt-2 text-xs text-zinc-500 truncate">
      {{ subtitle }}
    </div>
  </div>
</template>
