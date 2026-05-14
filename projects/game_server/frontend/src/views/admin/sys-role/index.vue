<template>
  <div class="main">
    <el-form :inline="true" :model="query" class="search-form bg-bg_color pl-8 pt-4 mb-3">
      <el-form-item label="角色名称"><el-input v-model="query.roleName" placeholder="请输入角色名称" clearable /></el-form-item>
      <el-form-item><el-button type="primary" @click="onSearch">查询</el-button><el-button @click="onReset">重置</el-button></el-form-item>
    </el-form>
    <div class="flex justify-between mb-3">
      <el-button type="primary" @click="openDialog()">新增角色</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="roleId" label="ID" width="60" />
      <el-table-column prop="roleName" label="角色名称" />
      <el-table-column prop="roleKey" label="角色标识" />
      <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column prop="remark" label="备注" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
          <el-button link type="warning" @click="openMenuDialog(row)">分配菜单</el-button>
          <el-popconfirm title="确认删除？" @confirm="onDelete(row)">
            <template #reference><el-button link type="danger">删除</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination class="mt-4 flex justify-end" v-model:current-page="query.pageIndex" v-model:page-size="query.pageSize" :total="total" layout="total, sizes, prev, pager, next" @change="loadData" />

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="form.roleId ? '编辑角色' : '新增角色'" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="角色名称" prop="roleName"><el-input v-model="form.roleName" /></el-form-item>
        <el-form-item label="角色标识" prop="roleKey"><el-input v-model="form.roleKey" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
        <el-form-item label="备注"><el-input v-model="form.remark" type="textarea" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="onSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 分配菜单弹窗 -->
    <el-dialog v-model="menuDialogVisible" title="分配菜单" width="500px">
      <el-tree
        ref="menuTreeRef"
        :data="menuTree"
        :props="{ label: 'title', children: 'children' }"
        show-checkbox
        node-key="menuId"
        :default-checked-keys="checkedMenuIds"
        check-strictly
      />
      <template #footer>
        <el-button @click="menuDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="menuSubmitting" @click="onSaveMenu">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { ElMessage } from "element-plus";
import type { FormInstance, ElTree } from "element-plus";
import { getRoleList, createRole, updateRole, deleteRole, getRoleMenuTree } from "@/api/system";
import { getMenuList } from "@/api/system";
import { http } from "@/utils/http";

defineOptions({ name: "SysRoleManage" });

const loading = ref(false), submitting = ref(false), list = ref<any[]>([]), total = ref(0);
const dialogVisible = ref(false), formRef = ref<FormInstance>();
const menuDialogVisible = ref(false), menuSubmitting = ref(false);
const menuTreeRef = ref<InstanceType<typeof ElTree>>();
const menuTree = ref<any[]>([]);
const checkedMenuIds = ref<number[]>([]);
const currentRoleId = ref<number>(0);

const query = reactive({ roleName: "", pageIndex: 1, pageSize: 10 });
const form = reactive<any>({ roleId: null, roleName: "", roleKey: "", sort: 0, remark: "" });
const rules = {
  roleName: [{ required: true, message: "请输入角色名称", trigger: "blur" }],
  roleKey: [{ required: true, message: "请输入角色标识", trigger: "blur" }]
};

async function loadData() {
  loading.value = true;
  try {
    const res: any = await getRoleList(query);
    list.value = res.data || [];
    total.value = res.count || 0;
  } finally { loading.value = false; }
}

function onSearch() { query.pageIndex = 1; loadData(); }
function onReset() { Object.assign(query, { roleName: "", pageIndex: 1 }); loadData(); }

function openDialog(row?: any) {
  Object.assign(form, { roleId: null, roleName: "", roleKey: "", sort: 0, remark: "" });
  if (row) Object.assign(form, row);
  dialogVisible.value = true;
}

async function onSubmit() {
  await formRef.value?.validate();
  submitting.value = true;
  try {
    form.roleId ? await updateRole(form.roleId, form) : await createRole(form);
    ElMessage.success("操作成功");
    dialogVisible.value = false;
    loadData();
  } finally { submitting.value = false; }
}

async function onDelete(row: any) {
  await deleteRole({ ids: [row.roleId] });
  ElMessage.success("删除成功");
  loadData();
}

async function openMenuDialog(row: any) {
  currentRoleId.value = row.roleId;
  // 加载菜单树
  const menuRes: any = await getMenuList({});
  menuTree.value = menuRes.data || [];
  // 加载角色已有菜单
  const roleMenuRes: any = await getRoleMenuTree(row.roleId);
  checkedMenuIds.value = roleMenuRes?.data?.checkedKeys || [];
  menuDialogVisible.value = true;
}

async function onSaveMenu() {
  menuSubmitting.value = true;
  try {
    const checkedKeys = menuTreeRef.value?.getCheckedKeys() as number[];
    const halfCheckedKeys = menuTreeRef.value?.getHalfCheckedKeys() as number[];
    const menuIds = [...checkedKeys, ...halfCheckedKeys];
    await http.request("put", `/api/v1/role/${currentRoleId.value}`, {
      data: { roleId: currentRoleId.value, menuIds }
    });
    ElMessage.success("菜单分配成功");
    menuDialogVisible.value = false;
  } finally { menuSubmitting.value = false; }
}

onMounted(loadData);
</script>
