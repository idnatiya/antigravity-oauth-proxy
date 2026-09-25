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
  <div class="flex h-screen w-screen overflow-hidden bg-[#141518]">
    <Sidebar />
    <div class="flex-1 flex flex-col min-w-0 overflow-hidden">
      <Header />
      <main ref="mainRef" class="flex-1 overflow-y-auto p-8 bg-[#141518]">
        <div class="max-w-7xl mx-auto space-y-8">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>
