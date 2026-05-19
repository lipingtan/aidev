<template>
  <div class="main">
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="operName" label="操作人员" />
      <el-table-column prop="operUrl" label="请求URL" show-overflow-tooltip />
      <el-table-column prop="requestMethod" label="请求方式" width="100" />
      <el-table-column prop="operIp" label="操作IP" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '成功' : '失败' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="latencyTime" label="耗时" width="100" />
      <el-table-column prop="operTime" label="操作时间" width="160" />
    </el-table>
    <el-pagination class="mt-4 flex justify-end" v-model:current-page="query.pageIndex" v-model:page-size="query.pageSize" :total="total" layout="total, sizes, prev, pager, next" @change="loadData" />
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { getOperaLogList } from "@/api/system";
defineOptions({ name: "OperLog" });
const loading = ref(false), list = ref<any[]>([]), total = ref(0);
const query = reactive({ pageIndex: 1, pageSize: 10 });
async function loadData() { loading.value = true; try { const res: any = await getOperaLogList(query); list.value = res.data || []; total.value = res.count || 0; } finally { loading.value = false; } }
onMounted(loadData);
</script>
