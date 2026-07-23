<template>
  <div class="field-perm-tab">
    <el-row :gutter="16">
      <!-- 左侧：对象列表 -->
      <el-col :span="6">
        <div class="object-panel">
          <div class="panel-header">
            <span>字段对象</span>
          </div>
          <el-menu :default-active="selectedObject?.object_code" @select="handleObjectSelect">
            <el-menu-item v-for="obj in objectList" :key="obj.object_code" :index="obj.object_code">
              {{ obj.object_name }}
            </el-menu-item>
          </el-menu>
          <el-empty v-if="objectList.length === 0" description="暂无已注册对象" :image-size="60" />
        </div>
      </el-col>

      <!-- 右侧：字段配置 -->
      <el-col :span="18">
        <div v-if="selectedObject" class="field-panel">
          <div class="panel-header">
            <span>{{ selectedObject.object_name }} - 字段权限</span>
            <el-button type="primary" size="small" :disabled="!dirty" @click="handleSave">保存</el-button>
          </div>
          <el-table :data="fieldTableData" border stripe size="small">
            <el-table-column label="字段" min-width="200">
              <template #default="{ row }">
                <span class="field-desc">{{ row.description }}</span>
                <span class="field-key">({{ row.field_name }})</span>
              </template>
            </el-table-column>
            <el-table-column label="权限" width="160">
              <template #default="{ row }">
                <el-select v-model="row.access" size="small" @change="dirty = true">
                  <el-option label="可编辑" value="EDITABLE" />
                  <el-option label="可见" value="VISIBLE" />
                  <el-option label="隐藏" value="HIDDEN" />
                </el-select>
              </template>
            </el-table-column>
          </el-table>
        </div>
        <el-empty v-else description="请在左侧选择一个对象" :image-size="80" />
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  listFieldObjects,
  listFieldDefinitions,
  getFieldPermissions,
  setFieldPermissions,
  type FieldObject,
  type FieldDefinition
} from '@/api/field-permission'

const props = defineProps<{
  roleId: string
}>()

// 对象列表
const objectList = ref<FieldObject[]>([])
const selectedObject = ref<FieldObject | null>(null)

// 字段表格
interface FieldRow {
  field_name: string
  description: string
  access: string
}
const fieldTableData = ref<FieldRow[]>([])
const dirty = ref(false)

async function loadObjects() {
  try {
    objectList.value = await listFieldObjects()
  } catch { /* 静默 */ }
}

async function loadFields(objectCode: string) {
  try {
    const definitions: FieldDefinition[] = await listFieldDefinitions(objectCode)
    let permMap: Record<string, string> = {}
    if (props.roleId) {
      const perms = await getFieldPermissions(props.roleId, objectCode)
      for (const p of perms) {
        permMap[p.field_name] = p.access
      }
    }
    fieldTableData.value = definitions.map((d) => ({
      field_name: d.field_name,
      description: d.description,
      access: permMap[d.field_name] || 'EDITABLE'
    }))
    dirty.value = false
  } catch { /* 静默 */ }
}

function handleObjectSelect(objectCode: string) {
  const obj = objectList.value.find((o) => o.object_code === objectCode) || null
  selectedObject.value = obj
  if (obj) loadFields(obj.object_code)
}

async function handleSave() {
  if (!props.roleId || !selectedObject.value) return
  try {
    const items = fieldTableData.value.map((row) => ({
      field_name: row.field_name,
      access: row.access
    }))
    await setFieldPermissions(props.roleId, selectedObject.value.object_code, items)
    dirty.value = false
    ElMessage.success('字段权限保存成功')
  } catch { /* 已由拦截器处理 */ }
}

// roleId 变更时重新加载
watch(() => props.roleId, () => {
  if (selectedObject.value) {
    loadFields(selectedObject.value.object_code)
  }
})

onMounted(() => {
  loadObjects()
})
</script>

<style scoped>
.field-perm-tab {
  min-height: 400px;
}
.object-panel, .field-panel {
  border: 1px solid #ebeef5;
  border-radius: 4px;
}
.panel-header {
  padding: 10px 16px;
  border-bottom: 1px solid #ebeef5;
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 500;
}
.object-panel .el-menu {
  border-right: none;
}
.field-desc {
  margin-right: 4px;
}
.field-key {
  color: #909399;
  font-size: 12px;
}
</style>
