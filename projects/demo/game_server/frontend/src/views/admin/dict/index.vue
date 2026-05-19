<template>
  <div class="main">
    <div class="flex justify-between mb-3">
      <el-button type="primary" @click="openDialog()">新增字典类型</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="dictId" label="ID" width="60" />
      <el-table-column prop="dictName" label="字典名称" />
      <el-table-column prop="dictType" label="字典类型" />
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
    <el-dialog v-model="dialogVisible" :title="form.dictId ? '编辑字典类型' : '新增字典类型'" width="450px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="字典名称" prop="dictName"><el-input v-model="form.dictName" /></el-form-item>
        <el-form-item label="字典类型" prop="dictType"><el-input v-model="form.dictType" /></el-form-item>
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
import { getDictTypeList, createDictType, updateDictType, deleteDictType } from "@/api/system";
defineOptions({ name: "Dict" });
const loading = ref(false), submitting = ref(false), list = ref<any[]>([]), total = ref(0), dialogVisible = ref(false), formRef = ref<FormInstance>();
const query = reactive({ pageIndex: 1, pageSize: 10 });
const form = reactive<any>({ dictId: null, dictName: "", dictType: "", remark: "" });
const rules = { dictName: [{ required: true, message: "请输入字典名称", trigger: "blur" }], dictType: [{ required: true, message: "请输入字典类型", trigger: "blur" }] };
async function loadData() { loading.value = true; try { const res: any = await getDictTypeList(query); list.value = res.data || []; total.value = res.count || 0; } finally { loading.value = false; } }
function openDialog(row?: any) { Object.assign(form, { dictId: null, dictName: "", dictType: "", remark: "" }); if (row) Object.assign(form, row); dialogVisible.value = true; }
async function onSubmit() { await formRef.value?.validate(); submitting.value = true; try { form.dictId ? await updateDictType(form.dictId, form) : await createDictType(form); ElMessage.success("操作成功"); dialogVisible.value = false; loadData(); } finally { submitting.value = false; } }
async function onDelete(row: any) { await deleteDictType({ ids: [row.dictId] }); ElMessage.success("删除成功"); loadData(); }
onMounted(loadData);
</script>
