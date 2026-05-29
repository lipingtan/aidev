const pptxgen = require("pptxgenjs");

let pres = new pptxgen();
pres.layout = "LAYOUT_16x9";
pres.author = "AI Dev";
pres.title = "主流AI模型概览";

let slide = pres.addSlide();
slide.background = { color: "0F172A" };

// 标题
slide.addText("主流 AI 大模型概览", {
  x: 0.6, y: 0.3, w: 8.8, h: 0.7,
  fontSize: 28, fontFace: "Arial", bold: true,
  color: "F8FAFC", margin: 0
});

// 副标题
slide.addText("2025 年主要厂商旗舰模型对比", {
  x: 0.6, y: 0.9, w: 6, h: 0.4,
  fontSize: 12, fontFace: "Arial",
  color: "94A3B8", margin: 0
});

// 卡片数据
const models = [
  { name: "GPT-4o", company: "OpenAI", color: "10B981", desc: "多模态旗舰，支持文本/图像/音频/视频" },
  { name: "Claude Opus 4", company: "Anthropic", color: "8B5CF6", desc: "超长上下文，深度推理与代码能力突出" },
  { name: "Gemini 2.5 Pro", company: "Google", color: "3B82F6", desc: "原生多模态，百万级 Token 上下文窗口" },
  { name: "DeepSeek-R1", company: "DeepSeek", color: "F59E0B", desc: "开源推理模型，数学/代码性能比肩闭源" },
  { name: "Llama 4", company: "Meta", color: "EF4444", desc: "开源生态标杆，Maverick 混合专家架构" },
  { name: "Qwen 3", company: "Alibaba", color: "06B6D4", desc: "中文能力领先，思考/非思考混合模式" },
];

const cardW = 2.8;
const cardH = 1.6;
const startX = 0.6;
const startY = 1.6;
const gapX = 0.3;
const gapY = 0.25;

models.forEach((m, i) => {
  const col = i % 3;
  const row = Math.floor(i / 3);
  const x = startX + col * (cardW + gapX);
  const y = startY + row * (cardH + gapY);

  // 卡片背景
  slide.addShape(pres.shapes.RECTANGLE, {
    x: x, y: y, w: cardW, h: cardH,
    fill: { color: "1E293B" },
    line: { color: "334155", width: 0.5 }
  });

  // 左侧色条
  slide.addShape(pres.shapes.RECTANGLE, {
    x: x, y: y, w: 0.06, h: cardH,
    fill: { color: m.color }
  });

  // 模型名称
  slide.addText(m.name, {
    x: x + 0.2, y: y + 0.15, w: cardW - 0.35, h: 0.4,
    fontSize: 14, fontFace: "Arial", bold: true,
    color: "F8FAFC", margin: 0, valign: "middle"
  });

  // 公司名
  slide.addText(m.company, {
    x: x + 0.2, y: y + 0.5, w: cardW - 0.35, h: 0.3,
    fontSize: 10, fontFace: "Arial",
    color: m.color, margin: 0, valign: "middle"
  });

  // 描述
  slide.addText(m.desc, {
    x: x + 0.2, y: y + 0.85, w: cardW - 0.35, h: 0.6,
    fontSize: 9, fontFace: "Arial",
    color: "94A3B8", margin: 0, valign: "top"
  });
});

// 底部备注
slide.addText("数据截至 2025 年 5 月 | 仅列出各厂商代表性旗舰模型", {
  x: 0.6, y: 5.2, w: 8.8, h: 0.3,
  fontSize: 8, fontFace: "Arial",
  color: "475569", margin: 0
});

pres.writeFile({ fileName: "ai_models_overview.pptx" });
