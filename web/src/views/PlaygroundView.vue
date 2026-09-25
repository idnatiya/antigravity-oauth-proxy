<script setup lang="ts">
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { marked } from 'marked'
import {
  Play,
  Square,
  Trash2,
  Copy,
  Check,
  Clock,
  Terminal,
  Zap,
  CheckCircle2,
  AlertCircle,
  Sparkles,
  Code,
  RefreshCw,
  ChevronDown,
  ChevronUp,
  FileText,
} from '@lucide/vue'
import { useUsageStore } from '@/stores/usageStore'
import ModelSelector from '@/components/playground/ModelSelector.vue'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Switch } from '@/components/ui/switch'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'

// Configure marked
marked.setOptions({
  gfm: true,
  breaks: true,
})

const route = useRoute()
const usageStore = useUsageStore()

// State
const selectedModel = ref('gemini-2.5-flash')
const streamMode = ref(true)
const viewMode = ref<'markdown' | 'raw'>('markdown')
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
    label: 'Math Calc',
    icon: Zap,
    prompt: 'Compute 12345 * 67890. Show the exact calculation and the final number.',
  },
  {
    label: 'Code Generator',
    icon: Code,
    prompt: 'Write a TypeScript function to deep clone an object safely, handling Dates and nested arrays.',
  },
  {
    label: 'Stream Speed',
    icon: Terminal,
    prompt: 'Write a 3-paragraph inspiring story about human exploration of the Alpha Centauri system.',
  },
]

function selectPreset(prompt: string) {
  promptText.value = prompt
}

const wordCount = computed(() => {
  return promptText.value.split(/\s+/).filter(Boolean).length
})

const estimatedTokens = computed(() => {
  if (!responseText.value) return 0
  return Math.round(responseText.value.length / 4)
})

const renderedMarkdown = computed(() => {
  if (!responseText.value) return ''
  try {
    return marked.parse(responseText.value, { async: false }) as string
  } catch {
    return responseText.value
  }
})

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

function clearPrompt() {
  promptText.value = ''
}

function clearOutput() {
  responseText.value = ''
  errorMessage.value = ''
  responseStatus.value = null
  latencyMs.value = null
}

