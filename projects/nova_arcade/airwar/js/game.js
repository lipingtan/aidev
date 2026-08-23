'use strict';
/* =========================================================
 * 雷霆突袭 THUNDER STRIKE —— 主程序
 * 状态机 / 玩家 / 武器 / 补给 / HUD / 渲染
 * ========================================================= */
const canvas = document.getElementById('game');
const ctx = canvas.getContext('2d');
const W = canvas.width, H = canvas.height;

function fitCanvas() {
  const s = Math.min(window.innerWidth / W, window.innerHeight / H) * 0.98;
  canvas.style.width = Math.floor(W * s) + 'px';
  canvas.style.height = Math.floor(H * s) + 'px';
}
window.addEventListener('resize', fitCanvas);
fitCanvas();

/* ---------------- 全局状态 ---------------- */
let state = 'menu'; // menu | intro | play | clear | over | win
let pause = false;
let stageIdx = 0, stageTime = 0, wavePtr = 0;
let score = 0;
let best = 0;
try { best = +(localStorage.getItem('ts_best') || 0); } catch (e) {}
function saveBest() { try { localStorage.setItem('ts_best', String(best)); } catch (e) {} }

let slowT = 0, timeScale = 1;
let introT = 0, clearT = 0, winT = 0, dieT = 0;
let bossPending = false, bossWarnT = 0;
let mutedMsgT = 0;
let laserTick = 0;

let boss = null;
const keys = {};

const player = {
  x: W / 2, y: H - 110, vx: 0, vy: 0, r: 11,
  hp: 100, maxHp: 100, speed: 340,
  weapon: 'normal',
  levels: { normal: 1, spread: 1, laser: 1, homing: 1 },
  bombs: 3, fireCd: 0, invuln: 2.5, shield: 0, tilt: 0, alive: true, trailT: 0,
};

const pbullets = [];
const drops = [];
const DROP_INFO = {
  health: ['HP', '#5eff9a'],
  power: ['W', '#ffd24d'],
  shield: ['S', '#7fe3ff'],
  bomb: ['B', '#ff9a5e'],
  score: ['$', '#ffe29a'],
};

/* ---------------- 背景 ---------------- */
const stars = [];
for (let i = 0; i < 110; i++) stars.push({ x: rand(0, W), y: rand(0, H), z: pick([1, 2, 3]) });
const clouds = [];
for (let i = 0; i < 7; i++) clouds.push({ x: rand(0, W), y: rand(40, H * 0.7), w: rand(180, 420), h: rand(30, 70), s: rand(8, 26) });
const MENU_BG = { bgTop: '#050a1c', bgBottom: '#0d1b3a', clouds: false };

function updateStars(dt) {
  for (const s of stars) {
    s.y += (26 + s.z * 46) * dt;
    if (s.y > H + 4) { s.y = -4; s.x = rand(0, W); }
  }
  for (const cl of clouds) {
    cl.x -= cl.s * dt;
    if (cl.x < -cl.w) { cl.x = W + cl.w; cl.y = rand(40, H * 0.7); }
  }
}

/* ---------------- 流程控制 ---------------- */
function startGame() {
  score = 0;
  stageIdx = 0;
  player.levels = { normal: 1, spread: 1, laser: 1, homing: 1 };
  player.bombs = 3;
  loadStage();
}

function loadStage() {
  stageTime = 0; wavePtr = 0;
  enemies.length = 0; ebullets.length = 0; pbullets.length = 0; drops.length = 0;
  particles.length = 0; bolts.length = 0; floaters.length = 0;
  boss = null; bossPending = false; bossWarnT = 0; slowT = 0; timeScale = 1; pause = false;
  Object.assign(player, { x: W / 2, y: H - 110, vx: 0, vy: 0, hp: player.maxHp, invuln: 2.5, alive: true, shield: 0 });
  state = 'intro'; introT = 0;
}

function nextStage() {
  // 未拾取的补给直接结算，避免换幕丢失奖励
  for (const d of drops) applyPickup(d);
  drops.length = 0;
  if (stageIdx >= STAGES.length - 1) {
    best = Math.max(best, score); saveBest();
    state = 'win'; winT = 0;
  } else {
    stageIdx++;
    loadStage();
  }
}

function switchWeapon(i) {
  if (i < 0 || i >= WEAPONS.length) return;
  player.weapon = WEAPONS[i].id;
  addFloater(player.x, player.y - 34, WEAPONS[i].name + ' Lv' + player.levels[WEAPONS[i].id], '#7fe3ff', 15);
  SFX.tone(600, 0.08, 'square', 0.1);
}

