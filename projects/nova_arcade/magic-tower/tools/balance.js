/* 攻略路线数值模拟：
 * - 必战(主线)：必须可胜且战后存活，且保留安全余量
 * - 可选(支线)：验证可胜；仅在战后 HP >= SAFETY 时才打（模拟聪明玩家）
 */
const { MONSTERS, PLAYER_START } = require('../js/monsters.js');

const SAFETY = 4000; // 战后最低安全血量

function fight(p, m) {
  const pdmg = Math.max(0, p.atk - m.def);
  if (pdmg <= 0) return { win: false, loss: Infinity, rounds: Infinity };
  const rounds = Math.ceil(m.hp / pdmg);
  const edmg = Math.max(0, m.atk - p.def);
  const loss = edmg * (rounds - 1);
  return { win: true, loss, rounds };
}

const P = { ...PLAYER_START, keys: { ...PLAYER_START.keys } };
let ok = true;
const line = (s) => console.log(s);

line(`起点: HP=${P.hp} ATK=${P.atk} DEF=${P.def}`);
line('');

function item(label, f) { f(P); line(`  [拾取] ${label}  -> HP=${P.hp} ATK=${P.atk} DEF=${P.def}`); }

function battle(key, label, optional = false) {
  const m = MONSTERS[key];
  const r = fight(P, m);
  if (!r.win) {
    ok = !optional; // 必战不可胜=失败；可选不可胜只提示
    line(`  [战斗] ${label}: 无法战胜 (ATK${P.atk}<=DEF${m.def})${optional ? '（可绕开）' : ''}`);
    return;
  }
  const after = P.hp - r.loss;
  if (!optional) {
    P.hp = after;
    if (after <= 0) ok = false;
    line(`  [必战] ${label}: ${r.rounds}回合, 损失${r.loss}, 剩HP=${after}${after <= 0 ? ' ← 死亡!' : ''}`);
    return;
  }
  if (after >= SAFETY) {
    P.hp = after;
    line(`  [可选·打] ${label}: ${r.rounds}回合, 损失${r.loss}, 剩HP=${after}`);
  } else {
    line(`  [可选·跳过] ${label}: 需${r.loss}伤害, 战后仅剩${after} < ${SAFETY}, 留待后期/放弃`);
  }
}

line('=== B1（新手层，怪物防御50<攻击100，全部白打） ===');
item('剑x3 (s s s)', p => { p.atk += 15; });
item('盾x2 (d d)',   p => { p.def += 10; });
battle('1', '史莱姆(左)');
battle('1', '史莱姆(右)');
item('药水x2', p => { p.hp += 3000; });
battle('2', '蝙蝠(中)');
battle('2', '蝙蝠(红门口袋)');
item('金币x3', p => { p.gold += 300; });

line('=== B2（骷髅层） ===');
item('魔剑S (入口)', p => { p.atk += 20; });
item('盾d (入口)',   p => { p.def += 5; });
battle('3', '骷髅(中厅)');
item('药水p+大药水q (中厅)', p => { p.hp += 5500; });
battle('4', '哥布林(右上)', true);
battle('4', '哥布林(左下)', true);
item('剑s (中厅)', p => { p.atk += 5; });
item('圣盾D+宝箱C (蓝门后,含U)', p => { p.def += 20; p.gold += 800; });

line('=== B3（龙王层） ===');
item('魔剑S (入口)', p => { p.atk += 20; });
item('大药水x2 (入口)', p => { p.hp += 8000; });
battle('5', '僵尸(中厅)');
item('圣盾D (中厅)', p => { p.def += 20; });
battle('6', '恶魔(左下)', true);
item('大药水q (龙王房,先拾取再开战)', p => { p.hp += 4000; });
battle('7', '★龙王(Boss)');

line('');
if (ok && P.hp > 0) {
  line(`✓ 主线可行：最终 HP=${P.hp} ATK=${P.atk} DEF=${P.def} GOLD=${P.gold}`);
} else {
  line('✗ 主线存在不可战胜/致死战斗');
  process.exit(1);
}
