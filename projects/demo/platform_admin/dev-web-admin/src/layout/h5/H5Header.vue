<!--
  H5 顶部标题栏
  使用 Vant NavBar 实现，包含后退按钮和汉堡菜单图标
-->
<script setup lang="ts">
defineProps<{
  title: string
  showBack?: boolean
}>()

const emit = defineEmits<{
  back: []
  menuClick: []
}>()
</script>

<template>
  <van-nav-bar
    :title="title"
    :left-arrow="showBack"
    fixed
    placeholder
    @click-left="emit('back')"
  >
    <template #right>
      <van-icon name="wap-nav" size="20" @click="emit('menuClick')" />
    </template>
  </van-nav-bar>
</template>

<style scoped>
/* Vant NavBar 样式由框架提供，此处可覆盖品牌色 */
:deep(.van-nav-bar) {
  background-color: var(--van-nav-bar-background, #fff);
}
</style>
