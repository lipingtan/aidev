# 本体层落地完整指南：给 AI Agent 提供什么 & 怎么实现

---

## 一、核心答案

AI Agent 需要的不是数据库 Schema，而是一份**业务语义描述文件**——告诉它"这个世界里有什么东西、它们是什么意思、它们之间怎么关联、能对它们做什么操作"。

---

## 二、本体层与 DDD 聚合根的关系

### 结论：本体层比 DDD 聚合根更宽

两者有关系，但解决的是不同层面的问题：

| 维度 | DDD 聚合根 | 本体层（Ontology） |
|------|-----------|-------------------|
| **服务对象** | 代码系统（开发者） | AI Agent + 人 + 代码系统 |
| **关注点** | 数据一致性边界、事务边界 | 语义理解、关系认知、业务规则 |
| **回答的问题** | "这组数据谁负责保护一致性？" | "这个东西是什么？和什么有关？意味着什么？" |
| **粒度** | 严格——聚合内部对外不可见 | 松散——所有实体平等暴露给认知层 |
| **关系表达** | 聚合内用直接引用，聚合间只用 ID | 所有关系显式定义，含语义描述 |

### 具体差异示例

拿"维修工单"场景：

**DDD 视角**：WorkOrder 是聚合根，内部包含 WorkOrderItem（维修项）和 PartConsumption（备件消耗）。外部不能直接操作 WorkOrderItem，必须通过 WorkOrder 的方法。Equipment 是另一个聚合根，WorkOrder 只持有 equipment_id 引用。

**本体层视角**：WorkOrder、Equipment、SparePart、Technician 都是平等的一级对象。AI 需要知道它们之间的所有关系——WorkOrder 关联哪台 Equipment、消耗了哪些 SparePart、指派给哪个 Technician。AI 不关心"谁是聚合根"，它需要的是完整的语义图谱。

### 包含关系

```
本体层 ⊃ DDD 领域模型 ⊃ 聚合根
```

本体层额外包含的内容（DDD 不覆盖的部分）：

1. **值对象/枚举的业务语义解释**——DDD 代码里只是类型定义，本体层要告诉 AI "这个枚举值在业务中意味着什么"
2. **跨聚合的关系语义**——DDD 只存 ID 引用，本体层要说清"这个 ID 代表什么关系、方向、含义"
3. **业务规则的自然语言表达**——DDD 把规则写成代码/领域服务，本体层要让 AI 能"读懂"规则
4. **操作的意图描述**——DDD 有 Command/Method，本体层要加上"什么场景下该调用、有什么副作用"

### 实际项目中两者配合方式

如果系统已经用 DDD 设计了，本体层不是重新做一套，而是**在 DDD 模型之上加一层语义标注**：

```
DDD 代码层（给机器执行用）
    ↓ 抽象 + 标注语义
本体层（给 AI 理解用）
```

具体映射：
1. 每个聚合根 → 本体层的一个 Object Type
2. 聚合内的关键值对象 → 也暴露为 Object Type（DDD 藏着的，本体层展开给 AI 看）
3. 聚合间的 ID 引用 → 本体层中显式定义关系 + 写明语义
4. Domain Service 中的规则逻辑 → 本体层中用自然语言复述一遍

**简单说**：DDD 是写给编译器和开发者看的领域模型，本体层是写给 AI 看的领域模型。内核相同，表达方式和完整度不同。如果 DDD 做得好，本体层就是把已有设计"翻译"一遍给 AI，成本很低。

---

## 三、你需要提供给 AI Agent 的四层信息

### 第一层：对象类型定义（Object Types）

告诉 AI "这个业务里有哪些核心实体"。

```yaml
# ontology/object_types.yaml
object_types:
  - name: Customer
    display_name: 客户
    description: "与我方签订服务合同的企业法人实体"
    primary_key: customer_id
    properties:
      - name: customer_id
        type: string
        description: "唯一标识符"
      - name: company_name
        type: string
        description: "企业全称（工商注册名）"
      - name: industry
        type: enum[manufacturing, finance, healthcare, education]
        description: "所属行业分类"
      - name: contract_status
        type: enum[active, expired, negotiating]
        description: "当前合同状态"
      - name: annual_revenue
        type: number
        unit: "万元"
        description: "客户年营收规模，用于判断服务等级"

  - name: Project
    display_name: 项目
    description: "为某个客户交付的一个具体工程，有明确起止时间和交付物"
    primary_key: project_id
    properties:
      - name: project_id
        type: string
      - name: name
        type: string
        description: "项目名称"
      - name: stage
        type: enum[requirement, design, development, testing, delivered, maintenance]
        description: "当前所处阶段"
      - name: health_status
        type: enum[green, yellow, red]
        description: "项目健康度。green=正常，yellow=有风险，red=严重延期或客户投诉"
```

