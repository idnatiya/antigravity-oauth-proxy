import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'login',
    component: () => import('@/views/LoginView.vue'),
    meta: { guestOnly: true },
  },
  {
    path: '/',
    redirect: '/overview',
  },
  {
    path: '/overview',
    name: 'overview',
    component: () => import('@/views/OverviewView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/playground',
    name: 'playground',
    component: () => import('@/views/PlaygroundView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/images',
    name: 'images',
    component: () => import('@/views/ImagesView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/requests',
    name: 'requests',
    component: () => import('@/views/RequestsView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/models',
    name: 'models',
    component: () => import('@/views/ModelsView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/accounts',
    name: 'accounts',
    component: () => import('@/views/AccountsView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/api-keys',
    name: 'api-keys',
    component: () => import('@/views/ApiKeysView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/security',
    name: 'security',
    component: () => import('@/views/SecurityView.vue'),
    meta: { requiresAuth: true },
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'not-found',
    component: () => import('@/views/NotFoundView.vue'),
  },
]

const router = createRouter({
  history: createWebHistory('/dashboard/'),
  routes,
})

router.beforeEach(async (to) => {
  const auth = useAuthStore()
  if (!auth.initialized) {
    await auth.initAuth()
  }

  if (to.meta.requiresAuth && !auth.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.guestOnly && auth.isAuthenticated) {
    return { name: 'overview' }
  }

  return true
})

export default router
