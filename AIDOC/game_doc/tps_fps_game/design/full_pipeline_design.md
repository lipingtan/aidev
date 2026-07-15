# 3D TPS 射击游戏 — 真人化角色全流程设计

> 工具链：Venice.ai · Character Creator 5.10 · Substance 3D Sampler · Substance 3D Painter · Blender 4.5 (UVPackMaster3) · Godot 4.5
> 视角：第三人称（FPS 机制 + TPS 展示）
> 核心诉求：皮肤/面部/身体高度真人化，各环节无缝衔接

---

## 全流程总览（7 大阶段）

```
┌──────────────────────────────────────────────────────────────────────────────┐
│ Stage 1: 概念与参考收集                                                       │
│   Venice.ai → 概念图 / 面部参考 / 体型参考                                    │
├──────────────────────────────────────────────────────────────────────────────┤
│ Stage 2: 基础人体与面部建模                                                    │
│   Character Creator 5.10 → 高精度基础人体 + 面部形态                          │
├──────────────────────────────────────────────────────────────────────────────┤
│ Stage 3: 拓扑 / UV / 细节雕刻                                                 │
│   Blender 4.5 + UVPackMaster3 → 重拓扑 / UV 展开 / 微表面细节                │
├──────────────────────────────────────────────────────────────────────────────┤
│ Stage 4: 皮肤材质制作                                                         │
│   Substance 3D Sampler → 皮肤扫描/采样基础                                   │
│   Substance 3D Painter → 完整 PBR 皮肤材质绘制                               │
├──────────────────────────────────────────────────────────────────────────────┤
│ Stage 5: 绑定与动画准备                                                       │
│   Blender 4.5 → 骨骼绑定 / 权重绘制 / 面部骨骼 / 形态键                     │
├──────────────────────────────────────────────────────────────────────────────┤
│ Stage 6: 引擎集成与渲染调优                                                    │
│   Godot 4.5 → 导入 / SSS 皮肤着色器 / TPS 相机 / 射击机制                   │
├──────────────────────────────────────────────────────────────────────────────┤
│ Stage 7: 迭代优化与性能验证                                                    │
│   全链路 → LOD / 纹理分档 / 性能预算对齐                                     │
└──────────────────────────────────────────────────────────────────────────────┘
```

---

## Stage 1: 概念与参考收集

### 工具：Venice.ai

### 目标
生成高质量角色概念参考图，确定面部特征、体型、肤色、着装风格。

### 步骤

| # | 操作 | 产出 | 质量要求 |
|---|------|------|----------|
| 1.1 | 生成面部正面/侧面/3/4角度参考 | 3-5 张面部参考图 | 五官比例准确、光照中性 |
| 1.2 | 生成全身体型参考（T-Pose 或 A-Pose） | 2-3 张体型参考图 | 比例正确、肌肉结构清晰 |
| 1.3 | 生成皮肤细节参考（毛孔、雀斑、血管） | 2-3 张微距参考 | 4K 分辨率、自然光 |
| 1.4 | 生成着装/装备概念 | 3-5 张装备概念 | 军事/战术风格、材质层次分明 |
| 1.5 | 建立参考板（Mood Board） | 1 份完整参考板 | 统一存入 `source/` |

### 衔接点
- 产出存入 `assets_source/{游戏名}/concept/character/`
- 面部参考图将作为 CC5 的 Headshot 输入或形态调整参考
- 体型参考用于 CC5 的 morph 调整

---

## Stage 2: 基础人体与面部建模

### 工具：Character Creator 5.10

### 目标
利用 CC5 的真人化基础网格生成高精度人体模型，含面部细节拓扑。

### 步骤

| # | 操作 | 产出 | 关键参数 |
|---|------|------|----------|
| 2.1 | 创建基础角色（选择 CC5 Reallusion 标准人体） | 基础人体 mesh | 选择真人比例 base |
| 2.2 | 使用 Headshot 插件或手动调整面部形态 | 面部高精度 mesh | 对照 Stage 1 面部参考 |
| 2.3 | 调整体型（身高/体重/肌肉/体脂比） | 完整体型 | 对照体型参考 |
| 2.4 | 调整皮肤属性（肤色基础、毛孔密度） | 基础皮肤配置 | CC5 Skin Gen 预设 |
| 2.5 | 添加眼睛/睫毛/眉毛/牙齿 | 面部附属组件 | CC5 原生资产 |
| 2.6 | 导出高模（FBX，含 morph targets） | `.fbx` 高精度模型 | 包含 BlendShape |
| 2.7 | 导出贴图基础（CC5 导出的 Diffuse/Normal/Spec） | CC5 原生贴图集 | 4K 分辨率 |