/* ---------------- 玩家 ---------------- */
function updatePlayer(dt) {
  if (!player.alive) return;
  player.invuln = Math.max(0, player.invuln - dt);
  player.shield = Math.max(0, player.shield - dt);
  let tx = 0, ty = 0;
  if (keys['a'] || keys['arrowleft']) tx -= 1;
  if (keys['d'] || keys['arrowright']) tx += 1;
  if (keys['w'] || keys['arrowup']) ty -= 1;
  if (keys['s'] || keys['arrowdown']) ty += 1;
  const len = Math.hypot(tx, ty) || 1;
  const k = 1 - Math.pow(0.002, dt);
  player.vx = lerp(player.vx, (tx / len) * player.speed, k);
  player.vy = lerp(player.vy, (ty / len) * player.speed, k);
  player.x = clamp(player.x + player.vx * dt, 18, W - 18);
  player.y = clamp(player.y + player.vy * dt, 30, H - 24);
  player.tilt = lerp(player.tilt, (player.vx / player.speed) * 0.4, k);

  // 引擎尾焰
  player.trailT -= dt;
  if (player.trailT <= 0) {
    player.trailT = 0.02;
    for (const sx of [-5, 5]) {
      addParticle({ type: 'spark', x: player.x + sx + rand(-2, 2), y: player.y + 16, vx: rand(-20, 20) - player.vx * 0.1, vy: rand(60, 140), life: rand(0.2, 0.4), max: 0.4, size: 2.5, color: pick(['#7fe3ff', '#ffd27f']) });
    }
  }

  // 自动开火
  player.fireCd -= dt;
  const lvl = player.levels[player.weapon];
  if (player.weapon === 'normal' && player.fireCd <= 0) {
    player.fireCd = 0.15;
    spawnNormal(lvl); SFX.shoot('normal');
  } else if (player.weapon === 'spread' && player.fireCd <= 0) {
    player.fireCd = 0.42;
    spawnSpread(lvl); SFX.shoot('spread');
  } else if (player.weapon === 'homing' && player.fireCd <= 0) {
    player.fireCd = Math.max(0.25, 0.6 - lvl * 0.06);
    spawnHoming(lvl); SFX.shoot('homing');
  } else if (player.weapon === 'laser') {
    laserTick -= dt;
    if (laserTick <= 0) { laserTick = 0.28; SFX.tone(180, 0.2, 'sawtooth', 0.05, 120); }
  }
}

function spawnNormal(lvl) {
  const dmg = 10 + lvl * 3, sp = 760;
  const patterns = [[0], [-9, 9], [-11, 0, 11], [-13, -4.5, 4.5, 13], [-15, -5, 5, 15]];
  for (const ox of patterns[lvl - 1]) {
    pbullets.push({ type: 'bullet', x: player.x + ox, y: player.y - 18, vx: 0, vy: -sp, r: 4, dmg, color: '#7fe3ff' });
  }
  if (lvl >= 5) {
    for (const s of [-1, 1]) {
      pbullets.push({ type: 'bullet', x: player.x + s * 20, y: player.y - 6, vx: Math.sin(s * 0.3) * sp, vy: -Math.cos(0.3) * sp, r: 4, dmg, color: '#7fe3ff' });
    }
  }
}

function spawnSpread(lvl) {
  const n = 4 + lvl; // 5~9 发
  for (let i = 0; i < n; i++) {
    const a = -Math.PI / 2 + (i / (n - 1) - 0.5) * 1.0;
    pbullets.push({ type: 'pellet', x: player.x, y: player.y - 14, vx: Math.cos(a) * 560, vy: Math.sin(a) * 560, r: 3.5, dmg: 8 + lvl * 2, color: '#ffd27f' });
  }
}

function spawnHoming(lvl) {
  const n = lvl; // 1~5 枚
  for (let i = 0; i < n; i++) {
    const ox = (i - (n - 1) / 2) * 14;
    pbullets.push({ type: 'missile', x: player.x + ox, y: player.y + 6, vx: rand(-30, 30), vy: -120, spd: 440, r: 6, dmg: 22 + lvl * 5, life: 2.6, color: '#ff9a3d' });
  }
}

function nearestTarget(x, y) {
  let t = null, td = 1e9;
  for (const e of enemies) {
    if (e.dead) continue;
    const d2 = (e.x - x) ** 2 + (e.y - y) ** 2;
    if (d2 < td) { td = d2; t = e; }
  }
  if (boss && boss.y > 0) {
    const d2 = (boss.x - x) ** 2 + (boss.y - y) ** 2;
    if (d2 < td) t = boss;
  }
  return t;
}

function useBomb() {
  if (!player.alive || player.bombs <= 0) return;
  player.bombs--;
  SFX.bomb(); addShake(24); flashScreen = 1; slowT = 0.5;
  addParticle({ type: 'ring', x: player.x, y: player.y, vx: 0, vy: 0, life: 0.7, max: 0.7, size: 24, grow: 1100, color: 'rgba(140,220,255,' });
  for (const e of enemies) if (!e.dead) damageEnemy(e, 80);
  if (boss && !boss.dead) damageBoss(80);
  for (const b of ebullets) { sparkBurst(b.x, b.y, 3, '#bfe8ff'); score += 5; }
  ebullets.length = 0;
  addFloater(player.x, player.y - 40, '核爆冲击!!', '#bfe8ff', 22);
}

function damagePlayer(dmg) {
  if (!player.alive || player.invuln > 0) return;
  if (player.shield > 0) {
    sparkBurst(player.x, player.y, 10, '#7fe3ff');
    SFX.hit();
    return;
  }
  player.hp -= dmg;
  player.invuln = 1.6;
  SFX.hit(); addShake(9); flashScreen = Math.max(flashScreen, 0.25);
  sparkBurst(player.x, player.y, 14, '#ff7a5e');
  if (player.hp <= 0) { player.hp = 0; killPlayer(); }
}

function killPlayer() {
  if (!player.alive) return;
  player.alive = false;
  dieT = 1.8;
  SFX.bigBoom(); addShake(26); flashScreen = 1;
  for (let i = 0; i < 7; i++) {
    explosion(player.x + rand(-50, 50), player.y + rand(-40, 40), rand(0.8, 1.8));
  }
}

