<template>
  <div class="main">
    <div class="flex justify-between mb-3">
      <el-button type="primary" @click="openDialog()">新增菜单</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border row-key="menuId" default-expand-all :tree-props="{ children: 'children' }">
      <el-table-column prop="title" label="菜单名称" />
      <el-table-column prop="menuType" label="类型" width="80">
        <template #default="{ row }">
          <el-tag :type="row.menuType === 'M' ? 'primary' : row.menuType === 'C' ? 'success' : 'info'">
            {{ row.menuType === 'M' ? '目录' : row.menuType === 'C' ? '菜单' : '按钮' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="icon" label="图标" width="80" />
      <el-table-column prop="sort" label="排序" width="80" />
      <el-table-column prop="path" label="路由路径" />
      <el-table-column prop="component" label="组件路径" />
      <el-table-column prop="permission" label="权限标识" />
      <el-table-column label="操作" width="150" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
          <el-popconfirm title="确认删除？" @confirm="onDelete(row)">
            <template #reference><el-button link type="danger">删除</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    <el-dialog v-model="dialogVisible" :title="form.menuId ? '编辑菜单' : '新增菜单'" width="600px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="90px">
        <el-form-item label="菜单类型">
          <el-radio-group v-model="form.menuType">
            <el-radio value="M">目录</el-radio>
            <el-radio value="C">菜单</el-radio>
            <el-radio value="F">按钮</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="菜单名称" prop="title"><el-input v-model="form.title" /></el-form-item>
        <el-form-item label="路由路径" prop="path"><el-input v-model="form.path" /></el-form-item>
        <el-form-item v-if="form.menuType === 'C'" label="组件路径"><el-input v-model="form.component" /></el-form-item>
        <el-form-item label="图标"><el-input v-model="form.icon" /></el-form-item>
        <el-form-item label="排序"><el-input-number v-model="form.sort" :min="0" /></el-form-item>
        <el-form-item label="权限标识"><el-input v-model="form.permission" /></el-form-item>
        <el-form-item label="是否显示">
          <el-radio-group v-model="form.visible">
            <el-radio value="0">显示</el-radio>
            <el-radio value="1">隐藏</el-radio>
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
import { getMenuList, createMenu, updateMenu, deleteMenu } from "@/api/system";
defineOptions({ name: "SysMenuManage" });
const loading = ref(false), submitting = ref(false), list = ref<any[]>([]), dialogVisible = ref(false), formRef = ref<FormInstance>();
const form = reactive<any>({ menuId: null, title: "", menuName: "", path: "", component: "", icon: "", sort: 0, menuType: "C", permission: "", visible: "0", parentId: 0 });
const rules = { title: [{ required: true, message: "请输入菜单名称", trigger: "blur" }] };
async function loadData() { loading.value = true; try { const res: any = await getMenuList({}); list.value = res.data || []; } finally { loading.value = false; } }
function openDialog(row?: any) { Object.assign(form, { menuId: null, title: "", menuName: "", path: "", component: "", icon: "", sort: 0, menuType: "C", permission: "", visible: "0", parentId: 0 }); if (row) Object.assign(form, row); dialogVisible.value = true; }
async function onSubmit() { await formRef.value?.validate(); submitting.value = true; try { form.menuId ? await updateMenu(form.menuId, form) : await createMenu(form); ElMessage.success("操作成功"); dialogVisible.value = false; loadData(); } finally { submitting.value = false; } }
async function onDelete(row: any) { await deleteMenu({ ids: [row.menuId] }); ElMessage.success("删除成功"); loadData(); }
onMounted(loadData);
</script>
