<script setup lang="ts">
/**
 * 首页 — 产品展示页
 */
import { ref } from 'vue'
import { useUserStore } from '@/stores/user'

const userStore = useUserStore()

// 产品数据
const products = ref([
  {
    id: 'rag-agent',
    icon: '🤖',
    badge: '热门',
    badgeType: 'hot',
    title: 'RAG Agent 智能客服',
    subtitle: '新一代企业级 AI 智能客服解决方案',
    description: '基于大语言模型与检索增强生成（RAG）技术，为企业提供 7×24 小时智能客户服务。精准理解用户意图，智能检索企业知识库，生成专业、准确的回复，显著提升客户满意度与服务效率。',
    features: [
      { icon: '🎯', text: '意图识别准确率 95%+' },
      { icon: '📚', text: '私有知识库智能检索' },
      { icon: '🔄', text: '多轮对话上下文理解' },
      { icon: '📊', text: '实时数据分析看板' },
      { icon: '🔗', text: '无缝对接主流客服系统' },
      { icon: '🛡️', text: '企业级数据安全保障' },
    ],
    highlights: [
      { value: '95%+', label: '问题解决率' },
      { value: '80%', label: '人工成本降低' },
      { value: '<1s', label: '平均响应时间' },
    ],
    gradient: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
  },
  {
    id: 'ai-delivery',
    icon: '⚡',
    badge: '新品',
    badgeType: 'new',
    title: 'AI 赋能软件交付',
    subtitle: '智能化软件开发与交付解决方案',
    description: '将 AI 技术深度融入软件开发全生命周期，从需求分析、架构设计、代码生成到测试部署，全面提升研发效能。智能代码助手辅助开发，自动化测试保障质量，让软件交付更快、更稳、更省心。',
    features: [
      { icon: '💡', text: 'AI 需求分析与拆解' },
      { icon: '🏗️', text: '智能架构设计推荐' },
      { icon: '✨', text: 'AI 代码生成与补全' },
      { icon: '🔍', text: '智能代码审查' },
      { icon: '🧪', text: '自动化测试用例生成' },
      { icon: '🚀', text: 'CI/CD 流水线自动化' },
    ],
    highlights: [
      { value: '60%', label: '开发效率提升' },
      { value: '40%', label: '缺陷率降低' },
      { value: '2x', label: '交付速度' },
    ],
    gradient: 'linear-gradient(135deg, #11998e 0%, #38ef7d 100%)',
  },
  {
    id: 'game-arcade',
    icon: '🎮',
    badge: '经典',
    badgeType: 'classic',
    title: '街机游戏厅',
    subtitle: '重温经典街机黄金时代',
    description: '汇聚数千款经典街机游戏，完美复刻原版体验。支持云端存档、多人联机、自定义按键，让您随时随地重温童年记忆。专业模拟器内核，画面音效完美还原，带您穿越回那个热血沸腾的街机年代。',
    features: [
      { icon: '🕹️', text: '3000+ 经典街机游戏' },
      { icon: '☁️', text: '云端存档随时续玩' },
      { icon: '👥', text: '本地/在线多人联机' },
      { icon: '🎨', text: '高清滤镜画面增强' },
      { icon: '⚡', text: '极速加载零等待' },
      { icon: '🔧', text: '自定义按键映射' },
    ],
    highlights: [
      { value: '3000+', label: '经典游戏' },
      { value: '100万+', label: '活跃玩家' },
      { value: '4.9', label: '用户评分' },
    ],
    gradient: 'linear-gradient(135deg, #f093fb 0%, #f5576c 100%)',
  },
])

// 联系方式
const contacts = ref([
  { icon: '📧', label: '商务合作', value: 'business@btcola.com' },
  { icon: '💬', label: '技术支持', value: 'support@btcola.com' },
  { icon: '📱', label: '客服热线', value: '400-888-9999' },
])
</script>

