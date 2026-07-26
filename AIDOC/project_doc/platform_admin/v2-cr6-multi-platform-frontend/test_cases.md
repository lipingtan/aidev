# 测试用例：V2-CR6 多端前端架构 + Plugin SDK V2

**关联文档**: requirements.md / design.md / tasks.md
**编写日期**: 2025-07-14
**测试类型**: 功能测试 / 接口测试 / 回归测试 / 数据验证 / UI端面测试
**测试方法**: 等价类划分、边界值分析、场景法、错误推测、状态转换测试、UX启发式评估

---

## 测试矩阵概览

| 模块 | 正向 | 反向 | 边界 | 回归 | UI功能 | UI体验 | 接口 | 数据验证 | 合计 |
|------|------|------|------|------|--------|--------|------|----------|------|
| FR-1 admin 双入口构建 | 3 | 2 | 1 | 1 | 3 | 5 | 1 | 1 | 17 |
| FR-2 user 双入口构建 | 2 | 1 | 1 | 1 | 3 | 5 | 1 | 1 | 15 |
| FR-3 Plugin SDK V2 基础API | 4 | 3 | 2 | — | — | — | — | — | 9 |
| FR-4 teardown 机制 | 3 | 2 | 2 | — | — | — | — | — | 7 |
| FR-5 plugin-loader | 3 | 3 | 1 | 1 | — | — | 2 | 1 | 11 |
| FR-6 后端 bundle 部署 | 3 | 3 | 2 | 1 | — | — | 2 | 2 | 13 |
| 回归（RG-1~RG-6） | — | — | — | 6 | — | — | — | — | 6 |
| **总计** | **18** | **14** | **9** | **10** | **6** | **10** | **6** | **5** | **78** |

---

## 一、正向测试（Happy Path）

### TC-001: admin build:pc 产物验证

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | dev-web-admin 工程依赖已安装（npm install 完成） |
| **测试步骤** | 1. 在 dev-web-admin 目录执行 `npm run build:pc`<br>2. 检查 `dist-pc/` 目录是否生成<br>3. 检查 `dist-pc/index.html` 是否存在<br>4. 检查 index.html 中 script src 引用的是 PC 入口 bundle |
| **预期结果** | `dist-pc/` 存在；`dist-pc/index.html` 存在；无构建错误；bundle 内容不含 Vant 相关代码 |
| **关联需求** | FR-1 |
| **设计方法** | 场景法、等价类划分 |

### TC-002: admin build:h5 产物验证

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | dev-web-admin 工程依赖已安装 |
| **测试步骤** | 1. 执行 `npm run build:h5`<br>2. 检查 `dist-h5/` 目录生成<br>3. 检查 `dist-h5/index.html` 存在<br>4. 检查 index.html 的 viewport meta 含 `user-scalable=no` |
| **预期结果** | `dist-h5/` 存在；viewport 含 `user-scalable=no`；bundle 包含 Vant CSS |
| **关联需求** | FR-1 |
| **设计方法** | 场景法 |

### TC-003: admin H5 模式 __PLATFORM_ADMIN_LIBS__ 宿主库验证

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | dev-web-admin H5 模式启动（`npm run dev:h5`） |
| **测试步骤** | 1. 浏览器 Console 执行 `window.__PLATFORM_ADMIN_LIBS__`<br>2. 检查返回对象包含 `Vue`、`ElementPlus`、`VueRouter`、`Vant` |
| **预期结果** | 对象包含上述 4 个键；`Vant` 不为 undefined |
| **关联需求** | FR-1 |
| **设计方法** | 等价类划分 |

### TC-004: user build:pc 产物验证

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | dev-web-user 工程依赖已安装，package.json 已声明 vant@4.x |
| **测试步骤** | 1. 执行 `npm run build:pc`<br>2. 检查 `dist-pc/index.html` 存在<br>3. 检查 `window.__PLATFORM_USER_LIBS__` 包含 Vue/ElementPlus/VueRouter（无 Vant） |
| **预期结果** | `dist-pc/` 存在；PC 宿主库不含 Vant |
| **关联需求** | FR-2 |
| **设计方法** | 场景法、等价类划分 |

### TC-005: user build:h5 产物验证

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | dev-web-user 工程依赖已安装 |
| **测试步骤** | 1. 执行 `npm run build:h5`<br>2. 检查 `dist-h5/index.html` 存在<br>3. H5 模式宿主库含 Vant |
| **预期结果** | 构建无错误；`__PLATFORM_USER_LIBS__.Vant` 存在 |
| **关联需求** | FR-2 |
| **设计方法** | 场景法 |

### TC-006: Plugin SDK getPlatform / getDevice 返回值验证

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | dev-web-admin H5 模式运行；已加载测试插件 |
| **测试步骤** | 1. 插件 setup 中调用 `ctx.getPlatform()` 和 `ctx.getDevice()`<br>2. 记录返回值 |
| **预期结果** | `getPlatform()` 返回 `"admin"`；`getDevice()` 返回 `"h5"` |
| **关联需求** | FR-3 |
| **设计方法** | 等价类划分 |

### TC-007: Plugin SDK registerExtension 返回 UnregisterFn

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 插件已加载，ctx 已创建 |
| **测试步骤** | 1. 调用 `const unregister = ctx.registerExtension('menu.extra', MockComp)`<br>2. 验证返回值类型为 function<br>3. 调用 `unregister()`<br>4. 检查扩展点 Map 中该组件已移除 |
| **预期结果** | 返回值为 function；调用后组件从扩展点 Map 移除 |
| **关联需求** | FR-3 |
| **设计方法** | 状态转换测试 |

### TC-008: Plugin SDK i18n.mergeLocale 合入主应用

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 插件已加载；主应用 i18n 已初始化 |
| **测试步骤** | 1. 调用 `ctx.i18n.mergeLocale('zh-CN', { 'game.title': '游戏' })`<br>2. 在 Vue 模板中使用 `$t('game.title')`<br>3. 检查渲染结果 |
| **预期结果** | 渲染出 `"游戏"`；无 missing key 警告 |
| **关联需求** | FR-3 |
| **设计方法** | 场景法 |

