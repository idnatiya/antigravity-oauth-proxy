<script setup lang="ts">
import { ref, computed } from 'vue'
import {
  Sparkles,
  Download,
  Copy,
  Check,
  RefreshCw,
  Image as ImageIcon,
  Upload,
  X,
  Wand2,
  Code2,
  ZoomIn,
  AlertCircle,
  Clock,
  ArrowRight,
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Card } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { imagesService } from '@/services/imagesService'

interface HistoryItem {
  id: string
  url: string
  b64: string
  prompt: string
  revisedPrompt?: string
  mode: 'generate' | 'edit'
  aspectRatio: string
  timestamp: string
  durationMs: number
}

// State
const activeTab = ref<'generate' | 'edit'>('generate')
const promptText = ref('')
const selectedRatio = ref('1:1')
const selectedStyle = ref('vivid')
const isGenerating = ref(false)
const isEnhancing = ref(false)
const errorMessage = ref('')
const currentImage = ref<HistoryItem | null>(null)
const history = ref<HistoryItem[]>([])

// Edit state
const editSourceImage = ref<string | null>(null)
const editSourceFileName = ref('')
const fileInputRef = ref<HTMLInputElement | null>(null)

// Action feedback states
const copiedOutput = ref(false)
const copiedBase64 = ref(false)
const copiedCurl = ref(false)
const showLightbox = ref(false)
const showCodeModal = ref(false)
const generationTimer = ref(0)
let timerInterval: ReturnType<typeof setInterval> | null = null

// Aspect ratio options
const aspectRatios = [
  { id: '1:1', label: '1:1', desc: 'Square', size: '1024x1024', icon: '■' },
  { id: '16:9', label: '16:9', desc: 'Landscape', size: '1792x1024', icon: '▬' },
  { id: '9:16', label: '9:16', desc: 'Portrait', size: '1024x1792', icon: '▮' },
  { id: '4:3', label: '4:3', desc: 'Photo', size: '1024x768', icon: '▭' },
  { id: '3:4', label: '3:4', desc: 'Tall', size: '768x1024', icon: '▯' },
]

// Prompt Presets
const promptPresets = [
  { label: 'Cyberpunk City', prompt: 'A neon-lit cyberpunk metropolis at rainy twilight, reflections in wet asphalt puddles, holographic billboards, cinematic lighting, ultra-detailed 8k render' },
  { label: 'Cute 3D Mascot', prompt: 'A cute fluffy robot creature sitting on a desk with a warm coffee mug, soft studio clay lighting, Pixar 3D style, vibrant friendly colors' },
  { label: 'Studio Product', prompt: 'Luxury ceramic coffee cup on a raw marble pedestal, soft directional morning sunlight, minimalist aesthetic, sharp focus, high-end commercial photography' },
  { label: 'Watercolor Landscape', prompt: 'Dreamy watercolor painting of misty pine mountains and a serene crystal lake at dawn, soft pastel hues, artistic brush bleed effects' },
  { label: 'Anime Sunset', prompt: 'Makoto Shinkai style anime scenery of a hilltop shrine overlook at golden hour, dramatic volumetric clouds, glowing lens flare, vivid emotions' },
]

// Handlers
function applyPreset(preset: { prompt: string }) {
  promptText.value = preset.prompt
}

// Enhance Prompt using Gemini 3.8
async function handleEnhancePrompt() {
  if (!promptText.value.trim() || isEnhancing.value) return
  isEnhancing.value = true
  errorMessage.value = ''
  try {
    const res = await imagesService.enhancePrompt({
      prompt: promptText.value,
      style: selectedStyle.value,
    })
    if (res && res.enhanced_prompt) {
      promptText.value = res.enhanced_prompt
    }
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : String(err)
    errorMessage.value = `Failed to enhance prompt with Gemini 3.8: ${msg}`
  } finally {
    isEnhancing.value = false
  }
}

