<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  KeyRound,
  Plus,
  Copy,
  Check,
  Eye,
  EyeOff,
  Trash2,
  ShieldCheck,
  Code2,
  Terminal,
  Cpu,
  RefreshCw,
  Sparkles,
} from '@lucide/vue'
import { Card } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import CreateKeyModal from '@/components/apikeys/CreateKeyModal.vue'
import DeleteKeyModal from '@/components/apikeys/DeleteKeyModal.vue'
import { apiKeysService } from '@/services/apiKeysService'
import type { APIKey } from '@/types'

const keys = ref<APIKey[]>([])
const loading = ref(true)
const error = ref('')

const isCreateModalOpen = ref(false)
const isDeleteModalOpen = ref(false)
const selectedKeyToDelete = ref<APIKey | null>(null)
const deleteBusy = ref(false)

// Track which keys are currently revealed
const revealedKeys = ref<Record<number, boolean>>({})
// Track copy feedback per key id
const copiedKeyId = ref<number | null>(null)
const copiedSnippet = ref<string | null>(null)

const curlSnippet = `curl http://localhost:8080/v1/chat/completions \\
  -H "Authorization: Bearer <YOUR_API_KEY>" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gemini-2.5-flash",
    "messages": [{"role": "user", "content": "Hello Antigravity!"}]
  }'`

const pythonSnippet = `from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:8080/v1",
    api_key="<YOUR_API_KEY>",
)

response = client.chat.completions.create(
    model="gemini-2.5-flash",
    messages=[{"role": "user", "content": "Hello world"}],
)
print(response.choices[0].message.content)`

const geminiSnippet = `curl http://localhost:8080/v1beta/models/gemini-2.5-flash:generateContent?key=<YOUR_API_KEY> \\
  -H "Content-Type: application/json" \\
  -d '{
    "contents": [{"parts": [{"text": "Hello Gemini"}]}]
  }'`

async function fetchKeys() {
  loading.value = true
  error.value = ''
  try {
    const res = await apiKeysService.getKeys()
    keys.value = res.keys || []
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : 'Failed to fetch API keys'
  } finally {
    loading.value = false
  }
}

function handleKeyCreated(newKey: APIKey) {
  keys.value.unshift(newKey)
}

function openDeleteModal(keyItem: APIKey) {
  selectedKeyToDelete.value = keyItem
  isDeleteModalOpen.value = true
}

function closeDeleteModal() {
  if (deleteBusy.value) return
  isDeleteModalOpen.value = false
  selectedKeyToDelete.value = null
}

async function confirmDeleteKey() {
  if (!selectedKeyToDelete.value) return
  deleteBusy.value = true
  try {
    await apiKeysService.deleteKey(selectedKeyToDelete.value.id)
    keys.value = keys.value.filter((k) => k.id !== selectedKeyToDelete.value?.id)
    closeDeleteModal()
  } catch (err: unknown) {
    error.value = err instanceof Error ? err.message : 'Failed to delete API key'
  } finally {
    deleteBusy.value = false
  }
}

function toggleReveal(id: number) {
  revealedKeys.value[id] = !revealedKeys.value[id]
}

async function copyText(text: string, id: number) {
  try {
    await navigator.clipboard.writeText(text)
    copiedKeyId.value = id
    setTimeout(() => {
      if (copiedKeyId.value === id) copiedKeyId.value = null
    }, 2000)
  } catch {
    const textarea = document.createElement('textarea')
    textarea.value = text
    document.body.appendChild(textarea)
    textarea.select()
    document.execCommand('copy')
    document.body.removeChild(textarea)
    copiedKeyId.value = id
    setTimeout(() => {
      if (copiedKeyId.value === id) copiedKeyId.value = null
    }, 2000)
  }
}

async function copySnippet(text: string, label: string) {
  try {
    await navigator.clipboard.writeText(text)
    copiedSnippet.value = label
    setTimeout(() => {
      if (copiedSnippet.value === label) copiedSnippet.value = null
    }, 2000)
  } catch {
    // ignore
  }
}

function formatDate(dateStr: string): string {
  if (!dateStr) return '-'
  try {
    const d = new Date(dateStr)
    return d.toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    })
  } catch {
    return dateStr
  }
}

