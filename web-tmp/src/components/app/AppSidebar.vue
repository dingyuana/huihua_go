<template>
  <div :class="['sidebar', { collapsed: appStore.sidebarCollapsed }]">
    <div class="sidebar-logo">
      <span v-if="!appStore.sidebarCollapsed">慧财财务</span>
      <span v-else>慧</span>
    </div>
    <el-menu
      :default-active="route.path"
      :collapse="appStore.sidebarCollapsed"
      :router="true"
      class="sidebar-menu"
    >
      <template v-for="item in menuItems" :key="item.path">
        <el-menu-item v-if="!item.children" :index="item.path">
          <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
          <template #title>{{ item.title }}</template>
        </el-menu-item>
        <el-sub-menu v-else :index="item.path">
          <template #title>
            <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
            <span>{{ item.title }}</span>
          </template>
          <el-menu-item v-for="child in item.children" :key="child.path" :index="child.path">
            {{ child.title }}
          </el-menu-item>
        </el-sub-menu>
      </template>
    </el-menu>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app.store'
import { useAuthStore } from '@/stores/auth.store'
import { roleMenuMap } from '@/config/menu.config'
import type { Role } from '@/types/enums'

const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()

const menuItems = computed(() => {
  const role = (authStore.user?.role || 'admin') as Role
  return roleMenuMap[role] || []
})
</script>

<style scoped lang="scss">
.sidebar {
  width: 200px;
  height: 100vh;
  background: #001529;
  transition: width 0.2s;
  overflow: hidden;

  &.collapsed {
    width: 64px;
  }

  .sidebar-logo {
    height: 48px;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #fff;
    font-size: 16px;
    font-weight: 600;
    background: rgba(0, 0, 0, 0.2);
  }

  .sidebar-menu {
    border-right: none;
    height: calc(100vh - 48px);
    overflow-y: auto;
  }
}
</style>
