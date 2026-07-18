<template>
  <el-dialog v-model="visible" :title="isEdit ? '编辑配置' : '新增配置'" width="500px" destroy-on-close>
    <el-form ref="formElRef" :model="form" :rules="rules" label-width="80px">
      <el-form-item label="配置名称" prop="config_name">
        <el-input v-model="form.config_name" placeholder="请输入配置名称" />
      </el-form-item>
      <el-form-item label="配置键" prop="config_key">
        <el-input v-model="form.config_key" placeholder="请输入配置键" />
      </el-form-item>
      <el-form-item label="配置值" prop="config_value">
        <el-input v-model="form.config_value" type="textarea" placeholder="请输入配置值" />
      </el-form-item>
      <el-form-item label="备注" prop="remark">
        <el-input v-model="form.remark" type="textarea" placeholder="请输入备注" />
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
import { createConfig, updateConfig } from '@/api/config'
import type { ConfigItem } from '@/api/config'

const emit = defineEmits<{ success: [] }>()

const visible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const formElRef = ref<FormInstance>()

const form = reactive({
  config_name: '',
  config_key: '',
  config_value: '',
  remark: ''
})

const rules: FormRules = {
  config_name: [{ required: true, message: '请输入配置名称', trigger: 'blur' }],
  config_key: [{ required: true, message: '请输入配置键', trigger: 'blur' }],
  config_value: [{ required: true, message: '请输入配置值', trigger: 'blur' }]
}

async function open(row?: ConfigItem) {
  visible.value = true
  isEdit.value = !!row
  await nextTick()
  formElRef.value?.resetFields()
  if (row) {
    editId.value = row.id
    form.config_name = row.config_name
    form.config_key = row.config_key
    form.config_value = row.config_value
    form.remark = row.remark || ''
  } else {
    editId.value = ''
    form.config_name = ''
    form.config_key = ''
    form.config_value = ''
    form.remark = ''
  }
}

async function handleSubmit() {
  await formElRef.value?.validate()
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateConfig(editId.value, { ...form })
    } else {
      await createConfig({ ...form })
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
