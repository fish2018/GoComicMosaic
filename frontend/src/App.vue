<template>
  <div class="app-container">
    <header class="app-header">
      <div class="container header-inner">
        <div class="brand">
          <router-link to="/" class="brand-link">
            <i class="bi bi-collection-play brand-icon"></i>
            <span class="brand-text">{{ siteInfo.logoText }}</span>
          </router-link>
        </div>
        
        <!-- 添加全局搜索组件 -->
        <LocalSearch class="header-search" />
        
        <div class="header-actions">
          <template v-if="isLoggedIn">
            <div class="user-greeting">
              <i class="bi bi-person-circle"></i>
              <span>{{ currentUser.username }}</span>
            </div>
            <div class="button-group">
              <router-link v-if="!isAdminPage" to="/admin" class="btn-custom btn-info">
                <i class="bi bi-gear-fill me-1"></i><span class="btn-text">管理后台</span>
              </router-link>
              <button @click="handleLogout" class="btn-custom btn-outline">
                <i class="bi bi-box-arrow-right me-1"></i><span class="btn-text">登出</span>
              </button>
              <router-link to="/submit" class="btn-custom btn-primary">
                <i class="bi bi-plus-circle me-1"></i><span class="btn-text">提交资源</span>
              </router-link>
            </div>
          </template>
          <template v-else>
            <div class="button-group">
              <router-link to="/login" class="btn-custom btn-outline" aria-label="管理员登录">
                <i class="bi bi-shield-lock me-1"></i><span class="btn-text">管理员登录</span>
              </router-link>
              <router-link to="/submit" class="btn-custom btn-primary" aria-label="提交资源">
                <i class="bi bi-plus-circle me-1"></i><span class="btn-text">提交资源</span>
              </router-link>
            </div>
          </template>
        </div>
      </div>
    </header>
    
    <main class="main-content">
      <div class="container content-container">
        <router-view />
      </div>
    </main>
    
    <!-- 隐藏的预加载元素 -->
    <div class="prefetch-footer" aria-hidden="true">
      <div class="footer-content">
        <div class="footer-brand"></div>
        <div class="footer-links"></div>
      </div>
    </div>
    
    <footer class="app-footer">
      <div class="container footer-inner">
        <!-- 页脚布局 -->
        <div class="footer-row">
          <template v-if="footerSettings">
            <!-- 动态生成链接 -->
            <template v-for="(link, index) in footerSettings.links" :key="index">
              <!-- 内部链接 -->
              <router-link v-if="link.type === 'internal'" :to="link.url" class="footer-link" :title="link.title">
                <i v-if="link.icon" :class="link.icon"></i>
                <span>{{ link.text }}</span>
              </router-link>
              
              <!-- 外部链接 -->
              <a v-else :href="link.url" target="_blank" class="footer-link" :title="link.title">
                <i v-if="link.icon" :class="link.icon"></i>
                <span v-if="!link.icon">{{ link.text }}</span>
              </a>
            </template>
            
            <!-- 访问统计 -->
            <span class="footer-link" v-if="footerSettings.show_visitor_count">总访问量 <span id="busuanzi_value_site_pv">0</span></span>
          </template>
          
          <!-- 在设置加载前的默认链接，或加载失败时的回退链接 -->
          <template v-else>
            <router-link to="/about" class="footer-link">关于我们</router-link>
            <a href="https://t.me/xueximeng" target="_blank" class="footer-link" title="加入Telegram群组">
              <i class="bi bi-telegram"></i>
            </a>
            <a href="https://github.com/fish2018/GoComicMosaic" target="_blank" class="footer-link" title="查看GitHub源码">
              <i class="bi bi-github"></i>
            </a>
          </template>
        </div>
        
        <!-- 分隔线 -->
        <div class="footer-divider"></div>
        
        <!-- 版权信息 -->
        <div class="copyright">
          <p>{{ footerSettings?.copyright || '&copy; 2025 美漫资源共建. 保留所有权利' }}</p>
        </div>
      </div>
    </footer>

    <!-- 全局悬浮按钮 -->
    <div class="floating-buttons">
      <router-link to="/" class="floating-btn home-btn" title="返回主页">
        <i class="bi bi-house-fill"></i>
      </router-link>
      <button @click="scrollToTop" class="floating-btn top-btn" title="回到顶部">
        <i class="bi bi-chevron-up"></i>
      </button>
      <button @click="scrollToBottom" class="floating-btn bottom-btn" title="回到底部">
        <i class="bi bi-chevron-down"></i>
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, nextTick, onUnmounted } from 'vue'
import { isAuthenticated, getCurrentUser, logout, setupAxiosInterceptors } from './utils/auth'
import { useRoute, useRouter } from 'vue-router'
import LocalSearch from './components/LocalSearch.vue'
import axios from 'axios'
import { getSiteSettings } from './utils/api'

