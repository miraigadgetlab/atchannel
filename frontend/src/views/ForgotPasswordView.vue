<script setup lang="ts">
import { ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import { authApi } from '@/api/auth'
import { errorMessage } from '@/lib/errors'

const router = useRouter()

// "email" sends a reset link; "code" uses one of the account's single-use
// recovery codes, which works without a mailbox.
const mode = ref<'email' | 'code'>('email')

// Email mode
const email = ref('')
const sent = ref(false)

// Recovery-code mode
const codeEmail = ref('')
const code = ref('')
const password = ref('')
const confirm = ref('')

const error = ref('')
const loading = ref(false)

function switchMode(next: 'email' | 'code'): void {
  mode.value = next
  error.value = ''
}

async function submitEmail(): Promise<void> {
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

async function submitCode(): Promise<void> {
  error.value = ''
  if (password.value !== confirm.value) {
    error.value = 'The two passwords do not match.'
    return
  }

  loading.value = true
  try {
    await authApi.resetPasswordWithCode(codeEmail.value, code.value, password.value)
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
    <h1 class="mb-1 text-2xl font-bold text-slate-900">Reset your password</h1>
    <p class="mb-6 text-sm text-slate-500">
      Use a reset link sent by email, or one of your recovery codes.
    </p>

    <div class="mb-6 flex gap-2">
      <button
        type="button"
        class="flex-1 rounded-lg border px-3 py-2 text-sm font-medium"
        :class="
          mode === 'email'
            ? 'border-indigo-500 text-indigo-600'
            : 'border-slate-300 text-slate-600 hover:bg-slate-50'
        "
        @click="switchMode('email')"
      >
        Email link
      </button>
      <button
        type="button"
        class="flex-1 rounded-lg border px-3 py-2 text-sm font-medium"
        :class="
          mode === 'code'
            ? 'border-indigo-500 text-indigo-600'
            : 'border-slate-300 text-slate-600 hover:bg-slate-50'
        "
        @click="switchMode('code')"
      >
        Recovery code
      </button>
    </div>

    <!-- Email link -->
    <template v-if="mode === 'email'">
      <div
        v-if="sent"
        class="rounded-lg border border-emerald-200 bg-emerald-50 px-4 py-3 text-sm text-emerald-800"
      >
        If an account exists for <span class="font-medium">{{ email }}</span
        >, a reset link has been sent.
      </div>

      <form v-else class="space-y-4" @submit.prevent="submitEmail">
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

        <p v-if="error" class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">
          {{ error }}
        </p>

        <button
          type="submit"
          class="w-full rounded-lg bg-indigo-600 px-3 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
          :disabled="loading"
        >
          {{ loading ? 'Sending…' : 'Send reset link' }}
        </button>
      </form>
    </template>

    <!-- Recovery code -->
    <form v-else class="space-y-4" @submit.prevent="submitCode">
      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700" for="code-email">
          Email
        </label>
        <input
          id="code-email"
          v-model="codeEmail"
          type="email"
          autocomplete="email"
          required
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
        />
      </div>

      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700" for="code">
          Recovery code
        </label>
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
        <label class="mb-1 block text-sm font-medium text-slate-700" for="code-password">
          New password
        </label>
        <input
          id="code-password"
          v-model="password"
          type="password"
          autocomplete="new-password"
          required
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
        />
      </div>

      <div>
        <label class="mb-1 block text-sm font-medium text-slate-700" for="code-confirm">
          Confirm new password
        </label>
        <input
          id="code-confirm"
          v-model="confirm"
          type="password"
          autocomplete="new-password"
          required
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 focus:outline-none"
        />
      </div>

      <p class="text-xs text-slate-500">
        Each code can be used once. Setting a new password signs you out everywhere.
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
