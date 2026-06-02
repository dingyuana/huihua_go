<template>
  <div class="page-container">
    <div class="page-header">
      <div class="page-title">
        <div
          v-if="icon"
          class="page-icon"
          :style="{ background: iconBg || 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)' }"
        >
          <slot name="icon">
            {{ icon }}
          </slot>
        </div>
        <div>
          <h3>{{ title }}</h3>
          <p
            v-if="subtitle"
            class="page-subtitle"
          >
            {{ subtitle }}
          </p>
        </div>
      </div>
      <div class="page-actions">
        <slot name="actions" />
      </div>
    </div>
    <div class="page-body">
      <slot />
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  title: string
  icon?: string
  iconBg?: string
  subtitle?: string
}>()
</script>

<style scoped lang="scss">
.page-container {
  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 20px;

    .page-title {
      display: flex;
      align-items: center;
      gap: 14px;

      .page-icon {
        display: flex;
        align-items: center;
        justify-content: center;
        width: 40px;
        height: 40px;
        border-radius: 10px;
        color: #fff;
        font-size: 18px;
        flex-shrink: 0;
      }

      h3 {
        font-size: 17px;
        font-weight: 600;
        color: #1d2129;
        margin: 0 0 2px;
      }

      .page-subtitle {
        font-size: 13px;
        color: #86909c;
        margin: 0;
      }
    }

    .page-actions {
      display: flex;
      gap: 10px;
    }
  }

  .page-body {
    border-radius: 10px;
    background: #fff;
    box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
    overflow: hidden;
  }
}
</style>
