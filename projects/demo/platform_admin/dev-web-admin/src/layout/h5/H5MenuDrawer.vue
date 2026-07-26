<!--
  H5 侧滑菜单
  使用 Vant Popup 从左侧滑出，显示树形菜单结构
-->
<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter } from 'vue-router'
import type { MenuItem } from '@/store/modules/menu'

interface Props {
  menus: MenuItem[]
}

const props = defineProps<Props>()
const emit = defineEmits<{
  close: []
}>()

const router = useRouter()
const activeKeys = ref<number[]>([])

// 展开的菜单项
const expandedKeys = computed({
  get: () => activeKeys.value,
  set: (val) => { activeKeys.value = val }
})

// 菜单点击处理
const handleMenuClick = (menu: MenuItem) => {
  if (menu.children && menu.children.length > 0) {
    // 有子菜单，切换展开/收起
    const index = activeKeys.value.indexOf(menu.id)
    if (index > -1) {
      activeKeys.value.splice(index, 1)
    } else {
      activeKeys.value.push(menu.id)
    }
  } else if (menu.path) {
    // 叶子菜单，跳转路由
    router.push(menu.path)
    emit('close')
  }
}

// 判断菜单项是否展开
const isExpanded = (menuId: number) => {
  return activeKeys.value.includes(menuId)
}
</script>

<template>
  <div class="h5-menu-drawer">
    <!-- 头部 -->
    <div class="drawer-header">
      <div class="header-title">菜单</div>
      <van-icon name="cross" size="20" @click="emit('close')" />
    </div>

    <!-- 菜单列表 -->
    <div class="drawer-body">
      <template v-for="menu in menus" :key="menu.id">
        <!-- 一级菜单 -->
        <div 
          class="menu-item level-1" 
          :class="{ 'has-children': menu.children && menu.children.length > 0 }"
          @click="handleMenuClick(menu)"
        >
          <div class="menu-item-content">
            <van-icon v-if="menu.icon" :name="menu.icon" class="menu-icon" />
            <span class="menu-title">{{ menu.name }}</span>
          </div>
          <van-icon 
            v-if="menu.children && menu.children.length > 0"
            :name="isExpanded(menu.id) ? 'arrow-down' : 'arrow'"
            class="expand-icon"
          />
        </div>

        <!-- 二级菜单（展开时显示） -->
        <div v-if="isExpanded(menu.id) && menu.children" class="submenu-wrapper">
          <div
            v-for="submenu in menu.children"
            :key="submenu.id"
            class="menu-item level-2"
            @click.stop="handleMenuClick(submenu)"
          >
            <div class="menu-item-content">
              <van-icon v-if="submenu.icon" :name="submenu.icon" class="menu-icon" />
              <span class="menu-title">{{ submenu.name }}</span>
            </div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.h5-menu-drawer {
  display: flex;
  flex-direction: column;
  height: 100%;
  background-color: #fff;
}

.drawer-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid #eee;
}

.header-title {
  font-size: 18px;
  font-weight: 600;
  color: #333;
}

.drawer-body {
  flex: 1;
  overflow-y: auto;
  padding: 8px 0;
  -webkit-overflow-scrolling: touch;
}

.menu-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  cursor: pointer;
  transition: background-color 0.2s;
}

.menu-item:active {
  background-color: #f5f5f5;
}

.menu-item.level-1 {
  font-size: 15px;
  font-weight: 500;
  color: #333;
}

.menu-item.level-2 {
  font-size: 14px;
  color: #666;
  padding-left: 40px;
  background-color: #fafafa;
}

.menu-item-content {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.menu-icon {
  font-size: 18px;
  color: #999;
}

.menu-title {
  flex: 1;
}

.expand-icon {
  color: #999;
  font-size: 14px;
  transition: transform 0.2s;
}

.submenu-wrapper {
  border-left: 2px solid #e8e8e8;
  margin-left: 16px;
}
</style>
