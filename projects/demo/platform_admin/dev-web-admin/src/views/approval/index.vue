<template>
  <div class="approval-page">
    <el-card shadow="never">
      <template #header>
        <el-tabs v-model="activeTab" @tab-change="handleTabChange">
          <el-tab-pane label="待我审批" name="pending" />
          <el-tab-pane label="我发起的" name="mine" />
        </el-tabs>
      </template>

      <el-table v-loading="loading" :data="tableData" border style="width: 100%">
        <el-table-column prop="flow_name" label="流程名称" min-width="140" show-overflow-tooltip />
        <el-table-column prop="biz_type" label="业务类型" width="160" />
        <el-table-column prop="biz_id" label="业务ID" width="120" show-overflow-tooltip />
        <el-table-column label="状态" width="100" align="center">
          <template #default="{ row }">
            <el-tag :type="statusTagType(row.status)" size="small">{{ statusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="applicant_name" label="发起人" width="100" />
        <el-table-column prop="created_at" label="发起时间" width="160" />
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button size="small" @click="goDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-empty v-if="!loading && tableData.length === 0" description="暂无审批记录" />

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
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { getApprovals, type ApprovalInstance } from '@/api/approval'

const router = useRouter()

const activeTab = ref<'pending' | 'mine'>('pending')
const loading = ref(false)
const tableData = ref<ApprovalInstance[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)

async function loadList() {
  loading.value = true
  try {
    const res: any = await getApprovals({
      view: activeTab.value,
      page: page.value,
      page_size: pageSize.value
    })
    tableData.value = Array.isArray(res) ? res : (res?.data || [])
    total.value = res?.total || tableData.value.length
  } catch {
    ElMessage.error('加载审批列表失败')
  } finally {
    loading.value = false
  }
}

function handleTabChange() {
  page.value = 1
  loadList()
}

function goDetail(row: ApprovalInstance) {
  router.push({ name: 'approval-detail', params: { id: row.id } })
}

function statusLabel(status: string) {
  const map: Record<string, string> = {
    PENDING: '审批中',
    APPROVED: '已通过',
    REJECTED: '已驳回',
    CANCELLED: '已撤回'
  }
  return map[status] || status
}

function statusTagType(status: string) {
  const map: Record<string, string> = {
    PENDING: 'warning',
    APPROVED: 'success',
    REJECTED: 'danger',
    CANCELLED: 'info'
  }
  return (map[status] || '') as any
}

onMounted(() => loadList())
</script>

<style scoped>
.approval-page {
  padding: 16px;
}

.pagination-wrap {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
}
</style>
