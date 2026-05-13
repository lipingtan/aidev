# DLC 下载服务器

## 快速启动

```bash
cd projects/dlc_server
npm install
npm start
```

服务器启动后监听 http://localhost:8787

## 准备 DLC 文件

1. 在 `projects/dlc_pack` 工程中导出 `poison_dlc.pck`
2. 将 `poison_dlc.pck` 复制到 `dlc_server/dlc_files/` 目录
3. 服务器会自动计算 SHA-256，无需手动填写 config.json

## API 接口

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /health | 健康检查 |
| GET | /api/dlc/list | 获取所有 DLC 列表 |
| GET | /api/dlc/:id/info | 获取 DLC 信息（含 SHA-256） |
| POST | /api/dlc/:id/verify | 验证授权 Token |
| GET | /api/dlc/:id/download | 下载 DLC PCK |

## 测试 Token

开发测试用 Token（在 config.json 中配置）：
- `test-token-12345`
- `dev-token-99999`

## 用 curl 测试

```bash
# 健康检查
curl http://localhost:8787/health

# 获取 DLC 信息（含 SHA-256）
curl http://localhost:8787/api/dlc/poison_dlc/info

# 验证 Token（免费 DLC 无需 Token）
curl -X POST http://localhost:8787/api/dlc/poison_dlc/verify \
  -H "Content-Type: application/json" \
  -d '{"device_id": "test-device"}'

# 下载 DLC（免费 DLC 无需 Token）
curl -o poison_dlc.pck http://localhost:8787/api/dlc/poison_dlc/download
```
