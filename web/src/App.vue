<script setup lang="ts">
import { RouterView } from 'vue-router'
import AppLayout from '@/components/layout/AppLayout.vue'
</script>

<template>
  <RouterView v-slot="{ Component, route }">
    <template v-if="Component">
      <component
        :is="Component"
        v-if="route.meta.guestOnly"
        :key="route.fullPath"
      />
      <AppLayout v-else>
        <component
          :is="Component"
          :key="route.fullPath"
        />
      </AppLayout>
    </template>
  </RouterView>
</template>
