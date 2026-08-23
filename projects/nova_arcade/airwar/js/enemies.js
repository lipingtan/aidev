'use strict';
/* =========================================================
 * 雷霆突袭 —— 敌机 / Boss / 敌方子弹
 * （player、boss、W、H、stageIdx 等由 game.js 提供，运行时可用）
 * ========================================================= */

const ENEMY_DEFS = {
  scout:   { hp: 20,  r: 14, score: 100, speed: 170, color: '#59d8ff' },
  fighter: { hp: 45,  r: 16, score: 200, speed: 130, color: '#ff5ecb' },
  gunship: { hp: 95,  r: 22, score: 350, speed: 75,  color: '#ffb84d' },
  bomber:  { hp: 230, r: 30, score: 600, speed: 55,  color: '#c77dff' },
  storm:   { hp: 130, r: 20, score: 500, speed: 90,  color: '#8ef0ff' },
};

const enemies = [];
const ebullets = [];

function spawnEnemy(type, x, y, opts = {}) {
  const d = ENEMY_DEFS[type];
  const hpMul = opts.hpMul || 1;
  const e = {
    type, x, y, vx: opts.vx || 0, vy: d.speed, r: d.r,
    hp: d.hp * hpMul, maxHp: d.hp * hpMul, score: d.score,
    t: rand(0, 6), phase: rand(0, TAU), fireCd: rand(1.2, 2.5),
    nextStrike: rand(1.6, 3.2), tele: false, flash: 0, dead: false,
  };
  enemies.push(e);
  return e;
}

function eShoot(x, y, angle, speed = 300, dmg = 10, type = 'ball', color = '#ff7a5e') {
  if (ebullets.length > 260) return;
  ebullets.push({
    x, y, vx: Math.cos(angle) * speed, vy: Math.sin(angle) * speed,
    r: type === 'shard' ? 4 : type === 'bolt' ? 7 : 5,
    dmg, type, color, t: 0,
  });
}

function updateEnemies(dt) {
  for (let i = enemies.length - 1; i >= 0; i--) {
    const e = enemies[i];
    if (e.dead) continue;
    e.t += dt;
    if (e.flash > 0) e.flash -= dt;
    // 侧翼入场：进入屏幕区域后逐渐转为垂直俯冲
    if (e.vx) {
      e.x += e.vx * dt;
      if (e.x > 90 && e.x < W - 90) e.vx *= Math.pow(0.03, dt);
    }
    switch (e.type) {
      case 'scout':
        e.x += Math.sin(e.t * 2.6 + e.phase) * 130 * dt;
        e.y += e.vy * dt;
        if (e.fireCd <= 0 && e.y > 30 && player.alive) {
          eShoot(e.x, e.y + e.r, Math.atan2(player.y - e.y, player.x - e.x), 260, 8, 'ball', '#7fe3ff');
          e.fireCd = rand(1.8, 3);
        }
        break;
      case 'fighter':
        e.y += e.vy * dt;
        e.x += (Math.sign(player.x - e.x) || 1) * 90 * dt;
        if (e.fireCd <= 0 && e.y > 30 && player.alive) {
          eShoot(e.x, e.y + e.r, Math.atan2(player.y - e.y, player.x - e.x), 320, 10, 'ball', '#ff9ad9');
          e.fireCd = rand(1.1, 1.8);
        }
        break;
      case 'gunship':
        e.y += e.vy * dt;
        e.x += Math.sin(e.t * 1.2 + e.phase) * 40 * dt;
        if (e.fireCd <= 0 && e.y > 30 && player.alive) {
          const base = Math.atan2(player.y - e.y, player.x - e.x);
          for (let k = -1; k <= 1; k++) eShoot(e.x, e.y + e.r * 0.6, base + k * 0.28, 280, 9, 'ball', '#ffc47a');
          e.fireCd = rand(2, 2.8);
        }
        break;
      case 'bomber':
        e.y += e.vy * dt;
        if (e.fireCd <= 0 && e.y > 40 && player.alive) {
          const base = Math.atan2(player.y - e.y, player.x - e.x);
          for (let k = -1.5; k <= 1.5; k += 1) eShoot(e.x, e.y + e.r * 0.6, base + k * 0.33, 240, 12, 'shard', '#d9a8ff');
          e.fireCd = rand(2.6, 3.4);
        }
        break;
      case 'storm': {
        // 在高空盘旋，蓄力后向玩家释放闪电
        const ty = 110 + Math.sin(e.t * 0.8 + e.phase) * 40;
        e.y += (ty - e.y) * Math.min(1, dt * 2);
        e.x += Math.sin(e.t * 0.6 + e.phase) * 70 * dt;
        e.nextStrike -= dt;
        if (e.nextStrike < 0.7 && !e.tele) { e.tele = true; SFX.zap(); }
        if (e.nextStrike <= 0) {
          e.tele = false;
          if (player.alive) {
            const a = Math.atan2(player.y - e.y, player.x - e.x);
            eShoot(e.x, e.y, a, 540, 13, 'bolt', '#bfe8ff');
            strike(e.x, e.y - 30, player.x + rand(-30, 30), player.y + rand(-20, 20), '#bfe8ff', 2);
            SFX.zap();
          }
          e.nextStrike = rand(2.6, 3.8);
        }
        break;
      }
    }
    if (e.y > H + 70 || e.x < -140 || e.x > W + 140) enemies.splice(i, 1);
  }
}

