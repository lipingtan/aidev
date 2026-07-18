<template>
  <div class="plugin-page">
    <div class="page-header">
      <h3>玩家管理</h3>
      <div></div>
    </div>
    <el-form :inline="true" :model="query" style="margin-bottom: 16px">
      <el-form-item label="游戏ID"><el-input v-model="query.gameId" placeholder="游戏ID" clearable /></el-form-item>
      <el-form-item label="UID"><el-input v-model="query.uid" placeholder="UID" clearable /></el-form-item>
      <el-form-item label="昵称"><el-input v-model="query.nickname" placeholder="昵称" clearable /></el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.status" placeholder="全部" clearable style="width:100px">
          <el-option label="正常" :value="1" />
          <el-option label="封禁" :value="2" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="onSearch">查询</el-button>
        <el-button @click="onReset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="gameId" label="游戏ID" width="80" />
      <el-table-column prop="uid" label="UID" width="120" />
      <el-table-column prop="nickname" label="昵称" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column prop="platform" label="平台" width="80" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '正常' : '封禁' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="lastLoginAt" label="最后登录时间" width="160" />
      <el-table-column prop="lastLoginIp" label="最后登录IP" width="130" />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDetail(row)">详情</el-button>
          <el-button v-if="row.status === 1" link type="danger" @click="openBanDialog(row)">封禁</el-button>
          <el-button v-if="row.status === 2" link type="success" @click="onUnban(row)">解封</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-pagination
      style="margin-top: 16px; display: flex; justify-content: flex-end"
      v-model:current-page="query.pageIndex"
      v-model:page-size="query.pageSize"
      :total="total"
      layout="total, sizes, prev, pager, next"
      @change="loadData"
    />

    <!-- 详情弹窗 -->
    <el-dialog v-model="detailVisible" title="玩家详情" width="520px">
      <el-descriptions :column="2" border>
        <el-descriptions-item label="ID">{{ detail.id }}</el-descriptions-item>
        <el-descriptions-item label="游戏ID">{{ detail.gameId }}</el-descriptions-item>
        <el-descriptions-item label="UID">{{ detail.uid }}</el-descriptions-item>
        <el-descriptions-item label="昵称">{{ detail.nickname }}</el-descriptions-item>
        <el-descriptions-item label="邮箱" :span="2">{{ detail.email }}</el-descriptions-item>
        <el-descriptions-item label="平台">{{ detail.platform }}</el-descriptions-item>
        <el-descriptions-item label="平台ID">{{ detail.platformId }}</el-descriptions-item>
        <el-descriptions-item label="设备ID" :span="2">{{ detail.deviceId }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="detail.status === 1 ? 'success' : 'danger'">{{ detail.status === 1 ? '正常' : '封禁' }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="封禁原因">{{ detail.banReason || '-' }}</el-descriptions-item>
        <el-descriptions-item label="最后登录时间">{{ detail.lastLoginAt }}</el-descriptions-item>
        <el-descriptions-item label="最后登录IP">{{ detail.lastLoginIp }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 封禁弹窗 -->
    <el-dialog v-model="banVisible" title="封禁玩家" width="400px">
      <el-form ref="banFormRef" :model="banForm" :rules="banRules" label-width="80px">
        <el-form-item label="封禁原因" prop="banReason">
          <el-input v-model="banForm.banReason" type="textarea" :rows="3" placeholder="请输入封禁原因" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="banVisible = false">取消</el-button>
        <el-button type="danger" :loading="submitting" @click="onBanSubmit">确认封禁</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { ElMessage } from "element-plus";
import type { FormInstance } from "element-plus";

const API_BASE = "/api/v1/plugin/game/player";

const loading = ref(false);
const submitting = ref(false);
const list = ref<any[]>([]);
const total = ref(0);
const detailVisible = ref(false);
const banVisible = ref(false);
const banFormRef = ref<FormInstance>();
const currentId = ref<number>(0);

const query = reactive({ gameId: "", uid: "", nickname: "", status: "" as any, pageIndex: 1, pageSize: 10 });
const detail = reactive<any>({});
const banForm = reactive({ banReason: "" });
const banRules = { banReason: [{ required: true, message: "请输入封禁原因", trigger: "blur" }] };

// 获取 token
function getToken(): string {
  try {
    // 优先从 Cookie 获取
    const cookieMatch = document.cookie.match(/authorized-token=([^;]+)/);
    if (cookieMatch) {
      const data = JSON.parse(decodeURIComponent(cookieMatch[1]));
      return data?.accessToken || "";
    }
    // 兜底从 localStorage 获取（pure-admin 使用 responsive- 前缀）
    const stored = localStorage.getItem("responsive-user-info");
    if (stored) {
      const data = JSON.parse(stored);
      return data?.accessToken || "";
    }
  } catch {}
  return "";
}

// 通用请求
async function request(method: string, url: string, body?: any) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${getToken()}`
  };
  const opts: RequestInit = { method, headers };
  if (body) {
    headers["Content-Type"] = "application/json";
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(url, opts);
  return res.json();
}

async function loadData() {
  loading.value = true;
  try {
    const params = new URLSearchParams();
    params.set("pageIndex", String(query.pageIndex));
    params.set("pageSize", String(query.pageSize));
    if (query.gameId) params.set("gameId", query.gameId);
    if (query.uid) params.set("uid", query.uid);
    if (query.nickname) params.set("nickname", query.nickname);
    if (query.status) params.set("status", String(query.status));
    const res = await request("GET", `${API_BASE}?${params.toString()}`);
    if (res.code === 200) {
      list.value = res.data?.list || res.data || [];
      total.value = res.data?.count || res.count || 0;
    }
  } finally {
    loading.value = false;
  }
}

function onSearch() { query.pageIndex = 1; loadData(); }
function onReset() { Object.assign(query, { gameId: "", uid: "", nickname: "", status: "", pageIndex: 1 }); loadData(); }

async function openDetail(row: any) {
  try {
    const res = await request("GET", `${API_BASE}/${row.id}`);
    if (res.code === 200) {
      Object.assign(detail, res.data || row);
    } else {
      Object.assign(detail, row);
    }
  } catch {
    Object.assign(detail, row);
  }
  detailVisible.value = true;
}

function openBanDialog(row: any) {
  currentId.value = row.id;
  banForm.banReason = "";
  banVisible.value = true;
}

async function onBanSubmit() {
  await banFormRef.value?.validate();
  submitting.value = true;
  try {
    const res = await request("PUT", `${API_BASE}/${currentId.value}/ban`, { status: 2, banReason: banForm.banReason });
    if (res.code === 200) {
      ElMessage.success("封禁成功");
      banVisible.value = false;
      loadData();
    } else {
      ElMessage.error(res.msg || "操作失败");
    }
  } finally {
    submitting.value = false;
  }
}

async function onUnban(row: any) {
  const res = await request("PUT", `${API_BASE}/${row.id}/ban`, { status: 1, banReason: "" });
  if (res.code === 200) {
    ElMessage.success("解封成功");
    loadData();
  } else {
    ElMessage.error(res.msg || "操作失败");
  }
}

onMounted(loadData);
</script>

<style scoped>
.plugin-page {
  padding: 20px;
}
.plugin-page .page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
  padding-bottom: 16px;
  border-bottom: 1px solid #ebeef5;
}
.plugin-page .page-header h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 600;
  color: #1e293b;
}
.plugin-page .el-table {
  border-radius: 8px;
}
</style>
