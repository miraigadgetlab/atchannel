<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { RouterLink, useRoute } from 'vue-router'

import { authApi } from '@/api/auth'
import { errorMessage } from '@/lib/errors'
import { useAuthStore } from '@/stores/auth'

type Status = 'working' | 'ok' | 'error' | 'missing'

const route = useRoute()
const auth = useAuthStore()

const status = ref<Status>('working')
const message = ref('')

onMounted(async () => {
  const token = typeof route.query.token === 'string' ? route.query.token : ''
  if (!token) {
    status.value = 'missing'
    return
  }

  try {
    await authApi.verifyEmail(token)
    status.value = 'ok'
    // Refresh the account so the "verify your email" banner disappears.
    if (auth.isAuthenticated) {
      await auth.fetchMe().catch(() => undefined)
    }
  } catch (err) {
    status.value = 'error'
    message.value = errorMessage(err)
  }
})
</script>

<template>
  <div class="mx-auto max-w-md text-center">
    <h1 class="mb-4 text-2xl font-bold text-slate-900">Email verification</h1>

    <p v-if="status === 'working'" class="text-sm text-slate-500">Verifying your email…</p>

    <div v-else-if="status === 'ok'" class="space-y-4">
      <p
        class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800"
      >
        Your email address has been verified. Thank you!
      </p>
      <RouterLink :to="{ name: 'home' }" class="font-medium text-indigo-600 hover:text-indigo-700">
        Go to the feed
      </RouterLink>
    </div>

    <div v-else-if="status === 'error'" class="space-y-4">
      <p class="rounded-lg border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700">
        {{ message || 'This verification link is invalid or has expired.' }}
      </p>
      <RouterLink :to="{ name: 'home' }" class="font-medium text-indigo-600 hover:text-indigo-700">
        Go to the feed
      </RouterLink>
    </div>

    <div v-else class="space-y-4">
      <p class="rounded-lg border border-amber-200 bg-amber-50 px-4 py-3 text-sm text-amber-800">
        This page needs a verification token. Open the link from your email.
      </p>
      <RouterLink :to="{ name: 'home' }" class="font-medium text-indigo-600 hover:text-indigo-700">
        Go to the feed
      </RouterLink>
    </div>
  </div>
</template>
