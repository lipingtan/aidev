# Ontology 驱动的 AI 系统落地流程设计

> 基于 Palantir Foundry 的本体论策略，提炼可具体执行的落地方案

---

## 一、核心理念提炼

Palantir 的本质洞察：企业数据基础设施的失败根源不在于"数据存不了"，而在于"数据只能看不能动"。

**传统模式**：数据源 → 数据湖/仓库 → BI 看板 → 人类看完后手动执行 → 又回到业务系统手动录入

**本体论模式**：数据源 → 本体层（语义+操作） → AI/应用直接执行业务动作 → 结果写回本体层

关键转变：把数据从"静态快照"变为"可操作的业务实体"。

---

## 二、落地流程总览（5 阶段）

```
阶段 1：业务建模        → 识别核心实体和操作（1-2周）
阶段 2：本体层构建      → 定义对象/关系/规则/操作（2-3周）
阶段 3：数据管道接入    → 打通数据源到本体层的索引（2-4周）
阶段 4：治理体系搭建    → 分支/审批/权限/审计日志（1-2周）
阶段 5：AI Agent 集成   → 让 AI 基于本体层安全执行（2-4周）
```

---

## 三、阶段 1：业务建模（识别"名词"和"动词"）

### 目标

把目标业务领域拆解为 Palantir 定义的两类元素：
- **语义元素（名词）**：Object Type、Property、Link Type
- **动力元素（动词）**：Action Type、Function、Dynamic Security

### 执行步骤

#### Step 1.1：业务领域圈定

选择一个边界清晰、数据可获取的业务子域作为试点。

选择标准：
- 数据源不超过 3-5 个系统
- 有明确的"人看完数据后手动操作"环节（即可以被自动化的操作闭环）
- 涉及的业务对象不超过 10-15 个

#### Step 1.2：实体识别工作坊

与业务专家进行 2-4 小时的结构化访谈：

| 问题 | 提取目标 |
|------|----------|
| "你们日常操作的核心'东西'有哪些？" | Object Type 候选 |
| "这些东西有哪些关键属性？" | Property 定义 |
| "它们之间什么关系？" | Link Type |
| "你日常对这些东西做什么操作？" | Action Type 候选 |
| "什么条件下做什么操作？" | Business Rule |
| "谁能做、谁不能做？" | Dynamic Security |

#### Step 1.3：产出物

```yaml
# 业务建模产出：domain_model.yaml
domain: "目标业务域名称"
objects:
  - name: EntityName
    display_name: "中文名"
    description: "业务含义"
    properties: [...]
    
links:
  - from: EntityA
    to: EntityB
    type: one_to_many
    description: "关系含义"

actions:
  - name: action_name
    description: "操作含义"
    modifies: [EntityName]
    preconditions: ["前置条件"]
    side_effects: ["副作用"]

rules:
  - scope: EntityName
    condition: "触发条件"
    action: "应该做什么"
```

---

## 四、阶段 2：本体层构建

### 目标

将业务模型转化为可执行的技术定义，让 AI 和应用能直接使用。

### 执行步骤

#### Step 2.1：对象类型定义

为每个实体编写完整的类型定义文件。关键是每个属性必须包含**业务语义描述**（给 AI 理解用）而非仅仅是技术类型。

```yaml
# ontology/objects/order.yaml
object_type:
  name: PurchaseOrder
  display_name: 采购订单
  description: "向供应商发出的物料采购请求，包含多个订单行"
  primary_key: order_id
  data_source: "erp.purchase_orders"
  
  properties:
    - name: order_id
      type: string
      description: "订单唯一编号，格式 PO-{年月}-{序号}"
      
    - name: status
      type: enum[draft, submitted, approved, shipped, received, closed]
      description: |
        订单状态流转：
        draft → submitted（提交审批）
        submitted → approved（审批通过）
        approved → shipped（供应商发货）
        shipped → received（仓库收货）
        received → closed（完成入库）
      constraints: "只能按顺序流转，不可跳跃或回退"
      
    - name: total_amount
      type: number
      unit: "元"
      description: "订单总金额，所有订单行金额之和"
      computed: true
```

#### Step 2.2：关系定义

显式定义所有对象间的关系，包含语义方向和业务含义。

```yaml
# ontology/relations.yaml
relations:
  - name: order_contains_lines
    from: PurchaseOrder
    to: OrderLine
    cardinality: one_to_many
    ownership: true  # 父子关系，删除订单则删除行
    description: "一个订单包含多个订单行（物料明细）"
    
  - name: order_from_supplier
    from: PurchaseOrder
    to: Supplier
    cardinality: many_to_one
    description: "每个订单对应一个供应商"
    reverse_description: "一个供应商可以有多个订单"
```

