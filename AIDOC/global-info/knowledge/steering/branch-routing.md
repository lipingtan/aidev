# 2D/3D 分支路由规范

## 概述

本文档定义了 Godot 项目中 2D 和 3D 内容的选择规则、实现差异和混合模式规范，确保在正确的场景使用正确的技术方案。

---

## 一、6 个分支节点的 2D/3D 对比表

### 1.1 场景搭建

| 维度 | 节点基础 | 坐标系 | 单位 | 场景组织 | 典型工作流 |
|------|----------|--------|------|----------|-----------|
| **3D** | Node3D | XYZ 三轴 | 1单位=1米 | 按功能分层（地形/建筑/道具/NPC） | 灰盒→白盒→美术替换 |
| **2D** | Node2D | XY 二轴 | 像素 | 按图层分层（背景/中景/前景/UI） | 概念图→切图→拼接→动画 |

**3D 场景搭建流程：**
```
1. 创建 Node3D 根节点
2. 添加 WorldEnvironment（天空盒、环境光）
3. 添加 DirectionalLight3D（主光源）
4. 搭建地形（MeshInstance3D / CSG / 导入模型）
5. 放置碰撞体（StaticBody3D + CollisionShape3D）
6. 配置导航网格（NavigationRegion3D）
7. 放置交互点（Area3D）
```

**2D 场景搭建流程：**
```
1. 创建 Node2D 根节点
2. 添加 ParallaxBackground（视差背景）
3. 添加 TileMap 层（地形/碰撞/装饰）
4. 放置 Sprite2D 对象
5. 配置碰撞（StaticBody2D + CollisionPolygon2D）
6. 配置导航（NavigationRegion2D）
7. 放置交互点（Area2D）
```

### 1.2 角色控制器

| 维度 | 基础节点 | 移动方式 | 碰撞形状 | 动画系统 | 输入处理 |
|------|----------|----------|----------|----------|----------|
| **3D** | CharacterBody3D | move_and_slide() 三轴 | CapsuleShape3D / BoxShape3D | AnimationTree + 骨骼动画 | 方向键→3D向量 + 鼠标→旋转 |
| **2D** | CharacterBody2D | move_and_slide() 二轴 | CapsuleShape2D / CircleShape2D | AnimationPlayer + SpriteFrames | 方向键→2D向量 |

**3D 角色控制器核心代码结构：**
```gdscript
# 3D 移动核心
var direction: Vector3 = Vector3.ZERO
direction.x = Input.get_axis("move_left", "move_right")
direction.z = Input.get_axis("move_forward", "move_backward")
direction = direction.rotated(Vector3.UP, camera_rotation.y)
velocity = direction.normalized() * speed
move_and_slide()
```

**2D 角色控制器核心代码结构：**
```gdscript
# 2D 移动核心
var direction: Vector2 = Vector2.ZERO
direction.x = Input.get_axis("move_left", "move_right")
direction.y = Input.get_axis("move_up", "move_down")
velocity = direction.normalized() * speed
move_and_slide()
```

### 1.3 相机系统

| 维度 | 基础节点 | 跟随方式 | 视角控制 | 常用配置 | 特殊处理 |
|------|----------|----------|----------|----------|----------|
| **3D** | Camera3D | SpringArm3D + 偏移 | 鼠标控制俯仰/旋转 | FOV、Near/Far、环境遮挡 | 防穿墙（SpringArm碰撞）、锁定目标 |
| **2D** | Camera2D | position_smoothing | 无旋转（或有限旋转） | Zoom、Limit、Drag | 边界限制、屏幕震动 |

**3D 相机典型配置：**
```
Camera3D
├── SpringArm3D (length=5.0, collision_mask=1)
│   └── Camera3D (fov=70, near=0.1, far=1000)
└── 脚本: 鼠标输入→旋转, 滚轮→距离, 锁定→插值朝向
```

**2D 相机典型配置：**
```
Camera2D
├── position_smoothing_enabled = true
├── position_smoothing_speed = 5.0
├── limit_left/right/top/bottom = 场景边界
└── 脚本: 跟随目标, 震动效果, 区域切换
```

### 1.4 物理系统

| 维度 | 物理空间 | 碰撞层 | 射线检测 | 力的应用 | 关节 |
|------|----------|--------|----------|----------|------|
| **3D** | PhysicsServer3D | 32层 3D碰撞 | RayCast3D / PhysicsRayQueryParameters3D | Vector3 力/冲量 | Generic6DOFJoint3D 等 |
| **2D** | PhysicsServer2D | 32层 2D碰撞 | RayCast2D / PhysicsRayQueryParameters2D | Vector2 力/冲量 | PinJoint2D / DampedSpringJoint2D |

