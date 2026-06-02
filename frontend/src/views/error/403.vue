<template>
  <div class="error-page">
    <h1>403</h1>
    <p>抱歉，您没有权限访问此页面</p>
    <el-button
      type="primary"
      @click="$router.push('/')"
    >
      返回首页
    </el-button>
  </div>
</template>

<style scoped lang="scss">
.error-page {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100vh;
  h1 { font-size: 72px; color: #ff4d4f; margin-bottom: 16px; }
  p { font-size: 16px; color: #666; margin-bottom: 24px; }
}
</style>