const route = useRoute()
const router = useRouter()
const isLoggedIn = ref(false)
const currentUser = ref({})
const footerPreloaded = ref(false)
const footerSettings = ref(null)
const siteInfo = ref({
  title: '美漫资源共建',
  logoText: '美漫资源共建',
  description: '美漫共建平台是一个开源的美漫资源共享网站，用户可以自由提交动漫信息，像马赛克一样，由多方贡献拼凑成完整资源。',
  keywords: '美漫, 动漫资源, 资源共享, 开源平台, 美漫共建'
})

// 计算当前是否在管理员页面
const isAdminPage = computed(() => {
  return route.path.startsWith('/admin')
})

const checkAuthState = () => {
  isLoggedIn.value = isAuthenticated()
  if (isLoggedIn.value) {
    currentUser.value = getCurrentUser() || {}
  }
}

const handleLogout = () => {
  logout()
  checkAuthState()
}

// 获取页脚设置
const loadFooterSettings = async () => {
  try {
    // 使用InfoManager获取缓存的信息
    const infoManager = (await import('./utils/InfoManager')).default;
    footerSettings.value = await infoManager.getFooterInfo();
    console.log('页脚设置加载成功:', footerSettings.value);
  } catch (error) {
    console.error('获取页脚设置失败:', error);
    // 使用默认设置
    footerSettings.value = {
      links: [
        { text: "关于我们", url: "/about", type: "internal" },
        { text: "Telegram", url: "https://t.me/xueximeng", icon: "bi bi-telegram", type: "external", title: "加入Telegram群组" },
        { text: "GitHub", url: "https://github.com/fish2018/GoComicMosaic", icon: "bi bi-github", type: "external", title: "查看GitHub源码" },
        { text: "在线点播", url: "/streams", type: "internal" },
        { text: "漫迪小站", url: "https://mdsub.top/", type: "external" },
        { text: "三次元成瘾者康复中心", url: "https://www.kangfuzhongx.in/", type: "external" },
      ],
      copyright: "© 2025 美漫资源共建. 保留所有权利",
      show_visitor_count: true
    };
  }
}

// 加载网站基本信息
const loadSiteInfo = async () => {
  try {
    const infoManager = (await import('./utils/InfoManager')).default;
    const info = await infoManager.getSiteBasicInfo();
    siteInfo.value = info;
    console.log('网站基本信息加载成功:', siteInfo.value);
    
    // 更新页面标题和meta信息
    updateMetaInfo(route);
  } catch (error) {
    console.error('获取网站基本信息失败:', error);
    // 默认值已在siteInfo的ref初始化中设置
  }
}

// 滚动到页面顶部
const scrollToTop = () => {
  window.scrollTo({
    top: 0,
    behavior: 'smooth'
  });
}

// 滚动到页面底部
const scrollToBottom = () => {
  // 确保底部元素已预加载
  if (!footerPreloaded.value) {
    preloadFooterContent();
  }
  
  // 使用平滑滚动到底部
  window.scrollTo({
    top: document.documentElement.scrollHeight,
    behavior: 'smooth'
  });
}

// 预加载页脚内容以防止渲染撕裂
const preloadFooterContent = () => {
  if (footerPreloaded.value) return;
  
  // 触发页脚资源预加载
  const footer = document.querySelector('.app-footer');
  if (footer) {
    // 强制重新计算布局
    footer.getBoundingClientRect();
    footerPreloaded.value = true;
  }
}

// 优化滚动性能的函数
const optimizeScrollPerformance = () => {
  let scrollTimeout;
  
  window.addEventListener('scroll', () => {
    // 滚动时检测是否接近底部
    const scrollPosition = window.scrollY + window.innerHeight;
    const documentHeight = document.documentElement.scrollHeight;
    
    // 如果接近底部，预加载页脚内容
    if (documentHeight - scrollPosition < 300 && !footerPreloaded.value) {
      preloadFooterContent();
    }
    
    // 滚动节流处理
    if (!scrollTimeout) {
      scrollTimeout = setTimeout(() => {
        scrollTimeout = null;
      }, 100);
    }
  }, { passive: true });
}

// 添加清理storage的函数
const clearPaginationStorage = () => {
  localStorage.removeItem('resourcesCurrentPage')
  localStorage.removeItem('resourcesPageSize')
  // 如果还有其他需要清理的分页相关数据，可以在这里添加
}

// 处理滚动事件，显示/隐藏回到顶部按钮
const handleScroll = () => {
  // 滚动时检测是否接近底部
  const scrollPosition = window.scrollY + window.innerHeight;
  const documentHeight = document.documentElement.scrollHeight;
  
  // 如果接近底部，预加载页脚内容
  if (documentHeight - scrollPosition < 300 && !footerPreloaded.value) {
    preloadFooterContent();
  }
}

