<template>
  <div class="main">
    <el-form :inline="true" :model="query" class="search-form bg-bg_color pl-8 pt-4 mb-3">
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
    <div class="flex justify-between mb-3">
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
      class="mt-4 flex justify-end"
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
import { getOrderList, refundOrder } from "@/api/game";
defineOptions({ name: "OrderManage" });

const loading = ref(false);
const list = ref<any[]>([]);
const total = ref(0);
const query = reactive({ gameId: "", orderNo: "", status: "", pageIndex: 1, pageSize: 10 });

const statusTagType = (status: string) => {
  const map: Record<string, string> = { pending: "warning", paid: "success", refunded: "info", failed: "danger" };
  return map[status] || "";
};
const statusLabel = (status: string) => {
  const map: Record<string, string> = { pending: "待支付", paid: "已支付", refunded: "已退款", failed: "失败" };
  return map[status] || status;
};

async function loadData() {
  loading.value = true;
  try {
    const res: any = await getOrderList(query);
    list.value = res.data || [];
    total.value = res.count || 0;
  } finally {
    loading.value = false;
  }
}
function onSearch() { query.pageIndex = 1; loadData(); }
function onReset() { Object.assign(query, { gameId: "", orderNo: "", status: "", pageIndex: 1 }); loadData(); }

async function onRefund(row: any) {
  await refundOrder(row.id);
  ElMessage.success("退款成功");
  loadData();
}

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
