<template>
  <div class="main">
    <!-- 搜索栏 -->
    <el-form :inline="true" :model="query" class="search-form bg-bg_color pl-8 pt-4 mb-3">
      <el-form-item label="用户名">
        <el-input v-model="query.username" placeholder="请输入用户名" clearable />
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.status" placeholder="请选择" clearable style="width:120px">
          <el-option label="正常" value="2" />
          <el-option label="停用" value="1" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="onSearch">查询</el-button>
        <el-button @click="onReset">重置</el-button>
      </el-form-item>
    </el-form>

    <!-- 操作栏 -->
    <div class="flex justify-between mb-3">
      <el-button type="primary" @click="openDialog()">新增用户</el-button>
    </div>

    <!-- 表格 -->
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="userId" label="ID" width="60" />
      <el-table-column prop="username" label="用户名" />
      <el-table-column prop="nickName" label="昵称" />
      <el-table-column prop="phone" label="手机号" />
      <el-table-column prop="email" label="邮箱" />
      <el-table-column label="状态" width="80">
        <template #default="{ row }">
          <el-tag :type="row.status === '2' ? 'success' : 'danger'">
            {{ row.status === '2' ? '正常' : '停用' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="createTime" label="创建时间" width="160" />
      <el-table-column label="操作" width="180" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
          <el-button link type="primary" @click="onResetPwd(row)">重置密码</el-button>
          <el-popconfirm title="确认删除？" @confirm="onDelete(row)">
            <template #reference>
              <el-button link type="danger">删除</el-button>
            </template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>

    <!-- 分页 -->
    <el-pagination
      class="mt-4 flex justify-end"
      v-model:current-page="query.pageIndex"
      v-model:page-size="query.pageSize"
      :total="total"
      :page-sizes="[10, 20, 50]"
      layout="total, sizes, prev, pager, next"
      @change="loadData"
    />

    <!-- 新增/编辑弹窗 -->
    <el-dialog v-model="dialogVisible" :title="form.userId ? '编辑用户' : '新增用户'" width="500px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="80px">
        <el-form-item label="用户名" prop="username">
          <el-input v-model="form.username" :disabled="!!form.userId" />
        </el-form-item>
        <el-form-item label="昵称" prop="nickName">
          <el-input v-model="form.nickName" />
        </el-form-item>
        <el-form-item v-if="!form.userId" label="密码" prop="password">
          <el-input v-model="form.password" type="password" show-password />
        </el-form-item>
        <el-form-item label="手机号" prop="phone">
          <el-input v-model="form.phone" />
        </el-form-item>
        <el-form-item label="邮箱" prop="email">
          <el-input v-model="form.email" />
        </el-form-item>
        <el-form-item label="状态">
          <el-radio-group v-model="form.status">
            <el-radio value="2">正常</el-radio>
            <el-radio value="1">停用</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="角色" prop="roleId">
          <el-select v-model="form.roleId" placeholder="请选择角色" style="width:100%">
            <el-option v-for="r in roleOptions" :key="r.roleId" :label="r.roleName" :value="r.roleId" />
          </el-select>
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
import { getUserList, createUser, updateUser, deleteUser, resetUserPwd } from "@/api/system";
import { getRoleList } from "@/api/system";

defineOptions({ name: "SysUserManage" });

const loading = ref(false);
const submitting = ref(false);
const list = ref<any[]>([]);
const total = ref(0);
const dialogVisible = ref(false);
const formRef = ref<FormInstance>();
const roleOptions = ref<any[]>([]);

const query = reactive({ username: "", status: "", pageIndex: 1, pageSize: 10 });
const form = reactive<any>({ userId: null, username: "", nickName: "", password: "", phone: "", email: "", status: "2", roleId: null });

const rules = {
  username: [{ required: true, message: "请输入用户名", trigger: "blur" }],
  password: [{ required: true, message: "请输入密码", trigger: "blur" }],
  roleId: [{ required: true, message: "请选择角色", trigger: "change" }]
};

async function loadData() {
  loading.value = true;
  try {
    const res: any = await getUserList(query);
    list.value = res.data || [];
    total.value = res.count || 0;
  } finally {
    loading.value = false;
  }
}

async function loadRoles() {
  const res: any = await getRoleList({ pageIndex: 1, pageSize: 100 });
  roleOptions.value = res.data || [];
}

function onSearch() { query.pageIndex = 1; loadData(); }
function onReset() { Object.assign(query, { username: "", status: "", pageIndex: 1 }); loadData(); }

function openDialog(row?: any) {
  Object.assign(form, { userId: null, username: "", nickName: "", password: "", phone: "", email: "", status: "2", roleId: null });
  if (row) Object.assign(form, row);
  dialogVisible.value = true;
}

async function onSubmit() {
  await formRef.value?.validate();
  submitting.value = true;
  try {
    if (form.userId) {
      await updateUser(form);
    } else {
      await createUser(form);
    }
    ElMessage.success("操作成功");
    dialogVisible.value = false;
    loadData();
  } finally {
    submitting.value = false;
  }
}

async function onDelete(row: any) {
  await deleteUser({ ids: [row.userId] });
  ElMessage.success("删除成功");
  loadData();
}

async function onResetPwd(row: any) {
  await resetUserPwd({ userId: row.userId, password: "123456" });
  ElMessage.success("密码已重置为 123456");
}

onMounted(() => { loadData(); loadRoles(); });
</script>
