<template>
  <div class="approval-detail-page">
    <el-card shadow="never" v-loading="loading">
      <template #header>
        <div class="card-header">
          <el-button :icon="ArrowLeft" text @click="router.back()">返回</el-button>
          <span class="title">审批详情</span>
        </div>
      </template>

      <template v-if="detail">
        <!-- 基本信息 -->
        <el-descriptions :column="2" border class="info-section">
          <el-descriptions-item label="流程名称">{{ detail.flow_name || detail.flow_code }}</el-descriptions-item>
          <el-descriptions-item label="业务类型">{{ detail.biz_type }}</el-descriptions-item>
          <el-descriptions-item label="业务ID">{{ detail.biz_id }}</el-descriptions-item>
          <el-descriptions-item label="发起人">{{ detail.applicant_name || detail.applicant_id }}</el-descriptions-item>
          <el-descriptions-item label="发起时间">{{ detail.created_at }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="statusTagType(detail.status)" size="small">{{ statusLabel(detail.status) }}</el-tag>
          </el-descriptions-item>
          <el-descriptions-item v-if="detail.cancel_reason" label="撤回原因" :span="2">
            {{ detail.cancel_reason }}
          </el-descriptions-item>
        </el-descriptions>

        <!-- 审批节点 -->
        <div class="section-title">审批节点</div>
        <el-timeline v-if="detail.nodes?.length">
          <el-timeline-item
            v-for="node in detail.nodes"
            :key="node.id"
            :type="nodeTimelineType(node.status)"
            :timestamp="`节点 ${node.node_order} · ${nodeTypeLabel(node.node_type)}`"
            placement="top"
          >
            <el-card shadow="never" class="node-card">
              <div class="node-info">
                <span class="node-status">
                  <el-tag :type="nodeTagType(node.status)" size="small">{{ nodeStatusLabel(node.status) }}</el-tag>
                </span>
                <span class="node-assignee-type">审批人类型：{{ assigneeTypeLabel(node.assignee_type) }}</span>
              </div>
              <div v-if="node.approve_comment" class="node-comment">
                <span class="label">审批意见：</span>{{ node.approve_comment }}
              </div>
              <div v-if="node.reject_reason" class="node-reject">
                <span class="label">驳回原因：</span>{{ node.reject_reason }}
              </div>
              <div v-if="node.assignee_note" class="node-note">
                <span class="label">备注：</span>{{ node.assignee_note }}
              </div>
            </el-card>
          </el-timeline-item>
        </el-timeline>
        <el-empty v-else description="暂无节点信息" />

        <!-- 操作按钮 -->
        <div class="action-bar">
          <!-- 撤回：发起人且实例 PENDING -->
          <el-button
            v-if="canCancel"
            type="warning"
            @click="handleCancel"
            :loading="submitting"
          >
            撤回申请
          </el-button>

          <!-- 审批通过 / 驳回：当前用户是合法审批人且节点 PENDING -->
          <template v-if="canApprove">
            <el-button type="success" :loading="submitting" @click="handleApprove">
              审批通过
            </el-button>
            <el-button type="danger" :loading="submitting" @click="rejectDialogVisible = true">
              驳回
            </el-button>
          </template>
        </div>
      </template>
    </el-card>

    <!-- 审批通过：可选填备注 -->
    <el-dialog v-model="approveDialogVisible" title="审批通过" width="440px" destroy-on-close>
      <el-form>
        <el-form-item label="审批备注">
          <el-input
            v-model="approveComment"
            type="textarea"
            :rows="3"
            placeholder="选填，可填写审批意见"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approveDialogVisible = false">取消</el-button>
        <el-button type="success" :loading="submitting" @click="submitApprove">确认通过</el-button>
      </template>
    </el-dialog>

    <!-- 驳回：必填原因 -->
    <el-dialog v-model="rejectDialogVisible" title="驳回申请" width="440px" destroy-on-close>
      <el-form ref="rejectFormRef" :model="rejectForm" :rules="rejectRules">
        <el-form-item label="驳回原因" prop="reason">
          <el-input
            v-model="rejectForm.reason"
            type="textarea"
            :rows="3"
            placeholder="请填写驳回原因（必填）"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="rejectDialogVisible = false">取消</el-button>
        <el-button type="danger" :loading="submitting" @click="submitReject">确认驳回</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ArrowLeft } from '@element-plus/icons-vue'
import { ElMessage, type FormInstance } from 'element-plus'
import {
  getApproval,
  approveApproval,
  rejectApproval,
  cancelApproval,
  type ApprovalDetail,
  type ApprovalNode
} from '@/api/approval'
import { useUserStore } from '@/store/modules/user'

const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const loading = ref(false)
const submitting = ref(false)
const detail = ref<ApprovalDetail | null>(null)

