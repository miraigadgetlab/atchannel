import { createRouter, createWebHistory } from 'vue-router'

import { setSeo } from '@/lib/seo'
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
    /** Meta description search engines show under the result. */
    description?: string
    /** Keeps the route out of search indexes (settings, auth, 404). */
    noindex?: boolean
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
      meta: {
        title: 'Home',
        description: 'Latest posts and channels on atchannel — read, reply, share what you know.',
      },
    },
    {
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView.vue'),
      meta: { guestOnly: true, title: 'Sign in', noindex: true },
    },
    {
      path: '/register',
      name: 'register',
      component: () => import('@/views/RegisterView.vue'),
      meta: { guestOnly: true, title: 'Register', noindex: true },
    },
    {
      path: '/forgot-password',
      name: 'forgot-password',
      component: () => import('@/views/ForgotPasswordView.vue'),
      meta: { guestOnly: true, title: 'Recover account', noindex: true },
    },
    {
      path: '/channels',
      name: 'channels',
      component: () => import('@/views/ChannelsView.vue'),
      meta: {
        title: 'Channels',
        description: 'Browse every discussion channel on atchannel.',
      },
    },
    {
      path: '/channelers',
      name: 'channelers',
      component: () => import('@/views/ChannelersView.vue'),
      meta: {
        title: 'Channelers',
        description: 'The most active channelers on atchannel, ranked by post count.',
      },
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
      meta: { requiresAuth: true, title: 'Settings', noindex: true },
    },
    {
      path: '/notifications',
      name: 'notifications',
      component: () => import('@/views/NotificationsView.vue'),
      meta: { requiresAuth: true, title: 'Notifications', noindex: true },
    },
    {
      path: '/admin',
      name: 'admin',
      component: () => import('@/views/AdminView.vue'),
      meta: { requiresAuth: true, roles: ['admin'], title: 'Admin', noindex: true },
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'not-found',
      component: () => import('@/views/NotFoundView.vue'),
      meta: { title: 'Page not found', noindex: true },
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

// Every navigation rewrites the tab title and the meta tags search engines
// read (description, robots, canonical, cards); content views then refine it
// with their own data — and JSON-LD — once loaded.
router.afterEach((to) => {
  setSeo({
    title: typeof to.meta.title === 'string' ? to.meta.title : undefined,
    description: typeof to.meta.description === 'string' ? to.meta.description : undefined,
    noindex: to.meta.noindex,
  })
})

export default router
