# Platform Admin E2E Tests

基于 Playwright 的端到端测试，验证完整用户流程。

## 前置条件

1. 后端运行在 `http://localhost:8000`
2. 前端运行在 `http://localhost:5173`（仅浏览器 UI 测试需要）

## 安装

```bash
cd projects/demo/platform_admin/e2e
npm install
npx playwright install chromium
```

## 运行

```bash
# 全部测试
npm test

# 带浏览器界面
npm run test:headed

# 调试模式
npm run test:debug

# 查看报告
npm run report
```

## 测试结构

```
tests/
├── auth.spec.ts           # 认证流程（登录/错误密码）
├── route-migration.spec.ts # 路由迁移验证（新路径/旧路径404）
├── crud.spec.ts           # CRUD 回归（角色/用户）
└── seed-data.spec.ts      # Seed 数据验证（应用/菜单/租户字段）
```
