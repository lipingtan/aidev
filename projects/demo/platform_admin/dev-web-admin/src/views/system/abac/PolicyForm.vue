<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? '编辑 ABAC 策略' : '新增 ABAC 策略'"
    width="760px"
    destroy-on-close
  >
    <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
      <!-- 基本信息 -->
      <el-form-item label="策略名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入策略名称" />
      </el-form-item>

      <el-form-item label="资源对象" prop="resource_type">
        <el-select v-model="form.resource_type" placeholder="选择资源对象" style="width: 100%">
          <el-option
            v-for="res in resourceList"
            :key="res.type"
            :label="res.display_name"
            :value="res.type"
          />
        </el-select>
      </el-form-item>

      <el-row :gutter="16">
        <el-col :span="12">
          <el-form-item label="主体类型" prop="subject_type">
            <el-select v-model="form.subject_type" placeholder="选择主体类型" style="width: 100%" @change="onSubjectTypeChange">
              <el-option label="角色" value="ROLE" />
              <el-option label="权限集" value="PERMISSION_SET" />
              <el-option label="用户" value="USER" />
              <el-option label="部门" value="DEPT" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="主体标识" prop="subject_id">
            <!-- ROLE / PERMISSION_SET：从角色列表下拉选择 -->
            <el-select
              v-if="form.subject_type === 'ROLE' || form.subject_type === 'PERMISSION_SET'"
              v-model="form.subject_id"
              placeholder="选择角色"
              filterable
              style="width: 100%"
            >
              <el-option
                v-for="r in filteredRoleOptions"
                :key="r.role_code"
                :label="`${r.role_name}（${r.role_code}）`"
                :value="r.role_code"
              />
            </el-select>
            <!-- USER / DEPT：手动输入 ID -->
            <el-input
              v-else
              v-model="form.subject_id"
              :placeholder="form.subject_type === 'USER' ? '用户雪花 ID' : '部门 ID'"
            />
          </el-form-item>
        </el-col>
      </el-row>

      <el-row :gutter="16">
        <el-col :span="8">
          <el-form-item label="效果" prop="effect">
            <el-select v-model="form.effect" style="width: 100%">
              <el-option label="允许（ALLOW）" value="ALLOW" />
              <el-option label="拒绝（DENY）" value="DENY" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="优先级">
            <el-input-number v-model="form.priority" :min="1" :max="9999" style="width: 100%" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item v-if="isEdit" label="状态">
            <el-switch v-model="form.status" :active-value="1" :inactive-value="0" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item label="描述">
        <el-input v-model="form.description" type="textarea" :rows="2" placeholder="可选" />
      </el-form-item>

      <!-- 行权限 Tab -->
      <el-divider content-position="left">行权限配置</el-divider>
      <div v-for="(rp, i) in form.row_policies" :key="i" class="row-policy-item">
        <div class="row-policy-header">
          <el-select v-model="rp.action" size="small" style="width: 120px">
            <el-option label="读取（read）" value="read" />
            <el-option label="创建（create）" value="create" />
            <el-option label="更新（update）" value="update" />
            <el-option label="删除（delete）" value="delete" />
          </el-select>
          <span class="rp-label">行过滤条件：</span>
          <el-button size="small" link type="danger" @click="removeRowPolicy(i)">移除</el-button>
        </div>
        <ConditionEditor
          v-model="rp.condition_expr"
          :resource-attrs="currentResourceAttrs"
          :subject-attrs="subjectAttrs"
        />
      </div>
      <el-button type="primary" plain size="small" @click="addRowPolicy">+ 添加行权限</el-button>

      <!-- 列权限 -->
      <el-divider content-position="left">列权限配置</el-divider>
      <div v-for="(cp, i) in form.col_policies" :key="i" class="col-policy-item">
        <el-input v-model="cp.field_name" size="small" placeholder="字段名（JSON key）" style="width: 160px" />
        <el-select v-model="cp.effect" size="small" style="width: 100px" @change="cp.mask_type = ''">
          <el-option label="显示" value="SHOW" />
          <el-option label="隐藏" value="HIDE" />
          <el-option label="脱敏" value="MASK" />
        </el-select>
        <el-select v-if="cp.effect === 'MASK'" v-model="cp.mask_type" size="small" style="width: 120px" placeholder="脱敏类型">
          <el-option label="手机号" value="phone" />
          <el-option label="邮箱" value="email" />
          <el-option label="身份证" value="id_card" />
          <el-option label="自定义" value="custom" />
        </el-select>
        <el-input
          v-if="cp.mask_type === 'custom'"
          v-model="cp.mask_pattern"
          size="small"
          placeholder="正则表达式"
          style="width: 160px"
        />
        <el-button size="small" link type="danger" @click="removeColPolicy(i)">×</el-button>
      </div>
      <el-button type="primary" plain size="small" @click="addColPolicy">+ 添加列权限</el-button>
    </el-form>

    <!-- 乐观锁：隐藏传参 -->
    <input type="hidden" :value="form.version" />

    <template #footer>
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import type { FormInstance } from 'element-plus'
import {
  createAbacPolicy,
  updateAbacPolicy,
  getAbacPolicy,
  type AbacPolicyItem,
  type ResourceVO,
  type SubjectAttrVO,
  type CondNode,
  type RowPolicyInput,
  type ColPolicyInput
} from '@/api/abac'
import { getRoleList, type RoleItem } from '@/api/role'
import ConditionEditor from './ConditionEditor.vue'

const props = defineProps<{
  resourceList: ResourceVO[]
  subjectAttrs: SubjectAttrVO[]
}>()

const emit = defineEmits<{ success: [] }>()

const visible = ref(false)
const isEdit = ref(false)
const submitting = ref(false)
const formRef = ref<FormInstance>()
const editingId = ref('')

