<template>
  <div class="main">
    <div class="flex justify-between mb-3">
      <el-button type="primary" @click="openDialog()">新增岗位</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="postId" label="ID" width="60" />
      <el-table-column prop="postName" label="岗位名称" />
      <el-table-column prop="postCode" label="岗位编码" />
      <el-table-column prop="sort" label="排序" width="80" />
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
    <el-dialog v-model="dialogVisible" :title="form.postId ? '编辑岗位' : '新增岗位'" width="450px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="岗位名称" prop="postName"><el-input v-model="form.postName" /></el-form-item>
        <el-form-item label="岗位编码" prop="postCode"><el-input v-model="form.postCode" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
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
import { getPostList, createPost, updatePost, deletePost } from "@/api/system";
defineOptions({ name: "SysPostManage" });
const loading = ref(false), submitting = ref(false), list = ref<any[]>([]), total = ref(0), dialogVisible = ref(false), formRef = ref<FormInstance>();
const query = reactive({ pageIndex: 1, pageSize: 10 });
const form = reactive<any>({ postId: null, postName: "", postCode: "", sort: 0, remark: "" });
const rules = { postName: [{ required: true, message: "请输入岗位名称", trigger: "blur" }], postCode: [{ required: true, message: "请输入岗位编码", trigger: "blur" }] };
async function loadData() { loading.value = true; try { const res: any = await getPostList(query); list.value = res.data || []; total.value = res.count || 0; } finally { loading.value = false; } }
function openDialog(row?: any) { Object.assign(form, { postId: null, postName: "", postCode: "", sort: 0, remark: "" }); if (row) Object.assign(form, row); dialogVisible.value = true; }
async function onSubmit() { await formRef.value?.validate(); submitting.value = true; try { form.postId ? await updatePost(form.postId, form) : await createPost(form); ElMessage.success("操作成功"); dialogVisible.value = false; loadData(); } finally { submitting.value = false; } }
async function onDelete(row: any) { await deletePost(row.postId); ElMessage.success("删除成功"); loadData(); }
onMounted(loadData);
</script>