**关键区别**：这不是数据库建表语句。`description` 字段是给 AI 看的语义解释，让它理解"这个字段在业务中意味着什么"。

---

### 第二层：关系定义（Relations）

告诉 AI "这些对象之间如何关联"。

```yaml
# ontology/relations.yaml
relations:
  - name: customer_owns_project
    from: Customer
    to: Project
    cardinality: one_to_many
    description: "一个客户可以有多个项目，每个项目只属于一个客户"

  - name: project_uses_modules
    from: Project
    to: PlatformModule
    cardinality: many_to_many
    description: "项目使用了哪些平台模块，一个模块可被多个项目复用"

  - name: project_assigned_engineer
    from: Project
    to: Engineer
    cardinality: many_to_many
    role_on_from: "负责工程师"
    description: "项目由哪些工程师负责，一个工程师可同时参与多个项目"

  - name: customer_has_tickets
    from: Customer
    to: SupportTicket
    cardinality: one_to_many
    description: "客户提交的支持工单"
```

**AI 拿到这个能做什么**：当用户问"A 客户最近有什么问题"，AI 知道要沿着 `customer_has_tickets` 关系查，而不是瞎猜表名。

---

### 第三层：业务规则与约束（Business Rules）

告诉 AI "什么情况下该怎么理解、什么操作是合法的"。

```yaml
# ontology/business_rules.yaml
rules:
  - scope: Customer
    rule: "contract_status=expired 超过 30 天且无续约意向的客户标记为流失风险"
    action_hint: "建议触发客户回访流程"

  - scope: Project
    rule: "health_status=red 的项目必须在 48 小时内产生一条处理记录"
    action_hint: "提醒项目负责人介入"

  - scope: SupportTicket
    rule: "优先级=urgent 且超过 4 小时未响应，自动升级给技术总监"
    action_hint: "可自动发送通知"

  - scope: PlatformModule
    rule: "被 3 个以上项目使用的模块视为'核心模块'，变更需要审批"
    action_hint: "修改前提示影响范围"

constraints:
  - "不允许删除有进行中项目的客户记录"
  - "项目 stage 只能按顺序流转，不能跳过（requirement→design→development→...）"
  - "annual_revenue 为空时，不可用于自动判断服务等级"
```

---

### 第四层：可执行操作（Actions / Tools）

告诉 AI "你能对这些对象做什么操作"。

```yaml
# ontology/actions.yaml
actions:
  - name: query_customers
    description: "按条件查询客户列表"
    parameters:
      - name: industry
        type: enum
        optional: true
      - name: contract_status
        type: enum
        optional: true
    returns: "Customer[]"

  - name: get_customer_health_report
    description: "获取某客户的整体健康状态，包含项目进度、工单情况、续费预测"
    parameters:
      - name: customer_id
        type: string
        required: true
    returns: "包含项目列表、未关闭工单数、续约风险评分的综合报告"

  - name: create_support_ticket
    description: "为客户创建一个支持工单"
    parameters:
      - name: customer_id
        type: string
        required: true
      - name: title
        type: string
        required: true
      - name: priority
        type: enum[low, medium, high, urgent]
        required: true
    side_effects: "urgent 优先级会触发即时通知"

  - name: change_project_stage
    description: "推进项目到下一阶段"
    parameters:
      - name: project_id
        type: string
        required: true
      - name: new_stage
        type: enum
        required: true
    constraints: "只允许向下一个阶段推进，不允许跳跃"
```

---

## 四、与纯数据库 Schema 的区别对比

| 维度 | 数据库 Schema | 本体层 |
|------|--------------|--------|
| 受众 | DBA/开发者 | AI Agent + 人 |
| 内容 | 字段名、类型、索引 | 语义描述、业务含义、使用场景 |
| 关系 | 外键约束 | 业务语义关系（含方向和含义描述） |
| 规则 | CHECK约束/触发器 | 业务规则 + 推荐动作 |
| 操作 | SQL | 带描述和约束的高层操作 |
| 示例 | `VARCHAR(100) NOT NULL` | `"企业全称（工商注册名），用于合同匹配和客户识别"` |

