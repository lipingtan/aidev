<template>
  <div class="plugin-page">
    <div class="page-header">
      <h3>支付配置</h3>
      <div>
        <el-button type="primary" @click="openDialog()">新增配置</el-button>
      </div>
    </div>
    <el-form :inline="true" :model="query" style="margin-bottom: 16px">
      <el-form-item label="游戏ID"><el-input v-model="query.gameId" placeholder="请输入游戏ID" clearable /></el-form-item>
      <el-form-item><el-button type="primary" @click="onSearch">查询</el-button><el-button @click="onReset">重置</el-button></el-form-item>
    </el-form>
    <el-table v-loading="loading" :data="list" stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="gameId" label="游戏ID" width="80" />
      <el-table-column prop="channel" label="渠道标识" width="100" />
      <el-table-column prop="channelName" label="渠道名称" />
      <el-table-column prop="env" label="环境" width="100" />
      <el-table-column label="启用状态" width="90">
        <template #default="{ row }">
          <el-tag :type="row.enabled === 1 ? 'success' : 'danger'">{{ row.enabled === 1 ? '启用' : '禁用' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="openDialog(row)">编辑</el-button>
        </template>
      </el-table-column>
    </el-table>
    <el-pagination
      style="margin-top: 16px; display: flex; justify-content: flex-end"
      v-model:current-page="query.pageIndex"
      v-model:page-size="query.pageSize"
      :total="total"
      layout="total, sizes, prev, pager, next"
      @change="loadData"
    />

    <!-- 新增/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="form.id ? '编辑支付配置' : '新增支付配置'" width="560px">
      <el-form ref="formRef" :model="form" :rules="rules" label-width="130px">
        <el-form-item label="游戏ID" prop="gameId"><el-input v-model="form.gameId" /></el-form-item>
        <el-form-item label="渠道" prop="channel">
          <el-select v-model="form.channel" style="width:100%" @change="onChannelChange">
            <el-option label="Stripe" value="stripe" />
            <el-option label="支付宝" value="alipay" />
            <el-option label="微信支付" value="wechat" />
          </el-select>
        </el-form-item>
        <el-form-item label="渠道名称" prop="channelName"><el-input v-model="form.channelName" /></el-form-item>
        <el-form-item label="启用状态" prop="enabled">
          <el-select v-model="form.enabled" style="width:100%">
            <el-option label="启用" :value="1" />
            <el-option label="禁用" :value="2" />
          </el-select>
        </el-form-item>
        <el-form-item label="环境" prop="env">
          <el-select v-model="form.env" style="width:100%">
            <el-option label="沙箱" value="sandbox" />
            <el-option label="生产" value="production" />
          </el-select>
        </el-form-item>

        <!-- Stripe 字段组 -->
        <template v-if="form.channel === 'stripe'">
          <el-form-item label="Publishable Key"><el-input v-model="form.stripePublishableKey" /></el-form-item>
          <el-form-item label="Secret Key"><el-input v-model="form.stripeSecretKey" show-password :placeholder="form.id ? '不修改请留空' : ''" /></el-form-item>
          <el-form-item label="Webhook Secret"><el-input v-model="form.stripeWebhookSecret" show-password :placeholder="form.id ? '不修改请留空' : ''" /></el-form-item>
        </template>

        <!-- 支付宝字段组 -->
        <template v-if="form.channel === 'alipay'">
          <el-form-item label="App ID"><el-input v-model="form.alipayAppId" /></el-form-item>
          <el-form-item label="应用私钥"><el-input v-model="form.alipayPrivateKey" type="textarea" :rows="4" :placeholder="form.id ? '不修改请留空' : ''" /></el-form-item>
          <el-form-item label="支付宝公钥"><el-input v-model="form.alipayPublicKey" type="textarea" :rows="4" /></el-form-item>
          <el-form-item label="回调地址"><el-input v-model="form.alipayNotifyUrl" /></el-form-item>
        </template>

        <!-- 微信字段组 -->
        <template v-if="form.channel === 'wechat'">
          <el-form-item label="App ID"><el-input v-model="form.wechatAppId" /></el-form-item>
          <el-form-item label="商户号"><el-input v-model="form.wechatMchId" /></el-form-item>
          <el-form-item label="API Key"><el-input v-model="form.wechatApiKey" show-password :placeholder="form.id ? '不修改请留空' : ''" /></el-form-item>
          <el-form-item label="回调地址"><el-input v-model="form.wechatNotifyUrl" /></el-form-item>
        </template>
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
import { request } from "../utils/request";

const API_BASE = "/api/v1/plugin/game/payment";

const loading = ref(false);
const submitting = ref(false);
const list = ref<any[]>([]);
const total = ref(0);
const dialogVisible = ref(false);
const formRef = ref<FormInstance>();

const query = reactive({ gameId: "", pageIndex: 1, pageSize: 10 });

const defaultForm = () => ({
  id: null, gameId: "", channel: "stripe", channelName: "", enabled: 1, env: "sandbox",
  stripePublishableKey: "", stripeSecretKey: "", stripeWebhookSecret: "",
  alipayAppId: "", alipayPrivateKey: "", alipayPublicKey: "", alipayNotifyUrl: "",
  wechatAppId: "", wechatMchId: "", wechatApiKey: "", wechatNotifyUrl: ""
});
const form = reactive<any>(defaultForm());
const rules = {
  gameId: [{ required: true, message: "请输入游戏ID", trigger: "blur" }],
  channel: [{ required: true, message: "请选择渠道", trigger: "change" }],
  channelName: [{ required: true, message: "请输入渠道名称", trigger: "blur" }],
  enabled: [{ required: true, message: "请选择启用状态", trigger: "change" }],
  env: [{ required: true, message: "请选择环境", trigger: "change" }]
};

async function loadData() {
  loading.value = true;
  try {
    const params = new URLSearchParams();
    params.set("pageIndex", String(query.pageIndex));
    params.set("pageSize", String(query.pageSize));
    if (query.gameId) params.set("gameId", query.gameId);
    const res = await request("GET", `${API_BASE}?${params.toString()}`);
    if (res.code === 200) {
      list.value = res.data?.list || [];
      total.value = res.data?.total || 0;
    }
  } finally {
    loading.value = false;
  }
}

function onSearch() { query.pageIndex = 1; loadData(); }
function onReset() { Object.assign(query, { gameId: "", pageIndex: 1 }); loadData(); }

function onChannelChange() {
  Object.assign(form, {
    stripePublishableKey: "", stripeSecretKey: "", stripeWebhookSecret: "",
    alipayAppId: "", alipayPrivateKey: "", alipayPublicKey: "", alipayNotifyUrl: "",
    wechatAppId: "", wechatMchId: "", wechatApiKey: "", wechatNotifyUrl: ""
  });
}

function openDialog(row?: any) {
  Object.assign(form, defaultForm());
  if (row) {
    Object.assign(form, row);
    // 服务端返回的敏感字段是 "***"，编辑时清空让用户按需填写
    if (form.stripeSecretKey === "***") form.stripeSecretKey = "";
    if (form.stripeWebhookSecret === "***") form.stripeWebhookSecret = "";
    if (form.alipayPrivateKey === "***") form.alipayPrivateKey = "";
    if (form.wechatApiKey === "***") form.wechatApiKey = "";
  }
  dialogVisible.value = true;
}

async function onSubmit() {
  await formRef.value?.validate();
  submitting.value = true;
  try {
    // 编辑用 PUT /payment/:id，新增用 POST /payment
    const res = form.id
      ? await request("PUT", `${API_BASE}/${form.id}`, form)
      : await request("POST", API_BASE, form);
    if (res.code === 200) {
      ElMessage.success("操作成功");
      dialogVisible.value = false;
      loadData();
    } else {
      ElMessage.error(res.msg || "操作失败");
    }
  } finally {
    submitting.value = false;
  }
}

onMounted(loadData);
</script>

<style scoped>
.plugin-page { padding: 20px; }
.plugin-page .page-header {
  display: flex; justify-content: space-between; align-items: center;
  margin-bottom: 20px; padding-bottom: 16px; border-bottom: 1px solid #ebeef5;
}
.plugin-page .page-header h3 { margin: 0; font-size: 18px; font-weight: 600; color: #1e293b; }
</style>
