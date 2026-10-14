<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import AppAvatar from '@/components/AppAvatar.vue'
import AppPagination from '@/components/AppPagination.vue'
import AppSpinner from '@/components/AppSpinner.vue'
import EmptyState from '@/components/EmptyState.vue'
import { commentsApi } from '@/api/comments'
import { postsApi } from '@/api/posts'
import { errorMessage, isNotFound } from '@/lib/errors'
import { formatRelative } from '@/lib/format'
import { useAuthStore } from '@/stores/auth'
import type { Comment, Post } from '@/types/api'

const props = defineProps<{ id: string }>()

const LIMIT = 20
const auth = useAuthStore()
const router = useRouter()

const post = ref<Post | null>(null)
const comments = ref<Comment[]>([])
const commentTotal = ref(0)
const offset = ref(0)

const loading = ref(true)
const notFound = ref(false)
const error = ref('')

const likes = ref(0)
const liked = ref(false)
const likePending = ref(false)

const newComment = ref('')
const posting = ref(false)
const commentError = ref('')

const canManage = computed(() => {
  if (!post.value || !auth.user) {
    return false
  }
  return auth.canModerate || post.value.user_id === auth.user.id
})

const canComment = computed(() => {
  if (!post.value || !auth.isAuthenticated) {
    return false
  }
  // A locked thread stays open to its author and moderators only.
  return !post.value.locked || auth.canModerate || post.value.user_id === auth.user?.id
})

async function loadPost(): Promise<void> {
  const data = await postsApi.get(props.id)
  post.value = data
  likes.value = data.likes
  liked.value = data.liked_by_me
}

async function loadComments(): Promise<void> {
  const data = await commentsApi.list(props.id, { limit: LIMIT, offset: offset.value })
  comments.value = data.items ?? []
  commentTotal.value = data.total
}

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  notFound.value = false
  try {
    await Promise.all([loadPost(), loadComments()])
  } catch (err) {
    if (isNotFound(err)) {
      notFound.value = true
    } else {
      error.value = errorMessage(err)
    }
  } finally {
    loading.value = false
  }
}

async function toggleLike(): Promise<void> {
  if (!auth.isAuthenticated) {
    void router.push({ name: 'login', query: { redirect: router.currentRoute.value.fullPath } })
    return
  }
  if (likePending.value) {
    return
  }

  likePending.value = true
  try {
    if (liked.value) {
      await postsApi.unlike(props.id)
      liked.value = false
      likes.value = Math.max(0, likes.value - 1)
    } else {
      await postsApi.like(props.id)
      liked.value = true
      likes.value += 1
    }
  } catch {
    // Optimistic state stands; a reload would correct it.
  } finally {
    likePending.value = false
  }
}

async function submitComment(): Promise<void> {
  commentError.value = ''
  if (!newComment.value.trim()) {
    return
  }

  posting.value = true
  try {
    await commentsApi.create(props.id, newComment.value)
    newComment.value = ''
    offset.value = 0
    await loadComments()
  } catch (err) {
    commentError.value = errorMessage(err)
  } finally {
    posting.value = false
  }
}

async function removeComment(id: number): Promise<void> {
  try {
    await commentsApi.remove(id)
    await loadComments()
  } catch (err) {
    commentError.value = errorMessage(err)
  }
}

function changeOffset(next: number): void {
  offset.value = next
  void loadComments()
}

async function removePost(): Promise<void> {
  if (!window.confirm('Delete this post and all of its comments?')) {
    return
  }
  try {
    await postsApi.remove(props.id)
    await router.push({ name: 'home' })
  } catch (err) {
    error.value = errorMessage(err)
  }
}

// Moderator actions mutate the post in place so the badges update immediately.
async function togglePin(): Promise<void> {
  if (!post.value) return
  if (post.value.pinned) {
    await postsApi.unpin(props.id)
    post.value.pinned = false
  } else {
    await postsApi.pin(props.id)
    post.value.pinned = true
  }
}

async function toggleLock(): Promise<void> {
  if (!post.value) return
  if (post.value.locked) {
    await postsApi.unlock(props.id)
    post.value.locked = false
  } else {
    await postsApi.lock(props.id)
    post.value.locked = true
  }
}

watch(() => props.id, load)
onMounted(load)
</script>