function formatRelativeTime(dateStr?: string | null): string {
  if (!dateStr) return 'Never used'
  try {
    const d = new Date(dateStr)
    const diffMs = Date.now() - d.getTime()
    const diffSec = Math.floor(diffMs / 1000)
    if (diffSec < 60) return 'Just now'
    const diffMin = Math.floor(diffSec / 60)
    if (diffMin < 60) return `${diffMin}m ago`
    const diffHr = Math.floor(diffMin / 60)
    if (diffHr < 24) return `${diffHr}h ago`
    const diffDays = Math.floor(diffHr / 24)
    if (diffDays < 30) return `${diffDays}d ago`
    return d.toLocaleDateString()
  } catch {
    return dateStr
  }
}

function maskKey(key: string): string {
  if (!key) return '••••••••••••••••'
  if (key.length <= 12) return '••••' + key.slice(-4)
  return key.slice(0, 7) + '••••••••••••••••' + key.slice(-4)
}

onMounted(() => {
  fetchKeys()
})
</script>

<template>
  <div class="space-y-6 pb-12">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
      <div>
        <h1 class="text-xl font-bold tracking-tight text-white flex items-center gap-2.5">
          <div class="h-8 w-8 rounded-lg bg-blue-500/10 border border-blue-500/20 text-blue-400 flex items-center justify-center">
            <KeyRound class="h-4.5 w-4.5" />
          </div>
          <span>API Keys</span>
        </h1>
        <p class="text-xs text-zinc-400 mt-1">
          Manage API secret keys to authenticate external AI clients (Cursor, Cline, Claude Code, OpenAI SDK, Gemini SDK, etc.)
        </p>
      </div>

      <div class="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          :disabled="loading"
          @click="fetchKeys"
          class="h-9 px-3 text-xs bg-[#121316] border-zinc-800 text-zinc-300 hover:text-white hover:bg-zinc-800/80 gap-1.5"
        >
          <RefreshCw class="h-3.5 w-3.5" :class="{ 'animate-spin': loading }" />
          <span>Refresh</span>
        </Button>
        <Button
          size="sm"
          @click="isCreateModalOpen = true"
          class="h-9 px-3.5 text-xs bg-blue-600 hover:bg-blue-500 text-white font-medium gap-1.5 rounded-lg shadow-sm"
        >
          <Plus class="h-4 w-4" />
          <span>Create API Key</span>
        </Button>
      </div>
    </div>

    <!-- Error Alert -->
    <div v-if="error" class="p-4 rounded-xl bg-red-500/10 border border-red-500/20 text-red-400 text-xs">
      {{ error }}
    </div>

    <!-- Metric Stats Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
      <!-- Active Keys -->
      <Card class="p-4 bg-[#121316] border-zinc-800/80 flex items-center justify-between shadow-xs">
        <div class="space-y-1">
          <span class="text-[11px] font-medium text-zinc-400 uppercase tracking-wider">Active Keys</span>
          <div class="text-2xl font-bold font-mono text-white">
            {{ keys.length }}
          </div>
          <span class="text-[10px] text-zinc-500 block">Registered API Tokens</span>
        </div>
        <div class="w-9 h-9 rounded-xl bg-blue-500/10 border border-blue-500/20 flex items-center justify-center text-blue-400">
          <KeyRound class="h-4.5 w-4.5" />
        </div>
      </Card>

      <!-- Auth Protocols -->
      <Card class="p-4 bg-[#121316] border-zinc-800/80 flex items-center justify-between shadow-xs">
        <div class="space-y-1">
          <span class="text-[11px] font-medium text-zinc-400 uppercase tracking-wider">Auth Protocols</span>
          <div class="text-xs font-bold font-mono text-emerald-400 pt-1">
            Bearer &bull; X-API-Key &bull; ?key=
          </div>
          <span class="text-[10px] text-zinc-500 block">OpenAI & Gemini Compatible</span>
        </div>
        <div class="w-9 h-9 rounded-xl bg-emerald-500/10 border border-emerald-500/20 flex items-center justify-center text-emerald-400">
          <ShieldCheck class="h-4.5 w-4.5" />
        </div>
      </Card>

      <!-- Security Guard -->
      <Card class="p-4 bg-[#121316] border-zinc-800/80 flex items-center justify-between shadow-xs">
        <div class="space-y-1">
          <span class="text-[11px] font-medium text-zinc-400 uppercase tracking-wider">Endpoint Guard</span>
          <div class="text-xs font-bold font-mono text-amber-400 pt-1">
            Active Protection
          </div>
          <span class="text-[10px] text-zinc-500 block">ADMIN_API_KEY fallback supported</span>
        </div>
        <div class="w-9 h-9 rounded-xl bg-amber-500/10 border border-amber-500/20 flex items-center justify-center text-amber-400">
          <Sparkles class="h-4.5 w-4.5" />
        </div>
      </Card>
    </div>

    <!-- API Keys Table Card -->
    <Card class="bg-[#121316] border-zinc-800/80 overflow-hidden shadow-xs">
      <div class="px-5 py-4 border-b border-zinc-800/80 flex items-center justify-between">
        <div>
          <h2 class="text-sm font-semibold text-white">Registered API Keys</h2>
          <p class="text-xs text-zinc-400">Secret keys configured to access proxy generation & MCP endpoints</p>
        </div>
        <Badge variant="outline" class="font-mono text-[11px] bg-blue-500/10 text-blue-400 border-blue-500/20">
          {{ keys.length }} Keys
        </Badge>
      </div>

      <div class="overflow-x-auto">
        <Table>
          <TableHeader class="bg-[#0e0f12]">
            <TableRow class="border-zinc-800/80 hover:bg-transparent">
              <TableHead class="text-zinc-400 text-xs font-semibold py-3 pl-5">Name / Label</TableHead>
              <TableHead class="text-zinc-400 text-xs font-semibold py-3">Secret Key</TableHead>
              <TableHead class="text-zinc-400 text-xs font-semibold py-3">Created</TableHead>
              <TableHead class="text-zinc-400 text-xs font-semibold py-3">Last Used</TableHead>
              <TableHead class="text-zinc-400 text-xs font-semibold py-3 text-right pr-5">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            <!-- Loading State -->
            <TableRow v-if="loading && keys.length === 0" class="border-zinc-800/60">
              <TableCell colspan="5" class="text-center py-8 text-xs text-zinc-500">
                <RefreshCw class="h-4 w-4 animate-spin mx-auto mb-2 text-zinc-400" />
                Loading API keys...
              </TableCell>
            </TableRow>

            <!-- Empty State -->
            <TableRow v-else-if="keys.length === 0" class="border-zinc-800/60">
              <TableCell colspan="5" class="text-center py-10 text-xs">
                <div class="h-10 w-10 rounded-full bg-zinc-800/50 border border-zinc-700/60 flex items-center justify-center mx-auto mb-3 text-zinc-400">
                  <KeyRound class="h-5 w-5" />
                </div>
                <div class="font-medium text-zinc-300">No API keys created yet</div>
                <p class="text-zinc-500 mt-1 max-w-sm mx-auto">
                  Create your first API key to start connecting tools like Cursor, Cline, or custom AI scripts.
                </p>
                <Button
                  size="sm"
                  @click="isCreateModalOpen = true"
                  class="mt-4 text-xs bg-blue-600 hover:bg-blue-500 text-white gap-1.5"
                >
                  <Plus class="h-3.5 w-3.5" />
                  <span>Create API Key</span>
                </Button>
              </TableCell>
            </TableRow>

            <!-- Keys Rows -->
            <TableRow
              v-for="k in keys"
              :key="k.id"
              class="border-zinc-800/60 hover:bg-zinc-800/20 transition-colors"
            >
              <!-- Name -->
              <TableCell class="py-3 pl-5">
                <div class="flex items-center gap-2.5">
                  <div class="h-7 w-7 rounded-lg bg-zinc-800/80 border border-zinc-700/60 text-zinc-300 flex items-center justify-center shrink-0">
                    <KeyRound class="h-3.5 w-3.5" />
                  </div>
                  <div class="min-w-0">
                    <div class="font-medium text-zinc-100 text-xs flex items-center gap-2">
                      <span class="truncate">{{ k.name }}</span>
                      <Badge variant="outline" class="font-mono text-[9px] px-1 py-0 bg-zinc-800 text-zinc-400 border-zinc-700">
                        ID: {{ k.id }}
                      </Badge>
                    </div>
                  </div>
                </div>
              </TableCell>

              <!-- Key Value -->
              <TableCell class="py-3">
                <div class="flex items-center gap-1.5 font-mono text-xs">
                  <span class="text-zinc-300 bg-[#0a0b0d] px-2.5 py-1 rounded border border-zinc-800 select-all">
                    {{ revealedKeys[k.id] ? k.key : maskKey(k.key) }}
                  </span>
                  <!-- Eye toggle -->
                  <button
                    type="button"
                    @click="toggleReveal(k.id)"
                    class="p-1 text-zinc-500 hover:text-zinc-300 transition-colors rounded cursor-pointer"
                    :title="revealedKeys[k.id] ? 'Hide secret key' : 'Show secret key'"
                  >
                    <EyeOff v-if="revealedKeys[k.id]" class="h-3.5 w-3.5" />
                    <Eye v-else class="h-3.5 w-3.5" />
                  </button>
                  <!-- Copy Button -->
                  <button
                    type="button"
                    @click="copyText(k.key, k.id)"
                    class="p-1 text-zinc-500 hover:text-blue-400 transition-colors rounded cursor-pointer"
                    :title="copiedKeyId === k.id ? 'Copied!' : 'Copy key'"
                  >
                    <Check v-if="copiedKeyId === k.id" class="h-3.5 w-3.5 text-emerald-400" />
                    <Copy v-else class="h-3.5 w-3.5" />
                  </button>
                </div>
              </TableCell>

              <!-- Created -->
              <TableCell class="py-3 text-xs text-zinc-400 font-sans">
                {{ formatDate(k.created_at) }}
              </TableCell>

              <!-- Last Used -->
              <TableCell class="py-3 text-xs">
                <span
                  class="font-mono text-[11px]"
                  :class="k.last_used_at ? 'text-emerald-400' : 'text-zinc-500'"
                >
                  {{ formatRelativeTime(k.last_used_at) }}
                </span>
              </TableCell>

              <!-- Actions -->
              <TableCell class="py-3 text-right pr-5">
                <Button
                  variant="ghost"
                  size="sm"
                  @click="openDeleteModal(k)"
                  class="h-8 w-8 p-0 text-zinc-500 hover:text-red-400 hover:bg-red-500/10 rounded-lg cursor-pointer"
                  title="Revoke and delete key"
                >
                  <Trash2 class="h-4 w-4" />
                </Button>
              </TableCell>
            </TableRow>
          </TableBody>
        </Table>
      </div>
    </Card>

    <!-- Quick Integration Guide -->
    <Card class="p-6 bg-[#121316] border-zinc-800/80 space-y-4 shadow-xs">
      <div class="flex items-center gap-3">
        <div class="h-9 w-9 rounded-xl bg-blue-500/10 border border-blue-500/20 text-blue-400 flex items-center justify-center shrink-0">
          <Code2 class="h-4.5 w-4.5" />
        </div>
        <div>
          <h3 class="text-sm font-semibold text-white">Client Integration Guide</h3>
          <p class="text-xs text-zinc-400">Configure your IDEs, code assistants, or HTTP clients with your API Key</p>
        </div>
      </div>

      <Tabs default-value="cursor" class="w-full">
        <TabsList class="bg-[#0e0f12] border border-zinc-800/80 p-0.5 h-9 rounded-lg">
          <TabsTrigger value="cursor" class="text-xs px-3 data-[state=active]:bg-zinc-800 data-[state=active]:text-white">
            <Cpu class="h-3.5 w-3.5 mr-1.5" />
            Cursor / IDE
          </TabsTrigger>
          <TabsTrigger value="curl" class="text-xs px-3 data-[state=active]:bg-zinc-800 data-[state=active]:text-white">
            <Terminal class="h-3.5 w-3.5 mr-1.5" />
            cURL / Shell
          </TabsTrigger>
          <TabsTrigger value="python" class="text-xs px-3 data-[state=active]:bg-zinc-800 data-[state=active]:text-white">
            Python (OpenAI)
          </TabsTrigger>
          <TabsTrigger value="gemini" class="text-xs px-3 data-[state=active]:bg-zinc-800 data-[state=active]:text-white">
            Gemini Native
          </TabsTrigger>
        </TabsList>

        <!-- Tab 1: Cursor -->
        <TabsContent value="cursor" class="pt-3">
          <div class="p-4 rounded-xl bg-[#0a0b0d] border border-zinc-800/80 space-y-3 text-xs font-mono">
            <div class="space-y-1.5">
              <div class="text-zinc-400 font-sans text-xs">Configure in Cursor Settings &rarr; Models &rarr; OpenAI API:</div>
              <div class="p-2.5 rounded bg-[#121316] border border-zinc-800 text-zinc-200 space-y-1">
                <div class="flex justify-between items-center">
                  <span class="text-zinc-500 font-sans">OpenAI Base URL:</span>
                  <span class="text-blue-300 font-medium select-all">http://localhost:8080/v1</span>
                </div>
                <div class="flex justify-between items-center">
                  <span class="text-zinc-500 font-sans">API Key:</span>
                  <span class="text-emerald-300 font-medium select-all">&lt;YOUR_API_KEY&gt;</span>
                </div>
              </div>
            </div>
            <div class="text-[11px] text-zinc-400 font-sans leading-relaxed">
              Enable models such as <code class="font-mono text-zinc-300">gemini-2.5-flash</code>, <code class="font-mono text-zinc-300">gemini-2.5-pro</code>, or <code class="font-mono text-zinc-300">claude-3-5-sonnet</code> in Cursor settings.
            </div>
          </div>
        </TabsContent>

        <!-- Tab 2: cURL -->
        <TabsContent value="curl" class="pt-3">
          <div class="relative p-4 rounded-xl bg-[#0a0b0d] border border-zinc-800/80 text-xs font-mono">
            <button
              type="button"
              @click="copySnippet(curlSnippet, 'curl')"
              class="absolute right-3 top-3 px-2 py-1 rounded bg-zinc-800/80 hover:bg-zinc-700 text-zinc-300 text-[11px] font-sans flex items-center gap-1 cursor-pointer transition-colors"
            >
              <Check v-if="copiedSnippet === 'curl'" class="h-3 w-3 text-emerald-400" />
              <Copy v-else class="h-3 w-3" />
              <span>{{ copiedSnippet === 'curl' ? 'Copied' : 'Copy' }}</span>
            </button>
            <pre class="text-zinc-300 leading-relaxed overflow-x-auto select-all"><code>{{ curlSnippet }}</code></pre>
          </div>
        </TabsContent>

        <!-- Tab 3: Python -->
        <TabsContent value="python" class="pt-3">
          <div class="relative p-4 rounded-xl bg-[#0a0b0d] border border-zinc-800/80 text-xs font-mono">
            <button
              type="button"
              @click="copySnippet(pythonSnippet, 'python')"
              class="absolute right-3 top-3 px-2 py-1 rounded bg-zinc-800/80 hover:bg-zinc-700 text-zinc-300 text-[11px] font-sans flex items-center gap-1 cursor-pointer transition-colors"
            >
              <Check v-if="copiedSnippet === 'python'" class="h-3 w-3 text-emerald-400" />
              <Copy v-else class="h-3 w-3" />
              <span>{{ copiedSnippet === 'python' ? 'Copied' : 'Copy' }}</span>
            </button>
            <pre class="text-zinc-300 leading-relaxed overflow-x-auto select-all"><code>{{ pythonSnippet }}</code></pre>
          </div>
        </TabsContent>

        <!-- Tab 4: Gemini Native -->
        <TabsContent value="gemini" class="pt-3">
          <div class="relative p-4 rounded-xl bg-[#0a0b0d] border border-zinc-800/80 text-xs font-mono">
            <button
              type="button"
              @click="copySnippet(geminiSnippet, 'gemini')"
              class="absolute right-3 top-3 px-2 py-1 rounded bg-zinc-800/80 hover:bg-zinc-700 text-zinc-300 text-[11px] font-sans flex items-center gap-1 cursor-pointer transition-colors"
            >
              <Check v-if="copiedSnippet === 'gemini'" class="h-3 w-3 text-emerald-400" />
              <Copy v-else class="h-3 w-3" />
              <span>{{ copiedSnippet === 'gemini' ? 'Copied' : 'Copy' }}</span>
            </button>
            <pre class="text-zinc-300 leading-relaxed overflow-x-auto select-all"><code>{{ geminiSnippet }}</code></pre>
          </div>
        </TabsContent>
      </Tabs>
    </Card>

    <!-- Modals -->
    <CreateKeyModal
      :open="isCreateModalOpen"
      @close="isCreateModalOpen = false"
      @created="handleKeyCreated"
    />

    <DeleteKeyModal
      :key-item="selectedKeyToDelete"
      :open="isDeleteModalOpen"
      :busy="deleteBusy"
      @close="closeDeleteModal"
      @confirm="confirmDeleteKey"
    />
  </div>
</template>
