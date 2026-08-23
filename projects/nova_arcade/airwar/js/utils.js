'use strict';
/* =========================================================
 * 雷霆突袭 THUNDER STRIKE —— 工具 / 粒子 / 闪电 / 音效
 * ========================================================= */
const TAU = Math.PI * 2;
const FONT = '"Microsoft YaHei","PingFang SC","Noto Sans SC",sans-serif';

const rand = (a, b) => a + Math.random() * (b - a);
const randInt = (a, b) => Math.floor(rand(a, b + 1));
const clamp = (v, a, b) => (v < a ? a : v > b ? b : v);
const lerp = (a, b, t) => a + (b - a) * t;
const dist = (ax, ay, bx, by) => Math.hypot(ax - bx, ay - by);
const pick = arr => (arr.length ? arr[Math.floor(Math.random() * arr.length)] : undefined);
function angDiff(a, b) { return ((a - b + Math.PI * 3) % TAU) - Math.PI; }

/* ---- 圆角矩形路径 ---- */
function rr(c, x, y, w, h, r) {
  c.beginPath();
  c.moveTo(x + r, y);
  c.arcTo(x + w, y, x + w, y + h, r);
  c.arcTo(x + w, y + h, x, y + h, r);
  c.arcTo(x, y + h, x, y, r);
  c.arcTo(x, y, x + w, y, r);
  c.closePath();
}

/* ---- 全局特效状态（各模块共享） ---- */
let shakeAmt = 0;
function addShake(v) { shakeAmt = Math.min(28, shakeAmt + v); }
let flashScreen = 0; // 屏幕白闪 0~1

/* ================= 音效（WebAudio 实时合成，无外部资源） ================= */
const SFX = {
  ctx: null, master: null, muted: false,
  init() {
    if (this.ctx) return;
    try {
      const AC = window.AudioContext || window.webkitAudioContext;
      this.ctx = new AC();
      this.master = this.ctx.createGain();
      this.master.gain.value = 0.32;
      this.master.connect(this.ctx.destination);
    } catch (e) { this.ctx = null; }
  },
  resume() { if (this.ctx && this.ctx.state === 'suspended') this.ctx.resume(); },
  _env(g, t0, vol, dur) {
    g.gain.setValueAtTime(0.0001, t0);
    g.gain.exponentialRampToValueAtTime(Math.max(0.001, vol), t0 + 0.012);
    g.gain.exponentialRampToValueAtTime(0.0001, t0 + dur);
  },
  tone(freq, dur = 0.15, type = 'square', vol = 0.2, slideTo = null) {
    if (!this.ctx || this.muted) return;
    try {
      const t0 = this.ctx.currentTime;
      const o = this.ctx.createOscillator();
      const g = this.ctx.createGain();
      o.type = type;
      o.frequency.setValueAtTime(freq, t0);
      if (slideTo) o.frequency.exponentialRampToValueAtTime(Math.max(20, slideTo), t0 + dur);
      this._env(g, t0, vol, dur);
      o.connect(g); g.connect(this.master);
      o.start(t0); o.stop(t0 + dur + 0.05);
    } catch (e) { /* 忽略音频错误 */ }
  },
  noise(dur = 0.3, vol = 0.3, freq = 800) {
    if (!this.ctx || this.muted) return;
    try {
      const t0 = this.ctx.currentTime;
      const len = Math.max(1, (this.ctx.sampleRate * dur) | 0);
      const buf = this.ctx.createBuffer(1, len, this.ctx.sampleRate);
      const d = buf.getChannelData(0);
      for (let i = 0; i < len; i++) d[i] = (Math.random() * 2 - 1) * (1 - i / len);
      const src = this.ctx.createBufferSource(); src.buffer = buf;
      const f = this.ctx.createBiquadFilter(); f.type = 'lowpass'; f.frequency.value = freq;
      const g = this.ctx.createGain(); this._env(g, t0, vol, dur);
      src.connect(f); f.connect(g); g.connect(this.master);
      src.start(t0);
    } catch (e) { /* 忽略音频错误 */ }
  },
  shoot(w) {
    if (w === 'normal') this.tone(880, 0.06, 'square', 0.05, 420);
    else if (w === 'spread') this.tone(520, 0.08, 'sawtooth', 0.06, 280);
    else if (w === 'homing') this.tone(300, 0.12, 'triangle', 0.07, 620);
  },
  boom(size = 1) { this.noise(0.25 * size, 0.5, 900); this.tone(120, 0.3 * size, 'sine', 0.4, 40); },
  bigBoom() { this.noise(0.8, 0.7, 500); this.tone(80, 0.9, 'sine', 0.6, 30); this.tone(160, 0.5, 'triangle', 0.3, 40); },
  zap() { this.noise(0.2, 0.4, 4000); this.tone(1400, 0.12, 'sawtooth', 0.12, 200); },
  pickup() { this.tone(660, 0.08, 'square', 0.15); setTimeout(() => this.tone(990, 0.1, 'square', 0.15), 70); },
  bomb() { this.noise(1.0, 0.8, 400); this.tone(60, 1.2, 'sine', 0.7, 25); },
  hit() { this.tone(200, 0.12, 'sawtooth', 0.2, 80); this.noise(0.1, 0.2, 1200); },
  warn() { this.tone(440, 0.25, 'square', 0.15); setTimeout(() => this.tone(440, 0.25, 'square', 0.15), 350); },
};