<template>
  <div class="home-page">
    <!-- Hero 区域 -->
    <section class="hero-section">
      <div class="hero-content">
        <img src="@/assets/logo.png" alt="千分软件" class="hero-logo" />
        <h1 class="hero-title">
          <span class="hero-subtitle-text">智能科技，赋能未来</span>
        </h1>
        <p class="hero-desc">
          专注于人工智能与互联网应用开发，致力于为企业和个人用户提供优质的软件产品与技术服务
        </p>
        <div class="hero-stats">
          <div class="stat-item">
            <span class="stat-value">7+</span>
            <span class="stat-label">年行业经验</span>
          </div>
          <div class="stat-item">
            <span class="stat-value">100万+</span>
            <span class="stat-label">服务用户</span>
          </div>
          <div class="stat-item">
            <span class="stat-value">99.9%</span>
            <span class="stat-label">服务可用性</span>
          </div>
        </div>
      </div>
    </section>

    <!-- 产品展示区 -->
    <section class="products-section">
      <div class="section-header">
        <h2 class="section-title">核心产品</h2>
        <p class="section-desc">为您提供专业、可靠的软件解决方案</p>
      </div>

      <div class="products-grid">
        <div 
          v-for="product in products" 
          :key="product.id" 
          class="product-card"
        >
          <!-- 卡片头部 -->
          <div class="product-header" :style="{ background: product.gradient }">
            <span class="product-icon">{{ product.icon }}</span>
            <span 
              class="product-badge" 
              :class="'badge-' + product.badgeType"
            >
              {{ product.badge }}
            </span>
          </div>

          <!-- 卡片内容 -->
          <div class="product-body">
            <h3 class="product-title">{{ product.title }}</h3>
            <p class="product-subtitle">{{ product.subtitle }}</p>
            <p class="product-desc">{{ product.description }}</p>

            <!-- 功能特性 -->
            <div class="product-features">
              <div 
                v-for="(feature, idx) in product.features" 
                :key="idx" 
                class="feature-item"
              >
                <span class="feature-icon">{{ feature.icon }}</span>
                <span class="feature-text">{{ feature.text }}</span>
              </div>
            </div>

            <!-- 数据亮点 -->
            <div class="product-highlights">
              <div 
                v-for="(hl, idx) in product.highlights" 
                :key="idx" 
                class="highlight-item"
              >
                <span class="highlight-value">{{ hl.value }}</span>
                <span class="highlight-label">{{ hl.label }}</span>
              </div>
            </div>

            <!-- 操作按钮 -->
            <div class="product-actions">
              <button class="btn-primary">了解更多</button>
              <button class="btn-secondary">免费试用</button>
            </div>
          </div>
        </div>
      </div>
    </section>

    <!-- 技术优势 -->
    <section class="advantage-section">
      <div class="section-header">
        <h2 class="section-title">为什么选择我们</h2>
        <p class="section-desc">专业团队，匠心打造每一款产品</p>
      </div>

      <div class="advantage-grid">
        <div class="advantage-item">
          <div class="advantage-icon">🚀</div>
          <h4 class="advantage-title">技术领先</h4>
          <p class="advantage-desc">紧跟前沿技术发展，将 AI 大模型、云计算等先进技术融入产品</p>
        </div>
        <div class="advantage-item">
          <div class="advantage-icon">🛡️</div>
          <h4 class="advantage-title">安全可靠</h4>
          <p class="advantage-desc">通过多项安全认证，数据加密存储，保障用户隐私与数据安全</p>
        </div>
        <div class="advantage-item">
          <div class="advantage-icon">⚡</div>
          <h4 class="advantage-title">极致体验</h4>
          <p class="advantage-desc">以用户为中心的产品设计理念，持续优化，追求极致用户体验</p>
        </div>
        <div class="advantage-item">
          <div class="advantage-icon">🤝</div>
          <h4 class="advantage-title">专业服务</h4>
          <p class="advantage-desc">7×24 小时技术支持，专属客户经理，快速响应您的每一个需求</p>
        </div>
      </div>
    </section>

    <!-- 联系我们 -->
    <section class="contact-section">
      <div class="section-header">
        <h2 class="section-title">联系我们</h2>
        <p class="section-desc">期待与您的合作</p>
      </div>

      <div class="contact-grid">
        <div 
          v-for="(contact, idx) in contacts" 
          :key="idx" 
          class="contact-item"
        >
          <span class="contact-icon">{{ contact.icon }}</span>
          <span class="contact-label">{{ contact.label }}</span>
          <span class="contact-value">{{ contact.value }}</span>
        </div>
      </div>
    </section>
  </div>
