<template>
  <div class="field-objects-container">
    <el-row :gutter="16">
      <!-- 左侧：对象列表 -->
      <el-col :span="8">
        <el-card shadow="never">
          <template #header>
            <el-row justify="space-between" align="middle">
              <span>字段对象</span>
              <el-button type="primary" :icon="Plus" size="small" @click="showAddObjectDialog = true">添加对象</el-button>
            </el-row>
          </template>
          <el-table :data="objectList" highlight-current-row size="small" @current-change="handleObjectSelect">
            <el-table-column prop="object_name" label="名称" />
            <el-table-column prop="object_code" label="编码" width="120" />
            <el-table-column prop="source" label="来源" width="70">
              <template #default="{ row }">
                <el-tag :type="row.source === 'AUTO' ? 'info' : 'success'" size="small">
                  {{ row.source === 'AUTO' ? '自动' : '手动' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-if="objectList.length === 0" description="暂无已注册对象" :image-size="60" />
        </el-card>
      </el-col>

      <!-- 右侧：字段列表 -->
      <el-col :span="16">
        <el-card shadow="never">
          <template #header>
            <el-row justify="space-between" align="middle">
              <span>{{ selectedObject ? `${selectedObject.object_name} - 字段定义` : '请选择左侧对象' }}</span>
              <div v-if="selectedObject">
                <el-button type="primary" :icon="Plus" size="small" @click="showAddFieldDialog = true">添加字段</el-button>
                <el-button size="small" @click="showHelpDrawer = true">权限配置演示</el-button>
              </div>
            </el-row>
          </template>

          <el-table v-if="selectedObject" :data="fieldList" border stripe size="small">
            <el-table-column prop="field_name" label="字段名" width="160" />
            <el-table-column label="描述" min-width="200">
              <template #default="{ row }">
                <template v-if="editingFieldName === row.field_name">
                  <el-input
                    v-model="editingDescription"
                    size="small"
                    style="width: 200px;"
                    @keyup.enter="confirmEditDescription(row)"
                    @blur="confirmEditDescription(row)"
                  />
                </template>
                <template v-else>
                  <span class="editable-text" @click="startEditDescription(row)">
                    {{ row.description }}
                  </span>
                </template>
              </template>
            </el-table-column>
            <el-table-column prop="source" label="来源" width="70">
              <template #default="{ row }">
                <el-tag :type="row.source === 'AUTO' ? 'info' : 'success'" size="small">
                  {{ row.source === 'AUTO' ? '自动' : '手动' }}
                </el-tag>
              </template>
            </el-table-column>
          </el-table>
          <el-empty v-else description="请先在左侧选择一个对象" :image-size="80" />
        </el-card>
      </el-col>
    </el-row>

    <!-- 添加对象弹窗 -->
    <el-dialog v-model="showAddObjectDialog" title="添加字段对象" width="440px">
      <el-form :model="addObjectForm" label-width="100px">
        <el-form-item label="对象编码" required>
          <el-input v-model="addObjectForm.object_code" placeholder="如 invoice、order" />
        </el-form-item>
        <el-form-item label="对象名称" required>
          <el-input v-model="addObjectForm.object_name" placeholder="如 发票、订单" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddObjectDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAddObject">确定</el-button>
      </template>
    </el-dialog>

    <!-- 添加字段弹窗 -->
    <el-dialog v-model="showAddFieldDialog" title="添加字段定义" width="440px">
      <el-form :model="addFieldForm" label-width="100px">
        <el-form-item label="字段名" required>
          <el-input v-model="addFieldForm.field_name" placeholder="如 phone、amount" />
        </el-form-item>
        <el-form-item label="描述" required>
          <el-input v-model="addFieldForm.description" placeholder="如 手机号、金额" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showAddFieldDialog = false">取消</el-button>
        <el-button type="primary" @click="handleAddField">确定</el-button>
      </template>
    </el-dialog>

    <!-- 帮助 Drawer：展示权限配置演示（原独立页面内容） -->
    <el-drawer v-model="showHelpDrawer" title="字段权限配置演示" size="70%" destroy-on-close>
      <FieldPermissionDemo />
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import {
  listFieldObjects,
  listFieldDefinitions,
  updateFieldDescription,
  manualRegisterObject,
  manualRegisterField,
  type FieldObject,
  type FieldDefinition
} from '@/api/field-permission'
import FieldPermissionDemo from '@/views/permission/field-permission/index.vue'

// 对象列表
const objectList = ref<FieldObject[]>([])
const selectedObject = ref<FieldObject | null>(null)

// 字段列表
const fieldList = ref<FieldDefinition[]>([])

// 编辑描述
const editingFieldName = ref('')
const editingDescription = ref('')

// 弹窗
const showAddObjectDialog = ref(false)
const showAddFieldDialog = ref(false)
const showHelpDrawer = ref(false)
const addObjectForm = reactive({ object_code: '', object_name: '' })
const addFieldForm = reactive({ field_name: '', description: '' })

async function loadObjects() {
  try {
    objectList.value = await listFieldObjects()
  } catch { /* 静默 */ }
}

async function loadFields(objectCode: string) {
  try {
    fieldList.value = await listFieldDefinitions(objectCode)
  } catch { /* 静默 */ }
}

function handleObjectSelect(row: FieldObject | null) {
  selectedObject.value = row
  if (row) loadFields(row.object_code)
}

function startEditDescription(row: FieldDefinition) {
  editingFieldName.value = row.field_name
  editingDescription.value = row.description
}

async function confirmEditDescription(row: FieldDefinition) {
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

onMounted(() => {
  loadObjects()
})
</script>

<style scoped>
.field-objects-container {
  padding: 16px;
}
.editable-text {
  cursor: pointer;
  border-bottom: 1px dashed #409eff;
  color: #409eff;
}
.editable-text:hover {
  color: #337ecc;
}
/* 窄屏：两栏改为上下堆叠 */
@media (max-width: 767px) {
  .field-objects-container :deep(.el-col-8),
  .field-objects-container :deep(.el-col-16) {
    width: 100% !important;
    max-width: 100% !important;
    flex: 0 0 100% !important;
  }
}
</style>