/* ---------------- 伤害 / 击杀 ---------------- */
function damageEnemy(e, dmg) {
  if (e.dead) return;
  e.hp -= dmg;
  e.flash = 0.12;
  if (e.hp <= 0) killEnemy(e);
}

function killEnemy(e) {
  e.dead = true;
  const size = clamp(e.r / 14, 0.7, 2);
  explosion(e.x, e.y, size);
  SFX.boom(size); addShake(3 * size);
  score += e.score;
  addFloater(e.x, e.y, '+' + e.score, '#ffe29a', 14);
  rollDrop(e.x, e.y);
}

function damageBoss(dmg) {
  if (!boss || boss.dead) return;
  boss.hp -= dmg;
  boss.flash = 0.1;
  if (boss.hp <= 0) killBoss();
}

function killBoss() {
  const b = boss;
  if (!b || b.dead) return;
  b.dead = true;
  SFX.bigBoom(); addShake(26); flashScreen = 1; slowT = 0.8;
  for (let i = 0; i < 9; i++) {
    explosion(b.x + rand(-b.r, b.r), b.y + rand(-b.r * 0.7, b.r * 0.7), rand(0.9, 1.8));
  }
  // 连锁闪电
  for (let i = 0; i < 5; i++) {
    strike(b.x + rand(-90, 90), b.y + rand(-70, 70), b.x + rand(-140, 140), b.y + rand(-100, 100), '#bfe8ff', 2);
  }
  const gained = 3000 + stageIdx * 2000;
  score += gained;
  addFloater(b.x, b.y, '+' + gained, '#ffd24d', 26);
  spawnDrop(b.x - 50, b.y, 'health');
  spawnDrop(b.x, b.y + 30, 'power');
  spawnDrop(b.x + 50, b.y, 'bomb');
  ebullets.length = 0;
  boss = null;
  state = 'clear'; clearT = 0;
}

/* ---------------- 补给掉落 ---------------- */
function rollDrop(x, y) {
  if (Math.random() >= 0.16) return;
  const r = Math.random();
  const type = r < 0.28 ? 'health' : r < 0.52 ? 'power' : r < 0.66 ? 'shield' : r < 0.8 ? 'bomb' : 'score';
  spawnDrop(x, y, type);
}

function spawnDrop(x, y, type) {
  drops.push({ x: clamp(x, 20, W - 20), y: clamp(y, 20, H - 20), t: rand(0, TAU), type });
}

function applyPickup(d) {
  const p = player;
  if (d.type === 'health') {
    if (p.hp >= p.maxHp) { score += 200; addFloater(p.x, p.y - 30, '+200', '#5eff9a'); }
    else { p.hp = Math.min(p.maxHp, p.hp + 30); addFloater(p.x, p.y - 30, '+30 HP', '#5eff9a'); }
  } else if (d.type === 'power') {
    const w = p.weapon;
    if (p.levels[w] < 5) { p.levels[w]++; addFloater(p.x, p.y - 30, WEAPONS.find(x => x.id === w).name + ' Lv' + p.levels[w], '#ffd24d'); }
    else { score += 500; addFloater(p.x, p.y - 30, '+500', '#ffd24d'); }
  } else if (d.type === 'shield') {
    p.shield = 8;
    addFloater(p.x, p.y - 30, '护盾展开!', '#7fe3ff');
  } else if (d.type === 'bomb') {
    if (p.bombs < 6) { p.bombs++; addFloater(p.x, p.y - 30, '+1 炸弹', '#ff9a5e'); }
    else { score += 300; addFloater(p.x, p.y - 30, '+300', '#ff9a5e'); }
  } else {
    score += 500;
    addFloater(p.x, p.y - 30, '+500', '#ffe29a');
  }
  SFX.pickup();
  sparkBurst(d.x, d.y, 10, DROP_INFO[d.type][1]);
}

function updateDrops(dt) {
  for (let i = drops.length - 1; i >= 0; i--) {
    const d = drops[i];
    d.t += dt;
    const dx = player.x - d.x, dy = player.y - d.y;
    const dd = Math.hypot(dx, dy) || 1;
    if (player.alive && dd < 150) { d.x += (dx / dd) * 240 * dt; d.y += (dy / dd) * 240 * dt; }
    else d.y += 65 * dt;
    if (d.y > H + 30 || d.x < -30 || d.x > W + 30) { drops.splice(i, 1); continue; }
    if (player.alive && dd < 24) { applyPickup(d); drops.splice(i, 1); }
  }
}

