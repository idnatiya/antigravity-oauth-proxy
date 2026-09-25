<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import {
  Play,
  Square,
  Trash2,
  Copy,
  Check,
  Clock,
  Terminal,
  Cpu,
  Zap,
  CheckCircle2,
  AlertCircle,
  Sparkles,
  Code,
  RefreshCw,
} from '@lucide/vue'
import { useUsageStore } from '@/stores/usageStore'

const usageStore = useUsageStore()

// State
const selectedModel = ref('gemini-2.5-flash')
const streamMode = ref(true)
const promptText = ref('')
const systemPrompt = ref('')
const showSystemPrompt = ref(false)
const showCurlModal = ref(false)

const isRunning = ref(false)
const responseText = ref('')
const latencyMs = ref<number | null>(null)
const responseStatus = ref<number | null>(null)
const errorMessage = ref('')
const copiedOutput = ref(false)
const copiedCurl = ref(false)

let abortController: AbortController | null = null

// Fallback models if store models not loaded yet
const defaultModelOptions = [
  'gemini-2.5-flash',
  'gemini-2.5-pro',
  'gemini-3.7-flash',
  'claude-3-7-sonnet',
  'gemini-2.0-flash',
]

const availableModels = computed(() => {
  if (usageStore.models.length > 0) {
    return usageStore.models.map(m => m.id)
  }
  return defaultModelOptions
})

const presets = [
  {
    label: 'Ping Test',
    icon: Sparkles,
    prompt: 'Reply with "Antigravity Proxy is fully operational!" and state your exact model name.',
  },
  {
    label: 'Math Calculation',
    icon: Zap,
    prompt: 'Compute 12345 * 67890. Show the exact calculation and the final number.',
  },
  {
    label: 'Code Generator',
    icon: Code,
    prompt: 'Write a TypeScript function to deep clone an object safely, handling Dates and nested arrays.',
  },
  {
    label: 'Streaming Speed',
    icon: Terminal,
    prompt: 'Write a 3-paragraph inspiring story about human exploration of the Alpha Centauri system.',
  },
]

function selectPreset(prompt: string) {
  promptText.value = prompt
}

async function runTest() {
  if (!promptText.value.trim() || isRunning.value) return

  isRunning.value = true
  errorMessage.value = ''
  responseText.value = ''
  responseStatus.value = null
  latencyMs.value = null

  abortController = new AbortController()
  const startTime = performance.now()

  try {
    const payload = {
      model: selectedModel.value,
      stream: streamMode.value,
      messages: [
        ...(systemPrompt.value.trim() ? [{ role: 'system', content: systemPrompt.value.trim() }] : []),
        { role: 'user', content: promptText.value.trim() },
      ],
    }

    const res = await fetch('/api/playground/chat', {
      method: 'POST',
      credentials: 'include',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(payload),
      signal: abortController.signal,
    })

    responseStatus.value = res.status

    if (!res.ok) {
      const errText = await res.text()
      throw new Error(`HTTP ${res.status}: ${errText || res.statusText}`)
    }

    if (!streamMode.value) {
      const data = await res.json()
      latencyMs.value = Math.round(performance.now() - startTime)
      responseText.value = data.choices?.[0]?.message?.content || JSON.stringify(data, null, 2)
      return
    }

    // Stream SSE Response
    const reader = res.body?.getReader()
    if (!reader) {
      throw new Error('Streaming not supported by browser environment')
    }

    const decoder = new TextDecoder()
    let buffer = ''

    while (true) {
      const { done, value } = await reader.read()
      if (done) break

      buffer += decoder.decode(value, { stream: true })
      const lines = buffer.split('\n')
      buffer = lines.pop() || ''

      for (const line of lines) {
        const trimmed = line.trim()
        if (!trimmed || !trimmed.startsWith('data:')) continue
        const dataStr = trimmed.slice(5).trim()
        if (dataStr === '[DONE]') continue
        try {
          const parsed = JSON.parse(dataStr)
          const delta = parsed.choices?.[0]?.delta?.content
          if (delta) {
            responseText.value += delta
            if (latencyMs.value === null) {
              latencyMs.value = Math.round(performance.now() - startTime)
            }
          }
        } catch {
          // ignore incomplete json chunk in SSE
        }
      }
    }

    latencyMs.value = Math.round(performance.now() - startTime)
  } catch (err: unknown) {
    if (err instanceof DOMException && err.name === 'AbortError') {
      responseText.value += '\n\n[Request interrupted by user]'
    } else {
      errorMessage.value = err instanceof Error ? err.message : 'Unknown request failure'
    }
  } finally {
    isRunning.value = false
    abortController = null
  }
}

