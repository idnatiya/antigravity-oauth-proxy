<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import Sidebar from './Sidebar.vue'
import Header from './Header.vue'

const route = useRoute()
const mainRef = ref<HTMLElement | null>(null)

watch(
  () => route.fullPath,
  () => {
    if (mainRef.value) {
      mainRef.value.scrollTop = 0
    }
  }
)
</script>

<template>
  <div class="flex h-screen w-screen overflow-hidden bg-[#0b0c0e]">
    <Sidebar />
    <div class="flex-1 flex flex-col min-w-0 overflow-hidden relative">
      <Header />
      <main ref="mainRef" class="flex-1 overflow-y-auto p-6 md:p-8 ambient-glow">
        <div class="max-w-7xl mx-auto space-y-8 pb-12">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>
