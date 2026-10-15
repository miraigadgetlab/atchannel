<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AppPagination from '@/components/AppPagination.vue'
import AppSpinner from '@/components/AppSpinner.vue'
import EmptyState from '@/components/EmptyState.vue'
import PostCard from '@/components/PostCard.vue'
import { postsApi } from '@/api/posts'
import { errorMessage } from '@/lib/errors'
import type { Post, PostSort } from '@/types/api'

const route = useRoute()
const router = useRouter()

const LIMIT = 20

const posts = ref<Post[]>([])
const total = ref(0)
const offset = ref(0)
const loading = ref(true)
const error = ref('')

const sort = ref<PostSort>((route.query.sort as PostSort) ?? 'new')
const search = ref(typeof route.query.q === 'string' ? route.query.q : '')

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const q = typeof route.query.q === 'string' ? route.query.q : undefined
    const data = await postsApi.list({ q, sort: sort.value, limit: LIMIT, offset: offset.value })
    posts.value = data.items ?? []
    total.value = data.total
  } catch (err) {
    error.value = errorMessage(err)
    posts.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

function applyFilters(): void {
  void router.push({
    name: 'home',
    query: {
      ...(search.value ? { q: search.value } : {}),
      ...(sort.value !== 'new' ? { sort: sort.value } : {}),
    },
  })
}

function changeOffset(next: number): void {
  offset.value = next
  void load()
}

// The URL is the source of truth for q/sort so a filtered feed is shareable
// and the browser back button works.
watch(
  () => route.query,
  () => {
    offset.value = 0
    search.value = typeof route.query.q === 'string' ? route.query.q : ''
    sort.value = (route.query.sort as PostSort) ?? 'new'
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
        placeholder="Search…"
        class="w-40 rounded-lg border border-slate-300 px-3 py-1.5 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
        @keyup.enter="applyFilters"
      />
      <select
        v-model="sort"
        class="rounded-lg border border-slate-300 px-2 py-1.5 text-sm focus:border-indigo-500 focus:outline-none"
        @change="applyFilters"
      >
        <option value="new">Newest</option>
        <option value="old">Oldest</option>
        <option value="top">Most liked</option>
      </select>
    </div>

    <AppSpinner v-if="loading" />

    <div
      v-else-if="error"
      class="rounded-xl border border-rose-200 bg-rose-50 p-4 text-sm text-rose-700"
    >
      {{ error }}
    </div>

    <EmptyState
      v-else-if="posts.length === 0"
      title="No posts found"
      description="Try a different search or start a new discussion."
    />

    <div v-else class="space-y-4">
      <PostCard v-for="post in posts" :key="post.id" :post="post" />
      <AppPagination :total="total" :limit="LIMIT" :offset="offset" @update="changeOffset" />
    </div>
  </div>
</template>