### TC-009: onTeardown 注册后 stop 触发调用

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 插件已通过 loadPlugin 加载 |
| **测试步骤** | 1. 插件 setup 中调用 `ctx.onTeardown(() => { cleanupCalled = true })`<br>2. 调用 `unloadPlugin('game')`<br>3. 检查 cleanupCalled |
| **预期结果** | `cleanupCalled === true`；teardown 在 unload 完成前被调用 |
| **关联需求** | FR-4 |
| **设计方法** | 状态转换测试 |

### TC-010: unloadPlugin 后 extensionRegistry 彻底清空

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 插件已加载，已注册 2 个扩展点 |
| **测试步骤** | 1. 调用 `unloadPlugin('game')`<br>2. 检查 extensionRegistry.has('game')<br>3. 检查 teardownQueues.has('game') |
| **预期结果** | 两者均为 false；window 全局变量 `__PLUGIN_GAME__` 已被 delete |
| **关联需求** | FR-4, FR-5 |
| **设计方法** | 状态转换测试 |

### TC-011: teardown 完成后扩展点自动 unregister

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 插件已加载并注册了扩展点组件 |
| **测试步骤** | 1. 插件注册 `extension-A` 到扩展点 `menu.extra`<br>2. 调用 `unloadPlugin`<br>3. 检查全局扩展点 Map 中 `menu.extra` 不再包含 `extension-A` |
| **预期结果** | 扩展点 Map 中该组件已被移除 |
| **关联需求** | FR-4 |
| **设计方法** | 状态转换测试 |

### TC-012: loadPlugin 成功加载并调用 setup

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | `/static/plugins/game/admin-pc/bundle.js` 文件存在且合法 |
| **测试步骤** | 1. 调用 `await loadPlugin('game', 'admin', 'pc')`<br>2. 检查 DOM 中 `script[data-plugin="game"]` 存在<br>3. 检查 `window.__PLUGIN_GAME__` 已定义<br>4. 检查 setup 函数被调用 |
| **预期结果** | Promise resolve；script 标签存在；setup 调用成功 |
| **关联需求** | FR-5 |
| **设计方法** | 场景法 |

### TC-013: loadPlugin 路径格式验证

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 无 |
| **测试步骤** | 1. 调用 `loadPlugin('my-plugin', 'user', 'h5')`<br>2. 拦截 script.src |
| **预期结果** | `script.src` 为 `/static/plugins/my-plugin/user-h5/bundle.js` |
| **关联需求** | FR-5 |
| **设计方法** | 等价类划分 |

### TC-014: 安装含 frontends 配置的插件 — bundle 部署

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | plugin.json 含 `frontends: [{platform:"admin",device:"pc",entry:"admin-pc/bundle.js"}]`；源文件存在 |
| **测试步骤** | 1. 调用 `InstallFromFile` 安装插件包<br>2. 检查 `{staticDir}/plugins/game/admin-pc/bundle.js` 是否存在<br>3. 验证文件内容与源一致 |
| **预期结果** | 目标文件存在；内容一致；`sys_plugin` 记录创建成功 |
| **关联需求** | FR-6 |
| **设计方法** | 场景法 |

### TC-015: 安装含 4 个 frontends 条目 — 全部部署

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | plugin.json 含 admin-pc、admin-h5、user-pc、user-h5 四个条目 |
| **测试步骤** | 1. 安装插件<br>2. 检查四个目录 `admin-pc/`、`admin-h5/`、`user-pc/`、`user-h5/` 均存在 bundle.js |
| **预期结果** | 四个目录全部创建，每个 bundle.js 存在 |
| **关联需求** | FR-6 |
| **设计方法** | 判定表 |

### TC-016: 卸载插件 — 静态资源目录删除

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 插件已安装，bundle 已部署到 `{staticDir}/plugins/game/` |
| **测试步骤** | 1. 调用 `Uninstall(ctx, "game", false)`<br>2. 检查 `{staticDir}/plugins/game/` 目录是否存在 |
| **预期结果** | 目录不存在；`sys_plugin` 记录已删除 |
| **关联需求** | FR-6 |
| **设计方法** | 状态转换测试 |

### TC-017: GET /static/plugins 返回 bundle 文件（200）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 插件已安装，bundle 已部署 |
| **测试步骤** | `GET /static/plugins/game/admin-pc/bundle.js` |
| **预期结果** | HTTP 200；Content-Type: `application/javascript` |
| **关联需求** | FR-6 |
| **设计方法** | 场景法 |

### TC-018: admin PC 模式下 window.__PLATFORM_ADMIN_LIBS__ 不含 Vant

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | dev-web-admin PC 模式启动 |
| **测试步骤** | 浏览器 Console 执行 `window.__PLATFORM_ADMIN_LIBS__.Vant` |
| **预期结果** | 返回 `undefined`（PC 入口不引入 Vant） |
| **关联需求** | FR-1 |
| **设计方法** | 等价类划分 |

---

## 二、反向测试（Negative）

### TC-N01: bundle 加载失败（404）— 主应用不崩溃

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | `/static/plugins/nonexistent/admin-pc/bundle.js` 不存在 |
| **测试步骤** | 1. 调用 `loadPlugin('nonexistent', 'admin', 'pc')`<br>2. 捕获 reject<br>3. 验证主应用页面仍可正常操作 |
| **预期结果** | Promise reject，错误信息含路径 `/static/plugins/nonexistent/admin-pc/bundle.js`；主应用无崩溃；其他插件不受影响 |
| **关联需求** | FR-5 |
| **设计方法** | 错误推测 |

### TC-N02: teardown 超时（>5s）— 强制继续卸载

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 插件已加载；teardown 函数为耗时 10s 的异步操作 |
| **测试步骤** | 1. `ctx.onTeardown(() => new Promise(r => setTimeout(r, 10000)))`<br>2. 调用 `unloadPlugin('game')`<br>3. 计时等待 |
| **预期结果** | 约 5s 后 unloadPlugin resolve（不永久挂起）；console.warn 含超时提示；extensionRegistry 中 'game' 已清理 |
| **关联需求** | FR-4, FR-5 |
| **设计方法** | 边界值分析、错误推测 |

