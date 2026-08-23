'use strict';
/* =========================================================
 * 雷霆突袭 —— 武器定义 / 关卡剧情 / Boss 数据
 * ========================================================= */

const WEAPONS = [
  { id: 'normal', name: '常规子弹', desc: '高速直射' },
  { id: 'spread', name: '排弹', desc: '扇形弹幕' },
  { id: 'laser', name: '激光', desc: '持续贯穿' },
  { id: 'homing', name: '追踪导弹', desc: '自动锁定' },
];

const BOSS_DEFS = [
  { name: '侦察旗舰「先锋」', hp: 1400, r: 56, color: '#7fd4ff' },
  { name: '风暴使者「雷翼」', hp: 2400, r: 62, color: '#9ef3ff' },
  { name: '空中堡垒「铁幕」', hp: 3800, r: 74, color: '#ffb84d' },
  { name: '雷皇「天罚」', hp: 6200, r: 80, color: '#ff6ec7' },
];

const STAGES = [
  {
    id: 1, name: '边境警报',
    story: [
      '凌晨 03:17，北部边境雷达亮起刺眼红点。',
      '敌方侦察机群越过国境线，正逼近沿海城市。',
      '王牌飞行员「苍隼」驾驶雷霆号升空拦截！',
    ],
    bgTop: '#071233', bgBottom: '#12224f', clouds: false, lightningRate: 0.10,
    waves: [
      { t: 1.0, type: 'scout', n: 4, pattern: 'line' },
      { t: 4.5, type: 'fighter', n: 2, pattern: 'v' },
      { t: 8.0, type: 'scout', n: 5, pattern: 'sides' },
      { t: 12.0, type: 'gunship', n: 1, pattern: 'line' },
      { t: 14.5, type: 'fighter', n: 3, pattern: 'v' },
      { t: 19.0, type: 'bomber', n: 1, pattern: 'line' },
      { t: 22.0, type: 'scout', n: 6, pattern: 'sides' },
      { t: 27.0, type: 'gunship', n: 2, pattern: 'v' },
      { t: 31.0, type: 'fighter', n: 4, pattern: 'line' },
      { t: 36.0, type: 'bomber', n: 2, pattern: 'sides' },
      { t: 41.0, type: 'scout', n: 8, pattern: 'sides' },
      { t: 45.5, type: 'gunship', n: 2, pattern: 'line' },
    ],
  },
  {
    id: 2, name: '风暴之眼',
    story: [
      '敌机残部遁入罕见雷暴云团，雷达信号全部丢失。',
      '云团中潜伏着敌方新型「风暴机」，以闪电为武器。',
      '穿过风暴之眼，摧毁它们的巢穴！',
    ],
    bgTop: '#0a0f2e', bgBottom: '#2a1a4d', clouds: true, lightningRate: 0.30,
    waves: [
      { t: 1.0, type: 'scout', n: 5, pattern: 'sides' },
      { t: 4.0, type: 'storm', n: 1, pattern: 'line' },
      { t: 7.5, type: 'fighter', n: 3, pattern: 'v' },
      { t: 11.0, type: 'storm', n: 2, pattern: 'sides' },
      { t: 15.0, type: 'gunship', n: 2, pattern: 'line' },
      { t: 19.0, type: 'scout', n: 6, pattern: 'sides' },
      { t: 23.0, type: 'storm', n: 2, pattern: 'v' },
      { t: 27.0, type: 'bomber', n: 1, pattern: 'line' },
      { t: 31.0, type: 'fighter', n: 4, pattern: 'sides' },
      { t: 35.0, type: 'storm', n: 3, pattern: 'line' },
      { t: 40.0, type: 'gunship', n: 2, pattern: 'v' },
      { t: 44.0, type: 'bomber', n: 2, pattern: 'sides' },
    ],
  },
  {
    id: 3, name: '钢铁堡垒',
    story: [
      '情报显示：敌军在高空部署了移动空中堡垒。',
      '厚重装甲与交叉火力网，是这片空域最大威胁。',
      '突破防线，直捣堡垒核心！',
    ],
    bgTop: '#101418', bgBottom: '#2c3440', clouds: false, lightningRate: 0.18,
    waves: [
      { t: 1.0, type: 'gunship', n: 2, pattern: 'line' },
      { t: 4.5, type: 'scout', n: 6, pattern: 'sides' },
      { t: 8.0, type: 'bomber', n: 1, pattern: 'line' },
      { t: 11.5, type: 'storm', n: 2, pattern: 'v' },
      { t: 15.0, type: 'fighter', n: 4, pattern: 'sides' },
      { t: 19.0, type: 'gunship', n: 3, pattern: 'v' },
      { t: 23.5, type: 'bomber', n: 2, pattern: 'line' },
      { t: 28.0, type: 'storm', n: 3, pattern: 'sides' },
      { t: 32.5, type: 'scout', n: 8, pattern: 'sides' },
      { t: 37.0, type: 'gunship', n: 2, pattern: 'line' },
      { t: 41.5, type: 'bomber', n: 2, pattern: 'v' },
      { t: 46.0, type: 'storm', n: 2, pattern: 'line' },
    ],
  },
  {
    id: 4, name: '雷霆之怒',
    story: [
      '堡垒深处，敌方最高指挥官「雷皇」现身。',
      '他驾驭风暴与闪电，誓要将天空化为焦土。',
      '最后的决战——用你的雷霆，终结这场风暴！',
    ],
    bgTop: '#1a0512', bgBottom: '#3d0f33', clouds: true, lightningRate: 0.35,
    waves: [
      { t: 1.0, type: 'scout', n: 8, pattern: 'sides' },
      { t: 5.0, type: 'storm', n: 2, pattern: 'line' },
      { t: 9.0, type: 'gunship', n: 3, pattern: 'v' },
      { t: 13.0, type: 'bomber', n: 2, pattern: 'sides' },
      { t: 17.0, type: 'fighter', n: 5, pattern: 'line' },
      { t: 21.5, type: 'storm', n: 3, pattern: 'v' },
      { t: 26.0, type: 'gunship', n: 3, pattern: 'sides' },
      { t: 30.5, type: 'bomber', n: 3, pattern: 'line' },
      { t: 35.0, type: 'scout', n: 10, pattern: 'sides' },
      { t: 40.0, type: 'storm', n: 3, pattern: 'v' },
      { t: 45.0, type: 'gunship', n: 3, pattern: 'line' },
      { t: 49.5, type: 'bomber', n: 2, pattern: 'sides' },
    ],
  },
];

const WIN_LINES = [
  '雷皇的旗舰在闪电中分崩离析，天空重归宁静。',
  '黎明破晓，城市的灯火重新亮起。',
  '谢谢你，飞行员。世界因你而平安。',
];
