<template>
  <div class="not-found-container">
    <div class="not-found-content">
      <!-- 404 图标和数字 -->
      <div class="error-visual">
        <div class="error-number">404</div>
        <div class="error-icon">
          <el-icon :size="120" color="rgba(255, 255, 255, 0.6)">
            <QuestionFilled />
          </el-icon>
        </div>
      </div>
      
      <!-- 错误信息 -->
      <div class="error-info">
        <h1 class="error-title">页面未找到</h1>
        <p class="error-description">
          抱歉，您访问的页面不存在或已被移动。
        </p>
        <p class="error-suggestion">
          请检查URL是否正确，或返回首页继续浏览。
        </p>
      </div>
      
      <!-- 操作按钮 -->
      <div class="error-actions">
        <el-button type="primary" size="large" @click="goHome">
          <el-icon><House /></el-icon>
          返回首页
        </el-button>
        <el-button size="large" @click="goBack">
          <el-icon><Back /></el-icon>
          返回上页
        </el-button>
        <el-button size="large" @click="goUpload">
          <el-icon><Upload /></el-icon>
          上传图片
        </el-button>
      </div>
      
      <!-- 快速链接 -->
      <div class="quick-links">
        <h3 class="links-title">您可能想要：</h3>
        <div class="links-grid">
          <router-link to="/" class="quick-link">
            <div class="link-icon">
              <el-icon><House /></el-icon>
            </div>
            <div class="link-content">
              <div class="link-title">首页</div>
              <div class="link-desc">查看系统概览</div>
            </div>
          </router-link>
          
          <router-link to="/upload" class="quick-link">
            <div class="link-icon">
              <el-icon><Upload /></el-icon>
            </div>
            <div class="link-content">
              <div class="link-title">上传图片</div>
              <div class="link-desc">上传新的图片</div>
            </div>
          </router-link>
          
          <router-link to="/gallery" class="quick-link">
            <div class="link-icon">
              <el-icon><Picture /></el-icon>
            </div>
            <div class="link-content">
              <div class="link-title">图片画廊</div>
              <div class="link-desc">浏览所有图片</div>
            </div>
          </router-link>
          
          <router-link to="/stats" class="quick-link">
            <div class="link-icon">
              <el-icon><DataAnalysis /></el-icon>
            </div>
            <div class="link-content">
              <div class="link-title">统计信息</div>
              <div class="link-desc">查看数据统计</div>
            </div>
          </router-link>
        </div>
      </div>
    </div>
    
    <!-- 装饰性元素 -->
    <div class="decoration">
      <div class="floating-element" v-for="i in 6" :key="i" :style="getFloatingStyle(i)">
        <el-icon :size="20 + i * 4" :color="getFloatingColor(i)">
          <Picture />
        </el-icon>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import {
  QuestionFilled,
  House,
  Back,
  Upload,
  Picture,
  DataAnalysis
} from '@element-plus/icons-vue'

const router = useRouter()

// 方法
const goHome = () => {
  router.push('/')
}

const goBack = () => {
  if (window.history.length > 1) {
    router.go(-1)
  } else {
    router.push('/')
  }
}

const goUpload = () => {
  router.push('/upload')
}

// 装饰性元素样式
const getFloatingStyle = (index: number) => {
  const positions = [
    { top: '10%', left: '10%', animationDelay: '0s' },
    { top: '20%', right: '15%', animationDelay: '1s' },
    { top: '60%', left: '5%', animationDelay: '2s' },
    { top: '70%', right: '10%', animationDelay: '3s' },
    { top: '40%', left: '15%', animationDelay: '4s' },
    { top: '80%', right: '20%', animationDelay: '5s' }
  ]
  
  return {
    ...positions[index - 1],
    animationDelay: positions[index - 1].animationDelay
  }
}