function stopTest() {
  if (abortController) {
    abortController.abort()
  }
}

function clearAll() {
  promptText.value = ''
  responseText.value = ''
  errorMessage.value = ''
  responseStatus.value = null
  latencyMs.value = null
}

function copyOutput() {
  if (!responseText.value) return
  navigator.clipboard.writeText(responseText.value)
  copiedOutput.value = true
  setTimeout(() => (copiedOutput.value = false), 2000)
}

const curlCommand = computed(() => {
  const origin = window.location.origin
  const escapedPrompt = promptText.value.replace(/"/g, '\\"').replace(/\n/g, ' ') || 'Hello!'
  return `curl -X POST "${origin}/v1/chat/completions" \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer <ADMIN_API_KEY>" \\
  -d '{
    "model": "${selectedModel.value}",
    "stream": ${streamMode.value},
    "messages": [
      {"role": "user", "content": "${escapedPrompt}"}
    ]
  }'`
})

function copyCurl() {
  navigator.clipboard.writeText(curlCommand.value)
  copiedCurl.value = true
  setTimeout(() => (copiedCurl.value = false), 2000)
}

function onKeydown(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
    e.preventDefault()
    runTest()
  }
}

onMounted(() => {
  if (usageStore.models.length === 0) {
    usageStore.fetchModels().catch(() => {})
  }
})
</script>

