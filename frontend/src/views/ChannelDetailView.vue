<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import AppPagination from '@/components/AppPagination.vue'
import AppSpinner from '@/components/AppSpinner.vue'
import EmptyState from '@/components/EmptyState.vue'
import PostCard from '@/components/PostCard.vue'
import { channelsApi } from '@/api/channels'
import { errorMessage, isNotFound } from '@/lib/errors'
import { formatDate } from '@/lib/format'
import { useAuthStore } from '@/stores/auth'
import type { ChannelDetail } from '@/types/api'

const props = defineProps<{ id: string }>()

const LIMIT = 20
const auth = useAuthStore()
const router = useRouter()

const channel = ref<ChannelDetail | null>(null)
const offset = ref(0)
const loading = ref(true)
const error = ref('')
const notFound = ref(false)

const canManage = computed(() => {
  if (!channel.value || !auth.user) {
    return false
  }
  return auth.canModerate || channel.value.created_by === auth.user.id
})

const editing = ref(false)
const editForm = ref({ name: '', title: '', description: '' })
const saving = ref(false)
const editError = ref('')

const postCount = computed(() => channel.value?.post_count ?? 0)

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  notFound.value = false
  try {
    channel.value = await channelsApi.get(props.id, { limit: LIMIT, offset: offset.value })
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

function changeOffset(next: number): void {
  offset.value = next
  void load()
}

function startEdit(): void {
  if (!channel.value) {
    return
  }
  editForm.value = {
    name: channel.value.name,
    title: channel.value.title,
    description: channel.value.description,
  }
  editing.value = true
}

async function saveEdit(): Promise<void> {
  editError.value = ''
  saving.value = true
  try {
    await channelsApi.update(props.id, editForm.value)
    editing.value = false
    await load()
  } catch (err) {
    editError.value = errorMessage(err)
  } finally {
    saving.value = false
  }
}

async function removeChannel(): Promise<void> {
  if (!window.confirm('Delete this channel and all of its posts?')) {
    return
  }
  try {
    await channelsApi.remove(props.id)
    await router.push({ name: 'channels' })
  } catch (err) {
    error.value = errorMessage(err)
  }
}

watch(
  () => props.id,
  () => {
    offset.value = 0
    void load()
  },
)

onMounted(load)
</script>

<template>
  <div>
    <AppSpinner v-if="loading" />

    <div v-else-if="notFound" class="text-center">
      <p class="text-lg font-semibold text-slate-800">Channel not found</p>
      <RouterLink :to="{ name: 'channels' }" class="mt-2 inline-block text-sm text-indigo-600">
        Back to channels
      </RouterLink>
    </div>

    <p v-else-if="error" class="rounded-lg bg-rose-50 px-4 py-3 text-sm text-rose-700">
      {{ error }}
    </p>

    <div v-else-if="channel">
      <div class="mb-6">
        <form
          v-if="editing"
          class="space-y-3 rounded-xl border border-slate-200 bg-white p-5"
          @submit.prevent="saveEdit"
        >
          <input
            v-model="editForm.name"
            type="text"
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
          />
          <input
            v-model="editForm.title"
            type="text"
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
          />
          <textarea
            v-model="editForm.description"
            rows="2"
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
          />
          <p v-if="editError" class="text-sm text-rose-600">{{ editError }}</p>
          <div class="flex gap-2">
            <button
              type="submit"
              class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
              :disabled="saving"
            >
              Save
            </button>
            <button
              type="button"
              class="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700"
              @click="editing = false"
            >
              Cancel
            </button>
          </div>
        </form>

        <div v-else>
          <div class="flex flex-wrap items-start justify-between gap-4">
            <div>
              <p class="text-xs font-medium text-slate-400">#{{ channel.name }}</p>
              <h1 class="text-2xl font-bold text-slate-900">{{ channel.title }}</h1>
              <p v-if="channel.description" class="mt-1 max-w-2xl text-sm text-slate-600">
                {{ channel.description }}
              </p>
              <p class="mt-2 text-xs text-slate-400">
                {{ postCount }} posts · created {{ formatDate(channel.created_at) }}
              </p>
            </div>

            <div class="flex gap-2">
              <RouterLink
                :to="{ name: 'post-new', query: { channel_id: String(channel.id) } }"
                class="rounded-lg bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white hover:bg-indigo-700"
              >
                New post
              </RouterLink>
              <template v-if="canManage">
                <button
                  type="button"
                  class="rounded-lg border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700 hover:bg-slate-50"
                  @click="startEdit"
                >
                  Edit
                </button>
                <button
                  type="button"
                  class="rounded-lg border border-rose-200 px-3 py-1.5 text-sm font-medium text-rose-600 hover:bg-rose-50"
                  @click="removeChannel"
                >
                  Delete
                </button>
              </template>
            </div>
          </div>
        </div>
      </div>

      <div class="space-y-3">
        <EmptyState v-if="channel.posts.length === 0" title="No posts in this channel yet" />
        <template v-else>
          <PostCard v-for="post in channel.posts" :key="post.id" :post="post" />
          <AppPagination
            :total="postCount"
            :limit="LIMIT"
            :offset="offset"
            @update="changeOffset"
          />
        </template>
      </div>
    </div>
  </div>
</template>
