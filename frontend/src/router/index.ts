import { createRouter, createWebHistory } from 'vue-router'
import type { RouteRecordRaw } from 'vue-router'
import baseRoutes from './routes/base'
import { useAuthStore } from '@/stores/auth.store'

export const routes: RouteRecordRaw[] = [
  ...baseRoutes,
]

const router = createRouter({
  history: createWebHistory(),
  routes,
})

router.beforeEach(async (to, _from, next) => {
  const authStore = useAuthStore()

  // 页面标题
  if (to.meta.title) {
    document.title = `${to.meta.title} - 慧财智能财务平台`
  }

  // 认证守卫
  if (!authStore.isLoggedIn && to.path !== '/login') {
    return next({ path: '/login', query: { redirect: to.fullPath } })
  }
  if (authStore.isLoggedIn && to.path === '/login') {
    return next({ path: '/dashboard' })
  }

  // 权限守卫
  const allowedRoles = to.meta.roles as string[] | undefined
  if (allowedRoles && authStore.user) {
    if (!allowedRoles.includes(authStore.user.role)) {
      return next({ path: '/403' })
    }
  }

  next()
})

export default router
