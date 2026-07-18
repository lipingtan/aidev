# 设计计划：DLC 管理功能完善

## 设计范围

基于 requirements.md 中的 FR-1 ~ FR-8，设计前后端改动方案。

## 澄清问题

[Question] Q1: FileStore 接口放在 `common/file_store/` 还是 `common/storage/`（已有 storage 目录）？
[Answer]
放common/storage/
[Question] Q2: DLC 统计面板是作为 dlc.vue 的子 Tab，还是独立页面（需要新增菜单）？
[Answer]
子tab
[Question] Q3: PCK 上传是在 DLC 列表的操作列点击"上传"，还是在编辑弹窗内嵌上传组件？
[Answer]
应该是新建和编辑都需要吧
[Question] Q4: 下载接口是直接返回文件流（浏览器直接下载），还是返回一个临时下载 URL？
[Answer]
按可扩展的方式实现，因为后面要存到OSS里面
[Question] Q5: 统计接口的"近 30 天下载趋势"数据来源——是从 game_order 表按日聚合，还是需要新增一张下载记录表？
[Answer]
先简化，直接聪game_order表聚合
