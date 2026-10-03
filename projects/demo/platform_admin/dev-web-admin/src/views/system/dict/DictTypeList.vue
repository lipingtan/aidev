<template>
  <div class="page-container" :class="{ 'mobile-layout': isMobile }">
    <el-card shadow="never" class="main-card">
      <!-- 固定头部区域（手机端） -->
      <div class="sticky-header">
        <div class="card-header">
          <span>字典类型列表</span>
          <el-button type="primary" size="small" @click="handleAdd">
            <el-icon><Plus /></el-icon>
            <span v-if="!isMobile">新增字典类型</span>
          </el-button>
        </div>

        <!-- 搜索栏 -->
        <el-form :model="queryParams" :inline="!isMobile" class="search-form">
          <el-form-item label="字典名称">
            <el-input v-model="queryParams.dictName" placeholder="请输入字典名称" clearable @keyup.enter="handleSearch" />
          </el-form-item>
          <el-form-item label="字典类型">
            <el-input v-model="queryParams.dictType" placeholder="请输入字典类型" clearable @keyup.enter="handleSearch" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 可滚动的列表区域 -->
      <div class="scrollable-content">
      <ResponsiveList
        ref="listRef"
        :data="tableData"
        :loading="loading"
        :table-props="{ 'row-key': 'dictId' }"
        row-key="dictId"
      >
        <!-- 桌面端表格 -->
        <template #table>
          <el-table-column prop="dictName" label="字典名称" min-width="160" />
          <el-table-column prop="dictType" label="字典类型" min-width="160" />
          <el-table-column prop="status" label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.status === 0 ? 'success' : 'danger'" size="small">
                {{ row.status === 0 ? '正常' : '停用' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="remark" label="备注" min-width="160" />
          <el-table-column prop="createTime" label="创建时间" width="180" />
          <el-table-column label="操作" width="180" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="handleData(row)">数据</el-button>
              <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
              <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </template>

        <!-- 手机端卡片 -->
        <template #card="{ row }">
          <div class="card-title">
            {{ row.dictName }}
            <el-tag :type="row.status === 0 ? 'success' : 'danger'" size="small" class="card-tag">
              {{ row.status === 0 ? '正常' : '停用' }}
            </el-tag>
          </div>
          <div class="card-row">
            <span class="card-label">字典类型</span>
            <span class="card-value">{{ row.dictType }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">备注</span>
            <span class="card-value">{{ row.remark || '-' }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">创建时间</span>
            <span class="card-value">{{ row.createTime }}</span>
          </div>
        </template>

        <!-- 手机端操作按钮 -->
        <template #actions="{ row }">
          <el-button size="small" type="primary" @click="handleData(row)">数据</el-button>
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
    </el-card>

    <DictTypeForm ref="formRef" @success="fetchData" />
    <DictDataList ref="dataRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { listDictTypes, deleteDictType } from '@/api/dict'
import type { DictTypeItem } from '@/api/dict'
import DictTypeForm from './DictTypeForm.vue'
import DictDataList from './DictDataList.vue'
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

const loading = ref(false)
const tableData = ref<DictTypeItem[]>([])
const total = ref(0)
const formRef = ref()
const dataRef = ref()

const queryParams = reactive({
  dictName: '',
  dictType: '',
  pageNum: 1,
  pageSize: 10
})

async function fetchData() {
  loading.value = true
  try {
    const { list, total: t } = await listDictTypes(queryParams)
    tableData.value = list
    total.value = t
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  queryParams.pageNum = 1
  fetchData()
}

function handleReset() {
  queryParams.dictName = ''
  queryParams.dictType = ''
  handleSearch()
}

function handleAdd() {
  formRef.value?.open()
}

function handleEdit(row: DictTypeItem) {
  formRef.value?.open(row)
}

function handleData(row: DictTypeItem) {
  dataRef.value?.open(row.dictType)
}

async function handleDelete(row: DictTypeItem) {
  await ElMessageBox.confirm(`确认删除字典类型「${row.dictName}」？`, '提示', { type: 'warning' })
  await deleteDictType(row.dictId)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(fetchData)
</script>

<style scoped>
.page-container { padding: 16px; }
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--el-border-color-light);
}
.search-form {
  padding: 12px 16px;
  margin-bottom: 0;
  border-bottom: 1px solid var(--el-border-color-lighter);
}
.scrollable-content { padding: 16px; }
.pagination { margin-top: 16px; justify-content: flex-end; }

/* 手机端布局 - 固定头部 */
@media (max-width: 768px) {
  .page-container.mobile-layout {
    padding: 0;
    height: 100vh;
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }

  .page-container.mobile-layout .main-card {
    flex: 1;
    display: flex;
    flex-direction: column;
    border-radius: 0;
    overflow: hidden;
  }

  .page-container.mobile-layout .main-card :deep(.el-card__body) {
    flex: 1;
    display: flex;
    flex-direction: column;
    padding: 0;
    overflow: hidden;
  }

  .sticky-header {
    flex-shrink: 0;
    background: var(--el-bg-color);
    z-index: 10;
  }

  .scrollable-content {
    flex: 1;
    overflow-y: auto;
    padding: 12px;
    -webkit-overflow-scrolling: touch;
  }

  .search-form {
    padding: 12px;
  }

  .search-form :deep(.el-form-item) {
    margin-bottom: 12px;
    margin-right: 0;
    width: 100%;
  }
  .search-form :deep(.el-input) {
    width: 100% !important;
  }
  .pagination { justify-content: center; }
}
</style>
