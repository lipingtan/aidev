<template>
  <div class="page-container">
    <el-card shadow="never" class="search-card">
      <el-form :model="queryParams" inline>
        <el-form-item label="岗位名称">
          <el-input v-model="queryParams.postName" placeholder="请输入岗位名称" clearable @keyup.enter="handleSearch" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="queryParams.status" placeholder="全部" clearable>
            <el-option label="启用" :value="0" />
            <el-option label="禁用" :value="1" />
          </el-select>
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
          <span>岗位列表</span>
          <el-button type="primary" @click="handleAdd">新增岗位</el-button>
        </div>
      </template>

      <el-table v-loading="loading" :data="tableData" row-key="postId">
        <el-table-column prop="postCode" label="岗位编码" width="140" />
        <el-table-column prop="postName" label="岗位名称" min-width="160" />
        <el-table-column prop="sort" label="排序" width="80" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="row.status === 0 ? '' : 'danger'" size="small">
              {{ row.status === 0 ? '启用' : '禁用' }}
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

    <PostForm ref="formRef" @success="fetchData" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { listPosts, deletePost } from '@/api/post'
import type { PostItem } from '@/api/post'
import PostForm from './PostForm.vue'

const loading = ref(false)
const tableData = ref<PostItem[]>([])
const total = ref(0)
const formRef = ref()

const queryParams = reactive({
  postName: '',
  status: undefined as number | undefined,
  pageNum: 1,
  pageSize: 10
})

async function fetchData() {
  loading.value = true
  try {
    const { list, total: t } = await listPosts(queryParams)
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
  queryParams.postName = ''
  queryParams.status = undefined
  handleSearch()
}

function handleAdd() {
  formRef.value?.open()
}

function handleEdit(row: PostItem) {
  formRef.value?.open(row)
}

async function handleDelete(row: PostItem) {
  await ElMessageBox.confirm(`确认删除岗位「${row.postName}」？`, '提示', { type: 'warning' })
  await deletePost(row.postId)
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