const getFloatingColor = (index: number) => {
  const colors = [
    'rgba(64, 158, 255, 0.3)',
    'rgba(103, 194, 58, 0.3)',
    'rgba(230, 162, 60, 0.3)',
    'rgba(245, 108, 108, 0.3)',
    'rgba(144, 147, 153, 0.3)',
    'rgba(255, 255, 255, 0.2)'
  ]
  return colors[index - 1]
}
</script>

<style lang="scss" scoped>
.not-found-container {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  padding: 2rem;
}

.not-found-content {
  text-align: center;
  max-width: 800px;
  z-index: 10;
}

// 错误视觉元素
.error-visual {
  position: relative;
  margin-bottom: 3rem;

  .error-number {
    font-size: 8rem;
    font-weight: 900;
    color: rgba(255, 255, 255, 0.1);
    line-height: 1;
    margin-bottom: -2rem;
    text-shadow: 0 0 20px rgba(255, 255, 255, 0.1);
  }

  .error-icon {
    position: relative;
    z-index: 2;
  }
}

// 错误信息
.error-info {
  margin-bottom: 3rem;

  .error-title {
    font-size: 2.5rem;
    font-weight: 700;
    color: white;
    margin-bottom: 1rem;
    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  }

  .error-description {
    font-size: 1.25rem;
    color: rgba(255, 255, 255, 0.8);
    margin-bottom: 1rem;
    line-height: 1.6;
  }

  .error-suggestion {
    font-size: 1rem;
    color: rgba(255, 255, 255, 0.6);
    line-height: 1.6;
  }
}

// 操作按钮
.error-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
  flex-wrap: wrap;
  margin-bottom: 4rem;
}

// 快速链接
.quick-links {
  .links-title {
    font-size: 1.25rem;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.8);
    margin-bottom: 2rem;
  }

  .links-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
    max-width: 600px;
    margin: 0 auto;
  }

  .quick-link {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1.5rem;
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 12px;
    text-decoration: none;
    transition: all 0.3s ease;

    &:hover {
      transform: translateY(-2px);
      background: rgba(255, 255, 255, 0.15);
      border-color: rgba(255, 255, 255, 0.3);
    }

    .link-icon {
      width: 40px;
      height: 40px;
      border-radius: 8px;
      background: rgba(255, 255, 255, 0.1);
      display: flex;
      align-items: center;
      justify-content: center;
      color: rgba(255, 255, 255, 0.8);
    }

    .link-content {
      text-align: left;

      .link-title {
        font-weight: 600;
        color: white;
        margin-bottom: 0.25rem;
      }

      .link-desc {
        font-size: 0.875rem;
        color: rgba(255, 255, 255, 0.6);
      }
    }
  }
}

// 装饰性元素
.decoration {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  pointer-events: none;
  z-index: 1;

  .floating-element {
    position: absolute;
    animation: float 6s ease-in-out infinite;
  }
}

// 动画
@keyframes float {
  0%, 100% {
    transform: translateY(0px) rotate(0deg);
  }
  25% {
    transform: translateY(-20px) rotate(5deg);
  }
  50% {
    transform: translateY(-10px) rotate(-5deg);
  }
  75% {
    transform: translateY(-15px) rotate(3deg);
  }
}

// 响应式设计
@media (max-width: 768px) {
  .not-found-container {
    padding: 1rem;
  }

  .error-visual {
    .error-number {
      font-size: 4rem;
    }
  }

  .error-info {
    .error-title {
      font-size: 2rem;
    }

    .error-description {
      font-size: 1.125rem;
    }
  }

  .error-actions {
    flex-direction: column;
    align-items: center;
  }

  .links-grid {
    grid-template-columns: 1fr;
  }

  .quick-link {
    padding: 1rem;
  }
}

@media (max-width: 480px) {
  .error-visual {
    .error-number {
      font-size: 3rem;
    }
  }

  .error-info {
    .error-title {
      font-size: 1.5rem;
    }

    .error-description {
      font-size: 1rem;
    }
  }

  .quick-link {
    flex-direction: column;
    text-align: center;
    gap: 0.5rem;

    .link-content {
      text-align: center;
    }
  }
}
</style>