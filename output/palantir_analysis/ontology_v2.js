const pptxgen = require("pptxgenjs");

// ── 配色方案（匹配原 PPT）──
const C = {
  navy:      "0B2344",  // 深蓝主色
  gold:      "D4AF70",  // 金色强调
  white:     "FFFFFF",
  offWhite:  "F5F7FA",
  lightGray: "E2E8F0",
  midGray:   "718096",  // 页码/辅助文字
  subtitle:  "4A5568",  // 副标题
  darkText:  "1A202C",  // 正文
  blue:      "4472C4",  // accent1
  teal:      "0D9488",
  green:     "38A169",
  orange:    "ED7D31",
  red:       "E53E3E",
};

const FONT_TITLE = "Georgia";
const FONT_BODY  = "Calibri";
const TOTAL_SLIDES = 20;

// ── Helper: 页脚 ──
function addFooter(slide, pageNum) {
  slide.addText("企业本体论架构 × AI Agent 落地方法论", {
    x: 0.5, y: 5.2, w: 7.0, h: 0.3,
    fontSize: 7, fontFace: FONT_BODY, color: C.gold, margin: 0,
  });
  slide.addText(`${pageNum} / ${TOTAL_SLIDES}`, {
    x: 8.6, y: 5.2, w: 0.9, h: 0.3,
    fontSize: 7, fontFace: FONT_BODY, color: C.midGray, align: "right", margin: 0,
  });
}

// ── Helper: 内容页标题 ──
function addPageTitle(slide, title, subtitle) {
  slide.addText(title, {
    x: 0.5, y: 0.3, w: 9.0, h: 0.5,
    fontSize: 22, fontFace: FONT_TITLE, color: C.navy, bold: true, margin: 0,
  });
  if (subtitle) {
    slide.addText(subtitle, {
      x: 0.5, y: 0.8, w: 9.0, h: 0.3,
      fontSize: 10, fontFace: FONT_BODY, color: C.subtitle, margin: 0,
    });
  }
}

// ── Helper: 分隔线 ──
function addDivider(slide, pres, y) {
  slide.addShape(pres.shapes.LINE, {
    x: 0.5, y: y, w: 9.0, h: 0,
    line: { color: C.lightGray, width: 0.5 },
  });
}

