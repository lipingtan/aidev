<template>
  <el-dialog v-model="visible" :title="isEdit ? '编辑字典数据' : '新增字典数据'" width="500px" destroy-on-close>
    <el-form ref="formElRef" :model="form" :rules="rules" label-width="90px">
      <el-form-item label="数据标签" prop="dictLabel">
        <el-input v-model="form.dictLabel" placeholder="请输入数据标签" />
      </el-form-item>
      <el-form-item label="数据键值" prop="dictValue">
        <el-input v-model="form.dictValue" placeholder="请输入数据键值" />
      </el-form-item>
      <el-form-item label="排序" prop="dictSort">
        <el-input-number v-model="form.dictSort" :min="0" />
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
import { createDictData, updateDictData } from '@/api/dict'
import type { DictDataItem } from '@/api/dict'

const emit = defineEmits<{ success: [] }>()

const visible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref(0)
const formElRef = ref<FormInstance>()
const currentDictType = ref('')

const form = reactive({
  dictLabel: '',
  dictValue: '',
  dictSort: 0,
  dictType: '',
  status: 0,
  remark: ''
})

const rules: FormRules = {
  dictLabel: [{ required: true, message: '请输入数据标签', trigger: 'blur' }],
  dictValue: [{ required: true, message: '请输入数据键值', trigger: 'blur' }]
}

async function open(row?: DictDataItem, dictType?: string) {
  visible.value = true
  isEdit.value = !!row
  if (dictType) currentDictType.value = dictType
  await nextTick()
  formElRef.value?.resetFields()
  if (row) {
    editId.value = row.dictCode
    form.dictLabel = row.dictLabel
    form.dictValue = row.dictValue
    form.dictSort = row.dictSort
    form.dictType = row.dictType
    form.status = row.status
    form.remark = row.remark || ''
  } else {
    editId.value = 0
    form.dictLabel = ''
    form.dictValue = ''
    form.dictSort = 0
    form.dictType = currentDictType.value
    form.status = 0
    form.remark = ''
  }
}

async function handleSubmit() {
  await formElRef.value?.validate()
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateDictData(editId.value, { ...form })
    } else {
      await createDictData({ ...form })
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
