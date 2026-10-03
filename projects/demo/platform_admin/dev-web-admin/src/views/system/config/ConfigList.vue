<template>
  <div class="page-container" :class="{ 'mobile-layout': isMobile }">
    <el-card shadow="never" class="main-card">
      <!-- 固定头部区域（手机端） -->
      <div class="sticky-header">
        <div class="card-header">
          <span>配置列表</span>
          <el-button type="primary" size="small" @click="handleAdd">
            <span v-if="!isMobile">新增配置</span>
            <span v-else>新增</span>
          </el-button>
        </div>

        <!-- 搜索栏 -->
        <el-form :model="queryParams" :inline="!isMobile" class="search-form">
          <el-form-item label="配置名称">
            <el-input v-model="queryParams.config_name" placeholder="请输入配置名称" clearable @keyup.enter="handleSearch" />
          </el-form-item>
          <el-form-item label="配置键">
            <el-input v-model="queryParams.config_key" placeholder="请输入配置键" clearable @keyup.enter="handleSearch" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="handleSearch">查询</el-button>
            <el-button @click="handleReset">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 可滚动的列表区域 -->
      <div class="scrollable-content">
      <!-- 响应式列表 -->
      <ResponsiveList :data="tableData" :loading="loading">
        <!-- 桌面端表格 -->
        <template #table>
          <el-table-column prop="config_name" label="配置名称" min-width="160" />
          <el-table-column prop="config_key" label="配置键" min-width="200" />
          <el-table-column prop="config_value" label="配置值" min-width="200" show-overflow-tooltip />
          <el-table-column prop="config_type" label="是否内置" width="100">
            <template #default="{ row }">
              <el-tag :type="row.config_type === 1 ? 'primary' : 'info'" size="small">
                {{ row.config_type === 1 ? '是' : '否' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="created_at" label="创建时间" width="180" />
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
            {{ row.config_name }}
            <el-tag :type="row.config_type === 1 ? 'primary' : 'info'" size="small" class="card-tag">
              {{ row.config_type === 1 ? '内置' : '自定义' }}
            </el-tag>
          </div>
          <div class="card-row">
            <span class="card-label">配置键</span>
            <span class="card-value">{{ row.config_key }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">配置值</span>
            <span class="card-value">{{ row.config_value }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">创建时间</span>
            <span class="card-value">{{ row.created_at }}</span>
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
        v-model:current-page="queryParams.page"
        v-model:page-size="queryParams.page_size"
        :total="total"
        :page-sizes="[10, 20, 50]"
        :layout="isMobile ? 'total, prev, pager, next' : 'total, sizes, prev, pager, next, jumper'"
        :small="isMobile"
        @size-change="fetchData"
        @current-change="fetchData"
      />
      </div>
    </el-card>

    <ConfigForm ref="formRef" @success="fetchData" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listConfigs, deleteConfig } from '@/api/config'
import type { ConfigItem } from '@/api/config'
import ResponsiveList from '@/components/ResponsiveList.vue'
import ConfigForm from './ConfigForm.vue'

// 响应式检测
const isMobile = ref(false)
const MOBILE_BREAKPOINT = 768

function checkMobile() {
  isMobile.value = window.innerWidth <= MOBILE_BREAKPOINT
}

const loading = ref(false)
const tableData = ref<ConfigItem[]>([])
const total = ref(0)
const formRef = ref()

const queryParams = reactive({
  config_name: '',
  config_key: '',
  page: 1,
  page_size: 10
})

async function fetchData() {
  loading.value = true
  try {
    const { list, total: t } = await listConfigs(queryParams)
    tableData.value = list
    total.value = t
  } finally {
    loading.value = false
  }
}

function handleSearch() {
  queryParams.page = 1
  fetchData()
}

function handleReset() {
  queryParams.config_name = ''
  queryParams.config_key = ''
  handleSearch()
}

function handleAdd() {
  formRef.value?.open()
}

function handleEdit(row: ConfigItem) {
  formRef.value?.open(row)
}

async function handleDelete(row: ConfigItem) {
  await ElMessageBox.confirm(`确认删除配置「${row.config_name}」？`, '提示', { type: 'warning' })
  await deleteConfig(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(() => {
  checkMobile()
  window.addEventListener('resize', checkMobile)
  fetchData()
})

onUnmounted(() => {
  window.removeEventListener('resize', checkMobile)
})
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