</template>

<style scoped>
/* ── 页面容器 ── */
.home-page {
  min-height: 100%;
  background: linear-gradient(180deg, #f8fafc 0%, #f1f5f9 100%);
}

/* ── Hero 区域 ── */
.hero-section {
  padding: 48px 24px 40px;
  text-align: center;
  background: linear-gradient(135deg, #1a1d2e 0%, #2d3148 100%);
  color: #fff;
}

.hero-content {
  max-width: 800px;
  margin: 0 auto;
}

.hero-logo {
  height: 70px;
  width: auto;
  margin-bottom: 20px;
}

.hero-title {
  font-size: 32px;
  font-weight: 800;
  line-height: 1.3;
  margin: 0 0 16px;
}

.hero-subtitle-text {
  font-size: 20px;
  font-weight: 400;
  opacity: 0.9;
}

.hero-desc {
  font-size: 15px;
  color: rgba(255,255,255,0.7);
  line-height: 1.7;
  margin: 0 0 32px;
  max-width: 600px;
  margin-left: auto;
  margin-right: auto;
}

.hero-stats {
  display: flex;
  justify-content: center;
  gap: 40px;
  flex-wrap: wrap;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.stat-value {
  font-size: 28px;
  font-weight: 700;
  color: #a5b4fc;
}

.stat-label {
  font-size: 13px;
  color: rgba(255,255,255,0.6);
}

/* ── 通用区块样式 ── */
.section-header {
  text-align: center;
  margin-bottom: 32px;
}

.section-title {
  font-size: 24px;
  font-weight: 700;
  color: #1a1d2e;
  margin: 0 0 8px;
}

.section-desc {
  font-size: 14px;
  color: #64748b;
  margin: 0;
}

/* ── 产品展示区 ── */
.products-section {
  padding: 48px 24px;
}

.products-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(340px, 1fr));
  gap: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.product-card {
  background: #fff;
  border-radius: 20px;
  overflow: hidden;
  box-shadow: 0 4px 24px rgba(0,0,0,0.08);
  transition: transform 0.3s, box-shadow 0.3s;
}

.product-card:hover {
  transform: translateY(-4px);
  box-shadow: 0 12px 40px rgba(0,0,0,0.12);
}

.product-header {
  padding: 32px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  position: relative;
}

.product-icon {
  font-size: 48px;
  filter: drop-shadow(0 4px 8px rgba(0,0,0,0.2));
}

.product-badge {
  padding: 6px 14px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  color: #fff;
  background: rgba(255,255,255,0.25);
  backdrop-filter: blur(8px);
}

.badge-hot {
  background: rgba(255,107,107,0.9);
}

.badge-new {
  background: rgba(56,239,125,0.9);
  color: #1a1d2e;
}

.badge-classic {
  background: rgba(255,193,7,0.9);
  color: #1a1d2e;
}

.product-body {
  padding: 24px;
}

.product-title {
  font-size: 22px;
  font-weight: 700;
  color: #1a1d2e;
  margin: 0 0 6px;
}

.product-subtitle {
  font-size: 14px;
  color: #667eea;
  font-weight: 500;
  margin: 0 0 16px;
}

.product-desc {
  font-size: 14px;
  color: #64748b;
  line-height: 1.7;
  margin: 0 0 20px;
}

/* 功能特性 */
.product-features {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  margin-bottom: 20px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: #f8fafc;
  border-radius: 10px;
  font-size: 13px;
  color: #475569;
}

.feature-icon {
  font-size: 14px;
}

.feature-text {
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* 数据亮点 */
.product-highlights {
  display: flex;
  justify-content: space-around;
  padding: 16px 0;
  border-top: 1px solid #f1f5f9;
  border-bottom: 1px solid #f1f5f9;
  margin-bottom: 20px;
}

.highlight-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}

.highlight-value {
  font-size: 20px;
  font-weight: 700;
  color: #1a1d2e;
}

.highlight-label {
  font-size: 12px;
  color: #94a3b8;
}

/* 操作按钮 */
.product-actions {
  display: flex;
  gap: 12px;
}

.btn-primary {
  flex: 1;
  height: 44px;
  border: none;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  cursor: pointer;
  transition: opacity 0.2s, transform 0.2s;
}

.btn-primary:hover {
  opacity: 0.9;
  transform: translateY(-1px);
}

.btn-secondary {
  flex: 1;
  height: 44px;
  border: 1.5px solid #e2e8f0;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 600;
  color: #64748b;
  background: #fff;
  cursor: pointer;
  transition: border-color 0.2s, color 0.2s;
}

.btn-secondary:hover {
  border-color: #667eea;
  color: #667eea;
}

/* ── 技术优势 ── */
.advantage-section {
  padding: 48px 24px;
  background: #fff;
}

.advantage-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(240px, 1fr));
  gap: 24px;
  max-width: 1000px;
  margin: 0 auto;
}