/* ---------------- 子弹 ---------------- */
function updatePbullets(dt) {
  for (let i = pbullets.length - 1; i >= 0; i--) {
    const b = pbullets[i];
    if (b.type === 'missile') {
      b.life -= dt;
      const t = nearestTarget(b.x, b.y);
      if (t) {
        const want = Math.atan2(t.y - b.y, t.x - b.x);
        const cur = Math.atan2(b.vy, b.vx);
        const d = angDiff(want, cur);
        const turn = 5.5 * dt;
        const na = cur + clamp(d, -turn, turn);
        b.vx = Math.cos(na) * b.spd; b.vy = Math.sin(na) * b.spd;
      }
      b.x += b.vx * dt; b.y += b.vy * dt;
      if (Math.random() < 0.8) {
        addParticle({ type: 'spark', x: b.x, y: b.y, vx: rand(-15, 15), vy: rand(-15, 15), life: 0.25, max: 0.25, size: 2, color: '#ffb15e' });
      }
      if (b.life <= 0 || b.y < -40 || b.x < -40 || b.x > W + 40) {
        explosion(b.x, b.y, 0.55, ['#ffd27f', '#ff9a3d']);
        SFX.boom(0.6);
        pbullets.splice(i, 1); continue;
      }
    } else {
      b.x += b.vx * dt; b.y += b.vy * dt;
      if (b.y < -30 || b.x < -30 || b.x > W + 30) { pbullets.splice(i, 1); continue; }
    }
    // 命中判定
    let hit = false;
    for (const e of enemies) {
      if (e.dead) continue;
      const rr2 = e.r + b.r;
      if ((b.x - e.x) ** 2 + (b.y - e.y) ** 2 < rr2 * rr2) {
        damageEnemy(e, b.dmg);
        if (b.type === 'missile') { explosion(b.x, b.y, 0.7, ['#ffd27f', '#ff9a3d']); SFX.boom(0.7); }
        else sparkBurst(b.x, b.y, 3, b.color);
        hit = true; break;
      }
    }
    if (!hit && boss && !boss.dead && boss.y > -10) {
      const rr2 = boss.r * 0.85 + b.r;
      if ((b.x - boss.x) ** 2 + (b.y - boss.y) ** 2 < rr2 * rr2) {
        damageBoss(b.dmg);
        if (b.type === 'missile') { explosion(b.x, b.y, 0.7, ['#ffd27f', '#ff9a3d']); SFX.boom(0.7); }
        else sparkBurst(b.x, b.y, 3, b.color);
        hit = true;
      }
    }
    if (hit) pbullets.splice(i, 1);
  }
}

function laserDamage(dt) {
  if (!player.alive || player.weapon !== 'laser') return;
  const lvl = player.levels.laser;
  const dps = 26 + lvl * 14;
  const beams = [[0, 7 + lvl * 3]];
  if (lvl >= 4) beams.push([-26, 5], [26, 5]);
  for (const [ox, hw] of beams) {
    const bx = player.x + ox;
    for (const e of enemies) {
      if (!e.dead && e.y < player.y && Math.abs(e.x - bx) < hw + e.r) {
        damageEnemy(e, dps * dt);
        if (Math.random() < 0.3) sparkBurst(bx, e.y + e.r, 2, '#7fe3ff');
      }
    }
    if (boss && !boss.dead && boss.y > 0 && boss.y < player.y && Math.abs(boss.x - bx) < hw + boss.r * 0.9) {
      damageBoss(dps * dt);
    }
  }
}

function updateEbullets(dt) {
  for (let i = ebullets.length - 1; i >= 0; i--) {
    const b = ebullets[i];
    b.x += b.vx * dt; b.y += b.vy * dt; b.t += dt;
    if (b.y < -40 || b.y > H + 40 || b.x < -40 || b.x > W + 40) { ebullets.splice(i, 1); continue; }
    if (player.alive && player.invuln <= 0) {
      const rr2 = b.r + player.r;
      if ((b.x - player.x) ** 2 + (b.y - player.y) ** 2 < rr2 * rr2) {
        ebullets.splice(i, 1);
        damagePlayer(b.dmg);
      }
    }
  }
}

/* ---------------- 波次 / Boss 触发 ---------------- */
function spawnWave(w) {
  if (enemies.length > 34) return;
  const hpMul = 1 + stageIdx * 0.35;
  for (let i = 0; i < w.n; i++) {
    let x, y, vx = 0;
    if (w.pattern === 'line') {
      x = W / 2 + (i - (w.n - 1) / 2) * 48; y = -30 - rand(0, 20);
    } else if (w.pattern === 'v') {
      const off = i - (w.n - 1) / 2;
      x = W / 2 + off * 54; y = -30 - Math.abs(off) * 26;
    } else {
      x = i % 2 === 0 ? -30 : W + 30;
      y = rand(50, 220);
      vx = (x < 0 ? 1 : -1) * rand(120, 170);
    }
    spawnEnemy(w.type, x, y, { vx, hpMul });
  }
}

/* ---------------- 天雷（随机环境闪电） ---------------- */
function ambientLightning(dt) {
  const rate = STAGES[stageIdx].lightningRate || 0;
  if (Math.random() >= rate * dt) return;
  let tx, ty;
  if (enemies.length && Math.random() < 0.7) {
    const e = pick(enemies);
    tx = e.x; ty = e.y;
  } else {
    tx = rand(60, W - 60); ty = rand(40, H * 0.5);
  }
  strike(tx + rand(-30, 30), -10, tx, ty, '#bfe8ff', 2.5);
  flashScreen = Math.max(flashScreen, 0.35);
  SFX.zap(); addShake(5);
  const others = enemies.filter(e => !e.dead);
  for (const e of others) {
    if (dist(e.x, e.y, tx, ty) < 120) {
      damageEnemy(e, 40);
      if (others.length > 1 && Math.random() < 0.8) {
        const o = pick(others.filter(x => x !== e));
        if (o) strike(tx, ty, o.x, o.y, '#bfe8ff', 1.5); // 连锁闪电
      }
    }
  }
  if (boss && !boss.dead && dist(boss.x, boss.y, tx, ty) < 140) damageBoss(30);
}

