# AI 接入架构 — 模型部署与 Action 调用机制

## 一、整体架构总览

```
┌─────────────────────────────────────────────────────────────┐
│                      用户/定时触发                            │
└──────────────────────────┬──────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│              AI Agent 编排层（核心调度）                       │
│  ┌─────────┐  ┌──────────────┐  ┌────────────────────┐     │
│  │ Prompt  │  │ Ontology     │  │ Action Executor    │     │
│  │ Builder │→ │ Context API  │→ │ (Tool Calling)     │     │
│  └─────────┘  └──────────────┘  └────────────────────┘     │
└──────────────────────────┬──────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│              LLM 推理层（可替换）                             │
│  方案A: 云端API（Claude/GPT-4/通义千问）                     │
│  方案B: 私有部署（Qwen2.5-72B / DeepSeek-V3）               │
└──────────────────────────┬──────────────────────────────────┘
                           ↓
┌─────────────────────────────────────────────────────────────┐
│              Action 执行层                                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │ MES 写回 │  │ WMS 写回 │  │ ERP 写回 │  │ 通知服务 │   │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘   │
└─────────────────────────────────────────────────────────────┘
```

---

## 二、是否需要单独部署 LLM？

### 三种方案对比

| 维度 | 方案A: 云端API | 方案B: 私有部署 | 方案C: 混合模式 |
|------|---------------|----------------|----------------|
| 代表 | Claude API / GPT-4 / 通义千问 | Qwen2.5-72B / DeepSeek-V3 | 敏感决策私有 + 通用分析云端 |
| 部署成本 | 0（按调用付费） | GPU服务器 30-80万/年 | 中等 |
| 数据安全 | 数据出境/出企业 | 数据不出内网 | 敏感数据不出 |
| 推理质量 | 最强（GPT-4/Claude） | 接近但略弱 | 按场景选最优 |
| 延迟 | 3-15秒 | 2-8秒（取决于硬件） | 视路由 |
| Function Calling | 原生支持，成熟 | 需要适配，部分模型支持 | 均可 |
| 适合阶段 | MVP/试点期 | 规模化+数据敏感 | 生产环境推荐 |

### 推荐策略

```
试点期（0-3个月）：方案A，用云端API快速验证
  → 通义千问（国内合规）或 Claude API（推理质量最好）
  → 月成本约 2000-8000 元（日均几十次决策调用）

规模化（3-12个月）：方案C，混合模式
  → 涉及客户数据/报价/财务的决策 → 私有部署模型
  → 通用分析/报表解读/知识问答 → 云端API
  → 私有部署推荐：Qwen2.5-72B（中文能力强，Function Calling 支持好）
```

### 私有部署最小硬件配置

```
Qwen2.5-72B（INT4量化）：
  - GPU: 2× A100 80GB 或 4× A6000 48GB
  - 内存: 128GB
  - 推理框架: vLLM / SGLang
  - 并发: 支持 ~10 并发请求

DeepSeek-V3（更大但更便宜的推理）：
  - 通过 DeepSeek API 调用（国内，数据不出境）
  - 成本约为 GPT-4 的 1/10
```

---

## 三、AI Agent 编排层 — 核心调度逻辑

这是整个系统的"大脑"，不是 LLM 本身，而是围绕 LLM 的编排代码。

### 3.1 调度流程

```
┌──────────────────────────────────────────────────────────┐
│                    Agent 主循环                            │
│                                                          │
│  1. 触发（定时/用户/事件）                                │
│           ↓                                              │
│  2. Ontology Context API → 获取当前业务状态               │
│           ↓                                              │
│  3. Prompt Builder → 组装 System + Context + Tools       │
│           ↓                                              │
│  4. 调用 LLM（带 Function Calling / Tool Use）           │
│           ↓                                              │
│  5. 解析 LLM 返回的 tool_calls                           │
│           ↓                                              │
│  6. 校验层 → 检查约束规则、权限、参数合法性               │
│           ↓                                              │
│  7a. 通过 → Action Executor 执行写回                     │
│  7b. 需审批 → 推送审批流，等待人工确认后执行              │
│  7c. 拒绝 → 记录原因，通知 AI 重新决策                   │
│           ↓                                              │
│  8. 执行结果回传 → 可选：让 LLM 确认/调整                │
│           ↓                                              │
│  9. 记录决策日志（审计追溯）                              │
└──────────────────────────────────────────────────────────┘
```

### 3.2 关键点：LLM 不直接操作业务系统

