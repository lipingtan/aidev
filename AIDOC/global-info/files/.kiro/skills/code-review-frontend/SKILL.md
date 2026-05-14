---
skillName: code-review-frontend
description: Vue 前端代码审查技能，检查组件规范、性能和安全性
version: 1.0.0
---

# Code Review 技能（前端版）

## 使用场景

在以下情况触发此 skill：

- **提交 PR 前**：开发完成后，提交 Pull Request 之前进行自查或请 AI 审查
- **完成任务后**：完成某个开发任务节点，进行阶段性代码质量检查
- **组件重构时**：对已有 Vue 组件进行重构，确保重构后质量不下降
- **新人代码审查**：对新加入团队成员的代码进行 Vue/TypeScript 规范性检查

## 核心审查规则

### Vue 组件规范

- 组件文件名使用 `PascalCase`，如 `UserProfile.vue`、`OrderList.vue`
- 单文件组件（SFC）结构顺序：`<script setup>` → `<template>` → `<style scoped>`
- 使用 Composition API（`<script setup>`），不使用 Options API（新代码）
- 组件职责单一，超过 300 行应考虑拆分为子组件
- 避免在 `<template>` 中写复杂逻辑，应提取为 computed 或方法

### Props / Events 规范

- Props 必须定义类型，使用 TypeScript 接口或 `defineProps<T>()` 泛型语法
- 必填 Props 不设默认值，可选 Props 必须有合理默认值
- Props 命名使用 `camelCase`，在模板中使用 `kebab-case`
- 自定义事件使用 `defineEmits` 明确声明，事件名使用 `camelCase`
- 不通过 `$parent` 或直接修改 Props 进行父子通信，使用 emit 或 provide/inject

### Composables 规范

- Composable 函数以 `use` 前缀命名，如 `useUserInfo`、`useTableData`
- 每个 Composable 职责单一，只处理一类逻辑
- Composable 内的响应式数据应在函数内部创建，避免共享状态导致的副作用
- 异步 Composable 需处理 loading、error 状态
- 在 `onUnmounted` 中清理副作用（事件监听、定时器、WebSocket 等）

### 安全规范

- 禁止使用 `v-html` 渲染用户输入内容，防止 XSS 攻击
- 动态路由参数需做类型校验和边界处理
- 敏感信息（Token、密码）不存储在 `localStorage`，使用 `sessionStorage` 或内存
- API 请求统一通过封装的 axios 实例发送，不直接使用 `fetch` 或裸 axios
- 表单提交前必须做前端校验，不依赖后端单独校验

## 审查清单

### 组件质量

- [ ] 组件文件名使用 PascalCase
- [ ] 使用 `<script setup>` + Composition API
- [ ] 组件行数不超过 300 行，超过则拆分
- [ ] 模板中无复杂逻辑表达式，已提取为 computed/方法
- [ ] 无注释掉的废弃代码块

### TypeScript 规范

- [ ] 无 `any` 类型（特殊情况需注释说明原因）
- [ ] Props 已定义完整类型
- [ ] API 响应数据已定义接口类型
- [ ] 函数参数和返回值有明确类型注解

### 性能

- [ ] 列表渲染使用唯一且稳定的 `:key`，不使用数组 index
- [ ] 大列表使用虚拟滚动（如 `vue-virtual-scroller`）
- [ ] 图片资源有懒加载处理
- [ ] 无不必要的 `watch`，优先使用 `computed`
- [ ] 组件按需引入，无全量引入 UI 库的情况

### 安全性

- [ ] 无 `v-html` 渲染用户输入
- [ ] 敏感信息未存储在 localStorage
- [ ] 表单有前端校验逻辑
- [ ] API 请求通过统一封装的实例发送

### 规范符合度

- [ ] Props/Emits 已明确声明类型
- [ ] Composable 以 `use` 前缀命名
- [ ] 副作用在 `onUnmounted` 中清理
- [ ] 代码通过 ESLint 检查无报错

## 示例

### 好的组件示例

```vue
<script setup lang="ts">
import { ref, computed, onUnmounted } from 'vue'
import { useUserStore } from '@/stores/user'
import type { UserInfo } from '@/types/user'

interface Props {
  userId: number
  showAvatar?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  showAvatar: true,
})

const emit = defineEmits<{
  (e: 'update', user: UserInfo): void
  (e: 'delete', userId: number): void
}>()

const userStore = useUserStore()
const loading = ref(false)

// ✅ 复杂逻辑提取为 computed
const displayName = computed(() =>
  userStore.currentUser?.nickname || userStore.currentUser?.username || '未知用户'
)

// ✅ 清理副作用
const timer = setInterval(() => {
  // 定时刷新逻辑
}, 30000)

onUnmounted(() => {
  clearInterval(timer)
})
</script>

<template>
  <div class="user-card">
    <!-- ✅ 使用唯一稳定的 key -->
    <el-avatar v-if="showAvatar" :src="userStore.currentUser?.avatar" />
    <span>{{ displayName }}</span>
  </div>
</template>

<style scoped>
.user-card {
  display: flex;
  align-items: center;
  gap: 8px;
}
</style>
```

### 需要改进的组件示例

```vue
<!-- ❌ 问题：Options API、无类型、v-html 风险、index 作为 key、无副作用清理 -->
<script>
export default {
  data() {
    return {
      users: [],
      desc: ''  // ❌ 无类型
    }
  },
  mounted() {
    this.timer = setInterval(this.refresh, 1000)
    // ❌ 未在 beforeUnmount 中清理 timer
  },
  methods: {
    refresh() { /* ... */ }
  }
}
</script>

<template>
  <div>
    <!-- ❌ 使用 index 作为 key，列表重排时会有问题 -->
    <div v-for="(user, index) in users" :key="index">
      <!-- ❌ v-html 渲染用户输入，XSS 风险 -->
      <span v-html="user.desc"></span>
    </div>
  </div>
</template>

<!-- ✅ 改进后 -->
<script setup lang="ts">
import { ref, onUnmounted } from 'vue'
import type { User } from '@/types'

const users = ref<User[]>([])

const timer = setInterval(refresh, 1000)
onUnmounted(() => clearInterval(timer))  // ✅ 清理副作用

function refresh() { /* ... */ }
</script>

<template>
  <div>
    <!-- ✅ 使用唯一 ID 作为 key -->
    <div v-for="user in users" :key="user.id">
      <!-- ✅ 使用文本插值，安全 -->
      <span>{{ user.desc }}</span>
    </div>
  </div>
</template>
```
