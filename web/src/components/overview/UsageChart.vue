<script setup lang="ts">
import { ref, computed } from 'vue'
import { Card } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import type { TimeSeriesPoint } from '@/types'

const props = defineProps<{
  points?: TimeSeriesPoint[]
  loading?: boolean
}>()

const activeMetric = ref<'requests' | 'tokens'>('requests')

// Interactive hover state
const hoveredIndex = ref<number | null>(null)
const chartSvgRef = ref<SVGSVGElement | null>(null)

const chartPoints = computed(() => {
  const pts = props.points || []
  if (pts.length === 0) return []
  if (pts.length === 1) {
    return [
      { ...pts[0], time: pts[0].time },
      { ...pts[0] },
    ]
  }
  return pts
})

function getPointValue(p: TimeSeriesPoint): number {
  if (activeMetric.value === 'tokens') {
    return Number(p.total_tokens || 0)
  }
  return Number(p.request_count || (p as any).requests || 0)
}

const maxVal = computed(() => {
  if (chartPoints.value.length === 0) return 10
  const max = Math.max(...chartPoints.value.map(p => getPointValue(p)))
  return max > 0 ? Math.ceil(max * 1.15) : 10
})

const totalMetricValue = computed(() => {
  return chartPoints.value.reduce((acc, p) => acc + getPointValue(p), 0)
})

const peakMetricValue = computed(() => {
  if (chartPoints.value.length === 0) return 0
  return Math.max(...chartPoints.value.map(p => getPointValue(p)))
})

const svgWidth = 800
const svgHeight = 220
const topPadding = 20
const bottomPadding = 20

const computedCoords = computed(() => {
  const pts = chartPoints.value
  if (pts.length < 2) return []

  return pts.map((p, i) => {
    const val = getPointValue(p)
    const x = (i / (pts.length - 1)) * svgWidth
    const y = svgHeight - bottomPadding - (val / maxVal.value) * (svgHeight - topPadding - bottomPadding)
    return { x, y, raw: p, val }
  })
})

const svgPath = computed(() => {
  const coords = computedCoords.value
  if (coords.length < 2) return ''

  // Smooth Bezier Curve
  return coords.reduce((acc, curr, i, arr) => {
    if (i === 0) return `M ${curr.x} ${curr.y}`
    const prev = arr[i - 1]
    const cpX = (prev.x + curr.x) / 2
    return `${acc} C ${cpX} ${prev.y}, ${cpX} ${curr.y}, ${curr.x} ${curr.y}`
  }, '')
})

const svgArea = computed(() => {
  if (!svgPath.value) return ''
  return `${svgPath.value} L ${svgWidth} ${svgHeight} L 0 ${svgHeight} Z`
})

function handleMouseMove(e: MouseEvent) {
  if (!chartSvgRef.value || computedCoords.value.length === 0) return
  const rect = chartSvgRef.value.getBoundingClientRect()
  const mouseX = e.clientX - rect.left
  const relativeX = (mouseX / rect.width) * svgWidth

  // Find closest point by X coordinate
  let closestIdx = 0
  let minDist = Infinity
  computedCoords.value.forEach((c, idx) => {
    const dist = Math.abs(c.x - relativeX)
    if (dist < minDist) {
      minDist = dist
      closestIdx = idx
    }
  })
  hoveredIndex.value = closestIdx
}

function handleMouseLeave() {
  hoveredIndex.value = null
}

const activeHoverCoord = computed(() => {
  if (hoveredIndex.value === null) return null
  return computedCoords.value[hoveredIndex.value] || null
})

function formatDisplayValue(val: number): string {
  if (activeMetric.value === 'tokens') {
    if (val >= 1_000_000) return (val / 1_000_000).toFixed(2) + 'M tokens'
    if (val >= 1_000) return (val / 1_000).toFixed(1) + 'k tokens'
    return val.toLocaleString() + ' tokens'
  }
  return val.toLocaleString() + ' requests'
}
</script>

