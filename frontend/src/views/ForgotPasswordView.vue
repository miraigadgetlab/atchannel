<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink } from 'vue-router'

import { authApi } from '@/api/auth'
import { errorMessage } from '@/lib/errors'

const email = ref('')
const sent = ref(false)
const error = ref('')
const loading = ref(false)

async function submit(): Promise<void> {
  error.value = ''
  loading.value = true
  try {
    // The backend always answers the same way (whether or not the address is
    // registered), so the UI must not imply anything about the result either.
    await authApi.forgotPassword(email.value)
    sent.value = true
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="mx-auto max-w-md">
    <h1 class="mb-1 text-2xl font-bold text-slate-900">Reset your password</h1>
    <p class="mb-6 text-sm text-slate-500">
      Enter the email on your account and we will send a reset link.
    </p>

    <div
      v-if="sent"
      class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800"
    >
      If an account exists for <span class="font-medium">{{ email }}</span
      >, a reset link has been sent.
    </div>

    <form v-else class="space-y-4" @submit.prevent="submit">
      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700" for="email">Email</label>
        <input
          id="email"
          v-model="email"
          type="email"
          autocomplete="email"
          required
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
        />
      </div>

      <p v-if="error" class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">{{ error }}</p>

      <button
        type="submit"
        class="w-full rounded-lg bg-indigo-600 px-3 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
        :disabled="loading"
      >
        {{ loading ? 'Sending…' : 'Send reset link' }}
      </button>
    </form>

    <p class="mt-6 text-sm text-slate-500">
      <RouterLink :to="{ name: 'login' }" class="font-medium text-indigo-600 hover:text-indigo-700">
        Back to sign in
      </RouterLink>
    </p>
  </div>
</template>