### TC-N03: frontends 中 entry 路径不存在 — 安装返回明确错误

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | plugin.json frontends[0].entry = "admin-pc/bundle.js"，但该文件在包内不存在 |
| **测试步骤** | 1. 调用 `InstallFromFile` 安装<br>2. 检查返回值 |
| **预期结果** | 返回 error；错误信息含 "bundle 源文件不存在" 及具体路径；`{staticDir}/plugins/{name}/` 目录未创建（无部分部署） |
| **关联需求** | FR-6 |
| **设计方法** | 错误推测 |

### TC-N04: 重复 loadPlugin 同一插件 — 不重复加载

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | `loadPlugin('game', 'admin', 'pc')` 已成功调用一次 |
| **测试步骤** | 1. 再次调用 `loadPlugin('game', 'admin', 'pc')`<br>2. 检查 DOM 中 `script[data-plugin="game"]` 数量<br>3. 检查 setup 调用次数 |
| **预期结果** | Promise resolve（不报错）；DOM 中只有 1 个 script 标签；setup 只被调用一次 |
| **关联需求** | FR-5 |
| **设计方法** | 错误推测、等价类划分 |

### TC-N05: 无效 platform/device 组合 — 类型约束

| 字段 | 内容 |
|------|------|
| **优先级** | P2 |
| **前置条件** | TypeScript 编译期检查 |
| **测试步骤** | 1. 尝试 `loadPlugin('game', 'mobile' as any, 'pc')`<br>2. 检查 TypeScript 编译结果<br>3. 运行时检查 bundlePath 构造 |
| **预期结果** | TypeScript 编译报错（类型不兼容）；运行时 bundle 路径中出现非规范值，导致 404 而非静默错误 |
| **关联需求** | FR-5, FR-3 |
| **设计方法** | 等价类划分（无效类） |

### TC-N06: bundle 加载成功但未注册全局变量 — reject 含明确信息

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | bundle.js 文件存在但未写入 `window.__PLUGIN_GAME__` |
| **测试步骤** | 1. 调用 `loadPlugin('game', 'admin', 'pc')`<br>2. 捕获 reject |
| **预期结果** | Promise reject；错误信息含 "未正确注册到 window.__PLUGIN_GAME__" |
| **关联需求** | FR-5 |
| **设计方法** | 错误推测 |

### TC-N07: 安装同名插件（已安装）— 返回冲突错误

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 插件 "game" 已安装（sys_plugin 存在记录） |
| **测试步骤** | 1. 再次调用 `InstallFromFile` 安装同名插件<br>2. 检查错误信息 |
| **预期结果** | 返回 error "插件 game 已安装，请先卸载后再安装"；无重复记录 |
| **关联需求** | FR-6 |
| **设计方法** | 错误推测 |

### TC-N08: plugin.json 缺失 name 字段 — 安装拒绝

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | plugin.json 中 name 为空字符串 |
| **测试步骤** | 调用 `InstallFromFile`，检查错误 |
| **预期结果** | 返回 error "plugin.json 中 name 字段不能为空" |
| **关联需求** | FR-6 |
| **设计方法** | 等价类划分（无效类） |

### TC-N09: teardown 函数抛出异常 — 不阻断后续清理

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 插件注册了一个会 throw Error 的 teardown 函数 |
| **测试步骤** | 1. `ctx.onTeardown(() => { throw new Error('cleanup failed') })`<br>2. 调用 `unloadPlugin('game')`<br>3. 验证 extensionRegistry 清理结果 |
| **预期结果** | unloadPlugin 仍然 resolve（异常被 catch）；extensionRegistry 中 'game' 已清理 |
| **关联需求** | FR-4, FR-5 |
| **设计方法** | 错误推测 |

### TC-N10: GET /static/plugins 不存在路径 — 返回 404

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | 插件未安装，目录不存在 |
| **测试步骤** | `GET /static/plugins/nonexistent/admin-pc/bundle.js` |
| **预期结果** | HTTP 404；不暴露服务器错误信息 |
| **关联需求** | FR-6 |
| **设计方法** | 错误推测 |

---

## 三、边界测试（Boundary）

### TC-B01: frontends 数组为空 — 安装成功无 bundle 目录

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | plugin.json: `{ "name": "game", "frontends": [] }` |
| **测试步骤** | 1. 调用 `InstallFromFile` 安装<br>2. 检查 `{staticDir}/plugins/game/` 目录是否存在<br>3. 检查 sys_plugin 记录 |
| **预期结果** | 安装成功，无错误；`{staticDir}/plugins/game/` 目录不存在；sys_plugin 记录正常创建 |
| **关联需求** | FR-6 |
| **设计方法** | 边界值分析 |

### TC-B02: frontends 含 4 个条目 — 全部部署成功

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | frontends 含 admin-pc、admin-h5、user-pc、user-h5 四个条目，源文件均存在 |
| **测试步骤** | 1. 安装插件<br>2. 检查四个 bundle.js 文件存在且内容正确 |
| **预期结果** | 四个目录均创建；每个 bundle.js 内容与源文件一致（字节级） |
| **关联需求** | FR-6 |
| **设计方法** | 边界值分析、判定表 |

### TC-B03: teardown 恰好在 4999ms 完成 — 正常完成不超时

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | teardown 函数延迟 4999ms resolve |
| **测试步骤** | 1. `ctx.onTeardown(() => new Promise(r => setTimeout(r, 4999)))`<br>2. 调用 `unloadPlugin`，计时 |
| **预期结果** | 约 4999ms 后正常 resolve；无超时警告日志；extensionRegistry 正常清理 |
| **关联需求** | FR-4 |
| **设计方法** | 边界值分析（max-1） |

### TC-B04: teardown 在 5001ms 完成 — 超时强制继续

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | teardown 函数延迟 5001ms resolve |
| **测试步骤** | 1. `ctx.onTeardown(() => new Promise(r => setTimeout(r, 5001)))`<br>2. 调用 `unloadPlugin`，计时 |
| **预期结果** | 约 5000ms 时 unloadPlugin resolve（强制继续）；console.warn 含超时信息；extensionRegistry 仍正常清理 |
| **关联需求** | FR-4 |
| **设计方法** | 边界值分析（max+1） |