```
LLM 的输出 ≠ 直接执行

LLM 输出的是"结构化意图"（tool_calls），
由 Agent 编排层做校验和执行，LLM 本身无权限直接写任何系统。
```

这个设计保证了：
- **安全性**：LLM 幻觉不会直接造成业务损失
- **可审计**：每个决策都有完整的 input→reasoning→action 日志
- **可回滚**：Action Executor 支持补偿操作

---

## 四、Function Calling 机制详解

### 4.1 什么是 Function Calling

LLM 原生支持的能力：你告诉模型"你有这些工具可以用"，模型在推理后会返回"我要调用哪个工具、传什么参数"，而不是直接返回文本。

```
你发给 LLM 的：
  - system prompt（角色+规则）
  - context（业务状态）
  - tools 定义（JSON Schema 描述每个 Action）
  - user message（"请生成今日排产建议"）

LLM 返回的：
  - 思考过程（可选）
  - tool_calls: [
      {name: "create_production_order", arguments: {...}},
      {name: "create_transfer_order", arguments: {...}},
      {name: "send_alert", arguments: {...}}
    ]
```

### 4.2 代码实现（Python，以通义千问为例）

```python
import json
from openai import OpenAI  # 通义千问兼容 OpenAI SDK

# 初始化客户端（通义千问）
client = OpenAI(
    api_key="sk-xxx",
    base_url="https://dashscope.aliyuncs.com/compatible-mode/v1"
)

# Tool 定义（从 Ontology Action 注册表自动生成）
TOOLS = [
    {
        "type": "function",
        "function": {
            "name": "create_production_order",
            "description": "在指定工厂产线上创建生产工单",
            "parameters": {
                "type": "object",
                "properties": {
                    "factory_id": {"type": "string", "description": "工厂ID"},
                    "line_id": {"type": "string", "description": "产线ID"},
                    "product_id": {"type": "string", "description": "产品SKU"},
                    "quantity": {"type": "number", "description": "生产数量"},
                    "priority": {"type": "string", "enum": ["urgent","normal","low"]},
                    "target_date": {"type": "string", "description": "YYYY-MM-DD"}
                },
                "required": ["factory_id","line_id","product_id","quantity","priority"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "create_transfer_order",
            "description": "从库存充足仓库调拨成品到紧缺仓库",
            "parameters": {
                "type": "object",
                "properties": {
                    "from_warehouse_id": {"type": "string"},
                    "to_warehouse_id": {"type": "string"},
                    "product_id": {"type": "string"},
                    "quantity": {"type": "number"},
                    "reason": {"type": "string"}
                },
                "required": ["from_warehouse_id","to_warehouse_id","product_id","quantity"]
            }
        }
    },
    {
        "type": "function",
        "function": {
            "name": "send_alert",
            "description": "发送业务预警通知",
            "parameters": {
                "type": "object",
                "properties": {
                    "alert_type": {"type": "string"},
                    "severity": {"type": "string", "enum": ["info","warning","critical"]},
                    "target_roles": {"type": "array", "items": {"type": "string"}},
                    "message": {"type": "string"}
                },
                "required": ["alert_type","severity","message"]
            }
        }
    }
]
```

### 4.3 Agent 主循环代码

