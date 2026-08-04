<!--
  ABAC 条件树可视化编辑器
  支持递归嵌套的 AND/OR 组合节点和叶子表达式节点
-->
<script setup lang="ts">
import CondNode from './CondNode.vue'
import type { CondNode as CondNodeType, ResourceAttrVO, SubjectAttrVO } from '@/api/abac'

const props = defineProps<{
  modelValue: CondNodeType | null
  resourceAttrs: ResourceAttrVO[]
  subjectAttrs: SubjectAttrVO[]
}>()

const emit = defineEmits<{
  'update:modelValue': [node: CondNodeType | null]
}>()

function addRoot() {
  emit('update:modelValue', {
    type: 'expr',
    left: { source: 'resource', attr: '' },
    op: 'eq',
    right: { source: 'const', value: '' }
  })
}

function handleRootUpdate(node: CondNodeType) {
  emit('update:modelValue', node)
}
</script>

<template>
  <div class="condition-editor">
    <CondNode
      v-if="modelValue"
      :node="modelValue"
      :resource-attrs="resourceAttrs"
      :subject-attrs="subjectAttrs"
      :depth="0"
      @update="handleRootUpdate"
      @remove="$emit('update:modelValue', null)"
    />
    <el-button v-if="!modelValue" size="small" type="primary" plain @click="addRoot">
      + 添加条件
    </el-button>
  </div>
</template>

<style scoped>
.condition-editor { padding: 4px 0; }
</style>
