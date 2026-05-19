# Ontology 驱动的 AI 决策系统 — 设计方法论与落地实施

## 企业背景：防水建材生产与销售企业

以东方雨虹/科顺股份类型的防水建材企业为原型：
- 产品线：防水卷材、防水涂料、密封胶、保温材料
- 业务链：原材料采购 → 生产制造 → 仓储物流 → 经销/直销 → 工程施工 → 售后质保
- 客户类型：房地产开发商、基建工程方、经销商、零售终端
- 痛点：产销协同难、工程回款慢、质量追溯链断裂、经销商管理粗放

---

# 一、方法论框架

## 1.1 整体分层架构

```
┌──────────────────────────────────────────────────┐
│  Layer 5: 用户交互层                              │  自然语言对话 / 看板 / 审批界面
├──────────────────────────────────────────────────┤
│  Layer 4: AI 决策层                               │  LLM 推理 + Tool Calling + 决策链
├──────────────────────────────────────────────────┤
│  Layer 3: 本体语义层（Ontology）                   │  对象/关系/属性/Action/规则
├──────────────────────────────────────────────────┤
│  Layer 2: 数据集成层                              │  多源接入 + 清洗 + 统一映射
├──────────────────────────────────────────────────┤
│  Layer 1: 业务系统层                              │  ERP/MES/WMS/CRM/项目管理/财务
└──────────────────────────────────────────────────┘
```

## 1.2 六步设计法

| 步骤 | 名称 | 核心产出 | 关键方法 |
|------|------|----------|----------|
| Step 1 | 业务场景锚定 | 优先场景清单（≤3个） | 决策密度×数据就绪度矩阵 |
| Step 2 | 核心对象识别 | Object Type 清单 | 名词/动词/形容词拆解法 |
| Step 3 | 对象关系建模 | Ontology Graph | 关系类型定义 + 约束规则 |
| Step 4 | 数据源映射 | 字段映射表 + ETL 规则 | 源系统盘点 + 语义对齐 |
| Step 5 | Action 定义 | 可执行操作清单 | 写回接口 + 权限 + 审批流 |
| Step 6 | AI 接入验证 | Prompt 模板 + 验证报告 | 上下文注入 + 人工校验循环 |

## 1.3 场景优先级评估矩阵

选场景的标准不是"哪个最大"，而是"哪个决策密度高 + 数据已有"：

```
                    数据就绪度（高）
                         │
         ┌───────────────┼───────────────┐
         │   ★ 首期落地   │   理想但需验证  │
         │  （产销协同）   │  （工程回款预测）│
决策频率  ├───────────────┼───────────────┤
  （高）  │   需先补数据   │   远期规划     │
         │  （施工质量追溯）│  （新品研发决策）│
         └───────────────┼───────────────┘
                         │
                    数据就绪度（低）
```

---

# 二、防水建材企业完整落地

## Step 1：业务场景锚定

### 场景分析

| 候选场景 | 决策频率 | 决策复杂度 | 数据就绪度 | 业务价值 | 优先级 |
|----------|----------|-----------|-----------|----------|--------|
| 产销协同（排产+备货） | 每日 | 高 | 高（ERP+MES已有） | 降库存积压、减断货 | ★★★ |
| 经销商信用与回款管理 | 每周 | 中 | 中（CRM+财务） | 降坏账、提现金流 | ★★☆ |
| 工程项目报价决策 | 每单 | 高 | 中（历史报价+成本） | 提中标率+利润率 | ★★☆ |
| 原材料采购时机 | 每周 | 中 | 高（采购+行情） | 降采购成本 | ★★☆ |
| 施工质量追溯 | 事件驱动 | 高 | 低（现场数据缺失） | 降质保赔付 | ★☆☆ |

**首期锚定：产销协同决策**

> 目标：AI 每日生成各工厂排产建议 + 各仓库补货计划，综合考虑订单、库存、产能、物流、季节性。

---

## Step 2：核心对象识别

### 名词 → Object Type

| 对象类型 | 英文标识 | 说明 |
|----------|----------|------|
| 工厂 | Factory | 生产基地，含产线信息 |
| 产线 | ProductionLine | 具体生产线，有产能上限和产品适配 |
| 产品 | Product | SKU 级别，如"SBS改性沥青防水卷材-4mm" |
| 产品族 | ProductFamily | 产品大类，如"卷材"、"涂料" |
| 原材料 | RawMaterial | 沥青、聚酯胎基、SBS改性剂等 |
| 仓库 | Warehouse | 成品仓/原材料仓，含地理位置 |
| 库存 | Inventory | 某仓库某产品的当前库存 |
| 客户 | Customer | 开发商/工程方/经销商 |
| 销售订单 | SalesOrder | 已确认的客户订单 |
| 需求预测 | DemandForecast | 基于历史+季节+项目管道的预测 |
| 生产工单 | ProductionOrder | 排产计划的执行单元 |
| 物流路线 | LogisticsRoute | 工厂→仓库→客户的运输路径 |
| 供应商 | Supplier | 原材料供应商 |

