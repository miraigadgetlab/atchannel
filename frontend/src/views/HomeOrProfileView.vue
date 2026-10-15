<script setup lang="ts">
import { computed, ref, watch } from 'vue'

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
//
// The same view backs the pretty profile route /:name, where the account comes
// from the route instead of the host; when neither is present this is the feed.
const props = defineProps<{ name?: string }>()

const sub = userSubdomain()
const account = computed(() => props.name ?? sub)
const profileId = ref<string | null>(null)
const loading = ref(false)
const notFound = ref(false)
const error = ref('')

async function resolve(): Promise<void> {
  const name = account.value
  profileId.value = null
  notFound.value = false
  error.value = ''

  if (!name) {
    loading.value = false
    return
  }

  loading.value = true
  try {
    const profile = await usersApi.getProfileByName(name)
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

// The component instance is reused when moving between /, /emirhan and another
// profile, so the account has to be watched rather than merely read once.
watch(account, resolve, { immediate: true })
</script>

<template>
  <AppSpinner v-if="loading" />
  <ProfileView v-else-if="profileId" :id="profileId" />
  <NotFoundView v-else-if="notFound" />
  <p v-else-if="error" class="bg-rose-50 px-4 py-3 text-sm text-rose-700">{{ error }}</p>
  <HomeView v-else />
</template>
