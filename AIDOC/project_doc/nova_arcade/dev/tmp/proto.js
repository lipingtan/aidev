
/* ================= data =================
   col  = neon 基色   colE = elegant 基色
   t2   = elegant 主题别名（氛围文案）      */
const GAMES=[
 {id:'tetra_nova',t:'TETRA NOVA',st:'新星方阵 · 变异消除',ic:'🧊',col:'#1a4a8f',colE:'#a8c9b8',cat:'puzzle',tags:['消除','Roguelike','离线'],pm:'trial',price:6,trial:3,rating:4.8,rc:23000,players:120000,ver:'1.0.0',dlc:false,alias:['俄罗斯方块','tetris','els'],py:['eluosifangkuai','elsfk'],desc:'Roguelike 变异俄罗斯方块：16 种可叠加变异、Boss 战、双子系统。SRS 手感 + 程序合成音频，全部离线可玩。'},
 {id:'neon2048',t:'2048 霓虹版',t2:'2048 青瓷版',st:'经典数字合成的双面演绎',ic:'🔢',col:'#2a6a4f',colE:'#c2d8a8',cat:'puzzle',tags:['数字','离线'],pm:'free',price:0,rating:4.6,rc:8900,players:45000,ver:'1.0.0',dlc:false,alias:['2048'],py:['erlingsibalian'],desc:'滑动合并数字，冲击 2048 与更高。随主题切换的釉色皮肤 + 无限模式。'},
 {id:'nova_snake',t:'霓虹贪吃蛇',t2:'竹影贪吃蛇',st:'经典蛇 · 传送门与障碍',ic:'🐍',col:'#4f1a6a',colE:'#bcb4d8',cat:'arcade',tags:['街机','离线'],pm:'ad',price:0,rating:4.4,rc:5400,players:30000,ver:'1.0.0',dlc:false,alias:['蛇','snake','贪吃蛇'],py:['tanchishe','tcs'],desc:'贪吃蛇重制：传送门、加速带、障碍关卡，支持激励视频复活。'},
 {id:'brick_reactor',t:'打砖块·反应堆',t2:'打砖块·陶窑',st:'物理弹球 清屏爽快',ic:'🧱',col:'#6a4a1a',colE:'#e0c49e',cat:'arcade',tags:['弹球','物理'],pm:'trial',price:3,trial:5,rating:4.5,rc:3200,players:18000,ver:'1.0.0',dlc:false,alias:['打砖块','brick'],py:['dazhuankuai','dzk'],desc:'角度决定一切的打砖块，连锁爆破与 Boss 砖阵。'},
 {id:'mine_sweeper_x',t:'扫雷·星际',t2:'扫雷·晴岚',st:'经典扫雷 双主题皮肤',ic:'💣',col:'#1a5a6a',colE:'#a4c6d4',cat:'puzzle',tags:['益智','离线'],pm:'free',price:0,rating:4.7,rc:6100,players:26000,ver:'1.0.0',dlc:false,alias:['扫雷','mine'],py:['saolei','sl'],desc:'零猜测盘面保证，每日挑战与经典难度，全部离线。'},
 {id:'bullet_waltz',t:'弹幕圆舞曲',st:'一人弹幕 五分钟心跳',ic:'🌀',col:'#6a1a3a',colE:'#d4a8ba',cat:'action',tags:['弹幕','硬核'],pm:'paid',price:6,rating:4.3,rc:1800,players:9000,ver:'1.0.0',dlc:true,alias:['弹幕','danmaku'],py:['danmuyuanwuqu','dmywq'],desc:'极简一人弹幕：五种弹幕编排的 Boss 轮舞，硬核向。'},
 {id:'idle_forge',t:'放置锻造屋',st:'睡前一键 收获神装',ic:'⚒️',col:'#5a5a1a',colE:'#d4c69c',cat:'casual',tags:['放置','挂机'],pm:'iap',price:0,rating:4.2,rc:4400,players:38000,ver:'1.0.0',dlc:true,alias:['放置','idle'],py:['fangzhiduzaowu','fzdz'],desc:'离线收益自动结算的锻造屋，收集 200+ 神装图鉴。'},
 {id:'dungeon_crawl_n',t:'新星地牢',t2:'新月地牢',st:'回合地牢 一局十分钟',ic:'🏰',col:'#3a1a6a',colE:'#b0aad0',cat:'roguelike',tags:['Roguelike','回合','离线'],pm:'paid',price:12,rating:4.9,rc:2900,players:11000,ver:'1.0.0',dlc:true,alias:['地牢','rogue'],py:['xinxindilao','xxdl'],desc:'程序生成地牢 + Build 组合，一局十分钟的手机 Roguelike。'},
 {id:'pixel_jump',t:'像素跳跳',t2:'苔原跳跃',st:'单键平台跳跃',ic:'🦘',col:'#1a6a5a',colE:'#9ed4c6',cat:'action',tags:['平台','离线'],pm:'free',price:0,rating:4.1,rc:7200,players:52000,ver:'1.0.0',dlc:false,alias:['跳跃','jump'],py:['xiangsutiaotiao','xstt'],desc:'一个键玩到底的平台跳跃：每日关卡 + 全球点赞关卡。'},
 {id:'match_nova',t:'新星消消',t2:'初雪消消',st:'三消 连击不停',ic:'✨',col:'#6a2a1a',colE:'#e0b0a4',cat:'puzzle',tags:['三消','离线'],pm:'trial',price:3,trial:5,rating:4.5,rc:9800,players:61000,ver:'1.0.0',dlc:false,alias:['消消乐','三消','match'],py:['xinxingxiaoxiao','xxxx'],desc:'无限连击的三消：步数模式 / 限时模式 / Boss 模式。'},
];
const CATS=[['all','全部'],['puzzle','消除益智'],['arcade','街机'],['action','动作'],['casual','休闲'],['roguelike','Roguelike']];
const REVIEWS={
 tetra_nova:[
  {nm:'星尘旅人',st:5,pt:'12.4 小时',dt:'2 天前',lk:328,tx:'变异叠加到后期真的会失控，黑洞+坍缩连锁一屏清空，爽感拉满。Boss 推垃圾行的压力刚刚好。'},
  {nm:'方块老炮',st:4,pt:'3.1 小时',dt:'5 天前',lk:156,tx:'手感是标准 SRS，锁定延迟可以微调这点好评。差评位给音效，BGM 后期有点吵（可关）。'},
  {nm:'NOVA玩家_1024',st:5,pt:'26 小时',dt:'1 周前',lk:98,tx:'第二个 Boss 之后每局的 build 都不一样，这是我手机里活最久的一个"俄罗斯方块"。'}],
 neon2048:[{nm:'数字搬运工',st:5,pt:'8 小时',dt:'3 天前',lk:77,tx:'皮肤随主题切换很惊喜，无限模式上头。'}],
};
const S={owned:new Set(['neon2048','mine_sweeper_x']),trialLeft:{tetra_nova:2,brick_reactor:5,match_nova:5},
 installed:new Set(GAMES.filter(g=>!g.dlc).map(g=>g.id)),records:{tetra_nova:{last:Date.now()-36e5,best:18220,pt:11520,finishes:9},
 neon2048:{last:Date.now()-86400e3,best:4096,pt:28800,finishes:14},mine_sweeper_x:{last:Date.now()-172800e3,best:0,pt:600,finishes:1}},
 hist:[],payChannel:'模拟支付',pendingOrder:null,myRev:{},curG:null,revStars:0};