/* ---------------- 主更新 ---------------- */
function updatePlay(dt, gdt) {
  stageTime += gdt;
  const st = STAGES[stageIdx];
  while (wavePtr < st.waves.length && st.waves[wavePtr].t <= stageTime) {
    spawnWave(st.waves[wavePtr]);
    wavePtr++;
  }
  if (!boss && !bossPending && wavePtr >= st.waves.length && enemies.length === 0) {
    bossPending = true; bossWarnT = 2.4; SFX.warn();
  }
  if (bossPending && !boss) {
    bossWarnT -= gdt;
    if (bossWarnT <= 0) { boss = makeBoss(stageIdx); bossPending = false; }
  }

  updatePlayer(gdt);
  laserDamage(gdt);
  updatePbullets(gdt);
  updateEnemies(gdt);
  updateBoss(gdt);
  updateEbullets(gdt);
  updateDrops(gdt);
  ambientLightning(gdt);

  // 机体碰撞
  if (player.alive && player.invuln <= 0) {
    for (const e of enemies) {
      if (!e.dead && dist(e.x, e.y, player.x, player.y) < e.r * 0.8 + player.r) {
        damagePlayer(15); damageEnemy(e, 30); break;
      }
    }
    if (boss && !boss.dead && boss.y > 0 && dist(boss.x, boss.y, player.x, player.y) < boss.r * 0.85 + player.r) {
      damagePlayer(25);
    }
  }

  for (let i = enemies.length - 1; i >= 0; i--) if (enemies[i].dead) enemies.splice(i, 1);

  updateParticles(gdt);
  updateFloaters(gdt);
  updateBolts(gdt);

  if (!player.alive) {
    dieT -= dt;
    if (dieT <= 0) {
      best = Math.max(best, score); saveBest();
      state = 'over';
    }
  }
}

function update(dt, gdt) {
  updateStars(dt);
  if (pause) return;
  switch (state) {
    case 'menu':
      if (Math.random() < dt * 0.4) strike(rand(80, W - 80), -10, rand(80, W - 80), rand(H * 0.2, H * 0.5), '#bfe8ff', 2);
      updateParticles(dt); updateFloaters(dt); updateBolts(dt);
      break;
    case 'intro':
      introT += dt;
      updateParticles(dt); updateBolts(dt);
      if (introT > 5.5) state = 'play';
      break;
    case 'play':
      updatePlay(dt, gdt);
      break;
    case 'clear':
      clearT += dt;
      updateDrops(gdt); updateParticles(gdt); updateFloaters(gdt); updateBolts(gdt);
      if (clearT > 3) nextStage();
      break;
    case 'over':
    case 'win':
      winT += dt;
      updateParticles(dt); updateFloaters(dt); updateBolts(dt);
      break;
  }
}

/* ---------------- 绘制 ---------------- */
function text(str, x, y, size, color = '#fff', align = 'center') {
  ctx.font = 'bold ' + size + 'px ' + FONT;
  ctx.fillStyle = color;
  ctx.textAlign = align;
  ctx.textBaseline = 'middle';
  ctx.fillText(str, x, y);
}

function drawBackground() {
  const st = state === 'menu' ? MENU_BG : STAGES[stageIdx];
  const g = ctx.createLinearGradient(0, 0, 0, H);
  g.addColorStop(0, st.bgTop);
  g.addColorStop(1, st.bgBottom);
  ctx.fillStyle = g;
  ctx.fillRect(0, 0, W, H);
  for (const s of stars) {
    ctx.globalAlpha = 0.25 + s.z * 0.25;
    ctx.fillStyle = s.z === 3 ? '#cfe8ff' : '#8fb4e8';
    const sz = s.z * 1.1;
    ctx.fillRect(s.x, s.y, sz, sz + (state === 'play' && !pause ? 6 : 0));
  }
  ctx.globalAlpha = 1;
  if (st.clouds) {
    for (const cl of clouds) {
      ctx.save();
      ctx.translate(cl.x, cl.y);
      ctx.scale(1, cl.h / cl.w);
      const cg = ctx.createRadialGradient(0, 0, 1, 0, 0, cl.w / 2);
      cg.addColorStop(0, 'rgba(140,150,200,0.12)');
      cg.addColorStop(1, 'rgba(140,150,200,0)');
      ctx.fillStyle = cg;
      ctx.beginPath(); ctx.arc(0, 0, cl.w / 2, 0, TAU); ctx.fill();
      ctx.restore();
    }
  }
}

function drawPlayer() {
  if (!player.alive) return;
  const c = ctx;
  if (player.shield > 0) {
    const a = player.shield < 2 ? (Math.sin(performance.now() / 50) > 0 ? 0.7 : 0.3) : 0.5;
    c.strokeStyle = `rgba(127,227,255,${a})`;
    c.lineWidth = 3;
    c.shadowColor = '#7fe3ff'; c.shadowBlur = 14;
    c.beginPath(); c.arc(player.x, player.y, 26, 0, TAU); c.stroke();
    c.shadowBlur = 0;
  }
  if (player.invuln > 0 && Math.floor(player.invuln * 14) % 2 === 0) return; // 受击闪烁
  c.save();
  c.translate(player.x, player.y);
  c.rotate(player.tilt);
  const eg = c.createRadialGradient(0, 18, 1, 0, 18, 14);
  eg.addColorStop(0, 'rgba(127,227,255,0.9)');
  eg.addColorStop(1, 'rgba(127,227,255,0)');
  c.fillStyle = eg;
  c.beginPath(); c.arc(0, 18, 14, 0, TAU); c.fill();
  const bg = c.createLinearGradient(0, -26, 0, 22);
  bg.addColorStop(0, '#eaffff');
  bg.addColorStop(0.35, '#7fe3ff');
  bg.addColorStop(1, '#1f6fb8');
  c.fillStyle = bg;
  c.strokeStyle = '#bfefff'; c.lineWidth = 1.5;
  c.shadowColor = '#7fe3ff'; c.shadowBlur = 12;
  c.beginPath();
  c.moveTo(0, -26);
  c.lineTo(5, -12); c.lineTo(24, 8); c.lineTo(26, 13); c.lineTo(9, 10);
  c.lineTo(7, 18); c.lineTo(12, 23); c.lineTo(0, 19);
  c.lineTo(-12, 23); c.lineTo(-7, 18); c.lineTo(-9, 10);
  c.lineTo(-26, 13); c.lineTo(-24, 8); c.lineTo(-5, -12);
  c.closePath();
  c.fill(); c.stroke();
  c.shadowBlur = 0;
  c.fillStyle = 'rgba(20,40,70,0.9)';
  c.beginPath(); c.ellipse(0, -8, 3.5, 7, 0, 0, TAU); c.fill();
  c.strokeStyle = '#bfefff'; c.lineWidth = 1; c.stroke();
  c.restore();
}