// 更新页面标题和meta信息的函数
const updateMetaInfo = (to) => {
  // 设置默认值
  const defaultTitle = siteInfo.value.title;
  const defaultDescription = siteInfo.value.description;
  const defaultKeywords = siteInfo.value.keywords;
  
  // 获取路由的meta信息
  const title = to.meta.title || defaultTitle;
  const description = to.meta.description || defaultDescription;
  const keywords = to.meta.keywords || defaultKeywords;
  
  // 更新页面标题
  document.title = title;
  
  // 更新meta描述
  let metaDescription = document.querySelector('meta[name="description"]');
  if (metaDescription) {
    metaDescription.setAttribute('content', description);
  }
  
  // 更新meta关键词
  let metaKeywords = document.querySelector('meta[name="keywords"]');
  if (metaKeywords) {
    metaKeywords.setAttribute('content', keywords);
  }
  
  // 更新Open Graph标签
  let ogTitle = document.querySelector('meta[property="og:title"]');
  if (ogTitle) {
    ogTitle.setAttribute('content', title);
  }
  
  let ogDescription = document.querySelector('meta[property="og:description"]');
  if (ogDescription) {
    ogDescription.setAttribute('content', description);
  }

  // 检查并更新favicon
  if (typeof window.checkFavicon === 'function') {
    window.checkFavicon();
  }
}

// 使用afterEach钩子监听路由变化
onMounted(() => {
  // 设置路由afterEach钩子
  router.afterEach((to) => {
    // 更新meta信息
    updateMetaInfo(to);
    
    // 回到页面顶部（可选）
    // window.scrollTo(0, 0)
  });
  
  checkAuthState();
  
  // 初始加载时设置meta信息
  updateMetaInfo(route);
  
  // 设置axios拦截器
  setupAxiosInterceptors(() => {
    logout();
    isLoggedIn.value = false;
  });
  
  // 加载页脚设置和网站基本信息
  loadFooterSettings();
  loadSiteInfo();
  
  // 监听滚动事件
  window.addEventListener('scroll', handleScroll, { passive: true });
  
  // 优化滚动性能
  optimizeScrollPerformance();
  
  // 确保初始渲染后预加载底部元素
  nextTick(() => {
    setTimeout(preloadFooterContent, 1000);
  });
  
  // 添加beforeunload事件监听器
  window.addEventListener('beforeunload', clearPaginationStorage);

  // 添加不蒜子访问统计脚本
  const bszScript = document.createElement('script');
  bszScript.async = true;
  bszScript.src = "//busuanzi.ibruce.info/busuanzi/2.3/busuanzi.pure.mini.js";
  document.head.appendChild(bszScript);
})

// 页面卸载时移除事件监听器
onUnmounted(() => {
  window.removeEventListener('scroll', handleScroll);
  window.removeEventListener('beforeunload', clearPaginationStorage);
});
</script>

<style>
/* 全局样式重置和基础样式 */
html, body {
  height: 100%;
  margin: 0;
  padding: 0;
  scroll-behavior: smooth; /* 添加平滑滚动 */
}

body {
  font-family: 'Poppins', 'Helvetica Neue', Arial, sans-serif;
  line-height: 1.6;
  color: var(--dark-color);
  background-color: var(--light-color);
  background-image: 
    radial-gradient(circle at 10% 10%, rgba(124, 58, 237, 0.06) 0%, transparent 500px),
    radial-gradient(circle at 90% 30%, rgba(6, 182, 212, 0.06) 0%, transparent 700px),
    radial-gradient(circle at 50% 80%, rgba(244, 63, 94, 0.04) 0%, transparent 600px);
  background-attachment: fixed;
  /* 优化渲染性能 */
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
}