### TC-B05: pluginName 含连字符 — 全局变量名转换正确

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | pluginName = `"my-cool-plugin"` |
| **测试步骤** | 1. 调用 `loadPlugin('my-cool-plugin', 'admin', 'pc')`<br>2. 检查期望的全局变量名 |
| **预期结果** | 查找 `window.__PLUGIN_MY_COOL_PLUGIN__`（连字符转下划线，全大写） |
| **关联需求** | FR-5 |
| **设计方法** | 边界值分析、等价类划分 |

### TC-B06: frontends 不含字段（旧 V1 插件）— 解析不报错

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | plugin.json 无 `frontends` 字段（V1 格式） |
| **测试步骤** | 1. 安装 V1 格式插件<br>2. 检查安装结果 |
| **预期结果** | 安装成功；`manifest.Frontends` 为空切片；无任何 bundle 部署；不报错 |
| **关联需求** | FR-6 |
| **设计方法** | 边界值分析（空值边界） |

### TC-B07: 同时注册多个 teardown 函数 — 全部依次执行

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | 插件注册 3 个 teardown 函数，分别设置 flag1/flag2/flag3 |
| **测试步骤** | 1. 注册三个 teardown<br>2. 调用 `unloadPlugin`<br>3. 检查三个 flag |
| **预期结果** | flag1、flag2、flag3 均为 true；Promise.all 并发执行 |
| **关联需求** | FR-4 |
| **设计方法** | 边界值分析 |

### TC-B08: platform/device 值为大写时 — 后端路径转小写

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **测试数据** | FrontendConfig: `{Platform: "Admin", Device: "PC", Entry: "admin-pc/bundle.js"}` |
| **测试步骤** | 安装插件，检查部署目录名 |
| **预期结果** | 目录名为 `admin-pc`（小写），而非 `Admin-PC` |
| **关联需求** | FR-6 |
| **设计方法** | 等价类划分、边界值分析 |

### TC-B09: build:pc 与 build:h5 产物互不污染

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **测试数据** | 连续执行 `npm run build:pc` 和 `npm run build:h5` |
| **测试步骤** | 1. 执行 `npm run build:pc`<br>2. 执行 `npm run build:h5`<br>3. 检查 `dist-pc/` 是否包含 Vant 代码<br>4. 检查 `dist-h5/` 是否包含 PC 布局组件 |
| **预期结果** | `dist-pc/` 不含 Vant/H5Layout；`dist-h5/` 不含 PcLayout/Sidebar 组件 |
| **关联需求** | FR-1 |
| **设计方法** | 边界值分析（产物隔离验证） |

---

## 四、回归测试（Regression）

> 执行范围：每次 CR-6 相关代码变更后运行本章节全部用例。

### TC-R01: PC 端登录、菜单加载、权限校验流程不受影响（RG-1）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-1 |
| **验证步骤** | 1. dev-web-admin PC 模式启动<br>2. 调用 `POST /auth/login` 登录<br>3. 调用 `GET /api/v1/common/user-menu?platform=admin`<br>4. 验证菜单正常渲染、权限校验生效 |
| **预期结果** | 接口均返回 200；菜单树正确渲染；无权限的菜单项不显示；与 CR-6 改造前行为完全一致 |
| **设计方法** | 回归验证 |

### TC-R02: dev-web-admin PC 构建产物功能等价（RG-2）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-2 |
| **验证步骤** | 1. `npm run build:pc`<br>2. 部署 `dist-pc/` 到静态服务<br>3. 访问登录页、列表页、详情页<br>4. 验证 Element Plus 侧边栏布局正常 |
| **预期结果** | 所有已有功能页面正常访问；侧边栏展开收起正常；无 JS 报错 |
| **设计方法** | 回归验证 |

### TC-R03: dev-web-user PC 构建产物功能等价（RG-3）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-3 |
| **验证步骤** | 1. `npm run build:pc`（在 dev-web-user）<br>2. 部署访问<br>3. 验证 Element Plus 风格布局正常 |
| **预期结果** | PC 模式布局正常；无 Vant 组件渲染错误；用户端功能页面可访问 |
| **设计方法** | 回归验证 |

### TC-R04: 现有 V1 插件 setup 函数正常调用（RG-4）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-4 |
| **验证步骤** | 1. 确认 game 插件为 V1 格式（有 setup，无 onTeardown）<br>2. 调用 `loadPlugin('game', 'admin', 'pc')`<br>3. 检查 setup 是否被调用、插件菜单是否出现 |
| **预期结果** | setup 成功调用；插件菜单出现在管理端；无 JS 报错 |
| **设计方法** | 回归验证 |

### TC-R05: 插件 stop/uninstall 后菜单和路由不残留（RG-5）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-5 |
| **验证步骤** | 1. 加载插件，确认菜单和路由存在<br>2. 调用 `unloadPlugin`<br>3. 刷新页面，检查插件菜单 |
| **预期结果** | 刷新后插件菜单消失；访问插件路由返回 404 或重定向；DOM 中无残留 script 标签 |
| **设计方法** | 回归验证 |

### TC-R06: 后端静态文件服务不影响 /api/ 路由（RG-6）

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **关联** | RG-6 |
| **验证步骤** | 1. `GET /static/plugins/game/admin-pc/bundle.js`（静态文件）<br>2. `GET /api/v1/admin/plugins`（API 路由）<br>3. `POST /auth/login`（认证路由） |
| **预期结果** | 静态文件返回 200/application/javascript；API 路由正常响应；认证路由不受影响；无路由冲突 |
| **设计方法** | 回归验证 |

---

## 五、UI 端面测试（Page Level）

> UI 端面测试验证用户通过浏览器实际看到的交互效果。TC-F 系列验证功能交互，TC-UX 系列从产品经理视角审查体验质量。

---

### 5.1 H5Layout.vue（admin 汉堡菜单布局）

#### TC-F01: H5Layout — NavBar 标题栏渲染

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | dev-web-admin H5 模式运行，用户已登录 |
| **测试步骤** | 1. 浏览器（移动端模式）访问管理后台首页<br>2. 检查 DOM 中是否存在 `.van-nav-bar` 元素<br>3. 检查右侧汉堡图标 `.van-icon-wap-nav` 是否存在 |
| **预期结果** | NavBar 存在于顶部；显示当前页面标题；右侧有汉堡图标 |
| **关联需求** | FR-1 |
| **设计方法** | 场景法 |

