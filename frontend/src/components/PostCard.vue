<script setup lang="ts">
import { ref, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import AppAvatar from './AppAvatar.vue'
import { postsApi } from '@/api/posts'
import { formatRelative } from '@/lib/format'
import { profileRoute } from '@/lib/profileUrl'
import { useAuthStore } from '@/stores/auth'
import type { Post } from '@/types/api'

const props = defineProps<{ post: Post }>()

const auth = useAuthStore()
const router = useRouter()

const likes = ref(props.post.likes)
const liked = ref(props.post.liked_by_me)
const pending = ref(false)

// The list can be replaced in place (pagination, refetch); keep the local
// like state in step with the incoming record.
watch(
  () => props.post,
  (post) => {
    likes.value = post.likes
    liked.value = post.liked_by_me
  },
)

async function toggleLike(): Promise<void> {
  if (!auth.isAuthenticated) {
    void router.push({ name: 'login', query: { redirect: router.currentRoute.value.fullPath } })
    return
  }
  if (pending.value) {
    return
  }

  pending.value = true
  try {
    if (liked.value) {
      await postsApi.unlike(props.post.id)
      liked.value = false
      likes.value = Math.max(0, likes.value - 1)
    } else {
      await postsApi.like(props.post.id)
      liked.value = true
      likes.value += 1
    }
  } catch {
    // Leave the optimistic state as-is; the next fetch will correct it.
  } finally {
    pending.value = false
  }
}
</script>

<template>
  <article class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
    <div class="flex items-start justify-between gap-4">
      <div class="min-w-0">
        <div class="mb-1 flex flex-wrap items-center gap-2">
          <span
            v-if="post.pinned"
            class="rounded-full bg-amber-100 px-2 py-0.5 text-xs font-semibold text-amber-700"
          >
            Pinned
          </span>
          <span
            v-if="post.locked"
            class="rounded-full bg-slate-100 px-2 py-0.5 text-xs font-semibold text-slate-600"
          >
            Locked
          </span>
          <RouterLink
            :to="{ name: 'channel', params: { id: post.channel_id } }"
            class="text-xs font-medium text-indigo-600 hover:text-indigo-700"
          >
            #{{ post.channel_name || post.channel_id }}
          </RouterLink>
        </div>

        <h2 class="truncate text-lg font-semibold text-slate-900">
          <RouterLink :to="{ name: 'post', params: { id: post.id } }" class="hover:text-indigo-700">
            {{ post.title }}
          </RouterLink>
        </h2>

        <p class="mt-2 whitespace-pre-wrap text-sm leading-relaxed text-slate-600">
          {{ post.content.length > 280 ? post.content.slice(0, 280) + '…' : post.content }}
        </p>
      </div>
    </div>

    <div class="mt-4 flex items-center justify-between border-t border-slate-100 pt-3">
      <RouterLink
        :to="profileRoute(post.user)"
        class="flex items-center gap-2 text-sm text-slate-600 hover:text-slate-900"
      >
        <AppAvatar :name="post.user.name" :avatar-url="post.user.avatar_url" size="sm" />
        <span class="font-medium">{{ post.user.name }}</span>
        <span class="text-slate-400">·</span>
        <span class="text-slate-500">{{ formatRelative(post.created_at) }}</span>
      </RouterLink>

      <button
        type="button"
        class="inline-flex items-center gap-1.5 rounded-lg border px-2.5 py-1 text-sm font-medium transition"
        :class="
          liked
            ? 'border-rose-200 bg-rose-50 text-rose-600'
            : 'border-slate-200 text-slate-600 hover:bg-slate-50'
        "
        :disabled="pending"
        @click="toggleLike"
      >
        <svg viewBox="0 0 20 20" fill="currentColor" class="h-4 w-4" aria-hidden="true">
          <path
            d="M10 17.25 2.9 10.6a4.5 4.5 0 1 1 6.36-6.36L10 4.94l.74-.7a4.5 4.5 0 1 1 6.36 6.36z"
          />
        </svg>
        <span>{{ likes }}</span>
      </button>
    </div>
  </article>
</template>