/* ---------- Boss ---------- */
function makeBoss(idx) {
  const d = BOSS_DEFS[idx];
  return {
    type: 'boss', key: 'b' + (idx + 1), name: d.name,
    x: W / 2, y: -100, targetY: 120, r: d.r, color: d.color,
    hp: d.hp, maxHp: d.hp, t: 0, flash: 0,
    patIdx: 0, patT: 2.5, lightT: 0, laser: null, enrage: false, dead: false,
  };
}

const BOSS_PATTERNS = {
  b1: ['radial', 'aimed', 'summon'],
  b2: ['lightning', 'aimed', 'radial', 'summon'],
  b3: ['laser', 'radial', 'shards', 'summon'],
  b4: ['lightning', 'laser', 'radial', 'shards', 'summon'],
};

function doBossPattern(b) {
  const list = BOSS_PATTERNS[b.key];
  const p = list[b.patIdx % list.length];
  b.patIdx++;
  if (p === 'radial') {
    const n = Math.round((14 + stageIdx * 4) * (b.enrage ? 1.5 : 1));
    for (let i = 0; i < n; i++) eShoot(b.x, b.y, (i / n) * TAU + b.t * 0.3, 230, 10, 'ball', b.color);
  } else if (p === 'aimed') {
    const base = Math.atan2(player.y - b.y, player.x - b.x);
    for (let i = 0; i < 5; i++) eShoot(b.x, b.y + 20, base + (i - 2) * 0.24, 330, 10, 'ball', b.color);
  } else if (p === 'shards') {
    const dir = b.t % TAU;
    for (let i = 0; i < 8; i++) eShoot(b.x, b.y, dir + (i - 3.5) * 0.14, 420, 9, 'shard', '#ffb84d');
  } else if (p === 'summon') {
    if (enemies.length < 26) {
      spawnEnemy('scout', b.x - 90, b.y + 20);
      spawnEnemy('scout', b.x + 90, b.y + 20);
    }
  } else if (p === 'lightning') {
    b.lightT = 1.1;
    SFX.warn();
  } else if (p === 'laser') {
    b.laser = { t: 2.8, dur: 2.8, base: Math.atan2(player.y - b.y, player.x - b.x), amp: 0.6, ang: 0 };
    SFX.zap();
  }
}

