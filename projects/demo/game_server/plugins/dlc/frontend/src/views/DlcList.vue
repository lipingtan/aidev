<template>
  <div class="dlc-plugin-page" style="padding: 20px">
    <div style="margin-bottom: 16px; display: flex; justify-content: space-between">
      <h3 style="margin: 0">DLC 管理</h3>
      <el-button type="primary" @click="dialogVisible = true">新增 DLC</el-button>
    </div>

    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="dlcKey" label="标识" width="120" />
      <el-table-column prop="version" label="版本" width="80" />
      <el-table-column prop="status" label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'info'">
            {{ row.status === 1 ? '上架' : '下架' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="price" label="价格(分)" width="90" />
      <el-table-column prop="downloadCount" label="下载量" width="80" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="onEdit(row)">编辑</el-button>
          <el-button link type="primary" @click="onUpload(row)">上传PCK</el-button>
          <el-popconfirm title="确认删除？" @confirm="onDelete(row)">
            <template #reference>
              <el-button link type="danger">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑 DLC' : '新增 DLC'" width="500px">
      <el-form :model="form" label-width="90px">
        <el-form-item label="游戏ID"><el-input-number v-model="form.gameId" :min="1" /></el-form-item>
        <el-form-item label="DLC标识"><el-input v-model="form.dlcKey" :disabled="!!form.id" /></el-form-item>
        <el-form-item label="名称"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="版本"><el-input v-model="form.version" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" /></el-form-item>
        <el-form-item label="价格(分)"><el-input-number v-model="form.price" :min="0" /></el-form-item>
        <el-form-item label="免费">
          <el-radio-group v-model="form.isFree">
            <el-radio :value="1">是</el-radio>
            <el-radio :value="2">否</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">上架</el-radio>
            <el-radio :value="2">下架</el-radio>
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

const API_BASE = "/api/v1/plugin/dlc";

const loading = ref(false);
const submitting = ref(false);
const list = ref<any[]>([]);
const dialogVisible = ref(false);

const form = reactive<any>({
  id: null,
  gameId: 1,
  dlcKey: "",
  name: "",
  version: "1.0.0",
  description: "",
  price: 0,
  isFree: 1,
  status: 2
});

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
async function request(method: string, path: string, body?: any) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${getToken()}`
  };
  const opts: RequestInit = { method, headers };
  if (body) {
    headers["Content-Type"] = "application/json";
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(API_BASE + path, opts);
  return res.json();
}

async function loadData() {
  loading.value = true;
  try {
    const res = await request("GET", "/list?page=1&pageSize=50");
    if (res.code === 200) {
      list.value = res.data?.list || [];
    }
  } finally {
    loading.value = false;
  }
}

function onEdit(row: any) {
  Object.assign(form, row);
  dialogVisible.value = true;
}

async function onSubmit() {
  submitting.value = true;
  try {
    const res = form.id
      ? await request("PUT", `/${form.id}`, form)
      : await request("POST", "", form);
    if (res.code === 200) {
      ElMessage.success(form.id ? "更新成功" : "创建成功");
      dialogVisible.value = false;
      resetForm();
      loadData();
    } else {
      ElMessage.error(res.msg || "操作失败");
    }
  } finally {
    submitting.value = false;
  }
}

async function onDelete(row: any) {
  const res = await request("DELETE", `/${row.id}`);
  if (res.code === 200) {
    ElMessage.success("删除成功");
    loadData();
  } else {
    ElMessage.error(res.msg || "删除失败");
  }
}

function onUpload(row: any) {
  // TODO: 实现文件上传对话框
  ElMessage.info("PCK 上传功能开发中");
}

function resetForm() {
  Object.assign(form, {
    id: null, gameId: 1, dlcKey: "", name: "", version: "1.0.0",
    description: "", price: 0, isFree: 1, status: 2
  });
}

onMounted(loadData);
</script>