<template>
  <div class="space-y-6">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <div class="flex items-center gap-2.5">
          <h2 class="text-base font-semibold text-white tracking-tight">API Playground & Functional Tester</h2>
          <span class="px-2 py-0.5 rounded-full text-[11px] font-mono font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20">
            OpenAI Compatible
          </span>
        </div>
        <p class="text-xs text-zinc-500 mt-0.5">
          Verify end-to-end proxy completions, test real-time SSE streaming, and benchmark models.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <button
          type="button"
          @click="showCurlModal = !showCurlModal"
          class="flex items-center gap-1.5 px-3 py-2 rounded-lg border border-[#2c2e36] bg-[#202227] text-zinc-300 hover:text-white hover:bg-[#282a32] text-xs font-medium transition-colors cursor-pointer shadow-xs"
        >
          <Terminal class="h-3.5 w-3.5 text-zinc-400" />
          <span>{{ showCurlModal ? 'Hide cURL' : 'cURL Command' }}</span>
        </button>
      </div>
    </div>

    <!-- cURL preview panel (expandable) -->
    <div
      v-if="showCurlModal"
      class="p-4 rounded-xl bg-[#18191d] border border-[#2c2e36] space-y-2 text-xs font-mono animate-in fade-in duration-150"
    >
      <div class="flex items-center justify-between text-zinc-400">
        <span class="text-[11px] uppercase tracking-wider font-sans font-medium text-zinc-500">Terminal Command</span>
        <button
          @click="copyCurl"
          class="flex items-center gap-1 px-2.5 py-1 rounded bg-[#202227] border border-[#2c2e36] text-zinc-300 hover:text-white transition-colors cursor-pointer text-[11px] font-sans"
        >
          <Check v-if="copiedCurl" class="h-3.5 w-3.5 text-emerald-400" />
          <Copy v-else class="h-3.5 w-3.5" />
          <span>{{ copiedCurl ? 'Copied' : 'Copy cURL' }}</span>
        </button>
      </div>
      <pre class="p-3 rounded-lg bg-[#101114] border border-[#24262e] text-zinc-300 text-[11px] overflow-x-auto selection:bg-blue-600/40"><code>{{ curlCommand }}</code></pre>
    </div>

    <!-- Workbench Grid (Controls & Input on Left, Output on Right) -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-5 items-start">
      <!-- Left Column: Settings & Input (5 cols) -->
      <div class="lg:col-span-5 space-y-4">
        <!-- Configuration Card -->
        <div class="p-4 rounded-xl bg-[#202227] border border-[#2c2e36] space-y-3.5 text-xs">
          <!-- Model Selection -->
          <div class="space-y-1.5">
            <label class="block font-medium text-zinc-400">Select Model</label>
            <div class="relative">
              <select
                v-model="selectedModel"
                :disabled="isRunning"
                class="w-full bg-[#18191d] border border-[#2c2e36] focus:border-blue-500 rounded-lg px-3 py-2 text-zinc-200 font-mono text-xs focus:outline-none transition-colors appearance-none cursor-pointer pr-8"
              >
                <option v-for="m in availableModels" :key="m" :value="m">
                  {{ m }}
                </option>
              </select>
              <Cpu class="h-4 w-4 absolute right-3 top-1/2 -translate-y-1/2 text-zinc-500 pointer-events-none" />
            </div>
          </div>

          <!-- Streaming Switch -->
          <div class="flex items-center justify-between pt-1">
            <div>
              <div class="font-medium text-white">Stream Output (SSE)</div>
              <div class="text-[11px] text-zinc-500">Stream response tokens incrementally</div>
            </div>
            <label class="relative inline-flex items-center cursor-pointer">
              <input
                type="checkbox"
                v-model="streamMode"
                :disabled="isRunning"
                class="sr-only peer"
              />
              <div
                class="w-9 h-5 bg-zinc-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-zinc-300 after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-blue-600"
              />
            </label>
          </div>

          <!-- Collapsible System Prompt -->
          <div class="pt-2 border-t border-[#2a2d34]">
            <button
              type="button"
              @click="showSystemPrompt = !showSystemPrompt"
              class="text-[11px] text-zinc-400 hover:text-zinc-200 flex items-center justify-between w-full cursor-pointer"
            >
              <span>System Prompt (Optional)</span>
              <span>{{ showSystemPrompt ? '▲' : '▼' }}</span>
            </button>
            <div v-if="showSystemPrompt" class="mt-2">
              <textarea
                v-model="systemPrompt"
                :disabled="isRunning"
                rows="2"
                placeholder="Optional instructions for the model..."
                class="w-full bg-[#18191d] border border-[#2c2e36] focus:border-blue-500 rounded-lg p-2.5 text-zinc-200 text-xs placeholder:text-zinc-600 focus:outline-none transition-colors"
              />
            </div>
          </div>
        </div>

        <!-- Quick Presets -->
        <div class="space-y-1.5">
          <span class="text-[11px] font-medium text-zinc-500 uppercase tracking-wider">Quick Presets</span>
          <div class="grid grid-cols-2 gap-2">
            <button
              v-for="p in presets"
              :key="p.label"
              type="button"
              :disabled="isRunning"
              @click="selectPreset(p.prompt)"
              class="flex items-center gap-2 p-2.5 rounded-lg bg-[#202227] border border-[#2c2e36] hover:border-blue-500/40 text-left text-xs text-zinc-300 hover:text-white transition-all cursor-pointer group"
            >
              <component :is="p.icon" class="h-3.5 w-3.5 text-blue-400 shrink-0 group-hover:scale-110 transition-transform" />
              <span class="truncate">{{ p.label }}</span>
            </button>
          </div>
        </div>

        <!-- Prompt Textarea Card -->
        <div class="p-4 rounded-xl bg-[#202227] border border-[#2c2e36] space-y-3">
          <div class="flex items-center justify-between">
            <label class="text-xs font-medium text-zinc-400">User Prompt</label>
            <button
              v-if="promptText"
              @click="clearAll"
              :disabled="isRunning"
              class="text-[11px] text-zinc-500 hover:text-zinc-300 flex items-center gap-1 cursor-pointer"
            >
              <Trash2 class="h-3 w-3" />
              <span>Clear</span>
            </button>
          </div>

          <textarea
            v-model="promptText"
            :disabled="isRunning"
            @keydown="onKeydown"
            rows="6"
            placeholder="Type your prompt here... (Press Cmd+Enter or Ctrl+Enter to run)"
            class="w-full bg-[#18191d] border border-[#2c2e36] focus:border-blue-500 rounded-lg p-3 text-xs text-zinc-200 placeholder:text-zinc-600 focus:outline-none transition-colors resize-y leading-relaxed font-sans"
          />

          <div class="flex items-center justify-between pt-1">
            <span class="text-[10px] text-zinc-500 font-mono">
              {{ promptText.length }} chars • {{ promptText.split(/\s+/).filter(Boolean).length }} words
            </span>

            <div class="flex items-center gap-2">
              <button
                v-if="isRunning"
                type="button"
                @click="stopTest"
                class="flex items-center gap-1.5 px-4 py-2 rounded-lg bg-red-600 hover:bg-red-500 text-white text-xs font-medium transition-colors cursor-pointer"
              >
                <Square class="h-3.5 w-3.5 fill-current" />
                <span>Stop</span>
              </button>

              <button
                v-else
                type="button"
                :disabled="!promptText.trim()"
                @click="runTest"
                class="flex items-center gap-1.5 px-5 py-2 rounded-lg bg-blue-600 hover:bg-blue-500 disabled:opacity-50 text-white text-xs font-medium transition-all shadow-sm shadow-blue-500/10 cursor-pointer"
              >
                <Play class="h-3.5 w-3.5 fill-current" />
                <span>Run Prompt</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Right Column: Live Output & Telemetry (7 cols) -->
      <div class="lg:col-span-7 space-y-4">
        <!-- Output Card -->
        <div class="rounded-xl bg-[#202227] border border-[#2c2e36] overflow-hidden flex flex-col min-h-[460px]">
          <!-- Card Header & Status Bar -->
          <div class="px-5 py-3 border-b border-[#282a32] flex items-center justify-between bg-[#1b1d22]">
            <div class="flex items-center gap-2.5">
              <span class="text-xs font-semibold text-white">Execution Output</span>
              <!-- Status indicator -->
              <span
                v-if="isRunning"
                class="inline-flex items-center gap-1.5 px-2 py-0.5 rounded-full text-[10px] font-mono font-medium bg-blue-500/10 text-blue-400 border border-blue-500/20"
              >
                <RefreshCw class="h-2.5 w-2.5 animate-spin" />
                Streaming...
              </span>
              <span
                v-else-if="responseStatus"
                class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-mono font-medium"
                :class="responseStatus < 400 ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20' : 'bg-red-500/10 text-red-400 border border-red-500/20'"
              >
                <CheckCircle2 v-if="responseStatus < 400" class="h-3 w-3" />
                <AlertCircle v-else class="h-3 w-3" />
                {{ responseStatus }} {{ responseStatus < 400 ? 'OK' : 'Error' }}
              </span>
            </div>

            <!-- Header Actions -->
            <div class="flex items-center gap-2">
              <span v-if="latencyMs !== null" class="text-[11px] font-mono text-zinc-400 flex items-center gap-1">
                <Clock class="h-3 w-3 text-zinc-500" />
                {{ latencyMs }}ms
              </span>
              <button
                v-if="responseText"
                @click="copyOutput"
                class="flex items-center gap-1 px-2.5 py-1 rounded bg-[#252830] hover:bg-[#2e313b] text-zinc-300 hover:text-white transition-colors text-[11px] cursor-pointer"
                title="Copy response text"
              >
                <Check v-if="copiedOutput" class="h-3 w-3 text-emerald-400" />
                <Copy v-else class="h-3 w-3" />
                <span>{{ copiedOutput ? 'Copied' : 'Copy' }}</span>
              </button>
            </div>
          </div>

          <!-- Error Alert Banner -->
          <div
            v-if="errorMessage"
            class="p-4 bg-red-500/10 border-b border-red-500/20 text-red-300 text-xs flex items-start gap-2.5 font-mono"
          >
            <AlertCircle class="h-4 w-4 text-red-400 shrink-0 mt-0.5" />
            <div class="space-y-1 flex-1 min-w-0">
              <div class="font-semibold text-red-300 font-sans">Execution Failure</div>
              <p class="text-[11px] text-red-200/90 break-all">{{ errorMessage }}</p>
            </div>
          </div>

          <!-- Output Body -->
          <div class="p-5 flex-1 flex flex-col overflow-y-auto max-h-[560px]">
            <!-- Empty state when no test has run yet -->
            <div
              v-if="!responseText && !isRunning && !errorMessage"
              class="flex-1 flex flex-col items-center justify-center text-center p-8 space-y-3 my-auto"
            >
              <div class="h-12 w-12 rounded-xl bg-zinc-800/80 border border-zinc-700/60 text-zinc-400 flex items-center justify-center">
                <Terminal class="h-6 w-6 text-zinc-500" />
              </div>
              <div class="space-y-1">
                <h4 class="text-xs font-semibold text-zinc-300">Ready to Test</h4>
                <p class="text-[11px] text-zinc-500 max-w-xs leading-relaxed">
                  Select a model, pick a preset or write a custom prompt, then click "Run Prompt" to view live response output.
                </p>
              </div>
            </div>

            <!-- Live Text Output -->
            <div v-else class="font-mono text-xs text-zinc-200 whitespace-pre-wrap leading-relaxed select-text">
              {{ responseText }}
              <span
                v-if="isRunning"
                class="inline-block w-2 h-4 ml-0.5 bg-blue-400 align-middle animate-pulse"
              />
            </div>
          </div>

          <!-- Bottom Telemetry Bar -->
          <div
            v-if="responseText || latencyMs !== null"
            class="px-5 py-2.5 border-t border-[#282a32] bg-[#1b1d22] flex flex-wrap items-center justify-between text-[11px] font-mono text-zinc-400 gap-2"
          >
            <div class="flex items-center gap-4">
              <span>Model: <strong class="text-zinc-200">{{ selectedModel }}</strong></span>
              <span>Mode: <strong class="text-zinc-200">{{ streamMode ? 'SSE Stream' : 'JSON' }}</strong></span>
            </div>
            <div class="flex items-center gap-4">
              <span>Output length: <strong class="text-zinc-200">{{ responseText.length }} chars</strong></span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
