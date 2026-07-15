<template>
  <div class="init-container">
    <div class="init-card">
      <h2 class="init-title">系统初始化</h2>
      <p class="init-desc">首次部署，请填写数据库连接信息完成系统初始化</p>

      <el-steps :active="currentStep" finish-status="success" align-center style="margin-bottom: 32px">
        <el-step title="填写配置" />
        <el-step title="测试连接" />
        <el-step title="执行初始化" />
      </el-steps>

      <!-- 步骤 1：填写配置 -->
      <el-form
        v-if="currentStep === 0"
        ref="configFormRef"
        :model="configForm"
        :rules="formRules"
        label-width="120px"
      >
        <h3>数据库配置</h3>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="地址" prop="dbHost">
              <el-input v-model="configForm.dbHost" placeholder="127.0.0.1" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="端口" prop="dbPort">
              <el-input-number v-model="configForm.dbPort" :min="1" :max="65535" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="数据库名" prop="dbName">
              <el-input v-model="configForm.dbName" placeholder="admindb" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="用户名" prop="dbUser">
              <el-input v-model="configForm.dbUser" placeholder="root" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="密码" prop="dbPassword">
          <el-input v-model="configForm.dbPassword" type="password" show-password placeholder="请输入数据库密码" />
        </el-form-item>

        <h3 style="margin-top: 24px">Redis 配置（可选）</h3>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="地址">
              <el-input v-model="configForm.redisHost" placeholder="127.0.0.1" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="端口">
              <el-input-number v-model="configForm.redisPort" :min="1" :max="65535" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="密码">
              <el-input v-model="configForm.redisPassword" type="password" show-password placeholder="无密码可留空" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="数据库编号">
              <el-input-number v-model="configForm.redisDb" :min="0" :max="15" />
            </el-form-item>
          </el-col>
        </el-row>

        <h3 style="margin-top: 24px">服务配置（可选）</h3>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="应用名称">
              <el-input v-model="configForm.appName" placeholder="管理平台" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="服务端口">
              <el-input-number v-model="configForm.appPort" :min="1" :max="65535" />
            </el-form-item>
          </el-col>
        </el-row>

        <el-form-item>
          <el-button type="primary" @click="handleTestConnection">下一步：测试连接</el-button>
        </el-form-item>
      </el-form>

      <!-- 步骤 2：测试连接 -->
      <div v-if="currentStep === 1" class="test-result">
        <div v-if="testing" style="text-align: center; padding: 40px">
          <el-icon class="is-loading" :size="48"><Loading /></el-icon>
          <p>正在测试连接...</p>
        </div>
        <el-result
          v-else
          :icon="testPassed ? 'success' : 'error'"
          :title="testPassed ? '连接测试通过' : '连接测试未通过'"
        >
          <template #sub-title>
            <p>{{ testMessage }}</p>
          </template>
          <template #extra>
            <el-button @click="currentStep = 0">返回修改</el-button>
            <el-button v-if="testPassed" type="primary" @click="handleExecuteInit">下一步：执行初始化</el-button>
            <el-button v-else type="warning" @click="handleTestConnection">重新测试</el-button>
          </template>
        </el-result>
      </div>

      <!-- 步骤 3：执行初始化 -->
      <div v-if="currentStep === 2" class="init-result">
        <div v-if="initializing" style="text-align: center; padding: 40px">
          <el-icon class="is-loading" :size="48"><Loading /></el-icon>
          <p>正在执行初始化，请稍候...</p>
        </div>
        <el-result v-else-if="initSuccess" icon="success" title="初始化完成">
          <template #sub-title>
            <p>{{ initMessage }}</p>
          </template>
        </el-result>
        <el-result v-else icon="error" title="初始化失败">
          <template #sub-title>
            <p>{{ initMessage }}</p>
          </template>
          <template #extra>
            <el-button @click="currentStep = 0">返回修改配置</el-button>
          </template>
        </el-result>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { Loading } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'
import type { FormInstance, FormRules } from 'element-plus'
import { testDBConnection, executeInit } from '@/api/init'
import type { SetupConfig } from '@/api/init'

const configFormRef = ref<FormInstance>()
const currentStep = ref(0)
const testing = ref(false)
const initializing = ref(false)
const testPassed = ref(false)
const testMessage = ref('')
const initSuccess = ref(false)
const initMessage = ref('')

const configForm = reactive<SetupConfig>({
  dbHost: '127.0.0.1',
  dbPort: 3306,
  dbName: 'admindb',
  dbUser: 'root',
  dbPassword: '',
  redisHost: '127.0.0.1',
  redisPort: 6379,
  redisPassword: '',
  redisDb: 0,
  appName: '管理平台',
  appPort: 8000
})

const formRules: FormRules = {
  dbHost: [{ required: true, message: '请输入数据库地址', trigger: 'blur' }],
  dbPort: [{ required: true, message: '请输入端口', trigger: 'blur' }],
  dbName: [{ required: true, message: '请输入数据库名称', trigger: 'blur' }],
  dbUser: [{ required: true, message: '请输入用户名', trigger: 'blur' }]
}

async function handleTestConnection() {
  if (configFormRef.value) {
    const valid = await configFormRef.value.validate().catch(() => false)
    if (!valid) return
  }
  testing.value = true
  currentStep.value = 1
  testPassed.value = false
  testMessage.value = ''
  try {
    const res = await testDBConnection({
      dbHost: configForm.dbHost,
      dbPort: configForm.dbPort,
      dbName: configForm.dbName,
      dbUser: configForm.dbUser,
      dbPassword: configForm.dbPassword
    })
    testPassed.value = res.code === 200
    testMessage.value = res.msg
  } catch (e: any) {
    testPassed.value = false
    testMessage.value = e.message || '请求失败'
  } finally {
    testing.value = false
  }
}

async function handleExecuteInit() {
  initializing.value = true
  currentStep.value = 2
  initSuccess.value = false
  initMessage.value = ''
  try {
    const res = await executeInit(configForm)
    initSuccess.value = res.code === 200
    initMessage.value = res.msg
    if (initSuccess.value) {
      ElMessage.success('初始化完成，正在跳转登录页')
      setTimeout(() => { window.location.href = '/login' }, 1500)
    }
  } catch (e: any) {
    initSuccess.value = false
    initMessage.value = e.message || '初始化失败'
  } finally {
    initializing.value = false
  }
}
</script>

<style scoped>
.init-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: #f0f2f5;
}
.init-card {
  width: 680px;
  padding: 40px;
  background: #fff;
  border-radius: 8px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}
.init-title {
  text-align: center;
  margin-bottom: 8px;
  color: #303133;
}
.init-desc {
  text-align: center;
  color: #909399;
  margin-bottom: 32px;
}
h3 { color: #606266; margin-bottom: 16px; font-size: 15px; }
</style>