### CC5 导出设置
```
格式：FBX (Binary)
目标：Blender（或 Other 3D Tools）
包含：
  ✓ Mesh（Subdivision Level 适中，不要最高）
  ✓ Morph Targets / BlendShapes
  ✓ Skeleton（标准 Humanoid）
  ✓ Textures（4K，PNG 格式）
  ✓ 面部骨骼（如使用 CC5 面部表情系统）
```

### 衔接点
- FBX 高模 → Stage 3 的 Blender 重拓扑输入
- CC5 导出的贴图 → Stage 4 的 Substance Painter 基础层
- CC5 骨骼 → Stage 5 的绑定参考（可能需要在 Blender 中重建/优化）

---

## Stage 3: 拓扑 / UV / 细节雕刻

### 工具：Blender 4.5 + UVPackMaster3

### 目标
将 CC5 高模转换为游戏可用的中低模拓扑，展开高效 UV，补充微表面细节。

### 步骤

| # | 操作 | 产出 | 关键参数 |
|---|------|------|----------|
| 3.1 | 导入 CC5 FBX 到 Blender | Blender 场景文件 | 检查缩放（CC5 默认 cm，Blender m） |
| 3.2 | 游戏模型重拓扑（若 CC5 拓扑已满足可跳过） | 低模 mesh（≤30K 面） | 面部保留高密度环线 |
| 3.3 | UV 展开 — 身体 | 身体 UV Layout | 接缝放在隐蔽处（腋下、裤线） |
| 3.4 | UV 展开 — 面部（独立 UV 岛，占比更大） | 面部 UV Layout | 面部占 UV 面积 ≥30% |
| 3.5 | UVPackMaster3 自动优化排布 | 最终 UV Atlas | 填充率 ≥92%，无重叠 |
| 3.6 | 多通道 UV（UV1=Diffuse/Normal, UV2=Lightmap） | 双通道 UV | UV2 无重叠 |
| 3.7 | 高模细节雕刻（Sculpt Mode） | 高模细节 mesh | 毛孔/皱纹/疤痕/微表面 |
| 3.8 | 烘焙法线贴图（高→低） | Normal Map（Tangent） | 4K (4096×4096) |
| 3.9 | 烘焙 AO 贴图 | AO Map | 4K |
| 3.10 | 烘焙 Curvature / Thickness | 辅助贴图 | 用于 Painter 遮罩 |
| 3.11 | 导出低模 FBX + 烘焙贴图 | 游戏模型 + 贴图集 | FBX 含 UV1+UV2 |

### UVPackMaster3 设置
```
- Pack Mode: Automatic
- Rotation Step: 任意旋转（最大填充）
- Margin: 4-8px（在 4K 下）
- 面部 UV 岛设置 Pin（固定较大尺寸）
- Overlap Check: Enabled
```

### Blender 烘焙设置
```
- Ray Distance: 适当（0.01-0.05m，视模型大小）
- Bake Type: Normal (Tangent Space) / AO / Curvature
- Output: 4096×4096, 32bit Float (EXR) → 后转 PNG/TGA
- Cage: 使用或 extrusion + max ray
```

### 衔接点
- 低模 FBX + 烘焙贴图 → Stage 4 的 Substance Painter 输入
- UV Layout 决定了后续所有贴图的空间分配
- 面部 UV 的高占比确保真人化细节精度

---

## Stage 4: 皮肤材质制作

### 工具：Substance 3D Sampler + Substance 3D Painter

### 目标
创建照片级真人皮肤 PBR 材质，含次表面散射、微观细节、区域化差异。

### 4A: Substance 3D Sampler（皮肤基础采样）

| # | 操作 | 产出 | 说明 |
|---|------|------|------|
| 4A.1 | 导入真实皮肤照片参考 | 分析结果 | 高质量皮肤微距照 |
| 4A.2 | 生成 Base Color 变体 | 皮肤基色贴图 | 去除光照信息 |
| 4A.3 | 生成 Normal / Height 细节 | 微表面法线 | 毛孔级别细节 |
| 4A.4 | 生成 Roughness 变体 | 粗糙度参考 | 皮肤不同区域粗糙度差异 |
| 4A.5 | 导出为 Substance Material (.sbsar) | 可复用材质 | 供 Painter 调用 |