<template>
  <Card class="p-6 bg-[#121316] border-zinc-800/80 shadow-sm flex flex-col gap-5">
    <!-- Header with Metric Toggles -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2">
          <h3 class="text-base font-semibold text-white tracking-tight">API Activity Trends</h3>
          <span class="inline-flex items-center px-2 py-0.5 rounded text-xs font-mono font-medium bg-zinc-800 text-zinc-300">
            Interactive
          </span>
        </div>
        <p class="text-sm text-zinc-400 mt-0.5">Real-time throughput metrics over the selected period</p>
      </div>

      <!-- Metric Toggle Buttons -->
      <div class="flex items-center p-1 rounded-lg bg-zinc-900/90 border border-zinc-800/80 self-start sm:self-auto">
        <button
          @click="activeMetric = 'requests'"
          class="px-3.5 py-1.5 rounded-md text-sm font-medium transition-all cursor-pointer"
          :class="[
            activeMetric === 'requests'
              ? 'bg-blue-600 text-white shadow-sm font-semibold'
              : 'text-zinc-400 hover:text-zinc-200'
          ]"
        >
          Requests
        </button>
        <button
          @click="activeMetric = 'tokens'"
          class="px-3.5 py-1.5 rounded-md text-sm font-medium transition-all cursor-pointer"
          :class="[
            activeMetric === 'tokens'
              ? 'bg-purple-600 text-white shadow-sm font-semibold'
              : 'text-zinc-400 hover:text-zinc-200'
          ]"
        >
          Tokens
        </button>
      </div>
    </div>

    <!-- Chart Container -->
    <div class="relative h-60 w-full flex items-center justify-center select-none">
      <!-- Loading Skeleton -->
      <div v-if="loading" class="w-full h-full flex flex-col gap-2 justify-center">
        <Skeleton class="w-full h-44 rounded-lg" />
        <div class="flex justify-between">
          <Skeleton class="w-20 h-4" />
          <Skeleton class="w-20 h-4" />
          <Skeleton class="w-20 h-4" />
        </div>
      </div>

      <!-- Empty State -->
      <div v-else-if="chartPoints.length === 0" class="text-zinc-500 text-xs font-mono py-12">
        No telemetry records logged for this time range.
      </div>

      <!-- Interactive SVG Chart -->
      <div
        v-else
        class="w-full h-full relative"
        @mousemove="handleMouseMove"
        @mouseleave="handleMouseLeave"
      >
        <svg
          ref="chartSvgRef"
          :viewBox="`0 0 ${svgWidth} ${svgHeight}`"
          preserveAspectRatio="none"
          class="w-full h-full overflow-visible cursor-crosshair"
        >
          <defs>
            <!-- Blue Gradient for Requests -->
            <linearGradient id="reqGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#3b82f6" stop-opacity="0.35" />
              <stop offset="85%" stop-color="#3b82f6" stop-opacity="0.02" />
              <stop offset="100%" stop-color="#3b82f6" stop-opacity="0" />
            </linearGradient>

            <!-- Purple Gradient for Tokens -->
            <linearGradient id="tokenGradient" x1="0" y1="0" x2="0" y2="1">
              <stop offset="0%" stop-color="#a855f7" stop-opacity="0.35" />
              <stop offset="85%" stop-color="#a855f7" stop-opacity="0.02" />
              <stop offset="100%" stop-color="#a855f7" stop-opacity="0" />
            </linearGradient>
          </defs>

          <!-- Horizontal Grid Lines -->
          <line x1="0" y1="40" :x2="svgWidth" y2="40" stroke="#26272e" stroke-dasharray="3 3" stroke-width="1" />
          <line x1="0" y1="100" :x2="svgWidth" y2="100" stroke="#26272e" stroke-dasharray="3 3" stroke-width="1" />
          <line x1="0" y1="160" :x2="svgWidth" y2="160" stroke="#26272e" stroke-dasharray="3 3" stroke-width="1" />
          <line x1="0" :y1="svgHeight - bottomPadding" :x2="svgWidth" :y2="svgHeight - bottomPadding" stroke="#26272e" stroke-width="1" />

          <!-- Area Fill -->
          <path
            :d="svgArea"
            :fill="activeMetric === 'requests' ? 'url(#reqGradient)' : 'url(#tokenGradient)'"
          />

          <!-- Main Curve -->
          <path
            :d="svgPath"
            fill="none"
            :stroke="activeMetric === 'requests' ? '#3b82f6' : '#a855f7'"
            stroke-width="2.5"
            stroke-linecap="round"
            stroke-linejoin="round"
          />

          <!-- Active Hover Guidelines & Indicator -->
          <g v-if="activeHoverCoord">
            <!-- Vertical Guideline -->
            <line
              :x1="activeHoverCoord.x"
              y1="0"
              :x2="activeHoverCoord.x"
              :y2="svgHeight - bottomPadding"
              stroke="#60a5fa"
              stroke-dasharray="3 3"
              stroke-width="1.5"
              opacity="0.8"
            />

            <!-- Outer pulsing ring -->
            <circle
              :cx="activeHoverCoord.x"
              :cy="activeHoverCoord.y"
              r="7"
              :fill="activeMetric === 'requests' ? '#3b82f6' : '#a855f7'"
              opacity="0.3"
            />
            <!-- Inner Dot -->
            <circle
              :cx="activeHoverCoord.x"
              :cy="activeHoverCoord.y"
              r="4"
              :fill="activeMetric === 'requests' ? '#60a5fa' : '#c084fc'"
              stroke="#0e0f12"
              stroke-width="2"
            />
          </g>
        </svg>

        <!-- Floating Tooltip Box -->
        <div
          v-if="activeHoverCoord"
          class="absolute pointer-events-none z-10 -translate-x-1/2 -translate-y-full px-3 py-2 rounded-lg bg-zinc-900/95 border border-zinc-700/80 shadow-xl backdrop-blur-sm text-xs transition-all duration-75 flex flex-col gap-0.5"
          :style="{
            left: `${(activeHoverCoord.x / svgWidth) * 100}%`,
            top: `${(activeHoverCoord.y / svgHeight) * 100 - 8}px`,
          }"
        >
          <div class="text-[10px] text-zinc-400 font-mono">{{ activeHoverCoord.raw.time }}</div>
          <div class="font-bold text-white font-mono flex items-center gap-1.5">
            <span
              class="h-2 w-2 rounded-full"
              :class="activeMetric === 'requests' ? 'bg-blue-400' : 'bg-purple-400'"
            />
            <span>{{ formatDisplayValue(activeHoverCoord.val) }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Chart Footer Metrics Strip -->
    <div
      v-if="!loading && chartPoints.length > 0"
      class="pt-3.5 border-t border-zinc-800/60 grid grid-cols-2 sm:grid-cols-3 gap-4 text-sm font-mono text-zinc-400"
    >
      <div>
        <span class="text-zinc-500 text-xs block">Total Volume:</span>
        <span class="text-zinc-200 font-semibold">{{ formatDisplayValue(totalMetricValue) }}</span>
      </div>
      <div>
        <span class="text-zinc-500 text-xs block">Peak Record:</span>
        <span class="text-zinc-200 font-semibold">{{ formatDisplayValue(peakMetricValue) }}</span>
      </div>
      <div class="hidden sm:block">
        <span class="text-zinc-500 text-xs block">Interval Points:</span>
        <span class="text-zinc-200 font-semibold">{{ chartPoints.length }} intervals</span>
      </div>
    </div>
  </Card>
</template>
