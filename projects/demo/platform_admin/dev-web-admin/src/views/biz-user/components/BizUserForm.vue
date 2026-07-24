<template>
  <el-dialog v-model="visible" :title="isEdit ? '编辑C端用户' : '新增C端用户'" width="500px" destroy-on-close>
    <el-form ref="formElRef" :model="form" :rules="rules" label-width="80px">
      <el-form-item label="手机号" prop="phone">
        <el-input v-model="form.phone" placeholder="请输入手机号" />
      </el-form-item>
      <el-form-item label="昵称" prop="nickname">
        <el-input v-model="form.nickname" placeholder="请输入昵称" />
      </el-form-item>
      <el-form-item v-if="!isEdit" label="密码" prop="password">
        <el-input v-model="form.password" type="password" show-password placeholder="留空则由系统生成" />
      </el-form-item>
      <el-form-item v-if="isEdit" label="头像" prop="avatar">
        <el-input v-model="form.avatar" placeholder="请输入头像URL" />
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
import { createBizUser, updateBizUser } from '@/api/biz-user'
import type { BizUser } from '@/api/biz-user'

const emit = defineEmits<{ success: [] }>()

const visible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const formElRef = ref<FormInstance>()

const form = reactive({
  phone: '',
  nickname: '',
  password: '',
  avatar: '',
  version: 1
})

const rules: FormRules = {
  phone: [
    { required: true, message: '请输入手机号', trigger: 'blur' },
    { pattern: /^1[3-9]\d{9}$/, message: '手机号格式不正确', trigger: 'blur' }
  ]
}

/** 打开弹窗，传入 row 为编辑模式 */
async function open(row?: BizUser) {
  visible.value = true
  isEdit.value = !!row
  await nextTick()
  formElRef.value?.resetFields()
  if (row) {
    editId.value = row.id
    form.phone = row.phone
    form.nickname = row.nickname || ''
    form.password = ''
    form.avatar = row.avatar || ''
    form.version = row.version
  } else {
    editId.value = ''
    form.phone = ''
    form.nickname = ''
    form.password = ''
    form.avatar = ''
    form.version = 1
  }
}

async function handleSubmit() {
  await formElRef.value?.validate()
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateBizUser(editId.value, {
        phone: form.phone || undefined,
        nickname: form.nickname || undefined,
        avatar: form.avatar || undefined,
        version: form.version
      })
    } else {
      await createBizUser({
        phone: form.phone,
        nickname: form.nickname || undefined,
        password: form.password || undefined
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
