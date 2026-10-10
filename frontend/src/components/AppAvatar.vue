<script setup lang="ts">
import { computed } from 'vue'

import { initials } from '@/lib/format'
import { mediaUrl } from '@/lib/media'

const props = withDefaults(
  defineProps<{
    name: string
    avatarUrl?: string
    size?: 'sm' | 'md' | 'lg'
  }>(),
  { avatarUrl: undefined, size: 'md' },
)

const sizeClasses = computed(() => {
  switch (props.size) {
    case 'sm':
      return 'h-8 w-8 text-xs'
    case 'lg':
      return 'h-16 w-16 text-xl'
    default:
      return 'h-10 w-10 text-sm'
  }
})

const src = computed(() => mediaUrl(props.avatarUrl))
</script>

<template>
  <img
    v-if="src"
    :src="src"
    :alt="name"
    :class="['rounded-full object-cover ring-1 ring-slate-200', sizeClasses]"
  />
  <span
    v-else
    :class="[
      'inline-flex items-center justify-center rounded-full bg-indigo-100 font-semibold text-indigo-700 ring-1 ring-indigo-200',
      sizeClasses,
    ]"
    :aria-label="name"
  >
    {{ initials(name) }}
  </span>
</template>
