const { MAPS } = require('../js/maps.js');
const fi = parseInt(process.argv[2] || '0');
const grid = MAPS[fi].map(r => [...r]);
let start = null;
for (let y = 0; y < 15; y++) for (let x = 0; x < 15; x++) if (grid[y][x] === '@') start = [x, y];
console.log('floor', fi + 1, 'start', start);
const seen = Array.from({ length: 15 }, () => new Array(15).fill(false));
const q = [];
if (start) { seen[start[1]][start[0]] = true; q.push(start); }
for (let y = 0; y < 15; y++) for (let x = 0; x < 15; x++) if (grid[y][x] === 'L') { seen[y][x] = true; q.push([x, y]); }
let steps = 0;
while (q.length) {
  const [x, y] = q.shift();
  for (const [dx, dy] of [[1, 0], [-1, 0], [0, 1], [0, -1]]) {
    const nx = x + dx, ny = y + dy;
    if (nx < 0 || ny < 0 || nx > 14 || ny > 14) continue;
    const c = grid[ny][nx];
    if (c === '#' || seen[ny][nx]) continue;
    seen[ny][nx] = true; q.push([nx, ny]);
    if (++steps % 25 === 0) console.log('step', steps, 'at', nx, ny, 'char', c);
  }
}
for (let y = 0; y < 15; y++) {
  let line = '';
  for (let x = 0; x < 15; x++) line += seen[y][x] ? grid[y][x] : '·';
  console.log(String(y).padStart(2), '|', line, '|');
}