function drawLaser() {
  if (!player.alive || player.weapon !== 'laser') return;
  const lvl = player.levels.laser;
  const flick = 0.9 + Math.sin(performance.now() / 25) * 0.1;
  const beams = [[0, (7 + lvl * 3) * flick]];
  if (lvl >= 4) beams.push([-26, 5 * flick], [26, 5 * flick]);
  for (const [ox, hw] of beams) {
    const x = player.x + ox;
    const yTop = player.y - 26;
    ctx.fillStyle = 'rgba(80,220,255,0.22)';
    ctx.fillRect(x - hw, 0, hw * 2, yTop);
    ctx.fillStyle = 'rgba(160,240,255,0.7)';
    ctx.fillRect(x - hw * 0.45, 0, hw * 0.9, yTop);
    ctx.fillStyle = '#fff';
    ctx.fillRect(x - 2, 0, 4, yTop);
  }
}

function drawPbullets() {
  const c = ctx;
  for (const b of pbullets) {
    if (b.type === 'missile') {
      const a = Math.atan2(b.vy, b.vx);
      c.save(); c.translate(b.x, b.y); c.rotate(a);
      c.fillStyle = '#ffd27f';
      c.shadowColor = '#ff9a3d'; c.shadowBlur = 10;
      c.beginPath(); c.moveTo(9, 0); c.lineTo(-6, 4.5); c.lineTo(-3, 0); c.lineTo(-6, -4.5);
      c.closePath(); c.fill();
      c.restore();
    } else {
      const a = Math.atan2(b.vy, b.vx);
      c.save(); c.translate(b.x, b.y); c.rotate(a);
      c.strokeStyle = b.color;
      c.lineWidth = b.type === 'pellet' ? 3 : 4.5;
      c.shadowColor = b.color; c.shadowBlur = 8;
      c.beginPath(); c.moveTo(-10, 0); c.lineTo(6, 0); c.stroke();
      c.fillStyle = '#fff';
      c.beginPath(); c.arc(6, 0, 2.2, 0, TAU); c.fill();
      c.restore();
    }
  }
  c.shadowBlur = 0;
}

function drawDrops() {
  for (const d of drops) {
    const [icon, color] = DROP_INFO[d.type];
    const bob = Math.sin(d.t * 3) * 3;
    ctx.save();
    ctx.translate(d.x, d.y + bob);
    ctx.shadowColor = color; ctx.shadowBlur = 12;
    ctx.fillStyle = 'rgba(8,16,34,0.9)';
    rr(ctx, -12, -12, 24, 24, 6); ctx.fill();
    ctx.strokeStyle = color; ctx.lineWidth = 2;
    rr(ctx, -12, -12, 24, 24, 6); ctx.stroke();
    ctx.shadowBlur = 0;
    text(icon, 0, 1, 13, color);
    ctx.restore();
  }
}

