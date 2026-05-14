<template>
  <div class="main">
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="ipaddr" label="IP地址" />
      <el-table-column prop="loginLocation" label="登录地点" />
      <el-table-column prop="browser" label="浏览器" />
      <el-table-column prop="os" label="操作系统" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === '2' ? 'success' : 'danger'">{{ row.status === '2' ? '成功' : '失败' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="msg" label="消息" />
      <el-table-column prop="loginTime" label="登录时间" width="160" />
    </el-table>
    <el-pagination class="mt-4 flex justify-end" v-model:current-page="query.pageIndex" v-model:page-size="query.pageSize" :total="total" layout="total, sizes, prev, pager, next" @change="loadData" />
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { getLoginLogList } from "@/api/system";
defineOptions({ name: "SysLoginLogManage" });
const loading = ref(false), list = ref<any[]>([]), total = ref(0);
const query = reactive({ pageIndex: 1, pageSize: 10 });
async function loadData() { loading.value = true; try { const res: any = await getLoginLogList(query); list.value = res.data || []; total.value = res.count || 0; } finally { loading.value = false; } }
onMounted(loadData);
</script>
