# 工单管理数据库设计

## 表清单

| 表名 | 说明 | 核心字段 |
|------|------|---------|
| wo_work_order | 工单主体 | order_no, order_type, status, reporter_id, repair_user_id, community_id |
| wo_work_order_image | 工单图片 | order_id, image_url, image_type(REPORT/REPAIR/REJECT) |
| wo_dispatch_record | 派发记录 | order_id, repair_user_id, dispatch_type(MANUAL/AUTO) |
| wo_repair_material | 维修材料 | order_id, material_name, quantity, unit_price, total_price |
| wo_evaluation | 评价记录 | order_id, score, content, evaluate_type(OWNER/AUTO) |
| wo_urge_record | 催单记录 | order_id, urger_id, urge_time |
| wo_work_order_log | 操作日志 | order_id, action, from_status, to_status, operator_id |

详细 DDL 参见 design.md 第三章。
