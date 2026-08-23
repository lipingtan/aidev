'use strict';
/* 无头冒烟测试：用桩件模拟 DOM，加载四个游戏脚本并模拟数千帧，
 * 覆盖菜单→剧情→战斗→Boss→关卡完成→下一幕全流程，捕获运行时错误。 */
const fs = require('fs');
const path = require('path');
const vm = require('vm');

let nowMs = 0;
let rafCb = null;
const handlers = { keydown: [], keyup: [] };

// ---- Canvas 2D 桩：所有方法 no-op，渐变返回带 addColorStop 的对象 ----
function makeCtx() {
  const target = {};
  return new Proxy(target, {
    get(t, k) {
      if (typeof k === 'symbol') return undefined;
      if (!(k in t)) {
        t[k] = (...a) => {
          if (k === 'createLinearGradient' || k === 'createRadialGradient') return { addColorStop() {} };
          if (k === 'measureText') return { width: 10 };
          return undefined;
        };
      }
      return t[k];
    },
    set(t, k, v) { t[k] = v; return true; },
  });
}
const ctxStub = makeCtx();
const canvasEl = {
  width: 960, height: 640, style: {},
  getContext: () => ctxStub,
  addEventListener: (t, f) => { handlers[t] = handlers[t] || []; handlers[t].push(f); },
};

const store = {};
const sandbox = {
  console, Math, JSON,
  performance: { now: () => nowMs },
  requestAnimationFrame: cb => { rafCb = cb; return 1; },
  innerWidth: 1280, innerHeight: 800,
  localStorage: {
    getItem: k => (k in store ? store[k] : null),
    setItem: (k, v) => { store[k] = String(v); },
  },
  setTimeout: () => 0, // 测试中不跑真实定时器
  addEventListener: (t, f) => { handlers[t] = handlers[t] || []; handlers[t].push(f); },
};
sandbox.window = sandbox;
sandbox.document = {
  getElementById: () => canvasEl,
  addEventListener: (t, f) => { handlers[t] = handlers[t] || []; handlers[t].push(f); },
};

const ctxObj = vm.createContext(sandbox);
const dir = path.join(__dirname, 'airwar');
for (const f of ['js/utils.js', 'js/stages.js', 'js/enemies.js', 'js/game.js']) {
  vm.runInContext(fs.readFileSync(path.join(dir, f), 'utf8'), ctxObj, { filename: f });
}

// 暴露测试钩子（同一上下文可访问顶层 let/const）
vm.runInContext(`
  globalThis.__T = {
    startGame, loadStage, useBomb, switchWeapon, nextStage, damageBoss,
    get state() { return state; }, set state(v) { state = v; },
    get stageTime() { return stageTime; }, set stageTime(v) { stageTime = v; },
    get wavePtr() { return wavePtr; }, set wavePtr(v) { wavePtr = v; },
    get boss() { return boss; },
    get score() { return score; },
    player, keys, enemies, pbullets, ebullets, drops,
  };
`, ctxObj);

const T = sandbox.__T;

function frames(n) {
  for (let i = 0; i < n; i++) {
    nowMs += 16.7;
    const cb = rafCb; rafCb = null;
    if (!cb) throw new Error('requestAnimationFrame 未重新注册');
    cb(nowMs);
  }
}
function key(k, down = true) {
  const list = down ? handlers.keydown : handlers.keyup;
  for (const f of list) f({ key: k, repeat: false, preventDefault() {} });
}

let failures = 0;
function check(name, cond) {
  console.log((cond ? 'PASS' : 'FAIL') + '  ' + name);
  if (!cond) failures++;
}

try {
  // 1. 菜单
  frames(60);
  check('初始为菜单状态', T.state === 'menu');

  // 2. 开始游戏 → 剧情 → 战斗
  key('Enter');
  frames(5);
  check('进入剧情状态', T.state === 'intro');
  frames(380); // ~6.3s > 5.5s 自动进入战斗
  if (T.state !== 'play') { key('Enter'); frames(400); } // 若提前失败则重开
  check('进入战斗状态', T.state === 'play' || T.state === 'over');

  // 3. 战斗中：移动 / 武器切换 / 炸弹
  const px = T.player.x, py = T.player.y;
  key('d'); frames(60); key('d', false);
  key('w'); frames(40); key('w', false);
  key('a'); frames(30); key('a', false);
  check('WASD 移动生效', Math.abs(T.player.x - px) > 5 || Math.abs(T.player.y - py) > 5);
  if (T.state === 'play') {
    key('2'); key('3'); key('4'); key('1');
    check('武器切换（回到常规）', T.player.weapon === 'normal');
    const bombs0 = T.player.bombs;
    key(' '); frames(10);
    check('炸弹可用且扣减', T.player.bombs === bombs0 - 1);
  }

  // 4. 战斗持续运行（随机移动 + 各武器轮流）
  const moves = ['w', 'a', 's', 'd'];
  for (let s = 0; s < 60 && T.state === 'play'; s++) {
    key(moves[s % 4]); frames(8); key(moves[s % 4], false);
    if (s % 15 === 7) key('1');
    if (s % 15 === 12) key('2');
  }
  check('战斗阶段无异常', T.state === 'play' || T.state === 'over');

  // 5. 若失败则重开，然后快进触发 Boss（玩家无敌保证确定性）
  if (T.state !== 'play') { key('Enter'); frames(400); }
  if (T.state === 'play') {
    T.player.invuln = 1e6; T.player.shield = 1e6; // 测试用无敌
    T.stageTime = 999; T.wavePtr = 999; // 清空波次 → Boss 警告
    for (let i = 0; i < 800 && !T.boss; i++) frames(1);
    check('Boss 出现', !!T.boss);
    if (T.boss) {
      // 等入场完成，用激光打几帧验证激光伤害路径
      for (let i = 0; i < 300 && T.boss && T.boss.y < 60; i++) frames(1);
      key('3');
      T.player.x = 480; T.player.y = 560;
      // 等 Boss 巡逻到激光正上方，保证伤害判定确定发生
      for (let i = 0; i < 900 && !(T.boss && Math.abs(T.boss.x - 480) < 30); i++) frames(1);
      const bhp = T.boss ? T.boss.hp : -1;
      frames(30);
      check('激光对 Boss 造成伤害', T.boss && T.boss.hp < bhp);
      // 直接结算 Boss → 关卡完成
      if (T.boss) { T.damageBoss(999999); frames(5); }
      check('Boss 被击败进入关卡完成', T.state === 'clear');
    }
  }

  // 6. 关卡完成 → 下一幕（第二幕：风暴之眼，含风暴机）
  if (T.state === 'clear') {
    key('Enter'); frames(400);
    check('进入第二幕', T.state === 'play' || T.state === 'intro' || T.state === 'over');
  }

  // 7. 若在第二幕失败 → 重开回菜单链路
  if (T.state === 'over') { key('Enter'); frames(400); }

  console.log('\n最终状态: ' + T.state + ' | 分数: ' + T.score + ' | 敌机: ' + T.enemies.length + ' | 玩家子弹: ' + T.pbullets.length + ' | 掉落: ' + T.drops.length);
} catch (err) {
  failures++;
  console.error('RUNTIME ERROR:', err && err.stack || err);
}

console.log(failures === 0 ? '\n=== SMOKE TEST PASSED ===' : `\n=== ${failures} FAILURES ===`);
process.exit(failures === 0 ? 0 : 1);