---

## 五、三种实现方式——完整落地代码

> 以一个实际业务场景贯穿：为一家制造业客户做的"设备维护管理系统 + AI 助手"

### 场景设定

客户是一家有 200 台 CNC 机床的工厂，需要：
- 管理设备信息、维修记录、备件库存
- AI 助手能回答"3号车间哪台设备该保养了"、"上个月故障最多的设备是哪台"
- AI 能自动创建维修工单

---

### 方式一：System Prompt 注入

#### 适用场景
- 业务对象不超过 10-15 个
- Agent 是对话式助手（ChatBot）
- 快速上线，不需要复杂架构

#### 具体做法

**Step 1：编写本体定义文件**

创建一个 `ontology.md` 或 `ontology.yaml`，作为项目交付物之一：

```markdown
# 设备维护管理系统 - 业务本体

## 核心对象

### Equipment（设备）
- equipment_id: 设备唯一编号，格式 "EQ-{车间号}-{序号}"，如 EQ-03-017
- name: 设备名称，如 "FANUC α-D14MiA5 立式加工中心"
- workshop: 所在车间编号（01-08）
- status: 运行状态
  - running: 正常运行中
  - idle: 空闲待命
  - maintenance: 维护保养中
  - fault: 故障停机
- last_maintenance_date: 上次保养完成日期
- maintenance_cycle_days: 保养周期（天），默认90天
- total_running_hours: 累计运行小时数
- purchase_date: 采购日期

### WorkOrder（维修工单）
- order_id: 工单编号，格式 "WO-{年月日}-{序号}"
- equipment_id: 关联设备
- type: 工单类型
  - preventive: 预防性保养（按周期触发）
  - corrective: 故障维修（设备出问题后创建）
  - emergency: 紧急抢修（影响产线停工）
- status: 工单状态
  - open: 已创建待指派
  - assigned: 已指派技师
  - in_progress: 维修中
  - completed: 已完成
  - closed: 已关闭（含验收）
- priority: 优先级
  - P1: 产线停工，4小时内必须响应
  - P2: 设备降级运行，24小时内处理
  - P3: 一般维护，本周内处理
  - P4: 改善类，排期处理
- assigned_technician: 指派的维修技师姓名
- fault_description: 故障描述
- resolution: 处理结果（完成后填写）
- created_at: 创建时间
- completed_at: 完成时间

### SparePart（备件）
- part_id: 备件编号
- name: 备件名称，如 "X轴伺服电机"、"主轴轴承 7014C"
- compatible_equipment: 适用设备型号列表
- stock_quantity: 当前库存数量
- min_stock: 安全库存下限
- lead_time_days: 采购提前期（天）

### Technician（维修技师）
- technician_id: 工号
- name: 姓名
- skills: 擅长领域，如 ["电气", "机械", "液压", "数控系统"]
- current_workload: 当前手头未完成工单数

## 关系
- Equipment → WorkOrder: 一台设备有多条维修记录（一对多）
- WorkOrder → Technician: 一个工单指派给一个技师（多对一）
- WorkOrder → SparePart: 一个工单可能消耗多个备件（多对多）
- SparePart → Equipment: 一个备件适用于多种设备（多对多）

## 业务规则
1. 设备保养判断：当前日期 - last_maintenance_date > maintenance_cycle_days → 该设备应安排保养
2. 紧急工单升级：P1工单创建超4小时未变为 in_progress → 自动通知车间主任
3. 备件预警：stock_quantity < min_stock → 触发采购提醒
4. 技师分配：优先分配 current_workload 最少 + skills 匹配的技师
5. 设备故障频率预警：同一设备30天内创建 >= 3 个 corrective 工单 → 标记为"需要大修评估"
```

**Step 2：将本体定义注入 System Prompt**

