<script setup lang="ts">
withDefaults(
  defineProps<{
    title?: string
    description?: string
  }>(),
  { title: 'Nothing here yet', description: undefined },
)
</script>

<template>
  <div class="rounded-xl border border-dashed border-slate-300 bg-white px-6 py-12 text-center">
    <p class="text-base font-medium text-slate-700">{{ title }}</p>
    <p v-if="description" class="mt-1 text-sm text-slate-500">{{ description }}</p>
    <slot />
  </div>
</template>
