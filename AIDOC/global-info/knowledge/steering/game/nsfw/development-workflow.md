# NSFW 开发工作流补充

> **本文件是 `development-workflow.md` 的 NSFW 扩展**。
> 仅当 `clarification.md` 声明项目含 NSFW/R-18 内容时生效。
> 通用工作流规范请参见 `../development-workflow.md`。

---

## Phase 0: 项目初始化 — NSFW 特有产出物

### 创意轨道

| 产出物 | 格式 | 验收标准 |
|--------|------|----------|
| 合规声明 | clarification.md NSFW 节 | 目标分级（R-18/R-18G）、目标平台（Steam/DLsite/itch.io）、All-Ages 版策略已确认 |
| 版本发布策略 | clarification.md NSFW 节 | Base Game + Adult DLC 拆分方案已确认，参考 `multi-version-strategy.md` |

### 验收标准

- [ ] clarification.md 中已填写目标平台合规声明
- [ ] clarification.md 中已填写 All-Ages vs R-18 版本发布策略

---

## Phase 3: 核心系统实现 — NSFW 特有产出物

### 技术轨道

| 产出物 | 格式 | 验收标准 |
|--------|------|----------|
| 关系数值系统 | 完整实现 | 参考 `relationship-stats.md`；好感/堕落/羞耻等数值可读写、信号可触发 |
| H-Scene 系统骨架 | 骨架实现 | 参考 `h-scene-system.md`；触发器合规检查、阶段状态机、DLC 门控、fallback 分支全部到位 |
| CG Gallery 系统 | 完整实现 | 参考 `cg-gallery.md`；解锁记录独立存档、UI 可正常浏览 |
| 服装状态机 | 完整实现 | 各服装层切换可用；H-Scene 触发时状态可读取 |
| DLC 动态挂接 | 完整实现 | `DLCManager` 可检测 Adult DLC 状态；NSFW 内容入口统一走 DLC 门控 |

### 验收标准

- [ ] 关系数值系统所有 stat 可正确计算 + 触发信号
- [ ] H-Scene 触发器合规检查通过（角色年龄声明验证）
- [ ] All-Ages 模式下 H-Scene 触发点正确 fallback，无报错

---

## Phase 4: 内容填充 — NSFW 特有产出物

### 创意轨道

| 产出物 | 格式 | 验收标准 |
|--------|------|----------|
| H-Scene 内容设计 | design/h-scenes/ | 每个 H-Scene 的触发条件、阶段设计、角色对话脚本完整；每个场景角色合规声明已确认 |
| CG 清单 | design/cg-list.md | 所有 CG 编号、触发场景、角色、内容标签完整登记 |
| NSFW 音频需求 | design/audio.md NSFW 节 | 每个 H-Scene 阶段所需语音状态机规格（喘息/呻吟/高潮语音库阶段切换）已定义 |

### 技术轨道

| 产出物 | 格式 | 验收标准 |
|--------|------|----------|
| H-Scene 内容实现 | H-Scene 场景目录 | 所有 H-Scene 触发器、动画、语音、CG 解锁集成完毕；合规检查通过 |
| NSFW 音频集成 | 音频资产/adult/ | 语音库按阶段状态机绑定；All-Ages/Adult 音频包切换可用 |
| 隐私功能实现 | 设置/全局单例 | 老板键、默认禁用截图 API、存档加密已实现 |
| 马赛克 Shader | 渲染/Shader 目录 | 实时马赛克 Shader 接入 CensorshipManager；日区自动打码验证通过 |

### 验收标准

- [ ] 所有 H-Scene 触发条件验证通过（包含合规检查）
- [ ] CG Gallery 所有 CG 解锁逻辑覆盖完整
- [ ] Adult DLC 禁用时，全年龄模式无 NSFW 内容泄露

---

## Phase 5: 打磨优化 — NSFW 特有产出物

### 技术轨道

| 产出物 | 格式 | 验收标准 |
|--------|------|----------|
| 多版本发布包 | export/（各版本子目录） | 参考 `multi-version-strategy.md` 打包矩阵完整输出 |
| 合规文档包 | compliance-docs/ | 角色年龄声明、内容审核日志、平台申报记录完整归档 |

### 验收标准

- [ ] All-Ages 版可在 Steam 普通标签下正常上架（通过平台内容审核）
- [ ] Adult DLC 版通过目标平台（Steam AO / itch.io / DLsite）成人内容审核
- [ ] 日本版马赛克验证通过
- [ ] 合规文档包完整，可应对平台投诉
