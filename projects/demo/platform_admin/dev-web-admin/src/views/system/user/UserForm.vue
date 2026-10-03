<template>
  <el-dialog v-model="visible" :title="isEdit ? '编辑用户' : '新增用户'" width="500px" destroy-on-close>
    <el-form ref="formElRef" :model="form" :rules="rules" label-width="80px">
      <el-form-item label="用户名" prop="username">
        <el-input v-model="form.username" :disabled="isEdit" placeholder="请输入用户名" />
      </el-form-item>
      <el-form-item v-if="!isEdit" label="密码" prop="password">
        <el-input v-model="form.password" type="password" show-password placeholder="请输入密码" />
      </el-form-item>
      <el-form-item label="邮箱" prop="email">
        <el-input v-model="form.email" placeholder="请输入邮箱" />
      </el-form-item>
      <el-form-item label="手机号" prop="phone">
        <el-input v-model="form.phone" placeholder="请输入手机号" />
      </el-form-item>
      <el-form-item label="Access Token过期时间(分钟)" prop="access_token_expire_minutes">
        <el-input-number
          v-model="form.access_token_expire_minutes"
          :min="1"
          :max="10080"
          :step="1"
          placeholder="留空使用全局默认值(60分钟)"
          style="width: 100%;"
        />
      </el-form-item>
      <el-form-item label="Refresh Token过期时间(分钟)" prop="refresh_token_expire_minutes">
        <el-input-number
          v-model="form.refresh_token_expire_minutes"
          :min="1"
          :max="10080"
          :step="1"
          placeholder="留空使用全局默认值(10080分钟)"
          style="width: 100%;"
        />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, nextTick } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { createUser, updateUser } from '@/api/user'
import type { UserPageItem } from '@/api/user'

const emit = defineEmits<{ success: [] }>()

const visible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const formElRef = ref<FormInstance>()

const form = reactive({
  username: '',
  password: '',
  email: '',
  phone: '',
  access_token_expire_minutes: null as number | null,
  refresh_token_expire_minutes: null as number | null,
  version: 1
})

const rules: FormRules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
  email: [
    { type: 'email', message: '邮箱格式不正确', trigger: 'blur' }
  ],
  phone: [
    { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确', trigger: 'blur' }
  ]
}

async function open(row?: UserPageItem) {
  visible.value = true
  isEdit.value = !!row
  await nextTick()
  formElRef.value?.resetFields()
  if (row) {
    editId.value = row.id
    form.username = row.username
    form.password = ''
    form.email = row.email || ''
    form.phone = row.phone || ''
    form.access_token_expire_minutes = row.access_token_expire_minutes ?? null
    form.refresh_token_expire_minutes = row.refresh_token_expire_minutes ?? null
    form.version = row.version || 1
  } else {
    editId.value = ''
    form.username = ''
    form.password = ''
    form.email = ''
    form.phone = ''
    form.access_token_expire_minutes = null
    form.refresh_token_expire_minutes = null
    form.version = 1
  }
}

async function handleSubmit() {
  await formElRef.value?.validate()
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateUser(editId.value, {
        email: form.email || undefined,
        phone: form.phone || undefined,
        access_token_expire_minutes: form.access_token_expire_minutes ?? undefined,
        refresh_token_expire_minutes: form.refresh_token_expire_minutes ?? undefined,
        version: form.version
      })
    } else {
      await createUser({
        username: form.username,
        password: form.password,
        email: form.email || undefined,
        phone: form.phone || undefined,
        access_token_expire_minutes: form.access_token_expire_minutes ?? undefined,
        refresh_token_expire_minutes: form.refresh_token_expire_minutes ?? undefined
      })
    }
    ElMessage.success(isEdit.value ? '编辑成功' : '新增成功')
    visible.value = false
    emit('success')
  } finally {
    submitting.value = false
  }
}

defineExpose({ open })
</script>