```python
class ProductionPlanningAgent:
    """产销协同 AI Agent"""
    
    def __init__(self):
        self.ontology = OntologyService()      # Ontology 查询服务
        self.executor = ActionExecutor()        # Action 执行器
        self.validator = ConstraintValidator()  # 约束校验器
        self.audit_log = AuditLogger()          # 审计日志
    
    def run_daily_planning(self, date: str):
        """每日产销决策主流程"""
        
        # Step 1: 从 Ontology 获取业务上下文
        context = self._build_context(date)
        
        # Step 2: 调用 LLM
        response = client.chat.completions.create(
            model="qwen-plus",  # 或 gpt-4o / claude-sonnet
            messages=[
                {"role": "system", "content": SYSTEM_PROMPT},
                {"role": "user", "content": context}
            ],
            tools=TOOLS,
            tool_choice="auto"  # 让模型自己决定调用哪些工具
        )
        
        # Step 3: 解析 tool_calls
        message = response.choices[0].message
        if not message.tool_calls:
            # 模型认为无需操作，记录并返回
            self.audit_log.record(date, "no_action", message.content)
            return
        
        # Step 4: 逐个校验和执行
        results = []
        for tool_call in message.tool_calls:
            action_name = tool_call.function.name
            arguments = json.loads(tool_call.function.arguments)
            
            # 4a. 约束校验
            validation = self.validator.check(action_name, arguments)
            
            if validation.status == "REJECTED":
                # 违反硬约束，拒绝执行
                results.append({
                    "action": action_name,
                    "status": "rejected",
                    "reason": validation.reason
                })
                continue
            
            if validation.status == "NEEDS_APPROVAL":
                # 需要人工审批
                approval_id = self.executor.submit_for_approval(
                    action_name, arguments, validation.approver
                )
                results.append({
                    "action": action_name,
                    "status": "pending_approval",
                    "approval_id": approval_id
                })
                continue
            
            # 4b. 直接执行
            exec_result = self.executor.execute(action_name, arguments)
            results.append({
                "action": action_name,
                "status": "executed",
                "result": exec_result
            })
        
        # Step 5: 记录完整决策日志
        self.audit_log.record(
            date=date,
            context_snapshot=context,
            llm_response=message,
            actions_taken=results
        )
        
        return results
    
    def _build_context(self, date: str) -> str:
        """从 Ontology 层构建业务上下文"""
        
        critical = self.ontology.query_inventory(health="CRITICAL")
        low = self.ontology.query_inventory(health="LOW")
        orders = self.ontology.query_orders(status="待排产", limit=20)
        capacity = self.ontology.query_factory_capacity()
        materials = self.ontology.query_material_risks()
        season = self.ontology.get_season_factor(date)
        
        # 格式化为 LLM 可读的文本
        return f"""
当前日期：{date}，季节系数：{season}

## 成品库存告警
### CRITICAL（<3天）
{self._format_inventory(critical)}

### LOW（3-7天）
{self._format_inventory(low)}

## 待排产订单（Top 20）
{self._format_orders(orders)}

## 工厂产能
{self._format_capacity(capacity)}

## 原材料风险
{self._format_materials(materials)}

请根据以上信息生成今日产销决策，调用相应工具执行。
"""
```

---

## 五、Action Executor — 写回业务系统

### 5.1 执行器架构

```
Agent 调用 executor.execute("create_production_order", {...})
                    ↓
┌─────────────────────────────────────────────┐
│           Action Executor                    │
│                                             │
│  1. 参数标准化（单位转换、ID解析）           │
│  2. 幂等性检查（防重复执行）                 │
│  3. 调用目标系统 API                         │
│  4. 等待响应 / 超时处理                      │
│  5. 记录执行结果                             │
│  6. 失败时触发补偿/重试                      │
└─────────────────────────────────────────────┘
                    ↓
          目标业务系统（MES/WMS/ERP）
```

### 5.2 执行器代码

```python
class ActionExecutor:
    """Action 执行器 — 负责将 AI 决策写回业务系统"""
    
    def __init__(self):
        self.handlers = {
            "create_production_order": self._handle_production_order,
            "create_transfer_order": self._handle_transfer_order,
            "adjust_safety_stock": self._handle_safety_stock,
            "send_alert": self._handle_alert,
        }
        self.executed_ids = set()  # 幂等性：防止重复执行
    
    def execute(self, action_name: str, arguments: dict) -> dict:
        """执行一个 Action"""
        
        # 幂等性检查
        idempotency_key = self._compute_key(action_name, arguments)
        if idempotency_key in self.executed_ids:
            return {"status": "skipped", "reason": "duplicate"}
        
        # 路由到对应 handler
        handler = self.handlers.get(action_name)
        if not handler:
            raise ValueError(f"未注册的 Action: {action_name}")
        
        try:
            result = handler(arguments)
            self.executed_ids.add(idempotency_key)
            return {"status": "success", **result}
        except Exception as e:
            return {"status": "failed", "error": str(e)}
    
    def _handle_production_order(self, args: dict) -> dict:
        """写回 MES 系统：创建生产工单"""
        
        # 调用 MES API
        response = mes_client.post("/api/v1/production-orders", json={
            "factoryCode": args["factory_id"],
            "lineCode": args["line_id"],
            "materialCode": args["product_id"],
            "plannedQty": args["quantity"],
            "priority": args["priority"],
            "targetDate": args.get("target_date"),
            "source": "AI_AGENT",  # 标记来源，便于追溯
            "requestId": generate_uuid()  # MES 侧幂等键
        })
        
        if response.status_code == 201:
            return {
                "order_id": response.json()["orderId"],
                "system": "MES",
                "message": f"工单已创建: {response.json()['orderId']}"
            }
        else:
            raise RuntimeError(f"MES 返回错误: {response.text}")
    
    def _handle_transfer_order(self, args: dict) -> dict:
        """写回 WMS + TMS：创建调拨单"""
        
        # 1. WMS 创建调拨出库单
        wms_resp = wms_client.post("/api/v1/transfer-out", json={
            "fromWarehouse": args["from_warehouse_id"],
            "toWarehouse": args["to_warehouse_id"],
            "sku": args["product_id"],
            "qty": args["quantity"],
            "reason": args.get("reason", "AI自动调拨"),
            "source": "AI_AGENT"
        })
        
        # 2. TMS 创建运输任务
        tms_resp = tms_client.post("/api/v1/shipments", json={
            "origin": args["from_warehouse_id"],
            "destination": args["to_warehouse_id"],
            "referenceId": wms_resp.json()["transferId"]
        })
        
        return {
            "transfer_id": wms_resp.json()["transferId"],
            "shipment_id": tms_resp.json()["shipmentId"],
            "system": "WMS+TMS"
        }
    
    def _handle_alert(self, args: dict) -> dict:
        """发送通知（企业微信/钉钉/短信）"""
        
        notification_service.send(
            channel="wecom",  # 企业微信
            targets=args.get("target_roles", ["运营经理"]),
            title=f"[{args['severity'].upper()}] {args['alert_type']}",
            content=args["message"],
            source="AI_AGENT"
        )
        return {"system": "notification", "message": "通知已发送"}
```

