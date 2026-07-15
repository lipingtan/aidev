# 需求文档：3D TPS 手机射击游戏（致命解药风格）

## 引言

本文档定义一款面向安卓平台的第三人称射击（TPS）手机游戏的完整需求，涵盖从概念图生成到可操作游戏的全流程。游戏玩法参照"致命解药"（Plague Inc: The Cure 衍生的生存射击类型），玩家在末日瘟疫背景下操作角色进行射击战斗与生存。

工具链：Venice.ai → Character Creator 5.10 → Substance 3D Sampler → Substance 3D Painter → Blender 4.5 → Godot 4.5

目标平台：Android（中端 SoC 及以上，骁龙 7+ Gen 2 / 天玑 8200 级）

## 术语表

- **Pipeline**：从概念图到引擎集成的完整资产制作管线
- **Venice_AI**：AI 图像生成工具，用于概念图和参考图生成
- **CC5**：Character Creator 5.10，人体与面部建模工具
- **Sampler**：Substance 3D Sampler，材质采样与生成工具
- **Painter**：Substance 3D Painter，PBR 材质绘制工具
- **Blender**：Blender 4.5，3D 建模、拓扑、UV、绑定、动画工具
- **Engine**：Godot 4.5 游戏引擎
- **TPS_Camera**：第三人称过肩射击相机系统
- **Touch_Controller**：触屏虚拟摇杆与按钮控制系统
- **LOD_System**：多级细节系统，根据距离切换模型精度
- **SSS_Shader**：次表面散射着色器，用于真人化皮肤渲染
- **AnimationTree**：Godot 动画混合树，处理上下半身分层动画
- **Forward_Mobile**：Godot 的 Forward Mobile 渲染后端
- **Performance_Monitor**：性能监控系统，实时追踪帧率与内存

## 需求

### 需求 1：概念图与参考资料生成

**用户故事：** 作为美术制作人员，我希望通过 AI 生成高质量角色概念图和参考资料，以便为后续建模提供准确的视觉基准。

#### 验收标准

1. WHEN 美术人员提供角色描述文本, THE Venice_AI SHALL 生成不少于 3 张面部多角度参考图（正面、侧面、3/4 角度），分辨率不低于 2048×2048 像素
2. WHEN 美术人员提供体型描述, THE Venice_AI SHALL 生成不少于 2 张全身参考图（T-Pose 或 A-Pose），人体比例符合写实标准
3. WHEN 美术人员提供皮肤细节需求, THE Venice_AI SHALL 生成不少于 2 张皮肤微距参考图，包含毛孔、雀斑、血管等微观细节
4. WHEN 美术人员提供装备风格描述, THE Venice_AI SHALL 生成不少于 3 张战术装备概念图，材质层次分明
5. THE Pipeline SHALL 将所有概念图产出存储至 `assets_source/{游戏名}/concept/character/` 目录，按角度和类型分文件夹归档

### 需求 2：基础人体与面部建模

**用户故事：** 作为美术制作人员，我希望利用 CC5 生成高精度真人化基础人体模型，以便获得符合写实标准的角色网格。

#### 验收标准

1. WHEN 概念参考图就绪, THE CC5 SHALL 基于参考图生成基础人体模型，面部形态与参考图匹配度经人工确认
2. THE CC5 SHALL 导出包含 Morph Targets（BlendShapes）的 FBX 文件，格式为 Binary FBX
3. THE CC5 SHALL 导出不低于 4096×4096 分辨率的基础贴图集（Diffuse、Normal、Specular），格式为 PNG
4. WHEN 导出 FBX 时, THE CC5 SHALL 包含标准 Humanoid 骨骼和面部骨骼数据
5. THE Pipeline SHALL 验证导出的 FBX 文件在 Blender 4.5 中可正确导入，缩放因子为 0.01（cm 转 m）

### 需求 3：拓扑、UV 与细节雕刻

**用户故事：** 作为技术美术，我希望将高模转换为移动端可用的游戏模型，以便在性能预算内保留最大视觉细节。

