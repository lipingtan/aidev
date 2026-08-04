<!--
  ABAC 条件节点（递归组件）
  type=group → AND/OR 逻辑组合，可添加子节点
  type=expr  → 叶子表达式：左值(资源属性) op 右值(主体属性/资源属性/常量)
-->
<script setup lang="ts">
import type { CondNode as CondNodeType, ResourceAttrVO, SubjectAttrVO } from '@/api/abac'

const props = defineProps<{
  node: CondNodeType
  resourceAttrs: ResourceAttrVO[]
  subjectAttrs: SubjectAttrVO[]
  depth: number
}>()

const emit = defineEmits<{
  update: [node: CondNodeType]
  remove: []
}>()

const ops = [
  { label: '等于', value: 'eq' },
  { label: '不等于', value: 'ne' },
  { label: '大于', value: 'gt' },
  { label: '小于', value: 'lt' },
  { label: '大于等于', value: 'gte' },
  { label: '小于等于', value: 'lte' },
  { label: '包含于', value: 'in' },
  { label: '不包含于', value: 'not_in' },
  { label: '包含', value: 'contains' },
  { label: '为空', value: 'is_null' },
  { label: '不为空', value: 'is_not_null' },
]

const noRightOps = ['is_null', 'is_not_null']

function update(patch: Partial<CondNodeType>) {
  emit('update', { ...props.node, ...patch })
}

function addChild(type: 'expr' | 'group') {
  const children = [...(props.node.children || [])]
  if (type === 'expr') {
    children.push({
      type: 'expr',
      left: { source: 'resource', attr: '' },
      op: 'eq',
      right: { source: 'const', value: '' }
    })
  } else {
    children.push({ type: 'group', operator: 'AND', children: [] })
  }
  update({ children })
}

function updateChild(index: number, newNode: CondNodeType) {
  const children = [...(props.node.children || [])]
  children[index] = newNode
  update({ children })
}

function removeChild(index: number) {
  const children = [...(props.node.children || [])]
  children.splice(index, 1)
  update({ children })
}

function updateLeftAttr(attr: string) {
  update({ left: { source: 'resource', attr } })
}

function updateOp(op: string) {
  update({ op })
}

function updateRightSource(source: string) {
  update({ right: { source: source as any, attr: '', value: '' } })
}

function updateRightAttr(attr: string) {
  update({ right: { ...props.node.right, source: props.node.right?.source as any, attr } })
}

function updateRightValue(value: string) {
  update({ right: { source: 'const', value } })
}
</script>

<template>
  <!-- 组合节点 -->
  <div v-if="node.type === 'group'" class="cond-group" :style="{ marginLeft: depth * 16 + 'px' }">
    <div class="group-header">
      <span class="group-label">逻辑：</span>
      <el-select
        :model-value="node.operator"
        size="small"
        style="width: 100px"
        @update:model-value="(v) => update({ operator: v as 'AND' | 'OR' })"
      >
        <el-option label="AND（且）" value="AND" />
        <el-option label="OR（或）" value="OR" />
      </el-select>
      <el-button size="small" link type="primary" @click="addChild('expr')">+ 条件</el-button>
      <el-button size="small" link type="primary" @click="addChild('group')">+ 组合</el-button>
      <el-button size="small" link type="danger" @click="$emit('remove')">删除</el-button>
    </div>
    <CondNode
      v-for="(child, i) in node.children"
      :key="i"
      :node="child"
      :resource-attrs="resourceAttrs"
      :subject-attrs="subjectAttrs"
      :depth="depth + 1"
      @update="(n) => updateChild(i, n)"
      @remove="removeChild(i)"
    />
  </div>

  <!-- 叶子表达式节点 -->
  <div v-else class="cond-expr" :style="{ marginLeft: depth * 16 + 'px' }">
    <!-- 左值：资源属性 -->
    <el-select
      :model-value="node.left?.attr"
      size="small"
      placeholder="选择资源属性"
      style="width: 160px"
      @update:model-value="updateLeftAttr"
    >
      <el-option
        v-for="a in resourceAttrs"
        :key="a.attr_name"
        :label="a.display"
        :value="a.attr_name"
      />
    </el-select>

    <!-- 操作符 -->
    <el-select
      :model-value="node.op"
      size="small"
      style="width: 110px"
      @update:model-value="updateOp"
    >
      <el-option v-for="o in ops" :key="o.value" :label="o.label" :value="o.value" />
    </el-select>

    <!-- 右值 source 选择（is_null/is_not_null 不需要右值） -->
    <template v-if="!noRightOps.includes(node.op || '')">
      <el-select
        :model-value="node.right?.source || 'const'"
        size="small"
        style="width: 100px"
        @update:model-value="updateRightSource"
      >
        <el-option label="常量" value="const" />
        <el-option label="主体属性" value="subject" />
        <el-option label="资源属性" value="resource" />
      </el-select>

      <!-- 主体属性右值 -->
      <el-select
        v-if="node.right?.source === 'subject'"
        :model-value="node.right?.attr"
        size="small"
        placeholder="选择主体属性"
        style="width: 140px"
        @update:model-value="updateRightAttr"
      >
        <el-option
          v-for="a in subjectAttrs"
          :key="a.attr_name"
          :label="a.display"
          :value="a.attr_name"
        />
      </el-select>

      <!-- 资源属性右值（两列互比） -->
      <el-select
        v-else-if="node.right?.source === 'resource'"
        :model-value="node.right?.attr"
        size="small"
        placeholder="选择资源属性"
        style="width: 140px"
        @update:model-value="updateRightAttr"
      >
        <el-option
          v-for="a in resourceAttrs"
          :key="a.attr_name"
          :label="a.display"
          :value="a.attr_name"
        />
      </el-select>

      <!-- 常量右值 -->
      <el-input
        v-else
        :model-value="String(node.right?.value ?? '')"
        size="small"
        placeholder="输入值"
        style="width: 140px"
        @update:model-value="updateRightValue"
      />
    </template>

    <el-button size="small" link type="danger" @click="$emit('remove')">×</el-button>
  </div>
</template>

<style scoped>
.cond-group {
  border: 1px dashed #dcdfe6;
  border-radius: 6px;
  padding: 8px 10px;
  margin-bottom: 8px;
  background: #fafafa;
}
.group-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}
.group-label { font-size: 13px; color: #606266; }
.cond-expr {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 8px;
  flex-wrap: wrap;
}
</style>