**碰撞层规划**：本工作台 **统一遵守** `godot/godot-engine.md` 第四节《碰撞层标准分配》（同样写入 `godot/code-generation.md` 第三节）。  
此处不再单独维护一份方案，避免与代码生成规范冲突。2D 项目沿用相同的 Layer 编号（1=PlayerHurtbox、2=EnemyHurtbox、3=PlayerHitbox、4=EnemyHitbox、5=Environment、6=Interactable、7=Projectile、8=Trigger），按 2D 节点替换实现类型即可。

### 1.5 光照系统

| 维度 | 光源类型 | 阴影 | 全局照明 | 后处理 | 性能考量 |
|------|----------|------|----------|--------|----------|
| **3D** | Directional/Omni/Spot Light3D | 实时阴影 + 烘焙 | SDFGI / VoxelGI / LightmapGI | Glow/SSAO/SSR/Fog | 阴影分辨率、级联数、光源数量 |
| **2D** | PointLight2D / DirectionalLight2D | 2D 阴影（遮挡体） | CanvasModulate | 自定义 Shader | Light2D 数量、遮挡体复杂度 |

**3D 光照标准配置：**
```
WorldEnvironment
├── Environment
│   ├── Background: Sky (ProceduralSkyMaterial)
│   ├── Ambient Light: Sky + 0.3 energy
│   ├── Tonemap: ACES
│   ├── SSAO: enabled (radius=1.0)
│   └── Glow: enabled (threshold=1.0)
└── DirectionalLight3D
    ├── shadow_enabled = true
    ├── directional_shadow_mode = PARALLEL_4_SPLITS
    └── light_energy = 1.0
```

**2D 光照标准配置：**
```
CanvasModulate (color = 环境基础色调)
├── PointLight2D (角色光源)
│   ├── texture = 光照纹理
│   ├── energy = 1.0
│   └── shadow_enabled = true
└── DirectionalLight2D (全局方向光)
```

### 1.6 UI 系统

| 维度 | 基础节点 | 布局方式 | 响应式 | 输入处理 | 与游戏世界的关系 |
|------|----------|----------|--------|----------|-----------------|
| **3D 游戏的 UI** | Control (CanvasLayer) | Container 布局 | 锚点 + 容器自适应 | 鼠标/手柄导航 | 通过 CanvasLayer 叠加在 3D 之上 |
| **2D 游戏的 UI** | Control (CanvasLayer) | Container 布局 | 锚点 + 容器自适应 | 鼠标/手柄导航 | 通过 CanvasLayer 叠加在 2D 之上 |
| **3D 世界内 UI** | SubViewport + Sprite3D | 3D 空间定位 | 固定尺寸 | 射线检测交互 | 存在于 3D 空间中（如血条、名牌） |

> **注意**：UI 系统本身始终是 2D（Control 节点），区别在于它叠加在 3D 还是 2D 游戏世界之上。

---

## 二、3D 资产导入流程

### 2.1 五步骤流程

```
步骤 1: 格式验证
    ├── 检查文件格式（.glb/.gltf 优先，.fbx/.obj 备选）
    ├── 检查文件大小（单模型 ≤50MB）
    ├── 检查多边形数（角色 ≤30K，环境 ≤100K，道具 ≤5K）
    └── 完成标准：文件可被 Godot 识别并预览

步骤 2: 导入配置
    ├── 设置导入选项（缩放、轴向、动画分离）
    ├── 配置材质映射（PBR 通道对应）
    ├── 配置 LOD 级别（如需要）
    └── 完成标准：模型在编辑器中显示正确，无材质丢失

步骤 3: 碰撞配置
    ├── 生成碰撞形状（简化凸包 / 三角网格 / 手动）
    ├── 设置碰撞层和掩码
    ├── 验证碰撞边界合理性
    └── 完成标准：碰撞形状贴合模型，无穿透

步骤 4: 场景集成
    ├── 创建 PackedScene（.tscn）
    ├── 添加必要子节点（AnimationPlayer、Area3D 等）
    ├── 配置脚本挂载点
    ├── 设置导出变量（可在 Inspector 中调整的参数）
    └── 完成标准：场景可实例化，脚本可挂载

步骤 5: 质量验证
    ├── 运行时性能检查（FPS 影响）
    ├── 视觉质量检查（材质、光照、阴影）
    ├── 交互功能检查（碰撞、触发、动画）
    ├── 内存占用检查
    └── 完成标准：满足性能预算，视觉达标，功能正常
```