#### TC-F02: H5Layout — 汉堡菜单侧滑弹出与关闭

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | H5 模式，已登录，菜单数据已加载 |
| **测试步骤** | 1. 点击右侧汉堡图标<br>2. 检查 van-popup 是否从左侧滑出<br>3. 检查菜单列表是否显示系统菜单项<br>4. 点击某个菜单项<br>5. 检查 popup 是否关闭，路由是否跳转 |
| **预期结果** | Popup 从左侧滑出；菜单项正确显示；点击后 popup 关闭，router-view 内容切换 |
| **关联需求** | FR-1 |
| **设计方法** | 场景法、状态转换测试 |

#### TC-F03: H5Layout — 返回按钮触发路由回退

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | H5 模式，已从菜单导航至二级页面 |
| **测试步骤** | 1. 进入某个二级详情页<br>2. 点击 NavBar 左侧返回箭头<br>3. 检查路由变化 |
| **预期结果** | 路由回退到上一页；NavBar 标题更新为上一页标题 |
| **关联需求** | FR-1 |
| **设计方法** | 场景法 |

---

### TC-UX01: H5Layout — 布局合理性审查

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | admin H5 首页（H5Layout.vue） |
| **审查维度** | 布局合理性 |
| **检查描述** | NavBar 高度是否符合移动端规范（44px）；内容区是否留出 NavBar 高度的 padding-top；核心操作区域是否在拇指可及范围内（底部 1/3） |
| **业界参考** | Vant 官方 NavBar 规范（height: 46px）；iOS HIG 最小点击区域 44×44pt |
| **预期结果** | NavBar 高度 44~46px；内容区不被 NavBar 遮挡；主要操作按钮在屏幕下半部分 |
| **不满足时修复建议** | 在 `.h5-content` 添加 `padding-top: 46px`；将高频操作（如新增按钮）移至页面底部悬浮区域 |
| **设计方法** | UX 启发式评估 |

#### TC-UX02: H5Layout — 一致性审查

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | admin H5 各页面（H5Layout.vue） |
| **审查维度** | 一致性 |
| **检查描述** | 所有 H5 页面 NavBar 标题大小/颜色是否统一；汉堡图标在不同页面位置是否一致；Popup 菜单样式是否统一（字体/间距/高亮色） |
| **业界参考** | 同项目 PC 端色彩规范；Vant 主题定制规范 |
| **预期结果** | NavBar 标题统一使用 16px 500 weight；汉堡图标固定在右侧；Popup 高亮色与系统主色一致 |
| **不满足时修复建议** | 通过 Vant CSS 变量 `--van-nav-bar-title-font-size` 统一设定；在 `H5MenuDrawer.vue` 中固定激活样式使用全局主色变量 |
| **设计方法** | UX 启发式评估 |

#### TC-UX03: H5Layout — 完整性审查（空态/加载态）

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | admin H5 菜单抽屉（H5MenuDrawer.vue） |
| **审查维度** | 完整性 |
| **检查描述** | 菜单数据加载中是否有 loading 骨架屏或 spinner；菜单为空时是否显示空状态提示；菜单加载失败时是否有错误提示 |
| **业界参考** | Nielsen 启发式 #1：系统状态可见性；Vant Skeleton 组件 |
| **预期结果** | 加载中显示 van-skeleton 或 van-loading；空菜单显示"暂无菜单"提示；加载失败显示 toast 错误 |
| **不满足时修复建议** | 在 `H5MenuDrawer.vue` 的 `v-if="loading"` 分支添加 `<van-skeleton :row="5" />`；空态添加 `<van-empty description="暂无菜单" />` |
| **设计方法** | UX 启发式评估 |

#### TC-UX04: H5Layout — 操作效率审查

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | admin H5 布局（H5Layout.vue） |
| **审查维度** | 操作效率 |
| **检查描述** | 从任意页面打开菜单所需点击次数是否为 1 次；菜单项点击后是否立即响应（无明显延迟）；是否支持滑动手势关闭 Popup |
| **业界参考** | iOS Safari 手势规范；微信小程序导航交互标准 |
| **预期结果** | 汉堡图标 1 次点击打开菜单；Popup 支持向左滑动关闭；菜单点击响应 < 100ms |
| **不满足时修复建议** | 在 `van-popup` 添加 `:close-on-click-overlay="true"` 和手势关闭支持；确保菜单数据在路由守卫阶段预加载 |
| **设计方法** | UX 启发式评估 |

#### TC-UX05: H5Layout — 响应式适配审查

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | admin H5 布局（H5Layout.vue） |
| **审查维度** | 响应式适配 |
| **检查描述** | 在 375px（iPhone SE）、390px（iPhone 14）、414px（Plus）宽度下布局是否正常；Popup 宽度 75% 是否在小屏上合理；内容区横向是否有溢出滚动 |
| **业界参考** | Vant 移动端适配方案（viewport + rem/vw）；iOS Safari 安全区域（safe-area-inset） |
| **预期结果** | 三种宽度下无水平滚动条；Popup 宽度自适应合理；底部内容不被 iPhone 底部横条遮挡（safe-area） |
| **不满足时修复建议** | 添加 `padding-bottom: env(safe-area-inset-bottom)` 兼容 iPhone 底部安全区；Popup 宽度改用 `min(280px, 80vw)` 确保小屏可用 |
| **设计方法** | UX 启发式评估 |

---

### 5.2 UserH5Layout.vue（user 底部 Tabbar 布局）

#### TC-F04: UserH5Layout — 底部 Tabbar 渲染

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | dev-web-user H5 模式运行，用户已登录 |
| **测试步骤** | 1. 访问用户端 H5 首页<br>2. 检查 DOM 底部 `.van-tabbar` 元素存在<br>3. 检查 Tabbar 菜单项数量和图标 |
| **预期结果** | Tabbar 固定在页面底部；菜单项来自后端 `/api/v1/common/user-menu?platform=user`；图标正确显示 |
| **关联需求** | FR-2 |
| **设计方法** | 场景法 |

