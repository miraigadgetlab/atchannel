<script setup lang="ts">
import { onMounted, ref } from 'vue'

import AppPagination from '@/components/AppPagination.vue'
import AppSpinner from '@/components/AppSpinner.vue'
import EmptyState from '@/components/EmptyState.vue'
import { adminApi } from '@/api/admin'
import { errorMessage } from '@/lib/errors'
import { formatDate } from '@/lib/format'
import { useAuthStore } from '@/stores/auth'
import type { Role, Stats, User } from '@/types/api'

const LIMIT = 20
const ALL_ROLES: Role[] = ['user', 'moderator', 'admin']

const auth = useAuthStore()

interface UserRow {
  user: User
  roles: Role[]
  busy: boolean
  notice: string
  codes: string[]
}

const rows = ref<UserRow[]>([])
const total = ref(0)
const offset = ref(0)
const loading = ref(true)
const error = ref('')

const stats = ref<Stats | null>(null)

async function load(): Promise<void> {
  loading.value = true
  error.value = ''
  try {
    const data = await adminApi.listUsers({ limit: LIMIT, offset: offset.value })
    rows.value = (data.items ?? []).map((user) => ({
      user,
      roles: [...user.roles],
      busy: false,
      notice: '',
      codes: [],
    }))
    total.value = data.total
  } catch (err) {
    error.value = errorMessage(err)
    rows.value = []
    total.value = 0
  } finally {
    loading.value = false
  }
}

async function loadStats(): Promise<void> {
  try {
    stats.value = await adminApi.stats()
  } catch {
    stats.value = null
  }
}

function changeOffset(next: number): void {
  offset.value = next
  void load()
}

async function saveRoles(row: UserRow): Promise<void> {
  row.busy = true
  row.notice = ''
  try {
    const result = await adminApi.setRoles(row.user.id, row.roles)
    row.user = result.user
    row.roles = [...result.user.roles]
    row.notice = result.warning
      ? result.warning
      : `Saved. ${result.sessions_revoked} session(s) revoked.`
  } catch (err) {
    row.notice = errorMessage(err)
  } finally {
    row.busy = false
  }
}

async function toggleBan(row: UserRow): Promise<void> {
  row.busy = true
  row.notice = ''
  try {
    const banned = row.user.banned_at === undefined
    await adminApi.setBanned(row.user.id, banned)
    await load()
  } catch (err) {
    row.notice = errorMessage(err)
  } finally {
    row.busy = false
  }
}

async function deleteUser(row: UserRow): Promise<void> {
  if (!window.confirm(`Permanently delete ${row.user.name}?`)) {
    return
  }
  row.busy = true
  try {
    await adminApi.deleteUser(row.user.id)
    await Promise.all([load(), loadStats()])
  } catch (err) {
    row.notice = errorMessage(err)
  } finally {
    row.busy = false
  }
}

// GenerateRecoveryCodes hands the plaintext codes to the admin exactly once.
// The note reminds them to pass the codes on before navigating away.
async function generateCodes(row: UserRow): Promise<void> {
  row.busy = true
  row.notice = ''
  row.codes = []
  try {
    const result = await adminApi.generateRecoveryCodes(row.user.id)
    row.codes = result.codes
    row.notice = 'New recovery codes generated. Copy them now — they are shown only once.'
  } catch (err) {
    row.notice = errorMessage(err)
  } finally {
    row.busy = false
  }
}

onMounted(async () => {
  await Promise.all([load(), loadStats()])
})
</script>

