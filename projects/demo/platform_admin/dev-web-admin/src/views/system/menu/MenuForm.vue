<template>
  <el-dialog v-model="visible" :title="isEdit ? '编辑菜单' : '新增菜单'" width="550px" destroy-on-close>
    <el-form ref="formElRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="类型" prop="type" v-if="!isEdit">
        <el-radio-group v-model="form.type">
          <el-radio value="MENU">菜单</el-radio>
          <el-radio value="BUTTON">按钮</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入名称" />
      </el-form-item>
      <el-form-item label="路由路径" v-if="form.type === 'MENU'">
        <el-input v-model="form.path" placeholder="请输入路由路径" />
      </el-form-item>
      <el-form-item label="组件路径" v-if="form.type === 'MENU'">
        <el-input v-model="form.component" placeholder="请输入组件路径" />
      </el-form-item>
      <el-form-item label="权限标识">
        <el-input v-model="form.permission_code" placeholder="如 system:user:list" />
      </el-form-item>
      <el-form-item label="图标" v-if="form.type === 'MENU'">
        <el-input v-model="form.icon" placeholder="请输入图标名称" />
      </el-form-item>
      <el-form-item label="排序" prop="sort_order">
        <el-input-number v-model="form.sort_order" :min="0" />
      </el-form-item>
      <el-form-item label="状态" v-if="isEdit">
        <el-radio-group v-model="form.status">
          <el-radio :value="1">启用</el-radio>
          <el-radio :value="0">禁用</el-radio>
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
import { createMenu, updateMenu } from '@/api/menu'
import type { MenuTreeItem } from '@/api/menu'

const emit = defineEmits<{ success: [] }>()
const visible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const formElRef = ref<FormInstance>()

const form = reactive({
  name: '',
  parent_id: null as string | null,
  type: 'MENU',
  path: '',
  component: '',
  permission_code: '',
  icon: '',
  sort_order: 0,
  status: 1,
  app_code: ''
})

const rules: FormRules = {
  name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }]
}

async function open(row?: MenuTreeItem, parentId?: string | null) {
  visible.value = true
  isEdit.value = !!row
  await nextTick()
  formElRef.value?.resetFields()
  if (row) {
    editId.value = row.id
    form.name = row.name
    form.type = row.type
    form.path = row.path || ''
    form.component = row.component || ''
    form.permission_code = row.permission_code || ''
    form.icon = row.icon || ''
    form.sort_order = row.sort_order
    form.status = row.status
    form.parent_id = row.parent_id
    form.app_code = row.app_code || ''
  } else {
    editId.value = ''
    form.name = ''
    form.type = 'MENU'
    form.path = ''
    form.component = ''
    form.permission_code = ''
    form.icon = ''
    form.sort_order = 0
    form.status = 1
    form.parent_id = parentId || null
    form.app_code = ''
  }
}

async function handleSubmit() {
  await formElRef.value?.validate()
  submitting.value = true
  try {
    const payload = {
      name: form.name,
      type: form.type,
      parent_id: form.parent_id,
      path: form.path || undefined,
      component: form.component || undefined,
      permission_code: form.permission_code || undefined,
      icon: form.icon || undefined,
      sort_order: form.sort_order,
      app_code: form.app_code || undefined,
      status: form.status
    }
    if (isEdit.value) {
      await updateMenu(editId.value, payload)
    } else {
      await createMenu(payload)
    }
    ElMessage.success(isEdit.value ? '编辑成功' : '新增成功')
    visible.value = false
    emit('success')
  } finally { submitting.value = false }
}

defineExpose({ open })
</script>
