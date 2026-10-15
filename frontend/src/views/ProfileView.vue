<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'

import AppAvatar from '@/components/AppAvatar.vue'
import AppSpinner from '@/components/AppSpinner.vue'
import { usersApi } from '@/api/users'
import { errorMessage, isNotFound } from '@/lib/errors'
import { formatDate } from '@/lib/format'
import { getBaseUrl, setSeo } from '@/lib/seo'
import type { PublicProfile } from '@/types/api'

const props = defineProps<{ id: string }>()

const profile = ref<PublicProfile | null>(null)
const loading = ref(true)
const notFound = ref(false)
const error = ref('')

async function load(): Promise<void> {
  loading.value = true
  notFound.value = false
  error.value = ''
  try {
    profile.value = await usersApi.getProfile(props.id)
    applySeo()
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

// Public profiles are indexable; ProfilePage JSON-LD tells search engines
// this page is a person's account, so "name atchannel" searches resolve here.
function applySeo(): void {
  const data = profile.value
  if (!data) {
    return
  }
  const url = `${getBaseUrl()}/users/${data.id}`
  const description =
    data.about_me?.trim().slice(0, 160) ||
    `${data.name} on atchannel — ${data.post_count} post${data.post_count === 1 ? '' : 's'}, ${data.comment_count} comment${data.comment_count === 1 ? '' : 's'}.`
  setSeo({
    title: data.name,
    description,
    type: 'profile',
    url,
    jsonLd: {
      '@context': 'https://schema.org',
      '@type': 'ProfilePage',
      mainEntity: {
        '@type': 'Person',
        name: data.name,
        description: data.about_me?.trim() || undefined,
        url,
      },
    },
  })
}

watch(() => props.id, load)
onMounted(load)
</script>

<template>
  <div class="mx-auto max-w-2xl">
    <AppSpinner v-if="loading" />

    <div v-else-if="notFound" class="text-center">
      <p class="text-lg font-semibold text-slate-800">User not found</p>
    </div>

    <p v-else-if="error" class="rounded-lg bg-rose-50 px-4 py-3 text-sm text-rose-700">
      {{ error }}
    </p>

    <div v-else-if="profile" class="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <div class="flex items-center gap-4">
        <AppAvatar :name="profile.name" :avatar-url="profile.avatar_url" size="lg" />
        <div>
          <h1 class="text-xl font-bold text-slate-900">{{ profile.name }}</h1>
          <p class="text-sm text-slate-500">Member since {{ formatDate(profile.created_at) }}</p>
        </div>
      </div>

      <p
        v-if="profile.about_me"
        class="mt-5 whitespace-pre-wrap text-sm leading-relaxed text-slate-700"
      >
        {{ profile.about_me }}
      </p>
      <p v-else class="mt-5 text-sm text-slate-400">This user has not written a bio yet.</p>

      <dl class="mt-6 grid grid-cols-2 gap-4 border-t border-slate-100 pt-4 text-center">
        <div>
          <dt class="text-xs font-medium tracking-wide text-slate-400 uppercase">Posts</dt>
          <dd class="text-2xl font-semibold text-slate-800">{{ profile.post_count }}</dd>
        </div>
        <div>
          <dt class="text-xs font-medium tracking-wide text-slate-400 uppercase">Comments</dt>
          <dd class="text-2xl font-semibold text-slate-800">{{ profile.comment_count }}</dd>
        </div>
      </dl>
    </div>
  </div>
</template>
