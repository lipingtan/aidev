<template>
  <div class="page-container">
    <el-card shadow="never" class="search-card">
      <el-form :model="queryParams" inline>
        <el-form-item label="配置名称">
          <el-input v-model="queryParams.configName" placeholder="请输入配置名称" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="配置键">
          <el-input v-model="queryParams.configKey" placeholder="请输入配置键" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="handleSearch">查询</el-button>
          <el-button @click="handleReset">重置</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card shadow="never" class="table-card">
      <template #header>
        <div class="card-header">
          <span>配置列表</span>
          <el-button type="primary" @click="handleAdd">新增配置</el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="tableData" row-key="configId">
        <el-table-column prop="configName" label="配置名称" min-width="160" />
        <el-table-column prop="configKey" label="配置键" min-width="200" />
        <el-table-column prop="configValue" label="配置值" min-width="200" />
        <el-table-column prop="configType" label="是否内置" width="100">
          <template #default="{ row }">
            <el-tag :type="row.configType === 1 ? '' : 'info'" size="small">
              {{ row.configType === 1 ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="createTime" label="创建时间" width="180" />
        <el-table-column label="操作" width="160" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        class="pagination"
        v-model:current-page="queryParams.pageNum"
        v-model:page-size="queryParams.pageSize"
        :total="total"
        :page-sizes="[10, 20, 50]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchData"
        @current-change="fetchData"
      />
    </el-card>

    <ConfigForm ref="formRef" @success="fetchData" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listConfigs, deleteConfig } from '@/api/config'
import type { ConfigItem } from '@/api/config'
import ConfigForm from './ConfigForm.vue'

const loading = ref(false)
const tableData = ref<ConfigItem[]>([])
const total = ref(0)
const formRef = ref()

const queryParams = reactive({
  configName: '',
  configKey: '',
  pageNum: 1,
  pageSize: 10
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
  queryParams.pageNum = 1
  fetchData()
}

function handleReset() {
  queryParams.configName = ''
  queryParams.configKey = ''
  handleSearch()
}

function handleAdd() {
  formRef.value?.open()
}

function handleEdit(row: ConfigItem) {
  formRef.value?.open(row)
}

async function handleDelete(row: ConfigItem) {
  await ElMessageBox.confirm(`确认删除配置「${row.configName}」？`, '提示', { type: 'warning' })
  await deleteConfig(row.configId)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(fetchData)
</script>

<style scoped>
.page-container { padding: 16px; }
.search-card { margin-bottom: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
