<template>
  <div class="main p-4">
    <el-alert title="Swagger API 文档" type="info" :closable="false" description="请访问后端 Swagger 文档查看接口详情" />
    <div class="mt-4">
      <el-link type="primary" :href="`http://localhost:8000/swagger/admin/index.html`" target="_blank">
        打开 Swagger 文档
      </el-link>
    </div>
  </div>
</template>
<script setup lang="ts">
defineOptions({ name: "Swagger" });
</script>
