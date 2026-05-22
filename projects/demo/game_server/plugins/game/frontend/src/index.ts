import GameList from "./views/GameList.vue";
import DlcList from "./views/DlcList.vue";
import PlayerList from "./views/PlayerList.vue";
import OrderList from "./views/OrderList.vue";
import PaymentConfig from "./views/PaymentConfig.vue";
import H5List from "./views/H5List.vue";

export const routes = [
  { path: "/plugin/game/list", name: "PluginGameList", component: GameList, meta: { title: "游戏列表" } },
  { path: "/plugin/game/dlc", name: "PluginGameDlc", component: DlcList, meta: { title: "DLC管理" } },
  { path: "/plugin/game/player", name: "PluginGamePlayer", component: PlayerList, meta: { title: "玩家管理" } },
  { path: "/plugin/game/order", name: "PluginGameOrder", component: OrderList, meta: { title: "订单管理" } },
  { path: "/plugin/game/payment", name: "PluginGamePayment", component: PaymentConfig, meta: { title: "支付配置" } },
  { path: "/plugin/game/h5", name: "PluginGameH5", component: H5List, meta: { title: "H5页面" } },
];

export const menus = [];