```python
import openai

# 读取本体定义文件
with open("ontology.md", "r", encoding="utf-8") as f:
    ontology_content = f.read()

system_prompt = f"""你是一个工厂设备维护管理AI助手。你可以帮助车间主任和维修技师查询设备状态、创建维修工单、查找备件信息。

## 你管理的业务领域

{ontology_content}

## 你的行为准则

1. 用户问设备状态时，根据上述对象定义理解字段含义
2. 判断"该保养了"时，使用规则：当前日期 - last_maintenance_date > maintenance_cycle_days
3. 创建工单时，根据故障严重程度自动判断优先级
4. 推荐技师时，考虑 skills 匹配度和 current_workload
5. 如果用户问的信息你不确定，明确说"我需要查询系统确认"，不要编造数据

## 你可以调用的操作（通过 function calling）

- query_equipment: 查询设备信息
- create_work_order: 创建维修工单
- query_spare_parts: 查询备件库存
- get_maintenance_schedule: 获取保养计划
"""

# 实际对话调用
response = openai.chat.completions.create(
    model="gpt-4o",
    messages=[
        {"role": "system", "content": system_prompt},
        {"role": "user", "content": "3号车间有哪些设备快到保养期了？"}
    ],
    tools=[...],  # function calling 定义（见方式二）
)
```

**Step 3：效果验证**

用户问：「3号车间哪台设备该保养了」

AI 的推理过程（因为有本体层）：
1. "3号车间" → 过滤 workshop = "03"
2. "该保养了" → 使用规则：当前日期 - last_maintenance_date > maintenance_cycle_days
3. 调用 query_equipment 查询，返回结果后用人类可读方式表达

**没有本体层时 AI 会怎样**：不知道"保养"对应哪个字段，不知道判断阈值是什么，可能瞎猜或要求用户提供更多信息。

---

### 方式二：Function Calling / Tool 定义

#### 适用场景
- Agent 需要执行实际操作（创建工单、修改数据）
- 需要严格约束 AI 的操作边界
- 使用 OpenAI / Claude / 开源模型的 Tool Use 能力

#### 具体做法

**Step 1：将本体层的"操作"映射为 Tool Schema**

```python
# tools_definition.py
# 这个文件就是你的本体层"操作定义"的代码实现

tools = [
    {
        "type": "function",
        "function": {
            "name": "query_equipment",
            "description": "查询设备信息。可按车间、状态、是否需要保养等条件筛选。返回设备列表及关键指标。",
            "parameters": {
                "type": "object",
                "properties": {
                    "workshop": {
                        "type": "string",
                        "description": "车间编号，如 '03' 表示3号车间。不传则查所有车间。",
                        "enum": ["01", "02", "03", "04", "05", "06", "07", "08"]
                    },
                    "status": {
                        "type": "string",
                        "description": "设备运行状态筛选。running=运行中，idle=空闲，maintenance=维护中，fault=故障停机",
                        "enum": ["running", "idle", "maintenance", "fault"]
                    },
                    "needs_maintenance": {
                        "type": "boolean",
                        "description": "是否只返回需要保养的设备（即超过保养周期的）。true=只看该保养的，false或不传=看全部"
                    },
                    "keyword": {
                        "type": "string",
                        "description": "按设备名称或编号模糊搜索"
                    }
                },
                "required": []
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "create_work_order",
            "description": "创建一个新的维修工单。创建后系统会根据优先级和技师技能自动推荐指派人选。P1紧急工单创建后会立即通知车间主任。",
            "parameters": {
                "type": "object",
                "properties": {
                    "equipment_id": {
                        "type": "string",
                        "description": "设备编号，格式 EQ-{车间号}-{序号}，如 EQ-03-017"
                    },
                    "type": {
                        "type": "string",
                        "description": "工单类型：preventive=预防性保养（定期），corrective=故障维修，emergency=紧急抢修（产线停了）",
                        "enum": ["preventive", "corrective", "emergency"]
                    },
                    "priority": {
                        "type": "string",
                        "description": "优先级：P1=产线停工4小时内响应，P2=设备降级24小时内，P3=一般本周内，P4=改善类排期",
                        "enum": ["P1", "P2", "P3", "P4"]
                    },
                    "fault_description": {
                        "type": "string",
                        "description": "故障现象描述，尽量包含：什么时候发现、具体表现、是否影响生产"
                    },
                    "preferred_technician": {
                        "type": "string",
                        "description": "优先指派的技师姓名（可选），不指定则系统自动匹配"
                    }
                },
                "required": ["equipment_id", "type", "priority", "fault_description"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "query_spare_parts",
            "description": "查询备件库存信息。可按备件名称搜索，也可查某台设备适用的备件。库存低于安全线的会标记预警。",
            "parameters": {
                "type": "object",
                "properties": {
                    "keyword": {
                        "type": "string",
                        "description": "备件名称关键词，如 '伺服电机'、'轴承'"
                    },
                    "equipment_id": {
                        "type": "string",
                        "description": "查询某台设备适用的所有备件"
                    },
                    "low_stock_only": {
                        "type": "boolean",
                        "description": "是否只看库存预警的备件（库存 < 安全下限）"
                    }
                },
                "required": []
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_equipment_fault_history",
            "description": "获取某台设备的故障维修历史。用于判断设备是否频繁故障、是否需要大修评估。30天内>=3次故障维修的设备会被标记为高风险。",
            "parameters": {
                "type": "object",
                "properties": {
                    "equipment_id": {
                        "type": "string",
                        "description": "设备编号"
                    },
                    "days": {
                        "type": "integer",
                        "description": "查最近多少天的记录，默认30天",
                        "default": 30
                    }
                },
                "required": ["equipment_id"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "get_technician_availability",
            "description": "查询维修技师的当前状态和可用性。返回技师列表含技能、当前工作量、是否可指派。用于工单分配决策。",
            "parameters": {
                "type": "object",
                "properties": {
                    "skill_required": {
                        "type": "string",
                        "description": "需要的技能类型",
                        "enum": ["电气", "机械", "液压", "数控系统", "PLC"]
                    },
                    "workshop": {
                        "type": "string",
                        "description": "优先查某车间的技师"
                    }
                },
                "required": []
            }
        }
    }
]
```

