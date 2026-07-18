<template>
  <div class="page-container">
    <el-card shadow="never" class="search-card">
      <el-form :model="queryParams" inline>
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
    </el-card>

    <el-card shadow="never" class="table-card">
      <template #header>
        <div class="card-header">
          <span>字典类型列表</span>
          <el-button type="primary" @click="handleAdd">新增字典类型</el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="tableData" row-key="dictId">
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
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleData(row)">数据管理</el-button>
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

    <DictTypeForm ref="formRef" @success="fetchData" />
    <DictDataList ref="dataRef" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listDictTypes, deleteDictType } from '@/api/dict'
import type { DictTypeItem } from '@/api/dict'
import DictTypeForm from './DictTypeForm.vue'
import DictDataList from './DictDataList.vue'

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
.search-card { margin-bottom: 16px; }
.card-header { display: flex; justify-content: space-between; align-items: center; }
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
