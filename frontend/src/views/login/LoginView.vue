<template>
  <div class="login-container">
    <div class="login-card">
      <h1 class="login-title">慧财智能财务平台</h1>
      <p class="login-subtitle">银行流水驱动业财一体化</p>
      <el-form ref="formRef" :model="form" :rules="rules" class="login-form" @keyup.enter="handleLogin">
        <el-form-item prop="account">
          <el-input v-model="form.account" placeholder="账号" size="large">
            <template #prefix><el-icon><User /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码" size="large" show-password>
            <template #prefix><el-icon><Lock /></el-icon></template>
          </el-input>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" size="large" class="login-btn" :loading="loading" @click="handleLogin">
            {{ loading ? '登录中...' : '登 录' }}
          </el-button>
        </el-form-item>
      </el-form>
      <p v-if="errorMsg" class="login-error">{{ errorMsg }}</p>
      <div class="login-hint">
        <p>演示账号：</p>
        <p><b>admin</b> / admin123（财务主管）</p>
        <p><b>cashier</b> / 123456（出纳）</p>
        <p><b>boss</b> / 123456（老板）</p>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import type { FormInstance } from 'element-plus'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth.store'
import { generateToken, verifyTokenAndFetchUser } from '@/utils/jwt'

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

async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  errorMsg.value = ''

  try {
    // 签发 JWT（与后端校验算法一致）
    const result = await generateToken(form.account, form.password)
    if (!result) {
      errorMsg.value = '账号或密码错误'
      loading.value = false
      return
    }

    // 存入 Store + localStorage
    authStore.setAuth(result.token, result.user)

    // 验证：调用后端 API 测试 token 是否有效
    try {
      const resp = await fetch('/api/v1/accounts/tree', {
        headers: { Authorization: `Bearer ${result.token}` },
      })
      if (resp.ok) {
        ElMessage.success('已连接后端服务')
      }
    } catch {
      // 后端不可用时仍可继续（Mock 模式兜底）
    }

    const redirect = (route.query.redirect as string) || '/dashboard'
    router.push(redirect)
  } catch (e) {
    errorMsg.value = '登录失败，请重试'
  } finally {
    loading.value = false
  }
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
  margin-bottom: 4px;
  color: #333;
}
.login-subtitle {
  text-align: center;
  color: #999;
  margin-bottom: 32px;
  font-size: 13px;
}
.login-form .login-btn {
  width: 100%;
}
.login-error {
  color: #ff4d4f;
  text-align: center;
  margin-top: 16px;
  font-size: 13px;
}
.login-hint {
  margin-top: 24px;
  padding: 12px;
  background: #f5f5f5;
  border-radius: 4px;
  font-size: 12px;
  color: #999;
  p { margin: 2px 0; }
  b { color: #333; }
}
</style>
