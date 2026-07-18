<template>
  <el-dialog v-model="visible" :title="isEdit ? '编辑岗位' : '新增岗位'" width="500px" destroy-on-close>
    <el-form ref="formElRef" :model="form" :rules="rules" label-width="80px">
      <el-form-item label="岗位编码" prop="postCode">
        <el-input v-model="form.postCode" placeholder="请输入岗位编码" />
      </el-form-item>
      <el-form-item label="岗位名称" prop="postName">
        <el-input v-model="form.postName" placeholder="请输入岗位名称" />
      </el-form-item>
      <el-form-item label="排序" prop="sort">
        <el-input-number v-model="form.sort" :min="0" />
      </el-form-item>
      <el-form-item label="状态" prop="status">
        <el-radio-group v-model="form.status">
          <el-radio :value="0">启用</el-radio>
          <el-radio :value="1">禁用</el-radio>
        </el-radio-group>
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
import { createPost, updatePost } from '@/api/post'
import type { PostItem } from '@/api/post'

const emit = defineEmits<{ success: [] }>()

const visible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref(0)
const formElRef = ref<FormInstance>()

const form = reactive({
  postCode: '',
  postName: '',
  sort: 0,
  status: 0
})

const rules: FormRules = {
  postCode: [{ required: true, message: '请输入岗位编码', trigger: 'blur' }],
  postName: [{ required: true, message: '请输入岗位名称', trigger: 'blur' }]
}

async function open(row?: PostItem) {
  visible.value = true
  isEdit.value = !!row
  await nextTick()
  formElRef.value?.resetFields()
  if (row) {
    editId.value = row.postId
    form.postCode = row.postCode
    form.postName = row.postName
    form.sort = row.sort
    form.status = row.status
  } else {
    editId.value = 0
    form.postCode = ''
    form.postName = ''
    form.sort = 0
    form.status = 0
  }
}

async function handleSubmit() {
  await formElRef.value?.validate()
  submitting.value = true
  try {
    if (isEdit.value) {
      await updatePost(editId.value, { ...form })
    } else {
      await createPost({ ...form })
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
