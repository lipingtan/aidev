<template>
  <div class="game-plugin-page" style="padding: 20px">
    <el-form :inline="true" :model="query" style="margin-bottom: 16px">
      <el-form-item label="游戏名称"><el-input v-model="query.name" placeholder="请输入游戏名称" clearable /></el-form-item>
      <el-form-item><el-button type="primary" @click="onSearch">查询</el-button><el-button @click="onReset">重置</el-button></el-form-item>
    </el-form>
    <div style="margin-bottom: 16px">
      <el-button type="primary" @click="openDialog()">新增游戏</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="游戏名称" />
      <el-table-column prop="appId" label="AppID" />
      <el-table-column prop="platform" label="平台" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '上线' : '下线' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createdAt" label="创建时间" width="160" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
          <el-button link type="warning" @click="onRegenSecret(row)">重置密钥</el-button>
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

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑游戏' : '新增游戏'" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="游戏名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="平台" prop="platform">
          <el-select v-model="form.platform" style="width:100%">
            <el-option label="PC" value="pc" />
            <el-option label="Android" value="android" />
            <el-option label="iOS" value="ios" />
            <el-option label="全平台" value="all" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" /></el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">上线</el-radio>
            <el-radio :value="0">下线</el-radio>
          </el-radio-group>
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
import { ref, reactive, onMounted } from "vue";
import { ElMessage } from "element-plus";
import type { FormInstance } from "element-plus";

const API_BASE = "/api/v1/plugin/game/game";

const loading = ref(false);
const submitting = ref(false);
const list = ref<any[]>([]);
const total = ref(0);
const dialogVisible = ref(false);
const formRef = ref<FormInstance>();

const query = reactive({ name: "", pageIndex: 1, pageSize: 10 });
const form = reactive<any>({ id: null, name: "", platform: "all", description: "", status: 1 });
const rules = {
  name: [{ required: true, message: "请输入游戏名称", trigger: "blur" }],
  platform: [{ required: true, message: "请选择平台", trigger: "change" }]
};

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

async function loadData() {
  loading.value = true;
  try {
    const params = new URLSearchParams();
    params.set("pageIndex", String(query.pageIndex));
    params.set("pageSize", String(query.pageSize));
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

function onSearch() { query.pageIndex = 1; loadData(); }
function onReset() { Object.assign(query, { name: "", pageIndex: 1 }); loadData(); }

function openDialog(row?: any) {
  Object.assign(form, { id: null, name: "", platform: "all", description: "", status: 1 });
  if (row) Object.assign(form, row);
  dialogVisible.value = true;
}

async function onSubmit() {
  await formRef.value?.validate();
  submitting.value = true;
  try {
    const res = form.id
      ? await request("PUT", `${API_BASE}/${form.id}`, form)
      : await request("POST", API_BASE, form);
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

async function onRegenSecret(row: any) {
  const res = await request("POST", `${API_BASE}/${row.id}/regen-secret`);
  if (res.code === 200) {
    ElMessage.success("密钥已重置");
    loadData();
  } else {
    ElMessage.error(res.msg || "操作失败");
  }
}

onMounted(loadData);
</script>