function updateBoss(dt) {
  const b = boss;
  if (!b || b.dead) return;
  b.t += dt;
  if (b.flash > 0) b.flash -= dt;
  // 入场
  if (b.y < b.targetY) { b.y += 60 * dt; return; }
  // 横向巡逻
  b.x = W / 2 + Math.sin(b.t * 0.45) * W * 0.3;
  // 狂暴阶段
  if (!b.enrage && b.hp < b.maxHp * 0.5) {
    b.enrage = true; SFX.warn(); addShake(8);
    addFloater(b.x, b.y - b.r - 20, '!! 狂暴化 !!', '#ff5e6e', 20);
  }
  // 闪电蓄力 → 释放
  if (b.lightT > 0) {
    b.lightT -= dt;
    if (b.lightT <= 0) {
      const a = Math.atan2(player.y - b.y, player.x - b.x);
      for (let i = -1; i <= 1; i++) eShoot(b.x, b.y + 10, a + i * 0.3, 560, 14, 'bolt', '#bfe8ff');
      strike(b.x, b.y - 20, player.x + rand(-40, 40), player.y + rand(-30, 30), '#bfe8ff', 2.5);
      SFX.zap(); flashScreen = Math.max(flashScreen, 0.5); addShake(8);
    }
  }
  // 横扫激光
  if (b.laser) {
    b.laser.t -= dt;
    if (b.laser.t <= 0) b.laser = null;
    else {
      const p = 1 - b.laser.t / b.laser.dur;
      b.laser.ang = b.laser.base + Math.sin(p * Math.PI * 2) * b.laser.amp;
      const dx = player.x - b.x, dy = player.y - b.y;
      const proj = dx * Math.cos(b.laser.ang) + dy * Math.sin(b.laser.ang);
      if (proj > 0 && player.alive) {
        const perp = Math.abs(-dx * Math.sin(b.laser.ang) + dy * Math.cos(b.laser.ang));
        if (perp < 16 + player.r && player.shield <= 0) {
          player.hp -= 30 * dt;
          if (player.hp <= 0) { player.hp = 0; killPlayer(); }
        }
      }
    }
  }
  // 攻击模式轮换
  b.patT -= dt;
  if (b.patT <= 0) {
    b.patT = b.enrage ? 2.4 : 3.6;
    doBossPattern(b);
  }
}

/* ---------- 绘制 ---------- */
function drawEnemyShape(c, e) {
  c.save();
  c.translate(e.x, e.y);
  c.rotate(clamp((e.vx || 0) * 0.002, -0.3, 0.3));
  const col = ENEMY_DEFS[e.type].color;
  switch (e.type) {
    case 'scout':
      c.fillStyle = '#123a52'; c.strokeStyle = col; c.lineWidth = 2;
      c.shadowColor = col; c.shadowBlur = 10;
      c.beginPath();
      c.moveTo(0, 16); c.lineTo(8, 4); c.lineTo(15, -9); c.lineTo(4, -6); c.lineTo(0, -13);
      c.lineTo(-4, -6); c.lineTo(-15, -9); c.lineTo(-8, 4);
      c.closePath(); c.fill(); c.stroke(); c.shadowBlur = 0;
      c.fillStyle = col; c.beginPath(); c.arc(0, 2, 3, 0, TAU); c.fill();
      break;
    case 'fighter':
      c.fillStyle = '#4a1240'; c.strokeStyle = col; c.lineWidth = 2;
      c.shadowColor = col; c.shadowBlur = 10;
      c.beginPath();
      c.moveTo(0, 18); c.lineTo(6, 8); c.lineTo(20, -6); c.lineTo(7, -9); c.lineTo(0, -16);
      c.lineTo(-7, -9); c.lineTo(-20, -6); c.lineTo(-6, 8);
      c.closePath(); c.fill(); c.stroke(); c.shadowBlur = 0;
      c.fillStyle = col; c.fillRect(-2, -10, 4, 16);
      break;
    case 'gunship':
      c.fillStyle = '#4a2c10'; c.strokeStyle = col; c.lineWidth = 2.5;
      c.shadowColor = col; c.shadowBlur = 12;
      c.beginPath();
      for (let i = 0; i < 6; i++) {
        const a = (i / 6) * TAU + Math.PI / 6;
        const x = Math.cos(a) * 20, y = Math.sin(a) * 20;
        i ? c.lineTo(x, y) : c.moveTo(x, y);
      }
      c.closePath(); c.fill(); c.stroke(); c.shadowBlur = 0;
      for (const s of [-1, 1]) { c.fillStyle = col; c.beginPath(); c.arc(s * 24, 6, 5, 0, TAU); c.fill(); }
      c.fillStyle = '#fff';
      c.beginPath(); c.arc(0, 0, 4 + Math.sin(e.t * 5) * 1.5, 0, TAU); c.fill();
      break;
    case 'bomber':
      c.fillStyle = '#2c1845'; c.strokeStyle = col; c.lineWidth = 2.5;
      c.shadowColor = col; c.shadowBlur = 14;
      rr(c, -30, -14, 60, 30, 10); c.fill(); c.stroke(); c.shadowBlur = 0;
      for (let i = 0; i < 4; i++) { c.fillStyle = col; c.beginPath(); c.arc(-21 + i * 14, 18, 4, 0, TAU); c.fill(); }
      c.fillStyle = col;
      c.beginPath(); c.moveTo(0, -14); c.lineTo(10, -26); c.lineTo(-10, -26); c.closePath(); c.fill();
      break;
    case 'storm': {
      const chg = e.nextStrike < 0.7;
      if (chg) {
        c.strokeStyle = `rgba(191,232,255,${0.4 + Math.sin(e.t * 25) * 0.3})`;
        c.lineWidth = 3;
        c.beginPath(); c.arc(0, 0, 26 + Math.sin(e.t * 20) * 4, 0, TAU); c.stroke();
      }
      c.fillStyle = '#0e3a48'; c.strokeStyle = col; c.lineWidth = 2;
      c.shadowColor = '#bfe8ff'; c.shadowBlur = chg ? 22 : 12;
      c.beginPath(); c.moveTo(0, -17); c.lineTo(13, 0); c.lineTo(0, 17); c.lineTo(-13, 0);
      c.closePath(); c.fill(); c.stroke(); c.shadowBlur = 0;
      c.fillStyle = chg ? '#ffffff' : col;
      c.beginPath(); c.arc(0, 0, 4.5, 0, TAU); c.fill();
      for (let i = 0; i < 3; i++) {
        const a = e.t * 3 + i * TAU / 3;
        c.fillStyle = '#bfe8ff';
        c.beginPath(); c.arc(Math.cos(a) * 21, Math.sin(a) * 21, 2.2, 0, TAU); c.fill();
      }
      break;
    }
  }
  if (e.flash > 0) {
    c.globalAlpha = Math.min(0.7, e.flash * 6);
    c.fillStyle = '#fff';
    c.beginPath(); c.arc(0, 0, e.r, 0, TAU); c.fill();
    c.globalAlpha = 1;
  }
  c.restore();
}

