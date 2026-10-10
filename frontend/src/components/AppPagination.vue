<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{
  total: number
  limit: number
  offset: number
}>()

const emit = defineEmits<{ (e: 'update', offset: number): void }>()

const hasPrevious = computed(() => props.offset > 0)
const hasNext = computed(() => props.offset + props.limit < props.total)
const rangeStart = computed(() => (props.total === 0 ? 0 : props.offset + 1))
const rangeEnd = computed(() => Math.min(props.offset + props.limit, props.total))

function go(delta: number): void {
  const next = Math.max(0, props.offset + delta * props.limit)
  if (next !== props.offset) {
    emit('update', next)
  }
}
</script>

<template>
  <div
    v-if="total > limit"
    class="flex flex-col items-center justify-between gap-3 pt-4 sm:flex-row"
  >
    <p class="text-sm text-slate-500">
      Showing <span class="font-medium text-slate-700">{{ rangeStart }}–{{ rangeEnd }}</span> of
      <span class="font-medium text-slate-700">{{ total }}</span>
    </p>
    <div class="flex gap-2">
      <button
        type="button"
        class="rounded-lg border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700 transition hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-40"
        :disabled="!hasPrevious"
        @click="go(-1)"
      >
        Previous
      </button>
      <button
        type="button"
        class="rounded-lg border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700 transition hover:bg-slate-100 disabled:cursor-not-allowed disabled:opacity-40"
        :disabled="!hasNext"
        @click="go(1)"
      >
        Next
      </button>
    </div>
  </div>
</template>
