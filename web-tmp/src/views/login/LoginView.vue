<template>
  <div class="login-container">
    <div class="login-card">
      <h1 class="login-title">慧财智能财务平台</h1>
      <p class="login-subtitle">银行流水驱动业财一体化</p>
      <el-form ref="formRef" :model="form" :rules="rules" class="login-form" @keyup.enter="handleLogin">
        <el-form-item prop="account">
          <el-input v-model="form.account" placeholder="账号" size="large" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" size="large" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="large" class="login-btn" :loading="loading" @click="handleLogin">
            登 录
          </el-button>
        </el-form-item>
      </el-form>
      <p v-if="errorMsg" class="login-error">{{ errorMsg }}</p>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import type { FormInstance } from 'element-plus'
import { useAuthStore } from '@/stores/auth.store'
import type { User } from '@/types/models/user'
import { Role } from '@/types/enums'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const formRef = ref<FormInstance>()
const loading = ref(false)
const errorMsg = ref('')

const form = reactive({
  account: '',
  password: '',
})

const rules = {
  account: [{ required: true, message: '请输入账号', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

function handleLogin() {
  formRef.value?.validate((valid) => {
    if (!valid) return
    loading.value = true
    errorMsg.value = ''

    // Mock 登录
    setTimeout(() => {
      const mockUser: User = {
        id: 'user-001',
        name: '张三',
        email: 'admin@example.com',
        role: Role.Admin,
        permissions: [
          'account:read', 'account:write',
          'voucher:read', 'voucher:write', 'voucher:submit', 'voucher:reverse',
          'bank:import', 'bank:classify',
          'report:read',
        ],
      }
      const mockToken = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.mock-token'
      authStore.setAuth(mockToken, mockUser)

      const redirect = (route.query.redirect as string) || '/dashboard'
      router.push(redirect)
      loading.value = false
    }, 800)
  })
}
</script>

<style scoped lang="scss">
.login-container {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100vh;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
.login-card {
  width: 400px;
  padding: 40px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
}
.login-title {
  font-size: 24px;
  text-align: center;
  margin-bottom: 8px;
  color: #333;
}
.login-subtitle {
  text-align: center;
  color: #999;
  margin-bottom: 32px;
  font-size: 14px;
}
.login-form {
  .login-btn {
    width: 100%;
  }
}
.login-error {
  color: #ff4d4f;
  text-align: center;
  margin-top: 16px;
  font-size: 13px;
}
</style>
