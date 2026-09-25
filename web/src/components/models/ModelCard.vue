<script setup lang="ts">
import { ref, computed } from 'vue'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Separator } from '@/components/ui/separator'
import {
  Sparkles,
  Bot,
  Layers,
  Cpu,
  FlaskConical,
  Code2,
  Copy,
  Check,
  Zap,
  Radio,
  Eye,
  Wrench,
  Brain,
} from '@lucide/vue'
import type { OpenAIModel } from '@/types'

const props = defineProps<{
  model: OpenAIModel
}>()

const emit = defineEmits<{
  (e: 'test', id: string): void
  (e: 'code', id: string): void
}>()

const copied = ref(false)

const familyInfo = computed(() => {
  const id = props.model.id.toLowerCase()
  if (id.startsWith('claude')) {
    return {
      name: 'Anthropic Claude',
      icon: Bot,
      pillClass: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
      badgeBorder: 'group-hover:border-amber-500/30',
      glow: 'group-hover:shadow-amber-500/5',
      context: '200K tokens',
    }
  }
  if (id.startsWith('gemini')) {
    return {
      name: 'Google Gemini',
      icon: Sparkles,
      pillClass: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
      badgeBorder: 'group-hover:border-blue-500/30',
      glow: 'group-hover:shadow-blue-500/5',
      context: '1M+ tokens',
    }
  }
  if (id.startsWith('gpt-oss') || id.startsWith('gpt') || id.startsWith('openai')) {
    return {
      name: 'Open-Weight / OSS',
      icon: Layers,
      pillClass: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
      badgeBorder: 'group-hover:border-purple-500/30',
      glow: 'group-hover:shadow-purple-500/5',
      context: '128K tokens',
    }
  }
  return {
    name: 'Universal LLM',
    icon: Cpu,
    pillClass: 'bg-zinc-500/10 text-zinc-400 border-zinc-500/20',
    badgeBorder: 'group-hover:border-zinc-500/30',
    glow: 'group-hover:shadow-zinc-500/5',
    context: '128K tokens',
  }
})

const tierInfo = computed(() => {
  const id = props.model.id.toLowerCase()
  if (id.includes('high')) {
    return { label: 'High Reasoning', class: 'bg-indigo-500/15 text-indigo-300 border-indigo-500/30', icon: Brain }
  }
  if (id.includes('medium')) {
    return { label: 'Medium Reasoning', class: 'bg-indigo-500/10 text-indigo-300 border-indigo-500/20', icon: Brain }
  }
  if (id.includes('low') || id.includes('flash-lite')) {
    return { label: 'Fast & Lite', class: 'bg-emerald-500/15 text-emerald-300 border-emerald-500/30', icon: Zap }
  }
  if (id.includes('thinking')) {
    return { label: 'Extended Thinking', class: 'bg-amber-500/15 text-amber-300 border-amber-500/30', icon: Brain }
  }
  if (id.includes('agent')) {
    return { label: 'Autonomous Agent', class: 'bg-sky-500/15 text-sky-300 border-sky-500/30', icon: Bot }
  }
  if (id.includes('pro')) {
    return { label: 'Pro Intelligence', class: 'bg-blue-500/15 text-blue-300 border-blue-500/30', icon: Sparkles }
  }
  return { label: 'General Production', class: 'bg-zinc-800 text-zinc-300 border-zinc-700/60', icon: Cpu }
})

const capabilities = computed(() => {
  const id = props.model.id.toLowerCase()
  const list: { label: string; icon: any }[] = [
    { label: 'Streaming SSE', icon: Radio },
    { label: 'Function Tools', icon: Wrench },
  ]
  if (id.startsWith('gemini') || id.startsWith('claude') || id.includes('4o')) {
    list.push({ label: 'Vision / Files', icon: Eye })
  }
  if (id.includes('thinking') || id.includes('high') || id.includes('pro')) {
    list.push({ label: 'Deep Reasoning', icon: Brain })
  }
  return list
})