#### 验收标准

1. WHEN CC5 高模导入 Blender, THE Blender SHALL 生成游戏用低模，移动档面数不超过 15000 三角面
2. THE Blender SHALL 为面部区域保留高密度环线拓扑，面部 UV 岛占总 UV 面积不低于 30%
3. WHEN UV 展开完成, THE Blender SHALL 使用 UVPackMaster3 优化排布，UV 填充率不低于 90%，无 UV 重叠
4. THE Blender SHALL 创建双通道 UV（UV1 用于 Diffuse/Normal，UV2 用于 Lightmap），UV2 无重叠
5. WHEN 高低模均就绪, THE Blender SHALL 烘焙 Tangent Space 法线贴图，分辨率为 4096×4096
6. THE Blender SHALL 烘焙 AO 贴图和 Curvature 贴图，供后续材质制作使用
7. THE Pipeline SHALL 全程使用 OpenGL（Y+）法线空间，与 CC5、Painter、Godot 保持一致

### 需求 4：皮肤材质制作

**用户故事：** 作为材质美术，我希望制作照片级真人皮肤 PBR 材质，以便角色在引擎中呈现逼真的皮肤效果。

#### 验收标准

1. WHEN 真实皮肤照片导入 Sampler, THE Sampler SHALL 生成去光照的 Base Color、Normal、Roughness 皮肤基础贴图
2. THE Sampler SHALL 导出可复用的 Substance Material（.sbsar 格式），供 Painter 直接调用
3. WHEN 低模和烘焙贴图导入 Painter, THE Painter SHALL 完成包含以下通道的完整 PBR 贴图集：Albedo、Normal、Roughness、AO、SSS Color、Thickness
4. THE Painter SHALL 按区域差异规范绘制皮肤材质（额头、鼻子、脸颊、下巴、耳朵、嘴唇、脖子、手部、身体各区域的 BaseColor/Roughness/Normal/SSS 参数不同）
5. WHEN 导出贴图时, THE Painter SHALL 导出两套分辨率：桌面档 4096×4096、移动档 2048×2048，格式为 PNG
6. THE Painter SHALL 使用 OpenGL（Y+）法线空间导出法线贴图，通道映射对应 Godot PBR 标准（Normal = OpenGL Y+）

### 需求 5：骨骼绑定与动画准备

**用户故事：** 作为技术美术，我希望为角色创建支持 TPS 射击的骨骼绑定系统，以便角色能正确执行移动和射击动画。

#### 验收标准

1. THE Blender SHALL 创建符合 Godot Humanoid 标准命名的骨骼系统，总骨骼数不超过 50 根（移动档）
2. THE Blender SHALL 实现上下半身分离绑定，分离点在 Spine2，支持上半身独立瞄准旋转
3. WHEN 骨骼创建完成, THE Blender SHALL 设置 IK 控制器（HandIK_L、HandIK_R、FootIK_L、FootIK_R、AimTarget）
4. THE Blender SHALL 完成身体自动权重绑定并手动修正关节区域，确保基础动作（Idle、Walk、Run、Aim、Shoot）无穿模和扭曲
5. WHEN 绑定完成, THE Blender SHALL 导出 glTF 2.0（.glb）格式模型，包含 Mesh、Armature、Shape Keys、Skinning 数据
6. THE Blender SHALL 使用 +Y Up、-Z Forward 坐标系导出，与 Godot 默认坐标系一致

### 需求 6：引擎集成与 TPS 系统

**用户故事：** 作为游戏开发者，我希望在 Godot 中完成角色集成和 TPS 射击系统搭建，以便获得可操作的游戏原型。

#### 验收标准