function drawEnemies(c) {
  for (const e of enemies) if (!e.dead) drawEnemyShape(c, e);
}

const BOSS_SHAPES = {
  b1: [[0, 52], [20, 18], [52, -6], [34, -34], [12, -26], [0, -40], [-12, -26], [-34, -34], [-52, -6], [-20, 18]],
  b2: [[0, 58], [14, 20], [44, 4], [62, -22], [30, -30], [16, -52], [-16, -52], [-30, -30], [-62, -22], [-44, 4], [-14, 20]],
  b3: [[-60, -36], [60, -36], [90, 0], [60, 36], [-60, 36], [-90, 0]],
  b4: [[0, 60], [18, 24], [52, 30], [34, 0], [58, -34], [20, -26], [0, -58], [-20, -26], [-58, -34], [-34, 0], [-52, 30], [-18, 24]],
};

function drawBoss(c) {
  const b = boss;
  if (!b || b.dead || b.y < -b.r) return;
  c.save();
  c.translate(b.x, b.y);
  const s = b.r / 56;
  // 狂暴光环
  if (b.enrage) {
    c.strokeStyle = `rgba(255,60,80,${0.35 + Math.sin(b.t * 8) * 0.2})`;
    c.lineWidth = 6;
    c.beginPath(); c.arc(0, 0, b.r * 1.25, 0, TAU); c.stroke();
  }
  // 主体
  const pts = BOSS_SHAPES[b.key];
  const g = c.createLinearGradient(0, -b.r, 0, b.r);
  g.addColorStop(0, 'rgba(20,30,50,0.95)');
  g.addColorStop(1, 'rgba(40,55,90,0.95)');
  c.fillStyle = g;
  c.strokeStyle = b.color;
  c.lineWidth = 3;
  c.shadowColor = b.color; c.shadowBlur = 22;
  c.beginPath();
  for (let i = 0; i < pts.length; i++) {
    const x = pts[i][0] * s, y = pts[i][1] * s;
    i ? c.lineTo(x, y) : c.moveTo(x, y);
  }
  c.closePath(); c.fill(); c.stroke();
  c.shadowBlur = 0;
  // 旋转炮塔
  for (let i = 0; i < 3; i++) {
    const a = b.t * 0.8 + i * TAU / 3;
    const tx = Math.cos(a) * b.r * 0.62, ty = Math.sin(a) * b.r * 0.5;
    c.fillStyle = b.color;
    c.beginPath(); c.arc(tx, ty, 9 * s * 0.8, 0, TAU); c.fill();
    c.fillStyle = 'rgba(255,255,255,0.7)';
    c.beginPath(); c.arc(tx, ty, 3.5, 0, TAU); c.fill();
  }
  // 核心
  const pulse = 0.8 + Math.sin(b.t * 4) * 0.2;
  const cg = c.createRadialGradient(0, 0, 1, 0, 0, 20 * s * 0.7);
  cg.addColorStop(0, '#ffffff');
  cg.addColorStop(0.5, b.color);
  cg.addColorStop(1, 'rgba(0,0,0,0)');
  c.fillStyle = cg;
  c.beginPath(); c.arc(0, 0, 20 * s * 0.7 * pulse + 6, 0, TAU); c.fill();
  // 受击闪白
  if (b.flash > 0) {
    c.globalAlpha = Math.min(0.6, b.flash * 5);
    c.fillStyle = '#fff';
    c.beginPath(); c.arc(0, 0, b.r, 0, TAU); c.fill();
    c.globalAlpha = 1;
  }
  c.restore();
}

