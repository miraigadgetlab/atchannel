<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'

import AppAvatar from '@/components/AppAvatar.vue'
import AppPagination from '@/components/AppPagination.vue'
import AppSpinner from '@/components/AppSpinner.vue'
import EmptyState from '@/components/EmptyState.vue'
import { notificationsApi } from '@/api/notifications'
import { errorMessage } from '@/lib/errors'
import { formatRelative } from '@/lib/format'
import { useNotificationsStore } from '@/stores/notifications'
import type { Notification } from '@/types/api'

const LIMIT = 20

const store = useNotificationsStore()
const items = ref<Notification[]>([])
const total = ref(0)
const offset = ref(0)
const loading = ref(true)
const error = ref('')

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const data = await notificationsApi.list({ limit: LIMIT, offset: offset.value })
    items.value = data.items ?? []
    total.value = data.total
  } catch (err) {
    error.value = errorMessage(err)
    items.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function changeOffset(next: number): void {
  offset.value = next
  void load()
}

async function markAllRead(): Promise<void> {
  try {
    await store.markAllRead()
  } catch {
    // Non-fatal: the list still renders, the badge just stays.
  }
}

onMounted(async () => {
  await load()
  await markAllRead()
})
</script>

<template>
  <div class="mx-auto max-w-3xl">
    <div class="mb-6 flex items-center justify-between">
      <h1 class="text-2xl font-bold text-slate-900">Notifications</h1>
      <button
        type="button"
        class="text-sm font-medium text-indigo-600 hover:text-indigo-700"
        @click="markAllRead"
      >
        Mark all as read
      </button>
    </div>

    <AppSpinner v-if="loading" />
    <p v-else-if="error" class="rounded-lg bg-rose-50 px-4 py-3 text-sm text-rose-700">
      {{ error }}
    </p>
    <EmptyState v-else-if="items.length === 0" title="No notifications yet" />

    <div v-else class="space-y-2">
      <div
        v-for="item in items"
        :key="item.id"
        class="flex items-center gap-3 rounded-xl border border-slate-200 bg-white p-4"
        :class="{ 'bg-indigo-50/40': !item.read_at }"
      >
        <RouterLink :to="{ name: 'profile', params: { id: item.actor.id } }">
          <AppAvatar :name="item.actor.name" :avatar-url="item.actor.avatar_url" size="sm" />
        </RouterLink>

        <p class="flex-1 text-sm text-slate-700">
          <RouterLink
            :to="{ name: 'profile', params: { id: item.actor.id } }"
            class="font-medium text-slate-900 hover:text-indigo-700"
          >
            {{ item.actor.name }}
          </RouterLink>
          <template v-if="item.type === 'comment'"> commented on your post.</template>
          <template v-else> liked your post.</template>
        </p>

        <RouterLink
          v-if="item.post_id"
          :to="{ name: 'post', params: { id: item.post_id } }"
          class="shrink-0 text-xs font-medium text-indigo-600 hover:text-indigo-700"
        >
          View
        </RouterLink>
        <span class="shrink-0 text-xs text-slate-400">{{ formatRelative(item.created_at) }}</span>
      </div>

      <AppPagination :total="total" :limit="LIMIT" :offset="offset" @update="changeOffset" />
    </div>
  </div>
</template>