// ==================== 加载详情 ====================

async function loadDetail() {
  const id = route.params.id as string
  if (!id) return
  loading.value = true
  try {
    const res: any = await getApproval(id)
    detail.value = res
  } catch {
    ElMessage.error('加载审批详情失败')
  } finally {
    loading.value = false
  }
}

// ==================== 权限判断 ====================

/** 当前待处理节点（PENDING 状态） */
const pendingNode = computed<ApprovalNode | null>(() => {
  return detail.value?.nodes?.find((n) => n.status === 'PENDING') ?? null
})

/** 当前用户是否是合法审批人 */
const canApprove = computed(() => {
  if (!pendingNode.value) return false
  return pendingNode.value.assignee_user_ids.includes(userStore.userId)
})

/** 当前用户是否可撤回（发起人且实例 PENDING） */
const canCancel = computed(() => {
  if (!detail.value) return false
  return (
    detail.value.status === 'PENDING' &&
    detail.value.applicant_id === userStore.userId
  )
})

// ==================== 审批通过 ====================

const approveDialogVisible = ref(false)
const approveComment = ref('')

function handleApprove() {
  approveComment.value = ''
  approveDialogVisible.value = true
}

async function submitApprove() {
  if (!detail.value) return
  submitting.value = true
  try {
    await approveApproval(detail.value.id, { comment: approveComment.value || undefined })
    ElMessage.success('审批通过')
    approveDialogVisible.value = false
    loadDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

// ==================== 驳回 ====================

const rejectDialogVisible = ref(false)
const rejectFormRef = ref<FormInstance>()
const rejectForm = reactive({ reason: '' })
const rejectRules = {
  reason: [{ required: true, message: '请填写驳回原因', trigger: 'blur' }]
}

async function submitReject() {
  await rejectFormRef.value?.validate()
  if (!detail.value) return
  submitting.value = true
  try {
    await rejectApproval(detail.value.id, { reason: rejectForm.reason })
    ElMessage.success('已驳回')
    rejectDialogVisible.value = false
    loadDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || '操作失败')
  } finally {
    submitting.value = false
  }
}

// ==================== 撤回 ====================

async function handleCancel() {
  if (!detail.value) return
  submitting.value = true
  try {
    await cancelApproval(detail.value.id)
    ElMessage.success('已撤回')
    loadDetail()
  } catch (e: any) {
    ElMessage.error(e?.message || '撤回失败')
  } finally {
    submitting.value = false
  }
}

// ==================== 工具函数 ====================

function statusLabel(s: string) {
  return { PENDING: '审批中', APPROVED: '已通过', REJECTED: '已驳回', CANCELLED: '已撤回' }[s] || s
}
function statusTagType(s: string) {
  return ({ PENDING: 'warning', APPROVED: 'success', REJECTED: 'danger', CANCELLED: 'info' }[s] || '') as any
}
function nodeStatusLabel(s: string) {
  return { WAITING: '等待中', PENDING: '待审批', APPROVED: '已通过', REJECTED: '已驳回', SKIPPED: '已跳过', TIMEOUT: '已超时' }[s] || s
}
function nodeTagType(s: string) {
  return ({ WAITING: 'info', PENDING: 'warning', APPROVED: 'success', REJECTED: 'danger', SKIPPED: 'info', TIMEOUT: 'danger' }[s] || '') as any
}
function nodeTimelineType(s: string) {
  return ({ WAITING: 'info', PENDING: 'warning', APPROVED: 'success', REJECTED: 'danger', SKIPPED: '', TIMEOUT: 'danger' }[s] || '') as any
}
function nodeTypeLabel(t: string) {
  return { SINGLE: '单人审批', AND_SIGN: '会签', OR_SIGN: '或签' }[t] || t
}
function assigneeTypeLabel(t: string) {
  return { USER: '指定用户', ROLE: '指定角色', DEPT_HEAD: '部门负责人' }[t] || t
}

onMounted(() => loadDetail())
</script>

<style scoped>
.approval-detail-page {
  padding: 16px;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 8px;
}

.title {
  font-size: 16px;
  font-weight: 600;
}

.info-section {
  margin-bottom: 24px;
}

.section-title {
  font-size: 15px;
  font-weight: 600;
  margin: 20px 0 12px;
  color: var(--el-text-color-primary);
}

.node-card {
  margin-bottom: 4px;
}

.node-info {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 6px;
}

.node-comment,
.node-reject,
.node-note {
  font-size: 13px;
  margin-top: 4px;
  color: var(--el-text-color-regular);
}

.node-reject {
  color: var(--el-color-danger);
}

.label {
  font-weight: 500;
  color: var(--el-text-color-secondary);
}

.action-bar {
  margin-top: 24px;
  display: flex;
  gap: 12px;
}
</style>