const $=s=>document.querySelector(s),$$=s=>document.querySelectorAll(s);
const fmt=n=>n>=1e4?(n/1e4).toFixed(1)+'万':n;
const g=id=>GAMES.find(x=>x.id===id);
const PMNAME={free:'免费',ad:'免费·含广告',trial:'试玩',paid:'买断',iap:'免费·内购'};

/* ================= theme system ================= */
const THEMES={
 neon:{iconEnd:'#0a1030',label:'霓虹 · 星穹',bgOut:'#020308'},
 elegant:{iconEnd:'#f4f1ea',label:'青瓷 · 素笺',bgOut:'#eae6dd'},
};
let THEME='neon';
function saveTheme(t){try{window.localStorage.setItem('nova_theme',t)}catch(e){}}
function loadTheme(){try{return window.localStorage.getItem('nova_theme')||'neon'}catch(e){return 'neon'}}
function getTheme(){return THEME}
function curCol(gm){return THEME==='elegant'?(gm.colE||gm.col):gm.col}
function gt(gm){return THEME==='elegant'&&gm.t2?gm.t2:gm.t}
function grad(gm,soft){const c=curCol(gm);return `linear-gradient(150deg,${soft?c+'55':c},${THEMES[THEME].iconEnd})`}
function setTheme(t){
  if(!THEMES[t])return;
  THEME=t;
  document.getElementById('frame').dataset.theme=t;
  document.getElementById('logoSub').textContent=THEMES[t].label;
  if(document.body)document.body.style.background=THEMES[t].bgOut;
  const a=document.getElementById('thNeon'),b=document.getElementById('thEleg');
  if(a&&b){a.classList.toggle('on',t==='neon');b.classList.toggle('on',t==='elegant')}
  saveTheme(t);
  rerenderAll();
}
function toggleTheme(){
  const next=THEME==='neon'?'elegant':'neon';
  setTheme(next);toast('已切换外观：'+THEMES[next].label);
}
function rerenderAll(){
  renderHome();renderCharts(chartMode);renderCat();renderSearchHome();renderLib(libMode);
  if(S.curG){try{renderDetail(S.curG)}catch(e){}}
}