.advantage-item {
  text-align: center;
  padding: 24px 20px;
  border-radius: 16px;
  background: #f8fafc;
  transition: background 0.2s;
}

.advantage-item:hover {
  background: #f1f5f9;
}

.advantage-icon {
  font-size: 36px;
  margin-bottom: 12px;
}

.advantage-title {
  font-size: 16px;
  font-weight: 600;
  color: #1a1d2e;
  margin: 0 0 8px;
}

.advantage-desc {
  font-size: 13px;
  color: #64748b;
  line-height: 1.6;
  margin: 0;
}

/* ── 联系我们 ── */
.contact-section {
  padding: 48px 24px;
}

.contact-grid {
  display: flex;
  justify-content: center;
  gap: 32px;
  flex-wrap: wrap;
  max-width: 800px;
  margin: 0 auto;
}

.contact-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 20px 32px;
  background: #fff;
  border-radius: 16px;
  box-shadow: 0 2px 12px rgba(0,0,0,0.06);
  min-width: 200px;
}

.contact-icon {
  font-size: 28px;
}

.contact-label {
  font-size: 13px;
  color: #94a3b8;
}

.contact-value {
  font-size: 14px;
  font-weight: 600;
  color: #1a1d2e;
}

/* ── 响应式 ── */
@media (max-width: 768px) {
  .hero-section {
    padding: 32px 16px;
  }

  .hero-title {
    font-size: 26px;
  }

  .hero-subtitle-text {
    font-size: 16px;
  }

  .hero-stats {
    gap: 24px;
  }

  .stat-value {
    font-size: 22px;
  }

  .products-section,
  .advantage-section,
  .contact-section {
    padding: 32px 16px;
  }

  .products-grid {
    grid-template-columns: 1fr;
  }

  .product-features {
    grid-template-columns: 1fr;
  }

  .section-title {
    font-size: 20px;
  }

  .contact-grid {
    flex-direction: column;
    align-items: center;
  }

  .contact-item {
    width: 100%;
    max-width: 300px;
  }
}

@media (max-width: 480px) {
  .hero-title {
    font-size: 22px;
  }

  .product-header {
    padding: 24px 20px;
  }

  .product-icon {
    font-size: 40px;
  }

  .product-body {
    padding: 20px;
  }

  .product-title {
    font-size: 18px;
  }

  .product-actions {
    flex-direction: column;
  }

  .btn-primary,
  .btn-secondary {
    width: 100%;
  }
}
</style>