### 动词 → Action

| 操作 | 写回系统 | 审批要求 |
|------|----------|----------|
| 创建生产工单 | MES | 产能超限时需审批 |
| 调整排产优先级 | MES | 自动 |
| 创建补货调拨单 | WMS | 跨区调拨需审批 |
| 创建采购申请 | ERP | 超额需审批 |
| 调整安全库存阈值 | WMS | 自动 |
| 发出缺货预警 | 通知系统 | 自动 |

### 形容词 → 计算属性/状态

| 属性 | 计算逻辑 |
|------|----------|
| 库存健康度 | 当前库存 / 未来14天预测消耗 |
| 产能利用率 | 已排产量 / 产线日产能 |
| 订单紧迫度 | (交期-今天) / 生产周期 |
| 供应风险度 | 原材料可用天数 × 供应商可靠度 |
| 季节性系数 | 基于历史同期销量的波动因子 |

---

## Step 3：对象关系建模

### 关系定义

```
Factory (工厂)
  ├── 拥有产线 → ProductionLine (1:N)
  ├── 配套仓库 → Warehouse (1:N, 含原材料仓+成品仓)
  └── 所属区域 → Region (N:1)

ProductionLine (产线)
  ├── 可生产产品 → Product (M:N, 含切换时间/日产能)
  ├── 当前工单 → ProductionOrder (1:N)
  └── 消耗原材料 → RawMaterial (M:N, 含BOM用量)

Product (产品)
  ├── 属于产品族 → ProductFamily (N:1)
  ├── BOM组成 → RawMaterial (M:N, 含用量配比)
  ├── 存放于 → Inventory (1:N, 每仓一条)
  └── 被订购 → SalesOrder (1:N)

Customer (客户)
  ├── 下单 → SalesOrder (1:N)
  ├── 所属区域 → Region (N:1)
  ├── 信用等级 → 属性
  └── 历史回款率 → 计算属性

SalesOrder (销售订单)
  ├── 包含产品 → Product (M:N, 含数量)
  ├── 交付仓库 → Warehouse (N:1)
  ├── 所属客户 → Customer (N:1)
  └── 状态 → 待排产/已排产/生产中/待发货/已发货

Warehouse (仓库)
  ├── 库存明细 → Inventory (1:N)
  ├── 物流路线 → LogisticsRoute (1:N)
  └── 覆盖区域 → Region (M:N)

Inventory (库存)
  ├── 所属仓库 → Warehouse (N:1)
  ├── 对应产品 → Product (N:1)
  └── 计算属性: 健康度、可用天数、补货紧迫度
```

### 关系图（核心链路）

```
DemandForecast ──预测──→ Product 需求量
         ↓
SalesOrder ──消耗──→ Inventory（成品）
         ↓                    ↓
   需要排产              触发补货/调拨
         ↓                    ↓
ProductionOrder ──占用──→ ProductionLine
         ↓
   消耗 RawMaterial ──触发──→ 采购申请
```

---

## Step 4：数据源映射

### 源系统盘点

| 源系统 | 提供的对象 | 接入方式 | 更新频率 |
|--------|-----------|----------|----------|
| ERP（SAP/用友） | 产品主数据、BOM、采购单、销售订单 | API/数据库直连 | 实时 |
| MES | 产线状态、工单进度、产能数据 | API | 每小时 |
| WMS | 库存明细、出入库记录 | API | 实时 |
| CRM | 客户信息、项目管道、回款记录 | API | 每日 |
| TMS（物流） | 运输路线、在途库存、运费 | API | 每日 |
| 外部数据 | 沥青价格行情、天气（影响施工季） | 爬虫/API | 每日 |

### 字段映射示例

```yaml
# Inventory 对象映射
source_system: WMS
source_table: wms_stock_detail
mapping:
  store_code     → warehouse_id    # 仓库编码
  material_code  → product_id      # 物料编码（需与ERP产品主数据关联）
  available_qty  → quantity         # 可用数量
  reserved_qty   → reserved        # 已预留数量（被订单锁定）
  uom            → unit            # 单位（卷/桶/吨）
  last_count_dt  → last_updated    # 最后盘点时间

# 计算属性（不来自源系统，由 Ontology 层计算）
computed:
  net_available: quantity - reserved
  days_of_supply: net_available / forecast_daily_demand(product_id, warehouse_id)
  health_status: 
    CRITICAL if days_of_supply < 3
    LOW if days_of_supply < 7
    NORMAL if days_of_supply < 21
    EXCESS if days_of_supply > 45
```

---

