<script setup lang="ts">
import { computed } from 'vue'
import type { TimeSeriesPoint } from '@/types'

const props = defineProps<{
  points?: TimeSeriesPoint[]
  loading?: boolean
}>()

const chartPoints = computed(() => {
  const pts = props.points || []
  if (pts.length === 0) return []
  if (pts.length === 1) {
    return [
      { ...pts[0], time: pts[0].time + ' ' },
      { ...pts[0] }
    ]
  }
  return pts
})

const maxVal = computed(() => {
  if (chartPoints.value.length === 0) return 10
  const max = Math.max(...chartPoints.value.map(p => Number(p.request_count || (p as any).requests || 0)))
  return max > 0 ? Math.ceil(max * 1.2) : 10
})

const svgPath = computed(() => {
  const pts = chartPoints.value
  if (pts.length < 2) return ''
  const width = 800
  const height = 200

  const coords = pts.map((p, i) => {
    const val = Number(p.request_count || (p as any).requests || 0)
    const x = (i / (pts.length - 1)) * width
    const y = height - (val / maxVal.value) * (height - 30) - 15
    return { x, y }
  })

  // Smooth SVG path
  return coords.reduce((acc, curr, i, arr) => {
    if (i === 0) return `M ${curr.x} ${curr.y}`
    const prev = arr[i - 1]
    const cpX = (prev.x + curr.x) / 2
    return `${acc} C ${cpX} ${prev.y}, ${cpX} ${curr.y}, ${curr.x} ${curr.y}`
  }, '')
})

const svgArea = computed(() => {
  if (!svgPath.value) return ''
  return `${svgPath.value} L 800 200 L 0 200 Z`
})
</script>

<template>
  <div class="p-6 rounded-xl bg-[#202227] border border-[#2c2e36] flex flex-col gap-4">
    <div class="flex items-center justify-between">
      <div>
        <h3 class="text-sm font-semibold text-white tracking-tight">Request Volume Over Time</h3>
        <p class="text-xs text-zinc-500">API throughput activity trends</p>
      </div>
      <div class="flex items-center gap-2 text-xs font-mono text-zinc-400">
        <span class="inline-block w-2.5 h-2.5 rounded-full bg-blue-500"></span>
        <span>Requests</span>
      </div>
    </div>

    <!-- Chart container -->
    <div class="h-56 w-full relative flex items-center justify-center">
      <div v-if="loading" class="text-zinc-500 text-xs animate-pulse">Loading chart data...</div>
      <div v-else-if="chartPoints.length === 0" class="text-zinc-500 text-xs font-mono">
        No telemetry records in selected range.
      </div>
      <svg
        v-else
        viewBox="0 0 800 200"
        preserveAspectRatio="none"
        class="w-full h-full overflow-visible"
      >
        <defs>
          <linearGradient id="areaGradient" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stop-color="#3b82f6" stop-opacity="0.3" />
            <stop offset="100%" stop-color="#3b82f6" stop-opacity="0.0" />
          </linearGradient>
        </defs>

        <!-- Grid Lines -->
        <line x1="0" y1="40" x2="800" y2="40" stroke="#2c2e36" stroke-dasharray="4 4" />
        <line x1="0" y1="100" x2="800" y2="100" stroke="#2c2e36" stroke-dasharray="4 4" />
        <line x1="0" y1="160" x2="800" y2="160" stroke="#2c2e36" stroke-dasharray="4 4" />

        <!-- Area Fill -->
        <path :d="svgArea" fill="url(#areaGradient)" />

        <!-- Stroke line -->
        <path :d="svgPath" fill="none" stroke="#3b82f6" stroke-width="2.5" />
      </svg>
    </div>
  </div>
</template>
