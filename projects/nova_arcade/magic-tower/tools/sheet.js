/* 渲染精灵联系表 PNG（无依赖：手写 PNG 编码 + zlib），并做结构校验 */
const path = require('path');
const zlib = require('zlib');
const fs = require('fs');
const { SPRITES } = require('../js/sprites.js');

// ---------- CRC32 / PNG ----------
let crcTable = null;
function crc32(buf) {
  if (!crcTable) {
    crcTable = new Int32Array(256);
    for (let n = 0; n < 256; n++) {
      let c = n;
      for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1;
      crcTable[n] = c;
    }
  }
  let c = ~0;
  for (let i = 0; i < buf.length; i++) c = crcTable[(c ^ buf[i]) & 0xff] ^ (c >>> 8);
  return ~c >>> 0;
}
function chunk(type, data) {
  const len = Buffer.alloc(4); len.writeUInt32BE(data.length, 0);
  const t = Buffer.from(type, 'ascii');
  const crc = Buffer.alloc(4); crc.writeUInt32BE(crc32(Buffer.concat([t, data])), 0);
  return Buffer.concat([len, t, data, crc]);
}
function encodePNG(w, h, rgba) {
  const sig = Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]);
  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(w, 0); ihdr.writeUInt32BE(h, 4);
  ihdr[8] = 8; ihdr[9] = 6; // 8-bit RGBA
  const raw = Buffer.alloc(h * (w * 4 + 1));
  for (let y = 0; y < h; y++) {
    raw[y * (w * 4 + 1)] = 0;
    rgba.copy(raw, y * (w * 4 + 1) + 1, y * w * 4, (y + 1) * w * 4);
  }
  return Buffer.concat([sig, chunk('IHDR', ihdr), chunk('IDAT', zlib.deflateSync(raw)), chunk('IEND', Buffer.alloc(0))]);
}

// ---------- 结构校验 ----------
const errors = [];
for (const [key, sp] of Object.entries(SPRITES)) {
  if (sp.rows.length !== 16) errors.push(`${key}: rows=${sp.rows.length}`);
  sp.rows.forEach((row, i) => {
    if (row.length !== 16) errors.push(`${key} row${i}: len=${row.length} "${row}"`);
    for (const ch of row) {
      if (ch !== '.' && !sp.pal[ch]) errors.push(`${key} row${i}: bad char '${ch}' in "${row}"`);
    }
  });
}
if (errors.length) {
  console.error('STRUCTURE ERRORS:');
  errors.forEach(e => console.error('  ' + e));
  process.exit(1);
}
console.log('structure OK:', Object.keys(SPRITES).length, 'sprites');

// ---------- 联系表 ----------
const SCALE = 8;
const TILE = 16 * SCALE;           // 128
const PAD = 24;
const LABEL = 34;                 // 底部留白（不画字，按位置对应）
const COLS = 4;
const names = Object.keys(SPRITES);
const ROWS = Math.ceil(names.length / COLS);
const W = COLS * (TILE + PAD) + PAD;
const H = ROWS * (TILE + LABEL + PAD) + PAD;
const px = Buffer.alloc(W * H * 4);
function put(x, y, r, g, b, a = 255) {
  if (x < 0 || y < 0 || x >= W || y >= H) return;
  const i = (y * W + x) * 4;
  px[i] = r; px[i + 1] = g; px[i + 2] = b; px[i + 3] = a;
}
function fillRect(x, y, w, h, col) {
  const [r, g, b] = col;
  for (let j = y; j < y + h; j++) for (let i = x; i < x + w; i++) put(i, j, r, g, b);
}
function hex(c) { return [parseInt(c.slice(1, 3), 16), parseInt(c.slice(3, 5), 16), parseInt(c.slice(5, 7), 16)]; }

fillRect(0, 0, W, H, [24, 25, 38]);
names.forEach((name, idx) => {
  const cx = idx % COLS, cy = Math.floor(idx / COLS);
  const ox = PAD + cx * (TILE + PAD), oy = PAD + cy * (TILE + LABEL + PAD);
  fillRect(ox - 4, oy - 4, TILE + 8, TILE + 8, [45, 47, 68]);
  const sp = SPRITES[name];
  for (let y = 0; y < 16; y++) for (let x = 0; x < 16; x++) {
    const ch = sp.rows[y][x];
    if (!ch || ch === '.') continue;
    const col = hex(sp.pal[ch]);
    fillRect(ox + x * SCALE, oy + y * SCALE, SCALE, SCALE, col);
  }
  // 调色板色块条（每格 8px）
  const keys = Object.keys(sp.pal).filter(k => k !== 'K' && sp.pal[k] !== '#131422');
  keys.slice(0, 10).forEach((k, i) => {
    fillRect(ox + i * 10, oy + TILE + 6, 8, 8, hex(sp.pal[k]));
  });
});
const out = path.join(__dirname, '..', 'tools', 'sheet.png');
fs.writeFileSync(out, encodePNG(W, H, px));
console.log('sheet:', out, W + 'x' + H);
names.forEach((n, i) => console.log(`  cell(${Math.floor(i / COLS)},${i % COLS}) = ${n}`));