### 2.2 导入配置速查

| 资产类型 | 缩放 | 动画 | 碰撞 | LOD | 材质 |
|----------|------|------|------|-----|------|
| 角色模型 | 1.0 | 分离为独立动画 | 胶囊体 | 3级 | PBR |
| 环境模型 | 1.0 | 无/简单 | 三角网格 | 2-3级 | PBR/烘焙 |
| 道具模型 | 1.0 | 无/简单 | 凸包 | 1-2级 | PBR |
| 武器模型 | 1.0 | 无 | 简化凸包 | 无 | PBR |
| 特效模型 | 1.0 | 顶点动画 | 无 | 无 | 自发光 |

---

## 三、2D AI 生成流程

### 3.1 五步骤流程

```
步骤 1: 提示词设计
    ├── 确定风格关键词（参考 style-keywords.md）
    ├── 确定内容描述（主体、姿态、表情、场景）
    ├── 确定技术参数（尺寸、比例、透明背景）
    ├── 添加负面提示词（参考 negative-prompts.md）
    └── 完成标准：提示词明确、无歧义、风格一致

步骤 2: 批量生成
    ├── 同风格资产使用相同种子/风格锁定
    ├── 每个资产生成 3-5 个变体供选择
    ├── 记录使用的提示词和参数
    └── 完成标准：每个资产有至少 1 个满意的变体

步骤 3: 后处理
    ├── 去除背景（如需透明）
    ├── 统一尺寸和分辨率
    ├── 色彩校正（确保风格一致）
    ├── 边缘处理（抗锯齿、描边）
    └── 完成标准：所有资产视觉风格统一，尺寸规范

步骤 4: 切图与动画
    ├── 精灵表切割（SpriteFrames 配置）
    ├── 动画帧序列定义
    ├── 九宫格切割（UI 元素）
    ├── 图集打包（TextureAtlas）
    └── 完成标准：动画流畅，切割精确，图集无浪费

步骤 5: 导入与验证
    ├── 导入到 Godot（设置过滤模式、重复模式）
    ├── 创建 SpriteFrames 资源
    ├── 在场景中预览动画效果
    ├── 检查内存占用
    └── 完成标准：显示正确，动画流畅，内存在预算内
```

### 3.2 AI 生成质量评估标准

| 评估维度 | 合格标准 | 不合格表现 |
|----------|----------|-----------|
| 风格一致性 | 与项目美术风格 90%+ 匹配 | 风格跳跃、混搭感明显 |
| 细节完整性 | 主体完整、无缺失部分 | 手指异常、物品残缺 |
| 边缘质量 | 边缘清晰、无毛刺 | 锯齿严重、背景残留 |
| 色彩协调 | 色调与项目调色板一致 | 色彩过饱和/过灰 |
| 尺寸适配 | 与游戏中其他元素比例协调 | 比例失调 |

---

## 四、混合模式规范

### 4.1 3D 世界 + 2D UI

**最常见的混合模式**，所有 3D 游戏都会使用。

```
场景树结构：
Root (Node3D)
├── World (Node3D)           # 3D 游戏世界
│   ├── Environment
│   ├── Characters
│   └── Level
├── CanvasLayer (layer=1)    # HUD 层
│   ├── HealthBar
│   ├── MiniMap
│   └── Crosshair
└── CanvasLayer (layer=10)   # 菜单层（最上层）
    ├── PauseMenu
    ├── InventoryUI
    └── DialogueBox
```

**规则：**
- HUD 使用 CanvasLayer layer=1（始终显示）
- 菜单使用 CanvasLayer layer=10（覆盖一切）
- 3D 世界内的 UI（血条、名牌）使用 SubViewport + Sprite3D
- UI 脚本放 `scripts/ui/`，游戏逻辑放 `scripts/gameplay/`

### 4.2 3D 世界 + 2D 特效

**用于粒子特效、技能特效等。**

```
场景树结构：
Character (CharacterBody3D)
├── Model (MeshInstance3D)
├── EffectsLayer (Node3D)
│   ├── GPUParticles3D        # 3D 粒子（烟雾、火焰）
│   └── Sprite3D              # 2D 特效贴片（Billboard 模式）
│       └── AnimatedSprite3D  # 序列帧特效
└── CanvasLayer               # 全屏特效（闪白、模糊）
    └── ColorRect (shader)
```