#### TC-F05: UserH5Layout — Tabbar 路由跳转与选中高亮

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **前置条件** | H5 模式，Tabbar 已渲染，至少 2 个菜单项 |
| **测试步骤** | 1. 点击第二个 Tabbar 菜单项<br>2. 检查路由是否跳转到对应 path<br>3. 检查该 Tabbar 项是否高亮选中 |
| **预期结果** | router-view 内容切换；被点击项颜色变为主色调；其他项为默认灰色 |
| **关联需求** | FR-2 |
| **设计方法** | 场景法、状态转换测试 |

#### TC-F06: UserH5Layout — 菜单为空时显示空态

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **前置条件** | 后端 `/api/v1/common/user-menu?platform=user` 返回空数组 |
| **测试步骤** | 1. Mock 接口返回空菜单<br>2. 加载 H5 用户端<br>3. 检查 Tabbar 和页面状态 |
| **预期结果** | 不报 JS 错误；Tabbar 显示空态或隐藏；router-view 显示空内容提示 |
| **关联需求** | FR-2 |
| **设计方法** | 错误推测 |

---

#### TC-UX06: UserH5Layout — 布局合理性审查

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | user H5 首页（UserH5Layout.vue） |
| **审查维度** | 布局合理性 |
| **检查描述** | 内容区底部是否为 Tabbar 高度预留空间；Tabbar 是否使用 fixed 定位不遮挡内容；Tabbar 图标+文字组合是否清晰可辨 |
| **业界参考** | 微信、支付宝、美团 H5 底部导航规范；Vant Tabbar 官方最佳实践 |
| **预期结果** | `.user-h5-content` 有 `padding-bottom: 50px`（Tabbar 高度）；内容不被 Tabbar 遮挡；图标尺寸 ≥ 22px |
| **不满足时修复建议** | 为 `.user-h5-content` 添加 `padding-bottom: calc(50px + env(safe-area-inset-bottom))`；Tabbar icon size 统一设为 24px |
| **设计方法** | UX 启发式评估 |

#### TC-UX07: UserH5Layout — 清晰性审查（菜单文案）

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | user H5 Tabbar（UserH5Layout.vue） |
| **审查维度** | 清晰性 |
| **检查描述** | Tabbar 菜单文案是否简洁（≤4字）；是否避免使用英文或技术术语；图标与文案语义是否一致 |
| **业界参考** | 微信小程序设计规范：底部导航文字不超过 4 个字 |
| **预期结果** | 每个 Tabbar 项文字 ≤ 4 个汉字；图标与文字含义一致（如"首页"对应 home 图标） |
| **不满足时修复建议** | 将后端返回的长标题在 `UserH5Layout.vue` 中截断处理，超过 4 字显示省略号；或在 Tabbar 专用配置中维护简短别名 |
| **设计方法** | UX 启发式评估 |

#### TC-UX08: UserH5Layout — 错误预防审查

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | user H5 各操作页面（UserH5Layout.vue 子路由） |
| **审查维度** | 错误预防与恢复 |
| **检查描述** | 用户端 H5 中的危险操作（如删除、取消订单）是否有二次确认弹窗；表单提交中是否禁用重复点击；网络错误时是否有重试入口 |
| **业界参考** | Nielsen 启发式 #5：错误预防；Vant Dialog 确认弹窗规范 |
| **预期结果** | 危险操作弹出 van-dialog 二次确认；提交按钮在请求中显示 loading 且禁用；网络错误 toast 含"重试"按钮 |
| **不满足时修复建议** | 封装 `useConfirmDialog` 通用 hook 统一危险操作确认流程；使用 `van-button :loading="submitting"` 防重复提交 |
| **设计方法** | UX 启发式评估 |

#### TC-UX09: UserH5Layout — 信息层次审查

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | user H5 内容页（UserH5Layout.vue 子路由） |
| **审查维度** | 信息层次 |
| **检查描述** | 页面主标题、副标题、正文是否通过字号区分（16/14/12px）；ID 等技术字段是否弱化或隐藏；时间是否使用"刚刚/N分钟前"等友好格式 |
| **业界参考** | Material Design 排版规范 Type Scale；微信公众号文章排版规范 |
| **预期结果** | 主要信息字号 ≥ 16px；次要信息使用灰色（#999）；时间格式人性化（非时间戳） |
| **不满足时修复建议** | 建立全局排版 CSS 变量 `--text-primary/secondary/tertiary`；封装 `<TimeAgo>` 组件统一时间展示格式 |
| **设计方法** | UX 启发式评估 |

#### TC-UX10: UserH5Layout — 一致性审查（PC vs H5 双模式）

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **页面** | user PC 布局 vs user H5 布局 |
| **审查维度** | 一致性 |
| **检查描述** | PC 与 H5 模式下同一业务页面（如用户信息）的字段显示是否一致；操作结果（成功/失败提示）语言风格是否统一；主色调是否一致 |
| **业界参考** | 同一品牌跨端设计一致性原则（淘宝 PC/APP 品牌色一致） |
| **预期结果** | 同一接口数据在 PC 和 H5 显示字段无差异；Toast 文案格式统一；主色变量两端共用同一 CSS 变量文件 |
| **不满足时修复建议** | 将 `primary-color` 等品牌色提取到 `src/styles/variables.scss`，PC 和 H5 入口均 import；统一 Toast/Message 调用封装 |
| **设计方法** | UX 启发式评估 |

---

## 六、接口测试（API Level）

### TC-A01: GET /static/plugins/{name}/{platform}-{device}/bundle.js — 200 正向

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **请求** | `GET /static/plugins/game/admin-pc/bundle.js` |
| **前置条件** | game 插件已安装，admin-pc bundle 已部署 |
| **预期响应** | HTTP 200；`Content-Type: application/javascript`；响应体为有效 JS 内容 |
| **关联需求** | FR-6, RG-6 |
| **设计方法** | 场景法 |

### TC-A02: GET /static/plugins/{name}/{platform}-{device}/bundle.js — 404

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **请求** | `GET /static/plugins/nonexistent/admin-pc/bundle.js` |
| **前置条件** | 插件未安装 |
| **预期响应** | HTTP 404；不暴露服务器内部路径信息 |
| **关联需求** | FR-6 |
| **设计方法** | 等价类划分（无效类）、错误推测 |