### 4B: Substance 3D Painter（完整材质绘制）

| # | 操作 | 产出 | 关键技法 |
|---|------|------|----------|
| 4B.1 | 新建项目，导入 Stage 3 低模 + 烘焙贴图 | Painter 项目 | 模板选 PBR Metallic/Roughness |
| 4B.2 | 导入 CC5 原始贴图作为基础层 | Base 层 | 混合模式 Normal |
| 4B.3 | 叠加 Sampler 生成的皮肤微观细节 | Detail Normal 层 | 混合强度 20-40% |
| 4B.4 | 绘制区域化皮肤差异 | — | 见下方区域规范 |
| 4B.5 | 制作 SSS（次表面散射）贴图 | SSS Color Map | 红色通道=血液分布 |
| 4B.6 | 制作 Thickness 贴图 | Thickness Map | 耳/鼻翼/手指薄处高值 |
| 4B.7 | 调整各通道最终效果 | 完整 PBR 通道集 | 逐通道检查 |
| 4B.8 | 导出最终贴图集 | 所有贴图文件 | 见导出规范 |

### 皮肤区域化差异规范

| 区域 | Base Color 特征 | Roughness | 法线细节 | SSS 强度 |
|------|-----------------|-----------|----------|----------|
| 额头 | 偏黄、油光 | 0.3-0.4（偏光滑） | 横纹/抬头纹 | 中 |
| 鼻子 | 偏红、毛孔大 | 0.25-0.35 | 大毛孔、黑头点 | 高（鼻翼薄） |
| 脸颊 | 偏粉红/红润 | 0.4-0.5 | 细毛孔 | 高 |
| 下巴 | 偏暗/青色（男性胡茬） | 0.4-0.5 | 胡茬微凸 | 中 |
| 耳朵 | 偏红/半透 | 0.5-0.6 | 软骨结构 | 极高（薄） |
| 嘴唇 | 深红/粉 | 0.2-0.3（湿润） | 唇纹竖向 | 极高 |
| 脖子 | 偏暗/红 | 0.5-0.6 | 横纹 | 中 |
| 手部 | 偏红/关节暗 | 0.5-0.7 | 指纹/皱纹深 | 低 |
| 身体 | 均匀肤色 | 0.5-0.6 | 大毛孔稀疏 | 中 |

### Painter 导出设置（对接 Godot 4.5）
```
输出模板：自定义 Godot PBR
通道映射：
  - Base Color（RGB）→ albedo_texture
  - Normal（OpenGL Y+）→ normal_texture
  - Roughness（灰度）→ roughness_texture
  - Metallic（灰度）→ metallic_texture（皮肤=0）
  - AO（灰度）→ ao_texture
  - SSS Color（RGB）→ 自定义通道，供 SSS Shader 使用
  - Thickness（灰度）→ 自定义通道
分辨率：4096×4096（桌面档），导出 2048 版本（移动档）
格式：PNG（无损）或 TGA
```

### 衔接点
- Painter 导出的贴图集 → Stage 6 Godot 材质配置
- SSS/Thickness 贴图 → Godot 自定义皮肤 Shader 输入
- 同一 Painter 项目可再次打开迭代 → Stage 7

---

## Stage 5: 绑定与动画准备

### 工具：Blender 4.5

### 目标
为游戏角色创建生产级骨骼绑定，支持 TPS 射击动画需求。

### 步骤

| # | 操作 | 产出 | 关键参数 |
|---|------|------|----------|
| 5.1 | 导入 Stage 3 低模（已有最终贴图的版本） | Blender 场景 | 确认 UV/法线无变 |
| 5.2 | 创建骨骼（Armature） | 骨骼系统 | Humanoid 标准命名 |
| 5.3 | 添加面部骨骼（或确认使用形态键方案） | 面部绑定 | ≤30 面部骨骼 |
| 5.4 | 身体自动权重绑定 + 手动修正 | 权重数据 | 关节无穿模 |
| 5.5 | 面部权重手动绘制 | 面部权重 | 嘴/眼区域精细 |
| 5.6 | 创建 IK 控制器（手/脚/瞄准） | IK Chain | 射击瞄准用 Upper Body IK |
| 5.7 | 测试基础动作（T-Pose/Idle/Walk/Run） | 测试动画 | 无穿模/扭曲 |
| 5.8 | 创建射击相关骨骼约束 | Aim Constraint | 上半身独立瞄准 |
| 5.9 | 导出 GLB/GLTF | 游戏用模型文件 | 含骨骼+权重+BlendShape |

