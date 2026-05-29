import { createApp } from 'vue'
import { createPinia } from 'pinia'
import ElementPlus from 'element-plus'
import 'element-plus/dist/index.css'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'
import * as ElementPlusIconsVue from '@element-plus/icons-vue'
import App from './App.vue'
import router from './router'
import { tenantResetPlugin } from './stores/plugins/tenant-reset'
import { permissionDirective } from './directives/permission'
import './styles/index.scss'

// 条件启动 Mock Service Worker
async function bootstrap() {
  if (import.meta.env.VITE_ENABLE_MOCK === 'true') {
    const { worker } = await import('./api/mock/browser')
    await worker.start({ onUnhandledRequest: 'bypass' })
    console.log('[MSW] Mock Service Worker started')
  }

  const app = createApp(App)
  const pinia = createPinia()

  pinia.use(tenantResetPlugin)

  app.directive('permission', permissionDirective)

  for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
    app.component(key, component)
  }

  app.use(pinia)
  app.use(router)
  app.use(ElementPlus, { locale: zhCn })
  app.mount('#app')
}

bootstrap()
