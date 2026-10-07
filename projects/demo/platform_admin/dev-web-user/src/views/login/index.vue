<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { sendCode } from '@/api/auth'
import { useUserStore } from '@/stores/user'
import { resolveTenantCodeAsync, resolveTenantCode } from '@/utils/tenant'

const router = useRouter()
const route  = useRoute()
const userStore = useUserStore()

const resolvingTenant = ref(true)
const tenantCode = ref('')
const tenantName = ref('')

onMounted(async () => {
  try {
    const result = await resolveTenantCodeAsync()
    tenantCode.value = result.tenantCode
    tenantName.value = result.tenantName
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

const canSend = computed(() => !resolvingTenant.value && !sendingCode.value && countdown.value === 0 && form.value.phone.length === 11)

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
    await userStore.login(form.value.phone, form.value.code, tenantCode.value, tenantName.value)
    ElMessage.success('登录成功')
    router.replace((route.query.redirect as string) || '/')
  } catch {
    // 拦截器已处理
  } finally {
    loading.value = false
  }
}

onUnmounted(() => { if (timer) clearInterval(timer) })

// ── 微信扫码登录 ──
const wechatQRLoading = ref(false)

async function handleWechatQRLogin() {
  wechatQRLoading.value = true
  try {
    const res = await fetch(`/api/v1/user/public/wechat/qr?tenant_id=${encodeURIComponent(tenantCode.value || '')}`)
    const data = await res.json()
    if (data.code === 0 && data.data?.auth_url) {
      window.location.href = data.data.auth_url
    } else {
      ElMessage.error(data.msg || '微信扫码登录暂不可用')
    }
  } catch {
    ElMessage.error('微信扫码登录暂不可用')
  } finally {
    wechatQRLoading.value = false
  }
}
</script>

<template>
  <div class="login-root">
    <!-- 左侧品牌区（PC 才显示） -->
    <div class="login-brand" aria-hidden="true">
      <div class="brand-inner">
        <div class="brand-logo">
          <svg width="48" height="48" viewBox="0 0 48 48" fill="none">
            <rect width="48" height="48" rx="14" fill="white" fill-opacity="0.15"/>
            <path d="M14 24C14 18.477 18.477 14 24 14s10 4.477 10 10-4.477 10-10 10S14 29.523 14 24z" fill="white" fill-opacity="0.3"/>
            <path d="M20 24l3 3 6-6" stroke="white" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <h1 class="brand-name">用户工作台</h1>
        <p class="brand-desc">安全、便捷的一站式服务平台</p>
        <div class="brand-features">
          <div class="feature-item" v-for="f in ['手机号快捷登录', '多租户隔离', '权限精细管控']" :key="f">
            <svg width="16" height="16" viewBox="0 0 16 16" fill="none">
              <circle cx="8" cy="8" r="8" fill="white" fill-opacity="0.2"/>
              <path d="M5 8l2 2 4-4" stroke="white" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <span>{{ f }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- 右侧表单区 -->
    <div class="login-form-area">
      <div class="login-card">
        <!-- 移动端 logo -->
        <div class="mobile-logo">
          <div class="mobile-logo-icon">U</div>
        </div>

        <div class="card-header">
          <h2 class="card-title">欢迎回来</h2>
          <p class="card-subtitle">{{ resolvingTenant ? '正在解析租户…' : '使用手机验证码安全登录' }}</p>
        </div>

        <div class="form-group">
          <label class="form-label" for="phone">手机号</label>
          <div class="input-wrapper">
            <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none">
              <path d="M6.62 10.79c1.44 2.83 3.76 5.14 6.59 6.59l2.2-2.2c.27-.27.67-.36 1.02-.24 1.12.37 2.33.57 3.57.57.55 0 1 .45 1 1V20c0 .55-.45 1-1 1-9.39 0-17-7.61-17-17 0-.55.45-1 1-1h3.5c.55 0 1 .45 1 1 0 1.25.2 2.45.57 3.57.11.35.03.74-.25 1.02l-2.2 2.2z" fill="currentColor"/>
            </svg>
            <input
              id="phone"
              v-model="form.phone"
              type="tel"
              inputmode="tel"
              maxlength="11"
              placeholder="请输入手机号"
              class="form-input"
              :disabled="resolvingTenant"
              autocomplete="tel"
            />
          </div>
        </div>

        <div class="form-group">
          <label class="form-label" for="code">验证码</label>
          <div class="code-row">
            <div class="input-wrapper code-input-wrapper">
              <svg class="input-icon" width="18" height="18" viewBox="0 0 24 24" fill="none">
                <path d="M20 4H4c-1.1 0-2 .9-2 2v12c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z" fill="currentColor"/>
              </svg>
              <input
                id="code"
                v-model="form.code"
                type="text"
                inputmode="numeric"
                maxlength="6"
                placeholder="请输入验证码"
                class="form-input"
                :disabled="resolvingTenant"
                autocomplete="one-time-code"
                @keyup.enter="handleLogin"
              />
            </div>
            <button
              class="send-btn"
              :class="{ 'send-btn--loading': sendingCode, 'send-btn--counting': countdown > 0 }"
              :disabled="!canSend"
              @click="handleSendCode"
            >
              <span v-if="sendingCode" class="spinner" />
              <span v-else-if="countdown > 0">{{ countdown }}s</span>
              <span v-else>获取验证码</span>
            </button>
          </div>
        </div>

        <button
          class="login-btn"
          :class="{ 'login-btn--loading': loading }"
          :disabled="resolvingTenant || loading"
          @click="handleLogin"
        >
          <span v-if="loading" class="spinner spinner--white" />
          <span v-else-if="resolvingTenant">加载中…</span>
          <span v-else>登&nbsp;&nbsp;录</span>
        </button>

        <p class="login-tip">登录即代表您同意<a href="#" class="link">服务协议</a>和<a href="#" class="link">隐私政策</a></p>

        <!-- 微信扫码登录入口（仅 PC，配置了才显示）-->
        <div class="divider"><span>或</span></div>
        <button class="wechat-btn" :disabled="wechatQRLoading" @click="handleWechatQRLogin">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor" style="margin-right:6px;flex-shrink:0">
            <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c-.156-.53-.24-1.082-.24-1.65 0-3.679 3.595-6.662 8.024-6.662.27 0 .534.015.796.04C16.084 4.957 12.679 2.188 8.69 2.188zm-1.947 3.87a1.094 1.094 0 1 1 0 2.188 1.094 1.094 0 0 1 0-2.188zm4.55 0a1.094 1.094 0 1 1 0 2.188 1.094 1.094 0 0 1 0-2.188zM23.3 15.375c0-3.379-3.237-6.12-7.228-6.12-4.001 0-7.24 2.741-7.24 6.12 0 3.38 3.239 6.12 7.24 6.12.774 0 1.518-.103 2.218-.293a.682.682 0 0 1 .564.075l1.487.873a.254.254 0 0 0 .131.044.23.23 0 0 0 .229-.23c0-.057-.024-.11-.037-.167l-.307-1.155a.465.465 0 0 1 .165-.522C22.254 19.228 23.3 17.385 23.3 15.375zm-9.565-.984a.86.86 0 1 1 0-1.72.86.86 0 0 1 0 1.72zm4.354 0a.86.86 0 1 1 0-1.72.86.86 0 0 1 0 1.72z"/>
          </svg>
          <span>{{ wechatQRLoading ? '跳转中...' : '微信扫码登录' }}</span>
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* ── 根布局 ── */
.login-root {
  display: flex;
  min-height: 100vh;
  min-height: 100dvh;
  background: #f0f2f5;
}

/* ── 左侧品牌区 ── */
.login-brand {
  display: none;
  flex: 0 0 420px;
  background: linear-gradient(145deg, #4f46e5 0%, #7c3aed 50%, #a855f7 100%);
  padding: 48px 40px;
  position: relative;
  overflow: hidden;
}

.login-brand::before {
  content: '';
  position: absolute;
  inset: 0;
  background: radial-gradient(ellipse at 30% 20%, rgba(255,255,255,0.1) 0%, transparent 60%),
              radial-gradient(ellipse at 80% 80%, rgba(255,255,255,0.08) 0%, transparent 50%);
}

@media (min-width: 960px) {
  .login-brand { display: flex; align-items: center; justify-content: center; }
}

.brand-inner {
  position: relative;
  z-index: 1;
}

.brand-logo {
  margin-bottom: 24px;
}

.brand-name {
  font-size: 28px;
  font-weight: 700;
  color: #fff;
  margin: 0 0 12px;
  letter-spacing: -0.3px;
}

.brand-desc {
  font-size: 15px;
  color: rgba(255,255,255,0.75);
  margin: 0 0 40px;
  line-height: 1.6;
}

.brand-features {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 10px;
  color: rgba(255,255,255,0.9);
  font-size: 14px;
}

/* ── 右侧表单区 ── */
.login-form-area {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px 20px;
}

.login-card {
  width: 100%;
  max-width: 400px;
  background: #fff;
  border-radius: 20px;
  padding: 40px 36px 32px;
  box-shadow: 0 4px 24px rgba(0,0,0,0.08), 0 1px 4px rgba(0,0,0,0.04);
}

/* ── 移动端 logo ── */
.mobile-logo {
  display: flex;
  justify-content: center;
  margin-bottom: 24px;
}

@media (min-width: 960px) {
  .mobile-logo { display: none; }
}

.mobile-logo-icon {
  width: 56px;
  height: 56px;
  border-radius: 16px;
  background: linear-gradient(135deg, #4f46e5, #7c3aed);
  color: #fff;
  font-size: 24px;
  font-weight: 700;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 20px rgba(79,70,229,0.35);
}

/* ── 卡片头部 ── */
.card-header {
  margin-bottom: 28px;
}

.card-title {
  font-size: 22px;
  font-weight: 700;
  color: #111827;
  margin: 0 0 6px;
  letter-spacing: -0.3px;
}

.card-subtitle {
  font-size: 14px;
  color: #6b7280;
  margin: 0;
}

/* ── 表单 ── */
.form-group {
  margin-bottom: 20px;
}

.form-label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: #374151;
  margin-bottom: 8px;
}

.input-wrapper {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 12px;
  color: #9ca3af;
  pointer-events: none;
  flex-shrink: 0;
}

.form-input {
  width: 100%;
  height: 44px;
  padding: 0 14px 0 40px;
  border: 1.5px solid #e5e7eb;
  border-radius: 10px;
  font-size: 15px;
  color: #111827;
  background: #fafafa;
  outline: none;
  transition: border-color 0.15s, box-shadow 0.15s, background 0.15s;
  box-sizing: border-box;
}

.form-input::placeholder { color: #d1d5db; }

.form-input:focus {
  border-color: #4f46e5;
  background: #fff;
  box-shadow: 0 0 0 3px rgba(79,70,229,0.12);
}

.form-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* ── 验证码行 ── */
.code-row {
  display: flex;
  gap: 10px;
}

.code-input-wrapper {
  flex: 1;
  min-width: 0;
}

.send-btn {
  flex-shrink: 0;
  width: 110px;
  height: 44px;
  border: 1.5px solid #4f46e5;
  border-radius: 10px;
  background: #fff;
  color: #4f46e5;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;
  white-space: nowrap;
}

.send-btn:hover:not(:disabled) {
  background: #f0effe;
}

.send-btn:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.send-btn--counting {
  border-color: #d1d5db;
  color: #9ca3af;
}

/* ── 登录按钮 ── */
.login-btn {
  width: 100%;
  height: 48px;
  margin-top: 8px;
  border: none;
  border-radius: 12px;
  background: linear-gradient(135deg, #4f46e5, #7c3aed);
  color: #fff;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.15s, transform 0.1s, box-shadow 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;
  letter-spacing: 0.5px;
  box-shadow: 0 4px 16px rgba(79,70,229,0.4);
}

.login-btn:hover:not(:disabled) {
  opacity: 0.92;
  transform: translateY(-1px);
  box-shadow: 0 6px 20px rgba(79,70,229,0.45);
}

.login-btn:active:not(:disabled) {
  transform: translateY(0);
}

.login-btn:disabled {
  opacity: 0.55;
  cursor: not-allowed;
  transform: none;
  box-shadow: none;
}

/* ── 底部提示 ── */
.login-tip {
  margin: 20px 0 0;
  text-align: center;
  font-size: 12px;
  color: #9ca3af;
}

.link {
  color: #4f46e5;
  text-decoration: none;
}
.link:hover { text-decoration: underline; }

/* ── Spinner ── */
.spinner {
  display: inline-block;
  width: 18px;
  height: 18px;
  border: 2px solid rgba(79,70,229,0.25);
  border-top-color: #4f46e5;
  border-radius: 50%;
  animation: spin 0.7s linear infinite;
}

.spinner--white {
  border-color: rgba(255,255,255,0.3);
  border-top-color: #fff;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* ── 移动端适配 ── */
@media (max-width: 480px) {
  .login-form-area {
    align-items: center;
    padding-top: 24px;
    background: linear-gradient(160deg, #4f46e5 0%, #7c3aed 30%, #f0f2f5 60%);
  }

  .login-card {
    padding: 32px 24px 28px;
    border-radius: 24px;
    box-shadow: 0 20px 60px rgba(0,0,0,0.15);
  }

  .mobile-logo { display: flex; }
  .card-header { margin-bottom: 24px; }
}

/* ── 分隔线 ── */
.divider {
  display: flex;
  align-items: center;
  gap: 12px;
  margin: 16px 0 12px;
  color: #9ca3af;
  font-size: 13px;
}
.divider::before, .divider::after {
  content: '';
  flex: 1;
  height: 1px;
  background: #e5e7eb;
}

/* ── 微信扫码按钮 ── */
.wechat-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  padding: 11px 0;
  border: 1px solid #07c160;
  border-radius: 10px;
  background: #fff;
  color: #07c160;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  transition: background .15s, color .15s;
}
.wechat-btn:hover:not(:disabled) { background: #07c160; color: #fff; }
.wechat-btn:disabled { opacity: .5; cursor: not-allowed; }
</style>
