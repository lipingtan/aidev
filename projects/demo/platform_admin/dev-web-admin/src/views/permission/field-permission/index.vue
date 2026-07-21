<template>
  <div class="field-permission-container">
    <!-- 顶部操作栏 -->
    <el-card shadow="never" class="top-bar">
      <el-row :gutter="16" align="middle">
        <el-col :span="8">
          <span class="label">选择角色：</span>
          <el-select v-model="selectedRoleId" placeholder="请选择角色" filterable style="width: 240px;">
            <el-option v-for="role in roleList" :key="role.id" :label="role.role_name" :value="role.id" />
          </el-select>
        </el-col>
        <el-col :span="16" style="text-align: right;">
          <el-button type="primary" :icon="Plus" @click="showAddObjectDialog = true">添加对象</el-button>
          <el-button type="success" :icon="Check" :disabled="!selectedRoleId || !selectedObject" @click="handleSave">
            保存权限
          </el-button>
        </el-col>
      </el-row>
    </el-card>

    <el-row :gutter="16" class="main-content">
      <!-- 左侧：对象列表 -->
      <el-col :span="6">
        <el-card shadow="never" class="object-list-card">
          <template #header>
            <span>字段对象</span>
          </template>
          <el-menu :default-active="selectedObject?.object_code" @select="handleObjectSelect">
            <el-menu-item v-for="obj in objectList" :key="obj.object_code" :index="obj.object_code">
              {{ obj.object_name }}({{ obj.object_code }})
            </el-menu-item>
          </el-menu>
          <el-empty v-if="objectList.length === 0" description="暂无已注册对象" />
        </el-card>
      </el-col>

      <!-- 右侧：字段列表 -->
      <el-col :span="18">
        <el-card shadow="never">
          <template #header>
            <el-row justify="space-between" align="middle">
              <span>{{ selectedObject ? `${selectedObject.object_name} - 字段列表` : '请选择左侧对象' }}</span>
              <el-button v-if="selectedObject" type="primary" :icon="Plus" size="small" @click="showAddFieldDialog = true">
                添加字段
              </el-button>
            </el-row>
          </template>

          <el-table v-if="selectedObject" :data="fieldTableData" border stripe>
            <el-table-column label="字段" min-width="240">
              <template #default="{ row }">
                <div class="field-name-cell">
                  <template v-if="editingFieldName === row.field_name">
                    <el-input
                      v-model="editingDescription"
                      size="small"
                      style="width: 160px;"
                      @keyup.enter="confirmEditDescription(row)"
                      @blur="confirmEditDescription(row)"
                    />
                  </template>
                  <template v-else>
                    <span class="description-text" @click="startEditDescription(row)">
                      {{ row.description }}
                    </span>
                  </template>
                  <span class="field-key">({{ row.field_name }})</span>
                </div>
              </template>
            </el-table-column>
            <el-table-column label="权限" width="200">
              <template #default="{ row }">
                <el-select v-model="row.access" size="small" style="width: 140px;">
                  <el-option label="可编辑" value="EDITABLE" />
                  <el-option label="可见" value="VISIBLE" />
                  <el-option label="隐藏" value="HIDDEN" />
                </el-select>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-else description="请先在左侧选择一个对象" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 添加对象弹窗 -->
    <el-dialog v-model="showAddObjectDialog" title="手动添加对象" width="440px">
      <el-form :model="addObjectForm" label-width="100px">
        <el-form-item label="对象编码" required>
          <el-input v-model="addObjectForm.object_code" placeholder="如 invoice" />
        </el-form-item>
        <el-form-item label="对象名称" required>
          <el-input v-model="addObjectForm.object_name" placeholder="如 发票" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddObjectDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAddObject">确定</el-button>
      </template>
    </el-dialog>

    <!-- 添加字段弹窗 -->
    <el-dialog v-model="showAddFieldDialog" title="手动添加字段" width="440px">
      <el-form :model="addFieldForm" label-width="100px">
        <el-form-item label="字段名" required>
          <el-input v-model="addFieldForm.field_name" placeholder="如 phone" />
        </el-form-item>
        <el-form-item label="描述" required>
          <el-input v-model="addFieldForm.description" placeholder="如 手机号" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddFieldDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAddField">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus, Check } from '@element-plus/icons-vue'
import { getRoleList, type RoleItem } from '@/api/role'
import {
  listFieldObjects,
  listFieldDefinitions,
  getFieldPermissions,
  setFieldPermissions,
  updateFieldDescription,
  manualRegisterObject,
  manualRegisterField,
  type FieldObject,
  type FieldDefinition
} from '@/api/field-permission'

// ==================== 角色 ====================
const roleList = ref<RoleItem[]>([])
const selectedRoleId = ref('')

// ==================== 对象 ====================
const objectList = ref<FieldObject[]>([])
const selectedObject = ref<FieldObject | null>(null)

