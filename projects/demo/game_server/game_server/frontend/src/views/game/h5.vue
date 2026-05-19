<template>
  <div class="main">
    <el-form :inline="true" :model="query" class="search-form bg-bg_color pl-8 pt-4 mb-3">
      <el-form-item label="游戏ID"><el-input v-model="query.gameId" placeholder="请输入游戏ID" clearable /></el-form-item>
      <el-form-item label="页面名称"><el-input v-model="query.name" placeholder="请输入页面名称" clearable /></el-form-item>
      <el-form-item><el-button type="primary" @click="onSearch">查询</el-button><el-button @click="onReset">重置</el-button></el-form-item>
    </el-form>
    <div class="flex justify-between mb-3">
      <el-button type="primary" @click="openDialog()">新增H5页面</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
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
          <el-popconfirm title="确认删除？" @confirm="onDelete(row)">
            <template #reference><el-button link type="danger">删除</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="mt-4 flex justify-end" v-model:current-page="query.pageIndex" v-model:page-size="query.pageSize" :total="total" layout="total, sizes, prev, pager, next" @change="loadData" />
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
        <el-form-item label="链接类型" prop="useExternal">
          <el-radio-group v-model="form.useExternal">
            <el-radio :value="1">外链</el-radio>
            <el-radio :value="2">内嵌</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.useExternal === 1" label="外链地址" prop="externalUrl">
          <el-input v-model="form.externalUrl" placeholder="请输入外链URL" />
        </el-form-item>
        <el-form-item v-if="form.useExternal === 2" label="页面内容" prop="content">
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
import { getH5PageList, createH5Page, updateH5Page, deleteH5Page } from "@/api/game";
defineOptions({ name: "H5PageManage" });
const loading = ref(false), submitting = ref(false), list = ref<any[]>([]), total = ref(0), dialogVisible = ref(false), formRef = ref<FormInstance>();
const query = reactive({ gameId: "", name: "", pageIndex: 1, pageSize: 10 });
const form = reactive<any>({ id: null, gameId: "", pageKey: "", name: "", pageType: "custom", useExternal: 1, content: "", externalUrl: "", status: 1, remark: "" });
const rules = {
  gameId: [{ required: true, message: "请输入游戏ID", trigger: "blur" }],
  pageKey: [{ required: true, message: "请输入页面标识", trigger: "blur" }],
  name: [{ required: true, message: "请输入名称", trigger: "blur" }],
  pageType: [{ required: true, message: "请选择页面类型", trigger: "change" }],
};
async function loadData() { loading.value = true; try { const res: any = await getH5PageList(query); list.value = res.data || []; total.value = res.count || 0; } finally { loading.value = false; } }
function onSearch() { query.pageIndex = 1; loadData(); }
function onReset() { Object.assign(query, { gameId: "", name: "", pageIndex: 1 }); loadData(); }
function openDialog(row?: any) {
  Object.assign(form, { id: null, gameId: "", pageKey: "", name: "", pageType: "custom", useExternal: 1, content: "", externalUrl: "", status: 1, remark: "" });
  if (row) Object.assign(form, row);
  dialogVisible.value = true;
}
async function onSubmit() {
  await formRef.value?.validate();
  submitting.value = true;
  try {
    form.id ? await updateH5Page(form.id, form) : await createH5Page(form);
    ElMessage.success("操作成功");
    dialogVisible.value = false;
    loadData();
  } finally { submitting.value = false; }
}
async function onDelete(row: any) { await deleteH5Page(row.id); ElMessage.success("删除成功"); loadData(); }
onMounted(loadData);
</script>
