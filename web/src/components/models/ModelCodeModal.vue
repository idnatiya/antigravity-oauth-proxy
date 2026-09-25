<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Copy, Check, Terminal, FileCode, Code2 } from '@lucide/vue'

const props = defineProps<{
  modelId: string | null
  open: boolean
}>()

const emit = defineEmits<{
  (e: 'update:open', val: boolean): void
}>()

const activeTab = ref<'curl' | 'python' | 'typescript'>('curl')
const copied = ref(false)

const baseUrl = computed(() => {
  if (typeof window !== 'undefined') {
    return `${window.location.origin}/v1`
  }
  return 'http://localhost:8080/v1'
})

const curlSnippet = computed(() => {
  const model = props.modelId || 'gemini-2.5-flash'
  return `curl -X POST ${baseUrl.value}/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer $ADMIN_API_KEY" \\
  -d '{
    "model": "${model}",
    "messages": [
      {
        "role": "user",
        "content": "Hello! Explain quantum computing in 2 sentences."
      }
    ],
    "stream": true
  }'`
})

const pythonSnippet = computed(() => {
  const model = props.modelId || 'gemini-2.5-flash'
  return `from openai import OpenAI

client = OpenAI(
    base_url="${baseUrl.value}",
    api_key="your-api-key",  # ADMIN_API_KEY
)

stream = client.chat.completions.create(
    model="${model}",
    messages=[
        {"role": "user", "content": "Hello! Explain quantum computing in 2 sentences."}
    ],
    stream=True,
)

for chunk in stream:
    content = chunk.choices[0].delta.content or ""
    print(content, end="", flush=True)
print()`
})

const typescriptSnippet = computed(() => {
  const model = props.modelId || 'gemini-2.5-flash'
  return `import OpenAI from "openai";

const openai = new OpenAI({
  baseURL: "${baseUrl.value}",
  apiKey: "your-api-key", // ADMIN_API_KEY
});

async function main() {
  const stream = await openai.chat.completions.create({
    model: "${model}",
    messages: [
      { role: "user", content: "Hello! Explain quantum computing in 2 sentences." }
    ],
    stream: true,
  });

  for await (const chunk of stream) {
    process.stdout.write(chunk.choices[0]?.delta?.content || "");
  }
  console.log();
}

main();`
})

const currentSnippet = computed(() => {
  if (activeTab.value === 'curl') return curlSnippet.value
  if (activeTab.value === 'python') return pythonSnippet.value
  return typescriptSnippet.value
})

async function copySnippet() {
  try {
    await navigator.clipboard.writeText(currentSnippet.value)
    copied.value = true
    setTimeout(() => {
      copied.value = false
    }, 2000)
  } catch (err) {
    console.error('Failed to copy snippet', err)
  }
}
</script>

<template>
  <Dialog :open="open" @update:open="emit('update:open', $event)">
    <DialogContent
      class="max-w-2xl bg-[#121316] border border-zinc-800/90 text-zinc-100 p-0 overflow-hidden shadow-2xl rounded-2xl"
    >
      <DialogHeader class="p-6 pb-4 border-b border-zinc-800/80">
        <div class="flex items-center gap-3">
          <div class="w-9 h-9 rounded-xl bg-blue-500/10 border border-blue-500/20 flex items-center justify-center text-blue-400 shrink-0">
            <Code2 class="h-4.5 w-4.5" />
          </div>
          <div class="min-w-0 flex-1">
            <DialogTitle class="text-sm font-semibold text-white flex items-center gap-2">
              <span>Integration Code Snippet</span>
              <Badge variant="outline" class="font-mono text-[10px] bg-blue-500/10 text-blue-400 border-blue-500/20 px-2 py-0.5">
                {{ modelId }}
              </Badge>
            </DialogTitle>
            <DialogDescription class="text-xs text-zinc-400 mt-1">
              Ready-to-use API call using OpenAI-compatible endpoints with streaming support.
            </DialogDescription>
          </div>
        </div>
      </DialogHeader>

      <div class="p-6 space-y-4">
        <!-- Tabs -->
        <Tabs v-model="activeTab" class="w-full">
          <div class="flex items-center justify-between gap-2 pb-2">
            <TabsList class="bg-[#0d0e11] border border-zinc-800/80 p-0.5 rounded-lg h-8">
              <TabsTrigger
                value="curl"
                class="text-xs px-3 py-1 data-[state=active]:bg-zinc-800 data-[state=active]:text-white text-zinc-400 rounded-md gap-1.5 flex items-center transition-all"
              >
                <Terminal class="h-3.5 w-3.5" />
                <span>cURL</span>
              </TabsTrigger>
              <TabsTrigger
                value="python"
                class="text-xs px-3 py-1 data-[state=active]:bg-zinc-800 data-[state=active]:text-white text-zinc-400 rounded-md gap-1.5 flex items-center transition-all"
              >
                <FileCode class="h-3.5 w-3.5" />
                <span>Python</span>
              </TabsTrigger>
              <TabsTrigger
                value="typescript"
                class="text-xs px-3 py-1 data-[state=active]:bg-zinc-800 data-[state=active]:text-white text-zinc-400 rounded-md gap-1.5 flex items-center transition-all"
              >
                <FileCode class="h-3.5 w-3.5" />
                <span>TypeScript</span>
              </TabsTrigger>
            </TabsList>

            <Button
              variant="outline"
              size="sm"
              @click="copySnippet"
              class="h-8 px-3 text-xs bg-zinc-900 border-zinc-800 hover:bg-zinc-800 hover:text-white text-zinc-300 gap-1.5 transition-all"
            >
              <Check v-if="copied" class="h-3.5 w-3.5 text-emerald-400" />
              <Copy v-else class="h-3.5 w-3.5" />
              <span>{{ copied ? 'Copied!' : 'Copy Code' }}</span>
            </Button>
          </div>

          <TabsContent value="curl" class="mt-2">
            <div class="relative rounded-xl border border-zinc-800/80 bg-[#090a0c] p-4 font-mono text-xs text-zinc-200 overflow-x-auto leading-relaxed selection:bg-blue-600/30">
              <pre>{{ curlSnippet }}</pre>
            </div>
          </TabsContent>

          <TabsContent value="python" class="mt-2">
            <div class="relative rounded-xl border border-zinc-800/80 bg-[#090a0c] p-4 font-mono text-xs text-zinc-200 overflow-x-auto leading-relaxed selection:bg-blue-600/30">
              <pre>{{ pythonSnippet }}</pre>
            </div>
          </TabsContent>

          <TabsContent value="typescript" class="mt-2">
            <div class="relative rounded-xl border border-zinc-800/80 bg-[#090a0c] p-4 font-mono text-xs text-zinc-200 overflow-x-auto leading-relaxed selection:bg-blue-600/30">
              <pre>{{ typescriptSnippet }}</pre>
            </div>
          </TabsContent>
        </Tabs>

        <!-- Informative callout -->
        <div class="rounded-xl bg-blue-500/5 border border-blue-500/15 p-3 text-xs text-zinc-400 flex items-start gap-2.5">
          <Terminal class="h-4 w-4 text-blue-400 mt-0.5 shrink-0" />
          <div class="space-y-1">
            <p class="text-zinc-300 font-medium">OpenAI Compatibility Layer</p>
            <p class="text-[11px] text-zinc-400 leading-normal">
              This endpoint seamlessly transforms standard OpenAI payload structures into Google CloudCode internal formats with OAuth tokens auto-refreshed.
            </p>
          </div>
        </div>
      </div>
    </DialogContent>
  </Dialog>
</template>
