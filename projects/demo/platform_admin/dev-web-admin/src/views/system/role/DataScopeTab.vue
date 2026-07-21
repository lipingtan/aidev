<template>
  <div class="perm-tab">
    <el-table v-loading="loading" :data="scopeRows" border style="width: 100%">
      <el-table-column prop="dimension_label" label="维度" width="150" />
      <el-table-column label="范围类型" width="160">
        <template #default="{ row }">
          <el-select
            v-model="row.scope_type"
            placeholder="选择范围类型"
            @change="handleScopeTypeChange(row); isDirty = true"
          >
            <el-option
              v-for="st in row.supported_scope_types"
              :key="st"
              :label="scopeTypeLabel(st)"
              :value="st"
            />
          </el-select>
        </template>
      </el-table-column>
      <el-table-column label="值">
        <template #default="{ row }">
          <!-- 仅 CUSTOM 需要输入值列表 -->
          <template v-if="row.scope_type === 'CUSTOM'">
            <el-select
              v-if="row.options && row.options.length > 0"
              v-model="row.values"
              multiple
              filterable
              placeholder="请选择"
              style="width: 100%"
              @change="isDirty = true"
            >
              <el-option
                v-for="opt in row.options"
                :key="opt.value"
                :label="opt.label"
                :value="opt.value"
              />
            </el-select>
            <el-select
              v-else
              v-model="row.values"
              multiple
              filterable
              allow-create
              default-first-option
              placeholder="请输入（回车添加）"
              style="width: 100%"
              @change="isDirty = true"
            />
          </template>
          <template v-else-if="row.scope_type === 'ALL'">
            <el-tag type="success" size="small">全部数据（不过滤）</el-tag>
          </template>
          <template v-else-if="row.scope_type === 'SELF'">
            <el-tag type="info" size="small">仅本人创建的数据</el-tag>
          </template>
          <template v-else-if="row.scope_type === 'DEPT'">
            <el-tag type="warning" size="small">本组织节点数据（运行时动态解析）</el-tag>
          </template>
          <template v-else-if="row.scope_type === 'DEPT_TREE'">
            <el-tag type="warning" size="small">本组织及下级节点数据（运行时动态解析）</el-tag>
          </template>
        </template>
      </el-table-column>
    </el-table>
    <div class="perm-tab-footer">
      <el-button type="primary" :loading="submitting" @click="handleSave">保存</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { getDataScopeConfigs, getRoleDataScopes, assignRoleDataScopes } from '@/api/role'
import type { DataScopeDimension, DataScopeConfig } from '@/api/role'

interface ScopeRow {
  dimension: string
  dimension_label: string
  scope_type: string
  supported_scope_types: string[]
  values: string[]
  options?: { value: string; label: string }[]
}

const props = defineProps<{ roleId: string }>()

const loading = ref(false)
const submitting = ref(false)
const scopeRows = ref<ScopeRow[]>([])
const isDirty = ref(false)

defineExpose({ isDirty })

const scopeTypeLabel = (type: string) => {
  const map: Record<string, string> = {
    ALL: '全部',
    SELF: '仅自己',
    DEPT: '本组织',
    DEPT_TREE: '本组织及下级',
    CUSTOM: '自定义',
  }
  return map[type] || type
}

const handleScopeTypeChange = (row: ScopeRow) => {
  // 切换到非 CUSTOM 时清空 values
  if (row.scope_type !== 'CUSTOM') {
    row.values = []
  }
}

async function loadData() {
  loading.value = true
  try {
    const [dimensions, currentScopes] = await Promise.all([
      getDataScopeConfigs(),
      getRoleDataScopes(props.roleId)
    ])
    // 将维度配置与角色当前值合并
    scopeRows.value = dimensions.map((dim: DataScopeDimension) => {
      const existing = currentScopes.find((s: DataScopeConfig) => s.dimension === dim.dimension_name)
      return {
        dimension: dim.dimension_name,
        dimension_label: dim.display_name,
        scope_type: existing?.scope_type || 'CUSTOM',
        supported_scope_types: dim.supported_scope_types || ['ALL', 'SELF', 'CUSTOM'],
        values: existing?.values || [],
        options: dim.options || []
      }
    })
  } finally {
    loading.value = false
  }
}

async function handleSave() {
  submitting.value = true
  try {
    const scopes: DataScopeConfig[] = scopeRows.value
      .filter((r) => r.scope_type !== '' && (r.scope_type !== 'CUSTOM' || r.values.length > 0))
      .map((r) => ({
        dimension: r.dimension,
        dimension_label: r.dimension_label,
        scope_type: r.scope_type,
        values: r.scope_type === 'CUSTOM' ? r.values : []
      }))
    await assignRoleDataScopes(props.roleId, scopes)
    isDirty.value = false
    ElMessage.success('数据权限保存成功')
  } finally {
    submitting.value = false
  }
}

watch(() => props.roleId, () => { isDirty.value = false; loadData() }, { immediate: true })
</script>

<style scoped>
.perm-tab { min-height: 200px; }
.perm-tab-footer { margin-top: 16px; text-align: right; }
</style>
