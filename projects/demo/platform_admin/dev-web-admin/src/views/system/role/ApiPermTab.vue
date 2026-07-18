<template>
  <div class="perm-tab">
    <!-- API 权限树 -->
    <el-tree
      ref="treeRef"
      v-loading="loading"
      :data="apiTree"
      show-checkbox
      node-key="id"
      :default-checked-keys="checkedKeys"
      :props="{ label: 'name', children: 'children' }"
      default-expand-all
      @check="handleCheck"
    >
      <template #default="{ data }">
        <span class="tree-node">
          <el-tag :type="data.type === 'GROUP' ? 'primary' : 'success'" size="small">
            {{ data.type === 'GROUP' ? '分组' : 'API' }}
          </el-tag>
          <span class="node-label">{{ data.type === 'GROUP' ? (data.display_name || data.name) : '' }}</span>
          <span v-if="data.type === 'ENDPOINT'" class="node-method">
            <span v-if="data.display_name" class="node-display-name">{{ data.display_name }}</span>
            <el-tag size="small" :type="getMethodTagType(data.http_method)">{{ data.http_method }}</el-tag>
            <span class="node-path">{{ data.url_pattern }}</span>
          </span>
          <el-switch
            v-model="data.visible"
            :active-value="1"
            :inactive-value="0"
            size="small"
            disabled
            style="margin-left: auto"
          />
        </span>
      </template>
    </el-tree>
    <div class="perm-tab-footer">
      <el-button type="primary" :loading="submitting" @click="handleSave">保存</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { ElTree } from 'element-plus'
import { getApiPermTree, getRoleApis, assignRoleApis } from '@/api/role'
import type { ApiPermTreeNode } from '@/api/role'

const props = defineProps<{ roleId: string; appCode: string }>()

const loading = ref(false)
const submitting = ref(false)
const apiTree = ref<ApiPermTreeNode[]>([])
const checkedKeys = ref<string[]>([])
const treeRef = ref<InstanceType<typeof ElTree>>()
const isDirty = ref(false)

defineExpose({ isDirty })

function getMethodTagType(method?: string): 'primary' | 'success' | 'warning' | 'danger' | 'info' {
  switch (method) {
    case 'GET': return 'success'
    case 'POST': return 'primary'
    case 'PUT': return 'warning'
    case 'DELETE': return 'danger'
    default: return 'info'
  }
}

async function loadData() {
  if (!props.appCode) return
  loading.value = true
  try {
    const [tree, ids] = await Promise.all([
      getApiPermTree(props.appCode),
      getRoleApis(props.roleId)
    ])
    // 过滤：显示 visible=1 的节点 + 已绑定(在 ids 中)的 visible=0 节点
    apiTree.value = filterVisibleTree(tree, ids)
    // 勾选叶子节点
    checkedKeys.value = filterLeafIds(apiTree.value, ids)
  } finally {
    loading.value = false
  }
}

// 过滤树：保留 visible=1 的 + 已绑定的 visible=0 的
function filterVisibleTree(nodes: ApiPermTreeNode[], boundIds: string[]): ApiPermTreeNode[] {
  return nodes
    .filter(n => n.visible === 1 || n.visible === undefined || boundIds.includes(n.id) || hasVisibleChildren(n, boundIds))
    .map(n => ({
      ...n,
      children: n.children ? filterVisibleTree(n.children, boundIds) : undefined
    }))
}

function hasVisibleChildren(node: ApiPermTreeNode, boundIds: string[]): boolean {
  if (!node.children) return false
  return node.children.some(c =>
    c.visible === 1 || c.visible === undefined || boundIds.includes(c.id) || hasVisibleChildren(c, boundIds)
  )
}

function filterLeafIds(nodes: ApiPermTreeNode[], ids: string[]): string[] {
  const leafIds: string[] = []
  function walk(list: ApiPermTreeNode[]) {
    for (const n of list) {
      if (!n.children || n.children.length === 0) {
        if (ids.includes(n.id)) leafIds.push(n.id)
      } else {
        walk(n.children)
      }
    }
  }
  walk(nodes)
  return leafIds
}

function handleCheck() {
  isDirty.value = true
}

async function handleSave() {
  submitting.value = true
  try {
    const checked = treeRef.value?.getCheckedKeys(false) as string[]
    const half = treeRef.value?.getHalfCheckedKeys() as string[]
    await assignRoleApis(props.roleId, [...checked, ...half])
    isDirty.value = false
    ElMessage.success('接口权限保存成功')
  } finally {
    submitting.value = false
  }
}

watch(() => props.appCode, () => {
  isDirty.value = false
  loadData()
}, { immediate: true })

watch(() => props.roleId, () => {
  isDirty.value = false
  if (props.appCode) loadData()
})
</script>

<style scoped>
.perm-tab { min-height: 200px; }
.perm-tab-footer { margin-top: 16px; text-align: right; }
.tree-node { display: flex; align-items: center; flex: 1; font-size: 13px; gap: 6px; }
.node-display-name { font-weight: 500; margin-right: 4px; }
.node-path { color: #909399; font-size: 12px; }
.node-label { margin-right: 4px; }
.node-method { display: inline-flex; align-items: center; gap: 4px; }
</style>
