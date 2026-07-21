<template>
  <el-dialog v-model="visible" :title="dialogTitle" width="520px" destroy-on-close>
    <el-form ref="formElRef" :model="form" :rules="rules" label-width="100px">
      <el-form-item label="角色名称" prop="role_name">
        <el-input v-model="form.role_name" placeholder="请输入角色名称" />
      </el-form-item>
      <el-form-item label="角色编码" prop="role_code">
        <el-input v-model="form.role_code" :disabled="isEdit" placeholder="请输入角色编码" />
      </el-form-item>
      <el-form-item v-if="!props.isPermissionSet" label="上级角色" prop="parent_id">
        <el-tree-select
          v-model="form.parent_id"
          :data="parentTreeData"
          :props="{ label: 'role_name', children: 'children', value: 'id' }"
          placeholder="请选择上级角色（留空为顶级）"
          clearable
          check-strictly
          :render-after-expand="false"
          style="width: 100%"
        />
      </el-form-item>
      <el-form-item label="排序" prop="sort_order">
        <el-input-number v-model="form.sort_order" :min="0" />
      </el-form-item>
      <el-form-item label="状态" prop="status">
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
import { ref, reactive, nextTick, computed } from 'vue'
import type { FormInstance, FormRules } from 'element-plus'
import { ElMessage } from 'element-plus'
import { createRole, updateRole } from '@/api/role'
import type { RoleItem } from '@/api/role'

const props = defineProps<{
  roleTree: RoleItem[]
  isPermissionSet?: boolean
}>()

const emit = defineEmits<{ success: [] }>()

const visible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const editId = ref('')
const editVersion = ref(1)
const formElRef = ref<FormInstance>()

const form = reactive({
  role_name: '',
  role_code: '',
  parent_id: '' as string | undefined,
  sort_order: 0,
  status: 1
})

const rules: FormRules = {
  role_name: [{ required: true, message: '请输入角色名称', trigger: 'blur' }],
  role_code: [{ required: true, message: '请输入角色编码', trigger: 'blur' }]
}

const dialogTitle = computed(() => {
  if (props.isPermissionSet) {
    return isEdit.value ? '编辑权限集' : '新增权限集'
  }
  return isEdit.value ? '编辑角色' : '新增角色'
})

/** 构建父角色下拉树数据，编辑时排除自身及子节点 */
const parentTreeData = computed(() => {
  if (!isEdit.value) return props.roleTree
  return filterSelfAndChildren(props.roleTree, editId.value)
})

function filterSelfAndChildren(nodes: RoleItem[], excludeId: string): RoleItem[] {
  return nodes
    .filter((n) => n.id !== excludeId)
    .map((n) => ({
      ...n,
      children: n.children ? filterSelfAndChildren(n.children, excludeId) : undefined
    }))
}

async function open(row?: RoleItem) {
  visible.value = true
  isEdit.value = !!row
  await nextTick()
  formElRef.value?.resetFields()
  if (row) {
    editId.value = row.id
    editVersion.value = row.version
    form.role_name = row.role_name
    form.role_code = row.role_code
    form.parent_id = row.parent_id || undefined
    form.sort_order = row.sort_order
    form.status = row.status
  } else {
    editId.value = ''
    editVersion.value = 1
    form.role_name = ''
    form.role_code = ''
    form.parent_id = undefined
    form.sort_order = 0
    form.status = 1
  }
}

async function handleSubmit() {
  await formElRef.value?.validate()
  submitting.value = true
  try {
    if (isEdit.value) {
      await updateRole(editId.value, {
        role_name: form.role_name,
        role_type: props.isPermissionSet ? 'PERMISSION_SET' : undefined,
        parent_id: props.isPermissionSet ? undefined : (form.parent_id || undefined),
        sort_order: form.sort_order,
        status: form.status,
        version: editVersion.value
      })
    } else {
      await createRole({
        role_name: form.role_name,
        role_code: form.role_code,
        role_type: props.isPermissionSet ? 'PERMISSION_SET' : undefined,
        parent_id: props.isPermissionSet ? undefined : (form.parent_id || undefined),
        sort_order: form.sort_order,
        status: form.status
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
