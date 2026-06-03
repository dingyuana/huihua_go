import type { Directive } from 'vue'
import { useAuthStore } from '@/stores/auth.store'

/** 按钮级权限指令：v-permission="['voucher:submit']" */
export const permissionDirective: Directive = {
  mounted(el, binding) {
    const authStore = useAuthStore()
    const requiredPermissions = binding.value as string[]
    const hasPermission = requiredPermissions.some(p => authStore.hasPermission(p))
    if (!hasPermission) {
      el.parentNode?.removeChild(el)
    }
  },
}