function drawHUD() {
  const c = ctx;
  const st = STAGES[stageIdx];
  c.fillStyle = 'rgba(4,8,20,0.6)';
  c.fillRect(0, 0, W, 44);
  // HP
  const pct = clamp(player.hp / player.maxHp, 0, 1);
  c.fillStyle = '#12203a';
  rr(c, 12, 13, 170, 18, 5); c.fill();
  if (pct > 0) {
    c.fillStyle = pct > 0.5 ? '#4dff8c' : pct > 0.25 ? '#ffd24d' : '#ff4d5e';
    rr(c, 14, 15, 166 * pct, 14, 4); c.fill();
  }
  text(`HP ${Math.max(0, Math.ceil(player.hp))}`, 20, 23, 12, '#08131f', 'left');
  // 武器
  const wi = WEAPONS.findIndex(w => w.id === player.weapon);
  c.fillStyle = 'rgba(10,20,44,0.8)';
  rr(c, 194, 7, 152, 30, 6); c.fill();
  c.strokeStyle = '#2a4a7a'; c.lineWidth = 1;
  rr(c, 194, 7, 152, 30, 6); c.stroke();
  text(WEAPONS[wi].name, 202, 16, 12, '#cfe8ff', 'left');
  const lvl = player.levels[player.weapon];
  for (let i = 0; i < 5; i++) {
    c.fillStyle = i < lvl ? '#ffd24d' : '#22345a';
    c.fillRect(202 + i * 18, 26, 14, 5);
  }
  // 炸弹
  text('炸弹', 358, 16, 12, '#cfe8ff', 'left');
  for (let i = 0; i < player.bombs; i++) {
    c.fillStyle = '#ff9a5e';
    c.shadowColor = '#ff9a5e'; c.shadowBlur = 6;
    c.beginPath(); c.arc(362 + i * 15, 31, 5, 0, TAU); c.fill();
  }
  c.shadowBlur = 0;
  // 关卡名 / 分数
  text(`第${st.id}幕 · ${st.name}`, W / 2, 22, 15, '#9fd8ff');
  text('SCORE ' + String(score).padStart(7, '0'), W - 14, 22, 16, '#ffe29a', 'right');
  // Boss 血条
  if (boss && !boss.dead && boss.y > 0) {
    const bw = 380, bx = W / 2 - bw / 2, by = 54;
    text(boss.name + (boss.enrage ? '（狂暴）' : ''), W / 2, by - 9, 13, '#ff9db4');
    c.fillStyle = 'rgba(20,8,16,0.7)';
    rr(c, bx - 4, by - 4, bw + 8, 14, 6); c.fill();
    const bp = clamp(boss.hp / boss.maxHp, 0, 1);
    if (bp > 0) {
      c.fillStyle = '#ff3d5e';
      rr(c, bx, by, bw * bp, 6, 3); c.fill();
    }
  }
}

function drawMenu() {
  const t = performance.now() / 1000;
  ctx.save();
  ctx.shadowColor = '#4da6ff'; ctx.shadowBlur = 30;
  text('雷 霆 突 击', W / 2, H * 0.22, 64, '#dff2ff');
  ctx.shadowBlur = 0;
  text('T H U N D E R   S T R I K E', W / 2, H * 0.22 + 54, 18, '#5f8fd9');
  text('敌机入侵 · 雷暴封锁 · 空中堡垒 · 终极决战', W / 2, H * 0.42, 17, '#9fb6d9');
  text(`武器系统：${WEAPONS.map(w => w.name).join(' / ')}`, W / 2, H * 0.51, 15, '#7fe3ff');
  text('WASD / 方向键 移动      1-4 切换武器      空格 炸弹      P 暂停      M 静音', W / 2, H * 0.59, 15, '#9fb6d9');
  text('击落敌机可掉落补给：生命 / 武器强化 / 护盾 / 炸弹 / 分数', W / 2, H * 0.65, 14, '#7f96b9');
  const a = 0.55 + Math.sin(t * 3) * 0.4;
  text('按 ENTER 开始任务', W / 2, H * 0.79, 26, `rgba(140,230,255,${a})`);
  if (best > 0) text(`最高分 ${best}`, W / 2, H * 0.88, 15, '#ffd24d');
  ctx.restore();
}

function drawIntro() {
  const st = STAGES[stageIdx];
  ctx.fillStyle = 'rgba(3,6,16,0.85)';
  ctx.fillRect(0, 0, W, H);
  text(`第 ${st.id} 幕`, W / 2, H * 0.30, 24, '#7fe3ff');
  ctx.save();
  ctx.shadowColor = '#4da6ff'; ctx.shadowBlur = 24;
  text(st.name, W / 2, H * 0.38, 46, '#eaf6ff');
  ctx.restore();
  st.story.forEach((line, i) => {
    const a = clamp(introT - 0.5 - i * 0.9, 0, 1);
    if (a > 0) text(line, W / 2, H * 0.52 + i * 36, 18, `rgba(190,215,245,${a})`);
  });
  const a = 0.4 + Math.sin(performance.now() / 300) * 0.4;
  text('按 ENTER 立即出击', W / 2, H * 0.85, 16, `rgba(140,230,255,${a})`);
}

function drawClear() {
  const st = STAGES[stageIdx];
  ctx.fillStyle = 'rgba(3,6,16,0.7)';
  ctx.fillRect(0, 0, W, H);
  ctx.save();
  ctx.shadowColor = '#4da6ff'; ctx.shadowBlur = 26;
  text('关卡完成', W / 2, H * 0.34, 48, '#7fe3ff');
  ctx.restore();
  text(`第 ${st.id} 幕 · ${st.name}`, W / 2, H * 0.45, 18, '#9fb6d9');
  text(`当前得分 ${score}`, W / 2, H * 0.53, 20, '#ffe29a');
  if (stageIdx < STAGES.length - 1) {
    const nx = STAGES[stageIdx + 1];
    text(`下一幕：${nx.name} —— ${nx.story[0]}`, W / 2, H * 0.62, 15, '#7f96b9');
  } else {
    text('最终决战即将落幕……', W / 2, H * 0.62, 15, '#7f96b9');
  }
  const a = 0.4 + Math.sin(performance.now() / 300) * 0.4;
  text('按 ENTER 继续', W / 2, H * 0.8, 18, `rgba(140,230,255,${a})`);
}

