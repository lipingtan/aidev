# 需求：DLC 管理功能完善

## 背景

当前 DLC 管理只有基础 CRUD，缺少文件上传/下载、游戏关联选择、业务校验等实用功能，需要补全以达到可用状态。

## 用户故事

- 作为运营人员，我希望在新增/编辑 DLC 时通过下拉选择游戏，以便避免手动输入 ID 出错
- 作为运营人员，我希望上传 PCK 文件并查看文件信息，以便管理 DLC 资源包
- 作为运营人员，我希望下载已上传的 PCK 文件，以便验证内容正确性
- 作为运营人员，我希望查看 DLC 的下载趋势和收入统计，以便评估运营效果
- 作为系统，我希望上架前校验 PCK 文件已上传，以便防止空包上架
- 作为系统，我希望阻止删除已有玩家购买的 DLC，以便保护用户权益

## 功能需求

### FR-1: 游戏下拉选择

**描述：** DLC 表单中 gameId 字段改为下拉选择已有游戏列表。

**验收标准：**
- WHEN 用户打开新增/编辑 DLC 弹窗 THEN 系统 SHALL 展示游戏下拉列表（调用 getGameList 获取）
- WHEN 下拉列表加载 THEN 系统 SHALL 显示游戏名称，值为游戏 ID

### FR-2: PCK 文件上传

**描述：** 在 DLC 编辑页面提供文件上传功能，上传后自动计算 SHA256 和文件大小。

**验收标准：**
- WHEN 用户在 DLC 详情/编辑页点击"上传 PCK" THEN 系统 SHALL 弹出文件选择器，限制 `.pck` 后缀
- WHEN 文件上传成功 THEN 系统 SHALL 更新 filePath、fileSize、sha256 字段并刷新列表
- WHEN 文件大小超过 500MB THEN 系统 SHALL 拒绝上传并提示"文件超过 500MB 限制"
- WHEN 存储层 THEN 系统 SHALL 通过 FileStore 接口存储，当前实现为本地磁盘，后续可替换为 S3/OSS

### FR-3: PCK 文件下载

**描述：** 管理端提供 DLC 文件下载链接（暂不鉴权）。

**验收标准：**
- WHEN DLC 已上传 PCK 文件 THEN 系统 SHALL 在列表操作列显示"下载"按钮
- WHEN 用户点击"下载" THEN 系统 SHALL 触发浏览器下载对应 PCK 文件
- WHEN DLC 未上传 PCK 文件 THEN 系统 SHALL 不显示"下载"按钮

### FR-4: 文件信息展示

**描述：** 列表和详情中展示 PCK 文件相关信息。

**验收标准：**
- WHEN DLC 列表加载 THEN 系统 SHALL 展示文件大小（格式化为 KB/MB）和上传状态
- WHEN 用户查看 DLC 详情 THEN 系统 SHALL 展示 filePath、fileSize、sha256

### FR-5: 免费/付费联动

**描述：** 选择"免费"时价格自动置 0 并禁用输入。

**验收标准：**
- WHEN 用户选择 isFree=1（免费） THEN 系统 SHALL 将 price 置为 0 并禁用价格输入框
- WHEN 用户选择 isFree=2（付费） THEN 系统 SHALL 启用价格输入框

### FR-6: 上架前校验

**描述：** DLC 上架时校验 PCK 文件是否已上传。

**验收标准：**
- WHEN 用户将 DLC 状态改为"上架"且 filePath 为空 THEN 系统 SHALL 拒绝操作并提示"请先上传 PCK 文件"
- WHEN filePath 不为空 THEN 系统 SHALL 允许上架

### FR-7: 删除保护

**描述：** 已有玩家购买的 DLC 不允许删除。

**验收标准：**
- WHEN 用户删除 DLC 且该 DLC 存在关联订单（game_order 中 productType='dlc' AND productId=dlcId） THEN 系统 SHALL 拒绝删除并提示"该 DLC 已有玩家购买，无法删除"
- WHEN 无关联订单 THEN 系统 SHALL 正常执行删除

### FR-8: DLC 统计面板

**描述：** 提供 DLC 的下载趋势和收入统计。

**验收标准：**
- WHEN 用户访问 DLC 统计页 THEN 系统 SHALL 展示：总下载次数、总收入、近 30 天下载趋势折线图、各 DLC 收入排行
- WHEN 用户按游戏筛选 THEN 系统 SHALL 过滤对应游戏的统计数据

## 非功能需求

- 文件大小限制：单个 PCK 文件最大 500MB
- 存储抽象：通过 FileStore 接口封装存储操作，当前实现本地磁盘，后续替换 S3/OSS 时只需实现新的 FileStore
- 不支持断点续传/分片上传（后续优化）
- 只保留最新版本，不做版本历史
