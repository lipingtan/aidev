<template>
  <div class="main">
    <el-form :inline="true" :model="query" class="search-form bg-bg_color pl-8 pt-4 mb-3">
      <el-form-item label="游戏ID"><el-input v-model="query.gameId" placeholder="请输入游戏ID" clearable /></el-form-item>
      <el-form-item label="DLC名称"><el-input v-model="query.name" placeholder="请输入DLC名称" clearable /></el-form-item>
      <el-form-item><el-button type="primary" @click="onSearch">查询</el-button><el-button @click="onReset">重置</el-button></el-form-item>
    </el-form>
    <div class="flex justify-between mb-3">
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
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === 1 ? 'success' : 'danger'">{{ row.status === 1 ? '上架' : '下架' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="downloadCount" label="下载次数" width="90" />
      <el-table-column prop="createdAt" label="创建时间" width="160" />
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

    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑DLC' : '新增DLC'" width="560px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="110px">
        <el-form-item label="游戏ID" prop="gameId"><el-input v-model.number="form.gameId" /></el-form-item>
        <el-form-item label="DLC Key" prop="dlcKey"><el-input v-model="form.dlcKey" /></el-form-item>
        <el-form-item label="DLC名称" prop="name"><el-input v-model="form.name" /></el-form-item>
        <el-form-item label="版本" prop="version"><el-input v-model="form.version" /></el-form-item>
        <el-form-item label="描述"><el-input v-model="form.description" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="价格(元)" prop="price"><el-input-number v-model="form.price" :min="0" :precision="2" style="width:100%" /></el-form-item>
        <el-form-item label="是否免费">
          <el-radio-group v-model="form.isFree">
            <el-radio :value="1">免费</el-radio>
            <el-radio :value="2">付费</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio :value="1">上架</el-radio>
            <el-radio :value="2">下架</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="最低游戏版本"><el-input v-model="form.minGameVersion" /></el-form-item>
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
import { getDlcList, createDlc, updateDlc, deleteDlc } from "@/api/game";
defineOptions({ name: "DlcManage" });

const loading = ref(false), submitting = ref(false), list = ref<any[]>([]), total = ref(0), dialogVisible = ref(false), formRef = ref<FormInstance>();
const query = reactive({ gameId: "", name: "", pageIndex: 1, pageSize: 10 });
const form = reactive<any>({ id: null, gameId: "", dlcKey: "", name: "", version: "", description: "", price: 0, isFree: 1, status: 1, minGameVersion: "" });
const rules = {
  gameId: [{ required: true, message: "请输入游戏ID", trigger: "blur" }],
  dlcKey: [{ required: true, message: "请输入DLC Key", trigger: "blur" }],
  name: [{ required: true, message: "请输入DLC名称", trigger: "blur" }],
  version: [{ required: true, message: "请输入版本", trigger: "blur" }],
  price: [{ required: true, message: "请输入价格", trigger: "blur" }]
};

async function loadData() {
  loading.value = true;
  try {
    const res: any = await getDlcList(query);
    list.value = res.data || [];
    total.value = res.count || 0;
  } finally {
    loading.value = false;
  }
}

function onSearch() { query.pageIndex = 1; loadData(); }
function onReset() { Object.assign(query, { gameId: "", name: "", pageIndex: 1 }); loadData(); }

function openDialog(row?: any) {
  Object.assign(form, { id: null, gameId: "", dlcKey: "", name: "", version: "", description: "", price: 0, isFree: 1, status: 1, minGameVersion: "" });
  if (row) {
    Object.assign(form, row);
    // 分→元
    form.price = row.price / 100;
  }
  dialogVisible.value = true;
}

async function onSubmit() {
  await formRef.value?.validate();
  submitting.value = true;
  try {
    const data = { ...form, price: Math.round(form.price * 100) };
    form.id ? await updateDlc(form.id, data) : await createDlc(data);
    ElMessage.success("操作成功");
    dialogVisible.value = false;
    loadData();
  } finally {
    submitting.value = false;
  }
}

async function onDelete(row: any) {
  await deleteDlc(row.id);
  ElMessage.success("删除成功");
  loadData();
}

onMounted(loadData);
</script>
