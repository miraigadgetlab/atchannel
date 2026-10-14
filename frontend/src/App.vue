<script setup lang="ts">
import { ref } from 'vue'
import { RouterView } from 'vue-router'

import AppNav from '@/components/AppNav.vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

const resending = ref(false)
const resent = ref(false)
const resendError = ref('')

async function resendVerification(): Promise<void> {
  resending.value = true
  resendError.value = ''
  try {
    await auth.resendVerification()
    resent.value = true
  } catch {
    resendError.value = 'Could not send the verification email.'
  } finally {
    resending.value = false
  }
}
</script>

<template>
  <div class="flex min-h-screen flex-col">
    <AppNav />

    <div
      v-if="auth.isAuthenticated && !auth.isEmailVerified"
      class="border-b border-amber-200 bg-amber-50"
    >
      <div
        class="mx-auto flex max-w-5xl flex-wrap items-center gap-x-3 gap-y-1 px-4 py-2 text-sm text-amber-800"
      >
        <span>Your email address is not verified yet.</span>
        <button
          v-if="!resent"
          type="button"
          class="font-semibold underline disabled:opacity-60"
          :disabled="resending"
          @click="resendVerification"
        >
          {{ resending ? 'Sending…' : 'Resend verification link' }}
        </button>
        <span v-else class="font-semibold">Verification email sent.</span>
        <span v-if="resendError" class="text-rose-700">{{ resendError }}</span>
      </div>
    </div>

    <main class="mx-auto w-full max-w-5xl flex-1 px-4 py-6">
      <RouterView />
    </main>

    <footer class="border-t border-slate-200 py-6 text-center text-xs text-slate-400">
      atchannel — a small discussion forum.
    </footer>
  </div>
</template>