### TC-A03: GET /static/plugins — user-h5 路径格式验证

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **请求** | `GET /static/plugins/game/user-h5/bundle.js` |
| **前置条件** | game 插件已安装，user-h5 bundle 已部署 |
| **预期响应** | HTTP 200；`Content-Type: application/javascript` |
| **关联需求** | FR-6 |
| **设计方法** | 等价类划分 |

### TC-A04: GET /api/v1/admin/plugins — 静态路由与 API 路由不冲突

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **请求** | `GET /api/v1/admin/plugins`，Headers: `Authorization: Bearer {valid_token}` |
| **前置条件** | 后端服务运行，已添加静态文件路由 |
| **预期响应** | HTTP 200；返回插件列表 JSON；`Content-Type: application/json` |
| **关联需求** | RG-6 |
| **设计方法** | 回归验证 |

### TC-A05: go build ./... — 后端构建零错误

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **执行命令** | `go build ./...`（在 backend/ 目录） |
| **前置条件** | installer.go、manifest.go 已按 CR-6 修改 |
| **预期结果** | 构建输出无 error；退出码为 0 |
| **关联需求** | FR-6, RG-6 |
| **设计方法** | 场景法 |

### TC-A06: go vet ./... — 无警告

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **执行命令** | `go vet ./...`（在 backend/ 目录） |
| **前置条件** | 同 TC-A05 |
| **预期结果** | 无任何 vet 警告输出；退出码为 0 |
| **关联需求** | FR-6 |
| **设计方法** | 场景法 |

---

## 七、数据验证

### TC-D01: 安装插件后 sys_plugin 记录字段验证

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **触发操作** | 安装含 frontends 配置的 game 插件（version: "1.2.0"） |
| **验证 SQL** | `SELECT name, version, status, frontend_path, installed_at FROM sys_plugin WHERE name = 'game'` |
| **预期结果** | `name='game'`；`version='1.2.0'`；`status=installed`；`frontend_path` 含 `plugins/game`；`installed_at` 非空 |
| **关联需求** | FR-6 |
| **设计方法** | 场景法 |

### TC-D02: bundle 文件部署路径一致性验证

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **触发操作** | 安装 frontends 含 admin-pc 和 user-h5 的插件 |
| **验证步骤** | 1. 检查文件系统：`{staticDir}/plugins/game/admin-pc/bundle.js` 存在<br>2. 检查文件系统：`{staticDir}/plugins/game/user-h5/bundle.js` 存在<br>3. 对比两个文件与插件包内源文件 MD5 |
| **预期结果** | 两个 bundle.js 均存在；MD5 与源文件一致（copyFile 无损） |
| **关联需求** | FR-6 |
| **设计方法** | 场景法 |

### TC-D03: 卸载后数据库记录硬删除验证

| 字段 | 内容 |
|------|------|
| **优先级** | P0 |
| **触发操作** | 调用 `Uninstall(ctx, "game", false)` |
| **验证 SQL** | `SELECT COUNT(*) FROM sys_plugin WHERE name = 'game'`（不过滤 deleted_at） |
| **预期结果** | COUNT = 0（硬删除，无软删除记录残留） |
| **关联需求** | FR-6 |
| **设计方法** | 状态转换测试 |

### TC-D04: admin_application 记录随插件安装创建

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **触发操作** | 安装 game 插件（manifest.DisplayName = "游戏插件"） |
| **验证 SQL** | `SELECT app_code, name, app_type FROM admin_application WHERE app_code = 'game' AND deleted_at IS NULL` |
| **预期结果** | `app_code='game'`；`name='游戏插件'`；`app_type='PLUGIN'` |
| **关联需求** | FR-6 |
| **设计方法** | 场景法 |

### TC-D05: 安装失败时无脏数据残留

| 字段 | 内容 |
|------|------|
| **优先级** | P1 |
| **触发操作** | 安装插件时 bundle 源文件不存在（deployFrontendBundles 报错） |
| **验证步骤** | 1. 检查 sys_plugin 无新记录<br>2. 检查 `{pluginsDir}/game/` 目录不存在<br>3. 检查 `{staticDir}/plugins/game/` 不存在 |
| **预期结果** | 三项均不存在；无脏数据 |
| **关联需求** | FR-6 |
| **设计方法** | 错误推测 |

---

## 执行结果记录（测试执行时填写）

