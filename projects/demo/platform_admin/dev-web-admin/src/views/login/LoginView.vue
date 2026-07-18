<!--
  登录页
  - 居中卡片布局，标题「快速开发平台」
  - el-form + el-input(用户名/密码/验证码) + el-button
  - 表单校验(非空)，调用 authStore.login()
  - 单租户：自动调用 selectTenant 后跳转首页
  - 多租户：跳转 /tenant-select 选择页
-->
<script setup lang="ts">
import { ref, reactive, onMounted, nextTick } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/store/modules/auth'
import { useUserStore } from '@/store/modules/user'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()
const userStore = useUserStore()

const formRef = ref<FormInstance>()
const passwordRef = ref()
const loading = ref(false)
const captchaImg = ref('')

const loginForm = reactive({
  username: '',
  password: '',
  captchaCode: '',
  captchaKey: ''
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
}

/** 加载验证码（如后端启用验证码则使用） */
async function loadCaptcha() {
  try {
    const res = await userStore.fetchCaptcha()
    captchaImg.value = res.data
    loginForm.captchaKey = res.id
  } catch {
    captchaImg.value = ''
  }
}

/** 提交登录 */
async function handleLogin() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  loading.value = true
  try {
    const result = await authStore.login(loginForm.username, loginForm.password)

    if (result.tenants.length === 1) {
      // 单租户：后端已直接返回 access_token，无需再调 selectTenant
      ElMessage.success('登录成功')
      const redirect = (route.query.redirect as string) || '/home'
      router.replace(redirect)
    } else if (result.tenants.length > 1) {
      // 多租户：跳转租户选择页
      router.replace('/tenant-select')
    } else {
      // 无租户（异常情况）
      ElMessage.warning('当前账号无可用租户，请联系管理员')
      authStore.logout()
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '登录失败，请重试')
    loadCaptcha()
    loginForm.captchaCode = ''
    loginForm.password = ''
    nextTick(() => passwordRef.value?.focus())
  } finally {
    loading.value = false
  }
}

onMounted(loadCaptcha)
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <h1 class="login-title">快速开发平台</h1>
      <p class="login-subtitle">PC 管理端</p>

      <el-form ref="formRef" :model="loginForm" :rules="rules" size="large" @keyup.enter="handleLogin">
        <el-form-item prop="username">
          <el-input v-model="loginForm.username" placeholder="请输入用户名" prefix-icon="User" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input ref="passwordRef" v-model="loginForm.password" type="password" placeholder="请输入密码" prefix-icon="Lock" show-password />
        </el-form-item>
        <el-form-item v-if="captchaImg" prop="captchaCode">
          <div class="captcha-row">
            <el-input v-model="loginForm.captchaCode" placeholder="请输入验证码" style="flex: 1" />
            <img :src="captchaImg" class="captcha-img" alt="验证码" @click="loadCaptcha" />
          </div>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" :loading="loading" class="login-btn" @click="handleLogin">登录</el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100vh;
  background-color: var(--color-bg-page);
}
.login-card {
  width: 400px;
  padding: 40px;
  background-color: var(--color-bg-white);
  border-radius: var(--border-radius-md);
  box-shadow: var(--shadow-lg);
}
.login-title {
  text-align: center;
  font-size: 24px;
  font-weight: 600;
  color: var(--color-primary);
  margin-bottom: var(--spacing-sm);
}
.login-subtitle {
  text-align: center;
  font-size: 14px;
  color: var(--color-text-secondary);
  margin-bottom: var(--spacing-xxl);
}
.login-btn { width: 100%; }
.captcha-row {
  display: flex;
  gap: 8px;
  width: 100%;
}
.captcha-img {
  height: 40px;
  cursor: pointer;
  border-radius: 4px;
  border: 1px solid #dcdfe6;
}
</style>