/* 设置颜色变量 */
:root {
  /* 主色调 - 炫彩紫蓝系 */
  --primary-color: #7c3aed;
  --primary-gradient: linear-gradient(145deg, #8b5cf6, #6366f1);
  --primary-hover: linear-gradient(145deg, #7c3aed, #4f46e5);
  
  /* 辅色调 - 鲜艳青绿系 */
  --secondary-color: #06b6d4;
  --secondary-gradient: linear-gradient(145deg, #0ea5e9, #0284c7);
  --secondary-hover: linear-gradient(145deg, #0891b2, #0369a1);
  
  /* 强调色 - 明亮橙粉系 */
  --accent-color: #f43f5e;
  --accent-gradient: linear-gradient(145deg, #f97316, #f43f5e);
  
  /* 中性色 */
  --light-color: #f0f4ff;
  --dark-color: #0f172a;
  --gray-color: #64748b;
  
  /* 功能色 */
  --success-color: #10b981;
  --info-color: #06b6d4;
  --warning-color: #f59e0b;
  --danger-color: #ef4444;
  
  /* 样式变量 */
  --border-radius: 14px;
  --card-radius: 18px;
  --box-shadow: 0 10px 25px -3px rgba(0, 0, 0, 0.15), 0 4px 10px -4px rgba(0, 0, 0, 0.1);
  --deep-shadow: 0 25px 30px -5px rgba(0, 0, 0, 0.15), 0 10px 15px -5px rgba(0, 0, 0, 0.1);
  --inner-shadow: inset 0 2px 4px 0 rgba(0, 0, 0, 0.05);
  --transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  --transition-bounce: all 0.5s cubic-bezier(0.34, 1.56, 0.64, 1);
  --glass-background: rgba(255, 255, 255, 0.6);
  --glass-border: 1px solid rgba(255, 255, 255, 0.8);
  --glass-blur: blur(12px);
}

.container {
  width: 100%;
  max-width: 1800px;
  margin: 0 auto;
  padding: 0 1rem;
  box-sizing: border-box;
}

.header-inner, .footer-inner {
  width: 100%;
  box-sizing: border-box;
}

/* 应用容器 */
.app-container {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  /* 优化移动设备上渲染性能 */
  -webkit-overflow-scrolling: touch;
  overflow-x: hidden; /* 防止水平滚动条 */
}

/* 头部导航样式 */
.app-header {
  background-color: var(--glass-background);
  backdrop-filter: var(--glass-blur);
  -webkit-backdrop-filter: var(--glass-blur);
  box-shadow: 0 8px 25px rgba(0, 0, 0, 0.06);
  position: sticky;
  top: 0;
  z-index: 100;
  border-bottom: var(--glass-border);
  border-radius: 0 0 var(--border-radius) var(--border-radius);
  margin: 0 1rem;
  padding: 0.75rem 0;
}

.header-inner {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 0.5rem 0;
  gap: 1.5rem;
  flex-wrap: wrap;
  width: 100%;
}

.brand {
  display: flex;
  align-items: center;
  flex-shrink: 0;
  flex: 0 0 auto;
  padding-left: 0.25rem; /* 默认添加一点左内边距 */
}

.brand-link {
  display: flex;
  align-items: center;
  text-decoration: none;
  color: var(--primary-color);
  font-weight: 700;
  font-size: 1.5rem;
  transition: var(--transition);
  max-width: 230px;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.brand-link:hover {
  color: var(--secondary-color);
  transform: translateY(-2px);
}

.brand-icon {
  margin-left: 1.5rem;
  font-size: 1.8rem;
  margin-right: 0.5rem;
  position: relative;
  z-index: 1;
}

.brand-icon::before {
  content: "";
  position: absolute;
  inset: -6px;
  background: var(--primary-gradient);
  border-radius: 50%;
  z-index: -1;
  opacity: 0.9;
  box-shadow: 0 4px 15px rgba(124, 58, 237, 0.35);
}

.brand-text {
  position: relative;
  font-weight: 800;
  letter-spacing: -0.5px;
  background: var(--primary-gradient);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  text-shadow: 3px 3px 5px rgba(124, 58, 237, 0.2);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.header-actions {
  display: flex;
  align-items: center;
  gap: 1.5rem;
  flex-shrink: 0;
  flex: 0 0 auto;
  padding-right: 1.25rem; /* 默认添加一点右内边距 */
}

.user-greeting {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  color: var(--dark-color);
  font-weight: 500;
  padding: 0.5rem 1.25rem;
  background-color: rgba(255, 255, 255, 0.5);
  border-radius: var(--border-radius);
  box-shadow: var(--inner-shadow);
  border: var(--glass-border);
  backdrop-filter: var(--glass-blur);
  -webkit-backdrop-filter: var(--glass-blur);
}

.button-group {
  display: flex;
  gap: 0.75rem;
  padding: 0.5rem 0;
}

.btn-custom {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 0.8rem 1.5rem;
  border-radius: var(--border-radius);
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition-bounce);
  text-decoration: none;
  border: none;
  font-size: 0.95rem;
  position: relative;
  overflow: hidden;
  box-shadow: 0 6px 15px rgba(0, 0, 0, 0.1);
  margin: 0.25rem 0;
}

.btn-custom::before {
  content: '';
  position: absolute;
  top: 0;
  left: -170%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
  transition: 0.5s;
}

.btn-custom:hover::before {
  left: 100%;
}

.btn-primary {
  background: var(--primary-gradient);
  color: white;
  box-shadow: 
    0 8px 20px rgba(124, 58, 237, 0.35),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
}

.btn-primary:hover {
  box-shadow: 
    0 12px 25px rgba(124, 58, 237, 0.5),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
  transform: translateY(-5px);
}

.btn-info {
  background: var(--secondary-gradient);
  color: white;
  box-shadow: 
    0 8px 20px rgba(6, 182, 212, 0.35),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
}

.btn-info:hover {
  box-shadow: 
    0 12px 25px rgba(6, 182, 212, 0.5),
    inset 0 1px 0 rgba(255, 255, 255, 0.3);
  transform: translateY(-5px);
}

.btn-outline {
  background: var(--glass-background);
  backdrop-filter: var(--glass-blur);
  -webkit-backdrop-filter: var(--glass-blur);
  border: 1px solid rgba(124, 58, 237, 0.3);
  color: var(--primary-color);
  box-shadow: 
    0 6px 15px rgba(124, 58, 237, 0.15),
    inset 0 1px 0 rgba(255, 255, 255, 0.7);
}

.btn-outline:hover {
  background: var(--primary-gradient);
  color: white;
  box-shadow: 
    0 8px 20px rgba(124, 58, 237, 0.35),
    inset 0 1px 0 rgba(255, 255, 255, 0.2);
  transform: translateY(-5px);
  border-color: transparent;
}

/* 主内容区域 */
.main-content {
  flex: 1;
  padding: 2rem 0;
}

.content-container {
  animation: fadeInUpScale 0.7s cubic-bezier(0.175, 0.885, 0.32, 1.275);
  width: 100%;
}

/* 页脚样式 */
.app-footer {
  background: var(--glass-background);
  backdrop-filter: var(--glass-blur);
  -webkit-backdrop-filter: var(--glass-blur);
  color: var(--dark-color);
  padding: 1.8rem 0 1.2rem; /* 增加上下内边距 */
  margin: 2rem 1rem 1rem; /* 减少上边距 */
  position: relative;
  overflow: hidden;
  box-shadow: 
    0 -8px 25px rgba(0, 0, 0, 0.06), 
    0 -2px 10px rgba(124, 58, 237, 0.04),
    0 -15px 40px rgba(0, 0, 0, 0.03); /* 增强阴影效果 */
  border: var(--glass-border);
  border-radius: var(--border-radius);
  will-change: transform;
  transform: translateZ(0);
  -webkit-backface-visibility: hidden;
  backface-visibility: hidden;
}

/* 添加顶部光晕效果增强立体感 */
.app-footer::before {
  content: "";
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  height: 2px;
  background: linear-gradient(90deg, 
    rgba(255, 255, 255, 0.05), 
    rgba(255, 255, 255, 0.8), 
    rgba(255, 255, 255, 0.05)
  );
  opacity: 0.6;
  z-index: 1;
}

/* 简化页脚内部容器样式 */
.footer-inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 0.7rem; /* 增加行间距 */
  padding: 0 0.8rem; /* 增加左右内边距 */
}

/* 页脚一行式布局 */
.footer-row {
  display: flex;
  justify-content: center; /* 居中显示所有元素 */
  align-items: center;
  flex-wrap: wrap;
  width: 100%;
  padding: 0.6rem 0; /* 增加内边距 */
  gap: 0.6rem; /* 元素之间的间距 */
}

/* 统一链接风格 */
.footer-link {
  color: var(--secondary-text);
  text-decoration: none;
  padding: 0.4rem 0.8rem;
  border-radius: 100px;
  font-size: 0.85rem;
  transition: all 0.2s ease;
  background-color: rgba(255, 255, 255, 0.3);
  /* border: 1px solid rgba(255, 255, 255, 0.5); //增强边框可见度 */
  white-space: nowrap;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  /* box-shadow: 0 2px 5px rgba(0, 0, 0, 0.04); 添加轻微阴影 */
}

.footer-link i {
  font-size: 0.95rem;
}

.footer-link:hover {
  background-color: var(--selection-bg);
  color: var(--primary-text);
  transform: translateY(-2px);
  box-shadow: 0 4px 8px rgba(0, 0, 0, 0.08); /* 悬停时增加阴影 */
}

/* 分隔线样式 */
.footer-divider {
  width: 100%;
  height: 1px;
  /* background: linear-gradient(90deg, 
    transparent, 
    rgba(124, 58, 237, 0.2), 
    rgba(124, 58, 237, 0.3), 
    rgba(124, 58, 237, 0.2), 
    transparent
  ); */
  border-bottom: 1px solid rgba(124, 58, 237, 0.07);
  margin: 0.6rem 0; /* 增加上下边距 */
}

/* 版权信息 */
.copyright {
  text-align: center;
  color: var(--secondary-text);
  font-size: 0.8rem;
  padding: 0.3rem 0;
}

.copyright p {
  margin: 0;
}

/* 响应式调整 */
@media (max-width: 768px) {
  .app-footer {
    padding: 1.4rem 0 1rem;
  }
  
  .footer-row {
    padding: 0.5rem 0;
    gap: 0.4rem;
  }
  
  .footer-link {
    font-size: 0.8rem;
    padding: 0.3rem 0.7rem;
  }
  
  .footer-divider {
    margin: 0.5rem 0;
  }
}

@media (max-width: 480px) {
  .app-footer {
    padding: 1.2rem 0 0.8rem;
    margin: 1.5rem 0.5rem 0.5rem;
  }
  
  .footer-inner {
    padding: 0 0.3rem;
  }
  
  .footer-row {
    padding: 0.3rem 0;
    gap: 0.3rem;
  }
  
  .footer-link {
    font-size: 0.75rem;
    padding: 0.25rem 0.5rem;
  }
  
  .footer-divider {
    margin: 0.4rem 0;
  }
}

/* 悬浮按钮样式 */
.floating-buttons {
  position: fixed;
  right: 25px;
  bottom: 70px;
  display: flex;
  flex-direction: column;
  gap: 15px;
  z-index: 100;
  /* 添加硬件加速，防止滚动时出现撕裂 */
  will-change: transform;
  transform: translateZ(0);
  -webkit-backface-visibility: hidden;
  backface-visibility: hidden;
  /* 确保按钮初始渲染完成 */
  opacity: 1;
  transition: var(--transition);
}

.floating-btn {
  width: 60px;
  height: 60px;
  border-radius: 30px;
  color: var(--primary-color);
  border: var(--glass-border);
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 1.3rem;
  cursor: pointer;
  transition: var(--transition-bounce);
  text-decoration: none;
  background: rgba(255, 255, 255, 0.7);
  backdrop-filter: var(--glass-blur);
  -webkit-backdrop-filter: var(--glass-blur);
  box-shadow: 
    0 10px 20px rgba(0, 0, 0, 0.1),
    0 5px 10px rgba(0, 0, 0, 0.05),
    inset 0 1px 0 rgba(255, 255, 255, 0.9),
    inset 0 -2px 5px rgba(0, 0, 0, 0.03);
  overflow: hidden;
}

.floating-btn::before {
  content: '';
  position: absolute;
  top: 0;
  left: -170%;
  width: 100%;
  height: 100%;
  background: linear-gradient(90deg, transparent, rgba(255, 255, 255, 0.2), transparent);
  transition: 0.5s;
  z-index: -1;
}

.floating-btn:hover {
  transform: translateY(-5px) scale(1.05);
  box-shadow: 
    0 15px 25px rgba(0, 0, 0, 0.15),
    0 8px 15px rgba(0, 0, 0, 0.1),
    inset 0 1px 0 rgba(255, 255, 255, 1);
}

.floating-btn:hover::before {
  left: 100%;
}

.floating-btn:active {
  transform: translateY(0) scale(0.95);
}

.floating-btn.home-btn {
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.8), rgba(240, 244, 255, 0.9));
  color: var(--primary-color);
  box-shadow: 
    0 10px 20px rgba(124, 58, 237, 0.15),
    0 5px 10px rgba(124, 58, 237, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 1);
}

.floating-btn.top-btn {
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.8), rgba(240, 244, 255, 0.9));
  color: var(--secondary-color);
  box-shadow: 
    0 10px 20px rgba(6, 182, 212, 0.15),
    0 5px 10px rgba(6, 182, 212, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 1);
}

.floating-btn.bottom-btn {
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.8), rgba(240, 244, 255, 0.9));
  color: var(--accent-color);
  box-shadow: 
    0 10px 20px rgba(244, 63, 94, 0.15),
    0 5px 10px rgba(244, 63, 94, 0.08),
    inset 0 1px 0 rgba(255, 255, 255, 1);
}

