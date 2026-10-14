import { createRouter, createWebHistory } from 'vue-router'

import { setTitle } from '@/lib/title'
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
    /** Tab title label; content views override it once their data loads. */
    title?: string
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
      meta: { guestOnly: true, title: 'Sign in' },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/RegisterView.vue'),
      meta: { guestOnly: true, title: 'Register' },
    },
    {
      path: '/forgot-password',
      name: 'forgot-password',
      component: () => import('@/views/ForgotPasswordView.vue'),
      meta: { guestOnly: true, title: 'Forgot password' },
    },
    {
      path: '/reset-password',
      name: 'reset-password',
      component: () => import('@/views/ResetPasswordView.vue'),
      meta: { guestOnly: true, title: 'Reset password' },
    },
    {
      path: '/channels',
      name: 'channels',
      component: () => import('@/views/ChannelsView.vue'),
      meta: { title: 'Channels' },
    },
    {
      path: '/channelers',
      name: 'channelers',
      component: () => import('@/views/ChannelersView.vue'),
      meta: { title: 'Channelers' },
    },
    {
      path: '/channels/:id',
      name: 'channel',
      component: () => import('@/views/ChannelDetailView.vue'),
      props: true,
      meta: { title: 'Channel' },
    },
    {
      path: '/posts/new',
      name: 'post-new',
      component: () => import('@/views/PostEditorView.vue'),
      meta: { requiresAuth: true, title: 'New post' },
    },
    {
      path: '/posts/:id/edit',
      name: 'post-edit',
      component: () => import('@/views/PostEditorView.vue'),
      props: true,
      meta: { requiresAuth: true, title: 'Edit post' },
    },
    {
      path: '/posts/:id',
      name: 'post',
      component: () => import('@/views/PostDetailView.vue'),
      props: true,
      meta: { title: 'Post' },
    },
    {
      path: '/users/:id',
      name: 'profile',
      component: () => import('@/views/ProfileView.vue'),
      props: true,
      meta: { title: 'Profile' },
    },
    {
      path: '/me',
      name: 'me',
      component: () => import('@/views/MeView.vue'),
      meta: { requiresAuth: true, title: 'Settings' },
    },
    {
      path: '/notifications',
      name: 'notifications',
      component: () => import('@/views/NotificationsView.vue'),
      meta: { requiresAuth: true, title: 'Notifications' },
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('@/views/AdminView.vue'),
      meta: { requiresAuth: true, roles: ['admin'], title: 'Admin' },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFoundView.vue'),
      meta: { title: 'Page not found' },
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

// Every navigation resets the tab title from the route's static label; content
// views then refine it (post title, channel title, profile name) once loaded.
router.afterEach((to) => {
  setTitle(typeof to.meta.title === 'string' ? to.meta.title : undefined)
})

export default router
