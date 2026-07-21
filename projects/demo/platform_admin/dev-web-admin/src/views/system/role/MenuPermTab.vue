<template>
  <div class="perm-tab">
    <!-- 资源树 -->
    <el-tree
      ref="treeRef"
      v-loading="loading"
      :data="resourceTree"
      show-checkbox
      node-key="id"
      :default-checked-keys="checkedKeys"
      :props="{ label: 'name', children: 'children', disabled: disabledFn }"
      default-expand-all
      @check="() => { isDirty = true }"
    />
    <div class="perm-tab-footer">
      <el-button type="primary" :loading="submitting" @click="handleSave">保存</el-button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { ElTree } from 'element-plus'
import { getResourceTree, getRoleResources, assignRoleResources, getAssignableResources } from '@/api/role'
import type { ResourceTreeNode, AssignResult } from '@/api/role'

const props = defineProps<{ roleId: string; appCode: string; parentId?: string | null }>()

const loading = ref(false)
const submitting = ref(false)
const resourceTree = ref<ResourceTreeNode[]>([])
const checkedKeys = ref<string[]>([])
const treeRef = ref<InstanceType<typeof ElTree>>()
const isDirty = ref(false)
const assignableIds = ref<string[] | null>(null)

defineExpose({ isDirty })

/** 判断节点是否不可选（不在父角色可分配范围内） */
function disabledFn(data: ResourceTreeNode): boolean {
  if (!assignableIds.value) return false
  return !assignableIds.value.includes(data.id)
}

async function loadData() {
  if (!props.appCode) return
  loading.value = true
  try {
    const requests: Promise<any>[] = [
      getResourceTree(props.appCode),
      getRoleResources(props.roleId)
    ]
    // 如果有 parentId，获取可分配范围
    if (props.parentId) {
      requests.push(getAssignableResources(props.parentId))
    }
    const results = await Promise.all(requests)
    const tree: ResourceTreeNode[] = results[0]
    const ids: string[] = results[1]
    assignableIds.value = results[2] || null

    resourceTree.value = tree
    // 只勾选叶子节点避免父节点自动全选
    checkedKeys.value = filterLeafIds(tree, ids)
  } finally {
    loading.value = false
  }
}

function filterLeafIds(nodes: ResourceTreeNode[], ids: string[]): string[] {
  const leafIds: string[] = []
  function walk(list: ResourceTreeNode[]) {
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

async function handleSave() {
  submitting.value = true
  try {
    const checked = treeRef.value?.getCheckedKeys(false) as string[]
    const half = treeRef.value?.getHalfCheckedKeys() as string[]
    const result: AssignResult = await assignRoleResources(props.roleId, [...checked, ...half])
    isDirty.value = false
    // 显示级联裁剪影响
    if (result?.affected_children?.length) {
      const totalRemoved = result.affected_children.reduce((sum, c) => sum + c.removed_count, 0)
      ElMessage.warning(`已影响 ${result.affected_children.length} 个子角色，共移除 ${totalRemoved} 项权限`)
    } else {
      ElMessage.success('菜单权限保存成功')
    }
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
</style>
