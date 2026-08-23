extends Node
## 全局事件总线（Autoload 名 EventBus，注册顺序第 1）
##
## 纯信号枢纽，无状态：盒子工程跨系统事件的统一发射点。
## 页面/服务只连信号、互不引用；信号表见 design.md §3.1。
##
## 依赖：
## - 无（无状态，不依赖其他单例）

# === 游戏生命周期 ===
## 游戏实例已启动（GameHost 就绪）
signal game_launched(gid: String)
## 游戏实例已结束；result 为结算数据（字段见 GameModule 协议）
signal game_finished(gid: String, result: Dictionary)
## 试用次数消耗一次；left 为剩余次数
signal trial_consumed(gid: String, left: int)

# === 商业与评价 ===
## 支付成功；order_id 为本地订单号
signal payment_succeeded(gid: String, order_id: String)
## 评价已提交（一人一游戏一条 active）
signal review_submitted(gid: String)
## 评价已删除
signal review_deleted(gid: String)

# === 成就与 DLC ===
## 成就解锁；def 为成就定义，points 为本次获得点数
signal achievement_unlocked(def: Dictionary, points: int)
## DLC 包安装完成（触发 Registry.reload()）
signal dlc_installed(gid: String)

# === 盒子 UI ===
## Tab 切换；idx 为 Tab 索引（0-3）
signal tab_changed(idx: int)
## 设置项变更（非主题类，如声音/震动开关）
signal settings_changed(key: StringName, value: Variant)
## 主题切换（neon/elegant）；BgLayer 仅监听此信号
signal theme_changed(theme_name: String)
## 游戏记录数据已更新
signal records_updated(gid: String)