// File Upload for Edit
function handleFileSelect(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files && target.files[0]) {
    loadFile(target.files[0])
  }
}

function handleDrop(e: DragEvent) {
  e.preventDefault()
  if (e.dataTransfer && e.dataTransfer.files[0]) {
    loadFile(e.dataTransfer.files[0])
  }
}

function loadFile(file: File) {
  if (!file.type.startsWith('image/')) {
    errorMessage.value = 'Please select a valid image file (PNG, JPEG, WebP)'
    return
  }
  editSourceFileName.value = file.name
  const reader = new FileReader()
  reader.onload = (ev) => {
    editSourceImage.value = ev.target?.result as string
  }
  reader.readAsDataURL(file)
}

function clearEditSource() {
  editSourceImage.value = null
  editSourceFileName.value = ''
  if (fileInputRef.value) {
    fileInputRef.value.value = ''
  }
}

// Send Current Image to Edit Tab
function useCurrentAsEditInput() {
  if (!currentImage.value) return
  editSourceImage.value = currentImage.value.url
  editSourceFileName.value = `image-${currentImage.value.id.slice(0, 6)}.png`
  activeTab.value = 'edit'
  promptText.value = ''
}

// Main Submit (Generate or Edit)
async function handleSubmit() {
  if (!promptText.value.trim() || isGenerating.value) return
  if (activeTab.value === 'edit' && !editSourceImage.value) {
    errorMessage.value = 'Please upload a source image to edit'
    return
  }

  isGenerating.value = true
  errorMessage.value = ''
  generationTimer.value = 0
  const startTime = Date.now()

  timerInterval = setInterval(() => {
    generationTimer.value = Math.floor((Date.now() - startTime) / 100) / 10
  }, 100)

  try {
    let res
    const selectedSize = aspectRatios.find(r => r.id === selectedRatio.value)?.size || '1024x1024'

    if (activeTab.value === 'generate') {
      res = await imagesService.generateImage({
        prompt: promptText.value,
        size: selectedSize,
        style: selectedStyle.value,
        model: 'gemini-3.1-flash-image',
      })
    } else {
      res = await imagesService.editImage({
        image: editSourceImage.value!,
        prompt: promptText.value,
        size: selectedSize,
        model: 'gemini-3.1-flash-image',
      })
    }

    const duration = Date.now() - startTime
    if (res && res.data && res.data.length > 0) {
      const item = res.data[0]
      const imageUrl = item.url || (item.b64_json ? `data:image/png;base64,${item.b64_json}` : '')
      const newHistoryItem: HistoryItem = {
        id: Math.random().toString(36).substring(2, 9),
        url: imageUrl,
        b64: item.b64_json || '',
        prompt: promptText.value,
        revisedPrompt: item.revised_prompt,
        mode: activeTab.value,
        aspectRatio: selectedRatio.value,
        timestamp: new Date().toLocaleTimeString(),
        durationMs: duration,
      }
      currentImage.value = newHistoryItem
      history.value.unshift(newHistoryItem)
    } else {
      throw new Error('No image returned from upstream service')
    }
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : String(err)
    errorMessage.value = msg || 'Image generation failed'
  } finally {
    isGenerating.value = false
    if (timerInterval) {
      clearInterval(timerInterval)
      timerInterval = null
    }
  }
}