**Step 2：实现 Tool 对应的后端函数**

```python
# tool_handlers.py
from datetime import datetime, timedelta
from database import db  # 你的数据库连接

def handle_query_equipment(params: dict) -> dict:
    """处理设备查询请求"""
    query = db.table("equipment").select("*")
    
    if params.get("workshop"):
        query = query.where("workshop", params["workshop"])
    
    if params.get("status"):
        query = query.where("status", params["status"])
    
    if params.get("keyword"):
        query = query.where_like("name", f"%{params['keyword']}%")
    
    results = query.execute()
    
    # 本体层的业务规则在这里实现
    if params.get("needs_maintenance"):
        today = datetime.now()
        results = [
            eq for eq in results
            if (today - eq["last_maintenance_date"]).days > eq["maintenance_cycle_days"]
        ]
    
    # 返回时附加业务语义信息（让AI能更好地解读）
    for eq in results:
        days_since_maintenance = (datetime.now() - eq["last_maintenance_date"]).days
        eq["maintenance_overdue_days"] = max(0, days_since_maintenance - eq["maintenance_cycle_days"])
        eq["maintenance_urgency"] = (
            "严重超期" if eq["maintenance_overdue_days"] > 30
            else "已超期" if eq["maintenance_overdue_days"] > 0
            else "正常"
        )
    
    return {
        "total": len(results),
        "equipment_list": results
    }


def handle_create_work_order(params: dict) -> dict:
    """创建维修工单"""
    # 验证设备存在
    equipment = db.table("equipment").where("equipment_id", params["equipment_id"]).first()
    if not equipment:
        return {"success": False, "error": f"设备 {params['equipment_id']} 不存在"}
    
    # 生成工单号
    today = datetime.now().strftime("%Y%m%d")
    count = db.table("work_orders").where_like("order_id", f"WO-{today}%").count()
    order_id = f"WO-{today}-{count + 1:03d}"
    
    # 自动匹配技师（本体层规则：优先 workload 最少 + skills 匹配）
    assigned_tech = params.get("preferred_technician")
    if not assigned_tech:
        # 根据工单类型确定需要的技能
        skill_map = {
            "preventive": "机械",
            "corrective": "电气",  # 默认，实际可根据故障描述判断
            "emergency": "电气"
        }
        techs = db.table("technicians") \
            .where_contains("skills", skill_map[params["type"]]) \
            .order_by("current_workload", "asc") \
            .first()
        assigned_tech = techs["name"] if techs else None
    
    # 创建工单
    work_order = {
        "order_id": order_id,
        "equipment_id": params["equipment_id"],
        "type": params["type"],
        "priority": params["priority"],
        "status": "assigned" if assigned_tech else "open",
        "fault_description": params["fault_description"],
        "assigned_technician": assigned_tech,
        "created_at": datetime.now().isoformat()
    }
    db.table("work_orders").insert(work_order)
    
    # P1工单触发通知（本体层规则）
    if params["priority"] == "P1":
        send_notification(
            to="车间主任",
            message=f"紧急工单 {order_id}：{equipment['name']} - {params['fault_description']}"
        )
    
    return {
        "success": True,
        "order_id": order_id,
        "assigned_technician": assigned_tech,
        "message": f"工单 {order_id} 已创建，{'已指派给' + assigned_tech if assigned_tech else '待指派技师'}"
    }
```

