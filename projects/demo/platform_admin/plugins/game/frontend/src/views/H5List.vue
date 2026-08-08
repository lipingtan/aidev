<template>
  <div class="plugin-page">
    <div class="page-header">
      <h3>H5页面管理</h3>
      <div>
        <el-button type="primary" @click="openDialog()">新增H5页面</el-button>
      </div>
    </div>
    <el-form :inline="true" :model="query" style="margin-bottom: 16px">
      <el-form-item label="游戏ID"><el-input v-model="query.gameId" placeholder="请输入游戏ID" clearable /></el-form-item>
      <el-form-item label="页面名称"><el-input v-model="query.name" placeholder="请输入页面名称" clearable /></el-form-item>
      <el-form-item><el-button type="primary" @click="onSearch">查询</el-button><el-button @click="onReset">重置</el-button></el-form-item>
    </el-form>
    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="gameId" label="游戏ID" width="80" />
      <el-table-column prop="pageKey" label="页面标识" />
      <el-table-column prop="name" label="名称" />
      <el-table-column prop="pageType" label="页面类型" width="100" />
      <el-table-column label="是否外链" width="90">
        <template #default="{ row }">
          <el-tag :type="row.useExternal === 1 ? 'warning' : 'info'">{{ row.useExternal === 1 ? '外链' : '内嵌' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '启用' : '禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="160" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
          <el-popconfirm :title="`确认删除H5页面「${row.name}」？`" @confirm="onDelete(row)">
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
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑H5页面' : '新增H5页面'" width="560px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="游戏ID" prop="gameId"><el-input v-model.number="form.gameId" /></el-form-item>
        <el-form-item label="页面标识" prop="pageKey"><el-input v-model="form.pageKey" /></el-form-item>
        <el-form-item label="名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="页面类型" prop="pageType">
          <el-select v-model="form.pageType" style="width:100%">
            <el-option label="custom" value="custom" />
            <el-option label="商城页" value="商城页" />
            <el-option label="活动页" value="活动页" />
            <el-option label="公告页" value="公告页" />
          </el-select>
        </el-form-item>
        <el-form-item label="链接类型">
          <el-radio-group v-model="form.useExternal">
            <el-radio :value="1">外链</el-radio>
            <el-radio :value="2">内嵌</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.useExternal === 1" label="外链地址" prop="externalUrl">
          <el-input v-model="form.externalUrl" placeholder="请输入外链URL" />
        </el-form-item>
        <el-form-item v-if="form.useExternal === 2" label="页面内容">
          <el-input v-model="form.content" type="textarea" :rows="8" placeholder="请输入页面内容" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">启用</el-radio>
            <el-radio :value="2">禁用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" :rows="2" /></el-form-item>
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
import { request } from "../utils/request";

const API_BASE = "/api/v1/plugin/game/h5";

const loading = ref(false);
const submitting = ref(false);
const list = ref<any[]>([]);
const total = ref(0);
const dialogVisible = ref(false);
const formRef = ref<FormInstance>();

const query = reactive({ gameId: "", name: "", pageIndex: 1, pageSize: 10 });
const form = reactive<any>({
  id: null, gameId: "", pageKey: "", name: "", pageType: "custom",
  useExternal: 1, content: "", externalUrl: "", status: 1, remark: ""
});
const rules = {
  gameId: [{ required: true, message: "请输入游戏ID", trigger: "blur" }],
  pageKey: [{ required: true, message: "请输入页面标识", trigger: "blur" }],
  name: [{ required: true, message: "请输入名称", trigger: "blur" }],
  pageType: [{ required: true, message: "请选择页面类型", trigger: "change" }]
};

async function loadData() {
  loading.value = true;
  try {
    const params = new URLSearchParams();
    params.set("pageIndex", String(query.pageIndex));
    params.set("pageSize", String(query.pageSize));
    if (query.gameId) params.set("gameId", query.gameId);
    if (query.name) params.set("name", query.name);
    const res = await request("GET", `${API_BASE}?${params.toString()}`);
    if (res.code === 200) {
      list.value = res.data?.list || [];
      total.value = res.data?.total || 0;
    }
  } finally {
    loading.value = false;
  }
}

function onSearch() { query.pageIndex = 1; loadData(); }
function onReset() { Object.assign(query, { gameId: "", name: "", pageIndex: 1 }); loadData(); }

function openDialog(row?: any) {
  Object.assign(form, {
    id: null, gameId: "", pageKey: "", name: "", pageType: "custom",
    useExternal: 1, content: "", externalUrl: "", status: 1, remark: ""
  });
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

onMounted(loadData);
</script>

<style scoped>
.plugin-page { padding: 20px; }
.plugin-page .page-header {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: 20px; padding-bottom: 16px; border-bottom: 1px solid #ebeef5;
}
.plugin-page .page-header h3 { margin: 0; font-size: 18px; font-weight: 600; color: #1e293b; }
</style>