// Download
function downloadImage(item: HistoryItem) {
  const link = document.createElement('a')
  link.href = item.url
  link.download = `antigravity-${item.mode}-${item.id}.png`
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

// Copy Image to Clipboard
async function copyImageToClipboard(item: HistoryItem) {
  try {
    const res = await fetch(item.url)
    const blob = await res.blob()
    await navigator.clipboard.write([
      new ClipboardItem({ [blob.type || 'image/png']: blob }),
    ])
    copiedOutput.value = true
    setTimeout(() => { copiedOutput.value = false }, 2000)
  } catch (err) {
    // Fallback: Copy data URL as text
    await navigator.clipboard.writeText(item.url)
    copiedOutput.value = true
    setTimeout(() => { copiedOutput.value = false }, 2000)
  }
}

// Copy Base64
async function copyBase64(item: HistoryItem) {
  const text = item.b64 || item.url.replace(/^data:image\/[a-z]+;base64,/, '')
  await navigator.clipboard.writeText(text)
  copiedBase64.value = true
  setTimeout(() => { copiedBase64.value = false }, 2000)
}

// Code Snippets
const curlSnippet = computed(() => {
  const prompt = promptText.value || 'A cute baby sea otter'
  if (activeTab.value === 'generate') {
    return `curl -X POST http://localhost:9878/v1/images/generations \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -d '{
    "model": "gemini-3.1-flash-image",
    "prompt": "${prompt.replace(/"/g, '\\"')}",
    "size": "${selectedRatio.value}",
    "response_format": "b64_json"
  }'`
  }
  return `curl -X POST http://localhost:9878/v1/images/edits \\
  -H "Authorization: Bearer YOUR_API_KEY" \\
  -F "image=@photo.png" \\
  -F "prompt=${prompt.replace(/"/g, '\\"')}" \\
  -F "model=gemini-3.1-flash-image"`
})

const pythonSnippet = computed(() => {
  const prompt = promptText.value || 'A cute baby sea otter'
  return `from openai import OpenAI

client = OpenAI(
    base_url="http://localhost:9878/v1",
    api_key="YOUR_API_KEY"
)

# Text-to-Image Generation
response = client.images.generate(
    model="gemini-3.1-flash-image",
    prompt="${prompt.replace(/"/g, '\\"')}",
    size="${selectedRatio.value}",
    response_format="b64_json"
)

print(response.data[0].url)`
})

async function copyCode(text: string) {
  await navigator.clipboard.writeText(text)
  copiedCurl.value = true
  setTimeout(() => { copiedCurl.value = false }, 2000)
}
</script>

<template>
  <div class="space-y-6 pb-12">
    <!-- Header -->
    <div class="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 border-b border-zinc-800/80 pb-5">
      <div>
        <div class="flex items-center gap-3">
          <h1 class="text-2xl font-bold tracking-tight text-white flex items-center gap-2">
            Image Studio
          </h1>
          <Badge variant="outline" class="bg-blue-500/10 text-blue-400 border-blue-500/30 gap-1.5 font-mono text-xs">
            <Sparkles class="h-3 w-3" />
            gemini-3.1-flash-image
          </Badge>
        </div>
        <p class="text-sm text-zinc-400 mt-1">
          Generate & edit images using Google CloudCode diffusion models with Gemini 3.8 prompt reasoning.
        </p>
      </div>

      <div class="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          class="border-zinc-800 text-zinc-300 hover:text-white bg-zinc-900/60"
          @click="showCodeModal = !showCodeModal"
        >
          <Code2 class="h-4 w-4 mr-1.5 text-blue-400" />
          API Snippets
        </Button>
      </div>
    </div>

    <!-- Error Banner -->
    <div
      v-if="errorMessage"
      class="flex items-start gap-3 p-4 rounded-xl bg-red-500/10 border border-red-500/30 text-red-200 text-sm"
    >
      <AlertCircle class="h-5 w-5 text-red-400 shrink-0 mt-0.5" />
      <div class="flex-1">{{ errorMessage }}</div>
      <button class="text-red-400 hover:text-red-200" @click="errorMessage = ''">
        <X class="h-4 w-4" />
      </button>
    </div>

    <!-- Main Workspace Grid -->
    <div class="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
      <!-- Left Column: Controls & Prompt Input (5 cols) -->
      <div class="lg:col-span-5 space-y-5">
        <Card class="bg-zinc-900/50 border-zinc-800/80 p-5 space-y-5 shadow-xl backdrop-blur-sm">
          <!-- Mode Tabs -->
          <div class="flex rounded-lg bg-zinc-950/80 p-1 border border-zinc-800/60">
            <button
              class="flex-1 py-2 px-3 rounded-md text-xs font-medium transition-all flex items-center justify-center gap-2"
              :class="activeTab === 'generate' ? 'bg-blue-600 text-white shadow' : 'text-zinc-400 hover:text-zinc-200'"
              @click="activeTab = 'generate'"
            >
              <Sparkles class="h-3.5 w-3.5" />
              <span>Generate (Text to Image)</span>
            </button>
            <button
              class="flex-1 py-2 px-3 rounded-md text-xs font-medium transition-all flex items-center justify-center gap-2"
              :class="activeTab === 'edit' ? 'bg-indigo-600 text-white shadow' : 'text-zinc-400 hover:text-zinc-200'"
              @click="activeTab = 'edit'"
            >
              <Wand2 class="h-3.5 w-3.5" />
              <span>Edit (Image to Image)</span>
            </button>
          </div>

          <!-- Edit Mode: Source Image Upload Dropzone -->
          <div v-if="activeTab === 'edit'" class="space-y-2">
            <label class="text-xs font-medium text-zinc-300 flex items-center justify-between">
              <span>Source Image to Edit</span>
              <span v-if="editSourceImage" class="text-zinc-500 font-mono text-[11px]">{{ editSourceFileName }}</span>
            </label>

            <!-- Upload Box -->
            <div
              v-if="!editSourceImage"
              class="border-2 border-dashed border-zinc-700/80 hover:border-indigo-500/60 rounded-xl p-6 text-center cursor-pointer transition-colors bg-zinc-950/40"
              @click="fileInputRef?.click()"
              @dragover.prevent
              @drop="handleDrop"
            >
              <input
                ref="fileInputRef"
                type="file"
                accept="image/*"
                class="hidden"
                @change="handleFileSelect"
              />
              <div class="flex flex-col items-center gap-2">
                <div class="h-10 w-10 rounded-full bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center text-indigo-400">
                  <Upload class="h-5 w-5" />
                </div>
                <div class="text-sm font-medium text-zinc-200">Click to upload or drag & drop</div>
                <div class="text-xs text-zinc-500">PNG, JPG, WebP up to 10MB</div>
              </div>
            </div>

            <!-- Uploaded Preview -->
            <div v-else class="relative rounded-xl border border-zinc-800 overflow-hidden group bg-zinc-950">
              <img :src="editSourceImage" class="w-full max-h-48 object-contain bg-zinc-950" />
              <button
                class="absolute top-2 right-2 p-1.5 rounded-lg bg-zinc-900/80 hover:bg-red-500/80 text-zinc-300 hover:text-white transition-colors"
                title="Remove image"
                @click="clearEditSource"
              >
                <X class="h-4 w-4" />
              </button>
            </div>
          </div>

          <!-- Prompt Input Area -->
          <div class="space-y-2">
            <div class="flex items-center justify-between">
              <label class="text-xs font-medium text-zinc-300 flex items-center gap-1.5">
                <span>{{ activeTab === 'generate' ? 'Image Prompt' : 'Edit Instructions' }}</span>
              </label>

              <!-- Gemini 3.8 Enhance Button -->
              <button
                type="button"
                class="inline-flex items-center gap-1.5 text-xs font-medium px-2.5 py-1 rounded-md bg-gradient-to-r from-blue-500/10 to-indigo-500/10 hover:from-blue-500/20 hover:to-indigo-500/20 text-blue-300 border border-blue-500/30 transition-all shadow-sm disabled:opacity-50"
                :disabled="isEnhancing || !promptText.trim()"
                @click="handleEnhancePrompt"
              >
                <RefreshCw v-if="isEnhancing" class="h-3.5 w-3.5 animate-spin text-blue-400" />
                <Sparkles v-else class="h-3.5 w-3.5 text-blue-400" />
                <span>{{ isEnhancing ? 'Enhancing with Gemini 3.8...' : '✨ Enhance Prompt (3.8)' }}</span>
              </button>
            </div>

            <textarea
              v-model="promptText"
              rows="4"
              class="w-full rounded-xl bg-zinc-950 border border-zinc-800 p-3.5 text-sm text-zinc-100 placeholder-zinc-500 focus:outline-none focus:ring-2 focus:ring-blue-500/50 focus:border-blue-500 transition-all resize-y"
              :placeholder="activeTab === 'generate'
                ? 'Describe the image you want to generate in detail (subject, style, lighting, camera angle, colors)...'
                : 'Describe what to modify (e.g. Change the background to a tropical beach at sunset, keep the subject same)...'"
              @keydown.enter.ctrl.prevent="handleSubmit"
              @keydown.enter.meta.prevent="handleSubmit"
            ></textarea>
            <div class="flex items-center justify-between text-[11px] text-zinc-500">
              <span>Press <kbd class="px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-300 font-mono">Ctrl+Enter</kbd> to run</span>
              <span>{{ promptText.length }} chars</span>
            </div>
          </div>

          <!-- Prompt Preset Pills (Generate Mode Only) -->
          <div v-if="activeTab === 'generate'" class="space-y-2">
            <div class="text-[11px] font-medium text-zinc-400">Quick Inspirations:</div>
            <div class="flex flex-wrap gap-1.5">
              <button
                v-for="preset in promptPresets"
                :key="preset.label"
                type="button"
                class="text-xs px-2.5 py-1 rounded-md bg-zinc-800/60 hover:bg-zinc-800 text-zinc-300 hover:text-white border border-zinc-700/50 transition-colors"
                @click="applyPreset(preset)"
              >
                {{ preset.label }}
              </button>
            </div>
          </div>

          <!-- Aspect Ratio Options -->
          <div class="space-y-2">
            <label class="text-xs font-medium text-zinc-300">Aspect Ratio</label>
            <div class="grid grid-cols-5 gap-1.5">
              <button
                v-for="ratio in aspectRatios"
                :key="ratio.id"
                type="button"
                class="flex flex-col items-center justify-center p-2 rounded-lg border text-center transition-all"
                :class="selectedRatio === ratio.id
                  ? 'bg-blue-500/15 border-blue-500/40 text-blue-300 font-semibold'
                  : 'bg-zinc-950/60 border-zinc-800 text-zinc-400 hover:text-zinc-200 hover:border-zinc-700'"
                @click="selectedRatio = ratio.id"
              >
                <span class="text-xs font-mono font-medium">{{ ratio.label }}</span>
                <span class="text-[10px] text-zinc-500">{{ ratio.desc }}</span>
              </button>
            </div>
          </div>

          <!-- Style Selector -->
          <div class="space-y-2">
            <label class="text-xs font-medium text-zinc-300">Aesthetic Style</label>
            <select
              v-model="selectedStyle"
              class="w-full rounded-lg bg-zinc-950 border border-zinc-800 p-2.5 text-xs text-zinc-200 focus:outline-none focus:ring-1 focus:ring-blue-500"
            >
              <option value="vivid">Vivid (Hyper-saturated, dramatic highlights)</option>
              <option value="natural">Natural (Subtle, realistic lighting)</option>
              <option value="cinematic">Cinematic (Anamorphic lens flare, moody color grading)</option>
              <option value="digital_art">Digital Art (Concept art, octane render 3D)</option>
              <option value="photorealistic">Photorealistic (Fine textures, candid photo look)</option>
            </select>
          </div>

          <!-- Action Button -->
          <Button
            type="button"
            class="w-full h-11 text-sm font-semibold rounded-xl transition-all shadow-lg"
            :class="activeTab === 'generate'
              ? 'bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white shadow-blue-500/20'
              : 'bg-gradient-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white shadow-purple-500/20'"
            :disabled="isGenerating || !promptText.trim() || (activeTab === 'edit' && !editSourceImage)"
            @click="handleSubmit"
          >
            <RefreshCw v-if="isGenerating" class="h-4 w-4 mr-2 animate-spin" />
            <Sparkles v-else-if="activeTab === 'generate'" class="h-4 w-4 mr-2" />
            <Wand2 v-else class="h-4 w-4 mr-2" />
            <span>
              {{ isGenerating
                ? `Rendering Pixels... (${generationTimer.toFixed(1)}s)`
                : activeTab === 'generate' ? 'Generate Image' : 'Apply Image Edit' }}
            </span>
          </Button>
        </Card>
      </div>

      <!-- Right Column: Canvas, Preview & History Strip (7 cols) -->
      <div class="lg:col-span-7 space-y-5">
        <!-- Canvas Card -->
        <Card class="bg-zinc-900/50 border-zinc-800/80 p-5 shadow-xl backdrop-blur-sm min-h-[480px] flex flex-col justify-between">
          <!-- Canvas Header / Status -->
          <div class="flex items-center justify-between pb-3 border-b border-zinc-800/60 text-xs">
            <div class="flex items-center gap-2">
              <span class="font-medium text-zinc-300">Canvas</span>
              <Badge v-if="currentImage" variant="outline" class="font-mono text-[10px] text-zinc-400 border-zinc-800">
                {{ currentImage.aspectRatio }}
              </Badge>
              <Badge v-if="currentImage" variant="outline" class="font-mono text-[10px] text-emerald-400 border-emerald-500/30">
                {{ (currentImage.durationMs / 1000).toFixed(2) }}s
              </Badge>
            </div>

            <!-- Toolbar Actions (When Image is Loaded) -->
            <div v-if="currentImage && !isGenerating" class="flex items-center gap-1.5">
              <Button
                variant="outline"
                size="sm"
                class="h-8 px-2.5 text-xs border-zinc-800 bg-zinc-950 text-zinc-300 hover:text-white"
                title="Use this image as input for Edit mode"
                @click="useCurrentAsEditInput"
              >
                <ArrowRight class="h-3.5 w-3.5 mr-1 text-indigo-400" />
                <span>Iterate Edit</span>
              </Button>

              <Button
                variant="outline"
                size="sm"
                class="h-8 px-2.5 text-xs border-zinc-800 bg-zinc-950 text-zinc-300 hover:text-white"
                title="Copy image bytes to clipboard"
                @click="copyImageToClipboard(currentImage)"
              >
                <Check v-if="copiedOutput" class="h-3.5 w-3.5 text-emerald-400 mr-1" />
                <Copy v-else class="h-3.5 w-3.5 mr-1" />
                <span>{{ copiedOutput ? 'Copied!' : 'Copy' }}</span>
              </Button>

              <Button
                variant="outline"
                size="sm"
                class="h-8 px-2.5 text-xs border-zinc-800 bg-zinc-950 text-zinc-300 hover:text-white"
                title="Download as PNG"
                @click="downloadImage(currentImage)"
              >
                <Download class="h-3.5 w-3.5 mr-1" />
                <span>PNG</span>
              </Button>

              <Button
                variant="outline"
                size="sm"
                class="h-8 px-2.5 text-xs border-zinc-800 bg-zinc-950 text-zinc-300 hover:text-white"
                title="Copy Base64 string"
                @click="copyBase64(currentImage)"
              >
                <Check v-if="copiedBase64" class="h-3.5 w-3.5 text-emerald-400 mr-1" />
                <Copy v-else class="h-3.5 w-3.5 mr-1 text-zinc-400" />
                <span>{{ copiedBase64 ? 'Copied B64!' : 'B64' }}</span>
              </Button>

              <Button
                variant="outline"
                size="sm"
                class="h-8 px-2 text-xs border-zinc-800 bg-zinc-950 text-zinc-300 hover:text-white"
                title="Zoom Lightbox"
                @click="showLightbox = true"
              >
                <ZoomIn class="h-3.5 w-3.5" />
              </Button>
            </div>
          </div>

          <!-- Main Canvas Display -->
          <div class="flex-1 flex items-center justify-center my-4 relative">
            <!-- Loading State Skeleton -->
            <div v-if="isGenerating" class="w-full flex flex-col items-center justify-center py-16 space-y-4">
              <div class="relative">
                <div class="h-20 w-20 rounded-2xl bg-gradient-to-tr from-blue-600/30 to-indigo-600/30 border border-blue-500/40 flex items-center justify-center animate-pulse">
                  <Sparkles class="h-10 w-10 text-blue-400 animate-spin" />
                </div>
                <div class="absolute -bottom-1 -right-1 h-6 w-6 rounded-full bg-emerald-500/20 border border-emerald-500/50 flex items-center justify-center text-emerald-400">
                  <Clock class="h-3 w-3" />
                </div>
              </div>
              <div class="text-center space-y-1">
                <div class="text-sm font-semibold text-white">Rendering with gemini-3.1-flash-image</div>
                <div class="text-xs text-zinc-500 font-mono">Elapsed time: {{ generationTimer.toFixed(1) }}s</div>
              </div>
            </div>

            <!-- Empty Initial State -->
            <div v-else-if="!currentImage" class="w-full flex flex-col items-center justify-center py-16 text-center space-y-3">
              <div class="h-16 w-16 rounded-2xl bg-zinc-950 border border-zinc-800 flex items-center justify-center text-zinc-600">
                <ImageIcon class="h-8 w-8" />
              </div>
              <div class="space-y-1 max-w-sm">
                <div class="text-sm font-medium text-zinc-300">Ready to create</div>
                <div class="text-xs text-zinc-500">
                  Type a prompt on the left or click one of the quick presets to render your first image.
                </div>
              </div>
            </div>

            <!-- Rendered Image -->
            <div v-else class="w-full flex flex-col items-center justify-center">
              <div
                class="rounded-xl overflow-hidden border border-zinc-800 shadow-2xl bg-zinc-950 cursor-pointer transition-transform hover:scale-[1.01]"
                @click="showLightbox = true"
              >
                <img
                  :src="currentImage.url"
                  class="max-h-[500px] w-auto object-contain mx-auto"
                  alt="Generated AI Image"
                />
              </div>

              <!-- Metadata Box Below Image -->
              <div class="w-full mt-4 p-3 rounded-lg bg-zinc-950/80 border border-zinc-800/80 space-y-1.5 text-left">
                <div class="flex items-center justify-between text-[11px] text-zinc-500 font-mono">
                  <span>PROMPT</span>
                  <span>{{ currentImage.timestamp }}</span>
                </div>
                <p class="text-xs text-zinc-300 leading-relaxed">{{ currentImage.prompt }}</p>
                <div v-if="currentImage.revisedPrompt && currentImage.revisedPrompt !== currentImage.prompt" class="pt-1.5 border-t border-zinc-900 text-[11px] text-zinc-400">
                  <span class="text-zinc-500 font-mono">REVISED: </span>{{ currentImage.revisedPrompt }}
                </div>
              </div>
            </div>
          </div>

          <!-- Session History Thumbnail Strip -->
          <div v-if="history.length > 1" class="pt-3 border-t border-zinc-800/60 space-y-2">
            <div class="flex items-center justify-between text-xs text-zinc-400">
              <span class="font-medium">Session History ({{ history.length }})</span>
            </div>
            <div class="flex items-center gap-2 overflow-x-auto pb-1 scrollbar-thin">
              <button
                v-for="item in history"
                :key="item.id"
                class="h-16 w-16 shrink-0 rounded-lg overflow-hidden border transition-all relative group"
                :class="currentImage?.id === item.id ? 'border-blue-500 ring-2 ring-blue-500/30' : 'border-zinc-800 hover:border-zinc-600 opacity-70 hover:opacity-100'"
                @click="currentImage = item"
              >
                <img :src="item.url" class="h-full w-full object-cover" />
                <span class="absolute bottom-0 right-0 px-1 text-[9px] font-mono bg-black/70 text-zinc-300">
                  {{ item.aspectRatio }}
                </span>
              </button>
            </div>
          </div>
        </Card>
      </div>
    </div>

    <!-- Lightbox Zoom Modal -->
    <div
      v-if="showLightbox && currentImage"
      class="fixed inset-0 z-50 bg-black/90 backdrop-blur-md flex items-center justify-center p-4"
      @click.self="showLightbox = false"
    >
      <div class="relative max-w-5xl max-h-[90vh] flex flex-col items-center">
        <button
          class="absolute -top-10 right-0 text-zinc-400 hover:text-white p-2"
          @click="showLightbox = false"
        >
          <X class="h-6 w-6" />
        </button>
        <img
          :src="currentImage.url"
          class="max-w-full max-h-[85vh] object-contain rounded-lg shadow-2xl border border-zinc-800"
        />
        <div class="mt-3 flex items-center gap-3">
          <Button size="sm" variant="outline" class="border-zinc-700 bg-zinc-900 text-white" @click="downloadImage(currentImage)">
            <Download class="h-4 w-4 mr-1.5" /> Download
          </Button>
          <Button size="sm" variant="outline" class="border-zinc-700 bg-zinc-900 text-white" @click="copyImageToClipboard(currentImage)">
            <Copy class="h-4 w-4 mr-1.5" /> Copy Image
          </Button>
        </div>
      </div>
    </div>

    <!-- API Code Modal -->
    <div
      v-if="showCodeModal"
      class="fixed inset-0 z-50 bg-black/80 backdrop-blur-sm flex items-center justify-center p-4"
      @click.self="showCodeModal = false"
    >
      <Card class="w-full max-w-2xl bg-zinc-900 border-zinc-800 p-6 space-y-4 shadow-2xl">
        <div class="flex items-center justify-between pb-3 border-b border-zinc-800">
          <div class="flex items-center gap-2">
            <Code2 class="h-5 w-5 text-blue-400" />
            <h3 class="text-base font-semibold text-white">API Integration Reference</h3>
          </div>
          <button class="text-zinc-400 hover:text-white" @click="showCodeModal = false">
            <X class="h-5 w-5" />
          </button>
        </div>

        <div class="space-y-4">
          <div>
            <div class="flex items-center justify-between text-xs text-zinc-400 mb-1.5">
              <span>cURL (OpenAI Compatible)</span>
              <button class="text-blue-400 hover:text-blue-300 font-mono text-[11px]" @click="copyCode(curlSnippet)">
                {{ copiedCurl ? 'Copied!' : 'Copy cURL' }}
              </button>
            </div>
            <pre class="p-3 rounded-lg bg-zinc-950 border border-zinc-800 text-xs font-mono text-zinc-300 overflow-x-auto">{{ curlSnippet }}</pre>
          </div>

          <div>
            <div class="flex items-center justify-between text-xs text-zinc-400 mb-1.5">
              <span>Python (OpenAI SDK)</span>
              <button class="text-blue-400 hover:text-blue-300 font-mono text-[11px]" @click="copyCode(pythonSnippet)">
                Copy Python
              </button>
            </div>
            <pre class="p-3 rounded-lg bg-zinc-950 border border-zinc-800 text-xs font-mono text-zinc-300 overflow-x-auto">{{ pythonSnippet }}</pre>
          </div>
        </div>
      </Card>
    </div>
  </div>
</template>
