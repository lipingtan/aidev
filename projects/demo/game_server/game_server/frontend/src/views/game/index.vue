<template>
  <div class="main">
    <el-form :inline="true" :model="query" class="search-form bg-bg_color pl-8 pt-4 mb-3">
      <el-form-item label="游戏名称"><el-input v-model="query.name" placeholder="请输入游戏名称" clearable /></el-form-item>
      <el-form-item><el-button type="primary" @click="onSearch">查询</el-button><el-button @click="onReset">重置</el-button></el-form-item>
    </el-form>
    <div class="flex justify-between mb-3">
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
    <el-pagination class="mt-4 flex justify-end" v-model:current-page="query.pageIndex" v-model:page-size="query.pageSize" :total="total" layout="total, sizes, prev, pager, next" @change="loadData" />
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
import { getGameList, createGame, updateGame, deleteGame, regenerateSecret } from "@/api/game";
defineOptions({ name: "GameManage" });
const loading = ref(false), submitting = ref(false), list = ref<any[]>([]), total = ref(0), dialogVisible = ref(false), formRef = ref<FormInstance>();
const query = reactive({ name: "", pageIndex: 1, pageSize: 10 });
const form = reactive<any>({ id: null, name: "", platform: "all", description: "", status: 1 });
const rules = { name: [{ required: true, message: "请输入游戏名称", trigger: "blur" }], platform: [{ required: true, message: "请选择平台", trigger: "change" }] };
async function loadData() { loading.value = true; try { const res: any = await getGameList(query); list.value = res.data || []; total.value = res.count || 0; } finally { loading.value = false; } }
function onSearch() { query.pageIndex = 1; loadData(); }
function onReset() { Object.assign(query, { name: "", pageIndex: 1 }); loadData(); }
function openDialog(row?: any) { Object.assign(form, { id: null, name: "", platform: "all", description: "", status: 1 }); if (row) Object.assign(form, row); dialogVisible.value = true; }
async function onSubmit() { await formRef.value?.validate(); submitting.value = true; try { form.id ? await updateGame(form.id, form) : await createGame(form); ElMessage.success("操作成功"); dialogVisible.value = false; loadData(); } finally { submitting.value = false; } }
async function onDelete(row: any) { await deleteGame(row.id); ElMessage.success("删除成功"); loadData(); }
async function onRegenSecret(row: any) { await regenerateSecret(row.id); ElMessage.success("密钥已重置"); loadData(); }
onMounted(loadData);
</script>
