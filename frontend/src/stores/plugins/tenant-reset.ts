/**
 * 多租户切换时自动重置业务 Store
 */
import type { PiniaPluginContext } from 'pinia'
import { watch } from 'vue'

export function tenantResetPlugin({ store }: PiniaPluginContext) {
  // 当 store 名不在系统列表中，标记为业务 store
  const systemStores = ['auth', 'tenant', 'app']

  if (!systemStores.includes(store.$id)) {
    // 监听 tenant store 的 currentTenantId 变化
    const tenantStore = useTenantStore?.() // 运行时获取
    if (tenantStore) {
      watch(() => tenantStore.currentTenantId, () => {
        store.$reset()
      })
    }
  }
}

// 延迟引用避免循环依赖
let useTenantStore: any
import('@/stores/tenant.store').then(m => {
  useTenantStore = m.useTenantStore
})