/* ================= router ================= */
let curTab='home',stack=[];
function switchTab(i){
  const t=$$('.tab')[i];curTab=t.dataset.pg;
  $$('.tab').forEach(x=>x.classList.remove('on'));t.classList.add('on');
  stack=[];showPage(curTab);
}
function go(pg,data){stack.push(curTab);curTab=pg;showPage(pg,data)}
function back(){curTab=stack.pop()||'home';showPage(curTab);
  const ti=['home','category','search','library'].indexOf(curTab);
  if(ti>=0){const t=$$('.tab')[ti];$$('.tab').forEach(x=>x.classList.remove('on'));t.classList.add('on')}}
function showPage(pg,data){
  $$('.page').forEach(p=>p.classList.remove('show'));
  const el=$('#pg-'+pg);el.classList.add('show');el.scrollTop=0;
  if(pg==='detail')renderDetail(data||S.curG);
  if(pg==='reviews')renderRevs(revSort);
}
window.addEventListener('keydown',e=>{if(e.key==='Backspace'&&document.activeElement!==$('#searchInput'))back()});

/* ================= render: cards ================= */
function badgesHtml(gm){
  let h='';
  if(gm.pm==='paid')h+=`<span class="badge paid">¥${gm.price}</span>`;
  else if(gm.pm==='trial'&&!S.owned.has(gm.id))h+=`<span class="badge trial">试玩</span>`;
  else if(gm.pm==='ad')h+=`<span class="badge off">AD</span>`;
  if(gm.dlc&&!S.installed.has(gm.id))h=`<span class="badge dl">⬇</span>`;
  return h;
}
function cardS(id){
  const gm=g(id);
  return `<div class="cardS" onclick="openDetail('${id}')">
    <div class="ic" style="background:${grad(gm)}">${gm.ic}${badgesHtml(gm)}</div>
    <div class="nm">${gt(gm)}</div>
    <div class="meta"><span class="star">★</span>${gm.rating}<span style="opacity:.5">·</span>${PMNAME[gm.pm]}</div></div>`;
}
function lrow(id,i){
  const gm=g(id);
  return `<div class="lrow r${i+1}" onclick="openDetail('${id}')">
    <div class="rank">${i+1}</div>
    <div class="ic" style="background:${grad(gm)}">${gm.ic}</div>
    <div class="mid"><div class="t">${gt(gm)}</div>
      <div class="s"><span class="star">★</span>${gm.rating} · ${fmt(gm.rc)}评价 · ${fmt(gm.players)}人玩过</div></div>
    <span class="go">›</span></div>`;
}
let chartMode='all',libMode='owned',revSort='useful',curCat='all';
function renderHome(){
  const cont=Object.keys(S.records).sort((a,b)=>S.records[b].last-S.records[a].last).slice(0,5);
  $('#rowContinue').innerHTML=cont.map(cardS).join('');
  const myTags=new Set();cont.forEach(id=>g(id)&&g(id).tags.forEach(t=>myTags.add(t)));
  const fy=GAMES.filter(x=>!S.owned.has(x.id)).filter(x=>{
    if(x.tags.some(t=>myTags.has(t)))return true;return x.rating>=4.5})
    .slice(0,8);
  $('#rowForYou').innerHTML=fy.map(x=>cardS(x.id)).join('')||GAMES.slice(0,6).map(x=>cardS(x.id)).join('');
}
function renderCharts(mode,el){
  chartMode=mode;
  if(el){$$('#chartTabs .chip').forEach(x=>x.classList.remove('on'));el.classList.add('on')}
  let list=[...GAMES];
  if(mode==='new')list.sort((a,b)=>b.ver.localeCompare(a.ver)||b.players-a.players);
  else if(mode==='good')list.sort((a,b)=>b.rating*b.rc-a.rating*a.rc);
  else list.sort((a,b)=>b.players-a.players);
  $('#chartList').innerHTML=list.slice(0,8).map((x,i)=>lrow(x.id,i)).join('');
}
let curFilters=new Set(['all']);
function renderCat(){
  $('#catSide').innerHTML=CATS.map(([k,n])=>{
    const cnt=k==='all'?GAMES.length:GAMES.filter(x=>x.cat===k).length;
    return `<div class="catItem ${k===curCat?'on':''}" onclick="curCat='${k}';renderCat()">${n}<div style="font-size:9px;opacity:.55;margin-top:2px">${cnt}</div></div>`}).join('');
  const list=GAMES.filter(x=>curCat==='all'||x.cat===curCat).filter(x=>{
    if(curFilters.has('free')&&!(x.pm==='free'||x.pm==='ad'||x.pm==='iap'))return false;
    if(curFilters.has('paid')&&!(x.pm==='paid'||x.pm==='trial'))return false;
    if(curFilters.has('offline')&&!x.tags.includes('离线'))return false;
    if(curFilters.has('r40')&&x.rating<4.0)return false;
    return true;});
  $('#catGrid').innerHTML=list.map(x=>`<div class="cardM" onclick="openDetail('${x.id}')">
    <div class="ic" style="background:${grad(x)}">${x.ic}${badgesHtml(x)}</div>
    <div class="nm">${gt(x)}</div>
    <div class="meta"><span class="star">★</span>${x.rating} · ${PMNAME[x.pm]}</div></div>`).join('')
    ||`<div style="grid-column:1/3" class="empty"><div class="e">🕳</div><p>该筛选下暂无游戏</p><button class="minibtn" onclick="curFilters=new Set(['all']);renderCat()">清除筛选</button></div>`;
}
$$('#catFilters .chip').forEach(c=>c.onclick=()=>{
  const f=c.dataset.f;
  if(f==='all'){curFilters=new Set(['all'])}else{curFilters.delete('all');curFilters.has(f)?curFilters.delete(f):curFilters.add(f)}
  $$('#catFilters .chip').forEach(x=>x.classList.toggle('on',curFilters.has(x.dataset.f)));
  renderCat();});
