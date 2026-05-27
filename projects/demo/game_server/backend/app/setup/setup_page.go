package setup

// setupPageHTML 安装向导页面 HTML
// 内嵌在二进制中，无需外部文件
const setupPageHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>管理平台开发底座 - 初始化安装</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; background: #0f172a; color: #e2e8f0; min-height: 100vh; display: flex; align-items: center; justify-content: center; }
  .container { width: 100%; max-width: 560px; padding: 24px; }
  .card { background: #1e293b; border: 1px solid #334155; border-radius: 12px; padding: 40px; }
  .logo { text-align: center; margin-bottom: 32px; }
  .logo h1 { font-size: 24px; font-weight: 700; color: #f1f5f9; }
  .logo p { color: #94a3b8; margin-top: 8px; font-size: 14px; }
  .steps { display: flex; justify-content: center; gap: 8px; margin-bottom: 32px; }
  .step { display: flex; align-items: center; gap: 6px; font-size: 13px; color: #64748b; }
  .step.active { color: #3b82f6; }
  .step.done { color: #22c55e; }
  .step-num { width: 24px; height: 24px; border-radius: 50%; border: 2px solid currentColor; display: flex; align-items: center; justify-content: center; font-size: 11px; font-weight: 600; }
  .step-line { width: 32px; height: 2px; background: #334155; }
  .section { margin-bottom: 24px; }
  .section-title { font-size: 13px; font-weight: 600; color: #94a3b8; text-transform: uppercase; letter-spacing: 0.05em; margin-bottom: 16px; padding-bottom: 8px; border-bottom: 1px solid #334155; }
  .form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; }
  .form-group { margin-bottom: 16px; }
  .form-group label { display: block; font-size: 13px; color: #94a3b8; margin-bottom: 6px; }
  .form-group input { width: 100%; padding: 10px 12px; background: #0f172a; border: 1px solid #334155; border-radius: 8px; color: #f1f5f9; font-size: 14px; outline: none; transition: border-color 0.2s; }
  .form-group input:focus { border-color: #3b82f6; }
  .form-group input.error { border-color: #ef4444; }
  .form-group .hint { font-size: 12px; color: #64748b; margin-top: 4px; }
  .test-btn { padding: 8px 16px; background: #1e40af; color: #bfdbfe; border: 1px solid #3b82f6; border-radius: 6px; font-size: 13px; cursor: pointer; transition: all 0.2s; }
  .test-btn:hover { background: #1d4ed8; }
  .test-result { margin-top: 8px; font-size: 13px; padding: 8px 12px; border-radius: 6px; display: none; }
  .test-result.success { background: #052e16; color: #4ade80; border: 1px solid #166534; display: block; }
  .test-result.error { background: #450a0a; color: #f87171; border: 1px solid #991b1b; display: block; }
  .submit-btn { width: 100%; padding: 14px; background: #3b82f6; color: white; border: none; border-radius: 8px; font-size: 15px; font-weight: 600; cursor: pointer; transition: background 0.2s; margin-top: 8px; }
  .submit-btn:hover { background: #2563eb; }
  .submit-btn:disabled { background: #1e3a5f; color: #64748b; cursor: not-allowed; }
  .progress { display: none; margin-top: 24px; }
  .progress-bar { height: 4px; background: #334155; border-radius: 2px; overflow: hidden; }
  .progress-fill { height: 100%; background: #3b82f6; border-radius: 2px; transition: width 0.5s; }
  .progress-text { font-size: 13px; color: #94a3b8; margin-top: 8px; text-align: center; }
  .success-panel { display: none; text-align: center; padding: 24px 0; }
  .success-icon { font-size: 48px; margin-bottom: 16px; }
  .success-panel h2 { font-size: 20px; color: #4ade80; margin-bottom: 8px; }
  .success-panel p { color: #94a3b8; font-size: 14px; }
  .goto-btn { display: inline-block; margin-top: 20px; padding: 12px 32px; background: #22c55e; color: white; border-radius: 8px; font-size: 14px; font-weight: 600; text-decoration: none; cursor: pointer; border: none; }
</style>
</head>
<body>
<div class="container">
  <div class="card">
    <div class="logo">
      <h1>⚡ 管理平台开发底座</h1>
      <p>首次运行，请完成初始化配置</p>
    </div>

    <div class="steps">
      <div class="step active" id="step1"><div class="step-num">1</div><span>数据库</span></div>
      <div class="step-line"></div>
      <div class="step" id="step3"><div class="step-num">2</div><span>完成</span></div>
    </div>

    <div id="installForm">
      <!-- 数据库配置 -->
      <div class="section">
        <div class="section-title">数据库配置</div>
        <div class="form-row">
          <div class="form-group">
            <label>数据库地址</label>
            <input type="text" id="dbHost" value="127.0.0.1" placeholder="127.0.0.1">
          </div>
          <div class="form-group">
            <label>端口</label>
            <input type="number" id="dbPort" value="3306" placeholder="3306">
          </div>
        </div>
        <div class="form-group">
          <label>数据库名</label>
          <input type="text" id="dbName" placeholder="platform_base" value="platform_base">
          <div class="hint">数据库需要提前创建，或确保用户有创建数据库的权限</div>
        </div>
        <div class="form-row">
          <div class="form-group">
            <label>用户名</label>
            <input type="text" id="dbUser" placeholder="root">
          </div>
          <div class="form-group">
            <label>密码</label>
            <input type="password" id="dbPassword" placeholder="（留空表示无密码）">
          </div>
        </div>
        <button class="test-btn" onclick="testDB()">🔌 测试连接</button>
        <div class="test-result" id="testResult"></div>
      </div>

      <!-- 管理员配置 -->
      <div class="section">
        <div class="section-title">管理员账号（默认）</div>
        <div style="background:#0f172a;border:1px solid #334155;border-radius:8px;padding:12px 16px;font-size:13px;color:#94a3b8;">
          <div>用户名：<strong style="color:#f1f5f9">admin</strong></div>
          <div style="margin-top:6px">密码：<strong style="color:#f1f5f9">admin123</strong></div>
          <div style="margin-top:8px;color:#64748b;font-size:12px">⚠️ 安装完成后请及时修改默认密码</div>
        </div>
      </div>

      <!-- 服务配置 -->
      <div class="section">
        <div class="section-title">服务配置</div>
        <div class="form-row">
          <div class="form-group">
            <label>平台名称</label>
            <input type="text" id="appName" value="管理平台开发底座" placeholder="管理平台开发底座">
          </div>
          <div class="form-group">
            <label>服务端口</label>
            <input type="number" id="appPort" value="8000" placeholder="8000">
          </div>
        </div>
      </div>

      <button class="submit-btn" id="submitBtn" onclick="doInstall()">🚀 开始安装</button>
    </div>

    <!-- 进度条 -->
    <div class="progress" id="progress">
      <div class="progress-bar"><div class="progress-fill" id="progressFill" style="width:0%"></div></div>
      <div class="progress-text" id="progressText">正在初始化...</div>
    </div>

    <!-- 成功面板 -->
    <div class="success-panel" id="successPanel">
      <div class="success-icon">✅</div>
      <h2>安装成功！</h2>
      <p>系统已完成初始化，即将跳转到管理后台</p>
      <button class="goto-btn" onclick="window.location.href='/'">进入管理后台 →</button>
    </div>
  </div>
</div>

<script>
async function testDB() {
  const btn = document.querySelector('.test-btn');
  const result = document.getElementById('testResult');
  btn.disabled = true;
  btn.textContent = '连接中...';
  result.className = 'test-result';
  result.style.display = 'none';

  try {
    const resp = await fetch('/setup/test-db', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({
        dbHost: document.getElementById('dbHost').value,
        dbPort: parseInt(document.getElementById('dbPort').value),
        dbName: document.getElementById('dbName').value,
        dbUser: document.getElementById('dbUser').value,
        dbPassword: document.getElementById('dbPassword').value,
      })
    });
    const data = await resp.json();
    result.textContent = data.code === 200 ? '✓ ' + data.msg : '✗ ' + data.msg;
    result.className = 'test-result ' + (data.code === 200 ? 'success' : 'error');
    result.style.display = 'block';
  } catch(e) {
    result.textContent = '✗ 请求失败: ' + e.message;
    result.className = 'test-result error';
    result.style.display = 'block';
  }
  btn.disabled = false;
  btn.textContent = '🔌 测试连接';
}

async function doInstall() {
  const btn = document.getElementById('submitBtn');
  btn.disabled = true;

  document.getElementById('installForm').style.display = 'none';
  const progress = document.getElementById('progress');
  progress.style.display = 'block';

  const steps = [
    [10, '正在验证数据库连接...'],
    [30, '正在创建数据库表结构...'],
    [60, '正在初始化系统数据...'],
    [80, '正在写入配置文件...'],
    [90, '正在启动服务...'],
  ];

  let stepIdx = 0;
  const timer = setInterval(() => {
    if (stepIdx < steps.length) {
      setProgress(steps[stepIdx][0], steps[stepIdx][1]);
      stepIdx++;
    }
  }, 600);

  try {
    const resp = await fetch('/setup/init', {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify({
        dbHost: document.getElementById('dbHost').value,
        dbPort: parseInt(document.getElementById('dbPort').value),
        dbName: document.getElementById('dbName').value,
        dbUser: document.getElementById('dbUser').value,
        dbPassword: document.getElementById('dbPassword').value,
        appName: document.getElementById('appName').value,
        appPort: parseInt(document.getElementById('appPort').value) || 8000,
      })
    });
    const data = await resp.json();
    clearInterval(timer);

    if (data.code === 200) {
      setProgress(100, '安装完成！');
      document.getElementById('step1').className = 'step done';
      document.getElementById('step3').className = 'step active';
      setTimeout(() => {
        progress.style.display = 'none';
        document.getElementById('successPanel').style.display = 'block';
        setTimeout(() => { window.location.href = '/'; }, 3000);
      }, 800);
    } else {
      clearInterval(timer);
      progress.style.display = 'none';
      document.getElementById('installForm').style.display = 'block';
      btn.disabled = false;
      alert('安装失败：' + data.msg);
    }
  } catch(e) {
    clearInterval(timer);
    progress.style.display = 'none';
    document.getElementById('installForm').style.display = 'block';
    btn.disabled = false;
    alert('请求失败：' + e.message);
  }
}

function setProgress(pct, text) {
  document.getElementById('progressFill').style.width = pct + '%';
  document.getElementById('progressText').textContent = text;
}
</script>
</body>
</html>`
