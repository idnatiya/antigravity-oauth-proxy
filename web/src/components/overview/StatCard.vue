<script setup lang="ts">
import type { Component } from 'vue'
import { Card } from '@/components/ui/card'

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
  <Card
    class="relative p-5 bg-[#121316] border-zinc-800/80 hover:border-zinc-700/80 flex flex-col justify-between transition-all duration-200 shadow-sm hover:shadow-lg hover:shadow-black/30 overflow-hidden group"
  >
    <!-- Top Accent Line -->
    <div
      class="absolute top-0 left-0 right-0 h-[2px] transition-opacity duration-200"
      :class="{
        'bg-gradient-to-r from-blue-500 to-indigo-500': color === 'blue' || !color,
        'bg-gradient-to-r from-emerald-500 to-teal-500': color === 'green',
        'bg-gradient-to-r from-purple-500 to-pink-500': color === 'purple',
        'bg-gradient-to-r from-amber-500 to-yellow-500': color === 'amber',
        'bg-gradient-to-r from-red-500 to-rose-500': color === 'red',
      }"
    />

    <div>
      <div class="flex items-center justify-between">
        <span class="text-xs font-semibold text-zinc-400 uppercase tracking-wider">{{ title }}</span>
        <div
          v-if="icon"
          class="h-9 w-9 rounded-lg flex items-center justify-center transition-transform group-hover:scale-105 duration-200"
          :class="{
            'bg-blue-500/10 text-blue-400 border border-blue-500/20 shadow-[0_0_12px_-2px_rgba(59,130,246,0.25)]': color === 'blue' || !color,
            'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 shadow-[0_0_12px_-2px_rgba(16,185,129,0.25)]': color === 'green',
            'bg-purple-500/10 text-purple-400 border border-purple-500/20 shadow-[0_0_12px_-2px_rgba(168,85,247,0.25)]': color === 'purple',
            'bg-amber-500/10 text-amber-400 border border-amber-500/20 shadow-[0_0_12px_-2px_rgba(245,158,11,0.25)]': color === 'amber',
            'bg-red-500/10 text-red-400 border border-red-500/20 shadow-[0_0_12px_-2px_rgba(239,68,68,0.25)]': color === 'red',
          }"
        >
          <component :is="icon" class="h-4.5 w-4.5" />
        </div>
      </div>

      <div class="mt-4 flex items-baseline justify-between">
        <div class="text-3xl font-bold tracking-tight text-white font-mono">{{ value }}</div>
        <div v-if="trend" class="text-xs font-mono text-emerald-400 font-medium flex items-center gap-1">
          <span>↑</span>
          <span>{{ trend }}</span>
        </div>
      </div>
    </div>

    <div v-if="subtitle" class="mt-3 text-sm text-zinc-400 font-mono truncate">
      {{ subtitle }}
    </div>
  </Card>
</template>
