/* ============================================================
 * 魔塔 — 游戏主体：状态 / 移动 / 开门 / 道具 / 楼梯 / UI
 * 依赖加载顺序: sprites.js monsters.js maps.js sfx.js battle.js
 * ============================================================ */
(function (root, factory) {
  const m = factory();
  if (typeof module !== 'undefined' && module.exports) module.exports = m;
  root.MT = Object.assign(root.MT || {}, { game: m });
})(typeof self !== 'undefined' ? self : globalThis, function () {

  const { SPRITES, HERO_FRAMES, makeSpriteCanvas } = MT;
  const { MONSTERS, ITEMS, DOORS, KEYS, PLAYER_START } = MT;
  const MAPS = MT.MAPS;

  const N = 15, TILE = 64; // 32px 精灵 × 2 = 64px 瓦片，画布 960×960
  const FLOOR_NAMES = ['B1', 'B2', 'B3'];
  const FLOOR_TITLES = ['B1 · 新手层', 'B2 · 骷髅层', 'B3 · 龙王层'];

  /* ---------- 状态 ---------- */
  let S = null; // { floors, floor, pos, player, log, over }
  let px = 0, py = 0;      // 玩家渲染位置（格子坐标，插值）
  let lastMove = 0;
  let dialogOpen = false;

  /* ---------- 行走动画 ---------- */
  const WALK_MS = 150;                 // 每步时长
  let walk = null;                     // { fx,fy,tx,ty,t,dir }
  let facing = 'd';                    // 当前朝向（空闲时保持）
  // 英雄各朝向双帧（32px×2=64px 瓦片），惰性预渲染（Node 环境无 document）
  let heroFrames = null;
  function getHeroFrames() {
    if (!heroFrames) {
      heroFrames = {};
      for (const d of ['d', 'u', 'l', 'r']) {
        heroFrames[d] = HERO_FRAMES[d].map(f => makeSpriteCanvas(f, 2));
      }
    }
    return heroFrames;
  }

  function freshState() {
    const floors = MAPS.map(rows => rows.map(r => [...r]));
    let sx = 1, sy = 1;
    for (let y = 0; y < N; y++) for (let x = 0; x < N; x++) if (floors[0][y][x] === '@') { sx = x; sy = y; floors[0][y][x] = '.'; }
    return {
      floors, floor: 0,
      pos: [{ x: sx, y: sy }, { x: -1, y: -1 }, { x: -1, y: -1 }],
      player: { ...PLAYER_START, keys: { ...PLAYER_START.keys } },
      log: [], over: false,
    };
  }

  function addLog(text, cls) {
    S.log.push({ text, cls: cls || '' });
    if (S.log.length > 40) S.log.shift();
    renderLog();
  }

  /* ---------- DOM ---------- */
  let canvas, ctx2d;
  let ui = {}; // 面板元素
  function bindUI() {
    canvas = document.getElementById('map');
    ctx2d = canvas.getContext('2d');
    ui = {
      floorBadge: document.getElementById('floor-badge'),
      gold: document.getElementById('hud-gold'),
      hpBar: document.getElementById('hp-bar'),
      hpTxt: document.getElementById('hp-txt'),
      atk: document.getElementById('stat-atk'),
      def: document.getElementById('stat-def'),
      keyR: document.getElementById('key-r'),
      keyB: document.getElementById('key-b'),
      keyG: document.getElementById('key-g'),
      log: document.getElementById('log'),
      muteBtn: document.getElementById('btn-mute'),
      helpBtn: document.getElementById('btn-help'),
      restartBtn: document.getElementById('btn-restart'),
    };
    const pc = document.getElementById('portrait');
    pc.getContext('2d').drawImage(makeSpriteCanvas(SPRITES.hero, 3), 0, 0); // 32px×3=96
    ui.muteBtn.onclick = () => { const m = MT.sfx.toggle(); ui.muteBtn.textContent = m ? '🔇 静音' : '🔊 音效'; };
    ui.helpBtn.onclick = showHelp;
    ui.restartBtn.onclick = restart;
    document.querySelectorAll('.dpad button').forEach(b => {
      b.addEventListener('pointerdown', (e) => { e.preventDefault(); const [dx, dy] = b.dataset.dir.split(','); move(dx | 0, dy | 0); });
    });
  }

  /* ---------- 渲染 ---------- */
  const spriteCache = new Map(); // sprite obj -> 3x canvas
  function tileSprite(sp) {
    let c = spriteCache.get(sp);
    if (!c) { c = makeSpriteCanvas(sp, 3); spriteCache.set(sp, c); }
    return c;
  }
  function drawWall(x, y) {
    const X = x * TILE, Y = y * TILE;
    ctx2d.fillStyle = '#1a1c2e';
    ctx2d.fillRect(X, Y, TILE, TILE);
    ctx2d.strokeStyle = '#262a44';
    ctx2d.lineWidth = 2;
    ctx2d.strokeRect(X + 1, Y + 1, TILE - 2, TILE - 2);
    ctx2d.strokeStyle = '#151726';
    ctx2d.beginPath();
    ctx2d.moveTo(X, Y + TILE / 2); ctx2d.lineTo(X + TILE, Y + TILE / 2);
    ctx2d.moveTo(X + TILE / 2, Y); ctx2d.lineTo(X + TILE / 2, Y + TILE / 2);
    ctx2d.moveTo(X + TILE / 4, Y + TILE / 2); ctx2d.lineTo(X + TILE / 4, Y + TILE);
    ctx2d.moveTo(X + 3 * TILE / 4, Y + TILE / 2); ctx2d.lineTo(X + 3 * TILE / 4, Y + TILE);
    ctx2d.stroke();
  }

  function drawFloor(x, y) {
    const X = x * TILE, Y = y * TILE;
    ctx2d.fillStyle = (x + y) % 2 === 0 ? '#2b2f4a' : '#282c46';
    ctx2d.fillRect(X, Y, TILE, TILE);
    ctx2d.strokeStyle = 'rgba(255,255,255,.04)';
    ctx2d.strokeRect(X + .5, Y + .5, TILE - 1, TILE - 1);
  }

  function draw() {
    const g = S.floors[S.floor];
    ctx2d.clearRect(0, 0, canvas.width, canvas.height);
    for (let y = 0; y < N; y++) for (let x = 0; x < N; x++) {
      const c = g[y][x];
      if (c === '#') drawWall(x, y);
      else drawFloor(x, y);
    }
    // 静态实体
    for (let y = 0; y < N; y++) for (let x = 0; x < N; x++) {
      const c = g[y][x];
      if (c === '#' || c === '.') continue;
      let sp = null;
      if (MONSTERS[c]) sp = SPRITES[MONSTERS[c].id];
      else if (ITEMS[c]) sp = SPRITES[ITEMS[c].id];
      else if (DOORS[c]) sp = SPRITES['door_' + { r: 'red', b: 'blue', g: 'gold' }[c]];
      else if (KEYS[c]) sp = SPRITES['key_' + { R: 'red', B: 'blue', G: 'gold' }[c]];
      else if (c === 'U') sp = SPRITES.stairsUp;
      else if (c === 'L') sp = SPRITES.stairsDown;
      if (sp) ctx2d.drawImage(tileSprite(sp), x * TILE, y * TILE);
    }
    // 玩家（插值位置 + 朝向帧 + 走路起伏）
    const hf = getHeroFrames();
    let dir = facing, frame = 0, bob = 0;
    if (walk) {
      dir = walk.dir;
      frame = walk.t < 0.5 ? 0 : 1;
      bob = Math.round(Math.sin(walk.t * Math.PI) * -2); // 步中上浮 2px
    }
    // 脚下柔和投影（角色落地感；bob 上浮时影子留地）
    ctx2d.save();
    ctx2d.globalAlpha = 0.28;
    ctx2d.fillStyle = '#000';
    ctx2d.beginPath();
    ctx2d.ellipse((px + 0.5) * TILE, (py + 1) * TILE - 4, TILE * 0.32, TILE * 0.14, 0, 0, Math.PI * 2);
    ctx2d.fill();
    ctx2d.restore();
    ctx2d.drawImage(hf[dir][frame], Math.round(px * TILE), Math.round(py * TILE) + bob);
  }

  let lastT = 0;
  function tick(now) {
    const t = S.pos[S.floor];
    const dt = lastT ? Math.min(50, now - lastT) : 16;
    lastT = now;
    if (walk) {
      walk.t += dt / WALK_MS;
      if (walk.t >= 1) {
        px = walk.tx; py = walk.ty;
        facing = walk.dir;
        walk = null;
      } else {
        const e = walk.t * (2 - walk.t); // easeOutQuad，起步快收尾缓
        px = walk.fx + (walk.tx - walk.fx) * e;
        py = walk.fy + (walk.ty - walk.fy) * e;
      }
    } else {
      // 非行走（楼梯滑入等）：平滑逼近目标
      px += (t.x - px) * 0.35;
      py += (t.y - py) * 0.35;
      if (Math.abs(t.x - px) < 0.02) px = t.x;
      if (Math.abs(t.y - py) < 0.02) py = t.y;
    }
    draw();
    requestAnimationFrame(tick);
  }

  function updatePanel() {
    const p = S.player;
    ui.floorBadge.textContent = FLOOR_NAMES[S.floor];
    ui.gold.textContent = '💰 ' + p.gold;
    ui.hpBar.style.width = Math.min(100, (p.hp / 25000) * 100) + '%';
    ui.hpTxt.textContent = p.hp;
    ui.atk.textContent = p.atk;
    ui.def.textContent = p.def;
    ui.keyR.textContent = p.keys.red;
    ui.keyB.textContent = p.keys.blue;
    ui.keyG.textContent = p.keys.gold;
  }

  function renderLog() {
    ui.log.innerHTML = S.log.slice(-8).map(l => `<div class="log-line ${l.cls}">${l.text}</div>`).join('');
  }

  function floatText(x, y, text, cls) {
    const d = document.createElement('div');
    d.className = 'float-txt ' + (cls || '');
    d.textContent = text;
    const k = canvas.clientWidth / 960; // 小屏缩放对齐
    d.style.left = (x * TILE + TILE / 2) * k + 'px';
    d.style.top = (y * TILE + 8) * k + 'px';
    document.getElementById('floaters').appendChild(d);
    setTimeout(() => d.remove(), 1000);
  }

  /* ---------- 动作 ---------- */
  // 从当前渲染位置发起一步行走（支持连续行走接链）
  function startWalk(nx, ny, dx, dy) {
    const dir = dx < 0 ? 'l' : dx > 0 ? 'r' : dy < 0 ? 'u' : 'd';
    walk = { fx: px, fy: py, tx: nx, ty: ny, t: 0, dir };
    facing = dir;
  }

  function move(dx, dy) {
    if (S.over || dialogOpen) return;
    const now = Date.now();
    if (now - lastMove < 130) return;
    lastMove = now;
    const t = S.pos[S.floor];
    const nx = t.x + dx, ny = t.y + dy;
    if (nx < 0 || ny < 0 || nx >= N || ny >= N) return;
    const g = S.floors[S.floor];
    const c = g[ny][nx];

    if (c === '#') { MT.sfx.play('step'); return; }

    if (MONSTERS[c]) { openBattle(c, nx, ny); return; }

    if (DOORS[c]) {
      const need = DOORS[c].need;
      if (S.player.keys[need] > 0) {
        S.player.keys[need]--;
        g[ny][nx] = '.';
        t.x = nx; t.y = ny; // 穿过门
        startWalk(nx, ny, dx, dy);
        MT.sfx.play('door');
        addLog(`打开${DOORS[c].name}`, 'sys');
        floatText(nx, ny, '开门', 'sys');
      } else {
        addLog(`需要${{ red: '红', blue: '蓝', gold: '金' }[need]}钥匙`, 'warn');
        floatText(nx, ny, '上锁', 'warn');
        MT.sfx.play('block');
      }
      updatePanel(); draw();
      return;
    }

    if (ITEMS[c]) {
      const it = ITEMS[c];
      if (it.kind === 'hp') S.player.hp += it.val;
      else if (it.kind === 'atk') S.player.atk += it.val;
      else if (it.kind === 'def') S.player.def += it.val;
      else if (it.kind === 'gold') S.player.gold += it.val;
      g[ny][nx] = '.';
      t.x = nx; t.y = ny; // 站到拾取格上
      startWalk(nx, ny, dx, dy);
      MT.sfx.play(it.kind === 'gold' ? 'gold' : 'pickup');
      addLog(`拾取 ${it.name}（${it.text}）`, 'item');
      floatText(nx, ny, it.text, 'item');
      updatePanel(); draw();
      return;
    }

    if (KEYS[c]) {
      const col = KEYS[c];
      S.player.keys[col]++;
      g[ny][nx] = '.';
      t.x = nx; t.y = ny; // 站到拾取格上
      startWalk(nx, ny, dx, dy);
      MT.sfx.play('pickup');
      addLog(`拾取${{ red: '红', blue: '蓝', gold: '金' }[col]}钥匙（共${S.player.keys[col]}把）`, 'item');
      floatText(nx, ny, `${{ red: '红', blue: '蓝', gold: '金' }[col]}钥匙+1`, 'item');
      updatePanel(); draw();
      return;
    }

    if (c === 'U') {
      if (S.floor < 2) { goUp(); return; }
      addLog('这里没有更上层了', 'warn');
      return;
    }
    if (c === 'L') {
      if (S.floor > 0) { goDown(); return; }
      addLog('已经在一楼了', 'warn');
      return;
    }

    // 空地
    t.x = nx; t.y = ny;
    startWalk(nx, ny, dx, dy);
    MT.sfx.play('step');
  }

  function findTile(floor, ch) {
    const g = S.floors[floor];
    for (let y = 0; y < N; y++) for (let x = 0; x < N; x++) if (g[y][x] === ch) return { x, y };
    return null;
  }

  function goUp() {
    const dest = S.floor + 1;
    const t = findTile(dest, 'L');
    S.floor = dest;
    if (t) S.pos[dest] = t; else S.pos[dest] = { x: 1, y: 1 };
    px = py = -99; // 从远处滑入
    MT.sfx.play('stairs');
    addLog(`来到 ${FLOOR_TITLES[S.floor]}`, 'sys');
    updatePanel(); draw();
  }

  function goDown() {
    const dest = S.floor - 1;
    const t = findTile(dest, 'U');
    S.floor = dest;
    if (t) S.pos[dest] = t; else S.pos[dest] = { x: 1, y: 1 };
    px = py = -99;
    MT.sfx.play('stairs');
    addLog(`回到 ${FLOOR_TITLES[S.floor]}`, 'sys');
    updatePanel(); draw();
  }

  /* ---------- 战斗 ---------- */
  function openBattle(mkey, x, y) {
    const m = MONSTERS[mkey];
    dialogOpen = true;
    MT.battle.start(mkey, S.player, (result) => {
      dialogOpen = false;
      if (result === 'win') {
        S.floors[S.floor][y][x] = '.';
        S.player.gold += m.gold;
        addLog(`击败${m.name}！获得 ${m.gold} 金币`, 'win');
        floatText(x, y, `+${m.gold}💰`, 'win');
        updatePanel(); draw();
        if (m.boss) showEnd(true);
      } else if (result === 'lose') {
        addLog('你被击败了…', 'bad');
        showEnd(false);
      }
    });
  }

  /* ---------- 对话框 ---------- */
  function mask() {
    const d = document.createElement('div');
    d.className = 'dlg-mask';
    document.body.appendChild(d);
    dialogOpen = true;
    return d;
  }
  function closeMask(d) { d.remove(); dialogOpen = false; }

  function showHelp() {
    if (dialogOpen) return;
    const d = mask();
    d.innerHTML = `
      <div class="dlg wide">
        <h3>帮助</h3>
        <p><b>移动</b>：方向键 / WASD / 屏幕方向键<br>
           <b>开战/加速</b>：空格 · 点击 &nbsp; <b>取消</b>：Esc<br>
           <b>重开</b>：R &nbsp; <b>帮助</b>：H</p>
        <hr>
        <p><b>战斗规则</b>（经典魔塔）<br>
           伤害 = max(0, 攻 − 防)，玩家先手；<br>
           击杀当回合敌人不再反击。防御高于攻击时伤害为 0。</p>
        <hr>
        <p><b>钥匙</b>：红/蓝/金钥匙各开一扇同色门（消耗）。</p>
        <p><b>道具</b>：药水+1500HP · 大药水+4000HP · 剑+5攻 · 魔剑+20攻<br>
           盾+5防 · 圣盾+20防 · 金币+100 · 宝箱+800</p>
        <hr>
        <p><b>目标</b>：从 B1 一路打到 B3，击败龙王通关！<br>
           怪物不会重生，谨慎安排战斗顺序。</p>
        <div class="dlg-btns"><button id="help-close" class="btn primary">知道了</button></div>
      </div>`;
    d.querySelector('#help-close').onclick = () => closeMask(d);
    d.addEventListener('pointerdown', (e) => { if (e.target === d) closeMask(d); });
  }

  function showEnd(win) {
    S.over = true;
    const d = mask();
    const p = S.player;
    d.innerHTML = `
      <div class="dlg end ${win ? 'win' : 'lose'}">
        <h2>${win ? '🏆 通关！' : '💀 冒险结束'}</h2>
        <p>${win ? '龙王倒下了，魔塔恢复了和平。' : '你倒在了魔塔之中…按 R 重新开始。'}</p>
        <div class="end-stats">
          <span>剩余 HP ${p.hp}</span><span>攻击 ${p.atk}</span><span>防御 ${p.def}</span>
          <span>金币 ${p.gold}</span><span>到达 ${FLOOR_NAMES[S.floor]}</span>
        </div>
        <div class="dlg-btns"><button id="end-restart" class="btn primary">重新开始</button></div>
      </div>`;
    d.querySelector('#end-restart').onclick = () => { closeMask(d); restart(); };
  }

  function restart() {
    if (MT.battle.busy) return; // 战斗中不可重开
    document.querySelectorAll('.dlg-mask').forEach(d => d.remove());
    S = freshState();
    px = S.pos[0].x; py = S.pos[0].y;
    walk = null; facing = 'd';
    dialogOpen = false;
    addLog('冒险开始！在 B1 找到钥匙和武器。', 'sys');
    updatePanel(); draw();
  }

  /* ---------- 键盘 ---------- */
  function bindKeys() {
    window.addEventListener('keydown', (e) => {
      const k = e.key;
      if (k === 'ArrowUp' || k === 'w' || k === 'W') { e.preventDefault(); move(0, -1); }
      else if (k === 'ArrowDown' || k === 's' || k === 'S') { e.preventDefault(); move(0, 1); }
      else if (k === 'ArrowLeft' || k === 'a' || k === 'A') { e.preventDefault(); move(-1, 0); }
      else if (k === 'ArrowRight' || k === 'd' || k === 'D') { e.preventDefault(); move(1, 0); }
      else if (k === 'h' || k === 'H') showHelp();
      else if (k === 'r' || k === 'R') { if (!MT.battle.busy) restart(); }
      else if (k === 'Escape') {
        document.querySelectorAll('.dlg-mask:not(.pv-mask)').forEach(d => d.remove());
        if (!document.querySelector('.pv-mask')) dialogOpen = false;
      }
    });
  }

  /* ---------- 启动 ---------- */
  function init() {
    bindUI();
    bindKeys();
    restart();
    requestAnimationFrame(tick);
  }

  if (typeof document !== 'undefined') {
    if (document.readyState === 'loading') document.addEventListener('DOMContentLoaded', init);
    else init();
  }

  return { get state() { return S; }, move };
});
