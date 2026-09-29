<script setup lang="ts">
import { ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import Sidebar from './Sidebar.vue'
import Header from './Header.vue'

const route = useRoute()
const mainRef = ref<HTMLElement | null>(null)
const isMobileSidebarOpen = ref(false)

function toggleMobileSidebar() {
  isMobileSidebarOpen.value = !isMobileSidebarOpen.value
}

function closeMobileSidebar() {
  isMobileSidebarOpen.value = false
}

watch(
  () => route.fullPath,
  () => {
    isMobileSidebarOpen.value = false
    if (mainRef.value) {
      mainRef.value.scrollTop = 0
    }
  }
)
</script>

<template>
  <div class="flex h-screen w-screen overflow-hidden bg-[#0b0c0e]">
    <!-- Mobile Backdrop -->
    <div
      v-if="isMobileSidebarOpen"
      class="fixed inset-0 z-40 bg-black/70 backdrop-blur-xs md:hidden transition-opacity"
      @click="closeMobileSidebar"
    />

    <!-- Sidebar Drawer -->
    <Sidebar
      :is-mobile-open="isMobileSidebarOpen"
      @close="closeMobileSidebar"
    />

    <!-- Main Content Area -->
    <div class="flex-1 flex flex-col min-w-0 overflow-hidden relative">
      <Header
        :is-mobile-sidebar-open="isMobileSidebarOpen"
        @toggle-sidebar="toggleMobileSidebar"
      />
      <main ref="mainRef" class="flex-1 overflow-y-auto p-4 sm:p-6 md:p-8 ambient-glow">
        <div class="w-full space-y-6 sm:space-y-8 pb-12">
          <slot />
        </div>
      </main>
    </div>
  </div>
</template>