**Step 3：Agent 主循环（完整示例）**

```python
# agent.py
import json
import openai

# 加载 tools 定义和 system prompt（含本体层）
from tools_definition import tools
from tool_handlers import handle_query_equipment, handle_create_work_order

TOOL_HANDLERS = {
    "query_equipment": handle_query_equipment,
    "create_work_order": handle_create_work_order,
    "query_spare_parts": handle_query_spare_parts,
    "get_equipment_fault_history": handle_get_equipment_fault_history,
    "get_technician_availability": handle_get_technician_availability,
}

def run_agent(user_message: str, conversation_history: list):
    """Agent 主循环：对话 → 调用工具 → 返回结果"""
    
    conversation_history.append({"role": "user", "content": user_message})
    
    while True:
        response = openai.chat.completions.create(
            model="gpt-4o",
            messages=conversation_history,
            tools=tools,
            tool_choice="auto"
        )
        
        message = response.choices[0].message
        conversation_history.append(message)
        
        # 如果AI决定调用工具
        if message.tool_calls:
            for tool_call in message.tool_calls:
                func_name = tool_call.function.name
                func_args = json.loads(tool_call.function.arguments)
                
                # 执行对应的handler
                handler = TOOL_HANDLERS.get(func_name)
                if handler:
                    result = handler(func_args)
                else:
                    result = {"error": f"未知操作: {func_name}"}
                
                # 将结果返回给AI
                conversation_history.append({
                    "role": "tool",
                    "tool_call_id": tool_call.id,
                    "content": json.dumps(result, ensure_ascii=False)
                })
            
            # 继续循环，让AI根据工具结果生成回复
            continue
        
        # AI生成了最终回复（没有工具调用）
        return message.content


# 使用示例
history = [{"role": "system", "content": system_prompt}]  # system_prompt 含本体层

answer = run_agent("3号车间有哪些设备快到保养期了？帮我给最紧急的那台创建个保养工单", history)
print(answer)

# AI 会：
# 1. 调用 query_equipment(workshop="03", needs_maintenance=True)
# 2. 看到结果后，找出 maintenance_overdue_days 最大的设备
# 3. 调用 create_work_order(equipment_id="EQ-03-017", type="preventive", priority="P3", ...)
# 4. 返回："3号车间有3台设备超过保养周期，最紧急的是 EQ-03-017（FANUC立式加工中心），
#          已超期15天。我已为它创建了保养工单 WO-20250725-001，指派给王技师。"
```

---

### 方式三：RAG 知识库的结构化索引

#### 适用场景
- 业务对象非常多（几十种甚至上百种）
- 无法全部塞进 System Prompt（Token 限制）
- 需要动态扩展（新增业务对象不需要改代码）
- Agent 需要处理的业务范围很广

#### 具体做法

**Step 1：将本体定义拆分为独立文档，按对象存储**

```
knowledge_base/
├── ontology/
│   ├── objects/
│   │   ├── equipment.md          # 设备对象定义
│   │   ├── work_order.md         # 工单对象定义
│   │   ├── spare_part.md         # 备件对象定义
│   │   ├── technician.md         # 技师对象定义
│   │   ├── production_line.md    # 产线对象定义
│   │   └── supplier.md           # 供应商对象定义
│   ├── relations/
│   │   ├── equipment_relations.md
│   │   └── work_order_relations.md
│   ├── rules/
│   │   ├── maintenance_rules.md  # 保养相关规则
│   │   ├── escalation_rules.md   # 升级相关规则
│   │   └── inventory_rules.md    # 库存相关规则
│   └── actions/
│       ├── equipment_actions.md
│       ├── work_order_actions.md
│       └── inventory_actions.md
└── domain_knowledge/
    ├── cnc_common_faults.md      # CNC 常见故障知识
    ├── maintenance_best_practices.md
    └── spare_parts_catalog.md
```

每个文件的内容格式：

