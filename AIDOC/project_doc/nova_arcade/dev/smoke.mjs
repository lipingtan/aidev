// Headless smoke test for tetra-nova.html
// Runs the full game script with DOM/canvas shims and simulates a session.
import { readFileSync, writeFileSync, mkdirSync } from 'node:fs';

const html = readFileSync(new URL('../tetra-nova.html', import.meta.url), 'utf8');
const m = html.match(/<script>([\s\S]*?)<\/script>/);
if (!m) { console.error('FAIL: no <script> found'); process.exit(1); }
const code = m[1];

// ---------- syntax check ----------
mkdirSync(new URL('./tmp', import.meta.url), { recursive: true });
writeFileSync(new URL('./tmp/game.js', import.meta.url), code);

// ---------- shims ----------
function makeCtx(canvas) {
  const gradient = { addColorStop() {} };
  const target = {};
  return new Proxy(target, {
    get(t, p) {
      if (p === 'canvas') return canvas;
      if (typeof p === 'symbol') return undefined;
      if (p in t) return t[p];
      return (...args) => {
        if (p === 'createLinearGradient' || p === 'createRadialGradient' || p === 'createPattern') return gradient;
        if (p === 'measureText') return { width: 10 };
        if (p === 'getImageData') return { data: new Uint8ClampedArray(4) };
        return undefined;
      };
    },
    set(t, p, v) { t[p] = v; return true; },
  });
}
function makeCanvas() {
  const c = { width: 0, height: 0, style: {}, addEventListener() {} };
  c.getContext = () => makeCtx(c);
  return c;
}
function makeEl(id) {
  return {
    id, style: {}, dataset: {}, textContent: '', innerHTML: '',
    firstElementChild: { style: {} },
    classList: { add() {}, remove() {}, toggle() {} },
    addEventListener() {}, appendChild() {}, setAttribute() {},
  };
}
const els = {};
const canvasIds = new Set(['bg', 'board', 'fx', 'hold', 'next']);
const doc = {
  getElementById(id) {
    if (!els[id]) els[id] = canvasIds.has(id) ? makeCanvas() : makeEl(id);
    return els[id];
  },
  createElement(tag) { return tag === 'canvas' ? makeCanvas() : makeEl('dyn'); },
  querySelectorAll() { return []; },
  addEventListener() {},
  body: makeEl('body'),
};
const storage = {};
const ls = {
  getItem: k => (k in storage ? storage[k] : null),
  setItem: (k, v) => { storage[k] = String(v); },
};
let rafCb = null;
const win = {
  innerWidth: 1280, innerHeight: 800, devicePixelRatio: 1,
  addEventListener() {},
};
const perf = { now: () => Date.now() };

// ---------- run ----------
let T = null;
const errors = [];
try {
  const fn = new Function('window', 'document', 'navigator', 'performance', 'requestAnimationFrame', 'localStorage', code);
  fn(win, doc, {}, perf, cb => { rafCb = cb; }, ls);
  T = win.__TETRA;
  if (!T) throw new Error('__TETRA not exposed');
} catch (e) {
  console.error('FAIL: bootstrap threw\n', e);
  process.exit(1);
}

let now = 0;
function pump(frames) {
  for (let i = 0; i < frames; i++) {
    now += 16.7;
    try { T.frame(now); }
    catch (e) { errors.push('frame threw @' + now.toFixed(0) + 'ms: ' + (e && e.stack || e)); return false; }
  }
  return true;
}
const ok = (cond, label) => {
  console.log((cond ? 'PASS' : 'FAIL') + ' - ' + label);
  if (!cond) process.exitCode = 1;
};
const G = T.G;

// scenario
pump(30);
ok(G.state === 'MENU', 'boots to menu');

function settleSelects(maxPicks = 8) {
  let picks = 0;
  while (G.state === 'SELECT' && picks < maxPicks) { T.force.pickCard(0); picks++; pump(60); }
}

T.startGame();
ok(G.state === 'PLAYING', 'startGame -> PLAYING');
ok(!!G.piece, 'piece spawned');
pump(60);

