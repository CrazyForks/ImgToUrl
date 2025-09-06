<template>
  <nav class="navbar">
    <div class="nav-content">
      <router-link to="/" class="logo">
        <el-icon :size="28" color="#fff">
          <Picture />
        </el-icon>
        <span class="logo-text">图床转换系统</span>
      </router-link>
      <div class="nav-menu">
        <router-link to="/" class="nav-link" active-class="active">
          <el-icon><House /></el-icon>
          首页
        </router-link>
        <router-link to="/upload" class="nav-link" active-class="active">
          <el-icon><Upload /></el-icon>
          上传图片
        </router-link>
        <router-link to="/gallery" class="nav-link" active-class="active">
          <el-icon><Picture /></el-icon>
          图片画廊
        </router-link>
        <router-link to="/stats" class="nav-link" active-class="active">
          <el-icon><DataAnalysis /></el-icon>
          统计信息
        </router-link>
        <router-link to="/storage" class="nav-link" active-class="active">
          <el-icon><Folder /></el-icon>
          存储内容
        </router-link>
      </div>
    </div>
  </nav>
</template>

<script setup lang="ts">
import { Picture, House, Upload, DataAnalysis, Folder } from '@element-plus/icons-vue'
</script>

<style lang="scss" scoped>
.navbar {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  padding: 0 2rem;
  position: sticky;
  top: 0;
  z-index: 100;

  .nav-content {
    max-width: 1400px;
    margin: 0 auto;
    display: flex;
    justify-content: space-between;
    align-items: center;
    height: 64px;
  }

  .logo {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    color: white;
    font-size: 1.125rem;
    font-weight: 600;
    text-decoration: none;
  }

  .nav-menu {
    display: flex;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .nav-link {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    color: rgba(255, 255, 255, 0.85);
    text-decoration: none;
    padding: 0.45rem 0.8rem;
    border-radius: 8px;
    transition: all 0.25s ease;

    &:hover,
    &.active {
      color: white;
      background: rgba(255, 255, 255, 0.12);
    }
  }
}

@media (max-width: 768px) {
  .navbar {
    padding: 0 1rem;
    .nav-menu {
      gap: 0.5rem;
    }
    .nav-link {
      padding: 0.4rem 0.6rem;
      font-size: 0.875rem;
    }
  }
}
</style>