// 角色列表（供 ROLE / PERMISSION_SET 下拉使用）
const allRoles = ref<RoleItem[]>([])

// 根据 subject_type 过滤角色：ROLE 显示普通角色，PERMISSION_SET 显示权限集
const filteredRoleOptions = computed(() => {
  if (form.subject_type === 'PERMISSION_SET') {
    return flattenRoles(allRoles.value).filter(r => r.role_type === 'PERMISSION_SET')
  }
  return flattenRoles(allRoles.value).filter(r => r.role_type !== 'PERMISSION_SET')
})

function flattenRoles(nodes: RoleItem[]): RoleItem[] {
  const result: RoleItem[] = []
  function walk(list: RoleItem[]) {
    for (const n of list) {
      result.push(n)
      if (n.children) walk(n.children)
    }
  }
  walk(nodes)
  return result
}

async function loadRoles() {
  try {
    allRoles.value = await getRoleList()
  } catch { /* 静默 */ }
}

function onSubjectTypeChange() {
  form.subject_id = ''
}

onMounted(() => {
  loadRoles()
})
interface FormState {
  name: string
  resource_type: string
  subject_type: string
  subject_id: string
  effect: string
  priority: number
  status: number
  version: number
  description: string
  row_policies: Array<{ action: string; condition_expr: CondNode | null }>
  col_policies: Array<ColPolicyInput & { mask_type?: string; mask_pattern?: string }>
}

const form = reactive<FormState>({
  name: '', resource_type: '', subject_type: 'ROLE', subject_id: '',
  effect: 'ALLOW', priority: 100, status: 1, version: 1, description: '',
  row_policies: [], col_policies: []
})

const rules = {
  name: [{ required: true, message: '请输入策略名称', trigger: 'blur' }],
  resource_type: [{ required: true, message: '请选择资源对象', trigger: 'change' }],
  subject_type: [{ required: true, message: '请选择主体类型', trigger: 'change' }],
  subject_id: [{ required: true, message: '请输入主体标识', trigger: 'blur' }],
  effect: [{ required: true, message: '请选择效果', trigger: 'change' }],
}

// 当前资源的属性列表
const currentResourceAttrs = computed(() => {
  const res = props.resourceList.find(r => r.type === form.resource_type)
  return res?.attributes || []
})

function addRowPolicy() {
  form.row_policies.push({ action: 'read', condition_expr: null })
}

function removeRowPolicy(i: number) {
  form.row_policies.splice(i, 1)
}

function addColPolicy() {
  form.col_policies.push({ field_name: '', effect: 'SHOW' })
}

function removeColPolicy(i: number) {
  form.col_policies.splice(i, 1)
}

function resetForm() {
  form.name = ''
  form.resource_type = ''
  form.subject_type = 'ROLE'
  form.subject_id = ''
  form.effect = 'ALLOW'
  form.priority = 100
  form.status = 1
  form.version = 1
  form.description = ''
  form.row_policies = []
  form.col_policies = []
}

async function open(row?: AbacPolicyItem) {
  resetForm()
  if (row) {
    isEdit.value = true
    editingId.value = row.id
    // 加载完整详情（含 row_policies + col_policies）
    const detail: any = await getAbacPolicy(row.id)
    form.name = detail.name
    form.resource_type = detail.resource_type
    form.subject_type = detail.subject_type
    form.subject_id = detail.subject_id
    form.effect = detail.effect
    form.priority = detail.priority
    form.status = detail.status
    form.version = detail.version  // 乐观锁
    form.description = detail.description || ''
    form.row_policies = (detail.row_policies || []).map((rp: any) => ({
      action: rp.action,
      condition_expr: rp.condition_expr || null
    }))
    form.col_policies = (detail.col_policies || []).map((cp: any) => ({
      field_name: cp.field_name,
      effect: cp.effect,
      mask_type: cp.mask_type || '',
      mask_pattern: cp.mask_pattern || ''
    }))
  } else {
    isEdit.value = false
    editingId.value = ''
  }
  visible.value = true
}

async function handleSubmit() {
  const valid = await formRef.value?.validate().catch(() => false)
  if (!valid) return

  submitting.value = true
  try {
    const rowPolicies: RowPolicyInput[] = form.row_policies.map(rp => ({
      action: rp.action,
      condition_expr: rp.condition_expr || undefined
    }))
    const colPolicies: ColPolicyInput[] = form.col_policies.map(cp => ({
      field_name: cp.field_name,
      effect: cp.effect,
      mask_type: cp.mask_type || undefined,
      mask_pattern: cp.mask_pattern || undefined
    }))

    if (isEdit.value) {
      await updateAbacPolicy(editingId.value, {
        name: form.name,
        effect: form.effect,
        priority: form.priority,
        status: form.status,
        description: form.description,
        version: form.version,  // 乐观锁
        row_policies: rowPolicies,
        col_policies: colPolicies
      })
      ElMessage.success('更新成功')
    } else {
      await createAbacPolicy({
        name: form.name,
        resource_type: form.resource_type,
        subject_type: form.subject_type,
        subject_id: form.subject_id,
        effect: form.effect,
        priority: form.priority,
        description: form.description,
        row_policies: rowPolicies,
        col_policies: colPolicies
      })
      ElMessage.success('创建成功')
    }
    visible.value = false
    emit('success')
  } finally {
    submitting.value = false
  }
}

defineExpose({ open })
</script>

<style scoped>
.row-policy-item {
  border: 1px solid #ebeef5;
  border-radius: 6px;
  padding: 10px 12px;
  margin-bottom: 10px;
  background: #fafafa;
}
.row-policy-header {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}
.rp-label { font-size: 13px; color: #606266; }
.col-policy-item {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}
</style>