<template>
  <div class="mx-auto max-w-3xl">
    <AppSpinner v-if="loading" />

    <div v-else-if="notFound" class="text-center">
      <p class="text-lg font-semibold text-slate-800">Post not found</p>
      <RouterLink :to="{ name: 'home' }" class="mt-2 inline-block text-sm text-indigo-600">
        Back to the feed
      </RouterLink>
    </div>

    <p v-else-if="error" class="rounded-lg bg-rose-50 px-4 py-3 text-sm text-rose-700">
      {{ error }}
    </p>

    <template v-else-if="post">
      <article class="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
        <div class="mb-2 flex flex-wrap items-center gap-2">
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

        <h1 class="text-2xl font-bold text-slate-900">{{ post.title }}</h1>

        <div class="mt-3 flex items-center gap-2 text-sm text-slate-500">
          <RouterLink
            :to="{ name: 'profile', params: { id: post.user.id } }"
            class="flex items-center gap-2 hover:text-slate-900"
          >
            <AppAvatar :name="post.user.name" :avatar-url="post.user.avatar_url" size="sm" />
            <span class="font-medium text-slate-700">{{ post.user.name }}</span>
          </RouterLink>
          <span>·</span>
          <span>{{ formatRelative(post.created_at) }}</span>
        </div>

        <div class="mt-5 whitespace-pre-wrap text-sm leading-relaxed text-slate-700">
          {{ post.content }}
        </div>

        <div class="mt-6 flex flex-wrap items-center gap-2 border-t border-slate-100 pt-4">
          <button
            type="button"
            class="inline-flex items-center gap-1.5 rounded-lg border px-3 py-1.5 text-sm font-medium transition"
            :class="
              liked
                ? 'border-rose-200 bg-rose-50 text-rose-600'
                : 'border-slate-200 text-slate-600 hover:bg-slate-50'
            "
            :disabled="likePending"
            @click="toggleLike"
          >
            <svg viewBox="0 0 20 20" fill="currentColor" class="h-4 w-4" aria-hidden="true">
              <path
                d="M10 17.25 2.9 10.6a4.5 4.5 0 1 1 6.36-6.36L10 4.94l.74-.7a4.5 4.5 0 1 1 6.36 6.36z"
              />
            </svg>
            {{ likes }} {{ liked ? 'Liked' : 'Like' }}
          </button>

          <template v-if="canManage">
            <button
              type="button"
              class="rounded-lg border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-700 hover:bg-slate-50"
              @click="router.push({ name: 'post-edit', params: { id: post.id } })"
            >
              Edit
            </button>
            <button
              type="button"
              class="rounded-lg border border-rose-200 px-3 py-1.5 text-sm font-medium text-rose-600 hover:bg-rose-50"
              @click="removePost"
            >
              Delete
            </button>
          </template>

          <template v-if="auth.canModerate">
            <button
              type="button"
              class="rounded-lg border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-700 hover:bg-slate-50"
              @click="togglePin"
            >
              {{ post.pinned ? 'Unpin' : 'Pin' }}
            </button>
            <button
              type="button"
              class="rounded-lg border border-slate-200 px-3 py-1.5 text-sm font-medium text-slate-700 hover:bg-slate-50"
              @click="toggleLock"
            >
              {{ post.locked ? 'Unlock' : 'Lock' }}
            </button>
          </template>
        </div>
      </article>

      <section class="mt-8">
        <h2 class="mb-4 text-lg font-semibold text-slate-900">
          Comments <span class="text-sm font-normal text-slate-400">({{ commentTotal }})</span>
        </h2>

        <form v-if="canComment" class="mb-6" @submit.prevent="submitComment">
          <textarea
            v-model="newComment"
            rows="3"
            placeholder="Add a comment…"
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
          />
          <div class="mt-2 flex items-center justify-between">
            <p v-if="commentError" class="text-sm text-rose-600">{{ commentError }}</p>
            <span v-else />
            <button
              type="submit"
              class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
              :disabled="posting"
            >
              {{ posting ? 'Posting…' : 'Comment' }}
            </button>
          </div>
        </form>

        <p
          v-else-if="post.locked"
          class="mb-6 rounded-lg bg-slate-100 px-4 py-3 text-sm text-slate-600"
        >
          This thread is locked. New comments are disabled.
        </p>

        <p
          v-else-if="!auth.isAuthenticated"
          class="mb-6 rounded-lg bg-slate-100 px-4 py-3 text-sm text-slate-600"
        >
          <RouterLink
            :to="{ name: 'login', query: { redirect: router.currentRoute.value.fullPath } }"
            class="font-medium text-indigo-600"
          >
            Sign in
          </RouterLink>
          to join the conversation.
        </p>

        <EmptyState v-if="comments.length === 0" title="No comments yet" />

        <div v-else class="space-y-4">
          <div
            v-for="comment in comments"
            :key="comment.id"
            class="rounded-xl border border-slate-200 bg-white p-4"
          >
            <div class="flex items-start gap-3">
              <RouterLink :to="{ name: 'profile', params: { id: comment.user.id } }">
                <AppAvatar
                  :name="comment.user.name"
                  :avatar-url="comment.user.avatar_url"
                  size="sm"
                />
              </RouterLink>
              <div class="min-w-0 flex-1">
                <div class="flex items-center gap-2 text-sm">
                  <RouterLink
                    :to="{ name: 'profile', params: { id: comment.user.id } }"
                    class="font-medium text-slate-800 hover:text-indigo-700"
                  >
                    {{ comment.user.name }}
                  </RouterLink>
                  <span class="text-slate-400">·</span>
                  <span class="text-slate-400">{{ formatRelative(comment.created_at) }}</span>
                </div>
                <p class="mt-1 whitespace-pre-wrap text-sm text-slate-700">{{ comment.content }}</p>
              </div>
              <button
                v-if="
                  auth.isAuthenticated && (auth.canModerate || auth.user?.id === comment.user_id)
                "
                type="button"
                class="text-xs font-medium text-rose-600 hover:text-rose-700"
                @click="removeComment(comment.id)"
              >
                Delete
              </button>
            </div>
          </div>

          <AppPagination
            :total="commentTotal"
            :limit="LIMIT"
            :offset="offset"
            @update="changeOffset"
          />
        </div>
      </section>
    </template>
  </div>
</template>
