<script setup lang="ts">
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import {
  Search,
  ChevronDown,
  Check,
  Cpu,
  Sparkles,
  X,
  PlusCircle,
} from '@lucide/vue'
import { Badge } from '@/components/ui/badge'

const props = defineProps<{
  modelValue: string
  models: string[]
  disabled?: boolean
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
}>()

const isOpen = ref(false)
const searchQuery = ref('')
const activeCategory = ref<'all' | 'gemini' | 'claude' | 'other'>('all')
const containerRef = ref<HTMLElement | null>(null)
const searchInputRef = ref<HTMLInputElement | null>(null)

function getFamily(id: string): string {
  if (id.startsWith('claude')) return 'Claude'
  if (id.startsWith('gemini')) return 'Gemini'
  if (id.startsWith('gpt-oss')) return 'OSS'
  return 'Other'
}

function getTierBadge(id: string): { label: string; variant: 'default' | 'secondary' | 'warning' | 'success' | 'outline' } {
  if (id.includes('high')) return { label: 'High Reasoning', variant: 'default' }
  if (id.includes('medium')) return { label: 'Medium Reasoning', variant: 'default' }
  if (id.includes('low') || id.includes('flash-lite')) return { label: 'Fast / Lite', variant: 'success' }
  if (id.includes('thinking')) return { label: 'Thinking', variant: 'warning' }
  if (id.includes('agent')) return { label: 'Agent', variant: 'default' }
  return { label: 'Standard', variant: 'secondary' }
}

const geminiCount = computed(() => props.models.filter(m => m.startsWith('gemini')).length)
const claudeCount = computed(() => props.models.filter(m => m.startsWith('claude')).length)
const otherCount = computed(() => props.models.filter(m => !m.startsWith('gemini') && !m.startsWith('claude')).length)

const categories = computed(() => [
  { key: 'all' as const, label: 'All', count: props.models.length },
  { key: 'gemini' as const, label: 'Gemini', count: geminiCount.value },
  { key: 'claude' as const, label: 'Claude', count: claudeCount.value },
  ...(otherCount.value > 0 ? [{ key: 'other' as const, label: 'Other', count: otherCount.value }] : []),
])

const filteredModels = computed(() => {
  let list = props.models

  if (activeCategory.value === 'gemini') {
    list = list.filter(m => m.startsWith('gemini'))
  } else if (activeCategory.value === 'claude') {
    list = list.filter(m => m.startsWith('claude'))
  } else if (activeCategory.value === 'other') {
    list = list.filter(m => !m.startsWith('gemini') && !m.startsWith('claude'))
  }

  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return list

  return list.filter(m => m.toLowerCase().includes(q) || getFamily(m).toLowerCase().includes(q))
})

const exactMatchExists = computed(() => {
  const q = searchQuery.value.trim().toLowerCase()
  if (!q) return true
  return props.models.some(m => m.toLowerCase() === q)
})

function toggleDropdown() {
  if (props.disabled) return
  isOpen.value = !isOpen.value
  if (isOpen.value) {
    searchQuery.value = ''
    nextTick(() => {
      searchInputRef.value?.focus()
    })
  }
}

function selectModel(model: string) {
  emit('update:modelValue', model)
  isOpen.value = false
  searchQuery.value = ''
}

function onSearchKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    isOpen.value = false
  } else if (e.key === 'Enter') {
    e.preventDefault()
    if (filteredModels.value.length > 0) {
      selectModel(filteredModels.value[0])
    } else if (searchQuery.value.trim()) {
      selectModel(searchQuery.value.trim())
    }
  }
}

function handleClickOutside(e: MouseEvent) {
  if (containerRef.value && !containerRef.value.contains(e.target as Node)) {
    isOpen.value = false
  }
}

watch(isOpen, (open) => {
  if (open) {
    document.addEventListener('click', handleClickOutside)
  } else {
    document.removeEventListener('click', handleClickOutside)
  }
})

onMounted(() => {
  // cleanup if unmounted
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})
</script>