const HOT=['俄罗斯方块','2048','roguelike','消消乐','离线','弹幕'];
function renderSearchHome(){
  $('#hotTags').innerHTML=HOT.map(h=>`<span class="histag hot" onclick="doSearch('${h}')">${h}</span>`).join('');
  $('#histTags').innerHTML=S.hist.length?S.hist.map(h=>`<span class="histag" onclick="doSearch('${h}')">${h}</span>`).join('')
    :`<span style="font-size:11px;color:var(--ink3)">暂无搜索历史</span>`;
}
function clearHist(){S.hist=[];renderSearchHome()}
function clearSearch(){$('#searchInput').value='';$('#searchHome').style.display='';$('#searchResults').style.display='none';$('#searchEmpty').style.display='none'}
let deb;
$('#searchInput').addEventListener('input',e=>{
  clearTimeout(deb);deb=setTimeout(()=>doSearch(e.target.value,true),150);});
function doSearch(q,fromInput){
  if(!fromInput){$('#searchInput').value=q}
  q=q.trim().toLowerCase();if(!q){clearSearch();return}
  if(!S.hist.includes(q)){S.hist.unshift(q);S.hist=S.hist.slice(0,10)}
  renderSearchHome();
  const scored=[];
  for(const gm of GAMES){let sc=0;
    if(gm.t.toLowerCase().startsWith(q))sc+=100;
    else if(gm.t.toLowerCase().includes(q))sc+=40;
    if(gm.t2&&gm.t2.toLowerCase().includes(q))sc+=35;
    if(gm.alias.some(a=>a.toLowerCase().includes(q)))sc+=35;
    if(gm.py.some(p=>p.startsWith(q)))sc+=30;
    if(gm.tags.some(t=>t.toLowerCase().includes(q)))sc+=15;
    if(sc>0)scored.push([sc,gm]);}
  scored.sort((a,b)=>b[0]-a[0]);
  const res=scored.slice(0,20).map(([,gm],i)=>lrow(gm.id,i)).join('');
  $('#searchHome').style.display='none';
  if(res){$('#searchEmpty').style.display='none';$('#searchResults').style.display='';$('#searchResults').innerHTML=res}
  else{$('#searchResults').style.display='none';$('#searchEmpty').style.display='';
    $('#searchGuess').innerHTML=GAMES.slice().sort((a,b)=>b.rating-a.rating).slice(0,6).map(x=>cardS(x.id)).join('')}
}
function renderLib(mode,el){
  libMode=mode;
  if(el){$$('#libTabs .chip').forEach(x=>x.classList.remove('on'));el.classList.add('on')}
  let ids;
  if(mode==='owned')ids=[...S.owned];
  else if(mode==='trial')ids=Object.keys(S.trialLeft).filter(id=>!S.owned.has(id)&&S.trialLeft[id]>0);
  else ids=Object.keys(S.records).sort((a,b)=>S.records[b].last-S.records[a].last);
  $('#libList').innerHTML=ids.map(id=>{const gm=g(id),r=S.records[id]||{};
    const sub=mode==='trial'?`试玩剩余 ${S.trialLeft[id]} 次`:r.pt?`最高 ${fmt(r.best)} · ${(r.pt/3600).toFixed(1)} 小时`:PMNAME[gm.pm];
    return `<div class="librow" onclick="openDetail('${id}')">
      <div class="ic" style="background:${grad(gm,1)}">${gm.ic}</div>
      <div class="t"><b>${gt(gm)}</b><span>${sub}</span></div><span class="arr">›</span></div>`}).join('')
    ||`<div class="empty"><div class="e">📦</div><p>这里还空空如也<br>去首页逛逛吧</p><button class="minibtn" onclick="switchTab(0)">去首页</button></div>`;
}
function ctaState(id){
  const gm=g(id);
  if(gm.dlc&&!S.installed.has(id))return{cls:'dl',txt:`⬇ 下载 · ${(gm.players%40+8)/10} MB`,hint:'DLC 下载后即可开玩（Wi-Fi 优先）'};
  if(S.owned.has(id)||gm.pm==='free'||gm.pm==='ad'||gm.pm==='iap')
    return{cls:'play',txt:'▶ 开玩',hint:S.records[id]?`上次最高 ${fmt(S.records[id].best)} · 继续`:'首次开玩'};
  if(gm.pm==='trial'){
    const left=S.trialLeft[id]??gm.trial;
    if(left>0)return{cls:'trial',txt:`▶ 试玩 · 剩 ${left} 次`,hint:`试玩结束后 ¥${gm.price} 解锁全部`};
    return{cls:'buy',txt:`¥${gm.price} 解锁完整版`,hint:'试玩次数已用完'};}
  return{cls:'buy',txt:`¥${gm.price} 购买`,hint:'买断制 · 一次购买永久拥有'};
}
function openDetail(id){S.curG=id;go('detail')}
function renderDetail(id){
  const gm=g(id);
  $('#shots').innerHTML=[gm.ic,'🎮','⚡','🏆'].map(e=>`<div class="shot">${e}</div>`).join('');
  $('#dIcon').textContent=gm.ic;$('#dIcon').style.background=grad(gm);
  $('#dTitle').textContent=gt(gm);$('#dSub').textContent=`${gm.st} · v${gm.ver}`;
  $('#dScore').innerHTML=`<b>${gm.rating}</b><span class="star">★</span><span>(${fmt(gm.rc)} 评价)</span><span style="opacity:.4">|</span><span>${fmt(gm.players)} 人玩过</span>`;
  $('#tagRow').innerHTML=gm.tags.map(t=>`<span class="tag">${t}</span>`).join('')+`<span class="tag dim">${PMNAME[gm.pm]}</span>`+(gm.dlc?'<span class="tag dim">DLC</span>':'');
  $('#dDescBody').textContent=gm.desc;
  const dist=[72,18,6,2,2];$('#rateBars').innerHTML=dist.map((p,i)=>
    `<div class="rbar"><span>${5-i}</span><div class="tr"><i style="width:${p}%"></i></div></div>`).join('');
  const revs=REVIEWS[id]||[];
  $('#dRevs').innerHTML=revs.slice(0,2).map(revHtml).join('')||
    `<div class="empty"><div class="e">✍️</div><p>还没有评价<br>来抢第一个沙发</p></div>`;
  const sim=GAMES.filter(x=>x.id!==id&&(x.cat===gm.cat||x.tags.some(t=>gm.tags.includes(t))))
    .sort((a,b)=>b.players-a.players).slice(0,4);
  $('#dSimilar').innerHTML=sim.map(x=>cardS(x.id)).join('');
  refreshCta();
}
function refreshCta(){
  const id=S.curG,st=ctaState(id);
  const btn=$('#ctaBtn');btn.className='cta '+st.cls;btn.textContent=st.txt;
  $('#ctaHint').textContent=st.hint;
}
function ctaClick(){
  const id=S.curG,gm=g(id),st=ctaState(id);
  if(st.cls==='dl'){openDownload(id);return}
  if(st.cls==='buy'){openPurchase(id);return}
  if(st.cls==='trial'){S.trialLeft[id]--;toast(`试玩已消耗 · 剩 ${S.trialLeft[id]} 次`);refreshCta()}
  const sc=Math.floor(800+Math.random()*14000);
  $('#resGame').textContent=gt(gm);$('#resScore').textContent=sc.toLocaleString();
  const r=S.records[id]||{best:0,pt:0};
  $('#resBest').textContent=`历史最高 ${Math.max(r.best,sc).toLocaleString()} · 本局 ${Math.floor(Math.random()*6+2)} 分 ${Math.floor(Math.random()*50+10)} 秒`;
  $('#resAch').style.display=Math.random()>.5?'':'none';
  const sim=GAMES.filter(x=>x.id!==id&&(x.cat===gm.cat||x.tags.some(t=>gm.tags.includes(t)))).slice(0,3);
  $('#resMore').innerHTML=sim.map(x=>`<div class="mini" onclick="closeOv('ov-result');openDetail('${x.id}')">
    <div class="ic" style="background:${grad(x)}">${x.ic}</div>${gt(x)}</div>`).join('');
  setTimeout(()=>openOv('ov-result'),400);
  r.best=Math.max(r.best,sc);r.pt+=200+r.best%400;r.last=Date.now();r.finishes=(r.finishes||0)+1;S.records[id]=r;
  renderHome();
}
function openPurchase(id){
  const gm=g(id);
  $('#purGame').innerHTML=`<div class="ic" style="background:${grad(gm)}">${gm.ic}</div>
    <div><b>${gt(gm)}</b><span>¥${gm.price} · 买断制</span></div>`;
  $('#purPrice').textContent=gm.price;
  $('#purBody').style.display='';$('#purPaying').style.display='none';
  openOv('ov-purchase');
}
function pickPay(el){$$('#payCh .opt').forEach(x=>x.classList.remove('on'));el.classList.add('on')}
function mockPay(){
  $('#purBody').style.display='none';$('#purPaying').style.display='';
  setTimeout(()=>{
    const id=S.curG;S.owned.add(id);delete S.trialLeft[id];
    closeOv('ov-purchase');toast('支付成功 · 已解锁 ▸');refreshCta();renderLib('owned');
    setTimeout(()=>toastAch('🏆 成就解锁：收藏家 I +10 点'),900);
  },900);
}
function openDownload(id){
  const gm=g(id);
  $('#dlGame').innerHTML=`<div class="ic" style="width:46px;height:46px;border-radius:13px;display:flex;align-items:center;justify-content:center;font-size:23px;background:${grad(gm)}">${gm.ic}</div><div style="font-size:13px;color:var(--ink)">${gt(gm)}<div style="font-size:10px;color:var(--ink3);margin-top:2px">扩展包 · 校验 SHA-256</div></div>`;
  openOv('ov-download');
  let p=0;const bar=$('#dlBar i'),pct=$('#dlPct');
  const tm=setInterval(()=>{p=Math.min(100,p+Math.random()*16);
    bar.style.width=p+'%';pct.textContent=`下载中 ${p|0}% · 预计 ${Math.ceil((100-p)/22)}s`;
    if(p>=100){clearInterval(tm);pct.textContent='✓ 安装完成';S.installed.add(id);
      setTimeout(()=>{closeOv('ov-download');toast('已安装 · 可开玩');refreshCta()},500)}},220);
}
function revHtml(r,mine){
  return `<div class="rev ${mine?'myrev':''}">
    <div class="hd"><div class="av">👾</div><span class="nm">${r.nm}</span>
      ${r.pt?`<span class="pt">游玩 ${r.pt}</span>`:''}<span class="stars">${'★'.repeat(r.st)}<span class="off">${'★'.repeat(5-r.st)}</span></span></div>
    <p>${r.tx||'<i style="color:var(--ink3)">（纯星级评价）</i>'}</p>
    <div class="ft"><span>👍 ${r.lk||0}</span><span>${r.dt}</span>${mine?`<span style="color:var(--gold)">我的</span>`:'<span>举报</span>'}
      ${mine?`<span style="color:var(--accent);cursor:pointer" onclick="openReview()">修改</span>`:''}</div>
    ${mine?`<div class="edit"><button class="minibtn" onclick="openReview()">修改</button><button class="minibtn" onclick="delMyRev()">删除</button></div>`:''}</div>`;
}
function renderRevs(sort,el){
  revSort=sort;
  if(el){$$('#revSort .chip').forEach(x=>x.classList.remove('on'));el.classList.add('on')}
  const id=S.curG,list=[...(REVIEWS[id]||[])];
  const mine=S.myRev[id];
  if(mine)list.unshift({...mine,nm:'NOVA玩家_8472',dt:'刚刚',lk:0,mine:true});
  if(sort==='good')list.sort((a,b)=>b.st-a.st);
  else if(sort==='bad')list.sort((a,b)=>a.st-b.st);
  else if(sort==='new');else list.sort((a,b)=>(b.lk||0)-(a.lk||0));
  $('#revList').innerHTML=list.map(r=>revHtml(r,r.mine)).join('')
    ||`<div class="empty"><div class="e">✍️</div><p>还没有评价<br>游玩 10 分钟后即可评价</p><button class="minibtn gold" onclick="openReview()">✎ 写评价</button></div>`;
}
function delMyRev(){delete S.myRev[S.curG];renderRevs();toast('评价已删除')}
function openReview(){
  const id=S.curG,gm=g(id);$('#revGameNm').textContent='· '+gt(gm);
  const pt=(S.records[id]?.pt)||0;
  const gate=$('#revGate'),btn=$('#revSubmit');
  if(pt<600){gate.textContent=`⏳ 需游玩满 10 分钟才能评价（当前 ${(pt/60).toFixed(0)} 分钟）`;btn.disabled=true}
  else{gate.textContent='';btn.disabled=false}
  const ex=S.myRev[id];$('#revText').value=ex?ex.tx:'';$('#revCount').textContent=`${ex?ex.tx.length:0} / 500`;
  S.revStars=ex?ex.st:0;paintStars(S.revStars);
  openOv('ov-review');
}
function paintStars(n){$$('#starPick span').forEach((s,i)=>s.classList.toggle('on',i<n))}
$$('#starPick span').forEach((s,i)=>s.onclick=()=>{S.revStars=i+1;paintStars(S.revStars)});
$('#revText').addEventListener('input',e=>$('#revCount').textContent=`${e.target.value.length} / 500`);
function submitReview(){
  if(!S.revStars){toast('请先选择星级');return}
  S.myRev[S.curG]={st:S.revStars,tx:$('#revText').value.trim(),pt:(((S.records[S.curG]?.pt)||0)/3600).toFixed(1)+' 小时'};
  closeOv('ov-review');renderRevs();toast('评价已发布（本地）');
  setTimeout(()=>toastAch('🏆 成就解锁：发声者 +10 点'),800);
}
function openOv(id){$('#'+id).classList.add('show')}
function closeOv(id){$('#'+id).classList.remove('show')}
let toastTm;
function toast(msg,ach){const t=$('#toast');t.textContent=msg;t.className=ach?'show ach':'show';
  clearTimeout(toastTm);toastTm=setTimeout(()=>t.classList.remove('show'),2200)}
function toastAch(msg){toast(msg,true)}

/* ================= init ================= */
THEME=loadTheme();
document.getElementById('frame').dataset.theme=THEME;
document.getElementById('logoSub').textContent=THEMES[THEME].label;
if(document.body)document.body.style.background=THEMES[THEME].bgOut;
const _thn=document.getElementById('thNeon'),_the=document.getElementById('thEleg');
if(_thn&&_the){_thn.classList.toggle('on',THEME==='neon');_the.classList.toggle('on',THEME==='elegant')}
renderHome();renderCharts('all');renderCat();renderSearchHome();renderLib('owned');
let bd=0;setInterval(()=>{bd=(bd+1)%3;$$('#banner .dots i').forEach((d,i)=>{d.className=i===bd?'on':'';
  d.style.width=i===bd?'14px':'5px'})},4000);