## Step 5：Action 定义

### Action 1：创建生产工单

```json
{
  "name": "create_production_order",
  "description": "在指定工厂的指定产线上创建生产工单",
  "parameters": {
    "factory_id": {"type": "string", "description": "工厂ID"},
    "line_id": {"type": "string", "description": "产线ID"},
    "product_id": {"type": "string", "description": "产品SKU"},
    "quantity": {"type": "number", "description": "生产数量（基本单位）"},
    "priority": {"type": "string", "enum": ["urgent", "normal", "low"]},
    "target_date": {"type": "string", "description": "期望完成日期"}
  },
  "constraints": [
    "产线必须支持该产品",
    "排产量不得超过产线日产能的120%",
    "原材料库存必须满足BOM需求，否则同时触发采购"
  ],
  "write_back": "MES.production_orders",
  "approval_rule": "priority=urgent 或 产能利用率>90% 时需主管审批"
}
```

### Action 2：创建仓库间调拨

```json
{
  "name": "create_transfer_order",
  "description": "从库存充足的仓库调拨成品到紧缺仓库",
  "parameters": {
    "from_warehouse_id": {"type": "string"},
    "to_warehouse_id": {"type": "string"},
    "product_id": {"type": "string"},
    "quantity": {"type": "number"},
    "reason": {"type": "string", "description": "调拨原因说明"}
  },
  "constraints": [
    "源仓库调出后库存健康度不得低于 NORMAL",
    "优先选择物流成本最低的路线",
    "跨大区调拨需区域总监审批"
  ],
  "write_back": "WMS.transfer_orders + TMS.shipment_request"
}
```

### Action 3：调整安全库存

```json
{
  "name": "adjust_safety_stock",
  "description": "根据季节性和需求变化调整某仓库某产品的安全库存水位",
  "parameters": {
    "warehouse_id": {"type": "string"},
    "product_id": {"type": "string"},
    "new_safety_stock": {"type": "number"},
    "reason": {"type": "string"}
  },
  "constraints": [
    "调整幅度不得超过当前值的±50%",
    "旺季（3-10月）允许上调，淡季允许下调"
  ],
  "write_back": "WMS.safety_stock_config"
}
```

### Action 4：发出预警通知

```json
{
  "name": "send_alert",
  "description": "向相关人员发送业务预警",
  "parameters": {
    "alert_type": {"type": "string", "enum": ["stockout_risk", "overstock", "capacity_bottleneck", "material_shortage"]},
    "severity": {"type": "string", "enum": ["info", "warning", "critical"]},
    "target_roles": {"type": "array", "items": {"type": "string"}},
    "message": {"type": "string"},
    "related_objects": {"type": "array", "description": "关联的 Ontology 对象ID"}
  },
  "write_back": "notification_service"
}
```

---

## Step 6：AI 接入与验证

### Prompt 模板设计

```python
def build_production_planning_prompt(date: str) -> tuple[str, str]:
    """构建产销协同决策的 Prompt"""
    
    # 1. 查询 Ontology 获取当前状态
    critical_inventory = InventoryObject.query(health_status="CRITICAL")
    low_inventory = InventoryObject.query(health_status="LOW")
    pending_orders = SalesOrderObject.query(status="待排产", sort="urgency DESC")
    factory_status = FactoryObject.query_all_with_capacity()
    material_risks = RawMaterialObject.query(supply_risk="HIGH")
    season_factor = SeasonService.get_factor(date)
    
    # 2. 组装业务上下文
    context = f"""
## 日期与季节
- 当前日期：{date}
- 季节性系数：{season_factor}（>1.0 表示旺季，防水施工高峰）

## 成品库存告警
### CRITICAL（<3天可用）
{format_inventory_table(critical_inventory)}

### LOW（3-7天可用）
{format_inventory_table(low_inventory)}

## 待排产订单（按紧迫度排序）
{format_orders_table(pending_orders[:20])}

## 工厂产能状态
{format_factory_capacity(factory_status)}

## 原材料供应风险
{format_material_risks(material_risks)}

## 可用操作
- create_production_order: 创建生产工单
- create_transfer_order: 仓库间调拨
- adjust_safety_stock: 调整安全库存
- send_alert: 发送预警通知
"""
    
    system_prompt = """你是防水建材企业的产销协同 AI 助手。

职责：根据当前库存状态、待排产订单、工厂产能和原材料供应情况，生成今日产销决策建议。

决策原则：
1. CRITICAL 库存必须当日安排生产或调拨
2. 排产优先级：紧急订单 > 低库存补货 > 安全库存回补
3. 产线切换有成本，同类产品尽量连续排产
4. 旺季（季节系数>1.2）时主动提高安全库存
5. 原材料有风险时，优先排产该材料相关产品（消耗库存避免过期）

输出格式：
1. 先给出决策摘要（3-5句话概括今日重点）
2. 然后逐条列出需要执行的 Action，每条说明决策依据
"""
    
    return system_prompt, context
```