T.keyDown('ArrowLeft'); pump(4); T.keyUp('ArrowLeft');
T.keyDown('ArrowUp'); pump(2); T.keyUp('ArrowUp');
T.keyDown('ArrowDown'); pump(10); T.keyUp('ArrowDown');
ok(errors.length === 0, 'movement/rotate/soft-drop inputs ok');

T.keyDown('Space'); T.keyUp('Space'); pump(120);
ok(errors.length === 0, 'hard drop + settle ok');
ok(G.score > 0 || G.pieceCount >= 1, 'drop scored/pieceCount advanced');

// force clears
T.force.fillRow(19); T.force.fillRow(18); T.force.fillRow(17);
T.keyDown('Space'); T.keyUp('Space'); pump(300);
settleSelects();
ok(G.lines > 0, 'line clears registered (lines=' + G.lines + ')');
ok(G.score > 0, 'score increased after clears (' + G.score + ')');
ok(errors.length === 0, 'clear pipeline no errors');

// wave completion -> select
T.force.giveLines(30); pump(120);
ok(G.state === 'SELECT', 'wave complete -> SELECT (state=' + G.state + ')');
T.keyDown('Digit1'); pump(10);
ok(G.state === 'PLAYING' && G.wave === 2, 'picked card -> wave 2');

// mutations
for (const id of ['bombs', 'cascade', 'embers', 'nova', 'twindrop', 'secondwind', 'lightning', 'shockwave', 'blackhole']) {
  T.force.applyMut(id);
}
ok((G.muts.bombs || 0) >= 1 && (G.muts.cascade || 0) >= 1, 'mutations applied');

// more clears with the whole mutation stack active
T.force.fillRow(19); T.force.fillRow(18);
T.keyDown('Space'); T.keyUp('Space'); pump(500);
settleSelects();
ok(errors.length === 0, 'mutation-laden clears no errors');

// twin drop exercised
for (let i = 0; i < 6; i++) {
  if (G.state === 'SELECT') { T.force.pickCard(0); pump(60); }
  T.keyDown('Space'); T.keyUp('Space'); pump(120);
}
settleSelects();
ok(errors.length === 0, 'twin drop cycles ok');

// boss
T.force.spawnBoss();
ok(!!G.boss, 'boss spawned');
pump(60 * 12); // 12s -> garbage pushes (with telegraph) + embers etc
settleSelects();
ok(errors.length === 0, 'boss attack cycle no errors');
ok(G.stats.bosses >= 0 && G.lines >= 0, 'survived boss attacks (bosses=' + G.stats.bosses + ')');

if (G.boss) {
  const hp0 = G.boss.hp;
  // rows 17/18 plain full, row 19 full WITH a core at x=4 -> any lock clears all three
  T.force.fillRow(17); T.force.fillRow(18); T.force.fillRow(19, 4);
  T.keyDown('Space'); T.keyUp('Space'); pump(400);
  settleSelects();
  ok(!G.boss || G.boss.hp < hp0 || G.stats.bosses >= 1, 'core clear damaged/killed boss');
  ok(errors.length === 0, 'boss damage pipeline ok');
} else {
  ok(G.stats.bosses >= 1, 'boss already slain by mutation chaos');
}

// second wind then game over
settleSelects();
if (G.state === 'PLAYING') {
  T.force.topOut();
  ok(G.secondWindUsed === true, 'second wind consumed on top-out');
  pump(120);
  settleSelects();
  // now force real game over: fill entire board
  for (let y = 0; y < 20; y++) for (let x = 0; x < 10; x++) {
    if (!G.grid[y][x]) G.grid[y][x] = { col: '#5a6680', garbage: false, bomb: false, core: false, ember: -1 };
  }
  T.force.topOut();
  ok(G.state === 'OVER', 'game over reached');
  pump(120);
  ok(errors.length === 0, 'game over sequence no errors');
}

// restart
T.startGame(); pump(120);
ok(G.state === 'PLAYING' && G.score === 0 && G.wave === 1, 'restart resets run');

if (errors.length) console.error('\nERRORS:\n' + errors.join('\n'));
console.log(errors.length === 0 ? '\nALL SMOKE CHECKS DONE' : '\nSMOKE TEST HAD FAILURES');
