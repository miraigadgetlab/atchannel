<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import { errorMessage } from '@/lib/errors'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
const router = useRouter()

const name = ref('')
const password = ref('')
const note = ref('')
const error = ref('')
const loading = ref(false)

async function submit(): Promise<void> {
  error.value = ''
  loading.value = true
  try {
    await auth.register(name.value, password.value, note.value)
    await router.push({ name: 'home' })
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="mx-auto max-w-md">
    <h1 class="mb-1 text-2xl font-bold text-slate-900">Create an account</h1>
    <p class="mb-6 text-sm text-slate-500">Join the discussion in a minute.</p>

    <form class="space-y-4" @submit.prevent="submit">
      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700" for="name">Name</label>
        <input
          id="name"
          v-model="name"
          type="text"
          autocomplete="nickname"
          required
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
        />
      </div>

      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700" for="password">Password</label>
        <input
          id="password"
          v-model="password"
          type="password"
          autocomplete="new-password"
          required
          minlength="8"
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
        />
        <p class="mt-1 text-xs text-slate-500">At least 8 characters.</p>
      </div>

      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700" for="note">
          Note for the admins (optional)
        </label>
        <textarea
          id="note"
          v-model="note"
          rows="3"
          maxlength="500"
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
        />
        <p class="mt-1 text-xs text-slate-500">
          Only you and the admins can read this. If you ever lose access, it helps us recognise you.
        </p>
      </div>

      <p v-if="error" class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">{{ error }}</p>

      <button
        type="submit"
        class="w-full rounded-lg bg-indigo-600 px-3 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
        :disabled="loading"
      >
        {{ loading ? 'Creating account…' : 'Create account' }}
      </button>
    </form>

    <p class="mt-6 text-sm text-slate-500">
      Already have an account?
      <RouterLink :to="{ name: 'login' }" class="font-medium text-indigo-600 hover:text-indigo-700">
        Sign in
      </RouterLink>
    </p>
  </div>
</template>
