<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import AppAvatar from '@/components/AppAvatar.vue'
import { errorMessage } from '@/lib/errors'
import { useAuthStore } from '@/stores/auth'
import { usersApi } from '@/api/users'
import type { UpdateProfileInput } from '@/api/users'

const auth = useAuthStore()
const router = useRouter()

// ---- Profile fields -------------------------------------------------------
const profileForm = reactive({
  name: auth.user?.name ?? '',
  email: auth.user?.email ?? '',
  about_me: auth.user?.about_me ?? '',
  note: auth.user?.note ?? '',
  current_password: '',
})
const profileSaving = ref(false)
const profileMessage = ref('')
const profileError = ref('')

async function saveProfile(): Promise<void> {
  profileError.value = ''
  profileMessage.value = ''

  // The email is only sent when it actually changed, because the backend
  // requires the current password to authorize an email change.
  const input: UpdateProfileInput = {
    name: profileForm.name,
    about_me: profileForm.about_me,
    note: profileForm.note,
  }
  if (profileForm.email !== auth.user?.email) {
    input.email = profileForm.email
    input.current_password = profileForm.current_password
  }

  profileSaving.value = true
  try {
    await auth.updateProfile(input)
    profileForm.current_password = ''
    profileMessage.value = 'Profile updated.'
  } catch (err) {
    profileError.value = errorMessage(err)
  } finally {
    profileSaving.value = false
  }
}

// ---- Avatar ---------------------------------------------------------------
const avatarUploading = ref(false)
const avatarError = ref('')

async function onAvatarChange(event: Event): Promise<void> {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) {
    return
  }

  avatarError.value = ''
  avatarUploading.value = true
  try {
    await auth.uploadAvatar(file)
  } catch (err) {
    avatarError.value = errorMessage(err)
  } finally {
    avatarUploading.value = false
    input.value = ''
  }
}

// ---- Password -------------------------------------------------------------
const passwordForm = reactive({ current: '', next: '', confirm: '' })
const passwordSaving = ref(false)
const passwordMessage = ref('')
const passwordError = ref('')

async function savePassword(): Promise<void> {
  passwordError.value = ''
  passwordMessage.value = ''

  if (passwordForm.next !== passwordForm.confirm) {
    passwordError.value = 'The two new passwords do not match.'
    return
  }

  passwordSaving.value = true
  try {
    await auth.changePassword(passwordForm.current, passwordForm.next)
    // A password change revokes every session server-side, so this one is
    // gone too: clear it and send the user to sign in again.
    auth.clearSession()
    await router.push({ name: 'login' })
  } catch (err) {
    passwordError.value = errorMessage(err)
  } finally {
    passwordSaving.value = false
  }
}

// ---- Sessions -------------------------------------------------------------
const signingOut = ref(false)

async function signOutEverywhere(): Promise<void> {
  signingOut.value = true
  try {
    await auth.logoutAll()
    await router.push({ name: 'home' })
  } finally {
    signingOut.value = false
  }
}

// ---- Email verification ---------------------------------------------------
const resending = ref(false)
const resent = ref(false)

async function resendVerification(): Promise<void> {
  resending.value = true
  try {
    await auth.resendVerification()
    resent.value = true
  } finally {
    resending.value = false
  }
}

// ---- Recovery codes -------------------------------------------------------
const codesRemaining = ref<number | null>(null)
const recoveryLoading = ref(false)
const recoveryError = ref('')
const generatedCodes = ref<string[]>([])
const codesCopied = ref(false)

async function loadRecoveryStatus(): Promise<void> {
  try {
    const { remaining } = await usersApi.recoveryCodeStatus()
    codesRemaining.value = remaining
  } catch {
    codesRemaining.value = null
  }
}

async function generateRecoveryCodes(): Promise<void> {
  recoveryError.value = ''
  codesCopied.value = false
  recoveryLoading.value = true
  try {
    const { codes } = await usersApi.generateRecoveryCodes()
    generatedCodes.value = codes
    codesRemaining.value = codes.length
  } catch (err) {
    recoveryError.value = errorMessage(err)
  } finally {
    recoveryLoading.value = false
  }
}

