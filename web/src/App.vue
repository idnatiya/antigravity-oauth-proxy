<script setup lang="ts">
import { useRoute, RouterView } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'

const route = useRoute()
</script>

<template>
  <RouterView v-if="route.meta.guestOnly" />
  <AppLayout v-else>
    <RouterView />
  </AppLayout>
</template>
