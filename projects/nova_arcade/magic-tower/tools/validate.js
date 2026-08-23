/* 地图结构 + 可达性验证（门视为可开启的过度近似） */
const { MAPS } = require('../js/maps.js');

const VALID = new Set(['#', '.', '@', 'r', 'b', 'g', 'R', 'B', 'G',
  '1', '2', '3', '4', '5', '6', '7',
  'p', 'q', 's', 'S', 'd', 'D', '$', 'C', 'U', 'L']);

let fail = false;
const err = (m) => { fail = true; console.error('  ✗ ' + m); };

MAPS.forEach((rows, fi) => {
  const floor = fi + 1;
  console.log(`\n=== Floor ${floor} ===`);
  if (rows.length !== 15) err(`rows=${rows.length}`);
  rows.forEach((row, y) => {
    if (row.length !== 15) err(`row${y} len=${row.length}: "${row}"`);
    for (const ch of row) if (!VALID.has(ch)) err(`row${y} bad char '${ch}'`);
    if (row[0] !== '#' || row[14] !== '#') err(`row${y} edge not wall: "${row}"`);
  });
  const grid = rows.map(r => [...r]);

  // 统计
  const counts = {};
  let start = null;
  for (let y = 0; y < 15; y++) for (let x = 0; x < 15; x++) {
    const c = grid[y][x];
    counts[c] = (counts[c] || 0) + 1;
    if (c === '@') start = [x, y];
  }
  console.log('  counts:', JSON.stringify(counts));

  // BFS（所有门视为可通行）
  const seen = Array.from({ length: 15 }, () => new Array(15).fill(false));
  const q = [];
  if (start) { seen[start[1]][start[0]] = true; q.push(start); }
  else if (floor === 1) err('no @ start');
  // 楼梯入口视为可达源（从下层上来）
  for (let y = 0; y < 15; y++) for (let x = 0; x < 15; x++) {
    if (grid[y][x] === 'L' && floor > 1) { seen[y][x] = true; q.push([x, y]); }
  }
  const DIRS = [[1, 0], [-1, 0], [0, 1], [0, -1]];
  while (q.length) {
    const [x, y] = q.shift();
    for (const [dx, dy] of DIRS) {
      const nx = x + dx, ny = y + dy;
      if (nx < 0 || ny < 0 || nx > 14 || ny > 14) continue;
      const c = grid[ny][nx];
      if (c === '#' || seen[ny][nx]) continue;
      seen[ny][nx] = true; q.push([nx, ny]);
    }
  }
  let unreachable = 0;
  for (let y = 0; y < 15; y++) for (let x = 0; x < 15; x++) {
    const c = grid[y][x];
    if (c !== '#' && !seen[y][x]) { unreachable++; err(`unreachable (${x},${y}) '${c}'`); }
  }
  if (!unreachable) console.log('  ✓ all tiles reachable');

  // 钥匙/门数量
  for (const [door, key] of [['r', 'R'], ['b', 'B'], ['g', 'G']]) {
    const d = counts[door] || 0, k = counts[key] || 0;
    if (d > 0 && k < d) err(`doors ${door}=${d} > keys ${key}=${k}`);
  }

  // 楼梯检查
  const hasU = !!counts['U'], hasL = !!counts['L'];
  if (floor === 1 && !hasU) err('F1 missing U');
  if (floor === 2 && !(hasU && hasL)) err('F2 needs U+L');
  if (floor === 3 && !hasL) err('F3 missing L');
});

process.exit(fail ? 1 : 0);
