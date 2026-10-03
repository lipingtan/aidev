<template>
  <div class="page-container" :class="{ 'mobile-layout': isMobile }">
    <el-card shadow="never" class="main-card">
      <!-- 固定头部区域（手机端） -->
      <div class="sticky-header">
        <div class="card-header">
          <span>菜单列表</span>
          <el-button type="primary" size="small" @click="handleAdd()">
            <el-icon><Plus /></el-icon>
            <span v-if="!isMobile">新增菜单</span>
          </el-button>
        </div>

        <!-- 搜索栏 -->
        <el-form :model="queryParams" :inline="!isMobile" class="search-form">
          <el-form-item label="菜单名称">
            <el-input v-model="queryParams.name" placeholder="请输入菜单名称" clearable @keyup.enter="fetchData" />
          </el-form-item>
          <el-form-item label="状态">
            <el-select v-model="queryParams.status" placeholder="全部" clearable>
              <el-option label="启用" :value="1" />
              <el-option label="禁用" :value="0" />
            </el-select>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="fetchData">查询</el-button>
            <el-button @click="queryParams.name = ''; queryParams.status = undefined; fetchData()">重置</el-button>
          </el-form-item>
        </el-form>
      </div>

      <!-- 可滚动的列表区域 -->
      <div class="scrollable-content">
      <ResponsiveList
        ref="listRef"
        :data="menuList"
        :loading="loading"
        :table-props="{ 'row-key': 'id', 'default-expand-all': true, 'tree-props': { children: 'children' } }"
      >
        <!-- 桌面端表格 -->
        <template #table>
          <el-table-column prop="name" label="菜单名称" min-width="180" />
          <el-table-column prop="icon" label="图标" width="80">
            <template #default="{ row }">{{ row.icon || '-' }}</template>
          </el-table-column>
          <el-table-column prop="type" label="类型" width="80">
            <template #default="{ row }">
              <el-tag :type="row.type === 'MENU' ? 'success' : 'warning'" size="small">
                {{ row.type === 'MENU' ? '菜单' : '按钮' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="sort_order" label="排序" width="80" />
          <el-table-column prop="permission_code" label="权限标识" min-width="160" />
          <el-table-column prop="path" label="路由路径" min-width="160" />
          <el-table-column prop="status" label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="160" fixed="right">
            <template #default="{ row }">
              <el-button v-if="row.type === 'MENU'" link type="primary" @click="handleAdd(row.id)">新增</el-button>
              <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
              <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
            </template>
          </el-table-column>
        </template>

        <!-- 手机端卡片 -->
        <template #card="{ row }">
          <div class="card-title">
            {{ row.name }}
            <el-tag :type="row.type === 'MENU' ? 'success' : 'warning'" size="small" class="card-tag">
              {{ row.type === 'MENU' ? '菜单' : '按钮' }}
            </el-tag>
            <el-tag :type="row.status === 1 ? 'success' : 'danger'" size="small" class="card-tag">
              {{ row.status === 1 ? '启用' : '禁用' }}
            </el-tag>
          </div>
          <div class="card-row">
            <span class="card-label">图标</span>
            <span class="card-value">{{ row.icon || '-' }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">排序</span>
            <span class="card-value">{{ row.sort_order }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">权限标识</span>
            <span class="card-value">{{ row.permission_code || '-' }}</span>
          </div>
          <div class="card-row">
            <span class="card-label">路由路径</span>
            <span class="card-value">{{ row.path || '-' }}</span>
          </div>
          <div v-if="row.children && row.children.length > 0" class="card-row">
            <span class="card-label">子菜单</span>
            <span class="card-value">{{ row.children.length }} 项</span>
          </div>
        </template>

        <!-- 手机端操作按钮 -->
        <template #actions="{ row }">
          <el-button v-if="row.type === 'MENU'" size="small" type="primary" @click="handleAdd(row.id)">新增</el-button>
          <el-button size="small" type="primary" @click="handleEdit(row)">编辑</el-button>
          <el-button size="small" type="danger" @click="handleDelete(row)">删除</el-button>
        </template>
      </ResponsiveList>
      </div>
    </el-card>

    <MenuForm ref="formRef" @success="fetchData" />
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { Plus } from '@element-plus/icons-vue'
import { getMenuTree, deleteMenu } from '@/api/menu'
import type { MenuTreeItem } from '@/api/menu'
import MenuForm from './MenuForm.vue'
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
const menuTree = ref<MenuTreeItem[]>([])
const formRef = ref()
const queryParams = reactive({ name: '', status: undefined as number | undefined })

// 手机端卡片展开树结构为扁平列表（带层级），桌面端保持树形
const menuList = computed<MenuTreeItem[]>(() => {
  if (!isMobile.value) return menuTree.value
  const flat: MenuTreeItem[] = []
  const walk = (items: MenuTreeItem[]) => {
    items.forEach(item => {
      flat.push(item)
      if (item.children && item.children.length > 0) walk(item.children)
    })
  }
  walk(menuTree.value)
  return flat
})

async function fetchData() {
  loading.value = true
  try { menuTree.value = await getMenuTree({ menuName: queryParams.name, status: queryParams.status }) }
  finally { loading.value = false }
}

function handleAdd(parentId?: string) { formRef.value?.open(undefined, parentId || null) }
function handleEdit(row: MenuTreeItem) { formRef.value?.open(row) }

async function handleDelete(row: MenuTreeItem) {
  await ElMessageBox.confirm(`确认删除菜单「${row.name}」？`, '提示', { type: 'warning' })
  await deleteMenu(row.id)
  ElMessage.success('删除成功')
  fetchData()
}

onMounted(fetchData)
</script>

<style scoped>
.page-container { padding: 16px; }
.main-card { }
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
  .search-form :deep(.el-select),
  .search-form :deep(.el-input) {
    width: 100% !important;
  }
}
</style>