#### Step 2.3：操作定义（Action Type）

这是 Palantir 本体论的核心差异——把"能做什么"也定义到数据模型中。

```yaml
# ontology/actions/approve_order.yaml
action_type:
  name: approve_purchase_order
  display_name: 审批采购订单
  description: "审批通过一个采购订单，使其可以发送给供应商"
  
  # 前置条件
  preconditions:
    - "订单状态必须为 submitted"
    - "审批人角色必须为 procurement_manager 或 finance_approver"
    - "订单金额 <= 审批人审批额度"
    
  # 修改的对象和属性
  modifications:
    - object: PurchaseOrder
      property: status
      change: "submitted → approved"
    - object: PurchaseOrder
      property: approved_by
      change: "设为当前操作人"
    - object: PurchaseOrder
      property: approved_at
      change: "设为当前时间"
      
  # 副作用
  side_effects:
    - "发送通知给供应商联系人"
    - "更新采购预算已用金额"
    - "生成审批日志记录"
    
  # 权限要求
  required_permissions:
    - "procurement:order:approve"
    
  # 失败回滚
  rollback: "不修改任何数据，返回错误原因"
```

#### Step 2.4：业务规则定义

```yaml
# ontology/rules/procurement_rules.yaml
rules:
  - name: auto_escalation
    scope: PurchaseOrder
    trigger: "status=submitted 且创建时间超过 48 小时未审批"
    action: "自动升级通知给部门总监"
    severity: warning
    
  - name: budget_check
    scope: PurchaseOrder
    trigger: "订单金额超过部门季度剩余预算的 80%"
    action: "标记为预算预警，审批时需要额外确认"
    severity: info
    
  - name: supplier_risk
    scope: Supplier
    trigger: "最近 90 天交付准时率 < 70%"
    action: "标记为高风险供应商，新订单需要二级审批"
    severity: critical
```

---

## 五、阶段 3：数据管道接入

### 目标

打通数据源到本体层的管道，让本体层对象实时反映业务真实状态。

### 执行步骤

#### Step 3.1：数据源清单与映射

```yaml
# pipeline/source_mapping.yaml
mappings:
  - object_type: PurchaseOrder
    data_source:
      type: database
      connection: "erp_postgres"
      table: "purchase_orders"
    field_mapping:
      order_id: "po_number"
      status: "order_status"  # 需要值映射：0=draft, 1=submitted, ...
      total_amount: "total_amt"
    refresh: 
      mode: streaming  # 或 batch
      interval: "实时"  # batch 模式下填 "每5分钟" / "每小时"
      
  - object_type: Supplier
    data_source:
      type: api
      endpoint: "https://internal.srm.com/api/suppliers"
    refresh:
      mode: batch
      interval: "每日 02:00"
```

#### Step 3.2：数据索引管道实现

两种模式按业务时效性选择：

| 模式 | 延迟 | 适用场景 |
|------|------|----------|
| Batch | 分钟~小时级 | 主数据同步、历史数据加载、报表数据 |
| Streaming | 秒~分钟级 | 订单状态变更、库存扣减、告警触发 |

```python
# pipeline/indexer.py（概念代码）
class OntologyIndexer:
    """将数据源变更索引到本体层"""
    
    def index_batch(self, object_type: str, source_data: list):
        """批量索引：全量或增量同步"""
        for record in source_data:
            mapped = self.apply_field_mapping(object_type, record)
            validated = self.validate_against_schema(object_type, mapped)
            self.upsert_object(object_type, validated)
            
    def index_stream(self, object_type: str, change_event: dict):
        """流式索引：监听变更事件实时更新"""
        mapped = self.apply_field_mapping(object_type, change_event["data"])
        validated = self.validate_against_schema(object_type, mapped)
        self.upsert_object(object_type, validated)
        
        # 检查是否触发业务规则
        triggered_rules = self.evaluate_rules(object_type, validated)
        for rule in triggered_rules:
            self.execute_rule_action(rule, validated)
```

#### Step 3.3：数据质量验证

索引前后必须验证：

| 检查项 | 方法 |
|--------|------|
| 完整性 | 源系统记录数 vs 本体层对象数，偏差 < 1% |
| 时效性 | 源系统变更到本体层可查询的延迟在 SLA 内 |
| 准确性 | 抽样 100 条对比源数据和本体层数据一致性 |
| 关系完整性 | 所有 Link 引用的目标对象都存在 |

