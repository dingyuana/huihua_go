import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { Tenant, Company } from '@/types/models/tenant'

export const useTenantStore = defineStore('tenant', () => {
  const currentTenantId = ref<string | null>(null)
  const currentCompany = ref<Company | null>(null)
  const tenantList = ref<Tenant[]>([])

  const watermark = computed(() => {
    if (currentCompany.value) {
      return `当前操作：${currentCompany.value.name}`
    }
    return ''
  })

  function switchTenant(id: string) {
    currentTenantId.value = id
    // 切换租户时重置所有业务 store（由 plugin 触发）
  }

  function setTenantList(list: Tenant[]) {
    tenantList.value = list
  }

  return { currentTenantId, currentCompany, tenantList, watermark, switchTenant, setTenantList }
})
