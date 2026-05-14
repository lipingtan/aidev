<template>
  <div class="main">
    <div class="flex justify-between mb-3">
      <el-button type="primary" @click="openDialog()">新增部门</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border row-key="deptId" default-expand-all :tree-props="{ children: 'children' }">
      <el-table-column prop="deptName" label="部门名称" />
      <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column prop="leader" label="负责人" />
      <el-table-column prop="phone" label="联系电话" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === '2' ? 'success' : 'danger'">{{ row.status === '2' ? '正常' : '停用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
          <el-popconfirm title="确认删除？" @confirm="onDelete(row)">
            <template #reference><el-button link type="danger">删除</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <el-dialog v-model="dialogVisible" :title="form.deptId ? '编辑部门' : '新增部门'" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="上级部门">
          <el-tree-select v-model="form.parentId" :data="treeData" :props="{ label: 'deptName', value: 'deptId', children: 'children' }" check-strictly clearable placeholder="请选择上级部门" style="width:100%" />
        </el-form-item>
        <el-form-item label="部门名称" prop="deptName"><el-input v-model="form.deptName" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
        <el-form-item label="负责人"><el-input v-model="form.leader" /></el-form-item>
        <el-form-item label="联系电话"><el-input v-model="form.phone" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="form.email" /></el-form-item>
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
import { getDeptList, createDept, updateDept, deleteDept } from "@/api/system";
defineOptions({ name: "SysDeptManage" });
const loading = ref(false), submitting = ref(false), list = ref<any[]>([]), treeData = ref<any[]>([]), dialogVisible = ref(false), formRef = ref<FormInstance>();
const form = reactive<any>({ deptId: null, parentId: 0, deptName: "", sort: 0, leader: "", phone: "", email: "", status: "2" });
const rules = { deptName: [{ required: true, message: "请输入部门名称", trigger: "blur" }] };
async function loadData() { loading.value = true; try { const res: any = await getDeptList({}); list.value = res.data || []; treeData.value = res.data || []; } finally { loading.value = false; } }
function openDialog(row?: any) { Object.assign(form, { deptId: null, parentId: 0, deptName: "", sort: 0, leader: "", phone: "", email: "", status: "2" }); if (row) Object.assign(form, row); dialogVisible.value = true; }
async function onSubmit() { await formRef.value?.validate(); submitting.value = true; try { form.deptId ? await updateDept(form.deptId, form) : await createDept(form); ElMessage.success("操作成功"); dialogVisible.value = false; loadData(); } finally { submitting.value = false; } }
async function onDelete(row: any) { await deleteDept(row.deptId); ElMessage.success("删除成功"); loadData(); }
onMounted(loadData);
</script>