/* ================= 粒子系统 ================= */
const particles = [];
function addParticle(p) { if (particles.length < 900) particles.push(p); }

function explosion(x, y, size = 1, palette = ['#ffe29a', '#ffb15e', '#ff6b4a']) {
  // 中心闪光
  addParticle({ type: 'flash', x, y, vx: 0, vy: 0, life: 0.18, max: 0.18, size: 46 * size, color: '#fff' });
  // 冲击波环
  addParticle({ type: 'ring', x, y, vx: 0, vy: 0, life: 0.5, max: 0.5, size: 12, grow: 280 * size, color: 'rgba(255,210,140,' });
  // 火花
  const n = Math.min(42, (16 * size) | 0);
  for (let i = 0; i < n; i++) {
    const a = rand(0, TAU), sp = rand(60, 340) * size;
    addParticle({ type: 'spark', x, y, vx: Math.cos(a) * sp, vy: Math.sin(a) * sp, life: rand(0.25, 0.6), max: 0.6, size: rand(1.5, 3.5) * Math.min(size, 1.6), color: pick(palette) });
  }
  // 金属碎片
  const m = Math.min(14, (5 * size) | 0);
  for (let i = 0; i < m; i++) {
    const a = rand(0, TAU), sp = rand(40, 200) * size;
    addParticle({ type: 'debris', x, y, vx: Math.cos(a) * sp, vy: Math.sin(a) * sp - 60, life: rand(0.5, 1.1), max: 1.1, size: rand(3, 7), rot: rand(0, TAU), vr: rand(-8, 8), color: pick(['#8a93a6', '#5c6577', palette[2]]) });
  }
  // 烟雾
  for (let i = 0; i < 6; i++) {
    addParticle({ type: 'smoke', x: x + rand(-8, 8), y: y + rand(-8, 8), vx: rand(-30, 30), vy: rand(-50, -10), life: rand(0.6, 1.4), max: 1.4, size: rand(6, 14) * Math.min(size, 2), color: '#9aa3b5' });
  }
}

function sparkBurst(x, y, n = 6, color = '#ffe29a') {
  for (let i = 0; i < n; i++) {
    const a = rand(0, TAU), sp = rand(40, 180);
    addParticle({ type: 'spark', x, y, vx: Math.cos(a) * sp, vy: Math.sin(a) * sp, life: rand(0.15, 0.4), max: 0.4, size: rand(1.2, 2.6), color });
  }
}

function updateParticles(dt) {
  for (let i = particles.length - 1; i >= 0; i--) {
    const p = particles[i];
    p.life -= dt;
    if (p.life <= 0) { particles.splice(i, 1); continue; }
    p.x += p.vx * dt; p.y += p.vy * dt;
    if (p.type === 'spark') {
      const drag = Math.pow(0.02, dt);
      p.vx *= drag; p.vy = p.vy * drag + 160 * dt;
    } else if (p.type === 'debris') {
      p.vy += 300 * dt; p.rot += p.vr * dt;
    } else if (p.type === 'smoke') {
      p.size += 14 * dt; p.vy -= 20 * dt;
    }
  }
}

