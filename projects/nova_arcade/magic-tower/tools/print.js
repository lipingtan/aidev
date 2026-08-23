/* 以文本形式打印精灵，便于无视觉模型审查形状 */
const { SPRITES } = require('../js/sprites.js');
const which = process.argv[2];
console.log('      ' + '0123456789ABCDEF'.replace(/./g, (c, i) => (i % 4 === 0 ? c : ' ')));
for (const [name, sp] of Object.entries(SPRITES)) {
  if (which && !name.includes(which)) continue;
  console.log('\n=== ' + name + ' (' + sp.label + ') ===');
  sp.rows.forEach((r, i) => console.log(String(i).padStart(2) + ' |' + r + '|'));
}
// 结构指标：包围盒 / 对称度 / 悬浮像素
console.log('\n--- metrics ---');
for (const [name, sp] of Object.entries(SPRITES)) {
  const cells = [];
  sp.rows.forEach((r, y) => [...r].forEach((ch, x) => { if (ch !== '.') cells.push([x, y]); }));
  if (!cells.length) continue;
  let minX = 99, maxX = -1, minY = 99, maxY = -1;
  const set = new Set(cells.map(c => c[0] + ',' + c[1]));
  let floating = 0;
  for (const [x, y] of cells) {
    minX = Math.min(minX, x); maxX = Math.max(maxX, x);
    minY = Math.min(minY, y); maxY = Math.max(maxY, y);
    if (!set.has((x - 1) + ',' + y) && !set.has(x + ',' + (y - 1)) && !set.has((x + 1) + ',' + y) && !set.has(x + ',' + (y + 1))) floating++;
  }
  let asym = 0;
  for (const [x, y] of cells) if (!set.has((15 - x) + ',' + y)) asym++;
  console.log(`${name}: bbox=${maxX - minX + 1}x${maxY - minY + 1} at(${minX},${minY}) floating=${floating} asymPixels=${asym}/${cells.length}`);
}