<template>
  <div>
    <h1 class="mb-6 text-2xl font-bold text-slate-900">Admin</h1>

    <div v-if="stats" class="mb-8 grid grid-cols-2 gap-4 sm:grid-cols-4">
      <div class="rounded-xl border border-slate-200 bg-white p-4 text-center">
        <p class="text-xs font-medium tracking-wide text-slate-400 uppercase">Users</p>
        <p class="text-2xl font-semibold text-slate-800">{{ stats.users }}</p>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white p-4 text-center">
        <p class="text-xs font-medium tracking-wide text-slate-400 uppercase">Channels</p>
        <p class="text-2xl font-semibold text-slate-800">{{ stats.channels }}</p>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white p-4 text-center">
        <p class="text-xs font-medium tracking-wide text-slate-400 uppercase">Posts</p>
        <p class="text-2xl font-semibold text-slate-800">{{ stats.posts }}</p>
      </div>
      <div class="rounded-xl border border-slate-200 bg-white p-4 text-center">
        <p class="text-xs font-medium tracking-wide text-slate-400 uppercase">Comments</p>
        <p class="text-2xl font-semibold text-slate-800">{{ stats.comments }}</p>
      </div>
    </div>

    <AppSpinner v-if="loading" />
    <p v-else-if="error" class="rounded-lg bg-rose-50 px-4 py-3 text-sm text-rose-700">
      {{ error }}
    </p>
    <EmptyState v-else-if="rows.length === 0" title="No users" />

    <div v-else class="space-y-3">
      <div
        v-for="row in rows"
        :key="row.user.id"
        class="rounded-xl border border-slate-200 bg-white p-4"
      >
        <div class="flex flex-wrap items-start justify-between gap-3">
          <div>
            <p class="font-medium text-slate-900">
              {{ row.user.name }}
              <span
                v-if="row.user.banned_at"
                class="ml-2 rounded-full bg-rose-100 px-2 py-0.5 text-xs font-semibold text-rose-700"
              >
                Banned
              </span>
              <span
                v-if="row.user.id === auth.user?.id"
                class="ml-2 rounded-full bg-indigo-100 px-2 py-0.5 text-xs font-semibold text-indigo-700"
              >
                You
              </span>
            </p>
            <p class="text-sm text-slate-500">{{ row.user.email }}</p>
            <p class="text-xs text-slate-400">Joined {{ formatDate(row.user.created_at) }}</p>
            <p v-if="row.user.note" class="mt-1 text-xs text-slate-500 whitespace-pre-wrap">
              <span class="font-medium text-slate-600">Note:</span> {{ row.user.note }}
            </p>
          </div>

          <div class="flex flex-wrap items-center gap-2">
            <button
              type="button"
              class="rounded-lg border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-60"
              :disabled="row.busy"
              @click="generateCodes(row)"
            >
              New codes
            </button>
            <template v-if="row.user.id !== auth.user?.id">
              <button
                type="button"
                class="rounded-lg border border-slate-300 px-3 py-1.5 text-sm font-medium text-slate-700 hover:bg-slate-50 disabled:opacity-60"
                :disabled="row.busy"
                @click="toggleBan(row)"
              >
                {{ row.user.banned_at ? 'Unban' : 'Ban' }}
              </button>
              <button
                type="button"
                class="rounded-lg border border-rose-200 px-3 py-1.5 text-sm font-medium text-rose-600 hover:bg-rose-50 disabled:opacity-60"
                :disabled="row.busy"
                @click="deleteUser(row)"
              >
                Delete
              </button>
            </template>
          </div>
        </div>

        <div class="mt-3 flex flex-wrap items-center gap-4 border-t border-slate-100 pt-3">
          <label
            v-for="role in ALL_ROLES"
            :key="role"
            class="inline-flex items-center gap-1.5 text-sm text-slate-600"
          >
            <input
              v-model="row.roles"
              type="checkbox"
              :value="role"
              class="rounded border-slate-300 text-indigo-600 focus:ring-indigo-500"
            />
            {{ role }}
          </label>

          <button
            type="button"
            class="rounded-lg bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white hover:bg-indigo-700 disabled:opacity-60"
            :disabled="row.busy || row.user.id === auth.user?.id"
            @click="saveRoles(row)"
          >
            Save roles
          </button>

          <span v-if="row.notice" class="text-xs text-slate-500">{{ row.notice }}</span>
        </div>

        <div
          v-if="row.codes.length"
          class="mt-3 rounded-lg border border-amber-200 bg-amber-50 p-3"
        >
          <p class="mb-2 text-xs font-medium text-amber-800">
            Shown only once — hand these to the user.
          </p>
          <ul class="grid grid-cols-2 gap-x-6 gap-y-1 text-sm text-slate-800">
            <li v-for="code in row.codes" :key="code">{{ code }}</li>
          </ul>
        </div>
      </div>

      <AppPagination :total="total" :limit="LIMIT" :offset="offset" @update="changeOffset" />
    </div>
  </div>
</template>
