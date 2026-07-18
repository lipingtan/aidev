<template>
  <el-dialog v-model="visible" :title="isEdit ? '编辑部门' : '新增部门'" width="500px" destroy-on-close>
    <el-form ref="formElRef" :model="form" :rules="rules" label-width="80px">
      <el-form-item label="上级部门" prop="parentId">
        <el-tree-select
          v-model="form.parentId"
          :data="deptTreeData"
          :props="{ label: 'deptName', value: 'deptId', children: 'children' }"
          check-strictly
          placeholder="请选择上级部门"
          style="width: 100%"
        />
      </el-form-item>
      <el-form-item label="部门名称" prop="deptName">
        <el-input v-model="form.deptName" placeholder="请输入部门名称" />
      </el-form-item>
      <el-form-item label="负责人" prop="leader">
        <el-input v-model="form.leader" placeholder="请输入负责人" />
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
import { createDept, updateDept, getDeptTree } from '@/api/dept'
import type { DeptItem } from '@/api/dept'

const emit = defineEmits<{ success: [] }>()

const visible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref(0)
const formElRef = ref<FormInstance>()
const deptTreeData = ref<DeptItem[]>([])

const form = reactive({
  deptName: '',
  parentId: 0,
  sort: 0,
  leader: '',
  status: 0
})

const rules: FormRules = {
  deptName: [{ required: true, message: '请输入部门名称', trigger: 'blur' }],
  parentId: [{ required: true, message: '请选择上级部门', trigger: 'change' }]
}

async function open(row?: DeptItem) {
  visible.value = true
  isEdit.value = !!row
  // 加载部门树
  const res: any = await getDeptTree()
  deptTreeData.value = Array.isArray(res) ? res : (res?.list || [])
  await nextTick()
  formElRef.value?.resetFields()
  if (row) {
    editId.value = row.deptId
    form.deptName = row.deptName
    form.parentId = row.parentId
    form.sort = row.sort
    form.leader = row.leader || ''
    form.status = row.status
  } else {
    editId.value = 0
    form.deptName = ''
    form.parentId = 0
    form.sort = 0
    form.leader = ''
    form.status = 0
  }
}

async function handleSubmit() {
  await formElRef.value?.validate()
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateDept(editId.value, { ...form })
    } else {
      await createDept({ ...form })
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
