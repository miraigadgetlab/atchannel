<script setup lang="ts">
withDefaults(defineProps<{ label?: string }>(), { label: 'Loading…' })
</script>

<template>
  <div class="flex items-center justify-center gap-3 py-10 text-slate-500">
    <span
      class="h-5 w-5 animate-spin rounded-full border-2 border-slate-300 border-t-indigo-600"
      aria-hidden="true"
    />
    <span class="text-sm">{{ label }}</span>
  </div>
</template>