function drawParticles(c) {
  for (const p of particles) {
    const a = clamp(p.life / p.max, 0, 1);
    if (p.type === 'flash') {
      const r = p.size * (1.6 - a * 0.6);
      const g = c.createRadialGradient(p.x, p.y, 0, p.x, p.y, r);
      g.addColorStop(0, `rgba(255,255,255,${a})`);
      g.addColorStop(0.4, `rgba(255,210,120,${a * 0.6})`);
      g.addColorStop(1, 'rgba(255,120,60,0)');
      c.fillStyle = g;
      c.fillRect(p.x - r, p.y - r, r * 2, r * 2);
    } else if (p.type === 'ring') {
      const r = lerp(p.size, p.size + p.grow, 1 - a);
      c.strokeStyle = p.color + (a * 0.9) + ')';
      c.lineWidth = 3 + a * 5;
      c.beginPath(); c.arc(p.x, p.y, r, 0, TAU); c.stroke();
    } else if (p.type === 'spark') {
      c.globalAlpha = a;
      c.strokeStyle = p.color; c.lineWidth = p.size;
      c.beginPath(); c.moveTo(p.x, p.y); c.lineTo(p.x - p.vx * 0.035, p.y - p.vy * 0.035); c.stroke();
      c.globalAlpha = 1;
    } else if (p.type === 'debris') {
      c.save(); c.translate(p.x, p.y); c.rotate(p.rot || 0);
      c.globalAlpha = a;
      c.fillStyle = p.color;
      c.fillRect(-p.size / 2, -p.size / 3, p.size, p.size * 0.66);
      c.restore(); c.globalAlpha = 1;
    } else if (p.type === 'smoke') {
      c.globalAlpha = a * 0.25;
      c.fillStyle = p.color;
      c.beginPath(); c.arc(p.x, p.y, p.size, 0, TAU); c.fill();
      c.globalAlpha = 1;
    }
  }
}

/* ================= 漂浮文字 ================= */
const floaters = [];
function addFloater(x, y, text, color = '#fff', size = 15) {
  floaters.push({ x, y, text, color, size, life: 1 });
}
function updateFloaters(dt) {
  for (let i = floaters.length - 1; i >= 0; i--) {
    const f = floaters[i];
    f.y -= 36 * dt; f.life -= dt * 0.8;
    if (f.life <= 0) floaters.splice(i, 1);
  }
}
function drawFloaters(c) {
  for (const f of floaters) {
    c.globalAlpha = clamp(f.life, 0, 1);
    c.font = 'bold ' + f.size + 'px ' + FONT;
    c.textAlign = 'center'; c.textBaseline = 'middle';
    c.fillStyle = f.color;
    c.fillText(f.text, f.x, f.y);
  }
  c.globalAlpha = 1;
}

/* ================= 闪电 ================= */
const bolts = [];
function boltPoints(x1, y1, x2, y2) {
  let pts = [[x1, y1], [x2, y2]];
  for (let depth = 0; depth < 5; depth++) {
    const next = [pts[0]];
    for (let i = 0; i < pts.length - 1; i++) {
      const ax = pts[i][0], ay = pts[i][1], bx = pts[i + 1][0], by = pts[i + 1][1];
      const mx = (ax + bx) / 2, my = (ay + by) / 2;
      const dx = bx - ax, dy = by - ay;
      const len = Math.hypot(dx, dy) || 1;
      const off = rand(-0.5, 0.5) * len * 0.45;
      next.push([mx + (-dy / len) * off, my + (dx / len) * off], [bx, by]);
    }
    pts = next;
  }
  return pts;
}
function strike(x1, y1, x2, y2, color = '#bfe8ff', w = 3) {
  bolts.push({ pts: boltPoints(x1, y1, x2, y2), life: 0.22, max: 0.22, color, w });
}
function updateBolts(dt) {
  for (let i = bolts.length - 1; i >= 0; i--) {
    bolts[i].life -= dt;
    if (bolts[i].life <= 0) bolts.splice(i, 1);
  }
}
function drawBolts(c) {
  for (const b of bolts) {
    const a = clamp(b.life / b.max, 0, 1);
    c.save();
    c.globalAlpha = a * 0.4;
    c.strokeStyle = b.color;
    c.shadowColor = b.color; c.shadowBlur = 18;
    c.lineJoin = 'round'; c.lineWidth = b.w + 4;
    c.beginPath();
    for (let i = 0; i < b.pts.length; i++) { const p = b.pts[i]; i ? c.lineTo(p[0], p[1]) : c.moveTo(p[0], p[1]); }
    c.stroke();
    c.globalAlpha = a;
    c.strokeStyle = '#ffffff'; c.lineWidth = Math.max(1, b.w - 1);
    c.beginPath();
    for (let i = 0; i < b.pts.length; i++) { const p = b.pts[i]; i ? c.lineTo(p[0], p[1]) : c.moveTo(p[0], p[1]); }
    c.stroke();
    c.restore();
  }
}