async function buildPresentation() {
  let pres = new pptxgen();
  pres.layout = "LAYOUT_16x9";
  pres.author = "AI Dev Workspace";
  pres.title  = "企业本体论架构 × AI Agent 落地方法论";

  // ═══════════════════════════════════════════
  // SLIDE 1: 封面
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.navy };

    slide.addText("企业本体论（Ontology）架构", {
      x: 0.5, y: 1.2, w: 9.0, h: 0.8,
      fontSize: 36, fontFace: FONT_TITLE, color: C.white, bold: true, margin: 0,
    });
    slide.addText("× AI Agent 落地方法论", {
      x: 0.5, y: 2.0, w: 9.0, h: 0.6,
      fontSize: 26, fontFace: FONT_TITLE, color: C.gold, margin: 0,
    });
    slide.addText("以大型防水材料集团产销数字化转型为例", {
      x: 0.5, y: 3.0, w: 9.0, h: 0.5,
      fontSize: 16, fontFace: FONT_BODY, color: C.white, margin: 0,
    });
    slide.addText("2026.06", {
      x: 0.5, y: 4.8, w: 9.0, h: 0.4,
      fontSize: 12, fontFace: FONT_BODY, color: C.midGray, margin: 0,
    });
  }

  // ═══════════════════════════════════════════
  // SLIDE 2: 目录
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "目录", null);

    const tocLeft = [
      "01  产销协同痛点",
      "02  本体层核心理念",
      "03  六层架构 + 治理侧柱",
      "04  语义层：对象与关系",
      "05  语义层：业务规则",
      "06  数据映射与实例化层",
      "07  功能层：可执行动作",
      "08  治理侧柱",
      "09  接口层 + AI Agent 适配",
      "10  应用层",
    ];
    const tocRight = [
      "11  持续学习闭环",
      "12  本体层 vs DDD",
      "13  AI Agent 三种落地方式",
      "14  推荐实施组合",
      "15  AI Agent 场景演示",
      "16  实施路线图",
      "17  成功关键 & 陷阱",
      "18  ROI 与衡量指标",
      "19  行动建议",
      "20  Q&A",
    ];

    slide.addText(
      tocLeft.map((t, i) => ({ text: t, options: { breakLine: i < tocLeft.length - 1 } })),
      { x: 0.5, y: 1.2, w: 4.2, h: 4.0, fontSize: 11, fontFace: FONT_BODY, color: C.darkText, valign: "top", margin: 0, paraSpaceAfter: 6 }
    );
    slide.addText(
      tocRight.map((t, i) => ({ text: t, options: { breakLine: i < tocRight.length - 1 } })),
      { x: 5.3, y: 1.2, w: 4.2, h: 4.0, fontSize: 11, fontFace: FONT_BODY, color: C.darkText, valign: "top", margin: 0, paraSpaceAfter: 6 }
    );
    addFooter(slide, 2);
  }

  // ═══════════════════════════════════════════
  // SLIDE 3: 产销协同痛点
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "大型防水材料集团产销数字化痛点", "业务背景与核心挑战");

    // 左侧业务背景
    slide.addText("业务背景", {
      x: 0.5, y: 1.2, w: 4.3, h: 0.3,
      fontSize: 14, fontFace: FONT_BODY, color: C.navy, bold: true, margin: 0,
    });
    slide.addText("集团具备制造+分销+项目交付能力，多个供货组织（工厂）和销售组织并存。已上线 ERP、MES、WMS、经销商系统，但产销协同仍高度依赖人工和 Excel。", {
      x: 0.5, y: 1.55, w: 4.3, h: 1.0,
      fontSize: 11, fontFace: FONT_BODY, color: C.darkText, valign: "top", margin: 0,
    });

    // 右侧痛点列表
    const pains = [
      ["货主与销售组织混淆", "库存归属不清，供货关系依赖人工判断"],
      ["数据孤岛", "同一批次在不同系统状态描述不一致"],
      ["语义不一致", "销售、生产、仓库对同一对象理解不同"],
      ["公司间交易不透明", "关联交易价格/额度/资格散落多系统"],
      ["产销计划脱节", "销售接单与生产释放不同步"],
      ["AI 难以落地", "模型结果无法直接驱动业务动作"],
    ];

    let py = 1.2;
    pains.forEach(([title, desc]) => {
      slide.addText(title, {
        x: 5.2, y: py, w: 4.3, h: 0.2,
        fontSize: 12, fontFace: FONT_BODY, color: C.navy, bold: true, margin: 0,
      });
      slide.addText(desc, {
        x: 5.2, y: py + 0.22, w: 4.3, h: 0.2,
        fontSize: 10, fontFace: FONT_BODY, color: C.subtitle, margin: 0,
      });
      py += 0.58;
    });

    addFooter(slide, 3);
  }

  // ═══════════════════════════════════════════
  // SLIDE 4: 核心理念
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "本体层核心理念", "不是 DB Schema，不是重做 DDD，是\"写给 AI 的领域模型\"");

    // 核心定义
    slide.addShape(pres.shapes.RECTANGLE, {
      x: 0.5, y: 1.2, w: 9.0, h: 0.8,
      fill: { color: C.offWhite },
    });
    slide.addText("AI Agent 需要的不是数据库 Schema，而是一份业务语义描述\u2014\u2014告诉它\u201C这个世界里有什么东西、它们之间怎么关联、能对它们做什么操作\u201D。", {
      x: 0.7, y: 1.3, w: 8.6, h: 0.6,
      fontSize: 12, fontFace: FONT_BODY, color: C.navy, italic: true, valign: "middle", margin: 0,
    });

    // 三列对比表
    const headerOpts = { fontSize: 9, fontFace: FONT_BODY, bold: true, color: C.white, fill: { color: C.navy }, valign: "middle", align: "center", border: [{ pt: 0.5, color: C.navy }], margin: [2, 3, 2, 3] };
    const cellOpts = (alt) => ({ fontSize: 9, fontFace: FONT_BODY, color: C.darkText, fill: { color: alt ? C.offWhite : C.white }, valign: "middle", align: "center", border: [{ pt: 0.5, color: C.lightGray }], margin: [2, 3, 2, 3] });

    const rows = [
      ["维度", "数据库 Schema", "DDD 聚合根", "本体层（Ontology）"],
      ["服务对象", "DBA / 开发者", "代码系统", "AI Agent + 人 + 系统"],
      ["关注点", "字段、索引、约束", "数据一致性边界", "语义理解、关系认知"],
      ["关系表达", "外键", "聚合间只用 ID", "所有关系显式 + 语义描述"],
      ["规则表达", "CHECK / 触发器", "领域服务代码", "自然语言 + 推荐动作"],
      ["操作表达", "SQL", "Command / Method", "带意图描述的 Tool Schema"],
    ].map((row, ri) => {
      if (ri === 0) return row.map(cell => ({ text: cell, options: headerOpts }));
      const alt = ri % 2 === 0;
      return row.map((cell, ci) => ({ text: cell, options: ci === 0 ? { ...cellOpts(alt), bold: true, align: "left" } : cellOpts(alt) }));
    });

    slide.addTable(rows, { x: 0.5, y: 2.2, w: 9.0, colW: [1.3, 2.2, 2.5, 3.0], rowH: 0.38 });

    // 底部关系公式
    slide.addText("本体层 ⊃ DDD 领域模型 ⊃ 聚合根     |     如果 DDD 做得好，本体层 = 在 DDD 上加一层语义标注", {
      x: 0.5, y: 4.7, w: 9.0, h: 0.3,
      fontSize: 10, fontFace: FONT_BODY, color: C.navy, bold: true, align: "center", margin: 0,
    });

    addFooter(slide, 4);
  }

  // ═══════════════════════════════════════════
  // SLIDE 5: 六层架构总览
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "六层核心架构 + 治理侧柱", "产销业务本体论完整架构");

    // 架构层叠（从下到上）
    const layers = [
      { name: "① 语义层", desc: "对象类型 + 关系 + 业务规则", color: C.navy },
      { name: "② 数据映射与实例化层", desc: "多源系统 → 统一对象实例", color: "1E3A5F" },
      { name: "③ 功能层", desc: "可执行的业务 Actions", color: "2B4C72" },
      { name: "④ 接口层 + AI Agent 适配", desc: "SDK / Tool Schema / 三种落地方式", color: "365E85" },
      { name: "⑤ 应用层", desc: "业务应用 + AI Agent 协作", color: "417098" },
      { name: "⑥ 持续学习闭环", desc: "决策捕获 → 规则优化 → 本体演进", color: "4C82AB" },
    ];

    let ly = 4.4;
    layers.forEach((l) => {
      slide.addShape(pres.shapes.RECTANGLE, {
        x: 1.5, y: ly, w: 6.5, h: 0.48,
        fill: { color: l.color },
      });
      slide.addText(l.name, {
        x: 1.7, y: ly, w: 2.5, h: 0.48,
        fontSize: 11, fontFace: FONT_BODY, color: C.white, bold: true, valign: "middle", margin: 0,
      });
      slide.addText(l.desc, {
        x: 4.2, y: ly, w: 3.6, h: 0.48,
        fontSize: 9, fontFace: FONT_BODY, color: C.white, valign: "middle", margin: 0,
      });
      ly -= 0.56;
    });

    // 治理侧柱
    slide.addShape(pres.shapes.RECTANGLE, {
      x: 8.3, y: 1.2, w: 1.0, h: 3.7,
      fill: { color: C.gold },
    });
    slide.addText("治\n理\n侧\n柱", {
      x: 8.3, y: 1.8, w: 1.0, h: 2.5,
      fontSize: 12, fontFace: FONT_BODY, color: C.navy, bold: true, align: "center", valign: "middle", margin: 0,
    });

    // 说明
    slide.addText("治理（权限/审计/合规）贯穿所有层，不是单独一个阶段", {
      x: 0.5, y: 5.0, w: 9.0, h: 0.2,
      fontSize: 9, fontFace: FONT_BODY, color: C.subtitle, italic: true, align: "center", margin: 0,
    });

    addFooter(slide, 5);
  }

  // ═══════════════════════════════════════════
  // SLIDE 6: 语义层 - 对象与关系
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "语义层：核心对象定义", "产销统一业务语言 — 对象类型");

    const headerOpts = { fontSize: 8, fontFace: FONT_BODY, bold: true, color: C.white, fill: { color: C.navy }, valign: "middle", align: "center", border: [{ pt: 0.5, color: C.navy }], margin: [1, 2, 1, 2] };
    const cellOpts = (alt) => ({ fontSize: 8, fontFace: FONT_BODY, color: C.darkText, fill: { color: alt ? C.offWhite : C.white }, valign: "middle", border: [{ pt: 0.5, color: C.lightGray }], margin: [1, 2, 1, 2] });

    const objData = [
      ["对象", "英文名", "核心职责", "关键属性"],
      ["供货组织", "SupplyOrganization", "货主，拥有库存和产能", "region, capacity, product_lines, status"],
      ["销售组织", "SalesOrganization", "卖货方，创建订单", "type(直销/经销/项目), region"],
      ["关联交易协议", "IntercompanyAgreement", "控制供货资格+转移价格+信用额度", "transfer_price, credit_limit_amount"],
      ["生产批次", "ProductionBatch", "归属供货组织的生产单元", "qc_status, expiry_date, available_qty"],
      ["成品库存", "FinishedGoodsInventory", "按货主×仓库×SKU 管理", "available_qty, locked_qty, safety_stock"],
      ["经销商订单", "DealerOrder", "连接销售组织和供货组织", "allocation_strategy, priority, status"],
      ["质检报告", "QualityInspectionReport", "批次放行依据", "result, metrics, disposition"],
      ["生产计划", "ProductionPlan", "补产排程", "trigger_source, planned_qty"],
    ].map((row, ri) => {
      if (ri === 0) return row.map(cell => ({ text: cell, options: headerOpts }));
      const alt = ri % 2 === 0;
      return row.map((cell, ci) => ({ text: cell, options: ci === 0 ? { ...cellOpts(alt), bold: true } : cellOpts(alt) }));
    });

    slide.addTable(objData, { x: 0.3, y: 1.2, w: 9.4, colW: [1.2, 2.2, 2.8, 3.2], rowH: 0.34 });

    // 关系简图
    slide.addText("核心关系", {
      x: 0.5, y: 4.4, w: 9.0, h: 0.25,
      fontSize: 10, fontFace: FONT_BODY, color: C.navy, bold: true, margin: 0,
    });
    slide.addText("供货组织 ──owns→ 库存 / 批次    |    销售组织 ←IntercompanyAgreement→ 供货组织    |    订单 ──allocates→ 批次", {
      x: 0.5, y: 4.65, w: 9.0, h: 0.3,
      fontSize: 9, fontFace: FONT_BODY, color: C.darkText, margin: 0,
    });

    addFooter(slide, 6);
  }

  // ═══════════════════════════════════════════
  // SLIDE 7: 语义层 - 业务规则
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "语义层：业务规则", "显式建模的判断逻辑与约束");

    const rules = [
      ["库存补货预警", "available_quantity < safety_stock → 触发对应供货组织排产", "AI 自动生成 ProductionPlan"],
      ["自动择供规则", "按距离最近 → transfer_price 最优 → 产能最充裕 排序，不拆行", "系统自动匹配"],
      ["信用额度管控", "未结算金额 > credit_limit_amount → 阻断分配", "通知财务审批"],
      ["质检放行前置", "qc_status=passed 后才能入库；failed 批次不可分配", "冻结并通知生产主管"],
      ["临期优先出库", "距 expiry_date < 30 天 → 分配优先级最高（FIFO）", "避免过期报废"],
      ["产能缺口预警", "预测需求 > 剩余产能 + 库存 → 标记产能不足", "建议跨组织调拨"],
      ["协议到期预警", "IntercompanyAgreement 到期前 30 天自动预警", "通知商务续签"],
    ];

    const headerOpts = { fontSize: 8.5, fontFace: FONT_BODY, bold: true, color: C.white, fill: { color: C.navy }, valign: "middle", border: [{ pt: 0.5, color: C.navy }], margin: [2, 3, 2, 3] };
    const cellOpts = (alt) => ({ fontSize: 8.5, fontFace: FONT_BODY, color: C.darkText, fill: { color: alt ? C.offWhite : C.white }, valign: "middle", border: [{ pt: 0.5, color: C.lightGray }], margin: [2, 3, 2, 3] });

    const tableData = [
      [{ text: "规则名称", options: headerOpts }, { text: "判断逻辑", options: headerOpts }, { text: "触发动作", options: headerOpts }],
      ...rules.map((row, i) => row.map((cell, ci) => ({
        text: cell,
        options: ci === 0 ? { ...cellOpts(i % 2 === 0), bold: true } : cellOpts(i % 2 === 0),
      }))),
    ];

    slide.addTable(tableData, { x: 0.3, y: 1.2, w: 9.4, colW: [1.8, 4.6, 3.0], rowH: 0.42 });

    slide.addText("约束：订单只能分配 passed 批次 | 库存扣减只能通过 Action | suspended 供货组织不接受订单", {
      x: 0.5, y: 4.8, w: 9.0, h: 0.25,
      fontSize: 8, fontFace: FONT_BODY, color: C.subtitle, italic: true, margin: 0,
    });

    addFooter(slide, 7);
  }

  // ═══════════════════════════════════════════
  // SLIDE 8: 数据映射与实例化层
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "数据映射与实例化层", "让多源数据成为可操作的业务对象");

    // 左侧方法论
    slide.addText("设计方法论", {
      x: 0.5, y: 1.2, w: 4.3, h: 0.3,
      fontSize: 13, fontFace: FONT_BODY, color: C.navy, bold: true, margin: 0,
    });
    const methods = [
      "识别权威数据源 + 实体解析（Entity Resolution）",
      "Federation（虚拟）vs 物化（复制）策略选择",
      "增量/实时管道设计（CDC + 流处理）",
      "数据质量规则前置到管道中",
    ];
    slide.addText(
      methods.map((t, i) => ({ text: t, options: { bullet: true, breakLine: i < methods.length - 1 } })),
      { x: 0.5, y: 1.55, w: 4.3, h: 1.5, fontSize: 10, fontFace: FONT_BODY, color: C.darkText, valign: "top", margin: 0, paraSpaceAfter: 4 }
    );

    // 右侧数据源
    slide.addText("防水产销数据源", {
      x: 5.2, y: 1.2, w: 4.3, h: 0.3,
      fontSize: 13, fontFace: FONT_BODY, color: C.navy, bold: true, margin: 0,
    });
    const sources = [
      "SAP ERP → 供货组织/销售组织/关联交易协议",
      "MES → 生产批次/车间执行数据",
      "WMS → 成品库存（按货主×仓库×SKU）",
      "经销商 CRM → DealerOrder/客户主数据",
      "质检系统 → QualityInspectionReport",
    ];
    slide.addText(
      sources.map((t, i) => ({ text: t, options: { bullet: true, breakLine: i < sources.length - 1 } })),
      { x: 5.2, y: 1.55, w: 4.3, h: 1.5, fontSize: 10, fontFace: FONT_BODY, color: C.darkText, valign: "top", margin: 0, paraSpaceAfter: 4 }
    );

    // 关键产出
    addDivider(slide, pres, 3.5);
    slide.addText("关键产出：可查询的实时产销对象实例 + 完整数据血缘", {
      x: 0.5, y: 3.6, w: 9.0, h: 0.3,
      fontSize: 12, fontFace: FONT_BODY, color: C.navy, bold: true, margin: 0,
    });
    slide.addText("原始数据不再是孤立的表格，而是业务团队和 AI 可以直接理解和操作的对象。库存归属于供货组织（货主），销售组织通过 IntercompanyAgreement 获得调货资格。", {
      x: 0.5, y: 3.95, w: 9.0, h: 0.7,
      fontSize: 11, fontFace: FONT_BODY, color: C.darkText, valign: "top", margin: 0,
    });

    addFooter(slide, 8);
  }

  // ═══════════════════════════════════════════
  // SLIDE 9: 功能层
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "功能层：可执行的产销业务动作", "围绕完整业务操作设计 Action");

    const headerOpts = { fontSize: 8.5, fontFace: FONT_BODY, bold: true, color: C.white, fill: { color: C.navy }, valign: "middle", border: [{ pt: 0.5, color: C.navy }], margin: [2, 3, 2, 3] };
    const cellOpts = (alt) => ({ fontSize: 8, fontFace: FONT_BODY, color: C.darkText, fill: { color: alt ? C.offWhite : C.white }, valign: "top", border: [{ pt: 0.5, color: C.lightGray }], margin: [2, 3, 2, 3] });

    const actions = [
      ["ReleaseBatchAfterQC", "质检通过→入库→增加库存→生成追溯码→通知销售组织", "batch_id, qc_report_id", "幂等；qc=passed 前置"],
      ["CreateDealerOrderAndAllocate", "下单→自动择供（不拆行）→锁定库存→库存不足触发补产", "sales_org_id, items[], strategy", "检查 Agreement 有效性+额度"],
      ["TriggerProductionReorder", "基于缺口/预测为供货组织生成补产计划", "supply_org_id, sku, qty, source", "AI 推荐需人工审批"],
      ["QueryAvailableInventoryForOrder", "查询销售组织可用库存（按供货组织分组+价格）", "sales_org_id, product_sku", "只返回有效协议范围内"],
      ["GetProductionCapacityGap", "计算供货组织未来N周产能缺口", "supply_org_id, weeks_ahead", "返回缺口明细+建议"],
    ];

    const tableData = [
      [{ text: "Action", options: headerOpts }, { text: "业务语义", options: headerOpts }, { text: "关键参数", options: headerOpts }, { text: "约束", options: headerOpts }],
      ...actions.map((row, i) => row.map((cell, ci) => ({
        text: cell,
        options: ci === 0 ? { ...cellOpts(i % 2 === 0), bold: true } : cellOpts(i % 2 === 0),
      }))),
    ];

    slide.addTable(tableData, { x: 0.2, y: 1.15, w: 9.6, colW: [2.4, 3.2, 2.2, 1.8], rowH: 0.6 });

    slide.addText("设计原则：Action 围绕完整业务操作（非单属性更新） | 区分人类决策/自动化/Agent Action | 幂等+补偿机制", {
      x: 0.5, y: 4.8, w: 9.0, h: 0.3,
      fontSize: 8.5, fontFace: FONT_BODY, color: C.subtitle, italic: true, margin: 0,
    });

    addFooter(slide, 9);
  }

  // ═══════════════════════════════════════════
  // SLIDE 10: 治理侧柱
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "治理侧柱：安全、合规与可审计", "贯穿所有层的横切关注点");

    // 方法论
    slide.addText("设计方法论", {
      x: 0.5, y: 1.2, w: 4.3, h: 0.3,
      fontSize: 13, fontFace: FONT_BODY, color: C.navy, bold: true, margin: 0,
    });
    const govMethods = [
      "基于对象/属性的细粒度访问控制（ABAC）",
      "目的驱动（Purpose-based）访问",
      "全链路审计 + 决策捕获",
      "数据主权与合规规则嵌入（数据安全法适配）",
    ];
    slide.addText(
      govMethods.map((t, i) => ({ text: t, options: { bullet: true, breakLine: i < govMethods.length - 1 } })),
      { x: 0.5, y: 1.55, w: 4.3, h: 1.3, fontSize: 10, fontFace: FONT_BODY, color: C.darkText, valign: "top", margin: 0, paraSpaceAfter: 4 }
    );

    // 产销案例
    slide.addText("产销案例", {
      x: 5.2, y: 1.2, w: 4.3, h: 0.3,
      fontSize: 13, fontFace: FONT_BODY, color: C.navy, bold: true, margin: 0,
    });
    slide.addText("• 区域销售经理只能看到自己负责区域的 DealerOrder 和对应供货组织数据\n\n• IntercompanyAgreement 的 transfer_price 仅财务和供应链总监可见\n\n• 所有 ReleaseBatchAfterQC 操作自动留痕（谁、何时、基于什么质检报告）\n\n• AI Agent 的每次 Action 调用记录完整上下文，可审计可回溯", {
      x: 5.2, y: 1.55, w: 4.3, h: 2.5,
      fontSize: 10, fontFace: FONT_BODY, color: C.darkText, valign: "top", margin: 0,
    });

    // 为何是侧柱
    addDivider(slide, pres, 4.2);
    slide.addText('为何不是独立层：治理不是"做完功能再加"，而是每一层设计时就嵌入的安全约束。放为侧柱更符合实际实施逻辑。', {
      x: 0.5, y: 4.35, w: 9.0, h: 0.5,
      fontSize: 10, fontFace: FONT_BODY, color: C.subtitle, italic: true, valign: "top", margin: 0,
    });

    addFooter(slide, 10);
  }

  // ═══════════════════════════════════════════
  // SLIDE 11: 接口层 + AI Agent 适配
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "接口层 + AI Agent 适配", "从 Ontology 自动生成 Agent 可消费的接口");

    // 左侧：接口层
    slide.addText("接口层职责", {
      x: 0.5, y: 1.2, w: 4.3, h: 0.3,
      fontSize: 13, fontFace: FONT_BODY, color: C.navy, bold: true, margin: 0,
    });
    const ifMethods = [
      "从 Ontology 定义自动生成强类型 SDK",
      "支持 REST / GraphQL / gRPC 协议",
      "统一鉴权、错误处理、版本管理",
      "自动暴露 Agent Tool Schema（Function Calling 格式）",
    ];
    slide.addText(
      ifMethods.map((t, i) => ({ text: t, options: { bullet: true, breakLine: i < ifMethods.length - 1 } })),
      { x: 0.5, y: 1.55, w: 4.3, h: 1.3, fontSize: 10, fontFace: FONT_BODY, color: C.darkText, valign: "top", margin: 0, paraSpaceAfter: 4 }
    );

    // 右侧：AI Agent 四层信息
    slide.addText("AI Agent 需要的四层信息", {
      x: 5.2, y: 1.2, w: 4.3, h: 0.3,
      fontSize: 13, fontFace: FONT_BODY, color: C.navy, bold: true, margin: 0,
    });

    const agentLayers = [
      ["① 对象类型", "有什么东西、什么含义"],
      ["② 关系定义", "对象之间如何关联"],
      ["③ 业务规则", "什么条件触发什么判断"],
      ["④ 可执行操作", "能做什么、参数、约束、副作用"],
    ];
    let ay = 1.6;
    agentLayers.forEach(([name, desc]) => {
      slide.addShape(pres.shapes.RECTANGLE, {
        x: 5.2, y: ay, w: 4.3, h: 0.38,
        fill: { color: C.offWhite },
      });
      slide.addText(name, {
        x: 5.4, y: ay, w: 1.5, h: 0.38,
        fontSize: 10, fontFace: FONT_BODY, color: C.navy, bold: true, valign: "middle", margin: 0,
      });
      slide.addText(desc, {
        x: 6.9, y: ay, w: 2.4, h: 0.38,
        fontSize: 9, fontFace: FONT_BODY, color: C.darkText, valign: "middle", margin: 0,
      });
      ay += 0.45;
    });

    // 产销案例
    addDivider(slide, pres, 3.7);
    slide.addText("产销案例：AI Agent 通过 Tool Schema 安全调用 QueryAvailableInventoryForOrder() 和 CreateDealerOrderAndAllocate()，\n无需关心底层 ERP/WMS/CRM 集成细节。", {
      x: 0.5, y: 3.85, w: 9.0, h: 0.7,
      fontSize: 10, fontFace: FONT_BODY, color: C.darkText, valign: "top", margin: 0,
    });

    addFooter(slide, 11);
  }

  // ═══════════════════════════════════════════
  // SLIDE 12: 应用层
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "应用层：业务应用与工作流", "业务人员和 AI Agent 在同一套对象模型上协作");

    const apps = [
      ["经销商实时库存看板", "按供货组织分组展示可用库存 + 价格 + 自动补货预警"],
      ["订单择供工作流", "自动择供（系统匹配）+ 手动拆行指定供货组织"],
      ["质检放行工作流", "质检通过 → 一键放行 → 自动入库 → 通知销售"],
      ["产销协同计划模拟器", "基于实时对象状态进行 what-if 分析（调产能/调库存）"],
      ["AI 产销助手", "自然语言查询库存/产能/订单状态 + 自动创建工单"],
      ["公司间结算报表", "按 IntercompanyAgreement 自动汇总调货量×价格"],
    ];

    let appY = 1.2;
    apps.forEach(([title, desc], i) => {
      const alt = i % 2 === 0;
      slide.addShape(pres.shapes.RECTANGLE, {
        x: 0.5, y: appY, w: 9.0, h: 0.55,
        fill: { color: alt ? C.offWhite : C.white },
      });
      slide.addText(title, {
        x: 0.7, y: appY, w: 2.8, h: 0.55,
        fontSize: 11, fontFace: FONT_BODY, color: C.navy, bold: true, valign: "middle", margin: 0,
      });
      slide.addText(desc, {
        x: 3.6, y: appY, w: 5.7, h: 0.55,
        fontSize: 10, fontFace: FONT_BODY, color: C.darkText, valign: "middle", margin: 0,
      });
      appY += 0.6;
    });

    slide.addText('关键价值：应用开发从"数据集成地狱"变成"业务逻辑编排"', {
      x: 0.5, y: 4.9, w: 9.0, h: 0.25,
      fontSize: 10, fontFace: FONT_BODY, color: C.navy, bold: true, italic: true, align: "center", margin: 0,
    });

    addFooter(slide, 12);
  }

  // ═══════════════════════════════════════════
  // SLIDE 13: 持续学习闭环
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "持续学习闭环", "决策捕获 → 规则优化 → 本体演进");

    const items = [
      "每次 Action 执行自动记录上下文 + 输入 + 输出 + 结果",
      "人类-in-the-loop 评审机制（AI 推荐的补产计划需人工审批）",
      "自动评估指标（择供准确率/库存周转/订单满足率）回写对象属性",
      "定期 Ontology 演进 + 规则阈值调整 + 模型再训练触发",
    ];
    slide.addText(
      items.map((t, i) => ({ text: t, options: { bullet: true, breakLine: i < items.length - 1 } })),
      { x: 0.5, y: 1.2, w: 9.0, h: 1.5, fontSize: 11, fontFace: FONT_BODY, color: C.darkText, valign: "top", margin: 0, paraSpaceAfter: 6 }
    );

    // 闭环案例
    addDivider(slide, pres, 2.9);
    slide.addText("产销闭环案例", {
      x: 0.5, y: 3.05, w: 9.0, h: 0.3,
      fontSize: 12, fontFace: FONT_BODY, color: C.navy, bold: true, margin: 0,
    });
    slide.addText("ReleaseBatchAfterQC 执行后 → 项目现场 App 反馈施工问题 → 自动回写 ProductionBatch 质量属性 → 影响未来质检规则阈值 → 生产计划模型优化 → 择供策略准确率持续提升", {
      x: 0.5, y: 3.4, w: 9.0, h: 0.8,
      fontSize: 11, fontFace: FONT_BODY, color: C.darkText, valign: "top", margin: 0,
    });

    addFooter(slide, 13);
  }

  // ═══════════════════════════════════════════
  // SLIDE 14: 本体层 vs DDD
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "本体层与 DDD 的关系", "互补而非替代，增量成本极低");

    const headerOpts = { fontSize: 9, fontFace: FONT_BODY, bold: true, color: C.white, fill: { color: C.navy }, valign: "middle", align: "center", border: [{ pt: 0.5, color: C.navy }], margin: [2, 3, 2, 3] };
    const cellOpts = (alt) => ({ fontSize: 9, fontFace: FONT_BODY, color: C.darkText, fill: { color: alt ? C.offWhite : C.white }, valign: "middle", border: [{ pt: 0.5, color: C.lightGray }], margin: [2, 3, 2, 3] });

    const dddData = [
      ["DDD 产出", "本体层映射", "增量工作"],
      ["聚合根", "→ Object Type", "加 description + 枚举语义说明"],
      ["聚合内值对象", "→ 也暴露为 Object Type", "DDD 藏着的，展开给 AI"],
      ["聚合间 ID 引用", "→ 显式关系定义 + 语义描述", "写明方向、含义、基数"],
      ["Domain Service 规则", "→ 自然语言复述 + action_hint", "翻译成 AI 可读格式"],
      ["Command / Method", "→ Action（含意图 + 约束 + 副作用）", "加 description + preconditions"],
    ].map((row, ri) => {
      if (ri === 0) return row.map(cell => ({ text: cell, options: headerOpts }));
      const alt = ri % 2 === 0;
      return row.map((cell, ci) => ({ text: cell, options: ci === 0 ? { ...cellOpts(alt), bold: true } : cellOpts(alt) }));
    });

    slide.addTable(dddData, { x: 0.5, y: 1.2, w: 9.0, colW: [2.5, 3.5, 3.0], rowH: 0.45 });

    slide.addText('结论：如果已有 DDD 设计，本体层就是把已有设计"翻译"给 AI，每个项目增量 2-4 小时。', {
      x: 0.5, y: 4.5, w: 9.0, h: 0.3,
      fontSize: 11, fontFace: FONT_BODY, color: C.navy, bold: true, align: "center", margin: 0,
    });

    addFooter(slide, 14);
  }

  // ═══════════════════════════════════════════
  // SLIDE 15: AI Agent 三种落地方式
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "AI Agent 三种落地方式", "按复杂度递进选择");

    const headerOpts = { fontSize: 9, fontFace: FONT_BODY, bold: true, color: C.white, fill: { color: C.navy }, valign: "middle", align: "center", border: [{ pt: 0.5, color: C.navy }], margin: [2, 3, 2, 3] };
    const cellOpts = (alt) => ({ fontSize: 8.5, fontFace: FONT_BODY, color: C.darkText, fill: { color: alt ? C.offWhite : C.white }, valign: "middle", align: "center", border: [{ pt: 0.5, color: C.lightGray }], margin: [2, 3, 2, 3] });

    const waysData = [
      ["维度", "System Prompt 注入", "Function Calling / Tool", "RAG 知识库索引"],
      ["做法", "完整本体定义塞入 Prompt", "Actions → Tool Schema", "本体文档向量化检索"],
      ["适合对象数", "< 15 个", "不限", "不限（几十到上百）"],
      ["Token 消耗", "高（每次全量）", "中（只传 Schema）", "低（按需检索）"],
      ["AI 理解准确度", "最高", "高", "取决于检索质量"],
      ["维护成本", "改文件即可", "改代码+部署", "改文件+重建索引"],
      ["适合阶段", "MVP / 试点", "生产环境", "大型平台"],
    ].map((row, ri) => {
      if (ri === 0) return row.map(cell => ({ text: cell, options: headerOpts }));
      const alt = ri % 2 === 0;
      return row.map((cell, ci) => ({ text: cell, options: ci === 0 ? { ...cellOpts(alt), bold: true, align: "left" } : cellOpts(alt) }));
    });

    slide.addTable(waysData, { x: 0.3, y: 1.15, w: 9.4, colW: [1.6, 2.5, 2.7, 2.6], rowH: 0.42 });

    addFooter(slide, 15);
  }

  // ═══════════════════════════════════════════
  // SLIDE 16: 推荐实施组合
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "推荐实施组合", "三种方式组合使用，按阶段递进");

    slide.addShape(pres.shapes.RECTANGLE, {
      x: 0.5, y: 1.2, w: 9.0, h: 1.2,
      fill: { color: C.offWhite },
    });
    slide.addText([
      { text: "System Prompt", options: { bold: true, breakLine: true } },
      { text: "→ 项目级业务语义描述（本体定义文件）" },
    ], { x: 0.7, y: 1.25, w: 8.6, h: 0.35, fontSize: 10, fontFace: FONT_BODY, color: C.darkText, margin: 0 });
    slide.addText([
      { text: "Function Calling", options: { bold: true, breakLine: true } },
      { text: "→ Agent 可执行的操作定义（Tool Schema）" },
    ], { x: 0.7, y: 1.65, w: 8.6, h: 0.35, fontSize: 10, fontFace: FONT_BODY, color: C.darkText, margin: 0 });
    slide.addText([
      { text: "RAG 知识库", options: { bold: true, breakLine: true } },
      { text: "→ 跨项目复用的行业知识积累（中期建设）" },
    ], { x: 0.7, y: 2.05, w: 8.6, h: 0.35, fontSize: 10, fontFace: FONT_BODY, color: C.darkText, margin: 0 });

    // 实施顺序
    slide.addText("实施顺序", {
      x: 0.5, y: 2.7, w: 9.0, h: 0.3,
      fontSize: 13, fontFace: FONT_BODY, color: C.navy, bold: true, margin: 0,
    });
    const steps = [
      "1. 现在就做：每个新项目多花 2-4 小时写 ontology.yaml（System Prompt）",
      "2. 每个项目标配：AI 功能后端接口规范化为 Tool Schema（Function Calling）",
      "3. 半年后建设：同行业项目积累 3+ 个后，构建行业知识向量库（RAG）",
    ];
    slide.addText(
      steps.map((t, i) => ({ text: t, options: { breakLine: i < steps.length - 1 } })),
      { x: 0.5, y: 3.05, w: 9.0, h: 1.2, fontSize: 11, fontFace: FONT_BODY, color: C.darkText, valign: "top", margin: 0, paraSpaceAfter: 8 }
    );

    addFooter(slide, 16);
  }

  // ═══════════════════════════════════════════
  // SLIDE 17: AI Agent 场景演示
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "AI Agent 场景演示", "产销协同中的真实对话场景");

    const scenarios = [
      {
        q: "华东区 SBS 卷材还有多少库存可以卖？",
        process: "查 IntercompanyAgreement → 确定可调货供货组织 → 分别查库存",
        answer: "华东工厂可用 85 吨（3200元/吨）；华南工厂可调 120 吨（3400元/吨，运输 3 天）。华南额度剩余 200 万。",
      },
      {
        q: "下个月华东工厂产能够不够？",
        process: "调用 GetProductionCapacityGap(SO-EAST-01, 4周)",
        answer: "第 2 周缺口 30 吨、第 3 周缺口 45 吨。建议华南工厂调拨支援或调整交付承诺。",
      },
      {
        q: "本月华东销售从各工厂调了多少货？结算多少？",
        process: "查 DealerOrder(本月) → 按 supply_org 分组 → 关联 Agreement 取 price",
        answer: "华东工厂 300 吨×3200=96 万；华南工厂 80 吨×3400=27.2 万。合计 123.2 万。",
      },
    ];

    let sy = 1.15;
    scenarios.forEach((s, i) => {
      // Question
      slide.addText(`场景 ${i + 1}：「${s.q}」`, {
        x: 0.5, y: sy, w: 9.0, h: 0.3,
        fontSize: 10, fontFace: FONT_BODY, color: C.navy, bold: true, margin: 0,
      });
      // Process
      slide.addText(`推理：${s.process}`, {
        x: 0.5, y: sy + 0.3, w: 9.0, h: 0.25,
        fontSize: 9, fontFace: FONT_BODY, color: C.subtitle, italic: true, margin: 0,
      });
      // Answer
      slide.addText(`回答：${s.answer}`, {
        x: 0.5, y: sy + 0.55, w: 9.0, h: 0.4,
        fontSize: 9, fontFace: FONT_BODY, color: C.darkText, margin: 0,
      });
      sy += 1.2;
    });

    addFooter(slide, 17);
  }

  // ═══════════════════════════════════════════
  // SLIDE 18: 实施路线图
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "实施路线图（四阶段落地）", "聚焦产销链条试点，快速见效");

    const phases = [
      { phase: "阶段 1", time: "4-8 周", content: "Event Storming + 核心域语义建模\n（SupplyOrg / DealerOrder / IntercompanyAgreement）" },
      { phase: "阶段 2", time: "8-12 周", content: "数据映射 + 最小 Actions 上线\n（ReleaseBatchAfterQC + CreateDealerOrderAndAllocate）\n+ System Prompt 注入 AI Agent" },
      { phase: "阶段 3", time: "并行", content: "治理 + 接口 + 首批应用\n（经销商库存看板、择供工作流、AI 产销助手）\n+ Function Calling 标准化" },
      { phase: "阶段 4", time: "持续", content: "反馈闭环 + RAG 知识库建设\n+ 全域推广（更多供货组织/销售组织接入）" },
    ];

    let py = 1.2;
    phases.forEach((p) => {
      // Phase label
      slide.addShape(pres.shapes.RECTANGLE, {
        x: 0.5, y: py, w: 1.3, h: 0.8,
        fill: { color: C.navy },
      });
      slide.addText(p.phase, {
        x: 0.5, y: py, w: 1.3, h: 0.5,
        fontSize: 11, fontFace: FONT_BODY, color: C.white, bold: true, align: "center", valign: "middle", margin: 0,
      });
      slide.addText(p.time, {
        x: 0.5, y: py + 0.45, w: 1.3, h: 0.35,
        fontSize: 9, fontFace: FONT_BODY, color: C.gold, align: "center", valign: "middle", margin: 0,
      });
      // Content
      slide.addText(p.content, {
        x: 2.0, y: py, w: 7.5, h: 0.8,
        fontSize: 10, fontFace: FONT_BODY, color: C.darkText, valign: "middle", margin: 0,
      });
      py += 0.95;
    });

    addFooter(slide, 18);
  }

  // ═══════════════════════════════════════════
  // SLIDE 19: 成功关键 & ROI
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.white };
    addPageTitle(slide, "成功关键因素 & ROI", null);

    // 左侧成功因素
    slide.addText("成功关键", {
      x: 0.5, y: 1.1, w: 4.3, h: 0.3,
      fontSize: 13, fontFace: FONT_BODY, color: C.navy, bold: true, margin: 0,
    });
    const success = [
      "业务 champion + 跨职能 Ontology 小组",
      "先从高价值产销域试点（供货组织+订单）",
      "Event Storming 作为语义建模起点",
      "Action 设计围绕完整业务操作",
      "IntercompanyAgreement 作为供货关系唯一真相源",
    ];
    slide.addText(
      success.map((t, i) => ({ text: t, options: { bullet: true, breakLine: i < success.length - 1 } })),
      { x: 0.5, y: 1.45, w: 4.3, h: 2.0, fontSize: 10, fontFace: FONT_BODY, color: C.darkText, valign: "top", margin: 0, paraSpaceAfter: 4 }
    );

    // 右侧 ROI
    slide.addText("核心 ROI 指标", {
      x: 5.2, y: 1.1, w: 4.3, h: 0.3,
      fontSize: 13, fontFace: FONT_BODY, color: C.navy, bold: true, margin: 0,
    });

    const roi = [
      ["产销协同响应速度", "提升 40%+"],
      ["批次释放周期", "3-5天 → 1天内"],
      ["库存周转率", "提升 15-25%"],
      ["择供准确率", "持续提升至 85%+"],
      ["新应用集成成本", "下降 60%（建一次用多次）"],
    ];

    const headerOpts = { fontSize: 9, fontFace: FONT_BODY, bold: true, color: C.white, fill: { color: C.navy }, valign: "middle", border: [{ pt: 0.5, color: C.navy }], margin: [2, 3, 2, 3] };
    const cellOpts = (alt) => ({ fontSize: 9, fontFace: FONT_BODY, color: C.darkText, fill: { color: alt ? C.offWhite : C.white }, valign: "middle", border: [{ pt: 0.5, color: C.lightGray }], margin: [2, 3, 2, 3] });

    const roiTable = [
      [{ text: "指标", options: headerOpts }, { text: "目标效果", options: headerOpts }],
      ...roi.map((row, i) => row.map((cell, ci) => ({
        text: cell,
        options: ci === 0 ? { ...cellOpts(i % 2 === 0), bold: true } : cellOpts(i % 2 === 0),
      }))),
    ];
    slide.addTable(roiTable, { x: 5.2, y: 1.45, w: 4.3, colW: [2.3, 2.0], rowH: 0.35 });

    // 陷阱
    addDivider(slide, pres, 3.8);
    slide.addText("常见陷阱（要避免）", {
      x: 0.5, y: 3.9, w: 9.0, h: 0.25,
      fontSize: 11, fontFace: FONT_BODY, color: C.red, bold: true, margin: 0,
    });
    slide.addText("System Silos（按源系统建 Object）| Kitchen Sink（属性堆砌）| God Object（一个类型塞太多）| Action Sprawl（零碎小动作）", {
      x: 0.5, y: 4.2, w: 9.0, h: 0.4,
      fontSize: 9, fontFace: FONT_BODY, color: C.darkText, margin: 0,
    });

    addFooter(slide, 19);
  }

  // ═══════════════════════════════════════════
  // SLIDE 20: 行动建议 + Q&A
  // ═══════════════════════════════════════════
  {
    let slide = pres.addSlide();
    slide.background = { color: C.navy };

    slide.addText("立即行动建议", {
      x: 0.5, y: 0.5, w: 9.0, h: 0.5,
      fontSize: 22, fontFace: FONT_TITLE, color: C.white, bold: true, margin: 0,
    });

    const actions = [
      "1. 本周：组织 1 次 Event Storming 工作坊（聚焦供货组织 + DealerOrder + IntercompanyAgreement）",
      "2. 本周：组建跨职能 Ontology 小组（业务+数据+架构+财务）",
      "3. 两周内：选择 1-2 个高价值产销场景作为 MVP 试点",
      "4. 一个月内：写出第一版 ontology.yaml + Tool Schema，AI Agent 开始试运行",
      "5. 持续：评估现有系统数据质量，为数据映射做准备",
    ];
    slide.addText(
      actions.map((t, i) => ({ text: t, options: { breakLine: i < actions.length - 1 } })),
      { x: 0.5, y: 1.2, w: 9.0, h: 2.5, fontSize: 13, fontFace: FONT_BODY, color: C.white, valign: "top", margin: 0, paraSpaceAfter: 10 }
    );

    // Q&A
    slide.addShape(pres.shapes.LINE, {
      x: 0.5, y: 4.0, w: 9.0, h: 0,
      line: { color: C.gold, width: 1 },
    });
    slide.addText("Q & A", {
      x: 0.5, y: 4.2, w: 9.0, h: 0.5,
      fontSize: 26, fontFace: FONT_TITLE, color: C.gold, align: "center", margin: 0,
    });
    slide.addText("欢迎讨论具体落地细节与下一步计划", {
      x: 0.5, y: 4.7, w: 9.0, h: 0.4,
      fontSize: 12, fontFace: FONT_BODY, color: C.midGray, align: "center", margin: 0,
    });
  }

  // ── 输出 ──
  const outputPath = "output/palantir_analysis/Ontology_AI_Agent_Methodology_v2.pptx";
  await pres.writeFile({ fileName: outputPath });
  console.log(`Generated: ${outputPath}`);
}

buildPresentation().catch(err => { console.error(err); process.exit(1); });
