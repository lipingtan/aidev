<template>
  <div class="perm-tab">
    <el-table v-loading="loading" :data="scopeRows" border style="width: 100%">
      <el-table-column prop="dimension_label" label="维度" width="150" />
      <el-table-column label="值">
        <template #default="{ row }">
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
  values: string[]
  options?: { value: string; label: string }[]
}

const props = defineProps<{ roleId: string }>()

const loading = ref(false)
const submitting = ref(false)
const scopeRows = ref<ScopeRow[]>([])
const isDirty = ref(false)

defineExpose({ isDirty })

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
      .filter((r) => r.values.length > 0)
      .map((r) => ({
        dimension: r.dimension,
        dimension_label: r.dimension_label,
        values: r.values
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
