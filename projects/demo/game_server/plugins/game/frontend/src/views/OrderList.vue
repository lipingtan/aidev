<template>
  <div class="game-plugin-page" style="padding: 20px">
    <el-form :inline="true" :model="query" style="margin-bottom: 16px">
      <el-form-item label="游戏ID"><el-input v-model="query.gameId" placeholder="请输入游戏ID" clearable /></el-form-item>
      <el-form-item label="订单号"><el-input v-model="query.orderNo" placeholder="请输入订单号" clearable /></el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.status" placeholder="全部" clearable style="width:120px">
          <el-option label="待支付" value="pending" />
          <el-option label="已支付" value="paid" />
          <el-option label="已退款" value="refunded" />
          <el-option label="失败" value="failed" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="onSearch">查询</el-button>
        <el-button @click="onReset">重置</el-button>
      </el-form-item>
    </el-form>
    <div style="margin-bottom: 16px">
      <el-button type="success" @click="onExport">导出 CSV</el-button>
    </div>
    <el-table v-loading="loading" :data="list" border stripe>
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="orderNo" label="订单号" min-width="180" />
      <el-table-column prop="gameId" label="游戏ID" width="80" />
      <el-table-column prop="playerId" label="玩家ID" width="80" />
      <el-table-column prop="productName" label="商品名称" min-width="120" />
      <el-table-column label="金额" width="120">
        <template #default="{ row }">
          {{ (row.amount / 100).toFixed(2) }} {{ row.currency }}
        </template>
      </el-table-column>
      <el-table-column prop="channel" label="支付渠道" width="100" />
      <el-table-column label="状态" width="90">
        <template #default="{ row }">
          <el-tag :type="statusTagType(row.status)">{{ statusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="paidAt" label="支付时间" width="160" />
      <el-table-column label="操作" width="100" fixed="right">
        <template #default="{ row }">
          <el-popconfirm v-if="row.status === 'paid'" title="确认退款？" @confirm="onRefund(row)">
            <template #reference><el-button link type="danger">退款</el-button></template>
          </el-popconfirm>
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
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from "vue";
import { ElMessage } from "element-plus";

const API_BASE = "/api/v1/plugin/game/order";

const loading = ref(false);
const list = ref<any[]>([]);
const total = ref(0);
const query = reactive({ gameId: "", orderNo: "", status: "", pageIndex: 1, pageSize: 10 });

// 状态映射
const statusTagType = (status: string) => {
  const map: Record<string, string> = { pending: "warning", paid: "success", refunded: "info", failed: "danger" };
  return map[status] || "";
};
const statusLabel = (status: string) => {
  const map: Record<string, string> = { pending: "待支付", paid: "已支付", refunded: "已退款", failed: "失败" };
  return map[status] || status;
};

// 获取 token
function getToken(): string {
  try {
    // 优先从 Cookie 获取
    const cookieMatch = document.cookie.match(/authorized-token=([^;]+)/);
    if (cookieMatch) {
      const data = JSON.parse(decodeURIComponent(cookieMatch[1]));
      return data?.accessToken || "";
    }
    // 兜底从 localStorage 获取（pure-admin 使用 responsive- 前缀）
    const stored = localStorage.getItem("responsive-user-info");
    if (stored) {
      const data = JSON.parse(stored);
      return data?.accessToken || "";
    }
  } catch {}
  return "";
}

// 通用请求
async function request(method: string, url: string, body?: any) {
  const headers: Record<string, string> = {
    Authorization: `Bearer ${getToken()}`
  };
  const opts: RequestInit = { method, headers };
  if (body) {
    headers["Content-Type"] = "application/json";
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(url, opts);
  return res.json();
}

async function loadData() {
  loading.value = true;
  try {
    const params = new URLSearchParams();
    params.set("pageIndex", String(query.pageIndex));
    params.set("pageSize", String(query.pageSize));
    if (query.gameId) params.set("gameId", query.gameId);
    if (query.orderNo) params.set("orderNo", query.orderNo);
    if (query.status) params.set("status", query.status);
    const res = await request("GET", `${API_BASE}?${params.toString()}`);
    if (res.code === 200) {
      list.value = res.data?.list || res.data || [];
      total.value = res.data?.count || res.count || 0;
    }
  } finally {
    loading.value = false;
  }
}

function onSearch() { query.pageIndex = 1; loadData(); }
function onReset() { Object.assign(query, { gameId: "", orderNo: "", status: "", pageIndex: 1 }); loadData(); }

async function onRefund(row: any) {
  const res = await request("POST", `${API_BASE}/${row.id}/refund`);
  if (res.code === 200) {
    ElMessage.success("退款成功");
    loadData();
  } else {
    ElMessage.error(res.msg || "退款失败");
  }
}

// 导出 CSV
function onExport() {
  const headers = ["ID", "订单号", "游戏ID", "玩家ID", "商品名称", "金额(元)", "货币", "渠道", "状态", "支付时间"];
  const rows = list.value.map(r => [
    r.id,
    r.orderNo,
    r.gameId,
    r.playerId,
    `"${(r.productName || "").replace(/"/g, '""')}"`,
    (r.amount / 100).toFixed(2),
    r.currency,
    r.channel,
    statusLabel(r.status),
    r.paidAt || ""
  ]);
  const csv = [headers.join(","), ...rows.map(r => r.join(","))].join("\n");
  const blob = new Blob(["\uFEFF" + csv], { type: "text/csv;charset=utf-8;" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  const now = new Date();
  const dateStr = `${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, "0")}${String(now.getDate()).padStart(2, "0")}`;
  a.href = url;
  a.download = `orders_${dateStr}.csv`;
  a.click();
  URL.revokeObjectURL(url);
}

onMounted(loadData);
</script>