async function copyCodes(): Promise<void> {
  try {
    await navigator.clipboard.writeText(generatedCodes.value.join('\n'))
    codesCopied.value = true
  } catch {
    codesCopied.value = false
  }
}

onMounted(loadRecoveryStatus)
</script>

<template>
  <div class="mx-auto max-w-2xl space-y-8">
    <div>
      <h1 class="text-2xl font-bold text-slate-900">Account settings</h1>
      <p class="text-sm text-slate-500">Manage your profile, password and sessions.</p>
    </div>

    <!-- Profile -->
    <section class="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <h2 class="mb-4 text-lg font-semibold text-slate-900">Profile</h2>

      <div class="mb-5 flex items-center gap-4">
        <AppAvatar :name="auth.user?.name ?? '?'" :avatar-url="auth.user?.avatar_url" size="lg" />
        <div>
          <label
            class="inline-block cursor-pointer rounded-lg border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700 hover:bg-slate-50"
          >
            {{ avatarUploading ? 'Uploading…' : 'Change avatar' }}
            <input
              type="file"
              accept="image/png,image/jpeg,image/gif,image/webp"
              class="hidden"
              :disabled="avatarUploading"
              @change="onAvatarChange"
            />
          </label>
          <p class="mt-1 text-xs text-slate-400">PNG, JPEG, GIF or WebP.</p>
          <p v-if="avatarError" class="mt-1 text-xs text-rose-600">{{ avatarError }}</p>
        </div>
      </div>

      <form class="space-y-4" @submit.prevent="saveProfile">
        <div>
          <label class="mb-1 block text-sm font-medium text-slate-700" for="name">Name</label>
          <input
            id="name"
            v-model="profileForm.name"
            type="text"
            required
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
          />
        </div>

        <div>
          <label class="mb-1 block text-sm font-medium text-slate-700" for="email">Email</label>
          <input
            id="email"
            v-model="profileForm.email"
            type="email"
            required
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
          />
        </div>

        <div v-if="profileForm.email !== auth.user?.email">
          <label class="mb-1 block text-sm font-medium text-slate-700" for="current-password">
            Current password (required to change email)
          </label>
          <input
            id="current-password"
            v-model="profileForm.current_password"
            type="password"
            autocomplete="current-password"
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
          />
        </div>

        <div>
          <label class="mb-1 block text-sm font-medium text-slate-700" for="about">About me</label>
          <textarea
            id="about"
            v-model="profileForm.about_me"
            rows="4"
            maxlength="1000"
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
          />
        </div>

        <div>
          <label class="mb-1 block text-sm font-medium text-slate-700" for="note">
            Note for the admins
          </label>
          <textarea
            id="note"
            v-model="profileForm.note"
            rows="3"
            maxlength="500"
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
          />
          <p class="mt-1 text-xs text-slate-500">
            Private: only you and the admins can read it. It helps us recognise you if you ever lose
            access.
          </p>
        </div>

        <p v-if="profileError" class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">
          {{ profileError }}
        </p>
        <p
          v-if="profileMessage"
          class="rounded-lg bg-emerald-50 px-3 py-2 text-sm text-emerald-700"
        >
          {{ profileMessage }}
        </p>

        <div class="flex items-center justify-between">
          <button
            type="submit"
            class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
            :disabled="profileSaving"
          >
            {{ profileSaving ? 'Saving…' : 'Save profile' }}
          </button>
          <RouterLink
            v-if="auth.user"
            :to="{ name: 'profile', params: { id: auth.user.id } }"
            class="text-sm font-medium text-indigo-600 hover:text-indigo-700"
          >
            View public profile
          </RouterLink>
        </div>
      </form>
    </section>

    <!-- Email verification -->
    <section class="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <h2 class="mb-2 text-lg font-semibold text-slate-900">Email verification</h2>
      <p v-if="auth.isEmailVerified" class="text-sm text-emerald-700">
        Your email address is verified.
      </p>
      <div v-else class="flex flex-wrap items-center gap-3">
        <p class="text-sm text-slate-600">Your email address is not verified yet.</p>
        <button
          type="button"
          class="text-sm font-semibold text-indigo-600 hover:text-indigo-700 disabled:opacity-60"
          :disabled="resending || resent"
          @click="resendVerification"
        >
          {{
            resent ? 'Verification email sent' : resending ? 'Sending…' : 'Send verification link'
          }}
        </button>
      </div>
    </section>

    <!-- Recovery codes -->
    <section class="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <h2 class="mb-2 text-lg font-semibold text-slate-900">Recovery codes</h2>
      <p class="mb-4 text-sm text-slate-600">
        Single-use codes that let you set a new password without a reset email. Store them somewhere
        safe — each code works only once.
      </p>

      <p v-if="codesRemaining !== null" class="mb-4 text-sm text-slate-500">
        {{
          codesRemaining > 0
            ? `${codesRemaining} unused code${codesRemaining === 1 ? '' : 's'} left.`
            : 'You have no unused codes.'
        }}
      </p>

      <div v-if="generatedCodes.length" class="mb-4 space-y-3">
        <p class="rounded-lg bg-amber-50 px-3 py-2 text-sm text-amber-800">
          Copy these now. They are shown only once, and generating a new batch retires every old
          code.
        </p>
        <ul
          class="grid grid-cols-2 gap-x-6 gap-y-1 rounded-lg border border-slate-300 bg-slate-50 p-4 text-sm text-slate-800"
        >
          <li v-for="code in generatedCodes" :key="code">{{ code }}</li>
        </ul>
        <button
          type="button"
          class="rounded-lg border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700 hover:bg-slate-50"
          @click="copyCodes"
        >
          {{ codesCopied ? 'Copied' : 'Copy codes' }}
        </button>
      </div>

      <p v-if="recoveryError" class="mb-4 rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">
        {{ recoveryError }}
      </p>

      <button
        type="button"
        class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
        :disabled="recoveryLoading"
        @click="generateRecoveryCodes"
      >
        {{
          recoveryLoading
            ? 'Generating…'
            : generatedCodes.length || (codesRemaining ?? 0) > 0
              ? 'Generate new codes'
              : 'Generate recovery codes'
        }}
      </button>
    </section>

    <!-- Password -->
    <section class="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <h2 class="mb-4 text-lg font-semibold text-slate-900">Change password</h2>
      <form class="space-y-4" @submit.prevent="savePassword">
        <div>
          <label class="mb-1 block text-sm font-medium text-slate-700" for="pw-current">
            Current password
          </label>
          <input
            id="pw-current"
            v-model="passwordForm.current"
            type="password"
            autocomplete="current-password"
            required
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
          />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium text-slate-700" for="pw-next">
            New password
          </label>
          <input
            id="pw-next"
            v-model="passwordForm.next"
            type="password"
            autocomplete="new-password"
            required
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
          />
        </div>
        <div>
          <label class="mb-1 block text-sm font-medium text-slate-700" for="pw-confirm">
            Confirm new password
          </label>
          <input
            id="pw-confirm"
            v-model="passwordForm.confirm"
            type="password"
            autocomplete="new-password"
            required
            class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm focus:border-indigo-500 focus:outline-none"
          />
        </div>

        <p v-if="passwordError" class="rounded-lg bg-rose-50 px-3 py-2 text-sm text-rose-700">
          {{ passwordError }}
        </p>
        <p
          v-if="passwordMessage"
          class="rounded-lg bg-emerald-50 px-3 py-2 text-sm text-emerald-700"
        >
          {{ passwordMessage }}
        </p>

        <button
          type="submit"
          class="rounded-lg bg-indigo-600 px-4 py-2 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
          :disabled="passwordSaving"
        >
          {{ passwordSaving ? 'Changing…' : 'Change password' }}
        </button>
      </form>
    </section>

    <!-- Sessions -->
    <section class="rounded-xl border border-slate-200 bg-white p-6 shadow-sm">
      <h2 class="mb-2 text-lg font-semibold text-slate-900">Sessions</h2>
      <p class="mb-4 text-sm text-slate-600">
        Signing out everywhere revokes every refresh token on every device.
      </p>
      <button
        type="button"
        class="rounded-lg border border-rose-200 px-4 py-2 text-sm font-semibold text-rose-600 hover:bg-rose-50 disabled:opacity-60"
        :disabled="signingOut"
        @click="signOutEverywhere"
      >
        {{ signingOut ? 'Signing out…' : 'Sign out everywhere' }}
      </button>
    </section>
  </div>
</template>