1. WHEN GLB 模型导入 Engine, THE Engine SHALL 正确解析骨骼、网格和 UV 数据，无报错
2. THE Engine SHALL 实现自定义 SSS_Shader，接收 Albedo、Normal、Roughness、AO、SSS Color、Thickness 共 6 张贴图输入
3. THE Engine SHALL 实现 TPS_Camera 系统（基于 SpringArm3D），正常模式角色居中偏左，ADS 瞄准模式切换为右肩过肩视角
4. THE Engine SHALL 实现基于 CharacterBody3D 的角色控制器，支持移动、冲刺、瞄准、射击四种基础状态
5. THE Engine SHALL 实现射击系统，包含 Raycast 命中检测、后坐力反馈、弹着点显示
6. THE Engine SHALL 配置 AnimationTree 实现上下半身分层混合，上半身处理瞄准/射击动画，下半身处理移动动画
7. WHEN 所有系统集成完毕, THE Engine SHALL 生成可在编辑器中运行的完整 TPS 射击原型场景

### 需求 7：致命解药风格玩法系统

**用户故事：** 作为游戏设计师，我希望实现参照致命解药的末日生存射击玩法，以便玩家在瘟疫末日场景中进行紧张的战斗与生存。

#### 验收标准

1. THE Engine SHALL 实现瘟疫感染区域系统，玩家进入感染区域时持续受到伤害，感染区域随时间扩张
2. THE Engine SHALL 实现感染度机制，角色持有感染度属性（0-100），达到 100 时角色死亡
3. WHEN 玩家击杀感染体敌人, THE Engine SHALL 减少玩家角色的感染度（每次击杀减少固定值）
4. THE Engine SHALL 实现波次敌人生成系统，每波敌人数量和强度递增，波次间有短暂安全间隔
5. THE Engine SHALL 实现资源拾取系统，包含弹药、医疗包（降低感染度）、武器升级组件三类拾取物
6. THE Engine SHALL 实现简易武器系统，支持至少 2 种武器类型（突击步枪、霰弹枪），可切换
7. WHEN 所有存活玩家死亡或感染度达到 100, THE Engine SHALL 触发游戏结束流程，显示存活波次数和击杀统计

### 需求 8：安卓平台性能优化

**用户故事：** 作为游戏开发者，我希望游戏在安卓中端设备上稳定运行 60fps，以便提供流畅的游戏体验。

#### 验收标准

1. THE Engine SHALL 使用 Forward_Mobile 渲染后端进行安卓导出
2. THE Engine SHALL 实现 LOD_System，角色模型支持至少 2 级 LOD（LOD0=15K 三角面，LOD1=5K 三角面）
3. THE Engine SHALL 将移动档单角色贴图总 VRAM 控制在 24MB 以内
4. THE Engine SHALL 将移动档单角色 Draw Call 控制在 3 次以内
5. WHILE 游戏运行在目标设备上, THE Engine SHALL 维持平均帧率不低于 60fps（1080p 分辨率）
6. THE Engine SHALL 将移动档场景纹理总 VRAM 控制在 350MB 以内
7. THE Engine SHALL 将移动档同屏骨骼角色数量限制在 8 个以内
8. THE Engine SHALL 关闭 SSAO、SSIL、Glow 等屏幕空间效果（移动档），使用简化的 SSS 或关闭 SSS
9. IF 帧率连续 3 秒低于 50fps, THEN THE Performance_Monitor SHALL 自动降低粒子数量和阴影距离

### 需求 9：触屏操作适配

**用户故事：** 作为手机玩家，我希望通过触屏虚拟控件流畅操作角色进行移动和射击，以便获得舒适的手机游戏体验。

#### 验收标准

1. THE Touch_Controller SHALL 在屏幕左侧提供虚拟摇杆，控制角色移动方向和速度
2. THE Touch_Controller SHALL 在屏幕右侧提供触摸区域，控制相机旋转（视角移动）
3. THE Touch_Controller SHALL 提供射击按钮，位于屏幕右下区域，支持长按连续射击
4. THE Touch_Controller SHALL 提供瞄准按钮（ADS 切换），位于射击按钮左侧
5. THE Touch_Controller SHALL 提供冲刺按钮，位于虚拟摇杆附近，双击摇杆方向触发冲刺
6. THE Touch_Controller SHALL 提供武器切换按钮和换弹按钮，位于屏幕右侧中部区域
7. WHEN 玩家触摸虚拟摇杆, THE Touch_Controller SHALL 在 16ms 内响应输入并更新角色移动状态
8. THE Touch_Controller SHALL 支持多点触控，允许同时移动、旋转视角和射击
9. THE Touch_Controller SHALL 提供 UI 布局自定义功能，玩家可调整各按钮的位置和大小
10. IF 设备屏幕分辨率或比例发生变化, THEN THE Touch_Controller SHALL 自适应调整控件布局，确保操作区域不超出屏幕边界