---

## 六、约束校验层 — 防止 AI 乱来

### 6.1 为什么需要这一层

LLM 会犯错：可能排产量超过产能、可能从库存不足的仓库调货、可能参数格式错误。
约束校验层是"AI 和业务系统之间的防火墙"。

### 6.2 校验规则实现

```python
class ConstraintValidator:
    """约束校验器 — 在 Action 执行前拦截不合理的决策"""
    
    def check(self, action_name: str, arguments: dict) -> ValidationResult:
        """校验一个 Action 是否可以执行"""
        
        rules = self._get_rules(action_name)
        for rule in rules:
            result = rule.evaluate(arguments)
            if result.status != "PASS":
                return result
        
        return ValidationResult(status="APPROVED")
    
    def _get_rules(self, action_name: str) -> list:
        """获取某个 Action 的所有校验规则"""
        
        RULES_REGISTRY = {
            "create_production_order": [
                CapacityCheckRule(),      # 产能不超限
                MaterialAvailableRule(),  # 原材料够用
                LineCompatibleRule(),     # 产线支持该产品
                ApprovalThresholdRule(), # 超阈值需审批
            ],
            "create_transfer_order": [
                SourceStockSufficientRule(),  # 源仓调出后不低于安全线
                RouteExistsRule(),            # 物流路线存在
                CrossRegionApprovalRule(),    # 跨区需审批
            ],
            "adjust_safety_stock": [
                AdjustmentRangeRule(),   # 调整幅度不超±50%
                SeasonalLogicRule(),     # 旺季不允许下调
            ]
        }
        return RULES_REGISTRY.get(action_name, [])


class CapacityCheckRule:
    """校验：排产量不超过产线日产能的120%"""
    
    def evaluate(self, args: dict) -> ValidationResult:
        line = ontology.get_production_line(args["line_id"])
        current_load = line.today_planned_qty
        new_total = current_load + args["quantity"]
        max_allowed = line.daily_capacity * 1.2
        
        if new_total > max_allowed:
            return ValidationResult(
                status="REJECTED",
                reason=f"产线{args['line_id']}今日已排{current_load}，"
                       f"新增{args['quantity']}后超过产能上限{max_allowed}"
            )
        
        # 超过100%但不超120%，需要审批
        if new_total > line.daily_capacity:
            return ValidationResult(
                status="NEEDS_APPROVAL",
                approver="生产主管",
                reason=f"产能利用率将达{new_total/line.daily_capacity*100:.0f}%"
            )
        
        return ValidationResult(status="PASS")
```

---

## 七、触发模式 — AI 什么时候运行

### 三种触发方式并存

| 模式 | 触发条件 | 典型场景 |
|------|----------|----------|
| 定时触发 | Cron 定时任务 | 每日6:00生成排产建议 |
| 事件触发 | 业务事件驱动 | 库存跌破阈值、大单进入、产线故障 |
| 人工触发 | 用户在界面提问 | "帮我看看下周华东区的备货情况" |

### 定时触发实现

