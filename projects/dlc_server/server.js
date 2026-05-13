/**
 * DLC 下载服务器
 *
 * 提供以下接口：
 *   GET  /api/dlc/list              — 获取所有可用 DLC 列表
 *   GET  /api/dlc/:id/info          — 获取指定 DLC 的信息（含 SHA-256）
 *   POST /api/dlc/:id/verify        — 验证授权 Token
 *   GET  /api/dlc/:id/download      — 下载 DLC PCK 文件（需要 Token）
 *   GET  /health                    — 健康检查
 *
 * 启动：node server.js
 * 默认端口：8787
 */

const express = require('express');
const cors = require('cors');
const fs = require('fs');
const path = require('path');
const crypto = require('crypto');

const app = express();
const config = JSON.parse(fs.readFileSync('./config.json', 'utf8'));

app.use(cors());
app.use(express.json());

// ─────────────────────────────────────────────
// 工具函数
// ─────────────────────────────────────────────

/** 从请求头提取 Bearer Token */
function extractToken(req) {
  const auth = req.headers['authorization'] || '';
  if (auth.startsWith('Bearer ')) {
    return auth.slice(7).trim();
  }
  return null;
}

/** 验证 Token 是否有效 */
function isValidToken(token) {
  if (!token) return false;
  return config.valid_tokens.includes(token);
}

/** 计算文件 SHA-256 */
function computeFileSha256(filePath) {
  const hash = crypto.createHash('sha256');
  const data = fs.readFileSync(filePath);
  hash.update(data);
  return hash.digest('hex');
}

/** 获取 DLC 文件的完整路径 */
function getDlcFilePath(dlcId) {
  const dlcInfo = config.dlcs[dlcId];
  if (!dlcInfo) return null;
  return path.join(config.dlc_files_dir, dlcInfo.file);
}

// ─────────────────────────────────────────────
// 路由
// ─────────────────────────────────────────────

/** 健康检查 */
app.get('/health', (req, res) => {
  res.json({ status: 'ok', time: new Date().toISOString() });
});

/** 获取所有可用 DLC 列表 */
app.get('/api/dlc/list', (req, res) => {
  const list = Object.values(config.dlcs).map(dlc => ({
    id: dlc.id,
    display_name: dlc.display_name,
    version: dlc.version,
    price: dlc.price,
    free: dlc.free,
  }));
  res.json({ dlcs: list });
});

/** 获取指定 DLC 的详细信息（含 SHA-256，用于客户端验证） */
app.get('/api/dlc/:id/info', (req, res) => {
  const dlcId = req.params.id;
  const dlcInfo = config.dlcs[dlcId];

  if (!dlcInfo) {
    return res.status(404).json({ error: `DLC 不存在: ${dlcId}` });
  }

  const filePath = getDlcFilePath(dlcId);
  let actualHash = dlcInfo.sha256;

  // 如果文件存在，实时计算哈希（开发模式）
  if (fs.existsSync(filePath)) {
    actualHash = computeFileSha256(filePath);
    const fileSize = fs.statSync(filePath).size;
    return res.json({
      id: dlcInfo.id,
      display_name: dlcInfo.display_name,
      version: dlcInfo.version,
      sha256: actualHash,
      file_size: fileSize,
      free: dlcInfo.free,
    });
  }

  // 文件不存在时返回配置中的哈希
  res.json({
    id: dlcInfo.id,
    display_name: dlcInfo.display_name,
    version: dlcInfo.version,
    sha256: actualHash,
    file_size: 0,
    free: dlcInfo.free,
    warning: 'PCK 文件尚未放置，请先导出 DLC PCK',
  });
});

/** 验证授权 Token（付费 DLC 使用） */
app.post('/api/dlc/:id/verify', (req, res) => {
  const dlcId = req.params.id;
  const dlcInfo = config.dlcs[dlcId];

  if (!dlcInfo) {
    return res.status(404).json({ authorized: false, error: 'DLC 不存在' });
  }

  // 免费 DLC 直接授权
  if (dlcInfo.free) {
    return res.json({ authorized: true, dlc_id: dlcId, free: true });
  }

  // 付费 DLC 验证 Token
  const token = extractToken(req);
  const { device_id } = req.body;

  if (!isValidToken(token)) {
    return res.status(401).json({
      authorized: false,
      error: '无效的授权 Token，请确认已购买此 DLC',
    });
  }

  console.log(`[Auth] DLC ${dlcId} 授权通过 | Token: ${token} | Device: ${device_id}`);
  res.json({
    authorized: true,
    dlc_id: dlcId,
    token_valid: true,
    device_id: device_id,
  });
});

/** 下载 DLC PCK 文件 */
app.get('/api/dlc/:id/download', (req, res) => {
  const dlcId = req.params.id;
  const dlcInfo = config.dlcs[dlcId];

  if (!dlcInfo) {
    return res.status(404).json({ error: 'DLC 不存在' });
  }

  // 付费 DLC 需要验证 Token
  if (!dlcInfo.free) {
    const token = extractToken(req);
    if (!isValidToken(token)) {
      return res.status(401).json({ error: '未授权，请先购买此 DLC' });
    }
  }

  const filePath = getDlcFilePath(dlcId);
  if (!filePath || !fs.existsSync(filePath)) {
    return res.status(404).json({
      error: `PCK 文件不存在: ${dlcInfo.file}`,
      hint: '请先在 dlc_pack 工程中导出 PCK，然后复制到 dlc_server/dlc_files/ 目录',
    });
  }

  const fileSize = fs.statSync(filePath).size;
  const sha256 = computeFileSha256(filePath);

  console.log(`[Download] DLC ${dlcId} | Size: ${fileSize} bytes | SHA256: ${sha256}`);

  // 设置响应头
  res.setHeader('Content-Type', 'application/octet-stream');
  res.setHeader('Content-Disposition', `attachment; filename="${dlcInfo.file}"`);
  res.setHeader('Content-Length', fileSize);
  res.setHeader('X-DLC-SHA256', sha256);  // 客户端可用此头验证完整性
  res.setHeader('X-DLC-Version', dlcInfo.version);

  // 流式发送文件
  const stream = fs.createReadStream(filePath);
  stream.pipe(res);
  stream.on('error', (err) => {
    console.error('[Download Error]', err);
    if (!res.headersSent) {
      res.status(500).json({ error: '文件读取失败' });
    }
  });
});

// ─────────────────────────────────────────────
// 启动服务器
// ─────────────────────────────────────────────

const PORT = config.port || 8787;
app.listen(PORT, () => {
  console.log(`\n🚀 DLC 服务器已启动`);
  console.log(`   地址: http://localhost:${PORT}`);
  console.log(`\n📋 可用接口:`);
  console.log(`   GET  http://localhost:${PORT}/health`);
  console.log(`   GET  http://localhost:${PORT}/api/dlc/list`);
  console.log(`   GET  http://localhost:${PORT}/api/dlc/poison_dlc/info`);
  console.log(`   POST http://localhost:${PORT}/api/dlc/poison_dlc/verify`);
  console.log(`   GET  http://localhost:${PORT}/api/dlc/poison_dlc/download`);
  console.log(`\n📁 DLC 文件目录: ${path.resolve(config.dlc_files_dir)}`);
  console.log(`\n⚠️  请先将 poison_dlc.pck 放入 dlc_files/ 目录\n`);
});
