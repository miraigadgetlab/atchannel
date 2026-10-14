<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'

import AppPagination from '@/components/AppPagination.vue'
import AppSpinner from '@/components/AppSpinner.vue'
import EmptyState from '@/components/EmptyState.vue'
import { channelsApi } from '@/api/channels'
import { errorMessage } from '@/lib/errors'
import { formatRelative } from '@/lib/format'
import { useAuthStore } from '@/stores/auth'
import type { Channel } from '@/types/api'

const LIMIT = 20

const auth = useAuthStore()
const route = useRoute()
const router = useRouter()

const channels = ref<Channel[]>([])
const total = ref(0)
const offset = ref(0)
const loading = ref(true)
const error = ref('')
const search = ref(typeof route.query.q === 'string' ? route.query.q : '')

const showForm = ref(false)
const form = ref({ name: '', title: '', description: '' })
const creating = ref(false)
const createError = ref('')
async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const q = typeof route.query.q === 'string' ? route.query.q : undefined
    const data = await channelsApi.list({ q, limit: LIMIT, offset: offset.value })
    channels.value = data.items ?? []
    total.value = data.total
  } catch (err) {
    error.value = errorMessage(err)
    channels.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function applySearch(): void {
  void router.push({ name: 'channels', query: search.value ? { q: search.value } : {} })
}

function changeOffset(next: number): void {
  offset.value = next
  void load()
}

async function createChannel(): Promise<void> {
  createError.value = ''
  creating.value = true
  try {
    const channel = await channelsApi.create(form.value)
    await router.push({ name: 'channel', params: { id: channel.id } })
  } catch (err) {
    createError.value = errorMessage(err)
  } finally {
    creating.value = false
  }
}

watch(
  () => route.query,
  () => {
    offset.value = 0
    search.value = typeof route.query.q === 'string' ? route.query.q : ''
    void load()
  },
)

onMounted(load)
</script>

<template>
  <div>
    <div class="mb-6 flex flex-wrap items-center gap-2">
      <input
        v-model="search"
        type="search"
        placeholder="Search channels…"
        class="w-44 rounded-lg border border-slate-300 px-3 py-1.5 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
        @keyup.enter="applySearch"
      />
      <button
        v-if="auth.isAdmin"
        type="button"
        class="rounded-lg bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white hover:bg-indigo-700"
        @click="showForm = !showForm"
      >
        {{ showForm ? 'Cancel' : 'New channel' }}
      </button>
    </div>

    <form
      v-if="showForm"
      class="mb-6 space-y-3 rounded-xl border border-slate-200 bg-white p-5"
      @submit.prevent="createChannel"
    >
      <h2 class="text-sm font-semibold text-slate-800">Create a channel</h2>
      <div class="grid gap-3 sm:grid-cols-2">
        <input
          v-model="form.name"
          type="text"
          placeholder="name (2-32 chars, e.g. go-lang)"
          required
          class="rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
        />
        <input
          v-model="form.title"
          type="text"
          placeholder="Title"
          required
          class="rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
        />
      </div>
      <textarea
        v-model="form.description"
        rows="2"
        placeholder="Description (optional)"
        class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
      />
      <p v-if="createError" class="text-sm text-rose-600">{{ createError }}</p>
      <button
        type="submit"
        class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
        :disabled="creating"
      >
        {{ creating ? 'Creating…' : 'Create' }}
      </button>
    </form>

    <AppSpinner v-if="loading" />
    <p v-else-if="error" class="rounded-lg bg-rose-50 px-4 py-3 text-sm text-rose-700">
      {{ error }}
    </p>
    <EmptyState v-else-if="channels.length === 0" title="No channels found" />

    <div v-else class="space-y-3">
      <RouterLink
        v-for="channel in channels"
        :key="channel.id"
        :to="{ name: 'channel', params: { id: channel.id } }"
        class="block rounded-xl border border-slate-200 bg-white p-5 transition hover:border-indigo-300 hover:shadow-sm"
      >
        <div class="flex items-center justify-between gap-4">
          <h2 class="font-semibold text-slate-900">{{ channel.title }}</h2>
          <span class="shrink-0 text-xs text-slate-400">#{{ channel.name }}</span>
        </div>
        <p v-if="channel.description" class="mt-1 line-clamp-2 text-sm text-slate-600">
          {{ channel.description }}
        </p>
        <p class="mt-2 text-xs text-slate-400">Created {{ formatRelative(channel.created_at) }}</p>
      </RouterLink>

      <AppPagination :total="total" :limit="LIMIT" :offset="offset" @update="changeOffset" />
    </div>
  </div>
</template>
