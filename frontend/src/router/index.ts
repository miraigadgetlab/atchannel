import { createRouter, createWebHistory } from 'vue-router'

import { useAuthStore } from '@/stores/auth'
import type { Role } from '@/types/api'

// Route meta is typed here so `to.meta.requiresAuth` etc. are checked by vue-tsc.
declare module 'vue-router' {
  interface RouteMeta {
    /** Redirects anonymous visitors to /login. */
    requiresAuth?: boolean
    /** Redirects signed-in visitors away (login/register/...). */
    guestOnly?: boolean
    /** Caller must hold at least one of these roles. */
    roles?: Role[]
  }
}

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  scrollBehavior(_to, _from, savedPosition) {
    return savedPosition ?? { top: 0 }
  },
  routes: [
    {
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView.vue'),
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/RegisterView.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/forgot-password',
      name: 'forgot-password',
      component: () => import('@/views/ForgotPasswordView.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/reset-password',
      name: 'reset-password',
      component: () => import('@/views/ResetPasswordView.vue'),
      meta: { guestOnly: true },
    },
    {
      path: '/verify-email',
      name: 'verify-email',
      component: () => import('@/views/VerifyEmailView.vue'),
    },
    {
      path: '/channels',
      name: 'channels',
      component: () => import('@/views/ChannelsView.vue'),
    },
    {
      path: '/channelers',
      name: 'channelers',
      component: () => import('@/views/ChannelersView.vue'),
    },
    {
      path: '/channels/:id',
      name: 'channel',
      component: () => import('@/views/ChannelDetailView.vue'),
      props: true,
    },
    {
      path: '/posts/new',
      name: 'post-new',
      component: () => import('@/views/PostEditorView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/posts/:id/edit',
      name: 'post-edit',
      component: () => import('@/views/PostEditorView.vue'),
      props: true,
      meta: { requiresAuth: true },
    },
    {
      path: '/posts/:id',
      name: 'post',
      component: () => import('@/views/PostDetailView.vue'),
      props: true,
    },
    {
      path: '/users/:id',
      name: 'profile',
      component: () => import('@/views/ProfileView.vue'),
      props: true,
    },
    {
      path: '/me',
      name: 'me',
      component: () => import('@/views/MeView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/notifications',
      name: 'notifications',
      component: () => import('@/views/NotificationsView.vue'),
      meta: { requiresAuth: true },
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('@/views/AdminView.vue'),
      meta: { requiresAuth: true, roles: ['admin'] },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFoundView.vue'),
    },
  ],
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()

  // The persisted session is restored before the app mounts; this is only a
  // safety net for navigations triggered before that finishes.
  if (!auth.initialized) {
    await auth.initialize()
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.guestOnly && auth.isAuthenticated) {
    return { name: 'home' }
  }

  if (to.meta.roles && !to.meta.roles.some((role) => auth.roles.includes(role))) {
    return { name: 'home' }
  }

  return true
})

export default router