```markdown
<!-- knowledge_base/ontology/objects/equipment.md -->
---
type: object_definition
object_name: Equipment
display_name: 设备
domain: 设备维护管理
keywords: [设备, 机床, CNC, 机器, 机台, 加工中心]
---

# Equipment（设备）

## 定义
工厂中需要维护管理的生产设备实体。每台设备有唯一编号、所属车间、运行状态和保养周期。

## 属性说明

| 属性 | 类型 | 业务含义 | 示例值 |
|------|------|----------|--------|
| equipment_id | string | 设备唯一编号，格式 EQ-{车间号}-{序号} | EQ-03-017 |
| name | string | 设备完整名称（品牌+型号+类型） | FANUC α-D14MiA5 立式加工中心 |
| workshop | string | 所在车间编号 | 03 |
| status | enum | 当前运行状态 | running/idle/maintenance/fault |
| last_maintenance_date | date | 上次完成保养的日期 | 2025-04-15 |
| maintenance_cycle_days | int | 保养周期天数，超过则应安排保养 | 90 |
| total_running_hours | float | 设备累计运行小时，用于寿命评估 | 12450.5 |

## 状态含义
- **running**：正在生产中，不可安排保养
- **idle**：空闲未生产，可安排保养
- **maintenance**：正在保养/维修中
- **fault**：发生故障停机，需要维修

## 常见查询场景
- "X号车间有多少台设备" → 按 workshop 筛选并计数
- "哪些设备该保养了" → 当前日期 - last_maintenance_date > maintenance_cycle_days
- "故障停机的设备" → status = fault
- "运行时间最长的设备" → 按 total_running_hours 排序
```

**Step 2：构建向量索引（使用任意向量库）**

```python
# build_ontology_index.py
import os
from pathlib import Path
from langchain.text_splitter import MarkdownHeaderTextSplitter
from langchain_openai import OpenAIEmbeddings
from langchain_community.vectorstores import Chroma  # 或 Milvus, Pinecone, Weaviate

KNOWLEDGE_BASE_DIR = "./knowledge_base"

def load_ontology_documents():
    """加载所有本体定义文档"""
    documents = []
    
    for md_file in Path(KNOWLEDGE_BASE_DIR).rglob("*.md"):
        with open(md_file, "r", encoding="utf-8") as f:
            content = f.read()
        
        # 提取 front-matter 中的元数据
        metadata = extract_frontmatter(content)
        metadata["source_file"] = str(md_file)
        
        # 按 markdown 标题分块，保留结构
        splitter = MarkdownHeaderTextSplitter(
            headers_to_split_on=[
                ("#", "h1"),
                ("##", "h2"),
                ("###", "h3"),
            ]
        )
        chunks = splitter.split_text(content)
        
        for chunk in chunks:
            chunk.metadata.update(metadata)
            documents.append(chunk)
    
    return documents


def build_vector_store():
    """构建向量索引"""
    docs = load_ontology_documents()
    
    embeddings = OpenAIEmbeddings(model="text-embedding-3-small")
    
    vectorstore = Chroma.from_documents(
        documents=docs,
        embedding=embeddings,
        persist_directory="./chroma_ontology_db",
        collection_name="ontology"
    )
    
    print(f"已索引 {len(docs)} 个文档块")
    return vectorstore


if __name__ == "__main__":
    build_vector_store()
```

**Step 3：Agent 运行时动态检索本体信息**