```python
# 使用 APScheduler 或 Celery Beat
from apscheduler.schedulers.background import BackgroundScheduler

scheduler = BackgroundScheduler()

# 每日早6点运行产销决策
@scheduler.scheduled_job('cron', hour=6, minute=0)
def daily_production_planning():
    agent = ProductionPlanningAgent()
    today = datetime.now().strftime("%Y-%m-%d")
    results = agent.run_daily_planning(today)
    # 结果推送到企业微信/钉钉
    notify_planning_results(results)

# 每4小时检查库存异常
@scheduler.scheduled_job('interval', hours=4)
def inventory_health_check():
    agent = ProductionPlanningAgent()
    # 只检查不执行，发现异常才触发完整决策
    alerts = agent.check_inventory_anomalies()
    if alerts:
        agent.run_emergency_planning(alerts)
```

### 事件触发实现

```python
# 监听业务系统的事件（通过消息队列）
import pika  # RabbitMQ

def on_inventory_alert(channel, method, properties, body):
    """库存告警事件触发"""
    event = json.loads(body)
    
    if event["type"] == "stock_below_threshold":
        agent = ProductionPlanningAgent()
        agent.run_emergency_planning(
            trigger_reason=f"仓库{event['warehouse_id']}产品{event['product_id']}库存告急",
            focus_products=[event["product_id"]]
        )

# 订阅 WMS 的库存变动事件
channel.basic_consume(queue='inventory.alerts', on_message_callback=on_inventory_alert)
```

---

## 八、完整数据流（一次决策的生命周期）

以"每日排产决策"为例，完整走一遍：

```
06:00 定时任务触发
  │
  ├─→ [Ontology Context API]
  │     ├── 查 WMS → 获取所有仓库库存状态
  │     ├── 查 ERP → 获取待排产订单
  │     ├── 查 MES → 获取产线当前负载
  │     ├── 查外部API → 获取天气/沥青行情
  │     └── 计算属性 → 库存健康度、紧迫度、季节系数
  │
  ├─→ [Prompt Builder]
  │     └── 组装: system_prompt + context + tools 定义
  │
  ├─→ [LLM 调用] (通义千问 qwen-plus)
  │     └── 返回: tool_calls = [
  │           {create_production_order, {factory:"岳阳", product:"SBS-4mm", qty:5000, priority:"urgent"}},
  │           {create_transfer_order, {from:"武汉仓", to:"南京仓", product:"JS涂料", qty:800}},
  │           {send_alert, {type:"material_shortage", severity:"warning", ...}}
  │         ]
  │
  ├─→ [Constraint Validator]
  │     ├── 工单1: 产能检查通过，原材料检查通过 → APPROVED
  │     ├── 调拨1: 源仓检查通过，同区域 → APPROVED
  │     └── 预警1: 无约束 → APPROVED
  │
  ├─→ [Action Executor]
  │     ├── → MES API: 创建工单 → 返回 order_id=PO20250418001
  │     ├── → WMS API: 创建调拨单 → 返回 transfer_id=TR001
  │     │   → TMS API: 创建运输任务 → 返回 shipment_id=SH001
  │     └── → 企业微信: 发送预警消息
  │
  ├─→ [Audit Logger]
  │     └── 记录: 日期、上下文快照、LLM原始响应、执行结果、耗时
  │
  └─→ [通知]
        └── 企业微信推送: "今日排产决策已生成，3项操作已执行，0项待审批"
```

---

## 九、人工交互模式

### 9.1 审批流

```
AI 决策 → 需审批 → 推送到审批人（企业微信/钉钉/Web界面）
                         ↓
              审批人看到：
              ┌─────────────────────────────────┐
              │ [紧急排产申请]                    │
              │                                 │
              │ AI 建议：岳阳工厂2号线排产        │
              │ 产品：SBS改性沥青卷材4mm          │
              │ 数量：5000卷                     │
              │ 原因：华东区库存仅剩2.1天         │
              │       有3笔订单4/20前交付         │
              │ 产能影响：利用率将达105%          │
              │                                 │
              │ [✓ 批准]  [✗ 拒绝]  [✎ 修改]    │
              └─────────────────────────────────┘
                         ↓
              批准 → Action Executor 执行
              拒绝 → 记录原因，AI 学习
              修改 → 人工调整参数后执行
```

### 9.2 对话式交互

除了自动决策，用户也可以直接和 AI 对话：

