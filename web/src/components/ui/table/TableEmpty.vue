<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'
import TableRow from './TableRow.vue'
import TableCell from './TableCell.vue'

const props = withDefaults(
  defineProps<{
    class?: HTMLAttributes['class']
    colspan?: number
  }>(),
  {
    colspan: 1,
  }
)
</script>

<template>
  <TableRow>
    <TableCell
      :class="
        cn(
          'p-8 text-center align-middle text-xs font-mono text-zinc-500 whitespace-nowrap',
          props.class
        )
      "
      :colspan="props.colspan"
    >
      <slot />
    </TableCell>
  </TableRow>
</template>