.floating-btn.home-btn:hover {
  color: white;
  background: var(--primary-gradient);
}

.floating-btn.top-btn:hover {
  color: white;
  background: var(--secondary-gradient);
}

.floating-btn.bottom-btn:hover {
  color: white;
  background: var(--accent-gradient);
}

/* 动画 */
@keyframes fadeInUpScale {
  from { 
    opacity: 0; 
    transform: translateY(20px) scale(0.98); 
  }
  to { 
    opacity: 1; 
    transform: translateY(0) scale(1); 
  }
}

@keyframes pulse {
  0% {
    transform: scale(1);
    box-shadow: 0 0 0 0 rgba(124, 58, 237, 0.4);
  }
  70% {
    transform: scale(1.05);
    box-shadow: 0 0 0 10px rgba(124, 58, 237, 0);
  }
  100% {
    transform: scale(1);
    box-shadow: 0 0 0 0 rgba(124, 58, 237, 0);
  }
}

/* 媒体查询 - 从大屏幕到小屏幕 */
@media (min-width: 1201px) {
  .header-search {
    margin: 0 2rem;
    max-width: 700px;
  }
}

@media (min-width: 993px) {
  .header-inner {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: 1.5rem;
  }
  
  .header-search {
    justify-self: center;
    width: 100%; /* 确保宽度填充可用空间 */
  }
}

