<template>
  <div class="main">
    <el-form :inline="true" :model="query" class="search-form bg-bg_color pl-8 pt-4 mb-3">
      <el-form-item label="参数名称"><el-input v-model="query.configName" placeholder="请输入参数名称" clearable /></el-form-item>
      <el-form-item label="参数键名"><el-input v-model="query.configKey" placeholder="请输入参数键名" clearable /></el-form-item>
      <el-form-item><el-button type="primary" @click="onSearch">查询</el-button><el-button @click="onReset">重置</el-button></el-form-item>
    </el-form>
    <div class="flex justify-between mb-3">
      <el-button type="primary" @click="openDialog()">新增参数</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="configId" label="ID" width="60" />
      <el-table-column prop="configName" label="参数名称" />
      <el-table-column prop="configKey" label="参数键名" />
      <el-table-column prop="configValue" label="参数键值" />
      <el-table-column prop="remark" label="备注" />
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
          <el-popconfirm title="确认删除？" @confirm="onDelete(row)">
            <template #reference><el-button link type="danger">删除</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="mt-4 flex justify-end" v-model:current-page="query.pageIndex" v-model:page-size="query.pageSize" :total="total" layout="total, sizes, prev, pager, next" @change="loadData" />
    <el-dialog v-model="dialogVisible" :title="form.configId ? '编辑参数' : '新增参数'" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="参数名称" prop="configName"><el-input v-model="form.configName" /></el-form-item>
        <el-form-item label="参数键名" prop="configKey"><el-input v-model="form.configKey" /></el-form-item>
        <el-form-item label="参数键值" prop="configValue"><el-input v-model="form.configValue" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
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
import { getConfigList, createConfig, updateConfig, deleteConfig } from "@/api/system";
defineOptions({ name: "SysConfigManage" });
const loading = ref(false), submitting = ref(false), list = ref<any[]>([]), total = ref(0), dialogVisible = ref(false), formRef = ref<FormInstance>();
const query = reactive({ configName: "", configKey: "", pageIndex: 1, pageSize: 10 });
const form = reactive<any>({ configId: null, configName: "", configKey: "", configValue: "", remark: "" });
const rules = { configName: [{ required: true, message: "请输入参数名称", trigger: "blur" }], configKey: [{ required: true, message: "请输入参数键名", trigger: "blur" }], configValue: [{ required: true, message: "请输入参数键值", trigger: "blur" }] };
async function loadData() { loading.value = true; try { const res: any = await getConfigList(query); list.value = res.data || []; total.value = res.count || 0; } finally { loading.value = false; } }
function onSearch() { query.pageIndex = 1; loadData(); }
function onReset() { Object.assign(query, { configName: "", configKey: "", pageIndex: 1 }); loadData(); }
function openDialog(row?: any) { Object.assign(form, { configId: null, configName: "", configKey: "", configValue: "", remark: "" }); if (row) Object.assign(form, row); dialogVisible.value = true; }
async function onSubmit() { await formRef.value?.validate(); submitting.value = true; try { form.configId ? await updateConfig(form.configId, form) : await createConfig(form); ElMessage.success("操作成功"); dialogVisible.value = false; loadData(); } finally { submitting.value = false; } }
async function onDelete(row: any) { await deleteConfig({ ids: [row.configId] }); ElMessage.success("删除成功"); loadData(); }
onMounted(loadData);
</script>