```
用户: "下周华东区 SBS 卷材够用吗？"

AI 内部流程:
  1. 意图识别 → 库存查询（不需要执行 Action）
  2. 查 Ontology → 华东区所有仓库的 SBS 卷材库存 + 未来7天预测
  3. 返回分析结果（纯文本，不触发 tool_call）

AI: "华东区 SBS-4mm 卷材当前总库存 12,000 卷，按近期日均消耗 1,800 卷计算，
     可支撑约 6.7 天。但下周有 2 笔大单（合计 4,500 卷）将交付，
     扣除后实际可用约 4.2 天。建议本周安排一次补货排产。
     需要我现在生成排产计划吗？"

用户: "好，安排吧"

AI: → 触发 create_production_order tool_call
```

---

## 十、系统部署架构

### 10.1 推荐部署方式

```
┌─────────────────────────────────────────────────────────┐
│                    企业内网                               │
│                                                         │
│  ┌──────────────┐    ┌──────────────┐                   │
│  │ Agent 服务    │    │ Ontology 服务 │                   │
│  │ (Python/Go)  │←──→│ (API Server) │                   │
│  │ K8s Pod      │    │ K8s Pod      │                   │
│  └──────┬───────┘    └──────┬───────┘                   │
│         │                   │                           │
│         │            ┌──────┴───────┐                   │
│         │            │ PostgreSQL   │ ← Ontology 数据    │
│         │            │ + Redis 缓存  │                   │
│         │            └──────────────┘                   │
│         │                                               │
│  ┌──────┴───────┐                                       │
│  │ Action       │──→ MES API                            │
│  │ Executor     │──→ WMS API                            │
│  │              │──→ ERP API                            │
│  │              │──→ 企业微信 Webhook                    │
│  └──────────────┘                                       │
│                                                         │
│  ┌──────────────┐                                       │
│  │ 消息队列      │ ← 事件触发（RabbitMQ/Kafka）          │
│  └──────────────┘                                       │
└────────────────────────────┬────────────────────────────┘
                             │ HTTPS（仅 LLM 调用出网）
                             ↓
                    ┌──────────────────┐
                    │ LLM API          │
                    │ (通义千问/DeepSeek)│
                    └──────────────────┘
```

### 10.2 关键设计决策

| 决策点 | 选择 | 原因 |
|--------|------|------|
| Agent 服务语言 | Python | LLM SDK 生态最好，开发效率高 |
| Ontology 存储 | PostgreSQL | 关系查询+JSON灵活性，够用 |
| 缓存 | Redis | 高频查询的计算属性缓存 |
| 消息队列 | RabbitMQ | 事件驱动触发，轻量够用 |
| LLM 调用 | HTTPS 出网 | 试点期用云端API，后期可切私有 |
| 部署 | K8s / Docker Compose | 按企业现有基础设施选择 |

### 10.3 最小部署（试点期）

如果企业没有 K8s，最小部署只需要一台服务器：

```
一台 Linux 服务器（4核8G即可）
├── Docker Compose 启动：
│   ├── agent-service (Python FastAPI)
│   ├── ontology-service (Python FastAPI)
│   ├── postgres (数据存储)
│   ├── redis (缓存)
│   └── rabbitmq (事件队列)
│
├── 定时任务: cron 触发 agent
├── LLM: 调用通义千问 API（无需本地GPU）
└── 对接: 通过 HTTP 调用现有 MES/WMS/ERP 的 API
```

---

## 十一、成本估算（试点期3个月）

| 项目 | 月成本 | 说明 |
|------|--------|------|
| LLM API 调用 | ¥3,000-8,000 | 日均50-100次决策调用 |
| 服务器 | ¥2,000 | 4核8G云服务器 |
| 开发人力 | 2人×3个月 | 1后端+1数据工程 |
| 总计（不含人力） | ~¥15,000-30,000/3个月 | |

规模化后如果切私有部署 LLM：
- GPU 服务器：¥30-50万/年（2×A100）
- 但 API 调用费降为 0

---

## 十二、总结：关键架构原则

1. **LLM 只输出意图，不直接执行** — Agent 编排层做校验和执行
2. **Ontology 是 LLM 的"眼睛"** — 没有 Ontology，LLM 看不懂业务
3. **约束校验是安全网** — 防止 AI 幻觉造成业务损失
4. **渐进式自动化** — 先建议→再审批→最后自动，逐步建立信任
5. **可审计可回滚** — 每个决策都有完整日志，出问题能追溯
6. **LLM 可替换** — Agent 层和 LLM 层解耦，随时切换模型供应商
