<script setup lang="ts">
import { ref, watch } from 'vue'
import { RouterLink, useRouter } from 'vue-router'

import AppAvatar from './AppAvatar.vue'
import { useAuthStore } from '@/stores/auth'
import { useNotificationsStore } from '@/stores/notifications'
import { profileRoute } from '@/lib/profileUrl'
import { userSubdomain } from '@/lib/subdomain'

const auth = useAuthStore()
const notifications = useNotificationsStore()
const router = useRouter()

const menuOpen = ref(false)

// On a user subdomain (kosero.atchannel.top) the root path is that account's
// profile, so links that lead away from it point back at the apex site: the
// brand and "Feed" keep the feed reachable, and "My profile" avoids rendering
// a profile page under someone else's subdomain.
const sub = userSubdomain()
const apexOrigin = sub
  ? `${window.location.protocol}//${window.location.host.slice(sub.length + 1)}`
  : ''
const apexHref = `${apexOrigin}/`
// On a subdomain "My profile" has to leave that subdomain, so the target is
// resolved to a path (/emirhan, or /users/4 when the name is taken) and then
// prefixed with the apex origin.
const myProfilePath = auth.user ? router.resolve(profileRoute(auth.user)).fullPath : '/'
const myProfileHref = `${apexOrigin}${myProfilePath}`

watch(
  () => auth.isAuthenticated,
  (signedIn) => {
    if (signedIn) {
      void notifications.refreshUnread()
    } else {
      notifications.reset()
    }
  },
  { immediate: true },
)

async function logout(): Promise<void> {
  menuOpen.value = false
  await auth.logout()
  void router.push({ name: 'home' })
}
</script>

<template>
  <header class="sticky top-0 z-40 border-b border-slate-200 bg-white/90 backdrop-blur">
    <div class="mx-auto flex max-w-5xl flex-wrap items-center gap-x-4 gap-y-2 px-4 py-3">
      <RouterLink v-if="!sub" to="/" class="order-1 brand text-indigo-700"> atchannel </RouterLink>
      <a v-else :href="apexHref" class="order-1 brand text-indigo-700"> atchannel </a>

      <!-- On narrow screens the nav drops to its own row (order-3, w-full)
           while brand and the account controls share the top row, so the two
           never collide the way they did when everything wrapped together. -->
      <nav
        class="order-3 flex w-full items-center justify-center gap-3 text-sm font-medium text-slate-600 sm:order-2 sm:w-auto sm:justify-start"
      >
        <RouterLink v-if="!sub" to="/" class="hover:text-slate-900" exact-active-class="nav-active">
          Feed
        </RouterLink>
        <a v-else :href="apexHref" class="hover:text-slate-900"> Feed </a>
        <RouterLink to="/channels" class="hover:text-slate-900" active-class="nav-active">
          Channels
        </RouterLink>
        <RouterLink to="/channelers" class="hover:text-slate-900" active-class="nav-active">
          Channelers
        </RouterLink>
      </nav>

      <div class="order-2 ml-auto flex items-center gap-2 sm:order-3">
        <template v-if="auth.isAuthenticated">
          <RouterLink
            :to="{ name: 'post-new' }"
            class="rounded-lg bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white hover:bg-indigo-700"
          >
            New post
          </RouterLink>

          <RouterLink
            :to="{ name: 'notifications' }"
            class="relative rounded-lg p-2 text-slate-600 hover:bg-slate-100"
            aria-label="Notifications"
          >
            <svg viewBox="0 0 20 20" fill="currentColor" class="h-5 w-5" aria-hidden="true">
              <path
                d="M10 2a6 6 0 0 0-6 6v2.6l-1.2 2.4A1 1 0 0 0 3.7 14.5h12.6a1 1 0 0 0 .9-1.5L16 10.6V8a6 6 0 0 0-6-6Zm0 16a2.5 2.5 0 0 0 2.45-2h-4.9A2.5 2.5 0 0 0 10 18Z"
              />
            </svg>
            <span
              v-if="notifications.unread > 0"
              class="absolute -top-0.5 -right-0.5 inline-flex min-w-[1.1rem] items-center justify-center rounded-full bg-rose-500 px-1 text-[10px] font-bold text-white"
            >
              {{ notifications.unread > 9 ? '9+' : notifications.unread }}
            </span>
          </RouterLink>

          <div class="relative">
            <button
              type="button"
              class="flex items-center gap-2 rounded-lg px-1.5 py-1 hover:bg-slate-100"
              @click="menuOpen = !menuOpen"
            >
              <AppAvatar
                :name="auth.user?.name ?? '?'"
                :avatar-url="auth.user?.avatar_url"
                size="sm"
              />
              <span class="hidden text-sm font-medium text-slate-700 sm:block">
                {{ auth.user?.name }}
              </span>
            </button>

            <template v-if="menuOpen">
              <button
                type="button"
                class="fixed inset-0 z-10 cursor-default"
                aria-hidden="true"
                @click="menuOpen = false"
              />
              <div
                class="absolute right-0 z-20 mt-2 w-48 overflow-hidden rounded-lg border border-slate-200 bg-white py-1 shadow-lg"
              >
                <a
                  v-if="sub"
                  :href="myProfileHref"
                  class="block px-4 py-2 text-sm text-slate-700 hover:bg-slate-50"
                  @click="menuOpen = false"
                >
                  My profile
                </a>
                <RouterLink
                  v-else
                  :to="auth.user ? profileRoute(auth.user) : { name: 'home' }"
                  class="block px-4 py-2 text-sm text-slate-700 hover:bg-slate-50"
                  @click="menuOpen = false"
                >
                  My profile
                </RouterLink>
                <RouterLink
                  :to="{ name: 'me' }"
                  class="block px-4 py-2 text-sm text-slate-700 hover:bg-slate-50"
                  @click="menuOpen = false"
                >
                  Settings
                </RouterLink>
                <RouterLink
                  v-if="auth.isAdmin"
                  :to="{ name: 'admin' }"
                  class="block px-4 py-2 text-sm text-slate-700 hover:bg-slate-50"
                  @click="menuOpen = false"
                >
                  Admin
                </RouterLink>
                <button
                  type="button"
                  class="block w-full px-4 py-2 text-left text-sm text-rose-600 hover:bg-rose-50"
                  @click="logout"
                >
                  Sign out
                </button>
              </div>
            </template>
          </div>
        </template>

        <template v-else>
          <RouterLink
            :to="{ name: 'login' }"
            class="rounded-lg px-3 py-1.5 text-sm font-medium text-slate-700 hover:bg-slate-100"
          >
            Sign in
          </RouterLink>
          <RouterLink
            :to="{ name: 'register' }"
            class="rounded-lg bg-indigo-600 px-3 py-1.5 text-sm font-semibold text-white hover:bg-indigo-700"
          >
            Register
          </RouterLink>
        </template>
      </div>
    </div>
  </header>
</template>