@media(max-width: 1200px) {
  .btn-custom {
    padding: 0.6rem 1.2rem;
  }
  
  .btn-custom .btn-text {
    font-size: 0.9rem;
  }
  
  .header-search {
    margin: 0 1.5rem;
    max-width: 550px;
  }
}

@media (max-width: 1200px) {
  .header-inner {
    display: flex;
    flex-wrap: wrap;
    gap: 1rem;
    padding: 0.75rem 0;
    justify-content: space-between;
  }
  
  .brand, .header-actions {
    flex: 0 0 auto;
  }
  
  /* 让搜索框在中等屏幕上与两侧元素同行 */
  .header-search {
    flex: 1;
    margin: 0 0.75rem;
    min-width: 200px;
    max-width: 100%;
  }
  
  .header-search :deep(.search-container) {
    padding: 0 0.25rem;
  }
  
  .btn-custom {
    padding: 0.6rem 1rem;
    font-size: 0.9rem;
  }
  
  .btn-custom .btn-text {
    display: none;
  }
  
  .btn-custom i {
    margin-right: 0 !important;
    font-size: 1.2rem;
  }
}

@media (max-width: 768px) {
  .app-header {
    padding: 0.5rem 0;
  }
  
  .header-inner {
    padding: 0.5rem 0;
    flex-direction: row;
    flex-wrap: wrap;
    justify-content: space-between;
    gap: 1rem;
  }
  
  .brand, .header-actions {
    width: auto;
  }
  
  /* 调整品牌区域右侧边距 */
  .brand {
    padding-left: 0.75rem;
  }
  
  .header-actions {
    padding-right: 0.75rem;
  }
  
  .button-group {
    padding: 0.25rem 0;
  }

  .brand-link {
    max-width: 180px;
    font-size: 1.3rem;
  }
  
  .user-greeting {
    justify-content: center;
    font-size: 0.9rem;
    padding: 0.4rem 0.8rem;
    margin: 0.25rem 0;
  }
  
  .button-group {
    justify-content: center;
    flex-wrap: wrap;
    margin-top: 0.25rem;
    gap: 0.5rem;
  }
  
  .btn-custom {
    padding: 0.6rem;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: flex;
    justify-content: center;
    align-items: center;
    margin: 0.1rem;
  }
  
  .btn-custom .btn-text {
    display: none;
  }
  
  .btn-custom i {
    margin-right: 0 !important;
    font-size: 1.1rem;
  }
  
  /* 小屏幕上搜索框样式 */
  .header-search {
    flex: 1 0 100%;
    margin: 0.75rem auto 0.5rem;
    order: 3;
    min-width: 0;
    max-width: 66%; /* 变为原来的三分之二 */
    width: 66%;
    justify-content: center; /* 内部元素居中 */
  }
  
  .header-search :deep(.search-container) {
    padding: 0 0.5rem;
  }
  
  .header-search :deep(.search-button) {
    padding-right: 0;
  }
  
  .header-search :deep(.search-icon) {
    margin-left: 2.75rem;
  }
  
  .footer-content {
    flex-direction: column;
    gap: 2rem;
    align-items: center;
  }
  
  .footer-brand {
    justify-content: center;
  }
  
  .footer-links {
    justify-content: center;
  }
  
  .app-header, .app-footer {
    margin: 0;
    border-radius: 0;
  }
  
  .container {
    padding: 0 0.75rem;
  }
  
  .floating-buttons {
    right: 15px;
    bottom: 60px;
    gap: 10px;
  }
  
  .floating-btn {
    width: 45px;
    height: 45px;
    font-size: 1.1rem;
  }
}

