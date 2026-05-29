<template>
  <header class="app-header">
    <div class="header-left">
      <el-button text @click="appStore.toggleSidebar">
        <el-icon :size="20">
          <component :is="appStore.sidebarCollapsed ? 'Expand' : 'Fold'" />
        </el-icon>
      </el-button>
      <el-breadcrumb class="header-breadcrumb">
        <el-breadcrumb-item :to="{ path: '/' }">首页</el-breadcrumb-item>
        <el-breadcrumb-item v-if="route.meta.title">{{ route.meta.title }}</el-breadcrumb-item>
      </el-breadcrumb>
    </div>
    <div class="header-right">
      <el-tag v-if="tenantStore.watermark" type="warning" size="small" effect="dark">
        {{ tenantStore.watermark }}
      </el-tag>
      <el-dropdown trigger="click">
        <span class="header-user">
          <el-avatar :size="28" icon="UserFilled" />
          <span class="user-name">{{ authStore.user?.name || '用户' }}</span>
        </span>
        <template #dropdown>
          <el-dropdown-menu>
            <el-dropdown-item @click="handleLogout">退出登录</el-dropdown-item>
          </el-dropdown-menu>
        </template>
      </el-dropdown>
    </div>
  </header>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app.store'
import { useAuthStore } from '@/stores/auth.store'
import { useTenantStore } from '@/stores/tenant.store'

const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()
const tenantStore = useTenantStore()

function handleLogout() {
  authStore.logout()
}
</script>

<style scoped lang="scss">
.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 48px;
  padding: 0 16px;
  background: #fff;
  border-bottom: 1px solid #e8e8e8;
  flex-shrink: 0;

  .header-left {
    display: flex;
    align-items: center;
    gap: 12px;
  }

  .header-breadcrumb {
    font-size: 13px;
  }

  .header-right {
    display: flex;
    align-items: center;
    gap: 16px;
  }

  .header-user {
    display: flex;
    align-items: center;
    gap: 8px;
    cursor: pointer;

    .user-name {
      font-size: 13px;
      color: #333;
    }
  }
}
</style>