async function copyModelId() {
  try {
    await navigator.clipboard.writeText(props.model.id)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (err) {
    console.error('Failed to copy model ID', err)
  }
}
</script>

<template>
  <Card
    class="p-5 bg-[#121316] border-zinc-800/80 transition-all duration-200 flex flex-col justify-between space-y-4 group shadow-sm hover:shadow-xl hover:border-zinc-700"
    :class="[familyInfo.badgeBorder, familyInfo.glow]"
  >
    <!-- Top Row: Provider & Tier -->
    <div class="space-y-3">
      <div class="flex items-center justify-between gap-2">
        <div class="flex items-center gap-1.5">
          <component :is="familyInfo.icon" class="h-3.5 w-3.5" :class="familyInfo.pillClass.split(' ')[1]" />
          <span class="text-[11px] font-semibold text-zinc-400">
            {{ familyInfo.name }}
          </span>
        </div>

        <Badge variant="outline" class="font-mono text-[10px] px-2 py-0.5 gap-1 shrink-0" :class="tierInfo.class">
          <component :is="tierInfo.icon" class="h-2.5 w-2.5" />
          <span>{{ tierInfo.label }}</span>
        </Badge>
      </div>

      <!-- Model Name with Copy Button -->
      <div class="flex items-start justify-between gap-2">
        <h3
          class="font-mono text-sm font-semibold text-white tracking-tight break-all group-hover:text-blue-300 transition-colors select-all leading-snug"
        >
          {{ model.id }}
        </h3>

        <button
          type="button"
          @click="copyModelId"
          class="p-1.5 rounded-lg text-zinc-500 hover:text-white hover:bg-zinc-800/80 transition-all shrink-0 cursor-pointer"
          :title="copied ? 'Copied!' : 'Copy Model ID'"
        >
          <Check v-if="copied" class="h-3.5 w-3.5 text-emerald-400" />
          <Copy v-else class="h-3.5 w-3.5" />
        </button>
      </div>

      <!-- Feature Chips & Context Limit -->
      <div class="flex flex-wrap items-center gap-1.5 pt-1">
        <span class="text-[10px] font-mono px-2 py-0.5 rounded-md bg-[#0d0e11] text-zinc-400 border border-zinc-800/70 flex items-center gap-1">
          <Cpu class="h-2.5 w-2.5 text-zinc-500" />
          {{ familyInfo.context }}
        </span>

        <span
          v-for="(cap, idx) in capabilities"
          :key="idx"
          class="text-[10px] font-mono px-1.5 py-0.5 rounded-md bg-zinc-900/60 text-zinc-400 border border-zinc-800/40 flex items-center gap-1"
        >
          <component :is="cap.icon" class="h-2.5 w-2.5 text-zinc-500" />
          {{ cap.label }}
        </span>
      </div>
    </div>

    <!-- Bottom Actions -->
    <div class="space-y-3 pt-1">
      <Separator class="bg-zinc-800/60" />
      <div class="flex items-center justify-between text-xs text-zinc-400 gap-2">
        <span class="text-zinc-500 text-[11px] truncate">
          Owner: <span class="text-zinc-300 font-mono">{{ model.owned_by }}</span>
        </span>

        <div class="flex items-center gap-1.5 shrink-0">
          <Button
            variant="outline"
            size="sm"
            @click="emit('code', model.id)"
            class="h-7 px-2 text-xs bg-zinc-900/80 border-zinc-800 hover:bg-zinc-800 hover:text-white text-zinc-300 gap-1 transition-all"
            title="View integration code snippet"
          >
            <Code2 class="h-3 w-3" />
            <span class="hidden sm:inline">Code</span>
          </Button>

          <Button
            variant="outline"
            size="sm"
            @click="emit('test', model.id)"
            class="h-7 px-2.5 text-xs bg-zinc-900/80 border-zinc-800 hover:bg-blue-600 hover:text-white hover:border-blue-500 gap-1.5 transition-all text-zinc-300"
          >
            <FlaskConical class="h-3 w-3" />
            <span>Test</span>
          </Button>
        </div>
      </div>
    </div>
  </Card>
</template>
