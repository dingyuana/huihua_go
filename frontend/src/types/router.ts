import type { LayoutType } from '@/types/enums'

/** 路由元信息 */
declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    layout?: LayoutType
    roles?: string[]
    permissions?: string[]
    keepAlive?: boolean
  }
}
