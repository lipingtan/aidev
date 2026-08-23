/* ============================================================
 * 魔塔 — WebAudio 合成音效（无外部资源）
 * MT.sfx.play(name) / MT.sfx.muted / MT.sfx.toggle()
 * ============================================================ */
(function (root, factory) {
  const m = factory();
  if (typeof module !== 'undefined' && module.exports) module.exports = m;
  root.MT = Object.assign(root.MT || {}, { sfx: m });
})(typeof self !== 'undefined' ? self : globalThis, function () {

  let ctx = null;
  let muted = false;

  function ac() {
    if (muted) return null;
    try {
      if (!ctx) ctx = new (window.AudioContext || window.webkitAudioContext)();
      if (ctx.state === 'suspended') ctx.resume();
      return ctx;
    } catch (e) { return null; }
  }

  function tone(freq, dur, type, vol, when = 0, slide = 0) {
    const c = ac();
    if (!c) return;
    const t0 = c.currentTime + when;
    const o = c.createOscillator();
    const g = c.createGain();
    o.type = type || 'square';
    o.frequency.setValueAtTime(freq, t0);
    if (slide) o.frequency.exponentialRampToValueAtTime(Math.max(30, freq + slide), t0 + dur);
    g.gain.setValueAtTime(vol || 0.08, t0);
    g.gain.exponentialRampToValueAtTime(0.001, t0 + dur);
    o.connect(g).connect(c.destination);
    o.start(t0);
    o.stop(t0 + dur + 0.02);
  }

  const SOUNDS = {
    step:    () => tone(220, 0.05, 'triangle', 0.03),
    hit:     () => { tone(180, 0.12, 'square', 0.1, 0, -120); tone(90, 0.18, 'sawtooth', 0.06, 0.02, -40); },
    block:   () => { tone(500, 0.06, 'triangle', 0.07); tone(380, 0.08, 'triangle', 0.05, 0.05); },
    pickup:  () => { tone(523, 0.08, 'sine', 0.09); tone(659, 0.08, 'sine', 0.09, 0.07); tone(784, 0.12, 'sine', 0.09, 0.14); },
    gold:    () => { tone(880, 0.06, 'square', 0.06); tone(1175, 0.1, 'square', 0.06, 0.05); },
    door:    () => { tone(120, 0.2, 'triangle', 0.12, 0, -60); tone(80, 0.25, 'sawtooth', 0.06, 0.05, -30); },
    stairs:  () => { tone(392, 0.1, 'triangle', 0.08); tone(523, 0.14, 'triangle', 0.08, 0.09); },
    battle:  () => { tone(330, 0.1, 'square', 0.08); tone(262, 0.14, 'square', 0.08, 0.1); },
    victory: () => { [523, 659, 784, 1047].forEach((f, i) => tone(f, 0.16, 'triangle', 0.1, i * 0.1)); },
    defeat:  () => { [392, 330, 262, 196].forEach((f, i) => tone(f, 0.22, 'sawtooth', 0.07, i * 0.16)); },
    boss:    () => { [110, 138, 110, 165].forEach((f, i) => tone(f, 0.2, 'sawtooth', 0.09, i * 0.14, -20)); },
  };

  function play(name) {
    const fn = SOUNDS[name];
    if (fn) { try { fn(); } catch (e) { /* ignore */ } }
  }

  return { play, get muted() { return muted; }, toggle() { muted = !muted; return muted; } };
});
