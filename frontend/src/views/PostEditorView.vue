<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'

import AppSpinner from '@/components/AppSpinner.vue'
import { channelsApi } from '@/api/channels'
import { postsApi } from '@/api/posts'
import { errorMessage } from '@/lib/errors'
import type { Channel } from '@/types/api'

const props = defineProps<{ id?: string }>()

const route = useRoute()
const router = useRouter()

const isEdit = computed(() => Boolean(props.id))

const title = ref('')
const content = ref('')
const channelId = ref<number | null>(null)
const channels = ref<Channel[]>([])

const loading = ref(true)
const submitting = ref(false)
const error = ref('')

onMounted(async () => {
  loading.value = true
  try {
    if (props.id) {
      const post = await postsApi.get(props.id)
      title.value = post.title
      content.value = post.content
      channelId.value = post.channel_id
    } else {
      const raw = route.query.channel_id
      if (typeof raw === 'string' && raw !== '' && !Number.isNaN(Number(raw))) {
        channelId.value = Number(raw)
      }
      const data = await channelsApi.list({ limit: 100 })
      channels.value = data.items ?? []
    }
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
})

async function submit(): Promise<void> {
  error.value = ''

  if (!isEdit.value && channelId.value === null) {
    error.value = 'Please choose a channel.'
    return
  }

  submitting.value = true
  try {
    if (props.id) {
      const post = await postsApi.update(props.id, { title: title.value, content: content.value })
      await router.push({ name: 'post', params: { id: post.id } })
    } else {
      const targetChannel = channelId.value
      if (targetChannel === null) {
        return
      }
      const post = await postsApi.create({
        channel_id: targetChannel,
        title: title.value,
        content: content.value,
      })
      await router.push({ name: 'post', params: { id: post.id } })
    }
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <div class="mx-auto max-w-3xl">
    <h1 class="mb-6 text-2xl font-bold text-slate-900">
      {{ isEdit ? 'Edit post' : 'New post' }}
    </h1>

    <AppSpinner v-if="loading" />

    <form v-else class="space-y-4" @submit.prevent="submit">
      <div v-if="!isEdit">
        <label class="mb-1 block text-sm font-medium text-slate-700" for="channel">Channel</label>
        <select
          id="channel"
          v-model.number="channelId"
          required
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
        >
          <option :value="null" disabled>Choose a channel…</option>
          <option v-for="channel in channels" :key="channel.id" :value="channel.id">
            {{ channel.title }} (#{{ channel.name }})
          </option>
        </select>
      </div>

      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700" for="title">Title</label>
        <input
          id="title"
          v-model="title"
          type="text"
          required
          maxlength="200"
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
        />
      </div>

      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700" for="content">Content</label>
        <textarea
          id="content"
          v-model="content"
          rows="12"
          required
          class="w-full rounded-lg border border-slate-300 px-3 py-2 font-mono text-sm focus:border-indigo-500 focus:outline-none"
        />
      </div>

      <p v-if="error" class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">{{ error }}</p>

      <div class="flex gap-2">
        <button
          type="submit"
          class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
          :disabled="submitting"
        >
          {{ submitting ? 'Saving…' : isEdit ? 'Save changes' : 'Publish' }}
        </button>
        <button
          type="button"
          class="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-700"
          @click="router.back()"
        >
          Cancel
        </button>
      </div>
    </form>
  </div>
</template>