function drawBossExtras(c) {
  const b = boss;
  if (!b || b.dead || b.y < 0) return;
  // 闪电蓄力警告线
  if (b.lightT > 0) {
    const a = 0.35 + Math.sin(performance.now() / 60) * 0.3;
    c.strokeStyle = `rgba(255,80,90,${a})`;
    c.lineWidth = 3;
    c.setLineDash([10, 8]);
    c.beginPath(); c.moveTo(b.x, b.y); c.lineTo(player.x, player.y); c.stroke();
    c.setLineDash([]);
  }
  // 横扫激光
  if (b.laser) {
    const ang = b.laser.ang;
    c.save();
    c.translate(b.x, b.y);
    c.rotate(ang);
    const g = c.createLinearGradient(0, 0, 720, 0);
    g.addColorStop(0, 'rgba(255,80,120,0.9)');
    g.addColorStop(1, 'rgba(255,80,120,0)');
    c.fillStyle = g;
    c.fillRect(0, -14, 720, 28);
    c.fillStyle = 'rgba(255,225,235,0.9)';
    c.fillRect(0, -4, 720, 8);
    c.restore();
  }
}

function drawEbullets(c) {
  for (const b of ebullets) {
    if (b.type === 'bolt') {
      const a = Math.atan2(b.vy, b.vx);
      c.save(); c.translate(b.x, b.y); c.rotate(a);
      c.strokeStyle = '#e8f9ff'; c.lineWidth = 3;
      c.shadowColor = '#7fd4ff'; c.shadowBlur = 12;
      c.beginPath();
      c.moveTo(-14, 0); c.lineTo(-7, rand(-3, 3)); c.lineTo(0, rand(-3, 3)); c.lineTo(10, 0);
      c.stroke();
      c.restore();
    } else if (b.type === 'shard') {
      const a = Math.atan2(b.vy, b.vx) + b.t * 8;
      c.save(); c.translate(b.x, b.y); c.rotate(a);
      c.fillStyle = b.color; c.shadowColor = b.color; c.shadowBlur = 8;
      c.fillRect(-7, -2, 14, 4);
      c.restore();
    } else {
      c.fillStyle = b.color;
      c.shadowColor = b.color; c.shadowBlur = 8;
      c.beginPath(); c.arc(b.x, b.y, b.r, 0, TAU); c.fill();
      c.shadowBlur = 0;
      c.fillStyle = 'rgba(255,255,255,0.8)';
      c.beginPath(); c.arc(b.x, b.y, b.r * 0.4, 0, TAU); c.fill();
    }
  }
  c.shadowBlur = 0;
}
