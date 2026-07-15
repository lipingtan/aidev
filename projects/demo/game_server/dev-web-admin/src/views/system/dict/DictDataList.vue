<template>
  <el-dialog v-model="visible" title="字典数据管理" width="800px" destroy-on-close>
    <div style="margin-bottom: 16px; display: flex; justify-content: space-between; align-items: center;">
      <span>字典类型：<el-tag>{{ currentDictType }}</el-tag></span>
      <el-button type="primary" size="small" @click="handleAdd">新增数据</el-button>
    </div>

    <el-table v-loading="loading" :data="tableData" row-key="dictCode">
      <el-table-column prop="dictSort" label="排序" width="80" />
      <el-table-column prop="dictLabel" label="数据标签" min-width="140" />
      <el-table-column prop="dictValue" label="数据键值" min-width="140" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 0 ? '' : 'danger'" size="small">
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

    <DictDataForm ref="dataFormRef" @success="fetchData" />
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listDictData, deleteDictData } from '@/api/dict'
import type { DictDataItem } from '@/api/dict'
import DictDataForm from './DictDataForm.vue'

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
.pagination { margin-top: 16px; justify-content: flex-end; }
</style>
