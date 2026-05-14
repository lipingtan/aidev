<template>
  <div class="main">
    <el-form :inline="true" :model="query" class="search-form bg-bg_color pl-8 pt-4 mb-3">
      <el-form-item label="接口名称"><el-input v-model="query.title" placeholder="请输入接口名称" clearable /></el-form-item>
      <el-form-item label="接口路径"><el-input v-model="query.path" placeholder="请输入接口路径" clearable /></el-form-item>
      <el-form-item><el-button type="primary" @click="onSearch">查询</el-button><el-button @click="onReset">重置</el-button></el-form-item>
    </el-form>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="title" label="接口名称" />
      <el-table-column prop="path" label="接口路径" show-overflow-tooltip />
      <el-table-column prop="action" label="请求方式" width="100" />
      <el-table-column prop="type" label="类型" width="80" />
    </el-table>
    <el-pagination class="mt-4 flex justify-end" v-model:current-page="query.pageIndex" v-model:page-size="query.pageSize" :total="total" layout="total, sizes, prev, pager, next" @change="loadData" />
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { getApiList } from "@/api/system";
defineOptions({ name: "SysApiManage" });
const loading = ref(false), list = ref<any[]>([]), total = ref(0);
const query = reactive({ title: "", path: "", pageIndex: 1, pageSize: 10 });
async function loadData() { loading.value = true; try { const res: any = await getApiList(query); list.value = res.data || []; total.value = res.count || 0; } finally { loading.value = false; } }
function onSearch() { query.pageIndex = 1; loadData(); }
function onReset() { Object.assign(query, { title: "", path: "", pageIndex: 1 }); loadData(); }
onMounted(loadData);
</script>