```python
# rag_agent.py
import json
import openai
from langchain_openai import OpenAIEmbeddings
from langchain_community.vectorstores import Chroma

# 加载向量库
embeddings = OpenAIEmbeddings(model="text-embedding-3-small")
vectorstore = Chroma(
    persist_directory="./chroma_ontology_db",
    embedding_function=embeddings,
    collection_name="ontology"
)

# 精简的 System Prompt（不塞完整本体，只放框架）
BASE_SYSTEM_PROMPT = """你是一个工厂设备维护管理AI助手。

你管理的业务领域包括：设备、维修工单、备件、技师。
当你需要理解某个业务概念、判断规则、或确认操作约束时，系统会为你提供相关的本体定义信息。

## 工作方式
1. 理解用户问题涉及哪些业务对象
2. 参考系统提供的本体信息来理解字段含义和业务规则
3. 调用工具执行查询或操作
4. 基于本体规则解读结果，给用户清晰的回答
"""


def retrieve_relevant_ontology(user_query: str, k: int = 3) -> str:
    """根据用户问题，检索相关的本体定义"""
    results = vectorstore.similarity_search(user_query, k=k)
    
    ontology_context = "## 相关业务定义（供你参考）\n\n"
    for doc in results:
        source = doc.metadata.get("source_file", "unknown")
        ontology_context += f"### 来源: {source}\n{doc.page_content}\n\n"
    
    return ontology_context


def run_rag_agent(user_message: str, conversation_history: list):
    """RAG 增强的 Agent 主循环"""
    
    # 1. 根据用户问题检索相关本体信息
    ontology_context = retrieve_relevant_ontology(user_message)
    
    # 2. 构建增强后的消息（将检索到的本体信息作为上下文注入）
    enhanced_messages = [
        {"role": "system", "content": BASE_SYSTEM_PROMPT},
        *conversation_history,
        # 在用户消息前注入本体上下文
        {"role": "system", "content": ontology_context},
        {"role": "user", "content": user_message}
    ]
    
    # 3. 调用模型（带 function calling）
    response = openai.chat.completions.create(
        model="gpt-4o",
        messages=enhanced_messages,
        tools=tools,  # 同方式二的 tools 定义
        tool_choice="auto"
    )
    
    # 4. 处理工具调用（同方式二的循环逻辑）
    message = response.choices[0].message
    
    if message.tool_calls:
        # ... 执行工具调用，获取结果
        # 工具结果返回后，再次检索本体（可能需要不同的规则来解读结果）
        tool_results_text = "..."  # 工具返回的原始数据
        interpretation_context = retrieve_relevant_ontology(
            f"如何解读以下数据：{tool_results_text[:200]}"
        )
        # 将解读上下文也注入，帮助AI正确解读工具返回的数据
        pass
    
    return message.content


# 使用示例
history = []
answer = run_rag_agent("EQ-03-017 这台设备最近是不是故障太频繁了？要不要安排大修？", history)

# AI 的处理流程：
# 1. RAG 检索到：equipment.md（设备定义）+ maintenance_rules.md（含"30天>=3次=需要大修评估"规则）
# 2. 调用 get_equipment_fault_history(equipment_id="EQ-03-017", days=30)
# 3. 得到结果：最近30天有4次 corrective 工单
# 4. 参考规则："30天内>=3次故障维修 → 标记为需要大修评估"
# 5. 回复："EQ-03-017 最近30天发生了4次故障维修，已超过大修评估阈值（3次/30天）。
#          建议安排一次全面检测，主要故障集中在X轴伺服系统。需要我创建一个大修评估工单吗？"
```

---

## 六、三种方式对比总结

| 维度 | 方式一：System Prompt | 方式二：Function Calling | 方式三：RAG 索引 |
|------|----------------------|--------------------------|------------------|
| 实现复杂度 | ★☆☆ 最简单 | ★★☆ 中等 | ★★★ 最复杂 |
| 适合对象数量 | < 15 个 | 不限（但操作数有限） | 不限 |
| Token 消耗 | 高（每次都传完整定义） | 中（只传 tool schema） | 低（按需检索） |
| 维护成本 | 改 prompt 即可 | 需改代码 + 重新部署 | 改文件 + 重建索引 |
| AI 理解准确度 | 最高（信息始终在上下文） | 高（工具描述够详细的话） | 取决于检索质量 |
| 适合阶段 | MVP / 小项目 | 生产环境 / 标准项目 | 大型平台 / 多行业 |

---

## 七、实际项目推荐组合与实施顺序

**对成长期 toB 定制开发 + AI 应用公司最务实的组合**：

```
方式一（System Prompt）用于：项目级的业务语义描述
    +
方式二（Function Calling）用于：Agent 可执行的操作定义
    +
方式三（RAG）用于：跨项目复用的行业知识积累（中期建设）
```

**实施顺序**：
1. **现在就做**：每个新项目交付时，多花 2-4 小时写一份 `ontology.md`（方式一）
2. **每个项目标配**：将 AI 功能的后端接口规范化为 Tool Schema（方式二）
3. **半年后建设**：当同行业项目积累 3+ 个后，开始构建行业知识向量库（方式三）

---

## 八、落地建议

1. **先从一个项目开始**：拿下一个新项目，在设计阶段多花 2-4 小时写本体定义文件
2. **格式不重要**：YAML/JSON/Markdown 都行，关键是信息完整
3. **随项目演进**：本体层不是一次写完的，随着对业务理解加深持续补充
4. **复用积累**：同行业项目的本体层 70% 可复用，这就是你的行业知识壁垒
5. **从 DDD 起步**：如果已有 DDD 设计，直接在其基础上加语义标注，成本最低