---

## 六、阶段 4：治理体系搭建

### 目标

确保本体层的变更安全可控、可追溯，对标 Palantir 的 Branch + Proposal + Approval 机制。

### 执行步骤

#### Step 4.1：变更分支机制

任何对本体层定义（Schema）的修改都必须走分支流程：

```
1. 创建工作分支（从 main 派生）
2. 在工作分支上修改本体定义
3. 自动化验证（Schema 校验、影响分析）
4. 提交 Proposal（描述变更目的和影响）
5. 指定审批人 Review
6. 审批通过 → 合并到 main → 立即生效
7. 审批拒绝 → 打回修改
```

落地实现（用 Git 仓库模拟）：

```
ontology-repo/
├── main branch          # 生产环境的本体定义
├── feature/add-xxx      # 新增对象类型
├── fix/update-rules     # 修改业务规则
└── .github/workflows/   # CI：自动校验 + 影响分析
```

#### Step 4.2：操作权限模型

```yaml
# governance/permissions.yaml
roles:
  - name: procurement_staff
    description: "采购专员"
    can_read: [PurchaseOrder, Supplier, OrderLine]
    can_execute: [create_draft_order, submit_order]
    cannot_execute: [approve_purchase_order, delete_order]
    
  - name: procurement_manager
    description: "采购经理"
    inherits: procurement_staff
    can_execute: [approve_purchase_order]
    approval_limit: 500000  # 50 万以内可审批
    
  - name: finance_approver
    description: "财务审批人"
    can_read: [PurchaseOrder]
    can_execute: [approve_purchase_order]
    approval_limit: 2000000  # 200 万以内

# 行列级数据权限
data_policies:
  - object: Supplier
    property: bank_account
    visible_to: [finance_approver]
    hidden_from: [procurement_staff]  # 采购看不到供应商银行账号
```

#### Step 4.3：审计日志（Action Log）

每个 Action 执行后自动生成不可篡改的日志对象：

```yaml
# 审计日志结构
action_log:
  log_id: "自动生成"
  action_type: "approve_purchase_order"
  operator: "张三 (procurement_manager)"
  timestamp: "2025-07-14T10:30:00+08:00"
  target_objects:
    - type: PurchaseOrder
      id: "PO-202507-0042"
  changes:
    - property: status
      before: "submitted"
      after: "approved"
    - property: approved_by
      before: null
      after: "张三"
  context:
    ip: "192.168.1.100"
    approval_comment: "金额合理，供应商资质已验证"
```

---

## 七、阶段 5：AI Agent 集成

### 目标

让 AI 基于本体层理解业务语义、执行受限操作、产出可靠结果。

### 核心设计原则（Palantir 的安全哲学）

1. **AI 只能在已定义的操作范围内行动**——不能发明新操作
2. **AI 继承操作人的权限**——不能越权
3. **写操作必须产生 Proposal**——人类审批后才执行
4. **所有 AI 行为产生审计日志**——可追溯

### 执行步骤

#### Step 5.1：AI 可用工具生成

从本体层的 Action Type 自动生成 AI 可调用的 Tool Schema：

```python
# ai/tool_generator.py
def generate_ai_tools(ontology_actions: list, user_permissions: list) -> list:
    """
    根据本体层操作定义 + 当前用户权限，
    动态生成 AI 可用的 Tool 列表（权限过滤）
    """
    tools = []
    for action in ontology_actions:
        # 权限过滤：AI 只能看到用户有权执行的操作
        if action["required_permissions"] not in user_permissions:
            continue
            
        tool = {
            "type": "function",
            "function": {
                "name": action["name"],
                "description": action["description"],
                # 将 preconditions 也写入描述，让 AI 知道约束
                "parameters": build_params_from_action(action)
            }
        }
        tools.append(tool)
    return tools
```

#### Step 5.2：AI 执行流程（读操作 vs 写操作）

```
用户自然语言请求
    ↓
AI 解析意图，匹配本体层中的 Object/Action
    ↓
┌─────────────────────────────────────┐
│ 读操作（查询）                        │
│ → 直接执行，返回结果                   │
│ → 结果基于本体层语义进行解读            │
└─────────────────────────────────────┘
┌─────────────────────────────────────┐
│ 写操作（创建/修改/删除）               │
│ → AI 生成 Action Proposal            │
│ → 展示给用户确认（变更内容 + 影响范围） │
│ → 用户批准 → 执行 Action              │
│ → 用户拒绝 → 取消                     │
│ → 生成 Action Log                    │
└─────────────────────────────────────┘
```