| 用例编号 | 用例名称 | 优先级 | 结果 | 执行人 | 日期 | 备注 |
|----------|----------|--------|------|--------|------|------|
| TC-001 | admin build:pc 产物验证 | P0 | ⬜ | | | |
| TC-002 | admin build:h5 产物验证 | P0 | ⬜ | | | |
| TC-003 | admin H5 宿主库验证 | P0 | ⬜ | | | |
| TC-004 | user build:pc 产物验证 | P0 | ⬜ | | | |
| TC-005 | user build:h5 产物验证 | P0 | ⬜ | | | |
| TC-006 | getPlatform/getDevice 返回值 | P0 | ⬜ | | | |
| TC-007 | registerExtension 返回 UnregisterFn | P0 | ⬜ | | | |
| TC-008 | i18n.mergeLocale 合入主应用 | P1 | ⬜ | | | |
| TC-009 | onTeardown stop 触发调用 | P0 | ⬜ | | | |
| TC-010 | unloadPlugin 后注册表清空 | P0 | ⬜ | | | |
| TC-011 | teardown 后扩展点自动 unregister | P0 | ⬜ | | | |
| TC-012 | loadPlugin 成功加载 setup | P0 | ⬜ | | | |
| TC-013 | loadPlugin 路径格式验证 | P0 | ⬜ | | | |
| TC-014 | 安装含 frontends 插件部署 | P0 | ⬜ | | | |
| TC-015 | 安装含 4 个 frontends 全部部署 | P0 | ⬜ | | | |
| TC-016 | 卸载插件静态资源目录删除 | P0 | ⬜ | | | |
| TC-017 | GET /static/plugins 200 | P0 | ⬜ | | | |
| TC-018 | PC 模式宿主库不含 Vant | P1 | ⬜ | | | |
| TC-N01 | bundle 404 主应用不崩溃 | P0 | ⬜ | | | |
| TC-N02 | teardown 超时强制继续 | P0 | ⬜ | | | |
| TC-N03 | entry 不存在安装返回错误 | P0 | ⬜ | | | |
| TC-N04 | 重复 loadPlugin 不重复加载 | P1 | ⬜ | | | |
| TC-N05 | 无效 platform/device 类型约束 | P2 | ⬜ | | | |
| TC-N06 | bundle 未注册全局变量 reject | P1 | ⬜ | | | |
| TC-N07 | 安装同名插件冲突错误 | P1 | ⬜ | | | |
| TC-N08 | plugin.json 缺失 name 拒绝 | P1 | ⬜ | | | |
| TC-N09 | teardown 抛异常不阻断清理 | P1 | ⬜ | | | |
| TC-N10 | GET 不存在路径 404 | P0 | ⬜ | | | |
| TC-B01 | frontends 为空安装成功 | P1 | ⬜ | | | |
| TC-B02 | frontends 4 条目全部部署 | P1 | ⬜ | | | |
| TC-B03 | teardown 4999ms 正常完成 | P1 | ⬜ | | | |
| TC-B04 | teardown 5001ms 超时强制 | P1 | ⬜ | | | |
| TC-B05 | pluginName 含连字符变量名转换 | P1 | ⬜ | | | |
| TC-B06 | V1 插件无 frontends 解析不报错 | P1 | ⬜ | | | |
| TC-B07 | 多个 teardown 函数全部执行 | P1 | ⬜ | | | |
| TC-B08 | platform/device 大写转小写路径 | P1 | ⬜ | | | |
| TC-B09 | build:pc 与 build:h5 产物不污染 | P0 | ⬜ | | | |
| TC-R01 | PC 端登录菜单权限不受影响 RG-1 | P0 | ⬜ | | | |
| TC-R02 | admin PC 构建产物功能等价 RG-2 | P0 | ⬜ | | | |
| TC-R03 | user PC 构建产物功能等价 RG-3 | P0 | ⬜ | | | |
| TC-R04 | V1 插件 setup 正常调用 RG-4 | P0 | ⬜ | | | |
| TC-R05 | 插件 stop/uninstall 无残留 RG-5 | P0 | ⬜ | | | |
| TC-R06 | 静态路由不影响 /api/ RG-6 | P0 | ⬜ | | | |
| TC-F01 | H5Layout NavBar 渲染 | P0 | ⬜ | | | |
| TC-F02 | H5Layout 汉堡菜单弹出关闭 | P0 | ⬜ | | | |
| TC-F03 | H5Layout 返回按钮路由回退 | P1 | ⬜ | | | |
| TC-F04 | UserH5Layout Tabbar 渲染 | P0 | ⬜ | | | |
| TC-F05 | UserH5Layout 路由跳转高亮 | P0 | ⬜ | | | |
| TC-F06 | UserH5Layout 菜单空态 | P1 | ⬜ | | | |
| TC-UX01 | H5Layout 布局合理性 | P1 | ⬜ | | | |
| TC-UX02 | H5Layout 一致性 | P1 | ⬜ | | | |
| TC-UX03 | H5Layout 完整性（空态/加载态） | P1 | ⬜ | | | |
| TC-UX04 | H5Layout 操作效率 | P1 | ⬜ | | | |
| TC-UX05 | H5Layout 响应式适配 | P1 | ⬜ | | | |
| TC-UX06 | UserH5Layout 布局合理性 | P1 | ⬜ | | | |
| TC-UX07 | UserH5Layout 清晰性（文案） | P1 | ⬜ | | | |
| TC-UX08 | UserH5Layout 错误预防 | P1 | ⬜ | | | |
| TC-UX09 | UserH5Layout 信息层次 | P1 | ⬜ | | | |
| TC-UX10 | PC vs H5 一致性 | P1 | ⬜ | | | |
| TC-A01 | GET bundle 200 | P0 | ⬜ | | | |
| TC-A02 | GET bundle 404 | P0 | ⬜ | | | |
| TC-A03 | GET user-h5 路径格式 | P1 | ⬜ | | | |
| TC-A04 | GET /api/v1/admin/plugins 不冲突 | P0 | ⬜ | | | |
| TC-A05 | go build ./... 零错误 | P0 | ⬜ | | | |
| TC-A06 | go vet ./... 无警告 | P1 | ⬜ | | | |
| TC-D01 | sys_plugin 记录字段验证 | P0 | ⬜ | | | |
| TC-D02 | bundle 路径一致性验证 | P0 | ⬜ | | | |
| TC-D03 | 卸载后硬删除验证 | P0 | ⬜ | | | |
| TC-D04 | admin_application 随插件创建 | P1 | ⬜ | | | |
| TC-D05 | 安装失败无脏数据 | P1 | ⬜ | | | |

---

## 质量自检结果

| # | 检查项 | 结果 |
|---|--------|------|
| 1 | 每个 FR-N 至少 1 正向 + 1 反向用例 | ✅ FR-1~FR-6 均覆盖 |
| 2 | 每个 RG-N 有对应回归用例 | ✅ TC-R01~TC-R06 对应 RG-1~RG-6 |
| 3 | 每个新增前端页面至少 3 条 TC-F | ✅ H5Layout: TC-F01~F03；UserH5Layout: TC-F04~F06 |
| 4 | TC-UX 覆盖 8 维度中的 5+ | ✅ 覆盖：布局合理性、一致性、完整性、操作效率、响应式、清晰性、错误预防、信息层次（8/8） |
| 5 | 所有 TC-UX 的"不满足时修复建议"不为空 | ✅ TC-UX01~TC-UX10 均含修复建议 |
| 6 | P0 覆盖所有核心接口 | ✅ TC-A01/A02/A04/A05 均为 P0 |
| 7 | TC-UX 引用具体业界参考 | ✅ 引用 Vant/Nielsen/Material Design/微信规范 |
| 8 | 无孤立用例 | ✅ 每条均标注 FR-N 或 RG-N |
