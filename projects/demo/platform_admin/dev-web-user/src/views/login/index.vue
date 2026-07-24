<script setup lang="ts">
/**
 * 登录页 — 响应式适配（PC 卡片居中 / 移动端全屏）
 * tenant_code 通过 API 从域名自动解析，用户无感知
 */
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { sendCode } from '@/api/auth'
import { useUserStore } from '@/stores/user'
import { resolveTenantCodeAsync, resolveTenantCode } from '@/utils/tenant'

const router = useRouter()
const route  = useRoute()
const userStore = useUserStore()

// 租户解析状态
const resolvingTenant = ref(true)
const tenantCode = ref('')

// 页面加载时异步解析租户
onMounted(async () => {
  try {
    const result = await resolveTenantCodeAsync()
    tenantCode.value = result.tenantCode
  } catch {
    tenantCode.value = resolveTenantCode()
  } finally {
    resolvingTenant.value = false
  }
})

const form = ref({ phone: '', code: '' })
const loading     = ref(false)
const sendingCode = ref(false)
const countdown   = ref(0)
let timer: ReturnType<typeof setInterval> | null = null

async function handleSendCode() {
  if (!form.value.phone) { ElMessage.warning('请输入手机号'); return }
  if (countdown.value > 0) return

  sendingCode.value = true
  try {
    await sendCode(form.value.phone, tenantCode.value)
    ElMessage.success('验证码已发送')
    countdown.value = 60
    timer = setInterval(() => {
      countdown.value--
      if (countdown.value <= 0) { clearInterval(timer!); timer = null }
    }, 1000)
  } catch {
    // 拦截器已处理
  } finally {
    sendingCode.value = false
  }
}

async function handleLogin() {
  if (!form.value.phone) { ElMessage.warning('请输入手机号'); return }
  if (!form.value.code)  { ElMessage.warning('请输入验证码'); return }

  loading.value = true
  try {
    await userStore.login(form.value.phone, form.value.code, tenantCode.value)
    ElMessage.success('登录成功')
    router.replace((route.query.redirect as string) || '/')
  } catch {
    // 拦截器已处理
  } finally {
    loading.value = false
  }
}

onUnmounted(() => { if (timer) clearInterval(timer) })
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <div class="login-header">
        <div class="login-logo">U</div>
        <h2 class="login-title">用户登录</h2>
        <p class="login-subtitle">请使用手机验证码登录</p>
      </div>

      <el-form :model="form" label-width="0" class="login-form">
        <!-- 手机号 -->
        <el-form-item>
          <el-input
            v-model="form.phone"
            placeholder="请输入手机号"
            prefix-icon="Phone"
            maxlength="11"
            size="large"
            inputmode="tel"
            :disabled="resolvingTenant"
          />
        </el-form-item>

        <!-- 验证码行 -->
        <el-form-item>
          <div class="code-row">
            <el-input
              v-model="form.code"
              placeholder="请输入验证码"
              prefix-icon="Message"
              maxlength="6"
              size="large"
              inputmode="numeric"
              :disabled="resolvingTenant"
              @keyup.enter="handleLogin"
            />
            <el-button
              type="primary"
              size="large"
              :disabled="resolvingTenant || countdown > 0 || sendingCode"
              :loading="sendingCode"
              class="send-btn"
              @click="handleSendCode"
            >
              {{ countdown > 0 ? `${countdown}s` : '获取验证码' }}
            </el-button>
          </div>
        </el-form-item>

        <!-- 登录按钮 -->
        <el-form-item>
          <el-button
            type="primary"
            size="large"
            :loading="loading"
            :disabled="resolvingTenant"
            class="login-btn"
            @click="handleLogin"
          >
            {{ resolvingTenant ? '加载中…' : '登 录' }}
          </el-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<style scoped>
/* ─── 页面背景 ─── */
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  min-height: 100dvh; /* 修复 iOS 动态地址栏 */
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  padding: 16px;
  box-sizing: border-box;
}

/* ─── 卡片 ─── */
.login-card {
  width: 100%;
  max-width: 420px;
  background: #fff;
  border-radius: 16px;
  padding: 40px 36px 32px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.2);
  box-sizing: border-box;
}

/* ─── 头部 ─── */
.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-logo {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: #fff;
  font-size: 24px;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 16px;
}

.login-title {
  margin: 0 0 6px;
  font-size: 22px;
  font-weight: 700;
  color: #1a1a2e;
}

.login-subtitle {
  margin: 0;
  font-size: 13px;
  color: #999;
}

/* ─── 表单 ─── */
.login-form {
  margin-top: 0;
}

.code-row {
  display: flex;
  width: 100%;
  gap: 10px;
}

.code-row .el-input {
  flex: 1;
  min-width: 0;
}

.send-btn {
  flex-shrink: 0;
  width: 110px;
  font-size: 13px;
  padding: 0 10px;
}

.login-btn {
  width: 100%;
  height: 48px;
  font-size: 16px;
  border-radius: 8px;
  margin-top: 4px;
}

/* ─── 移动端断点（≤480px）：去除多余 padding，输入框更高 ─── */
@media (max-width: 480px) {
  .login-page {
    align-items: flex-start;
    padding-top: 60px;
  }

  .login-card {
    border-radius: 20px;
    padding: 32px 24px 28px;
    box-shadow: 0 10px 40px rgba(0, 0, 0, 0.18);
  }

  .login-logo {
    width: 52px;
    height: 52px;
    font-size: 22px;
  }

  .login-title {
    font-size: 20px;
  }

  /* 加大触摸目标 */
  .login-btn {
    height: 52px;
    font-size: 17px;
  }

  .send-btn {
    width: 96px;
    font-size: 12px;
  }
}

/* ─── 超小屏（≤360px） ─── */
@media (max-width: 360px) {
  .login-page {
    padding: 40px 12px 12px;
  }

  .login-card {
    padding: 28px 16px 24px;
  }
}
</style>