function drawOver() {
  ctx.fillStyle = 'rgba(10,2,6,0.8)';
  ctx.fillRect(0, 0, W, H);
  ctx.save();
  ctx.shadowColor = '#ff3d5e'; ctx.shadowBlur = 28;
  text('任务失败', W / 2, H * 0.34, 52, '#ff5e6e');
  ctx.restore();
  text('MISSION FAILED', W / 2, H * 0.34 + 48, 16, '#a05560');
  text(`得分 ${score}`, W / 2, H * 0.52, 22, '#ffe29a');
  if (score >= best && score > 0) {
    const a = 0.5 + Math.sin(performance.now() / 200) * 0.4;
    text('★ 新纪录 ★', W / 2, H * 0.60, 18, `rgba(255,210,77,${a})`);
  }
  const a = 0.4 + Math.sin(performance.now() / 300) * 0.4;
  text('按 ENTER 重新开始', W / 2, H * 0.78, 18, `rgba(140,230,255,${a})`);
}

function drawWin() {
  ctx.fillStyle = 'rgba(6,4,12,0.82)';
  ctx.fillRect(0, 0, W, H);
  ctx.save();
  ctx.shadowColor = '#ffd24d'; ctx.shadowBlur = 30;
  text('胜 利', W / 2, H * 0.18, 56, '#ffd24d');
  ctx.restore();
  text('VICTORY', W / 2, H * 0.18 + 46, 16, '#a8894a');
  WIN_LINES.forEach((line, i) => {
    const a = clamp(winT - 0.6 - i * 1.0, 0, 1);
    if (a > 0) text(line, W / 2, H * 0.38 + i * 34, 17, `rgba(220,235,255,${a})`);
  });
  text(`最终得分 ${score}`, W / 2, H * 0.68, 22, '#ffe29a');
  if (score >= best && score > 0) text('★ 新纪录 ★', W / 2, H * 0.75, 18, '#ffd24d');
  const a = 0.4 + Math.sin(performance.now() / 300) * 0.4;
  text('按 ENTER 返回主菜单', W / 2, H * 0.88, 18, `rgba(140,230,255,${a})`);
}

function drawPause() {
  ctx.fillStyle = 'rgba(3,6,16,0.6)';
  ctx.fillRect(0, 0, W, H);
  text('已暂停', W / 2, H * 0.45, 40, '#cfe8ff');
  text('按 P 继续', W / 2, H * 0.45 + 46, 16, '#7f96b9');
}

function render() {
  drawBackground();
  ctx.save();
  if (shakeAmt > 0) ctx.translate(rand(-shakeAmt, shakeAmt), rand(-shakeAmt, shakeAmt));
  if (state === 'menu') {
    drawParticles(ctx);
    drawBolts(ctx);
    drawMenu();
  } else {
    drawDrops();
    drawEbullets(ctx);
    drawEnemies(ctx);
    drawBossExtras(ctx);
    drawBoss(ctx);
    drawPbullets();
    drawLaser();
    drawPlayer();
    drawParticles(ctx);
    drawBolts(ctx);
    drawFloaters(ctx);
    if (state === 'play') {
      if (bossPending) {
        const a = Math.sin(performance.now() / 90) > 0 ? 0.9 : 0.3;
        text('!! 警告 WARNING !!', W / 2, H * 0.4, 40, `rgba(255,70,90,${a})`);
      }
      drawHUD();
    }
    if (state === 'intro') drawIntro();
    if (state === 'clear') drawClear();
    if (state === 'over') drawOver();
    if (state === 'win') drawWin();
  }
  ctx.restore();
  // 白闪
  if (flashScreen > 0) {
    ctx.fillStyle = `rgba(235,245,255,${clamp(flashScreen, 0, 1) * 0.8})`;
    ctx.fillRect(0, 0, W, H);
  }
  if (mutedMsgT > 0) text(SFX.muted ? '已静音' : '声音开启', 14, H - 20, 13, '#7f96b9', 'left');
  if (pause && state === 'play') drawPause();
}

/* ---------------- 输入 ---------------- */
window.addEventListener('keydown', (e) => {
  SFX.init(); SFX.resume();
  const k = e.key.toLowerCase();
  if (['arrowup', 'arrowdown', 'arrowleft', 'arrowright', ' '].includes(k)) e.preventDefault();
  keys[k] = true;
  if (e.repeat) return;
  if (k === 'm') {
    SFX.muted = !SFX.muted;
    mutedMsgT = 1.5;
    return;
  }
  switch (state) {
    case 'menu':
      if (k === 'enter') startGame();
      break;
    case 'intro':
      if (k === 'enter' || k === ' ') state = 'play';
      break;
    case 'play':
      if (k === 'p' || k === 'escape') pause = !pause;
      else if (!pause) {
        if (k === '1' || k === '2' || k === '3' || k === '4') switchWeapon(+k - 1);
        else if (k === ' ') useBomb();
      }
      break;
    case 'clear':
      if (k === 'enter' || k === ' ') nextStage();
      break;
    case 'over':
      if (k === 'enter' || k === ' ') startGame();
      break;
    case 'win':
      if (k === 'enter') state = 'menu';
      break;
  }
});
window.addEventListener('keyup', (e) => { keys[e.key.toLowerCase()] = false; });

/* ---------------- 主循环 ---------------- */
let lastT = performance.now();
function frame(now) {
  requestAnimationFrame(frame);
  const dt = Math.min(0.033, (now - lastT) / 1000);
  lastT = now;
  if (slowT > 0) { slowT -= dt; timeScale = 0.35; } else timeScale = 1;
  const gdt = dt * timeScale;
  shakeAmt = Math.max(0, shakeAmt - dt * 45);
  flashScreen = Math.max(0, flashScreen - dt * 2.5);
  mutedMsgT = Math.max(0, mutedMsgT - dt);
  update(dt, gdt);
  render();
}
requestAnimationFrame(frame);
