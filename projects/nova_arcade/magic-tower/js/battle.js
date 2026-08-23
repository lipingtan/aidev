/* ============================================================
 * 魔塔 — 战斗系统：预览 + 回合动画 + 胜负演出
 * MT.battle.start(mkey, { player, onEnd })
 *   player: 状态对象（含 hp/atk/def，战斗中直接扣减 hp）
 *   onEnd(result): 'win' | 'lose' | 'retreat'
 * ============================================================ */
(function (root, factory) {
  const m = factory();
  if (typeof module !== 'undefined' && module.exports) module.exports = m;
  root.MT = Object.assign(root.MT || {}, { battle: m });
})(typeof self !== 'undefined' ? self : globalThis, function () {

  const { SPRITES, makeSpriteCanvas } = MT;
  const sfx = MT.sfx;

  let skipFlag = false;
  let busy = false;

  function sleep(ms) {
    return new Promise((resolve) => {
      if (skipFlag) return resolve();
      const t0 = Date.now();
      const iv = setInterval(() => {
        if (skipFlag || Date.now() - t0 >= ms) { clearInterval(iv); resolve(); }
      }, 40);
    });
  }

  /* ---------- DOM（首次使用时构建） ---------- */
  let el = null; // { root, pFighter, eFighter, pCanvas, eCanvas, pBar, eBar, pHpTxt, eHpTxt, fx, banner }
  function build() {
    if (el) return el;
    const root = document.createElement('div');
    root.id = 'battle';
    root.innerHTML = `
      <div class="b-stage">
        <div class="fighter" id="bf-p">
          <div class="f-sprite"></div>
          <div class="f-plate">
            <div class="f-name">勇者</div>
            <div class="f-hpbar"><div class="f-hpfill"></div><span class="f-hptxt"></span></div>
            <div class="f-chips"><span class="chip atk"></span><span class="chip def"></span></div>
          </div>
        </div>
        <div class="b-vs">VS</div>
        <div class="fighter" id="bf-e">
          <div class="f-sprite"></div>
          <div class="f-plate">
            <div class="f-name"></div>
            <div class="f-hpbar"><div class="f-hpfill"></div><span class="f-hptxt"></span></div>
            <div class="f-chips"><span class="chip atk"></span><span class="chip def"></span></div>
          </div>
        </div>
      </div>
      <div class="b-fx"></div>
      <div class="b-banner"></div>
      <div class="b-hint">空格 / 点击：加速</div>`;
    document.body.appendChild(root);
    el = {
      root,
      pFighter: root.querySelector('#bf-p'),
      eFighter: root.querySelector('#bf-e'),
      fx: root.querySelector('.b-fx'),
      banner: root.querySelector('.b-banner'),
    };
    el.pSprite = el.pFighter.querySelector('.f-sprite');
    el.eSprite = el.eFighter.querySelector('.f-sprite');
    el.pBar = el.pFighter.querySelector('.f-hpfill');
    el.eBar = el.eFighter.querySelector('.f-hpfill');
    el.pHpTxt = el.pFighter.querySelector('.f-hptxt');
    el.eHpTxt = el.eFighter.querySelector('.f-hptxt');
    el.pAtk = el.pFighter.querySelector('.chip.atk');
    el.pDef = el.pFighter.querySelector('.chip.def');
    el.eName = el.eFighter.querySelector('.f-name');
    el.eAtk = el.eFighter.querySelector('.chip.atk');
    el.eDef = el.eFighter.querySelector('.chip.def');

    const onSkip = () => { skipFlag = true; };
    root.addEventListener('pointerdown', onSkip);
    window.addEventListener('keydown', (e) => { if (e.code === 'Space') onSkip(); });
    return el;
  }

  function setBar(fill, txt, cur, max) {
    fill.style.width = Math.max(0, Math.min(100, (cur / max) * 100)) + '%';
    txt.textContent = Math.max(0, cur) + ' / ' + max;
  }

  function floatText(targetEl, text, cls) {
    const d = build().fx;
    const r = targetEl.getBoundingClientRect();
    const b = d.getBoundingClientRect();
    const s = document.createElement('div');
    s.className = 'dmg-float ' + (cls || '');
    s.textContent = text;
    s.style.left = (r.left - b.left + r.width / 2 + (Math.random() * 40 - 20)) + 'px';
    s.style.top = (r.top - b.top + r.height * 0.3) + 'px';
    d.appendChild(s);
    setTimeout(() => s.remove(), 1000);
  }

  function sparks(targetEl, n, color) {
    const d = build().fx;
    const r = targetEl.getBoundingClientRect();
    const b = d.getBoundingClientRect();
    for (let i = 0; i < (n || 8); i++) {
      const s = document.createElement('div');
      s.className = 'spark';
      if (color) s.style.background = color;
      const ang = Math.random() * Math.PI * 2;
      const dist = 30 + Math.random() * 60;
      s.style.setProperty('--dx', Math.cos(ang) * dist + 'px');
      s.style.setProperty('--dy', Math.sin(ang * 1.3) * dist + 'px');
      s.style.left = (r.left - b.left + r.width / 2) + 'px';
      s.style.top = (r.top - b.top + r.height / 2) + 'px';
      d.appendChild(s);
      setTimeout(() => s.remove(), 700);
    }
  }

  function slash(targetEl, color) {
    const d = build().fx;
    const r = targetEl.getBoundingClientRect();
    const b = d.getBoundingClientRect();
    const s = document.createElement('div');
    s.className = 'slash-arc';
    if (color) s.style.borderColor = color;
    s.style.left = (r.left - b.left + r.width / 2 - 55) + 'px';
    s.style.top = (r.top - b.top + r.height / 2 - 55) + 'px';
    d.appendChild(s);
    setTimeout(() => s.remove(), 400);
  }

  function lunge(fighter, dir) { // dir: 1 向右冲, -1 向左冲
    fighter.style.transition = 'transform .16s ease-in';
    fighter.style.transform = `translateX(${dir * 64}px)`;
    setTimeout(() => {
      fighter.style.transition = 'transform .22s ease-out';
      fighter.style.transform = 'translateX(0)';
    }, 180);
  }

  function shakeScreen() {
    const root = build().root;
    root.classList.remove('shake');
    void root.offsetWidth;
    root.classList.add('shake');
  }

  function flash(targetEl) {
    targetEl.classList.remove('hitflash');
    void targetEl.offsetWidth;
    targetEl.classList.add('hitflash');
  }

  /* ---------- 预览对话框 ---------- */
  function showPreview(mkey, player, onFight, onRetreat) {
    const m = MT.MONSTERS[mkey];
    const pdmg = Math.max(0, player.atk - m.def);
    let verdict, vcls;
    if (pdmg <= 0) { verdict = '你的攻击无法破防 —— 无法战胜！'; vcls = 'bad'; }
    else {
      const rounds = Math.ceil(m.hp / pdmg);
      const loss = Math.max(0, m.atk - player.def) * (rounds - 1);
      if (loss >= player.hp) { verdict = `预计 ${rounds} 回合，损失 ${loss} HP —— 会死亡！`; vcls = 'bad'; }
      else { verdict = `预计 ${rounds} 回合，损失 ${loss} HP`; vcls = loss > player.hp * 0.5 ? 'warn' : 'ok'; }
    }
    const dlg = document.createElement('div');
    dlg.className = 'dlg-mask pv-mask';
    dlg.innerHTML = `
      <div class="dlg">
        <h3>遭遇 ${m.boss ? '★ ' : ''}${m.name}</h3>
        <div class="pv-row">
          <div class="pv-side">
            <div class="pv-sprite"></div>
            <div class="pv-name">勇者</div>
            <div class="pv-stat">HP ${player.hp}</div>
            <div class="pv-stat">攻 ${player.atk} · 防 ${player.def}</div>
          </div>
          <div class="pv-vs">→</div>
          <div class="pv-side">
            <div class="pv-sprite"></div>
            <div class="pv-name">${m.name}${m.boss ? '（Boss）' : ''}</div>
            <div class="pv-stat">HP ${m.hp}</div>
            <div class="pv-stat">攻 ${m.atk} · 防 ${m.def}</div>
          </div>
        </div>
        <p class="pv-verdict ${vcls}">${verdict}</p>
        <div class="dlg-btns">
          <button id="pv-fight" class="btn primary">开战</button>
          <button id="pv-retreat" class="btn">撤退</button>
        </div>
      </div>`;
    document.body.appendChild(dlg);
    const ps = dlg.querySelectorAll('.pv-sprite');
    ps[0].appendChild(makeSpriteCanvas(SPRITES.hero, 3)); // 32px×3=96
    ps[1].appendChild(makeSpriteCanvas(SPRITES[m.id], m.boss ? 4 : 3)); // 96 / Boss 128
    const fightBtn = dlg.querySelector('#pv-fight');
    const retreatBtn = dlg.querySelector('#pv-retreat');
    let closed = false;
    const close = (fn) => { if (closed) return; closed = true; window.removeEventListener('keydown', onKey); dlg.remove(); fn(); };
    fightBtn.onclick = () => close(onFight);
    retreatBtn.onclick = () => close(onRetreat);
    if (pdmg <= 0) fightBtn.disabled = true;
    function onKey(e) {
      if (e.code === 'Space') { e.preventDefault(); if (!fightBtn.disabled) close(onFight); }
      else if (e.key === 'Escape') close(onRetreat);
    }
    window.addEventListener('keydown', onKey);
  }

  /* ---------- 主流程 ---------- */
  async function start(mkey, player, onEnd) {
    if (busy) return;
    busy = true;
    skipFlag = false;
    const m = MT.MONSTERS[mkey];

    showPreview(mkey, player, () => runFight(), () => { busy = false; onEnd('retreat'); });

    async function runFight() {
      const D = build();
      D.root.classList.add('show');
      D.pSprite.innerHTML = '';
      D.eSprite.innerHTML = '';
      D.pSprite.appendChild(makeSpriteCanvas(SPRITES.hero, 4)); // 32px×4=128
      D.eSprite.appendChild(makeSpriteCanvas(SPRITES[m.id], m.boss ? 5 : 4)); // Boss 160 / 其余 128
      D.eName.textContent = (m.boss ? '★ ' : '') + m.name;
      D.pAtk.textContent = '攻 ' + player.atk;
      D.pDef.textContent = '防 ' + player.def;
      D.eAtk.textContent = '攻 ' + m.atk;
      D.eDef.textContent = '防 ' + m.def;
      const phpStart = player.hp;
      setBar(D.pBar, D.pHpTxt, player.hp, phpStart);
      let ehp = m.hp;
      setBar(D.eBar, D.eHpTxt, ehp, m.hp);

      sfx.play(m.boss ? 'boss' : 'battle');
      await sleep(700);

      const pdmg = Math.max(0, player.atk - m.def);
      const edmg = Math.max(0, m.atk - player.def);
      let result;

      while (true) {
        /* 玩家攻击 */
        lunge(D.pFighter, 1);
        await sleep(200);
        slash(D.eSprite, '#ffe9a8');
        sparks(D.eSprite, pdmg > 0 ? 10 : 4, pdmg > 0 ? '#ffd24a' : '#9aa3b8');
        sfx.play(pdmg > 0 ? 'hit' : 'block');
        await sleep(450);
        if (pdmg > 0) {
          ehp -= pdmg;
          flash(D.eSprite);
          floatText(D.eFighter, '-' + pdmg, 'dmg');
        } else {
          floatText(D.eFighter, '格挡', 'block');
        }
        setBar(D.eBar, D.eHpTxt, ehp, m.hp);
        await sleep(350);

        if (ehp <= 0) { result = 'win'; break; }

        /* 敌人攻击 */
        lunge(D.eFighter, -1);
        await sleep(200);
        slash(D.pSprite, '#ff8a7a');
        sparks(D.pSprite, edmg > 0 ? 10 : 4, edmg > 0 ? '#ff5c5c' : '#9aa3b8');
        sfx.play(edmg > 0 ? 'hit' : 'block');
        await sleep(450);
        if (edmg > 0) {
          player.hp -= edmg;
          flash(D.pSprite);
          shakeScreen();
          floatText(D.pFighter, '-' + edmg, 'dmg-enemy');
        } else {
          floatText(D.pFighter, '格挡', 'block');
        }
        setBar(D.pBar, D.pHpTxt, Math.max(0, player.hp), phpStart);
        await sleep(350);

        if (player.hp <= 0) { result = 'lose'; break; }
      }

      /* 胜负演出 */
      D.banner.textContent = result === 'win' ? '胜利！' : '你被击败了…';
      D.banner.className = 'b-banner show ' + (result === 'win' ? 'win' : 'lose');
      if (result === 'win') {
        sfx.play('victory');
        D.eSprite.classList.add('shatter');
        sparks(D.eSprite, 18, '#ffd24a');
        for (let i = 0; i < 12; i++) {
          const c = document.createElement('div');
          c.className = 'coin';
          c.style.left = (30 + Math.random() * 40) + '%';
          c.style.top = '30%';
          c.style.setProperty('--cx', (Math.random() * 200 - 100) + 'px');
          c.style.setProperty('--cy', (150 + Math.random() * 200) + 'px');
          c.style.animationDelay = (Math.random() * 0.3) + 's';
          D.fx.appendChild(c);
          setTimeout(() => c.remove(), 1600);
        }
      } else {
        sfx.play('defeat');
        D.pSprite.classList.add('shatter');
      }
      await sleep(result === 'win' ? 1400 : 1800);

      D.root.classList.remove('show');
      D.eSprite.classList.remove('shatter');
      D.pSprite.classList.remove('shatter');
      D.banner.className = 'b-banner';
      busy = false;
      onEnd(result);
    }
  }

  return { start, get busy() { return busy; } };
});