### AI 输出示例

```
## 决策摘要

今日处于施工旺季（季节系数1.35），华东区SBS卷材库存告急（仅剩2.1天），
需紧急安排岳阳工厂2号线排产。同时武汉仓涂料积压（>60天），建议调拨至
南京仓消化。沥青供应商本周报价上涨8%，建议加速消耗当前沥青库存。

## 执行操作

### 1. 紧急排产 - SBS改性沥青卷材4mm
create_production_order(
  factory_id="F003_岳阳",
  line_id="L002_卷材线",
  product_id="SBS-4mm-001",
  quantity=5000,  # 卷
  priority="urgent",
  target_date="2025-04-18"
)
依据：华东区库存仅2.1天，有3笔订单4月20日前交付，合计需4200卷。

### 2. 库存调拨 - JS防水涂料
create_transfer_order(
  from_warehouse_id="WH_武汉",
  to_warehouse_id="WH_南京",
  product_id="JS-COAT-002",
  quantity=800,  # 桶
  reason="武汉仓积压60天+，南京仓仅剩8天，调拨平衡"
)
依据：武汉仓该产品65天库存，南京仓仅8天且有在手订单。

### 3. 调整安全库存 - 旺季上调
adjust_safety_stock(
  warehouse_id="WH_上海",
  product_id="SBS-4mm-001",
  new_safety_stock=3000,  # 从2000上调
  reason="进入施工旺季，季节系数1.35，历史同期日均消耗量+40%"
)

### 4. 预警 - 沥青价格风险
send_alert(
  alert_type="material_shortage",
  severity="warning",
  target_roles=["采购经理", "生产总监"],
  message="沥青现货价格本周上涨8%，当前库存可支撑12天生产。建议评估是否提前锁价采购。",
  related_objects=["RM_沥青_70号", "F003_岳阳", "F001_深圳"]
)
```

---

# 三、实施路线图

## 阶段划分

```
Phase 0: 准备期（2周）
├── 组建团队：1产品+2后端+1数据工程+1业务顾问
├── 源系统权限申请
└── 确认首期场景和成功标准

Phase 1: 数据基座（4周）
├── Week 1-2: 源系统数据接入（ERP/MES/WMS）
├── Week 3: 数据清洗规则 + 字段映射
└── Week 4: Ontology 对象模型实现 + 计算属性

Phase 2: 语义服务（3周）
├── Week 5: Ontology 查询 API 开发
├── Week 6: Action 接口开发（对接 MES/WMS）
└── Week 7: 权限/审批流集成

Phase 3: AI 接入（3周）
├── Week 8: Prompt 模板设计 + LLM 选型接入
├── Week 9: 3家工厂试点，人工审核模式
└── Week 10: 反馈收集 + Prompt/模型调优

Phase 4: 规模化（4周）
├── Week 11-12: 全量工厂/仓库接入
├── Week 13: 低风险操作自动执行
└── Week 14: 接入第二个场景（经销商信用管理）
```

## 成功标准

| 指标 | 基线 | 目标 |
|------|------|------|
| 成品断货率 | 8% | <3% |
| 库存周转天数 | 45天 | 35天 |
| 排产计划制定耗时 | 4小时/天（人工） | 15分钟/天（AI+审核） |
| 产线切换次数 | 12次/周 | 8次/周 |

---

# 四、技术选型建议

| 层级 | 推荐方案 | 备选 |
|------|----------|------|
| 数据集成 | Apache Airflow + dbt | Flink（实时场景） |
| Ontology 存储 | PostgreSQL + Neo4j（关系图） | 纯 PG + JSONB |
| Ontology API | Go/Python FastAPI | GraphQL |
| AI 推理 | Claude/GPT-4 + Function Calling | 本地部署 Qwen2.5 |
| Action 执行 | 消息队列(RabbitMQ) + 幂等写回 | 直接 API 调用 |
| 前端交互 | React + 审批流组件 | 企业微信/钉钉集成 |

---

# 五、关键风险与应对

| 风险 | 影响 | 应对 |
|------|------|------|
| 源系统数据质量差 | Ontology 计算属性不准 | Phase 1 加入数据质量看板，设置可信度标记 |
| AI 幻觉导致错误决策 | 排产错误/库存异常 | 首期全部人工审核，逐步放开自动执行 |
| 业务部门不信任 AI | 推广受阻 | 先做"建议模式"而非"自动模式"，用数据证明价值 |
| 系统接口不稳定 | Action 写回失败 | 消息队列 + 重试 + 补偿机制 |
| 跨部门数据权限 | 无法获取完整上下文 | 高层推动 + Ontology 层内置行级权限控制 |
