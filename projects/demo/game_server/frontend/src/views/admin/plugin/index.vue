<template>
  <div class="main">
    <el-card shadow="never" class="mb-4">
      <template #header>
        <div class="flex justify-between items-center">
          <span class="text-lg font-medium">插件管理</span>
          <el-button type="primary" @click="installDialogVisible = true">安装插件</el-button>
        </div>
      </template>
      <el-table v-loading="loading" :data="list" stripe style="width: 100%">
      <el-table-column prop="name" label="名称" width="160" />
      <el-table-column prop="version" label="版本" width="100" />
      <el-table-column prop="status" label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="description" label="描述" />
      <el-table-column prop="installedAt" label="安装时间" width="180" />
      <el-table-column label="操作" width="200" fixed="right">
        <template #default="{ row }">
          <el-button v-if="row.status === 0 || row.status === 2" link type="primary" @click="onStart(row)">启动</el-button>
          <el-button v-if="row.status === 1" link type="warning" @click="onStop(row)">停止</el-button>
          <el-popconfirm title="确认卸载该插件？" @confirm="onUninstall(row)">
            <template #reference><el-button link type="danger">卸载</el-button></template>
          </el-popconfirm>
        </template>
      </el-table-column>
    </el-table>
    </el-card>

    <!-- 安装插件对话框 -->
    <el-dialog v-model="installDialogVisible" title="安装插件" width="500px">
      <el-form :model="installForm" label-width="90px">
        <el-form-item label="安装方式">
          <el-radio-group v-model="installForm.mode">
            <el-radio value="file">本地上传</el-radio>
            <el-radio value="url">远程 URL</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="installForm.mode === 'url'" label="插件名称">
          <el-input v-model="installForm.name" placeholder="请输入插件名称" />
        </el-form-item>
        <el-form-item v-if="installForm.mode === 'file'" label="插件文件">
          <el-upload
            ref="uploadRef"
            :auto-upload="false"
            :limit="1"
            accept=".zip,.tar.gz,.tgz"
            :on-change="onFileChange"
          >
            <el-button type="primary">选择文件</el-button>
            <template #tip>
              <div class="el-upload__tip">支持 .zip / .tar.gz / .tgz 格式</div>
            </template>
          </el-upload>
        </el-form-item>
        <el-form-item v-if="installForm.mode === 'url'" label="远程地址">
          <el-input v-model="installForm.url" placeholder="请输入插件下载 URL" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="installDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="installing" @click="onInstall">确认安装</el-button>
      </template>
    </el-dialog>
  </div>
</template>
<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { ElMessage } from "element-plus";
import type { UploadFile } from "element-plus";
import {
  getPluginList,
  installPluginByUrl,
  installPluginByFile,
  startPlugin,
  stopPlugin,
  uninstallPlugin
} from "@/api/plugin";

defineOptions({ name: "PluginManage" });

const loading = ref(false);
const installing = ref(false);
const list = ref<any[]>([]);
const installDialogVisible = ref(false);
const uploadRef = ref();

const installForm = reactive({
  mode: "file" as "file" | "url",
  name: "",
  url: "",
  file: null as File | null
});

/** 状态标签类型映射 */
function statusTagType(status: number) {
  const map: Record<number, string> = { 0: "info", 1: "success", 2: "warning", 3: "danger" };
  return map[status] ?? "info";
}

/** 状态文本映射 */
function statusLabel(status: number) {
  const map: Record<number, string> = { 0: "已安装", 1: "运行中", 2: "已停止", 3: "异常" };
  return map[status] ?? "未知";
}

/** 加载插件列表 */
async function loadData() {
  loading.value = true;
  try {
    const res: any = await getPluginList();
    list.value = res.data || [];
  } finally {
    loading.value = false;
  }
}

/** 文件选择回调 */
function onFileChange(file: UploadFile) {
  installForm.file = file.raw || null;
}

/** 安装插件 */
async function onInstall() {
  installing.value = true;
  try {
    if (installForm.mode === "url") {
      if (!installForm.name) {
        ElMessage.warning("请输入插件名称");
        return;
      }
      if (!installForm.url) {
        ElMessage.warning("请输入远程 URL");
        return;
      }
      await installPluginByUrl({ name: installForm.name, url: installForm.url });
    } else {
      if (!installForm.file) {
        ElMessage.warning("请选择插件文件");
        return;
      }
      // 本地上传不需要填写名称，后端自动从 plugin.json 提取
      await installPluginByFile("auto", installForm.file);
    }
    ElMessage.success("安装成功");
    installDialogVisible.value = false;
    resetInstallForm();
    loadData();
  } catch (e: any) {
    ElMessage.error(e?.response?.data?.msg || e?.message || "安装失败");
  } finally {
    installing.value = false;
  }
}

/** 重置安装表单 */
function resetInstallForm() {
  installForm.mode = "file";
  installForm.name = "";
  installForm.url = "";
  installForm.file = null;
}

/** 启动插件 */
async function onStart(row: any) {
  const res: any = await startPlugin(row.name);
  if (res?.code === 200) {
    ElMessage.success("启动成功，正在刷新...");
    loadData();
    setTimeout(() => window.location.reload(), 800);
  } else {
    ElMessage.error(res?.msg || "启动失败");
  }
}

/** 停止插件 */
async function onStop(row: any) {
  await stopPlugin(row.name);
  ElMessage.success("停止成功");
  loadData();
  // 刷新侧边栏菜单（插件停止后会注销菜单）
  setTimeout(() => window.location.reload(), 500);
}

/** 卸载插件 */
async function onUninstall(row: any) {
  await uninstallPlugin(row.name, false);
  ElMessage.success("卸载成功");
  loadData();
}

onMounted(loadData);
</script>