#### Step 5.3：System Prompt 构建策略

```python
def build_ai_context(user_query: str, user_role: str) -> str:
    """构建 AI 上下文：本体层语义 + 用户权限 + 检索增强"""
    
    # 1. 基础本体信息（对象定义 + 关系）
    relevant_objects = retrieve_relevant_ontology(user_query)
    
    # 2. 适用的业务规则
    relevant_rules = retrieve_relevant_rules(user_query)
    
    # 3. 用户权限边界
    permissions = get_user_permissions(user_role)
    
    context = f"""
## 业务本体（你的认知边界）
{relevant_objects}

## 适用的业务规则
{relevant_rules}

## 你的操作权限
你当前代表角色：{user_role}
可执行操作：{permissions['can_execute']}
禁止操作：{permissions['cannot_execute']}

## 安全准则
- 写操作必须先展示 Proposal 让用户确认
- 不确定时询问用户，不要猜测
- 数据查询结果中的敏感字段已按权限过滤
"""
    return context
```

#### Step 5.4：幻觉防护机制

Palantir 通过本体层约束 AI 的"想象空间"：

| 防护层 | 机制 |
|--------|------|
| 语义约束 | AI 只能引用本体层已定义的对象和属性 |
| 操作约束 | AI 只能调用已定义的 Action，不能自创操作 |
| 值域约束 | enum 类型属性限制了可选值范围 |
| 规则约束 | AI 推理必须基于已定义的业务规则 |
| 权限约束 | AI 看不到也执行不了超权限的内容 |
| 审批约束 | 写操作产生 Proposal，人类兜底 |

---

## 八、项目交付物清单

每个使用本方法论的项目应产出以下制品：

```
project_ontology/
├── README.md                    # 本体层概述和使用说明
├── domain_model.yaml            # 阶段1：业务建模结果
├── objects/                     # 阶段2：对象类型定义
│   ├── entity_a.yaml
│   └── entity_b.yaml
├── relations.yaml               # 阶段2：关系定义
├── actions/                     # 阶段2：操作定义
│   ├── action_x.yaml
│   └── action_y.yaml
├── rules/                       # 阶段2：业务规则
│   └── domain_rules.yaml
├── pipeline/                    # 阶段3：数据管道配置
│   ├── source_mapping.yaml
│   └── indexer_config.yaml
├── governance/                  # 阶段4：治理配置
│   ├── permissions.yaml
│   ├── approval_workflow.yaml
│   └── audit_policy.yaml
├── ai/                          # 阶段5：AI 集成配置
│   ├── tool_schema.json         # 自动生成的 AI Tool 定义
│   ├── system_prompt.md         # AI 上下文模板
│   └── safety_config.yaml       # 安全策略
└── tests/                       # 验证用例
    ├── data_quality_checks.yaml
    └── permission_test_cases.yaml
```

---

## 九、与已有技术栈的对应关系

| Palantir 组件 | 开源/自建等价物 | 说明 |
|---------------|-----------------|------|
| OMS（元数据服务） | YAML 定义文件 + Git 版本管理 | 本体 Schema 存储和版本化 |
| Object Storage V2 | PostgreSQL / MongoDB + Redis 缓存 | 对象实例存储和查询 |
| OSS（对象集服务） | Elasticsearch / 自建查询层 | 搜索、筛选、聚合 |
| Actions 服务 | 自建 API 层（Go/Python） | 操作执行和事务管理 |
| Funnel（数据管道） | Debezium + Kafka / 定时任务 | 数据索引管道 |
| Branching/Proposal | Git PR + CI/CD | 变更审批流 |
| Approvals App | 审批流引擎（自建或集成） | 多级审批 |
| Action Log | 审计日志表（append-only） | 不可篡改的操作记录 |
| AIP | LLM API + Function Calling | AI 执行层 |
| OSDK | 自建 SDK / OpenAPI | 外部应用集成 |

---

## 十、落地优先级建议

**第一优先（本周启动）**：
- 选定试点业务域
- 完成阶段 1 的实体识别

**第二优先（月内完成）**：
- 阶段 2 本体层定义
- 阶段 5 的 AI Tool Schema 生成（让 AI 先能"读"）

**第三优先（季度内完成）**：
- 阶段 3 数据管道
- 阶段 4 治理体系
- 阶段 5 完整的写操作+审批流

**核心原则**：先让 AI 能"看懂"业务（读），再让 AI 能"安全地做"业务（写）。