### 需求 10：工具链衔接与管线验证

**用户故事：** 作为项目负责人，我希望各阶段工具之间的数据交接无损且可验证，以便避免返工和数据丢失。

#### 验收标准

1. THE Pipeline SHALL 在 Venice_AI 产出概念图后验证：分辨率不低于 2048×2048、角度覆盖齐全（正面/侧面/3-4 角度）
2. THE Pipeline SHALL 在 CC5 导出后验证：FBX 文件在 Blender 中可正确导入，单位缩放 0.01 后尺寸正常
3. THE Pipeline SHALL 在 Blender 拓扑完成后验证：面数符合移动档预算（≤15K tri）、UV 填充率≥90%、无 UV 重叠
4. THE Pipeline SHALL 在 Painter 导出后验证：所有贴图通道齐全（6-7 张）、法线空间为 OpenGL Y+、分辨率匹配目标档位
5. THE Pipeline SHALL 在 Blender 绑定导出后验证：GLB 文件在 Godot 中可正确导入、骨骼命名符合 Humanoid 标准、无报错
6. THE Pipeline SHALL 在 Engine 集成后验证：角色可操作、贴图无错位、动画无穿模
7. THE Pipeline SHALL 使用非破坏性工作流，各阶段保留原始项目文件（CC5 项目、Painter 分层文件、Blender 修改器栈），支持任意阶段回溯修改
8. WHEN 任一阶段的 UV 布局发生修改, THE Pipeline SHALL 重新执行该阶段之后的所有烘焙和贴图绘制步骤

### 需求 11：敌人 AI 系统

**用户故事：** 作为游戏设计师，我希望感染体敌人具备基础 AI 行为，以便为玩家提供有挑战性的战斗体验。

#### 验收标准

1. THE Engine SHALL 实现敌人状态机，包含 Idle、Patrol、Chase、Attack、Death 五种状态
2. WHEN 玩家进入敌人警戒距离（5-10 单位）, THE Engine SHALL 触发敌人从 Idle/Patrol 切换为 Chase 状态
3. WHEN 敌人追击距离超过 15 单位, THE Engine SHALL 令敌人放弃追击返回巡逻状态
4. THE Engine SHALL 限制同时攻击玩家的敌人数量不超过 3 个，其余敌人保持包围待命状态
5. WHEN 敌人发起攻击, THE Engine SHALL 播放 0.3-0.8 秒的攻击前摇动画，给予玩家反应时间
6. THE Engine SHALL 实现至少 2 种敌人类型：近战感染体（快速冲锋）和远程感染体（投射物攻击）

### 需求 12：HUD 与游戏界面

**用户故事：** 作为手机玩家，我希望在游戏中清晰看到关键信息，以便快速做出决策。

#### 验收标准

1. THE Engine SHALL 在屏幕顶部显示玩家血量条和感染度条
2. THE Engine SHALL 在屏幕右下方显示当前武器图标、弹匣余量和备弹量
3. THE Engine SHALL 在屏幕中央显示准星 UI，瞄准模式下准星收缩
4. WHEN 新一波敌人生成, THE Engine SHALL 在屏幕中央显示波次提示（停留 2-3 秒后消失）
5. THE Engine SHALL 在屏幕左上角显示当前波次数和存活时间
6. WHEN 玩家受到伤害, THE Engine SHALL 在受击方向显示红色方向指示（持续 0.5-1.0 秒）
7. THE Engine SHALL 确保所有 HUD 元素在不同屏幕比例下正确适配，不遮挡关键游戏画面
