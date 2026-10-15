<script setup lang="ts">
import { ref } from 'vue'

import AppSpinner from '@/components/AppSpinner.vue'
import HomeView from '@/views/HomeView.vue'
import NotFoundView from '@/views/NotFoundView.vue'
import ProfileView from '@/views/ProfileView.vue'
import { usersApi } from '@/api/users'
import { errorMessage, isNotFound } from '@/lib/errors'
import { userSubdomain } from '@/lib/subdomain'

// The root path means two things depending on the host: on the apex it is the
// feed, but on a user subdomain (kosero.atchannel.top) it is that account's
// public profile. Resolving it here — instead of redirecting to /users/:id —
// keeps the address bar on the subdomain root, which is the URL people share.
const sub = userSubdomain()
const profileId = ref<string | null>(null)
const loading = ref(Boolean(sub))
const notFound = ref(false)
const error = ref('')

async function resolveSubdomain(): Promise<void> {
  if (!sub) {
    return
  }
  try {
    const profile = await usersApi.getProfileByName(sub)
    profileId.value = String(profile.id)
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

void resolveSubdomain()
</script>

<template>
  <AppSpinner v-if="loading" />
  <ProfileView v-else-if="profileId" :id="profileId" />
  <NotFoundView v-else-if="notFound" />
  <p v-else-if="error" class="bg-rose-50 px-4 py-3 text-sm text-rose-700">{{ error }}</p>
  <HomeView v-else />
</template>
