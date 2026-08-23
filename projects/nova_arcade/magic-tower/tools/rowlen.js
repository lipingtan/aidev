// 报告每个精灵每行的实际长度（所有 SPRITES 现在都应是 32 宽）
const s = require('../js/sprites.js');
let any = false;
function check(sp) {
  const lens = sp.rows.map(r => r.length);
  const uniq = [...new Set(lens)];
  if (uniq.length > 1 || uniq[0] !== 32) {
    any = true;
    console.log(`${sp.name}: rowlens=${uniq.join('/')} (want all 32)`);
    sp.rows.forEach((r, i) => { if (r.length !== 32) console.log(`   row ${i} (${r.length}): ${JSON.stringify(r)}`); });
  }
}
for (const [k, sp] of Object.entries(s.SPRITES)) if (sp.rows) check(sp);
for (const d of ['d', 'u', 'l', 'r']) for (const sp of s.HERO_FRAMES[d]) check(sp);
console.log(any ? '--- HAS BAD ROWS ---' : '--- ALL ROWS 32 OK ---');
