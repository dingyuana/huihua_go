<template>
  <div class="error-page">
    <h1>404</h1>
    <p>页面不存在</p>
    <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
  </div>
</template>

<style scoped lang="scss">
.error-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  h1 { font-size: 72px; color: #faad14; margin-bottom: 16px; }
  p { font-size: 16px; color: #666; margin-bottom: 24px; }
}
</style>
