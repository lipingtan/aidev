<template>
  <el-dialog v-model="visible" :title="isEdit ? '编辑配置' : '新增配置'" width="500px" destroy-on-close>
    <el-form ref="formElRef" :model="form" :rules="rules" label-width="80px">
      <el-form-item label="配置名称" prop="configName">
        <el-input v-model="form.configName" placeholder="请输入配置名称" />
      </el-form-item>
      <el-form-item label="配置键" prop="configKey">
        <el-input v-model="form.configKey" placeholder="请输入配置键" />
      </el-form-item>
      <el-form-item label="配置值" prop="configValue">
        <el-input v-model="form.configValue" type="textarea" placeholder="请输入配置值" />
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
const editId = ref(0)
const formElRef = ref<FormInstance>()

const form = reactive({
  configName: '',
  configKey: '',
  configValue: '',
  remark: ''
})

const rules: FormRules = {
  configName: [{ required: true, message: '请输入配置名称', trigger: 'blur' }],
  configKey: [{ required: true, message: '请输入配置键', trigger: 'blur' }],
  configValue: [{ required: true, message: '请输入配置值', trigger: 'blur' }]
}

async function open(row?: ConfigItem) {
  visible.value = true
  isEdit.value = !!row
  await nextTick()
  formElRef.value?.resetFields()
  if (row) {
    editId.value = row.configId
    form.configName = row.configName
    form.configKey = row.configKey
    form.configValue = row.configValue
    form.remark = row.remark || ''
  } else {
    editId.value = 0
    form.configName = ''
    form.configKey = ''
    form.configValue = ''
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
