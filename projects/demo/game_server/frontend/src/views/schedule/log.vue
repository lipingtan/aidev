<template>
  <div class="main">
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="jobId" label="任务ID" width="80" />
      <el-table-column prop="jobName" label="任务名称" />
      <el-table-column prop="jobGroup" label="任务组" />
      <el-table-column prop="invokeTarget" label="调用目标" show-overflow-tooltip />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '成功' : '失败' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="执行时间" width="160" />
    </el-table>
    <el-pagination class="mt-4 flex justify-end" v-model:current-page="query.pageIndex" v-model:page-size="query.pageSize" :total="total" layout="total, sizes, prev, pager, next" @change="loadData" />
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { http } from "@/utils/http";
defineOptions({ name: "JobLog" });
const loading = ref(false), list = ref<any[]>([]), total = ref(0);
const query = reactive({ pageIndex: 1, pageSize: 10 });
async function loadData() {
  loading.value = true;
  try {
    const res: any = await http.request("get", "/api/v1/sysjob/log", { params: query });
    if (res?.data?.list) { list.value = res.data.list; total.value = res.data.count || 0; }
    else if (Array.isArray(res?.data)) { list.value = res.data; }
  } finally { loading.value = false; }
}
onMounted(loadData);
</script>
