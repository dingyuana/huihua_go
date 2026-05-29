import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('@/views/auth/Login.vue'),
    meta: { public: true }
  },
  {
    path: '/',
    redirect: '/dashboard'
  },
  {
    path: '/dashboard',
    name: 'Dashboard',
    component: () => import('@/views/dashboard/Dashboard.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/vouchers',
    name: 'VoucherList',
    component: () => import('@/views/voucher/VoucherList.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/vouchers/new',
    name: 'VoucherNew',
    component: () => import('@/views/voucher/VoucherForm.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/vouchers/:id/edit',
    name: 'VoucherEdit',
    component: () => import('@/views/voucher/VoucherForm.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/bank-transactions',
    name: 'BankTxnList',
    component: () => import('@/views/bank/BankTxnList.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/invoices',
    name: 'InvoiceList',
    component: () => import('@/views/invoice/InvoiceList.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/parties',
    name: 'PartyList',
    component: () => import('@/views/party/PartyList.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/reports/trial-balance',
    name: 'TrialBalance',
    component: () => import('@/views/report/TrialBalance.vue'),
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()
  
  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
  } else if (to.path === '/login' && authStore.isAuthenticated) {
    next('/dashboard')
  } else {
    next()
  }
})

export default router