### 骨骼层级标准（TPS 射击需求）
```
Root
├── Hips
│   ├── Spine
│   │   ├── Spine1
│   │   │   ├── Spine2 (上下半身分离点)
│   │   │   │   ├── Neck
│   │   │   │   │   └── Head
│   │   │   │   │       ├── Eye_L / Eye_R
│   │   │   │   │       └── Jaw
│   │   │   │   ├── Shoulder_L → UpperArm_L → LowerArm_L → Hand_L → Fingers
│   │   │   │   └── Shoulder_R → UpperArm_R → LowerArm_R → Hand_R → Fingers
│   │   │   └── (IK Target: AimTarget)
│   ├── UpperLeg_L → LowerLeg_L → Foot_L → Toe_L
│   └── UpperLeg_R → LowerLeg_R → Foot_R → Toe_R
└── (IK Targets: FootIK_L, FootIK_R, HandIK_L, HandIK_R)
```

### Blender → Godot 导出设置
```
格式：glTF 2.0 (.glb)
包含：
  ✓ Mesh + Normals + UVs
  ✓ Armature（骨骼）
  ✓ Shape Keys（面部形态键）
  ✓ Skinning（权重）
  ✗ 动画（单独导出或分段导出）
Transform: +Y Up, -Z Forward（Godot 默认）
Scale: Apply All Transforms before export
```

### 衔接点
- GLB 模型 → Stage 6 Godot 导入
- 骨骼命名 → Godot AnimationTree 自动映射
- IK Target 骨骼 → Godot 中的程序化瞄准系统
- 上下半身分离点 → AnimationTree 混合空间上半身独立层

---

## Stage 6: 引擎集成与渲染调优

### 工具：Godot 4.5

### 目标
在 Godot 中完成角色集成、皮肤渲染、TPS 相机、射击机制。

### 步骤

| # | 操作 | 产出 | 关键配置 |
|---|------|------|----------|
| 6.1 | 导入 GLB 模型到 Godot | .tscn 场景 | 检查骨骼/网格/UV 完整 |
| 6.2 | 配置材质 — 创建皮肤 ShaderMaterial | 皮肤 Shader | 自定义 SSS Shader |
| 6.3 | 配置贴图通道映射 | 材质参数 | 对应 Stage 4 导出的贴图 |
| 6.4 | 搭建 TPS 相机系统（SpringArm3D） | 相机场景 | 避免穿墙、平滑跟随 |
| 6.5 | 实现角色控制器（CharacterBody3D） | 移动脚本 | WASD + 鼠标控制朝向 |
| 6.6 | 实现射击系统（Raycast + 弹道） | 射击脚本 | 准星、后坐力、弹着点 |
| 6.7 | 配置 AnimationTree | 动画混合 | 上下半身分层混合 |
| 6.8 | 配置环境光照（SDFGI / SSAO / SSR） | 渲染环境 | 皮肤表现最优光照 |
| 6.9 | 测试完整循环 | 可玩场景 | 移动+射击+动画流畅 |

### Godot 皮肤 SSS Shader 核心结构
```gdshader
shader_type spatial;
render_mode blend_mix, depth_draw_opaque, cull_back, diffuse_burley, specular_schlick_ggx;

// 贴图输入
uniform sampler2D albedo_tex : source_color;
uniform sampler2D normal_tex : hint_normal;
uniform sampler2D roughness_tex : hint_roughness_r;
uniform sampler2D ao_tex : hint_default_white;
uniform sampler2D sss_color_tex : source_color;
uniform sampler2D thickness_tex : hint_default_white;

// SSS 参数
uniform float sss_strength : hint_range(0.0, 1.0) = 0.4;
uniform vec3 sss_transmittance_color : source_color = vec3(1.0, 0.4, 0.3);

void fragment() {
    ALBEDO = texture(albedo_tex, UV).rgb;
    NORMAL_MAP = texture(normal_tex, UV).rgb;
    ROUGHNESS = texture(roughness_tex, UV).r;
    AO = texture(ao_tex, UV).r;
    
    // Godot 4.x 内置 SSS
    SSS_STRENGTH = sss_strength;
    SSS_TRANSMITTANCE_COLOR = sss_transmittance_color;
    SSS_TRANSMITTANCE_DEPTH = texture(thickness_tex, UV).r * 0.5;
}
```

