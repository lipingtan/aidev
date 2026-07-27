<template>
  <div class="flow-config-page">
    <el-card shadow="never">
      <template #header>
        <div class="card-header">
          <span>审批流定义</span>
          <el-button type="primary" :icon="Plus" @click="openCreate">新建流程</el-button>
        </div>
      </template>

      <!-- 列表 -->
      <el-table
        v-loading="loading"
        :data="tableData"
        border
        style="width: 100%"
      >
        <el-table-column prop="flow_code" label="流程编码" width="160" />
        <el-table-column prop="flow_name" label="流程名称" min-width="160">
          <template #default="{ row }">
            {{ row.flow_name }}
            <el-tag v-if="row.is_global" size="small" type="info" style="margin-left: 6px">系统默认</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="节点数" width="80" align="center">
          <template #default="{ row }">{{ row.flow_config?.length || 0 }}</template>
        </el-table-column>
        <el-table-column prop="created_at" label="创建时间" width="160" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <template v-if="!row.is_global">
              <el-button size="small" @click="openEdit(row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
            </template>
            <el-button v-else size="small" @click="openView(row)">查看</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 空态 -->
      <el-empty v-if="!loading && tableData.length === 0" description="暂无审批流定义" />

      <!-- 分页 -->
      <div v-if="total > 0" class="pagination-wrap">
        <el-pagination
          v-model:current-page="page"
          v-model:page-size="pageSize"
          :total="total"
          :page-sizes="[10, 20, 50]"
          layout="total, sizes, prev, pager, next"
          @change="loadList"
        />
      </div>
    </el-card>

    <!-- 创建/编辑 弹窗 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="760px"
      destroy-on-close
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="100px"
        :disabled="viewOnly"
      >
        <el-form-item label="流程编码" prop="flow_code">
          <el-input v-model="form.flow_code" placeholder="唯一标识，如 tenant_app_subscription" :disabled="!!form.id" />
        </el-form-item>
        <el-form-item label="流程名称" prop="flow_name">
          <el-input v-model="form.flow_name" placeholder="流程名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="2" />
        </el-form-item>

        <!-- 节点列表 -->
        <el-divider content-position="left">审批节点</el-divider>
        <div
          v-for="(node, idx) in form.flow_config"
          :key="idx"
          class="node-item"
        >
          <div class="node-item__title">
            <span>节点 {{ idx + 1 }}</span>
            <el-button
              v-if="!viewOnly"
              size="small"
              type="danger"
              link
              @click="removeNode(idx)"
            >移除</el-button>
          </div>
          <el-row :gutter="12">
            <el-col :span="12">
              <el-form-item
                :label="'节点类型'"
                :prop="`flow_config.${idx}.node_type`"
                :rules="{ required: true, message: '请选择节点类型' }"
              >
                <el-select v-model="node.node_type" placeholder="节点类型" style="width: 100%">
                  <el-option label="单人审批 (SINGLE)" value="SINGLE" />
                  <el-option label="会签 (AND_SIGN)" value="AND_SIGN" />
                  <el-option label="或签 (OR_SIGN)" value="OR_SIGN" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="12">
              <el-form-item
                label="审批人类型"
                :prop="`flow_config.${idx}.assignee_type`"
                :rules="{ required: true, message: '请选择审批人类型' }"
              >
                <el-select
                  v-model="node.assignee_type"
                  placeholder="审批人类型"
                  style="width: 100%"
                  @change="() => { node.assignee_ids = [] }"
                >
                  <el-option label="指定用户 (USER)" value="USER" />
                  <el-option label="指定角色 (ROLE)" value="ROLE" />
                  <el-option label="部门负责人 (DEPT_HEAD)" value="DEPT_HEAD" />
                </el-select>
              </el-form-item>
            </el-col>
          </el-row>

          <!-- 审批人选择 -->
          <el-form-item
            v-if="node.assignee_type === 'USER'"
            label="审批用户"
            :prop="`flow_config.${idx}.assignee_ids`"
            :rules="{ required: true, type: 'array', min: 1, message: '请选择审批用户' }"
          >
            <el-select
              v-model="node.assignee_ids"
              multiple
              filterable
              placeholder="选择审批用户"
              style="width: 100%"
            >
              <el-option
                v-for="u in userOptions"
                :key="u.id"
                :label="u.nick_name || u.username"
                :value="String(u.id)"
              />
            </el-select>
          </el-form-item>

          <el-form-item
            v-else-if="node.assignee_type === 'ROLE'"
            label="审批角色"
            :prop="`flow_config.${idx}.assignee_ids`"
            :rules="{ required: true, type: 'array', min: 1, message: '请选择审批角色' }"
          >
            <el-select
              v-model="node.assignee_ids"
              multiple
              filterable
              placeholder="选择审批角色"
              style="width: 100%"
            >
              <el-option
                v-for="r in roleOptions"
                :key="r.id"
                :label="r.name"
                :value="String(r.id)"
              />
            </el-select>
          </el-form-item>

          <el-form-item
            v-else-if="node.assignee_type === 'DEPT_HEAD'"
            label="说明"
          >
            <span class="text-secondary">自动解析为部门负责人，无需手动指定</span>
          </el-form-item>

          <!-- 超时配置 -->
          <el-row :gutter="12">
            <el-col :span="8">
              <el-form-item label="超时时限(h)">
                <el-input-number
                  v-model="node.timeout_hours"
                  :min="0"
                  placeholder="0=不超时"
                  style="width: 100%"
                />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="超时动作">
                <el-select v-model="node.timeout_action" clearable placeholder="超时动作" style="width: 100%">
                  <el-option label="自动通过" value="AUTO_APPROVE" />
                  <el-option label="自动驳回" value="AUTO_REJECT" />
                  <el-option label="升级转派" value="ESCALATE" />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item v-if="node.timeout_action === 'ESCALATE'" label="升级目标">
                <el-input v-model="node.escalate_to" placeholder="用户/角色 ID" />
              </el-form-item>
            </el-col>
          </el-row>
        </div>

        <el-button
          v-if="!viewOnly"
          type="primary"
          plain
          :icon="Plus"
          style="width: 100%; margin-top: 4px"
          @click="addNode"
        >
          添加节点
        </el-button>
      </el-form>

      <template #footer>
        <el-button @click="dialogVisible = false">{{ viewOnly ? '关闭' : '取消' }}</el-button>
        <el-button v-if="!viewOnly" type="primary" :loading="saving" @click="handleSubmit">
          保存
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Plus } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, type FormInstance } from 'element-plus'
import {
  getApprovalFlows,
  createApprovalFlow,
  updateApprovalFlow,
  deleteApprovalFlow,
  type ApprovalFlow,
  type FlowNodeConfig
} from '@/api/approval'
import request from '@/utils/request'