**规则：**
- 3D 空间中的 2D 特效使用 Sprite3D（billboard=true）
- 全屏后处理特效使用 CanvasLayer + Shader
- 粒子优先使用 GPUParticles3D（性能更好）
- 序列帧特效用于复杂手绘动画效果

### 4.3 小地图

**3D 世界的 2D 俯视图表示。**

```
实现方案 A：SubViewport 实时渲染
├── SubViewportContainer (Control, 在 CanvasLayer 中)
│   └── SubViewport
│       └── Camera3D (正交投影, 俯视)
└── 优点：实时更新，自动同步
    缺点：性能开销较大

实现方案 B：预渲染 + 标记点
├── TextureRect (预渲染的地图图片)
├── 玩家标记 (Sprite2D, 根据3D位置换算2D坐标)
├── 敌人标记 (Sprite2D)
└── 优点：性能好
    缺点：地图变化时需重新渲染

推荐：小型关卡用方案 A，大型开放世界用方案 B
```

**坐标转换公式：**
```gdscript
# 3D 世界坐标 → 2D 小地图坐标
func world_to_minimap(world_pos: Vector3) -> Vector2:
    var map_pos: Vector2 = Vector2.ZERO
    map_pos.x = (world_pos.x - map_origin.x) / map_scale * minimap_size.x
    map_pos.y = (world_pos.z - map_origin.z) / map_scale * minimap_size.y
    return map_pos
```

---

## 五、节点选择速查表

### 5.1 按功能选择

| 功能需求 | 3D 节点 | 2D 节点 | 选择依据 |
|----------|---------|---------|----------|
| 可控角色 | CharacterBody3D | CharacterBody2D | 项目主维度 |
| 静态障碍 | StaticBody3D | StaticBody2D | 不会移动的碰撞体 |
| 物理物体 | RigidBody3D | RigidBody2D | 需要物理模拟 |
| 区域检测 | Area3D | Area2D | 触发器、伤害区域 |
| 射线检测 | RayCast3D | RayCast2D | 视线、地面检测 |
| 导航代理 | NavigationAgent3D | NavigationAgent2D | AI 寻路 |
| 路径跟随 | PathFollow3D | PathFollow2D | 沿路径移动 |
| 动画播放 | AnimationPlayer | AnimationPlayer | 通用，2D/3D 共用 |
| 动画树 | AnimationTree | AnimationTree | 复杂动画混合 |
| 粒子效果 | GPUParticles3D | GPUParticles2D | 视觉特效 |
| 音频 | AudioStreamPlayer3D | AudioStreamPlayer2D | 空间音频 |
| 光源(点) | OmniLight3D | PointLight2D | 局部照明 |
| 光源(方向) | DirectionalLight3D | DirectionalLight2D | 全局照明 |
| 相机 | Camera3D | Camera2D | 视角控制 |
| 骨骼动画 | Skeleton3D | Skeleton2D | 骨骼驱动变形 |
| 地形 | MeshInstance3D / CSG | TileMap | 关卡地形 |

### 5.2 决策流程图

```
需要创建新内容？
    │
    ├── 是 UI 界面？ → 使用 Control 节点（始终 2D）
    │
    ├── 是游戏世界内容？
    │   ├── 项目主维度是 3D？ → 使用 3D 节点
    │   ├── 项目主维度是 2D？ → 使用 2D 节点
    │   └── 混合项目？ → 根据该内容的空间需求决定
    │
    ├── 是特效？
    │   ├── 需要深度/遮挡？ → GPUParticles3D / Sprite3D
    │   └── 全屏/叠加？ → CanvasLayer + 2D 节点
    │
    └── 是音频？
        ├── 需要空间衰减？ → AudioStreamPlayer3D/2D
        └── 全局音频？ → AudioStreamPlayer（无空间）
```

### 5.3 常见错误与修正

| 错误 | 正确做法 | 原因 |
|------|----------|------|
| 在 3D 场景中用 Sprite2D 做血条 | 用 SubViewport+Sprite3D 或 CanvasLayer | Sprite2D 不参与 3D 深度排序 |
| 用 RigidBody3D 做角色 | 用 CharacterBody3D | RigidBody 难以精确控制 |
| 在 2D 游戏中用 Camera3D | 用 Camera2D | Camera3D 在 2D 场景中无效 |
| UI 放在游戏世界节点下 | UI 放在 CanvasLayer 下 | 避免被游戏世界变换影响 |
| 用 Area3D 做地面碰撞 | 用 StaticBody3D | Area 不阻挡物体 |
