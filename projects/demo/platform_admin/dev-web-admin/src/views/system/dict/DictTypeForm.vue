<template>
  <el-dialog v-model="visible" :title="isEdit ? '编辑字典类型' : '新增字典类型'" width="500px" destroy-on-close>
    <el-form ref="formElRef" :model="form" :rules="rules" label-width="90px">
      <el-form-item label="字典名称" prop="dictName">
        <el-input v-model="form.dictName" placeholder="请输入字典名称" />
      </el-form-item>
      <el-form-item label="字典类型" prop="dictType">
        <el-input v-model="form.dictType" placeholder="请输入字典类型" :disabled="isEdit" />
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-radio-group v-model="form.status">
          <el-radio :value="0">正常</el-radio>
          <el-radio :value="1">停用</el-radio>
        </el-radio-group>
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
import { createDictType, updateDictType } from '@/api/dict'
import type { DictTypeItem } from '@/api/dict'

const emit = defineEmits<{ success: [] }>()

const visible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref(0)
const formElRef = ref<FormInstance>()

const form = reactive({
  dictName: '',
  dictType: '',
  status: 0,
  remark: ''
})

const rules: FormRules = {
  dictName: [{ required: true, message: '请输入字典名称', trigger: 'blur' }],
  dictType: [{ required: true, message: '请输入字典类型', trigger: 'blur' }]
}

async function open(row?: DictTypeItem) {
  visible.value = true
  isEdit.value = !!row
  await nextTick()
  formElRef.value?.resetFields()
  if (row) {
    editId.value = row.dictId
    form.dictName = row.dictName
    form.dictType = row.dictType
    form.status = row.status
    form.remark = row.remark || ''
  } else {
    editId.value = 0
    form.dictName = ''
    form.dictType = ''
    form.status = 0
    form.remark = ''
  }
}

async function handleSubmit() {
  await formElRef.value?.validate()
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateDictType(editId.value, { ...form })
    } else {
      await createDictType({ ...form })
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