<template>
  <div ref="containerRef" class="relative w-full">
    <!-- Trigger Button -->
    <button
      type="button"
      :disabled="disabled"
      @click="toggleDropdown"
      class="w-full flex items-center justify-between p-2.5 rounded-lg bg-[#0d0e11] border border-zinc-800/80 hover:border-zinc-700 focus:border-blue-500/80 focus:outline-none transition-all cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed group text-left"
    >
      <div class="flex items-center gap-2.5 min-w-0 pr-2">
        <div class="p-1.5 rounded-md bg-zinc-900 border border-zinc-800 text-blue-400 shrink-0 group-hover:scale-105 transition-transform">
          <Cpu class="h-3.5 w-3.5" />
        </div>
        <div class="min-w-0 flex-1">
          <div class="flex items-center gap-2">
            <span class="font-mono text-xs font-semibold text-white truncate">{{ modelValue }}</span>
            <Badge variant="outline" class="font-mono text-[9px] px-1.5 py-0 h-4 shrink-0 text-zinc-400 border-zinc-800">
              {{ getFamily(modelValue) }}
            </Badge>
          </div>
          <div class="text-[10px] text-zinc-500 truncate">
            {{ getTierBadge(modelValue).label }}
          </div>
        </div>
      </div>

      <ChevronDown
        class="h-4 w-4 text-zinc-500 shrink-0 transition-transform duration-150"
        :class="{ 'rotate-180 text-blue-400': isOpen }"
      />
    </button>

    <!-- Search & Selection Popover Dropdown -->
    <div
      v-if="isOpen"
      class="absolute left-0 top-full mt-2 w-full z-50 rounded-xl bg-[#121316] border border-zinc-700/80 shadow-2xl shadow-black ring-1 ring-white/5 overflow-hidden animate-in fade-in zoom-in-95 duration-100 flex flex-col max-h-[420px]"
    >
      <!-- Search Bar -->
      <div class="p-3 border-b border-zinc-800/80 bg-zinc-900/50 flex items-center gap-2.5">
        <Search class="h-3.5 w-3.5 text-zinc-400 shrink-0" />
        <input
          ref="searchInputRef"
          v-model="searchQuery"
          type="text"
          placeholder="Filter models by name, provider, or tier..."
          @keydown="onSearchKeydown"
          class="w-full bg-transparent text-xs text-zinc-100 placeholder:text-zinc-500 focus:outline-none font-sans"
        />
        <button
          v-if="searchQuery"
          type="button"
          @click="searchQuery = ''"
          class="text-zinc-500 hover:text-zinc-300 p-0.5 rounded cursor-pointer"
        >
          <X class="h-3.5 w-3.5" />
        </button>
        <span
          v-else
          class="hidden sm:inline-block text-[10px] font-mono text-zinc-500 bg-zinc-900 border border-zinc-800 px-1.5 py-0.5 rounded"
        >
          ESC
        </span>
      </div>

      <!-- Category Filter Pills Bar -->
      <div class="px-3 py-2 border-b border-zinc-800/80 bg-zinc-900/30 flex items-center gap-1.5 overflow-x-auto no-scrollbar">
        <button
          v-for="cat in categories"
          :key="cat.key"
          type="button"
          @click="activeCategory = cat.key"
          class="px-2.5 py-1 rounded-full text-xs transition-all flex items-center gap-1.5 cursor-pointer shrink-0"
          :class="activeCategory === cat.key
            ? 'bg-blue-600 text-white font-medium shadow-xs shadow-blue-500/25'
            : 'bg-zinc-900/80 hover:bg-zinc-800 text-zinc-400 hover:text-zinc-200 border border-zinc-800/80'"
        >
          <span>{{ cat.label }}</span>
          <span
            class="text-[10px] font-mono rounded-full px-1.5 py-0.2"
            :class="activeCategory === cat.key ? 'bg-white/20 text-white' : 'bg-zinc-800 text-zinc-500'"
          >
            {{ cat.count }}
          </span>
        </button>
      </div>

      <!-- Models List (Single-Line Compact High-Density Rows) -->
      <div class="flex-1 overflow-y-auto p-1.5 space-y-0.5 max-h-64">
        <button
          v-for="m in filteredModels"
          :key="m"
          type="button"
          @click="selectModel(m)"
          class="w-full flex items-center justify-between px-3 py-2 rounded-lg text-left transition-colors cursor-pointer group"
          :class="m === modelValue
            ? 'bg-blue-600/10 text-white border border-blue-500/30'
            : 'hover:bg-zinc-800/60 text-zinc-300 border border-transparent'"
        >
          <!-- Left: Provider Icon & Model Name -->
          <div class="flex items-center gap-2.5 min-w-0 pr-2">
            <div class="shrink-0">
              <Sparkles
                v-if="m.startsWith('gemini')"
                class="h-3.5 w-3.5 text-blue-400"
              />
              <Sparkles
                v-else-if="m.startsWith('claude')"
                class="h-3.5 w-3.5 text-amber-400"
              />
              <Cpu
                v-else
                class="h-3.5 w-3.5 text-emerald-400"
              />
            </div>
            <span
              class="font-mono text-xs font-semibold truncate group-hover:text-blue-300 transition-colors"
              :class="m === modelValue ? 'text-blue-400' : 'text-zinc-200'"
            >
              {{ m }}
            </span>
          </div>

          <!-- Right: Tier Badge & Active Checkmark -->
          <div class="flex items-center gap-2 shrink-0">
            <Badge
              :variant="getTierBadge(m).variant"
              class="font-mono text-[9px] px-1.5 py-0 h-4"
            >
              {{ getTierBadge(m).label }}
            </Badge>
            <Check
              v-if="m === modelValue"
              class="h-3.5 w-3.5 text-blue-400 shrink-0"
            />
            <span v-else class="w-3.5 shrink-0" />
          </div>
        </button>

        <!-- No Filter Results State -->
        <div v-if="filteredModels.length === 0" class="py-6 px-4 text-center space-y-1">
          <Sparkles class="h-4 w-4 text-zinc-500 mx-auto" />
          <p class="text-xs text-zinc-400 font-medium">No matching models found</p>
          <p class="text-[11px] text-zinc-500">Try refining your search keyword.</p>
        </div>
      </div>

      <!-- Custom Model Fallback option -->
      <div
        v-if="searchQuery.trim() && !exactMatchExists"
        class="p-2 border-t border-zinc-800/80 bg-zinc-900/40"
      >
        <button
          type="button"
          @click="selectModel(searchQuery.trim())"
          class="w-full flex items-center justify-between p-2 rounded-lg bg-blue-500/10 hover:bg-blue-500/20 border border-blue-500/30 text-blue-400 text-xs transition-colors cursor-pointer"
        >
          <span class="flex items-center gap-1.5 truncate">
            <PlusCircle class="h-3.5 w-3.5 shrink-0" />
            <span>Use custom model: <strong class="font-mono text-white">{{ searchQuery.trim() }}</strong></span>
          </span>
          <span class="text-[10px] text-blue-300 font-mono shrink-0">Enter ↵</span>
        </button>
      </div>

      <!-- Popover Footer Status Bar -->
      <div
        v-else
        class="px-3 py-1.5 border-t border-zinc-800/80 bg-zinc-900/40 flex items-center justify-between text-[10px] font-mono text-zinc-500"
      >
        <span>{{ filteredModels.length }} models</span>
        <span class="hidden sm:inline">Press Esc to close</span>
      </div>
    </div>
  </div>
</template>
