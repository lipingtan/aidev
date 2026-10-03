<template>
  <el-dialog v-model="visible" title="字典数据管理" :width="isMobile ? '95%' : '800px'" destroy-on-close class="dict-data-dialog">
    <div class="dialog-body" :class="{ 'mobile-layout': isMobile }">
    <div class="sticky-header">
    <div class="dialog-toolbar">
      <span>字典类型：<el-tag>{{ currentDictType }}</el-tag></span>
      <el-button type="primary" size="small" @click="handleAdd">新增数据</el-button>
    </div>
    </div>

    <div class="scrollable-content">
    <ResponsiveList
      ref="listRef"
      :data="tableData"
      :loading="loading"
      :table-props="{ 'row-key': 'dictCode' }"
      row-key="dictCode"
    >
      <!-- 桌面端表格 -->
      <template #table>
        <el-table-column prop="dictSort" label="排序" width="80" />
        <el-table-column prop="dictLabel" label="数据标签" min-width="140" />
        <el-table-column prop="dictValue" label="数据键值" min-width="140" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 0 ? 'success' : 'danger'" size="small">
              {{ row.status === 0 ? '正常' : '停用' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </template>

      <!-- 手机端卡片 -->
      <template #card="{ row }">
        <div class="card-title">
          {{ row.dictLabel }}
          <el-tag :type="row.status === 0 ? 'success' : 'danger'" size="small" class="card-tag">
            {{ row.status === 0 ? '正常' : '停用' }}
          </el-tag>
        </div>
        <div class="card-row">
          <span class="card-label">数据键值</span>
          <span class="card-value">{{ row.dictValue }}</span>
        </div>
        <div class="card-row">
          <span class="card-label">排序</span>
          <span class="card-value">{{ row.dictSort }}</span>
        </div>
      </template>

      <!-- 手机端操作按钮 -->
      <template #actions="{ row }">
        <el-button size="small" type="primary" @click="handleEdit(row)">编辑</el-button>
        <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
      </template>
    </ResponsiveList>

    <el-pagination
      class="pagination"
      v-model:current-page="queryParams.pageNum"
      v-model:page-size="queryParams.pageSize"
      :total="total"
      :page-sizes="[10, 20, 50]"
      :layout="isMobile ? 'total, prev, pager, next' : 'total, sizes, prev, pager, next, jumper'"
      :small="isMobile"
      @size-change="fetchData"
      @current-change="fetchData"
    />
    </div>
    </div>

    <DictDataForm ref="dataFormRef" @success="fetchData" />
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listDictData, deleteDictData } from '@/api/dict'
import type { DictDataItem } from '@/api/dict'
import DictDataForm from './DictDataForm.vue'
import ResponsiveList from '@/components/ResponsiveList.vue'

// 响应式检测
const isMobile = ref(false)
const MOBILE_BREAKPOINT = 768
const listRef = ref()

function checkMobile() {
  isMobile.value = window.innerWidth <= MOBILE_BREAKPOINT
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})

const visible = ref(false)
const loading = ref(false)
const tableData = ref<DictDataItem[]>([])
const total = ref(0)
const currentDictType = ref('')
const dataFormRef = ref()

const queryParams = reactive({
  dictType: '',
  pageNum: 1,
  pageSize: 10
})

async function fetchData() {
  loading.value = true
  try {
    const { list, total: t } = await listDictData({ ...queryParams, dictType: currentDictType.value })
    tableData.value = list
    total.value = t
  } finally {
    loading.value = false
  }
}

function open(dictType: string) {
  visible.value = true
  currentDictType.value = dictType
  queryParams.pageNum = 1
  fetchData()
}

function handleAdd() {
  dataFormRef.value?.open(undefined, currentDictType.value)
}

function handleEdit(row: DictDataItem) {
  dataFormRef.value?.open(row, currentDictType.value)
}

async function handleDelete(row: DictDataItem) {
  await ElMessageBox.confirm(`确认删除数据「${row.dictLabel}」？`, '提示', { type: 'warning' })
  await deleteDictData(row.dictCode)
  ElMessage.success('删除成功')
  fetchData()
}

defineExpose({ open })
</script>

<style scoped>
.dialog-toolbar {
  margin-bottom: 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.pagination { margin-top: 16px; justify-content: flex-end; }

@media (max-width: 768px) {
  /* 对话框固定头部布局 */
  :deep(.dict-data-dialog .el-dialog__body) {
    padding: 0;
  }

  .dialog-body.mobile-layout {
    display: flex;
    flex-direction: column;
    height: 70vh;
    overflow: hidden;
  }

  .dialog-body.mobile-layout .sticky-header {
    flex-shrink: 0;
    background: var(--el-bg-color);
    z-index: 10;
    padding: 12px;
    border-bottom: 1px solid var(--el-border-color-lighter);
  }

  .dialog-body.mobile-layout .dialog-toolbar {
    margin-bottom: 0;
  }

  .dialog-body.mobile-layout .scrollable-content {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    -webkit-overflow-scrolling: touch;
  }

  .pagination { justify-content: center; }
}
</style>
