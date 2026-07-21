<template>
  <el-dialog v-model="dialogVisible" title="记录共享" width="640px" @close="handleClose">
    <!-- 已有共享规则列表 -->
    <div class="section-title">已有共享规则</div>
    <el-table :data="shareList" border size="small" style="margin-bottom: 20px;">
      <el-table-column label="共享目标类型" width="120">
        <template #default="{ row }">
          {{ shareToTypeLabel(row.share_to_type) }}
        </template>
      </el-table-column>
      <el-table-column prop="share_to_id" label="目标ID" width="120" />
      <el-table-column label="权限级别" width="100">
        <template #default="{ row }">
          <el-tag :type="row.access_level === 'EDIT' ? 'warning' : 'info'" size="small">
            {{ row.access_level === 'EDIT' ? '可编辑' : '只读' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="过期时间" width="160">
        <template #default="{ row }">
          {{ row.expire_at || '永久' }}
        </template>
      </el-table-column>
      <el-table-column label="操作" width="80" fixed="right">
        <template #default="{ row }">
          <el-button type="danger" link size="small" @click="handleDelete(row.id)">删除</el-button>
        </template>
      </el-table-column>
      <template #empty>
        <span>暂无共享规则</span>
      </template>
    </el-table>

    <!-- 新增共享表单 -->
    <div class="section-title">新增共享</div>
    <el-form :model="form" label-width="100px" size="default">
      <el-form-item label="目标类型">
        <el-radio-group v-model="form.share_to_type" @change="handleTargetTypeChange">
          <el-radio value="USER">用户</el-radio>
          <el-radio value="ROLE">角色</el-radio>
          <el-radio value="DEPT">部门</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="共享目标">
        <el-select v-model="form.share_to_id" filterable placeholder="请选择目标" style="width: 100%;">
          <el-option
            v-for="opt in targetOptions"
            :key="opt.value"
            :label="opt.label"
            :value="opt.value"
          />
        </el-select>
      </el-form-item>
      <el-form-item label="权限级别">
        <el-radio-group v-model="form.access_level">
          <el-radio value="READ">只读</el-radio>
          <el-radio value="EDIT">可编辑</el-radio>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="过期时间">
        <el-date-picker
          v-model="form.expire_at"
          type="datetime"
          placeholder="留空表示永久有效"
          style="width: 100%;"
          value-format="YYYY-MM-DD HH:mm:ss"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="dialogVisible = false">关闭</el-button>
      <el-button type="primary" :disabled="!canSubmit" @click="handleSubmit">添加共享</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, computed, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  createRecordShare,
  listRecordShares,
  deleteRecordShare,
  type RecordShareItem
} from '@/api/record-share'
import { listUsers, type UserPageItem } from '@/api/user'
import { getRoleList, type RoleItem } from '@/api/role'

// ==================== Props / Emits ====================
const props = defineProps<{
  objectCode: string
  recordId: string
  visible: boolean
}>()

const emit = defineEmits<{
  (e: 'update:visible', val: boolean): void
}>()

// ==================== 对话框可见性 ====================
const dialogVisible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val)
})

// ==================== 共享列表 ====================
const shareList = ref<RecordShareItem[]>([])

async function loadShareList() {
  if (!props.objectCode || !props.recordId) return
  try {
    shareList.value = await listRecordShares(props.objectCode, props.recordId)
  } catch { /* 静默 */ }
}

// ==================== 表单 ====================
const form = reactive({
  share_to_type: 'USER' as string,
  share_to_id: '' as string,
  access_level: 'READ' as string,
  expire_at: null as string | null
})

const canSubmit = computed(() => form.share_to_id !== '')

// ==================== 目标选项 ====================
interface SelectOption {
  value: string
  label: string
}
const targetOptions = ref<SelectOption[]>([])

async function loadTargetOptions() {
  targetOptions.value = []
  form.share_to_id = ''
  try {
    if (form.share_to_type === 'USER') {
      const res = await listUsers({ page: 1, page_size: 200 })
      targetOptions.value = (res.list || []).map((u: UserPageItem) => ({
        value: u.id,
        label: u.username
      }))
    } else if (form.share_to_type === 'ROLE') {
      const list = await getRoleList()
      const flat: SelectOption[] = []
      function flatten(nodes: RoleItem[]) {
        for (const n of nodes) {
          flat.push({ value: n.id, label: n.role_name })
          if (n.children) flatten(n.children)
        }
      }
      flatten(list)
      targetOptions.value = flat
    } else if (form.share_to_type === 'DEPT') {
      // 部门暂无独立 API，展示占位提示
      targetOptions.value = []
    }
  } catch { /* 静默 */ }
}

function handleTargetTypeChange() {
  loadTargetOptions()
}

// ==================== 事件处理 ====================
async function handleSubmit() {
  if (!form.share_to_id) {
    ElMessage.warning('请选择共享目标')
    return
  }
  try {
    await createRecordShare({
      object_code: props.objectCode,
      record_id: props.recordId,
      share_to_type: form.share_to_type,
      share_to_id: form.share_to_id,
      access_level: form.access_level,
      expire_at: form.expire_at || null
    })
    ElMessage.success('共享规则已添加')
    form.share_to_id = ''
    form.expire_at = null
    await loadShareList()
  } catch { /* 已由拦截器处理 */ }
}

async function handleDelete(id: string) {
  try {
    await ElMessageBox.confirm('确定删除该共享规则？', '提示', { type: 'warning' })
    await deleteRecordShare(id)
    ElMessage.success('已删除')
    await loadShareList()
  } catch { /* 取消或错误 */ }
}

function handleClose() {
  emit('update:visible', false)
}

function shareToTypeLabel(type: string): string {
  const map: Record<string, string> = { USER: '用户', ROLE: '角色', DEPT: '部门' }
  return map[type] || type
}

// ==================== 生命周期 ====================
watch(() => props.visible, (val) => {
  if (val) {
    loadShareList()
    loadTargetOptions()
  }
})
</script>

<style scoped>
.section-title {
  font-weight: 600;
  font-size: 14px;
  margin-bottom: 12px;
  color: #303133;
}
</style>