function clearAll() {
  clearPrompt()
  clearOutput()
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

function applyRouteModel() {
  if (route.query.model && typeof route.query.model === 'string') {
    selectedModel.value = route.query.model
  }
}

watch(() => route.query.model, () => {
  applyRouteModel()
})

onMounted(() => {
  applyRouteModel()
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
          <Badge variant="outline" class="font-mono text-blue-400 bg-blue-500/10 border-blue-500/20 text-[11px]">
            OpenAI Compatible
          </Badge>
        </div>
        <p class="text-xs text-zinc-400 mt-0.5">
          Verify end-to-end proxy completions, test real-time SSE streaming, and benchmark model responses.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          @click="showCurlModal = true"
          class="h-8 text-xs bg-zinc-900/60 border-zinc-800 hover:bg-zinc-800 text-zinc-300 gap-1.5"
        >
          <Terminal class="h-3.5 w-3.5 text-zinc-400" />
          <span>cURL Command</span>
        </Button>
        <Button
          v-if="promptText || responseText || errorMessage"
          variant="ghost"
          size="sm"
          @click="clearAll"
          :disabled="isRunning"
          class="h-8 text-xs text-zinc-400 hover:text-zinc-200 hover:bg-zinc-800/60 gap-1.5"
        >
          <Trash2 class="h-3.5 w-3.5 text-zinc-500" />
          <span>Reset All</span>
        </Button>
      </div>
    </div>

    <!-- Workbench Grid (Controls & Input on Left, Output on Right) -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-5 items-stretch">
      <!-- Left Column: Settings & Input Studio (5 cols) -->
      <div class="lg:col-span-5 flex flex-col space-y-4">
        <!-- Configuration & Prompt Card -->
        <Card class="p-4 sm:p-5 bg-[#121316] border-zinc-800/80 shadow-sm flex-1 flex flex-col space-y-4">
          <!-- Searchable Model Selection & Stream Row -->
          <div class="space-y-3">
            <div class="space-y-1.5">
              <label class="block text-xs font-medium text-zinc-400">Select Model</label>
              <ModelSelector
                v-model="selectedModel"
                :models="availableModels"
                :disabled="isRunning"
              />
            </div>

            <!-- Streaming Switch Row -->
            <div class="flex items-center justify-between p-3 rounded-lg bg-[#0d0e11] border border-zinc-800/80">
              <div class="space-y-0.5">
                <div class="text-xs font-medium text-zinc-200 flex items-center gap-2">
                  <span>Stream Output (SSE)</span>
                  <Badge
                    variant="outline"
                    class="font-mono text-[9px] px-1.5 py-0 h-4"
                    :class="streamMode ? 'text-blue-400 border-blue-500/30 bg-blue-500/10' : 'text-zinc-500 border-zinc-800'"
                  >
                    {{ streamMode ? 'Real-time SSE' : 'Buffered JSON' }}
                  </Badge>
                </div>
                <div class="text-[11px] text-zinc-500">Stream tokens as they are generated by upstream</div>
              </div>
              <Switch
                :checked="streamMode"
                @update:checked="streamMode = $event"
                :disabled="isRunning"
              />
            </div>
          </div>

          <!-- Quick Presets -->
          <div class="space-y-1.5">
            <span class="text-[11px] font-medium text-zinc-500 uppercase tracking-wider">Quick Presets</span>
            <div class="flex items-center gap-2 overflow-x-auto pb-1 no-scrollbar">
              <button
                v-for="p in presets"
                :key="p.label"
                type="button"
                :disabled="isRunning"
                @click="selectPreset(p.prompt)"
                class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-[#0d0e11] border border-zinc-800/80 hover:border-blue-500/40 text-xs text-zinc-300 hover:text-white transition-all shrink-0 cursor-pointer group disabled:opacity-50"
              >
                <component :is="p.icon" class="h-3.5 w-3.5 text-blue-400 group-hover:scale-110 transition-transform" />
                <span>{{ p.label }}</span>
              </button>
            </div>
          </div>

          <!-- Collapsible System Prompt Drawer -->
          <div class="rounded-lg bg-[#0d0e11] border border-zinc-800/80 overflow-hidden">
            <button
              type="button"
              @click="showSystemPrompt = !showSystemPrompt"
              class="w-full px-3 py-2 text-xs text-zinc-400 hover:text-zinc-200 flex items-center justify-between cursor-pointer transition-colors bg-zinc-900/30"
            >
              <span class="flex items-center gap-1.5">
                <Sparkles class="h-3 w-3 text-purple-400" />
                <span>System Instructions</span>
                <span v-if="systemPrompt.trim()" class="text-[10px] text-purple-400 font-mono">(active)</span>
                <span v-else class="text-[10px] text-zinc-600">(optional)</span>
              </span>
              <component :is="showSystemPrompt ? ChevronUp : ChevronDown" class="h-3.5 w-3.5 text-zinc-500" />
            </button>
            <div v-if="showSystemPrompt" class="p-3 border-t border-zinc-800/80">
              <textarea
                v-model="systemPrompt"
                :disabled="isRunning"
                rows="2"
                placeholder="Optional instructions guiding tone, formatting, or constraints..."
                class="w-full bg-[#121316] border border-zinc-800/80 focus:border-blue-500/80 rounded-lg p-2.5 text-zinc-200 text-xs placeholder:text-zinc-600 focus:outline-none transition-colors resize-y font-sans leading-relaxed"
              />
            </div>
          </div>

          <!-- User Prompt Textarea Area -->
          <div class="space-y-2 flex-1 flex flex-col pt-1">
            <div class="flex items-center justify-between">
              <label class="text-xs font-medium text-zinc-300">User Prompt</label>
              <Button
                v-if="promptText"
                variant="ghost"
                size="sm"
                @click="clearPrompt"
                :disabled="isRunning"
                class="h-6 px-2 text-[11px] text-zinc-500 hover:text-zinc-300 gap-1"
              >
                <Trash2 class="h-3 w-3" />
                <span>Clear Input</span>
              </Button>
            </div>

            <textarea
              v-model="promptText"
              :disabled="isRunning"
              @keydown="onKeydown"
              rows="6"
              placeholder="Type your prompt here... (Press Cmd+Enter or Ctrl+Enter to run)"
              class="w-full flex-1 min-h-[160px] bg-[#0d0e11] border border-zinc-800/80 focus:border-blue-500/80 rounded-xl p-3.5 text-xs text-zinc-100 placeholder:text-zinc-600 focus:outline-none transition-all resize-y leading-relaxed font-sans"
            />

            <!-- Prompt Card Bottom Toolbar -->
            <div class="flex items-center justify-between pt-1">
              <div class="flex items-center gap-2 text-[11px] text-zinc-500 font-mono">
                <span>{{ promptText.length }} chars</span>
                <span>•</span>
                <span>{{ wordCount }} words</span>
                <span class="hidden sm:inline-flex items-center px-1.5 py-0.5 text-[10px] bg-zinc-900 border border-zinc-800 rounded text-zinc-400">
                  ⌘↵
                </span>
              </div>

              <div class="flex items-center gap-2">
                <Button
                  v-if="isRunning"
                  variant="destructive"
                  size="sm"
                  @click="stopTest"
                  class="gap-1.5 h-8 text-xs font-medium"
                >
                  <Square class="h-3 w-3 fill-current" />
                  <span>Stop</span>
                </Button>

                <Button
                  v-else
                  variant="default"
                  size="sm"
                  :disabled="!promptText.trim()"
                  @click="runTest"
                  class="gap-1.5 h-8 text-xs font-medium bg-blue-600 hover:bg-blue-500 text-white shadow-sm shadow-blue-500/20 disabled:opacity-50"
                >
                  <Play class="h-3 w-3 fill-current" />
                  <span>Run Prompt</span>
                </Button>
              </div>
            </div>
          </div>
        </Card>
      </div>

      <!-- Right Column: Live Output & Telemetry Console (7 cols) -->
      <div class="lg:col-span-7 flex flex-col">
        <Card class="bg-[#121316] border-zinc-800/80 overflow-hidden flex flex-col flex-1 min-h-[520px] shadow-sm">
          <!-- Console Chrome Header -->
          <div class="px-5 py-3 border-b border-zinc-800/80 flex items-center justify-between bg-zinc-900/40">
            <div class="flex items-center gap-2.5">
              <span class="text-xs font-semibold text-white">Execution Output</span>
              <!-- Status Indicator -->
              <Badge
                v-if="isRunning"
                variant="outline"
                class="font-mono text-[10px] bg-blue-500/10 text-blue-400 border-blue-500/20 gap-1.5"
              >
                <RefreshCw class="h-2.5 w-2.5 animate-spin" />
                Streaming...
              </Badge>
              <Badge
                v-else-if="responseStatus"
                variant="outline"
                :class="responseStatus < 400 ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' : 'bg-red-500/10 text-red-400 border-red-500/20'"
                class="font-mono text-[10px] gap-1"
              >
                <CheckCircle2 v-if="responseStatus < 400" class="h-3 w-3" />
                <AlertCircle v-else class="h-3 w-3" />
                {{ responseStatus }} {{ responseStatus < 400 ? 'OK' : 'Error' }}
              </Badge>
              <Badge
                v-else
                variant="outline"
                class="font-mono text-[10px] text-zinc-500 border-zinc-800"
              >
                Idle
              </Badge>
            </div>

            <!-- Header Actions, Mode Toggle & Latency -->
            <div class="flex items-center gap-2">
              <!-- View Mode Toggle (Markdown vs Raw) -->
              <div
                v-if="responseText"
                class="flex items-center rounded-lg bg-zinc-900 border border-zinc-800 p-0.5 text-[10px] font-medium"
              >
                <button
                  type="button"
                  @click="viewMode = 'markdown'"
                  class="px-2 py-0.5 rounded cursor-pointer transition-colors"
                  :class="viewMode === 'markdown' ? 'bg-zinc-800 text-white font-semibold shadow-xs' : 'text-zinc-400 hover:text-zinc-200'"
                >
                  Markdown
                </button>
                <button
                  type="button"
                  @click="viewMode = 'raw'"
                  class="px-2 py-0.5 rounded cursor-pointer transition-colors"
                  :class="viewMode === 'raw' ? 'bg-zinc-800 text-white font-semibold shadow-xs' : 'text-zinc-400 hover:text-zinc-200'"
                >
                  Raw
                </button>
              </div>

              <Badge
                v-if="latencyMs !== null"
                variant="outline"
                class="font-mono text-[10px] text-zinc-400 border-zinc-800 bg-zinc-900/60 gap-1"
              >
                <Clock class="h-3 w-3 text-zinc-500" />
                <span>{{ latencyMs }}ms</span>
              </Badge>

              <Button
                v-if="responseText"
                variant="secondary"
                size="sm"
                @click="copyOutput"
                class="h-7 text-[11px] gap-1 bg-zinc-800 hover:bg-zinc-700 text-zinc-200"
                title="Copy response text"
              >
                <Check v-if="copiedOutput" class="h-3 w-3 text-emerald-400" />
                <Copy v-else class="h-3 w-3" />
                <span>{{ copiedOutput ? 'Copied' : 'Copy' }}</span>
              </Button>

              <Button
                v-if="responseText || errorMessage"
                variant="ghost"
                size="sm"
                @click="clearOutput"
                class="h-7 px-2 text-[11px] text-zinc-500 hover:text-zinc-300"
                title="Clear output console"
              >
                <Trash2 class="h-3 w-3" />
              </Button>
            </div>
          </div>

          <!-- Error Alert Banner -->
          <div
            v-if="errorMessage"
            class="p-4 bg-red-500/10 border-b border-red-500/20 text-red-300 text-xs flex items-start justify-between gap-3 font-mono"
          >
            <div class="flex items-start gap-2.5 min-w-0">
              <AlertCircle class="h-4 w-4 text-red-400 shrink-0 mt-0.5" />
              <div class="space-y-1 min-w-0">
                <div class="font-semibold text-red-300 font-sans">Execution Failure</div>
                <p class="text-[11px] text-red-200/90 break-all leading-relaxed">{{ errorMessage }}</p>
              </div>
            </div>
            <Button
              variant="outline"
              size="sm"
              @click="runTest"
              class="h-7 text-xs border-red-500/30 hover:bg-red-500/20 text-red-200 shrink-0"
            >
              Retry
            </Button>
          </div>

          <!-- Output Body (Terminal Window) -->
          <div class="p-5 flex-1 flex flex-col overflow-y-auto bg-[#090a0c]">
            <!-- Empty state when no test has run yet -->
            <div
              v-if="!responseText && !isRunning && !errorMessage"
              class="flex-1 flex flex-col items-center justify-center text-center p-8 space-y-3 my-auto"
            >
              <div class="h-12 w-12 rounded-xl bg-zinc-900 border border-zinc-800 text-zinc-400 flex items-center justify-center">
                <Terminal class="h-6 w-6 text-zinc-500" />
              </div>
              <div class="space-y-1">
                <h4 class="text-xs font-semibold text-zinc-300">Awaiting Execution</h4>
                <p class="text-[11px] text-zinc-500 max-w-sm leading-relaxed">
                  Select a model, pick a quick preset or type your custom prompt, then press
                  <kbd class="px-1 py-0.5 text-[10px] bg-zinc-800 border border-zinc-700 rounded text-zinc-400 font-mono">⌘↵</kbd>
                  or click "Run Prompt" to view real-time streaming output.
                </p>
              </div>
            </div>

            <!-- Rendered Markdown Output -->
            <div v-else-if="viewMode === 'markdown'" class="relative flex-1 select-text">
              <div class="markdown-body" v-html="renderedMarkdown" />
              <span
                v-if="isRunning"
                class="inline-block w-2 h-4 ml-1 bg-blue-400 align-middle animate-pulse"
              />
            </div>

            <!-- Raw Monospace Output -->
            <div v-else class="font-mono text-xs text-zinc-200 whitespace-pre-wrap leading-relaxed select-text flex-1">
              {{ responseText }}
              <span
                v-if="isRunning"
                class="inline-block w-2 h-4 ml-0.5 bg-blue-400 align-middle animate-pulse"
              />
            </div>
          </div>

          <!-- Bottom Telemetry Bar -->
          <div
            class="px-5 py-2.5 border-t border-zinc-800/80 bg-zinc-900/40 flex flex-wrap items-center justify-between text-[11px] font-mono text-zinc-500 gap-2"
          >
            <div class="flex items-center gap-3">
              <span>Model: <strong class="text-zinc-300 font-semibold">{{ selectedModel }}</strong></span>
              <span>•</span>
              <span>Mode: <strong class="text-zinc-300 font-semibold">{{ streamMode ? 'SSE Stream' : 'Buffered JSON' }}</strong></span>
              <span>•</span>
              <span class="flex items-center gap-1">
                <FileText class="h-3 w-3 text-zinc-500" />
                <span>Format: <strong class="text-zinc-300 capitalize">{{ viewMode }}</strong></span>
              </span>
            </div>
            <div class="flex items-center gap-3">
              <span v-if="responseText">
                {{ responseText.length }} chars <span class="text-zinc-600">(~{{ estimatedTokens }} tokens)</span>
              </span>
              <span v-else>Ready</span>
            </div>
          </div>
        </Card>
      </div>
    </div>

    <!-- cURL Command Dialog Modal -->
    <Dialog :open="showCurlModal" @update:open="(val: boolean) => showCurlModal = val">
      <DialogContent class="max-w-2xl bg-[#121316] border-zinc-800 p-0 overflow-hidden">
        <DialogHeader class="px-6 py-4 border-b border-zinc-800/80 bg-zinc-900/40">
          <div class="flex items-center gap-2.5">
            <div class="p-2 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400">
              <Terminal class="h-4 w-4" />
            </div>
            <div>
              <DialogTitle class="text-sm font-semibold text-white">cURL Command Snippet</DialogTitle>
              <DialogDescription class="text-xs text-zinc-400 mt-0.5">
                Execute requests against this proxy instance from terminal or CLI scripts
              </DialogDescription>
            </div>
          </div>
        </DialogHeader>

        <div class="p-6 space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-[11px] font-mono text-zinc-400 uppercase tracking-wider">Terminal Command</span>
            <Button
              variant="secondary"
              size="sm"
              @click="copyCurl"
              class="h-7 text-xs gap-1.5 bg-zinc-800 hover:bg-zinc-700 text-zinc-200"
            >
              <Check v-if="copiedCurl" class="h-3 w-3 text-emerald-400" />
              <Copy v-else class="h-3 w-3" />
              <span>{{ copiedCurl ? 'Copied to Clipboard' : 'Copy cURL' }}</span>
            </Button>
          </div>

          <pre class="p-4 rounded-xl bg-[#090a0c] border border-zinc-800/80 text-zinc-300 font-mono text-xs overflow-x-auto selection:bg-blue-600/40 leading-relaxed"><code>{{ curlCommand }}</code></pre>

          <p class="text-[11px] text-zinc-500">
            Replace <code class="text-blue-400 font-mono">&lt;ADMIN_API_KEY&gt;</code> with your proxy admin key or API key configured in environment variables.
          </p>
        </div>

        <DialogFooter class="px-6 py-3 border-t border-zinc-800/80 bg-zinc-900/30 flex justify-end">
          <Button variant="outline" size="sm" @click="showCurlModal = false">
            Close
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </div>
</template>
