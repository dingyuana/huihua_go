<template>
  <n-layout style="height: 100vh">
    <!-- 顶部导航 -->
    <n-layout-header style="height: 60px; padding: 0 24px; display: flex; align-items: center; justify-content: space-between; border-bottom: 1px solid #f0f0f0; background: #fff;">
      <div style="display: flex; align-items: center; gap: 16px;">
        <h2 style="margin: 0; font-size: 18px; font-weight: 600;">慧话财务</h2>
        <n-button quaternary @click="toggleSidebar">
          <template #icon>
            <n-icon><MenuOutline /></n-icon>
          </template>
        </n-button>
      </div>
      <div style="display: flex; align-items: center; gap: 16px;">
        <span style="color: #666;">{{ authStore.currentUser?.username }}</span>
        <n-dropdown :options="userMenuOptions" @select="handleUserMenuSelect">
          <n-button quaternary>
            <template #icon>
              <n-icon><person-outline /></n-icon>
            </template>
          </n-button>
        </n-dropdown>
      </div>
    </n-layout-header>

    <n-layout style="top: 60px;">
      <!-- 侧边菜单 -->
      <n-layout-sider
        v-if="!sidebarCollapsed"
        width="220"
        :native-scrollbar="false"
        bordered
        style="background: #fff;"
      >
        <n-menu
          v-model:value="activeKey"
          :options="menuOptions"
          @update:value="handleMenuSelect"
        />
      </n-layout-sider>

      <!-- 主内容 -->
      <n-layout-content style="padding: 16px; background: #f5f5f5;">
        <router-view />
      </n-layout-content>
    </n-layout>
  </n-layout>
</template>

<script setup lang="ts">
import { ref, computed, h } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { NLayout, NLayoutHeader, NLayoutSider, NLayoutContent, NMenu, NButton, NIcon, NDropdown, type MenuOption } from 'naive-ui'
import { 
  MenuOutline, 
  PersonOutline,
  LogOutOutline,
  HomeOutline,
  DocumentTextOutline,
  CardOutline,
  ReceiptOutline,
  PeopleOutline,
  BarChartOutline,
  SettingsOutline
} from '@vicons/ionicons5'
import { useAuthStore } from '@/stores/auth'
import { useAppStore } from '@/stores/app'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const appStore = useAppStore()

const activeKey = ref(route.name as string)

const sidebarCollapsed = computed(() => appStore.sidebarCollapsed)

const toggleSidebar = () => {
  appStore.toggleSidebar()
}

const menuOptions: MenuOption[] = [
  {
    label: '工作台',
    key: 'Dashboard',
    icon: () => h(NIcon, null, { default: () => h(HomeOutline) })
  },
  {
    label: '凭证管理',
    key: 'VoucherList',
    icon: () => h(NIcon, null, { default: () => h(DocumentTextOutline) })
  },
  {
    label: '银行流水',
    key: 'BankTxnList',
    icon: () => h(NIcon, null, { default: () => h(CardOutline) })
  },
  {
    label: '发票',
    key: 'InvoiceList',
    icon: () => h(NIcon, null, { default: () => h(ReceiptOutline) })
  },
  {
    label: '往来单位',
    key: 'PartyList',
    icon: () => h(NIcon, null, { default: () => h(PeopleOutline) })
  },
  {
    label: '报表',
    key: 'TrialBalance',
    icon: () => h(NIcon, null, { default: () => h(BarChartOutline) })
  },
  {
    label: '账套初始化',
    key: 'Setup',
    icon: () => h(NIcon, null, { default: () => h(SettingsOutline) })
  }
]

const userMenuOptions = [
  {
    label: '退出登录',
    key: 'logout',
    icon: () => h(NIcon, null, { default: () => h(LogOutOutline) })
  }
]

const handleMenuSelect = (key: string) => {
  router.push({ name: key })
}

const handleUserMenuSelect = async (key: string) => {
  if (key === 'logout') {
    await authStore.clearAuth()
    router.push('/login')
  }
}
</script>