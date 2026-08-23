/* ============================================================
 * 魔塔 Magic Tower — 像素精灵数据 (16x16)
 * 浏览器: window.MT.SPRITES / MT.makeSpriteCanvas
 * Node:   module.exports (用于 tools/ 下的 QA 脚本)
 * ============================================================ */
(function (root, factory) {
  const m = factory();
  if (typeof module !== 'undefined' && module.exports) module.exports = m;
  root.MT = Object.assign(root.MT || {}, m);
})(typeof self !== 'undefined' ? self : globalThis, function () {

  const K = '#131422'; // 通用描边黑

  /* ---------- 工具：16x16 放大到 32x32 / 左右镜像 ---------- */
  function up16(sp) {
    const rows = [];
    for (let y = 0; y < sp.rows.length; y++) {
      let out = '';
      for (const ch of sp.rows[y]) out += ch + ch;
      rows.push(out);
    }
    return Object.assign({}, sp, { rows });
  }
  function mirror(sp) {
    return Object.assign({}, sp, { rows: sp.rows.map(r => [...r].reverse().join('')) });
  }

  /* ---------- 英雄：骑士（32x32 高清，4朝向 × 2帧走路） ---------- */
  const HERO_PAL = {
    K,
    R: '#ff5a4f', r: '#b02a22',            // 红缨 亮/暗
    S: '#eef2fb', s: '#8b96b5', w: '#ffffff', // 钢甲 亮/暗/白色高光
    E: '#bffaff', e: '#37d3f0',           // 面甲辉光 外/内
    G: '#ffe066', g: '#c98a1e',           // 金 亮/暗
    B: '#5a74e0', b: '#2f3f9e',          // 披风 亮/暗
    L: '#8a5636', l: '#5e3a22',          // 靴 亮/暗
  };
  // 行辅助：c32=内容居中到 32 列（对称美术），l32=左对齐右补点（侧面）
  function c32(s) {
    if (s.length > 32) throw new Error('c32 too wide: ' + s);
    const pad = 32 - s.length, l = Math.floor(pad / 2), r = pad - l;
    return '.'.repeat(l) + s + '.'.repeat(r);
  }
  function l32(s) {
    if (s.length > 32) throw new Error('l32 too wide: ' + s);
    return s + '.'.repeat(32 - s.length);
  }
  const E32 = '................................'; // 32 个透明点

  /* 朝下（正面，站立） */
  const heroD = {
    name: 'hero_d', label: '勇者·前', pal: HERO_PAL,
    rows: [
      E32,
      c32('RR'),                          // 1 缨尖
      c32('RRRr'),                        // 2 红缨（右缘暗）
      c32('KRRrrK'),                      // 3 缨基
      c32('KwwSSSsSsssK'),               // 4 盔顶（白高光+抖动过渡）
      c32('KwSSSsSsssssK'),              // 5 盔体（明暗交界抖动）
      c32('KSSEEEwwESSK'),               // 6 发光面甲（白色亮核）
      c32('KSSEEEwwESSK'),               // 7 面甲辉光带
      c32('KwSSSsSsssssK'),              // 8 盔下（抖动）
      c32('KSSSsSsssK'),                 // 9 盔底收窄（抖动）
      c32('KGgGGggGK'),                  // 10 金色护颊（金/暗抖动）
      c32('KBBBwSSSsSssBBBK'),           // 11 肩（披风+钢胸，抖动）
      c32('KBBBSSSsSsssBBBK'),           // 12
      c32('KBBBwGGGggBBBK'),             // 13 胸前金徽（白高光+抖动）
      c32('KBBBGgGGggBBBK'),             // 14 金徽（棋盘抖动）
      c32('KBBBsSsSSsBBBK'),             // 15 下胸（抖入阴影）
      c32('KBBBssssssBBBK'),             // 16
      c32('KgggGwwGggK'),                // 17 腰带+白色金扣
      c32('KSSsK....KSSsK'),             // 18 双腿（上亮下暗）
      c32('KSSsK....KSSsK'),             // 19
      c32('KSSsK....KSSsK'),             // 20
      c32('KSSsK....KSSsK'),             // 21
      c32('KSSsK....KSSsK'),             // 22
      c32('KSSsK....KSSsK'),             // 23
      c32('KSSsK....KSSsK'),             // 24
      c32('KSSsK....KSSsK'),             // 25
      c32('KLLlK....KLLlK'),             // 26 靴（亮/暗）
      c32('KLLlK....KLLlK'),             // 27
      E32,                               // 28
      E32,                               // 29
    ],
  };

  /* 正面迈步（腿交叉）：复用 heroD 上身，替换腿/靴 */
  const D1_LEGS = [
    '........KssssK....KssssK........', // 18 双腿张开
    '........KssssK....KssssK........', // 19
    '........KssssK....KssssK........', // 20
    '........KssssK.......KsssK......', // 21 右腿抬起
    '........KssssK.......KsssK......', // 22
    '........KssssK..................', // 23 左腿
    '........KssssK..................', // 24
    '......KLLLLLK........KLLLLLK....', // 25 靴（前/后）
    E32,                                  // 26
    E32,                                  // 27
  ];
  const heroD1 = Object.assign({}, heroD, {
    name: 'hero_d1', label: '勇者·前(走)',
    rows: heroD.rows.slice(0, 18).concat(D1_LEGS),
  });

  /* 朝上（背面）：复用 heroD，替换头部（缨垂背后、无面甲） */
  const U_HEAD = [
    c32('KwwSSSSssssK'),           // 4 盔顶（白高光）
    c32('KwSSSSSSssssK'),          // 5 盔体
    c32('KSRRRRrrSK'),             // 6 背后红缨（右暗）
    c32('KSRRRRrrSK'),             // 7 红缨
    c32('KwSSSSSSssssK'),          // 8 盔下
    c32('KRRRrrrRK'),              // 9 缨垂颈后
    c32('KGgggGGgggK'),            // 10 金领
  ];
  const heroU = Object.assign({}, heroD, {
    name: 'hero_u', label: '勇者·后',
    rows: heroD.rows.map((r, i) => (i >= 4 && i <= 10 ? U_HEAD[i - 4] : r)),
  });

  /* 背面迈步：复用 heroU，替换腿/靴 */
  const heroU1 = Object.assign({}, heroU, {
    name: 'hero_u1', label: '勇者·后(走)',
    rows: heroU.rows.slice(0, 18).concat(D1_LEGS),
  });

  /* 朝左（侧面，站立） */
  const heroL = {
    name: 'hero_l', label: '勇者·左', pal: HERO_PAL,
    rows: [
      E32,
      l32('..............RR'),            // 1 缨尖
      l32('.............RRRRR'),          // 2 红缨
      l32('............KRRRRK'),         // 3 缨基
      l32('..........KwwSSSSssssK'),     // 4 盔顶（白高光/右暗）
      l32('.........KEEwESSSssssK'),     // 5 面甲（前亮核/后暗）
      l32('.........KEEwESSSssssK'),     // 6 面甲
      l32('..........KGwGGggK'),         // 7 金护颊（高光/暗）
      l32('.......KBBBwSSSSssssK'),       // 8 肩/披风（白高光）
      l32('........KBBSSSSSssssK'),       // 9 躯干（右暗）
      l32('........KBBSSSSSssssK'),       // 10
      l32('........KBBSSSSSssssK'),       // 11
      l32('........KBBSSSSSssssK'),       // 12
      l32('........KBBSSSSSssssK'),       // 13
      l32('........KBBSSSSSssssK'),       // 14
      l32('........KgggGGGgggK'),         // 15 腰带
      l32('.......KssssssssK'),           // 16 胯
      l32('..........KssssK.KssK'),       // 17 近腿+远腿
      l32('..........KssssK.KssK'),       // 18
      l32('..........KssssK.KssK'),       // 19
      l32('..........KssssK.KssK'),       // 20
      l32('..........KssssK.KssK'),       // 21
      l32('..........KssssK.KssK'),       // 22
      l32('..........KssssK.KssK'),       // 23
      l32('..........KssssK.KssK'),       // 24
      l32('.........KLLLLLK..KLK'),       // 25 近靴+远靴
      l32('.........KLLLLLK..KLK'),       // 26
      E32,                                // 27
      E32,                                // 28
      E32,                                // 29
    ],
  };

  /* 侧面迈步（腿交叉）：复用 heroL 上身，替换腿/靴 */
  const L1_LEGS = [
    l32('.........KssssK...KssK'),   // 17 近腿前+远腿后
    l32('.........KssssK...KssK'),   // 18
    l32('.........KssssK...KssK'),   // 19
    l32('.........KssssK...KssK'),   // 20
    l32('.........KssssK'),          // 21 远腿抬起
    l32('.........KssssK'),          // 22
    l32('.........KssssK'),          // 23
    l32('.........KssssK'),          // 24
    l32('........KLLLLLK...KLK'),    // 25 靴
    E32,                             // 26
  ];
  const heroL1 = Object.assign({}, heroL, {
    name: 'hero_l1', label: '勇者·左(走)',
    rows: heroL.rows.slice(0, 17).concat(L1_LEGS),
  });

  /* 朝右 = 朝左镜像 */
  const heroR = Object.assign(mirror(heroL), { name: 'hero_r', label: '勇者·右' });
  const heroR1 = Object.assign(mirror(heroL1), { name: 'hero_r1', label: '勇者·右(走)' });

  const HERO_FRAMES = {
    d: [heroD, heroD1], u: [heroU, heroU1], l: [heroL, heroL1], r: [heroR, heroR1],
  };

  /* ---------- 史莱姆（绿色史莱姆） ---------- */
  const slime = {
    name: 'slime', label: '史莱姆',
    pal: { K, G: '#59d84e', g: '#2f9c30', H: '#b6f5a8', W: '#ffffff', M: '#17301a' },
    rows: [
      "................",
      "................",
      "................",
      "......KKKK......",
      "....KGWGGGGK....",
      "...KGHHHGGGGK...",
      "..KGHHGGGGGGK...",
      ".KGHGGGGGGGGGK..",
      ".KGGWWGGGWWGGK..",
      ".KGGWWGGGWWGGK..",
      ".KGGGMGGGGMGGK..",
      ".KGGGGGGGGGGGK..",
      "..KGGGGGGGGGK...",
      "...KKgggggKK....",
      "....KKKKKK......",
      "................",
    ],
  };

  /* ---------- 蝙蝠（紫翼蝙蝠） ---------- */
  const bat = {
    name: 'bat', label: '蝙蝠',
    pal: { K, P: '#8a5cf5', p: '#5b3aa8', D: '#3a2470', E: '#ff5c5c', W: '#ffffff' },
    rows: [
      "................",
      "....KK....KK....",
      "...KPPK..KPPK...",
      "..KPPKKKKKKPPK..",
      ".KPPPEPPEPPPPK..",
      ".KPWDDDDDPPK....",
      "..KPPPPPPPPPPK..",
      "KPPPKDDDDDDKPPPK",
      "KPPPKDDDDDDKPPPK",
      "K..K..DDDD..K..K",
      ".....KDDDK......",
      "....KPK..KPK....",
      "....KK....KK....",
      "................",
      "................",
      "................",
    ],
  };

  /* ---------- 骷髅兵 ---------- */
  const skeleton = {
    name: 'skeleton', label: '骷髅兵',
    pal: { K, W: '#e8ecf4', w: '#b9c2d4', D: '#6a7388', B: '#5a4632' },
    rows: [
      "................",
      "....KKKKKK......",
      "...KWWWWWWK.....",
      "..KWWWWWWWWK....",
      "..KWDDWWDDWK....",
      "..KWWWWWWWWK....",
      "...KWWDWWDWK....",
      "....KKKKKK......",
      ".....KWWK.......",
      "...KWWWWWWK.....",
      "..KWKWWKWWK.....",
      "..KWKWWKWWK.....",
      "...KWWKWWK......",
      "....KW..WK......",
      "...KWK..KWK.....",
      "................",
    ],
  };

  /* ---------- 哥布林（绿皮持棒） ---------- */
  const goblin = {
    name: 'goblin', label: '哥布林',
    pal: { K, G: '#6fae4a', g: '#41702a', E: '#ff4030', W: '#ffffff', L: '#8a5a2b' },
    rows: [
      "...........KLK..",
      "....KKKKKK..L...",
      "...KGWGGGK..L...",
      "..KGEGGGGEGKL...",
      "..KGWKGGGWGKL...",
      "...KKKKKKKKK....",
      "....KLLLLLLK....",
      "....KLLLLLLK....",
      "...KGGLLGGLGK...",
      "...KGGLLGGLGK...",
      "....KGG..GGK....",
      "....KGK..KGK....",
      "...KKKK..KKKK...",
      "................",
      "................",
      "................",
    ],
  };

  /* ---------- 僵尸 ---------- */
  const zombie = {
    name: 'zombie', label: '僵尸',
    pal: { K, Z: '#7a9c5a', z: '#4e6b38', D: '#2f4224', W: '#d8e0c8', E: '#ffd040' },
    rows: [
      "................",
      ".....KKKKKK.....",
      "...KZZZZZZZK....",
      "..KWZZZZZZZK....",
      "..KEZZZZZEZK....",
      "..KZZDDDDZZK....",
      "...KZZZZZZK.....",
      "....KWWWK.......",
      "...KZWWWWZK.....",
      "...KZWZZWZK.....",
      "...KZZZZZZK.....",
      ".....KZ.ZK......",
      ".....KZ.ZK......",
      ".....KK..KK.....",
      "................",
      "................",
    ],
  };

  /* ---------- 恶魔（红角蝠翼） ---------- */
  const demon = {
    name: 'demon', label: '恶魔',
    pal: { K, R: '#c0392b', r: '#7e1f16', E: '#ffd040', W: '#ffffff', D: '#3a1410' },
    rows: [
      "....KK....KK....",
      "...KRRK..KRRK...",
      "..KRRKKKKKKRRK..",
      ".KRREERRREERRRK.",
      ".KRRDDDDDDDRRK..",
      ".KRRRWRRRRWRRK..",
      "..KWRRRRRRRRK...",
      "..KRrRRRRRrRK...",
      "...KRRRRRRRK....",
      "...KRrRRRrRK....",
      "....KRRRRRK.....",
      "....KRK.KRK.....",
      "....KKK.KKK.....",
      "................",
      "................",
      "................",
    ],
  };

  /* ---------- 龙王（最终Boss，金冠红龙） ---------- */
  const dragon = {
    name: 'dragon', label: '龙王',
    pal: { K, R: '#d8433a', r: '#96241d', G: '#ffd24a', g: '#b8860b', E: '#ffe066', W: '#ffffff', D: '#5c130f' },
    rows: [
      "....G.G..G.G....",
      "...KGGK..KGGK...",
      "...KGgK..KGgK...",
      "..KKRRKKKKRRKK..",
      ".KWRRRRRRRRRRK..",
      ".KRRREEEERRRK...",
      ".KRRDDDDDDDRRK..",
      ".KRRRWRRRRWRRK..",
      "..KRRRRRRRRRK...",
      "..KrRRrrrRRrRK..",
      "...KRRRRRRRK....",
      "....KRK.KRK.....",
      "....KKK.KKK.....",
      "................",
      "................",
      "................",
    ],
  };

  /* ---------- 狼人（狼魔，参考 refimg/wolf.png） ---------- */
  const wolfman = {
    name: 'wolfman', label: '狼人',
    pal: { K:'#1c1512', F:'#a8a8b0', f:'#72727c', d:'#46464e', D:'#2a2a30', E:'#ff3b2e', e:'#ff9a5c', W:'#e8e4d6', L:'#7a4a2b', l:'#4a2c16', M:'#c0bcb2' },
    rows: [
      E32,                                       // 0
      c32('KfFK......KfFK'),                     // 1 耳尖（刺毛）
      c32('KFfFK....KFfFK'),                     // 2 双耳
      c32('KFFFKKKKKKFFFK'),                     // 3 头顶毛冠（中间凹）
      c32('KFFFFFddddFFFFK'),                    // 4 额头（浅毛+中央阴影）
      c32('KFFffFddFFFfFK'),                     // 5 眉弓（毛+暗线）
      c32('KFFeEEFFFEEeFFK'),                    // 6 发红双眼（外圈辉光 e）
      c32('KFFfffffFFK'),                        // 7 脸颊（毛，收窄成狼吻）
      c32('KMWWDDDWMK'),                         // 8 吻部：浅口鼻 M + 白獠牙 W + 暗口 D
      c32('KfMWWWWWfK'),                         // 9 下颚一排利齿（尖收）
      c32('KffddddddffK'),                       // 10 下颌/颈（深毛）
      c32('KFFfFFFFLLLLFFFFfFFK'),               // 11 宽肩+上臂（毛+皮带束胸）
      c32('KFfffFLLffLLFfffFK'),                 // 12 胸（皮带交叉+手臂毛）
      c32('KFFffFddddFFffFK'),                   // 13 下胸（毛+暗）
      c32('KffLLLLLLLLffK'),                     // 14 皮腰带（宽）
      c32('KfdllllllldfK'),                      // 15 手臂垂下（毛 d+皮带 l）
      c32('KfWdFddFdWfK'),                       // 16 爪手（白爪尖 W 在外侧）
      c32('KfdddddddffK'),                       // 17
      c32('KffddddffK'),                         // 18 髋收窄
      c32('KffdffK..KffdffK'),                   // 19 双腿分叉
      c32('KffK....KffK'),                       // 20 大腿
      c32('KfdK....KfdK'),                       // 21
      c32('KfKK....KfKK'),                       // 22 小腿
      c32('KdKK....KdKK'),                       // 23
      c32('KDWK....KWDK'),                       // 24 爪足（白爪尖 W）
      c32('KDDK....KDDK'),                       // 25 足底
      c32('KDDDK...KDDDK'),                      // 26 足跟
      E32,                                       // 27
      E32,                                       // 28
      E32,                                       // 29
      E32,                                       // 30
      E32,                                       // 31
    ],
  };

  /* ---------- 门（红/蓝/金） ---------- */
  function door(color, dark, label) {
    return {
      name: 'door_' + color, label,
      pal: { K, W: '#f0e6d2', L: '#8a5a2b', l: '#5c3a18', C: color, c: dark },
      rows: [
        "................",
        "....KKKKKKK.....",
        "...KCCCCCCK.....",
        "..KCLLLLLLCK....",
        "..KLWLLWLLLCK...",
        "..KLWLLWLLLCK...",
        "..KLLLLLLLWCK...",
        "..KLWLLWLLLCK...",
        "..KLWLLWLLLCK...",
        "..KLLLLLLLWCK...",
        "..KLWLLWLLLCK...",
        "..KLWLLWLLLCK...",
        "...KKKKKKKKK....",
        "................",
        "................",
        "................",
      ],
    };
  }

  /* ---------- 钥匙（红/蓝/金） ---------- */
  function key(color, dark, label) {
    return {
      name: 'key_' + color, label,
      pal: { K, C: color, c: dark, W: '#ffffff' },
      rows: [
        "................",
        "................",
        "....KKKK........",
        "...KCCCKK.......",
        "..KCWWCCKK......",
        "..KCWWCCCCK.....",
        "..KCCCCCCCCK....",
        "...KCCCKCCKK....",
        "....KKKK.CK.....",
        ".........CK.....",
        "........CCK.....",
        ".........CK.....",
        "........CCK.....",
        ".........CK.....",
        "................",
        "................",
      ],
    };
  }

  /* ---------- 药水（小/大） ---------- */
  const potion = {
    name: 'potion', label: '生命药水',
    pal: { K, L: '#7a4a2b', G: '#bcd8f0', R: '#e8433a', r: '#9c2620', W: '#ffffff' },
    rows: [
      "................",
      ".......KK.......",
      "......KLLK......",
      ".....KGGGK......",
      "....KGRRGK......",
      "...KGRRRRGK.....",
      "..KGWRRRRRGK....",
      "..KGWRRRRRGK....",
      "..KGRRRRRRGK....",
      "..KGrRRRRrGK....",
      "...KKrrrrKK.....",
      "....KKKKK.......",
      "................",
      "................",
      "................",
      "................",
    ],
  };

  const bigPotion = {
    name: 'bigpotion', label: '大生命药水',
    pal: { K, L: '#7a4a2b', G: '#c8b8f0', P: '#9c5ce8', p: '#6a2fa8', W: '#ffffff' },
    rows: [
      "................",
      "......KKK.......",
      ".....KLLLK......",
      "....KGGGGK......",
      "...KGPPPPGK.....",
      "..KGWPPPPPGK....",
      ".KGWPPPPPPPGK...",
      ".KGPPPPPPPPGK...",
      ".KGpPPPPPPpGK...",
      ".KGppPPPpppGK...",
      "..KKppppppKK....",
      "...KKKKKKK......",
      "................",
      "................",
      "................",
      "................",
    ],
  };

  /* ---------- 剑（小/大） ---------- */
  const sword = {
    name: 'sword', label: '攻击+5',
    pal: { K, S: '#dfe4f0', W: '#ffffff', G: '#ffd24a', L: '#7a4a2b' },
    rows: [
      "..........K.....",
      ".........KSK....",
      ".........KSK....",
      ".........KWK....",
      ".........KSK....",
      ".........KSK....",
      ".........KSK....",
      ".........KSK....",
      ".........KSK....",
      "......KGGGGK....",
      "........KLLK....",
      "........KLLK....",
      ".......KGGK.....",
      "................",
      "................",
      "................",
    ],
  };

  const bigSword = {
    name: 'bigsword', label: '攻击+20',
    pal: { K, S: '#e8f4ff', E: '#7df3ff', G: '#ffd24a', L: '#7a4a2b' },
    rows: [
      ".........EE.....",
      "........ESSK....",
      "........ESSK....",
      "........ESSK....",
      "........ESSK....",
      "........ESSK....",
      "........ESSK....",
      "........ESSK....",
      ".......KESSK....",
      ".KKKKKEGGGKKK...",
      "........KLLK....",
      "........KLLK....",
      ".......KGGK.....",
      "................",
      "................",
      "................",
    ],
  };

  /* ---------- 盾（小/大） ---------- */
  const shield = {
    name: 'shield', label: '防御+5',
    pal: { K, B: '#4a63d8', b: '#2e3d97', G: '#ffd24a' },
    rows: [
      "................",
      "....KKKKKK......",
      "...KBBBBBK......",
      "..KBBGGGBBK.....",
      "..KBGKKKGBK.....",
      "..KBGKGGKGBK....",
      "..KBGKGGKGBK....",
      "..KBGKKKGBK.....",
      "..KBBGGGBBK.....",
      "...KBBBBBK......",
      "....KBBBBK......",
      ".....KBBK.......",
      "......KK........",
      "................",
      "................",
      "................",
    ],
  };

  const bigShield = {
    name: 'bigshield', label: '防御+20',
    pal: { K, R: '#d8433a', r: '#96241d', G: '#ffd24a', E: '#7df3ff' },
    rows: [
      "................",
      ".....EEE........",
      "....KRRRRK......",
      "...KRGGGGRK.....",
      "..KRGGGGGGRK....",
      "..KRGGKKKGGRK...",
      "..KRGGKGGGGRK...",
      "..KRGGKGGGGRK...",
      "..KRGGKKKGGRK...",
      "..KRGGGGGGRK....",
      "...KRGGGGRK.....",
      "....KRRRRK......",
      ".....KKKK.......",
      "................",
      "................",
      "................",
    ],
  };

  /* ---------- 金币堆 ---------- */
  const coins = {
    name: 'coins', label: '金币',
    pal: { K, G: '#ffd24a', g: '#b8860b', W: '#fff2c0' },
    rows: [
      "................",
      "................",
      "................",
      "................",
      "................",
      "......KKK.......",
      ".....KGGGK......",
      "....KGWGGK......",
      "..KGGKGGGK......",
      ".KGGGGGGGGK.....",
      ".KGgGGGWgGK.....",
      "..KKKKKKKK......",
      "................",
      "................",
      "................",
      "................",
    ],
  };

  /* ---------- 宝箱 ---------- */
  const chest = {
    name: 'chest', label: '宝箱',
    pal: { K, L: '#8a5a2b', l: '#5c3a18', G: '#ffd24a', g: '#b8860b' },
    rows: [
      "................",
      "................",
      ".....KKKKK......",
      "....KLLLLLK.....",
      "...KLGLLGLLK....",
      "..KLLGGGGLLK....",
      "..KGGLLLLGGK....",
      "..KLLGLLGLLK....",
      "..KLLGLLGLLK....",
      "..KLLGLLGLLK....",
      "...KKKKKKKK.....",
      "................",
      "................",
      "................",
      "................",
      "................",
    ],
  };

  /* ---------- 楼梯（上/下） ---------- */
  const stairsUp = {
    name: 'stairsup', label: '上楼',
    pal: { K, S: '#8a93a7', s: '#565d72', G: '#ffd24a' },
    rows: [
      "................",
      "....GGG.........",
      "...KGGGK........",
      "....KKK.........",
      "..KSSSSSK.......",
      ".KSSSSSSSK......",
      "KSSSSSSSSSK.....",
      "KssssssssssK....",
      "KssssssssssK....",
      "KssssssssssK....",
      "KSSSSSSSSSK.....",
      "KssssssssssK....",
      ".KKKKKKKKK......",
      "................",
      "................",
      "................",
    ],
  };

  const stairsDown = {
    name: 'stairsdown', label: '下楼',
    pal: { K, S: '#8a93a7', s: '#565d72', G: '#ffd24a' },
    rows: [
      "................",
      "....KKK.........",
      "...KGGGK........",
      "....GGG.........",
      "..KSSSSSK.......",
      ".KsssssssK......",
      "KsssssssssK.....",
      "KSSSSSSSSSK.....",
      "KsssssssssK.....",
      "KSSSSSSSSSK.....",
      "KsssssssssK.....",
      "KSSSSSSSSSK.....",
      ".KKKKKKKKK......",
      "................",
      "................",
      "................",
    ],
  };

  const SPRITES = {
    hero: heroD,
    slime: up16(slime), bat: up16(bat), skeleton: up16(skeleton), goblin: up16(goblin),
    zombie: up16(zombie), demon: up16(demon), dragon: up16(dragon), wolfman: wolfman,
    door_red: up16(door('#e8433a', '#9c2620', '红门')),
    door_blue: up16(door('#4a7cf5', '#2c4fc0', '蓝门')),
    door_gold: up16(door('#ffd24a', '#b8860b', '金门')),
    key_red: up16(key('#e8433a', '#9c2620', '红钥匙')),
    key_blue: up16(key('#4a7cf5', '#2c4fc0', '蓝钥匙')),
    key_gold: up16(key('#ffd24a', '#b8860b', '金钥匙')),
    potion: up16(potion), bigPotion: up16(bigPotion), sword: up16(sword), bigSword: up16(bigSword),
    shield: up16(shield), bigShield: up16(bigShield), coins: up16(coins), chest: up16(chest),
    stairsUp: up16(stairsUp), stairsDown: up16(stairsDown),
  };
  // 小写别名（与 monsters.js 的 ITEMS.id 一致）
  SPRITES.bigpotion = SPRITES.bigPotion;
  SPRITES.bigsword = SPRITES.bigSword;
  SPRITES.bigshield = SPRITES.bigShield;

  /* 把精灵画到 canvas（整数倍缩放，保持像素锐利；支持 16/32 网格） */
  function makeSpriteCanvas(sp, scale) {
    const G = sp.rows.length; // 16 或 32
    const c = document.createElement('canvas');
    c.width = G * scale;
    c.height = G * scale;
    const ctx = c.getContext('2d');
    for (let y = 0; y < G; y++) {
      const row = sp.rows[y];
      for (let x = 0; x < G; x++) {
        const ch = row[x];
        if (!ch || ch === '.') continue;
        const col = sp.pal[ch];
        if (!col) continue;
        ctx.fillStyle = col;
        ctx.fillRect(x * scale, y * scale, scale, scale);
      }
    }
    return c;
  }

  return { SPRITES, HERO_FRAMES, makeSpriteCanvas };
});
