<script setup lang="ts">
import { computed, type HTMLAttributes } from 'vue'
import { TabsList, type TabsListProps } from 'radix-vue'
import { cn } from '@/lib/utils'

const props = defineProps<TabsListProps & { class?: HTMLAttributes['class'] }>()

const delegatedProps = computed(() => {
  const { class: _, ...delegated } = props
  return delegated
})
</script>

<template>
  <TabsList
    v-bind="delegatedProps"
    :class="
      cn(
        'inline-flex items-center justify-center rounded-xl bg-[#202227] p-1 border border-[#2c2e36] text-zinc-400',
        props.class
      )
    "
  >
    <slot />
  </TabsList>
</template>
