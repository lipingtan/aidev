<template>
  <div class="game-plugin-page" style="padding: 20px">
    <el-tabs v-model="activeTab" @tab-change="onTabChange">
      <el-tab-pane label="列表管理" name="list">
        <el-form :inline="true" :model="query" style="margin-bottom: 16px">
          <el-form-item label="游戏">
            <el-select v-model="query.gameId" placeholder="请选择游戏" clearable style="width: 200px">
              <el-option v-for="g in gameList" :key="g.id" :label="g.name" :value="g.id" />
            </el-select>
          </el-form-item>
          <el-form-item label="DLC名称"><el-input v-model="query.name" placeholder="请输入DLC名称" clearable /></el-form-item>
          <el-form-item><el-button type="primary" @click="onSearch">查询</el-button><el-button @click="onReset">重置</el-button></el-form-item>
        </el-form>
        <div style="margin-bottom: 16px">
          <el-button type="primary" @click="openDialog()">新增DLC</el-button>
        </div>
        <el-table v-loading="loading" :data="list" border stripe>
          <el-table-column prop="id" label="ID" width="60" />
          <el-table-column prop="gameId" label="游戏ID" width="80" />
          <el-table-column prop="name" label="DLC名称" />
          <el-table-column prop="dlcKey" label="DLC Key" />
          <el-table-column prop="version" label="版本" width="80" />
          <el-table-column label="价格(元)" width="100">
            <template #default="{ row }">{{ (row.price / 100).toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="是否免费" width="90">
            <template #default="{ row }">
              <el-tag :type="row.isFree === 1 ? 'success' : 'warning'">{{ row.isFree === 1 ? '免费' : '付费' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="文件大小" width="100">
            <template #default="{ row }">{{ row.fileSize > 0 ? formatSize(row.fileSize) : '-' }}</template>
          </el-table-column>
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '上架' : '下架' }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="downloadCount" label="下载次数" width="90" />
          <el-table-column prop="createdAt" label="创建时间" width="160" />
          <el-table-column label="操作" width="220" fixed="right">
            <template #default="{ row }">
              <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
              <el-button v-if="row.filePath" link type="success" @click="onDownload(row)">下载</el-button>
              <el-popconfirm title="确认删除？" @confirm="onDelete(row)">
                <template #reference><el-button link type="danger">删除</el-button></template>
              </el-popconfirm>
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
      </el-tab-pane>

      <el-tab-pane label="统计分析" name="stats">
        <div style="display: flex; gap: 16px; margin-bottom: 16px">
          <el-card shadow="hover" style="width: 240px">
            <el-statistic title="总下载次数" :value="statsData.totalDownloads" />
          </el-card>
          <el-card shadow="hover" style="width: 240px">
            <el-statistic title="总收入(元)" :value="statsData.totalRevenue / 100" :precision="2" />
          </el-card>
        </div>
        <el-table :data="statsData.ranking" border stripe>
          <el-table-column prop="productName" label="DLC名称" />
          <el-table-column prop="downloads" label="下载次数" width="120" />
          <el-table-column label="收入(元)" width="120">
            <template #default="{ row }">{{ (row.revenue / 100).toFixed(2) }}</template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑DLC' : '新增DLC'" width="560px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="游戏" prop="gameId">
          <el-select v-model="form.gameId" placeholder="请选择游戏" style="width: 100%">
            <el-option v-for="g in gameList" :key="g.id" :label="g.name" :value="g.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="DLC Key" prop="dlcKey"><el-input v-model="form.dlcKey" /></el-form-item>
        <el-form-item label="DLC名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="版本" prop="version"><el-input v-model="form.version" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="是否免费">
          <el-radio-group v-model="form.isFree">
            <el-radio :value="1">免费</el-radio>
            <el-radio :value="2">付费</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="价格(元)" prop="price">
          <el-input-number v-model="form.price" :min="0" :precision="2" :disabled="form.isFree === 1" style="width:100%" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">上架</el-radio>
            <el-radio :value="2">下架</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="最低游戏版本"><el-input v-model="form.minGameVersion" /></el-form-item>
        <el-form-item v-if="form.id" label="PCK文件">
          <el-upload :auto-upload="false" :limit="1" accept=".pck" :on-change="handlePckChange" :file-list="pckFileList">
            <el-button type="primary">选择文件</el-button>
            <template #tip><div class="el-upload__tip">仅支持 .pck 文件，大小不超过 500MB</div></template>
          </el-upload>
          <el-button v-if="pckFile" type="success" :loading="uploading" @click="onUploadPck" style="margin-top: 8px">上传</el-button>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from "vue";
import { ElMessage } from "element-plus";
import type { FormInstance, UploadFile } from "element-plus";

const API_BASE = "/api/v1/plugin/game/dlc";
const GAME_API = "/api/v1/plugin/game/game";

const activeTab = ref("list");
const loading = ref(false);
const submitting = ref(false);
const uploading = ref(false);
const dialogVisible = ref(false);
const formRef = ref<FormInstance>();
const list = ref<any[]>([]);
const total = ref(0);
const gameList = ref<any[]>([]);

const query = reactive({ gameId: "" as any, name: "", pageIndex: 1, pageSize: 10 });
const form = reactive<any>({ id: null, gameId: "", dlcKey: "", name: "", version: "", description: "", price: 0, isFree: 1, status: 1, minGameVersion: "" });
const rules = {
  gameId: [{ required: true, message: "请选择游戏", trigger: "change" }],
  dlcKey: [{ required: true, message: "请输入DLC Key", trigger: "blur" }],
  name: [{ required: true, message: "请输入DLC名称", trigger: "blur" }],
  version: [{ required: true, message: "请输入版本", trigger: "blur" }],
  price: [{ required: true, message: "请输入价格", trigger: "blur" }]
};

// PCK 上传相关
const pckFile = ref<File | null>(null);
const pckFileList = ref<UploadFile[]>([]);
const MAX_PCK_SIZE = 500 * 1024 * 1024;

// 统计数据
const statsData = reactive<any>({ totalDownloads: 0, totalRevenue: 0, ranking: [] });

// 获取 token
function getToken(): string {
  try {
    const data = JSON.parse(localStorage.getItem("user-info") || "{}");
    return data?.accessToken || "";
  } catch {
    return "";
  }
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

// 文件大小格式化
function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + " B";
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + " KB";
  return (bytes / (1024 * 1024)).toFixed(1) + " MB";
}

async function loadData() {
  loading.value = true;
  try {
    const params = new URLSearchParams();
    params.set("pageIndex", String(query.pageIndex));
    params.set("pageSize", String(query.pageSize));
    if (query.gameId) params.set("gameId", String(query.gameId));
    if (query.name) params.set("name", query.name);
    const res = await request("GET", `${API_BASE}?${params.toString()}`);
    if (res.code === 200) {
      list.value = res.data?.list || res.data || [];
      total.value = res.data?.count || res.count || 0;
    }
  } finally {
    loading.value = false;
  }
}

async function loadGameList() {
  try {
    const res = await request("GET", `${GAME_API}?pageSize=100`);
    if (res.code === 200) {
      gameList.value = res.data?.list || res.data || [];
    }
  } catch {
    gameList.value = [];
  }
}

async function loadStats() {
  try {
    const params = query.gameId ? `?gameId=${query.gameId}` : "";
    const res = await request("GET", `${API_BASE}/stats${params}`);
    if (res.code === 200) {
      const d = res.data || {};
      statsData.totalDownloads = d.totalDownloads || 0;
      statsData.totalRevenue = d.totalRevenue || 0;
      statsData.ranking = d.ranking || [];
    }
  } catch {
    // 静默处理
  }
}

function onSearch() { query.pageIndex = 1; loadData(); }
function onReset() { Object.assign(query, { gameId: "", name: "", pageIndex: 1 }); loadData(); }

function onTabChange(tab: string) {
  if (tab === "stats") loadStats();
}

function openDialog(row?: any) {
  Object.assign(form, { id: null, gameId: "", dlcKey: "", name: "", version: "", description: "", price: 0, isFree: 1, status: 1, minGameVersion: "" });
  pckFile.value = null;
  pckFileList.value = [];
  if (row) {
    Object.assign(form, row);
    form.price = row.price / 100; // 分→元
  }
  dialogVisible.value = true;
}

async function onSubmit() {
  await formRef.value?.validate();
  submitting.value = true;
  try {
    const data = { ...form, price: Math.round(form.price * 100) }; // 元→分
    const res = form.id
      ? await request("PUT", `${API_BASE}/${form.id}`, data)
      : await request("POST", API_BASE, data);
    if (res.code === 200) {
      ElMessage.success("操作成功");
      dialogVisible.value = false;
      loadData();
    } else {
      ElMessage.error(res.msg || "操作失败");
    }
  } finally {
    submitting.value = false;
  }
}

async function onDelete(row: any) {
  const res = await request("DELETE", `${API_BASE}/${row.id}`);
  if (res.code === 200) {
    ElMessage.success("删除成功");
    loadData();
  } else {
    ElMessage.error(res.msg || "删除失败");
  }
}

async function onDownload(row: any) {
  try {
    const res = await request("GET", `${API_BASE}/${row.id}/download-url`);
    if (res.code === 200) {
      const url = res.data?.url || res.url;
      if (url) window.open(url);
      else ElMessage.error("获取下载地址失败");
    } else {
      ElMessage.error(res.msg || "获取下载地址失败");
    }
  } catch {
    ElMessage.error("获取下载地址失败");
  }
}

function handlePckChange(file: UploadFile) {
  if (file.raw && file.raw.size > MAX_PCK_SIZE) {
    ElMessage.error("文件大小不能超过 500MB");
    pckFileList.value = [];
    pckFile.value = null;
    return;
  }
  pckFile.value = file.raw || null;
  pckFileList.value = file.raw ? [file] : [];
}

async function onUploadPck() {
  if (!pckFile.value || !form.id) return;
  uploading.value = true;
  try {
    const formData = new FormData();
    formData.append("file", pckFile.value);
    const res = await fetch(`${API_BASE}/${form.id}/upload`, {
      method: "POST",
      headers: { Authorization: `Bearer ${getToken()}` },
      body: formData
    });
    const json = await res.json();
    if (json.code === 200) {
      ElMessage.success("上传成功");
      pckFile.value = null;
      pckFileList.value = [];
      loadData();
    } else {
      ElMessage.error(json.msg || "上传失败");
    }
  } catch {
    ElMessage.error("上传失败");
  } finally {
    uploading.value = false;
  }
}

// isFree 联动 price
watch(() => form.isFree, (val) => {
  if (val === 1) form.price = 0;
});

onMounted(() => {
  loadGameList();
  loadData();
});
</script>