// ==================== 字段表格 ====================
interface FieldRow {
  field_name: string
  description: string
  access: string
}
const fieldTableData = ref<FieldRow[]>([])

// ==================== 描述编辑 ====================
const editingFieldName = ref('')
const editingDescription = ref('')

// ==================== 弹窗 ====================
const showAddObjectDialog = ref(false)
const showAddFieldDialog = ref(false)
const addObjectForm = reactive({ object_code: '', object_name: '' })
const addFieldForm = reactive({ field_name: '', description: '' })

// ==================== 初始化 ====================
async function loadRoles() {
  try {
    const list = await getRoleList()
    // 展平树形角色
    const flat: RoleItem[] = []
    function flatten(nodes: RoleItem[]) {
      for (const n of nodes) {
        flat.push(n)
        if (n.children) flatten(n.children)
      }
    }
    flatten(list)
    roleList.value = flat
  } catch { /* 静默 */ }
}

async function loadObjects() {
  try {
    objectList.value = await listFieldObjects()
  } catch { /* 静默 */ }
}

async function loadFields(objectCode: string) {
  try {
    const definitions: FieldDefinition[] = await listFieldDefinitions(objectCode)
    // 加载权限配置
    let permMap: Record<string, string> = {}
    if (selectedRoleId.value) {
      const perms = await getFieldPermissions(selectedRoleId.value, objectCode)
      for (const p of perms) {
        permMap[p.field_name] = p.access
      }
    }
    fieldTableData.value = definitions.map((d) => ({
      field_name: d.field_name,
      description: d.description,
      access: permMap[d.field_name] || 'VISIBLE'
    }))
  } catch { /* 静默 */ }
}

// ==================== 事件处理 ====================
function handleObjectSelect(objectCode: string) {
  const obj = objectList.value.find((o) => o.object_code === objectCode) || null
  selectedObject.value = obj
  if (obj) {
    loadFields(obj.object_code)
  }
}

// 角色切换时重新加载权限
watch(selectedRoleId, () => {
  if (selectedObject.value) {
    loadFields(selectedObject.value.object_code)
  }
})

// 保存权限
async function handleSave() {
  if (!selectedRoleId.value || !selectedObject.value) return
  try {
    const items = fieldTableData.value.map((row) => ({
      field_name: row.field_name,
      access: row.access
    }))
    await setFieldPermissions(selectedRoleId.value, selectedObject.value.object_code, items)
    ElMessage.success('权限保存成功')
  } catch { /* 已由拦截器处理 */ }
}

// 描述编辑
function startEditDescription(row: FieldRow) {
  editingFieldName.value = row.field_name
  editingDescription.value = row.description
}

async function confirmEditDescription(row: FieldRow) {
  if (!selectedObject.value) return
  if (editingDescription.value && editingDescription.value !== row.description) {
    try {
      await updateFieldDescription(selectedObject.value.object_code, row.field_name, editingDescription.value)
      row.description = editingDescription.value
      ElMessage.success('描述已更新')
    } catch { /* 已由拦截器处理 */ }
  }
  editingFieldName.value = ''
}

// 添加对象
async function handleAddObject() {
  if (!addObjectForm.object_code || !addObjectForm.object_name) {
    ElMessage.warning('请填写完整信息')
    return
  }
  try {
    await manualRegisterObject(addObjectForm.object_code, addObjectForm.object_name)
    ElMessage.success('对象添加成功')
    showAddObjectDialog.value = false
    addObjectForm.object_code = ''
    addObjectForm.object_name = ''
    await loadObjects()
  } catch { /* 已由拦截器处理 */ }
}

// 添加字段
async function handleAddField() {
  if (!selectedObject.value || !addFieldForm.field_name || !addFieldForm.description) {
    ElMessage.warning('请填写完整信息')
    return
  }
  try {
    await manualRegisterField(selectedObject.value.object_code, addFieldForm.field_name, addFieldForm.description)
    ElMessage.success('字段添加成功')
    showAddFieldDialog.value = false
    addFieldForm.field_name = ''
    addFieldForm.description = ''
    await loadFields(selectedObject.value.object_code)
  } catch { /* 已由拦截器处理 */ }
}

// 页面挂载
loadRoles()
loadObjects()
</script>

<style scoped>
.field-permission-container {
  padding: 16px;
}
.top-bar {
  margin-bottom: 16px;
}
.top-bar .label {
  font-weight: 500;
  margin-right: 8px;
}
.main-content {
  min-height: 500px;
}
.object-list-card .el-menu {
  border-right: none;
}
.field-name-cell {
  display: flex;
  align-items: center;
  gap: 4px;
}
.description-text {
  cursor: pointer;
  border-bottom: 1px dashed #409eff;
  color: #409eff;
}
.description-text:hover {
  color: #337ecc;
}
.field-key {
  color: #909399;
  font-size: 12px;
}
</style>
