<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import { authApi } from '@/api/auth'
import { errorMessage } from '@/lib/errors'

const router = useRouter()

// Recovery is code-based only. A single-use code, generated from the account
// settings (or by an admin), stands in for the old emailed reset link.
const name = ref('')
const code = ref('')
const password = ref('')
const confirm = ref('')

const error = ref('')
const loading = ref(false)

async function submit(): Promise<void> {
  error.value = ''

  if (password.value !== confirm.value) {
    error.value = 'The two passwords do not match.'
    return
  }

  loading.value = true
  try {
    await authApi.resetPasswordWithCode(name.value, code.value, password.value)
    await router.push({ name: 'login' })
  } catch (err) {
    error.value = errorMessage(err)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="mx-auto max-w-md">
    <h1 class="mb-1 text-2xl font-bold text-slate-900">Recover your account</h1>
    <p class="mb-6 text-sm text-slate-500">
      Enter your name and one of your single-use recovery codes to set a new password.
    </p>

    <form class="space-y-4" @submit.prevent="submit">
      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700" for="name">Name</label>
        <input
          id="name"
          v-model="name"
          type="text"
          autocomplete="username"
          required
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
        />
      </div>

      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700" for="code">Recovery code</label>
        <input
          id="code"
          v-model="code"
          type="text"
          autocomplete="one-time-code"
          placeholder="XXXXX-XXXXX"
          required
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
        />
      </div>

      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700" for="password">New password</label>
        <input
          id="password"
          v-model="password"
          type="password"
          autocomplete="new-password"
          required
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
        />
      </div>

      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700" for="confirm">
          Confirm new password
        </label>
        <input
          id="confirm"
          v-model="confirm"
          type="password"
          autocomplete="new-password"
          required
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
        />
      </div>

      <p class="text-xs text-slate-500">
        Each code can be used once. Setting a new password signs you out everywhere. No code? Ask an
        admin to generate one for you.
      </p>

      <p v-if="error" class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">{{ error }}</p>

      <button
        type="submit"
        class="w-full rounded-lg bg-indigo-600 px-3 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
        :disabled="loading"
      >
        {{ loading ? 'Saving…' : 'Set new password' }}
      </button>
    </form>

    <p class="mt-6 text-sm text-slate-500">
      <RouterLink :to="{ name: 'login' }" class="font-medium text-indigo-600 hover:text-indigo-700">
        Back to sign in
      </RouterLink>
    </p>
  </div>
</template>