### TPS 相机关键参数
```
SpringArm3D:
  - Length: 2.5-3.5m
  - Collision Mask: 环境层
  - Margin: 0.2
Camera3D:
  - FOV: 65-75°
  - Near: 0.05（近距离皮肤细节可见）
  - Far: 500
瞄准偏移:
  - 正常: 角色居中偏左
  - 瞄准(ADS): 右肩过肩视角, FOV 收窄至 50-55°
```

### 衔接点
- 完成后 → Stage 7 性能优化
- SSS Shader 参数 → 根据实际渲染效果迭代调整
- 相机参数 → 根据手感体验迭代

---

## Stage 7: 迭代优化与性能验证

### 工具：全链路

### 目标
确保角色在目标性能预算内运行，视觉质量满足真人化标准。

### 步骤

| # | 操作 | 工具 | 验收标准 |
|---|------|------|----------|
| 7.1 | LOD 生成（3 级） | Blender Decimate | LOD0=30K, LOD1=15K, LOD2=5K |
| 7.2 | 纹理分档导出 | Painter 重导出 | Desktop=4K, Mobile=2K |
| 7.3 | 帧率压测（单角色特写） | Godot Profiler | ≥60fps@1080p |
| 7.4 | 帧率压测（同屏多角色） | Godot Profiler | 5 角色同屏 ≥60fps |
| 7.5 | VRAM 占用检查 | Godot Monitor | 单角色纹理 ≤64MB VRAM |
| 7.6 | 视觉质量审查 | 截图对比 | 皮肤无塑料感、SSS 自然 |
| 7.7 | 动画质量审查 | 实际操作 | 无穿模/权重错误/滑步 |
| 7.8 | 射击手感审查 | 实际操作 | 后坐力/准星/反馈自然 |

### 性能预算（单个主角）

| 项目 | 桌面档 | 移动档 |
|------|--------|--------|
| 面数 | ≤30K tri | ≤15K tri |
| 骨骼数 | ≤80 | ≤50 |
| 贴图总 VRAM | ≤64MB | ≤24MB |
| Draw Call | ≤5 | ≤3 |
| SSS | 启用 | 简化/关闭 |
| LOD 级别 | 3 | 2 |

---

## 阶段衔接矩阵（无缝衔接关键）

| 从 → 到 | 交付物 | 格式 | 检查项 |
|----------|--------|------|--------|
| 1→2 | 概念参考图 | PNG/JPG | 分辨率≥2K、角度齐全 |
| 2→3 | CC5 高模 + 基础贴图 | FBX + PNG (4K) | 缩放单位确认(cm→m) |
| 3→4 | 低模 + 烘焙贴图 + UV | FBX + EXR/PNG | UV 无重叠、法线方向正确 |
| 4→5 | 已贴图的低模 | — (同一 Blender 文件) | 贴图预览无错位 |
| 4→6 | 完整 PBR 贴图集 | PNG/TGA 4K | 通道齐全(6-7 张) |
| 5→6 | 绑定完成的模型 | GLB | 骨骼命名标准、权重正确 |
| 6→7 | 引擎中可运行的角色 | Godot .tscn | 无报错、可操作 |

---

## 关键注意事项

1. **单位统一**：CC5 使用 cm，Blender/Godot 使用 m。导入 Blender 时缩放 ×0.01。
2. **法线空间**：全程使用 OpenGL（Y+）法线空间，CC5/Blender/Painter/Godot 保持一致。
3. **UV 一致性**：Stage 3 确定 UV 后，后续所有贴图操作基于同一 UV。修改 UV 需重做所有烘焙和绘制。
4. **骨骼命名**：使用 Godot 可识别的 Humanoid 标准命名，确保 AnimationTree 自动重定向。
5. **非破坏性工作流**：CC5 项目保留原始文件，Painter 项目保留分层，Blender 保留修改器栈。方便任何阶段回溯修改。
6. **版本管理**：每个 Stage 完成时对交付物做版本快照，避免后续修改导致无法回退。
