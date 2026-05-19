<template>
  <div class="main p-4">
    <el-form :model="form" label-width="120px" v-loading="loading">
      <el-form-item label="系统名称">
        <el-input v-model="form.sys_app_name" style="width:300px" />
      </el-form-item>
      <el-form-item label="系统Logo">
        <el-input v-model="form.sys_app_logo" style="width:400px" />
      </el-form-item>
      <el-form-item>
        <el-button type="primary" :loading="submitting" @click="onSave">保存</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { ElMessage } from "element-plus";
import { http } from "@/utils/http";
defineOptions({ name: "SysConfigSet" });
const loading = ref(false), submitting = ref(false);
const form = reactive<any>({ sys_app_name: "", sys_app_logo: "" });
onMounted(async () => {
  loading.value = true;
  try {
    const res: any = await http.request("get", "/api/v1/set-config");
    if (res?.data) Object.assign(form, res.data);
  } finally { loading.value = false; }
});
async function onSave() {
  submitting.value = true;
  try {
    await http.request("put", "/api/v1/set-config", { data: Object.entries(form).map(([k, v]) => ({ configKey: k, configValue: v })) });
    ElMessage.success("保存成功");
  } finally { submitting.value = false; }
}
</script>
