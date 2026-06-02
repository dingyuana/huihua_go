<template>
  <div :class="['sidebar', { collapsed: appStore.sidebarCollapsed }]">
    <div class="sidebar-logo">
      <div
        v-if="!appStore.sidebarCollapsed"
        class="logo-content"
      >
        <span class="logo-brand">慧财财务</span>
        <span
          v-if="companyName"
          class="logo-company"
        >{{ companyName }}</span>
      </div>
      <span
        v-else
        class="logo-icon"
      >慧</span>
    </div>
    <el-menu
      :default-active="route.path"
      :collapse="appStore.sidebarCollapsed"
      :router="true"
      class="sidebar-menu"
    >
      <template
        v-for="item in menuItems"
        :key="item.path"
      >
        <el-menu-item
          v-if="!item.children"
          :index="item.path"
        >
          <el-icon v-if="item.icon">
            <component :is="item.icon" />
          </el-icon>
          <template #title>
            {{ item.title }}
          </template>
        </el-menu-item>
        <el-sub-menu
          v-else
          :index="item.path"
        >
          <template #title>
            <el-icon v-if="item.icon">
              <component :is="item.icon" />
            </el-icon>
            <span>{{ item.title }}</span>
          </template>
          <el-menu-item
            v-for="child in item.children"
            :key="child.path"
            :index="child.path"
          >
            {{ child.title }}
          </el-menu-item>
        </el-sub-menu>
      </template>
    </el-menu>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAppStore } from '@/stores/app.store'
import { useAuthStore } from '@/stores/auth.store'
import { roleMenuMap } from '@/config/menu.config'
import type { Role } from '@/types/enums'
import request from '@/api/request'

const route = useRoute()
const appStore = useAppStore()
const authStore = useAuthStore()

const companyName = ref('')

const menuItems = computed(() => {
  const role = (authStore.user?.role || 'admin') as Role
  return roleMenuMap[role] || []
})

async function loadCompany() {
  try {
    const res = await request.get('/account-setup/status')
    const data = res.data || res
    if (data.company?.company_name) {
      companyName.value = data.company.company_name
    }
  } catch {
    // Ignore errors
  }
}

onMounted(loadCompany)
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
    background: rgba(0, 0, 0, 0.2);

    .logo-content {
      display: flex;
      flex-direction: column;
      align-items: center;
      line-height: 1.2;

      .logo-brand {
        font-size: 14px;
        font-weight: 600;
      }

      .logo-company {
        font-size: 11px;
        color: rgba(255, 255, 255, 0.65);
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
        max-width: 160px;
      }
    }

    .logo-icon {
      font-size: 18px;
      font-weight: 600;
    }
  }

  .sidebar-menu {
    border-right: none;
    height: calc(100vh - 48px);
    overflow-y: auto;
  }
}
</style>