@media (max-width: 480px) {
  .footer-links {
    flex-direction: row;
    align-items: center;
    gap: 0.75rem;
    justify-content: center;
    flex-wrap: wrap;
  }
  
  .footer-link {
    width: auto;
    text-align: center;
    padding: 0.4rem 0.8rem;
    font-size: 0.9rem;
  }
  
  .floating-buttons {
    right: 10px;
    bottom: 50px;
  }
  
  .floating-btn {
    width: 40px;
    height: 40px;
    font-size: 1rem;
  }
  
  .brand {
    padding-left: 0.75rem;
  }
  
  .brand-link {
    max-width: 160px;
    font-size: 1.2rem;
  }
  
  .brand-icon {
    font-size: 1.5rem;
  }
  
  /* 小屏幕上搜索框居中显示 */
  .header-search {
    max-width: 66%;
    width: 66%;
    margin-left: auto;
    margin-right: auto;
  }
  
  .header-search :deep(.search-container) {
    padding: 0 0.5rem;
  }
  
  .header-search :deep(.search-icon) {
    margin-left: 0.85rem;
  }
}

/* 超小屏幕适配 */
@media (max-width: 360px) {
  .container {
    padding: 0 0.5rem;
  }
  
  .brand {
    padding-left: 0.5rem;
  }
  
  .brand-link {
    max-width: 130px;
    font-size: 1.1rem;
  }
  
  .header-search {
    max-width: 70%;
    width: 70%;
  }
  
  .header-search :deep(.search-icon) {
    margin-left: 1rem;
  }
  
  .floating-buttons {
    right: 8px;
    bottom: 40px;
    gap: 8px;
  }
  
  .floating-btn {
    width: 35px;
    height: 35px;
    font-size: 0.9rem;
  }
}

