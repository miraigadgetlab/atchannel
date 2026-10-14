<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'

import AppAvatar from '@/components/AppAvatar.vue'
import AppPagination from '@/components/AppPagination.vue'
import AppSpinner from '@/components/AppSpinner.vue'
import EmptyState from '@/components/EmptyState.vue'
import { usersApi } from '@/api/users'
import { errorMessage } from '@/lib/errors'
import type { Channeler } from '@/types/api'

const LIMIT = 20

const channelers = ref<Channeler[]>([])
const total = ref(0)
const offset = ref(0)
const loading = ref(true)
const error = ref('')

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const data = await usersApi.listChannelers({ limit: LIMIT, offset: offset.value })
    channelers.value = data.items ?? []
    total.value = data.total
  } catch (err) {
    error.value = errorMessage(err)
    channelers.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function changeOffset(next: number): void {
  offset.value = next
  void load()
}

onMounted(load)
</script>

<template>
  <div>
    <AppSpinner v-if="loading" />
    <p v-else-if="error" class="rounded-lg bg-rose-50 px-4 py-3 text-sm text-rose-700">
      {{ error }}
    </p>
    <EmptyState
      v-else-if="channelers.length === 0"
      title="No channelers yet"
      description="Once people start posting, they will show up here."
    />

    <div v-else class="space-y-3">
      <RouterLink
        v-for="(channeler, index) in channelers"
        :key="channeler.id"
        :to="{ name: 'profile', params: { id: channeler.id } }"
        class="flex items-center gap-4 rounded-xl border border-slate-200 bg-white p-4 transition hover:border-indigo-300 hover:shadow-sm"
      >
        <span class="w-8 shrink-0 text-center text-lg font-bold text-slate-400">
          {{ offset + index + 1 }}
        </span>

        <AppAvatar :name="channeler.name" :avatar-url="channeler.avatar_url" />

        <div class="min-w-0 flex-1">
          <p class="truncate font-semibold text-slate-900">{{ channeler.name }}</p>
          <p v-if="channeler.about_me" class="truncate text-sm text-slate-500">
            {{ channeler.about_me }}
          </p>
        </div>

        <div class="shrink-0 text-right">
          <p class="text-lg font-bold text-slate-800">{{ channeler.post_count }}</p>
          <p class="text-xs text-slate-400">posts</p>
        </div>
      </RouterLink>

      <AppPagination :total="total" :limit="LIMIT" :offset="offset" @update="changeOffset" />
    </div>
  </div>
</template>