// ==================== 列表 ====================

const loading = ref(false)
const tableData = ref<(ApprovalFlow & { is_global?: boolean })[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

async function loadList() {
  loading.value = true
  try {
    const res: any = await getApprovalFlows({ page: page.value, page_size: pageSize.value })
    tableData.value = Array.isArray(res) ? res : (res?.data || [])
    total.value = res?.total || tableData.value.length
  } catch {
    ElMessage.error('加载审批流列表失败')
  } finally {
    loading.value = false
  }
}

// ==================== 用户/角色选项 ====================

const userOptions = ref<any[]>([])
const roleOptions = ref<any[]>([])

async function loadOptions() {
  try {
    const [uRes, rRes]: any[] = await Promise.all([
      request.get('/api/v1/admin/sys-user', { params: { page: 1, page_size: 100 } }),
      request.get('/api/v1/admin/role', { params: { page: 1, page_size: 100 } })
    ])
    userOptions.value = Array.isArray(uRes) ? uRes : (uRes?.data || uRes?.list || [])
    roleOptions.value = Array.isArray(rRes) ? rRes : (rRes?.data || rRes?.list || [])
  } catch {
    // 选项加载失败不阻断主流程
  }
}

// ==================== 弹窗 ====================

const dialogVisible = ref(false)
const dialogTitle = ref('新建流程')
const viewOnly = ref(false)
const saving = ref(false)
const formRef = ref<FormInstance>()

const defaultNode = (): FlowNodeConfig => ({
  node_order: 1,
  node_type: 'SINGLE',
  assignee_type: 'USER',
  assignee_ids: [],
  timeout_hours: 0
})

const form = reactive<{
  id: string
  flow_code: string
  flow_name: string
  description: string
  flow_config: FlowNodeConfig[]
}>({
  id: '',
  flow_code: '',
  flow_name: '',
  description: '',
  flow_config: [defaultNode()]
})

const rules = {
  flow_code: [{ required: true, message: '请填写流程编码', trigger: 'blur' }],
  flow_name: [{ required: true, message: '请填写流程名称', trigger: 'blur' }]
}

function resetForm() {
  form.id = ''
  form.flow_code = ''
  form.flow_name = ''
  form.description = ''
  form.flow_config = [defaultNode()]
}

function openCreate() {
  resetForm()
  viewOnly.value = false
  dialogTitle.value = '新建流程'
  dialogVisible.value = true
}

function openEdit(row: ApprovalFlow) {
  form.id = row.id
  form.flow_code = row.flow_code
  form.flow_name = row.flow_name
  form.description = row.description || ''
  form.flow_config = row.flow_config?.length
    ? row.flow_config.map((n, i) => ({ ...n, node_order: i + 1 }))
    : [defaultNode()]
  viewOnly.value = false
  dialogTitle.value = '编辑流程'
  dialogVisible.value = true
}

function openView(row: ApprovalFlow) {
  openEdit(row)
  viewOnly.value = true
  dialogTitle.value = '查看流程（系统默认，只读）'
}

function addNode() {
  form.flow_config.push({
    ...defaultNode(),
    node_order: form.flow_config.length + 1
  })
}

function removeNode(idx: number) {
  form.flow_config.splice(idx, 1)
  form.flow_config.forEach((n, i) => { n.node_order = i + 1 })
}

async function handleSubmit() {
  await formRef.value?.validate()
  if (form.flow_config.length === 0) {
    ElMessage.warning('请至少添加一个审批节点')
    return
  }
  saving.value = true
  try {
    const payload = {
      flow_code: form.flow_code,
      flow_name: form.flow_name,
      description: form.description,
      flow_config: form.flow_config
    }
    if (form.id) {
      await updateApprovalFlow(form.id, payload)
      ElMessage.success('更新成功')
    } else {
      await createApprovalFlow(payload)
      ElMessage.success('创建成功')
    }
    dialogVisible.value = false
    loadList()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    saving.value = false
  }
}

async function handleDelete(row: ApprovalFlow) {
  try {
    await ElMessageBox.confirm(
      `确定删除流程「${row.flow_name}」？删除后不可恢复。`,
      '删除确认',
      { type: 'warning', confirmButtonText: '确定删除', cancelButtonText: '取消' }
    )
    await deleteApprovalFlow(row.id)
    ElMessage.success('删除成功')
    loadList()
  } catch (e: any) {
    if (e === 'cancel' || e?.toString?.().includes('cancel')) return
    ElMessage.error(e?.message || '删除失败')
  }
}

onMounted(() => {
  loadList()
  loadOptions()
})
</script>

<style scoped>
.flow-config-page {
  padding: 16px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}

.node-item {
  border: 1px solid var(--el-border-color);
  border-radius: 6px;
  padding: 12px 16px 0;
  margin-bottom: 12px;
  background: var(--el-fill-color-lighter);
}

.node-item__title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--el-text-color-primary);
}

.text-secondary {
  color: var(--el-text-color-secondary);
  font-size: 13px;
}
</style>
