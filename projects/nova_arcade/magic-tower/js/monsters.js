/* ============================================================
 * 魔塔 — 怪物 / 道具数值表
 * 战斗公式（经典魔塔）：伤害 = max(0, 攻 - 防)
 * 玩家先手；敌人被击杀当回合不再反击。
 * ============================================================ */
(function (root, factory) {
  const m = factory();
  if (typeof module !== 'undefined' && module.exports) module.exports = m;
  root.MT = Object.assign(root.MT || {}, m);
})(typeof self !== 'undefined' ? self : globalThis, function () {

  const MONSTERS = {
    1: { id: 'slime',    name: '史莱姆',   hp: 600,  atk: 70,  def: 50,  gold: 50,  boss: false },
    2: { id: 'bat',      name: '蝙蝠',     hp: 700,  atk: 90,  def: 50,  gold: 60,  boss: false },
    3: { id: 'skeleton', name: '骷髅兵',   hp: 2800, atk: 180, def: 90,  gold: 150, boss: false },
    4: { id: 'goblin',   name: '哥布林',   hp: 2400, atk: 190, def: 100, gold: 180, boss: false },
    5: { id: 'zombie',   name: '僵尸',     hp: 2200, atk: 240, def: 110, gold: 300, boss: false },
    6: { id: 'demon',    name: '恶魔',     hp: 2200, atk: 280, def: 110, gold: 400, boss: false },
    7: { id: 'dragon',   name: '龙王',     hp: 4000, atk: 300, def: 80,  gold: 2000, boss: true },
  };

  const ITEMS = {
    p: { id: 'potion',    name: '生命药水',   kind: 'hp',   val: 1500, text: '+1500 HP' },
    q: { id: 'bigpotion', name: '大生命药水', kind: 'hp',   val: 4000, text: '+4000 HP' },
    s: { id: 'sword',     name: '宝剑',       kind: 'atk',  val: 5,    text: '+5 攻击' },
    S: { id: 'bigsword',  name: '魔剑',       kind: 'atk',  val: 20,   text: '+20 攻击' },
    d: { id: 'shield',    name: '盾牌',       kind: 'def',  val: 5,    text: '+5 防御' },
    D: { id: 'bigshield', name: '圣盾',       kind: 'def',  val: 20,   text: '+20 防御' },
    $: { id: 'coins',     name: '金币',       kind: 'gold', val: 100,  text: '+100 金币' },
    C: { id: 'chest',     name: '宝箱',       kind: 'gold', val: 800,  text: '+800 金币' },
  };

  const DOORS = {
    r: { need: 'red',   name: '红门' },
    b: { need: 'blue',  name: '蓝门' },
    g: { need: 'gold',  name: '金门' },
  };
  const KEYS = {
    R: 'red', B: 'blue', G: 'gold',
  };

  /* 玩家初始状态 */
  const PLAYER_START = { hp: 8000, atk: 100, def: 100, gold: 0, keys: { red: 0, blue: 0, gold: 0 } };

  return { MONSTERS, ITEMS, DOORS, KEYS, PLAYER_START };
});
