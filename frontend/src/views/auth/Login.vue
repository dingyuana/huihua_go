<template>
  <div class="login-container">
    <n-card style="width: 400px;" :bordered="false" size="large">
      <template #header>
        <div style="text-align: center;">
          <h1 style="font-size: 24px; margin: 0;">慧话财务系统</h1>
          <p style="color: #666; margin-top: 8px;">请登录您的账户</p>
        </div>
      </template>

      <n-form ref="formRef" :model="formValue" :rules="rules" size="large">
        <n-form-item path="username" label="用户名">
          <n-input v-model:value="formValue.username" placeholder="请输入用户名" @keydown.enter="handleLogin">
            <template #prefix>
              <n-icon><person-outline /></n-icon>
            </template>
          </n-input>
        </n-form-item>

        <n-form-item path="password" label="密码">
          <n-input
            v-model:value="formValue.password"
            type="password"
            placeholder="请输入密码"
            show-password-on="click"
            @keydown.enter="handleLogin"
          >
            <template #prefix>
              <n-icon><lock-closed-outline /></n-icon>
            </template>
          </n-input>
        </n-form-item>
      </n-form>

      <template #footer>
        <n-button
          type="primary"
          block
          :loading="loading"
          @click="handleLogin"
        >
          登录
        </n-button>
      </template>

      <template #action>
        <div style="text-align: center; color: #999; font-size: 12px;">
          测试账号: testuser / admin123
        </div>
      </template>
    </n-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { NCard, NForm, NFormItem, NInput, NButton, NIcon, useMessage, useDialog } from 'naive-ui'
import { PersonOutline, LockClosedOutline } from '@vicons/ionicons5'
import { authService } from '@/api/auth'

const router = useRouter()
const message = useMessage()

const formRef = ref()
const loading = ref(false)
const formValue = ref({
  username: '',
  password: ''
})

const rules = {
  username: { required: true, message: '请输入用户名', trigger: 'blur' },
  password: { required: true, message: '请输入密码', trigger: 'blur' }
}

const handleLogin = async () => {
  if (loading.value) return

  try {
    await formRef.value?.validate()
  } catch {
    return
  }

  loading.value = true
  try {
    await authService.login(formValue.value)
    message.success('登录成功')
    router.push('/dashboard')
  } catch (error: any) {
    message.error(error?.response?.data?.message || '登录失败，请检查用户名和密码')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
</style>