/* 底部内容预加载的占位符类 */
.prefetch-footer {
  position: absolute;
  height: 0;
  width: 0;
  opacity: 0;
  pointer-events: none;
}

/* 搜索框样式 - 与Hero宽度一致 */
.header-search {
  flex: 1;
  margin: 0 1.5rem;
  min-width: 280px;
  max-width: 700px; /* 限制最大宽度，与Hero区域接近 */
  position: relative;
  z-index: 110;
  width: auto;
  align-self: center; /* 确保垂直居中 */
}

/* 搜索框内部样式 - 应用到LocalSearch组件内部 */
.header-search :deep(.search-container) {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}

.header-search :deep(.search-input) {
  flex: 1;
}

.header-search :deep(.search-button) {
  margin-left: auto;
  padding-right: 0.5rem;
  display: flex;
  align-items: center;
  justify-content: center;
}

.header-search :deep(.search-icon) {
  margin-left: 0.5rem;
}

/* 移动端搜索框特殊样式 */
@media (max-width: 768px) {
  .header-search :deep(.search-container) {
    padding: 0 0.5rem;
  }
  
  .header-search :deep(.search-button) {
    padding-right: 0;
  }
  
  .header-search :deep(.search-icon) {
    margin-left: 0.75rem;
  }
  
  .app-header {
    padding: 0.5rem 0;
  }
  
  .brand, .header-actions {
    width: auto;
  }
  
  .brand {
    padding-left: 0.75rem;
  }
  
  .header-actions {
    padding-right: 0.75rem;
  }
  
  .button-group {
    padding: 0.25rem 0;
  }

  .brand-link {
    max-width: 180px;
    font-size: 1.3rem;
  }
  
  .user-greeting {
    justify-content: center;
    font-size: 0.9rem;
    padding: 0.4rem 0.8rem;
    margin: 0.25rem 0;
    margin-left: 1.5rem;
  }
  
  .button-group {
    justify-content: center;
    flex-wrap: wrap;
    margin-top: 0.25rem;
    gap: 0.5rem;
  }
  
  .btn-custom {
    padding: 0.6rem;
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: flex;
    justify-content: center;
    align-items: center;
    margin: 0.1rem;
  }
  
  .btn-custom .btn-text {
    display: none;
  }
  
  .btn-custom i {
    margin-right: 0 !important;
    font-size: 1.1rem;
  }
  
  .footer-content {
    flex-direction: column;
    gap: 2rem;
    align-items: center;
  }
  
  .footer-brand {
    justify-content: center;
  }
  
  .footer-links {
    justify-content: center;
  }
  
  .app-header, .app-footer {
    margin: 0;
    border-radius: 0;
  }
  
  .container {
    padding: 0 0.75rem;
  }
  
  .floating-buttons {
    right: 15px;
    bottom: 60px;
    gap: 10px;
  }
  
  .floating-btn {
    width: 45px;
    height: 45px;
    font-size: 1.1rem;
  }
}

/* 删除不再使用的页脚相关样式 */
.friend-link {
  display: none; /* 这个不再使用 */
}

.site-stats {
  display: none; /* 这个不再使用 */
}

.site-stats span {
  display: none; /* 这个不再使用 */
}

.social-links {
  display: none; /* 这个不再使用 */
}

.social-link {
  display: none; /* 这个不再使用 */
}

.left-section {
  display: none; /* 这个不再使用 */
}

.links-section {
  display: none; /* 这个不再使用 */
}

.links-container {
  display: none; /* 这个不再使用 */
}

.footer-section {
  display: none; /* 这个不再使用 */
}

.footer-button {
  display: none; /* 这个不再使用 */
}

.visit-count {
  display: none; /* 这个不再使用 */
}
</style> 