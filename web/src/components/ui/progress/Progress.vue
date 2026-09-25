<script setup lang="ts">
import { computed, type HTMLAttributes } from 'vue'
import {
  ProgressIndicator,
  ProgressRoot,
  type ProgressRootProps,
} from 'radix-vue'
import { cn } from '@/lib/utils'

const props = withDefaults(
  defineProps<
    ProgressRootProps & {
      class?: HTMLAttributes['class']
      indicatorClass?: HTMLAttributes['class']
    }
  >(),
  {
    modelValue: 0,
  }
)

const delegatedProps = computed(() => {
  const { class: _, indicatorClass: __, ...delegated } = props
  return delegated
})
</script>

<template>
  <ProgressRoot
    v-bind="delegatedProps"
    :class="
      cn(
        'relative h-2 w-full overflow-hidden rounded-full bg-[#18191d]',
        props.class
      )
    "
  >
    <ProgressIndicator
      :class="
        cn(
          'h-full w-full flex-1 transition-all duration-300',
          props.indicatorClass
        )
      "
      :style="`transform: translateX(-${100 - (props.modelValue ?? 0)}%);`"
    />
  </ProgressRoot>
</template>
