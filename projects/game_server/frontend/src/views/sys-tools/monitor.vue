<template>
  <div class="main p-4">
    <el-row :gutter="16" v-if="info">
      <el-col :span="12">
        <el-card header="CPU 信息">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="核心数">{{ info.cpu?.cpuNum }}</el-descriptions-item>
            <el-descriptions-item label="使用率">{{ info.cpu?.used }}%</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card header="内存信息">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="总内存">{{ info.mem?.total }}</el-descriptions-item>
            <el-descriptions-item label="已用">{{ info.mem?.used }}</el-descriptions-item>
            <el-descriptions-item label="使用率">{{ info.mem?.usage }}%</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
      <el-col :span="24" class="mt-4">
        <el-card header="服务器信息">
          <el-descriptions :column="3" border>
            <el-descriptions-item label="服务器名称">{{ info.sys?.computerName }}</el-descriptions-item>
            <el-descriptions-item label="操作系统">{{ info.sys?.osName }}</el-descriptions-item>
            <el-descriptions-item label="服务器IP">{{ info.sys?.computerIp }}</el-descriptions-item>
          </el-descriptions>
        </el-card>
      </el-col>
    </el-row>
    <el-empty v-else description="暂无数据" />
  </div>
</template>
<script setup lang="ts">
import { ref, onMounted } from "vue";
import { http } from "@/utils/http";
defineOptions({ name: "ServerMonitor" });
const info = ref<any>(null);
onMounted(async () => { try { const res: any = await http.request("get", "/api/v1/server-monitor"); info.value = res.data; } catch (e) { console.warn("服务监控接口不可用", e); } });
</script>
