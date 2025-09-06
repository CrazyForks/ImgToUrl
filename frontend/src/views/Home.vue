<template>
  <div class="home-container">
    <NavBar />

    <!-- 主要内容 -->
    <main class="main-content">
      <!-- 英雄区域 -->
      <section class="hero-section">
        <div class="hero-content">
          <h1 class="hero-title">
            高性能图床转换系统
          </h1>
          <p class="hero-subtitle">
            基于 Cloudflare R2 的快速、安全、可靠的图片存储服务
          </p>
          <div class="hero-features">
            <div class="feature-item">
              <el-icon :size="24" color="#67C23A">
                <Lightning />
              </el-icon>
              <span>极速上传</span>
            </div>
            <div class="feature-item">
              <el-icon :size="24" color="#409EFF">
                <Lock />
              </el-icon>
              <span>安全可靠</span>
            </div>
            <div class="feature-item">
              <el-icon :size="24" color="#E6A23C">
                <Star />
              </el-icon>
              <span>高性能</span>
            </div>
          </div>
          <div class="hero-actions">
            <el-button type="primary" size="large" @click="goToUpload">
              <el-icon><Upload /></el-icon>
              开始上传
            </el-button>
            <el-button size="large" @click="viewStats">
              <el-icon><DataAnalysis /></el-icon>
              查看统计
            </el-button>
          </div>
        </div>
      </section>

      <!-- 统计信息卡片 -->
      <section class="stats-section">
        <div class="stats-container">
          <div class="stat-card">
            <div class="stat-icon">
              <el-icon :size="32" color="#409EFF">
                <Picture />
              </el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-number">{{ stats.total_images.toLocaleString() }}</div>
              <div class="stat-label">总图片数</div>
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-icon">
              <el-icon :size="32" color="#67C23A">
                <Folder />
              </el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-number">{{ formatFileSize(stats.total_size) }}</div>
              <div class="stat-label">总存储量</div>
            </div>
          </div>
          <div class="stat-card">
            <div class="stat-icon">
              <el-icon :size="32" color="#E6A23C">
                <Calendar />
              </el-icon>
            </div>
            <div class="stat-content">
              <div class="stat-number">{{ stats.today_images.toLocaleString() }}</div>
              <div class="stat-label">今日上传</div>
            </div>
          </div>
        </div>
      </section>

      <!-- 快速上传区域 -->
      <section class="quick-upload-section">
        <div class="upload-container">
          <h2 class="section-title">快速上传</h2>
          <div class="upload-area" @drop="handleDrop" @dragover="handleDragOver" @dragleave="handleDragLeave">
            <div class="upload-content">
              <el-icon :size="48" color="#909399">
                <UploadFilled />
              </el-icon>
              <p class="upload-text">拖拽图片到此处或点击上传</p>
              <p class="upload-hint">支持 JPG、PNG、GIF、WebP 格式，单个文件最大 10MB</p>
              <el-button type="primary" @click="selectFiles">
                选择文件
              </el-button>
            </div>
          </div>
          <input ref="fileInput" type="file" multiple accept="image/*" @change="handleFileSelect" style="display: none;" />
        </div>
      </section>
    </main>

    <!-- 页脚 -->
    <footer class="footer">
      <div class="footer-content">
        <p>&copy; 2024 图床转换系统. 基于 Vue3 + Go + MySQL + Cloudflare R2</p>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Picture,
  Upload,
  DataAnalysis,
  Lightning,
  Lock,
  Star,
  Folder,
  Calendar,
  UploadFilled
} from '@element-plus/icons-vue'
import NavBar from '@/components/NavBar.vue'
import { useUploadStore } from '@/stores/upload'

const router = useRouter()
const uploadStore = useUploadStore()
const fileInput = ref<HTMLInputElement>()

// 统计信息
const stats = ref({
  total_images: 0,
  total_size: 0,
  today_images: 0
})

// 格式化文件大小
const formatFileSize = (bytes: number): string => {
  return uploadStore.formatFileSize(bytes)
}

// 导航方法
const goToUpload = () => {
  router.push('/upload')
}

const viewStats = () => {
  router.push('/stats')
}

// 文件选择
const selectFiles = () => {
  fileInput.value?.click()
}

// 处理文件选择
const handleFileSelect = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const files = target.files
  if (files && files.length > 0) {
    await handleFiles(Array.from(files))
  }
}

// 拖拽处理
const handleDragOver = (event: DragEvent) => {
  event.preventDefault()
  event.dataTransfer!.dropEffect = 'copy'
}

const handleDragLeave = (event: DragEvent) => {
  event.preventDefault()
}

const handleDrop = async (event: DragEvent) => {
  event.preventDefault()
  const files = event.dataTransfer?.files
  if (files && files.length > 0) {
    await handleFiles(Array.from(files))
  }
}

// 处理文件上传
const handleFiles = async (files: File[]) => {
  // 过滤图片文件
  const imageFiles = files.filter(file => file.type.startsWith('image/'))
  
  if (imageFiles.length === 0) {
    ElMessage.warning('请选择图片文件')
    return
  }

  if (imageFiles.length > 10) {
    ElMessage.warning('最多只能同时上传10个文件')
    return
  }

  try {
    let result
    if (imageFiles.length === 1) {
      result = await uploadStore.uploadSingleImage(imageFiles[0])
    } else {
      result = await uploadStore.uploadMultipleImages(imageFiles)
    }

    if (result.success) {
      ElMessage.success('上传成功！')
      await loadStats() // 重新加载统计信息
      router.push('/upload') // 跳转到上传页面查看结果
    } else {
      ElMessage.error(result.error || '上传失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '上传失败')
  }
}

// 加载统计信息
const loadStats = async () => {
  await uploadStore.fetchStats()
  stats.value = uploadStore.stats
}

// 组件挂载时加载统计信息
onMounted(() => {
  loadStats()
})
</script>

<style lang="scss" scoped>
.home-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

// 导航栏样式
.navbar {
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(10px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.2);
  padding: 0 2rem;
  position: sticky;
  top: 0;
  z-index: 100;

  .nav-content {
    max-width: 1200px;
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
    font-size: 1.25rem;
    font-weight: 600;
  }

  .nav-menu {
    display: flex;
    gap: 2rem;
  }

  .nav-link {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    color: rgba(255, 255, 255, 0.8);
    text-decoration: none;
    padding: 0.5rem 1rem;
    border-radius: 8px;
    transition: all 0.3s ease;

    &:hover {
      color: white;
      background: rgba(255, 255, 255, 0.1);
    }
  }
}

// 主要内容
.main-content {
  flex: 1;
  padding: 2rem;
}

// 英雄区域
.hero-section {
  text-align: center;
  padding: 4rem 0;
  max-width: 800px;
  margin: 0 auto;

  .hero-title {
    font-size: 3rem;
    font-weight: 700;
    color: white;
    margin-bottom: 1rem;
    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  }

  .hero-subtitle {
    font-size: 1.25rem;
    color: rgba(255, 255, 255, 0.9);
    margin-bottom: 2rem;
    line-height: 1.6;
  }

  .hero-features {
    display: flex;
    justify-content: center;
    gap: 2rem;
    margin-bottom: 2rem;

    .feature-item {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      color: white;
      font-weight: 500;
    }
  }

  .hero-actions {
    display: flex;
    justify-content: center;
    gap: 1rem;
  }
}

// 统计信息
.stats-section {
  margin: 4rem 0;

  .stats-container {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 2rem;
    max-width: 1000px;
    margin: 0 auto;
  }

  .stat-card {
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 16px;
    padding: 2rem;
    display: flex;
    align-items: center;
    gap: 1rem;
    transition: transform 0.3s ease;

    &:hover {
      transform: translateY(-4px);
    }

    .stat-icon {
      flex-shrink: 0;
    }

    .stat-content {
      .stat-number {
        font-size: 2rem;
        font-weight: 700;
        color: white;
        line-height: 1;
      }

      .stat-label {
        font-size: 0.875rem;
        color: rgba(255, 255, 255, 0.7);
        margin-top: 0.25rem;
      }
    }
  }
}

// 快速上传区域
.quick-upload-section {
  margin: 4rem 0;

  .upload-container {
    max-width: 600px;
    margin: 0 auto;
  }

  .section-title {
    font-size: 2rem;
    font-weight: 600;
    color: white;
    text-align: center;
    margin-bottom: 2rem;
  }

  .upload-area {
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    border: 2px dashed rgba(255, 255, 255, 0.3);
    border-radius: 16px;
    padding: 3rem;
    text-align: center;
    transition: all 0.3s ease;
    cursor: pointer;

    &:hover {
      border-color: rgba(255, 255, 255, 0.5);
      background: rgba(255, 255, 255, 0.15);
    }

    .upload-content {
      .upload-text {
        font-size: 1.125rem;
        color: white;
        margin: 1rem 0 0.5rem;
      }

      .upload-hint {
        font-size: 0.875rem;
        color: rgba(255, 255, 255, 0.7);
        margin-bottom: 1.5rem;
      }
    }
  }
}

// 页脚
.footer {
  background: rgba(0, 0, 0, 0.2);
  padding: 2rem;
  text-align: center;

  .footer-content {
    color: rgba(255, 255, 255, 0.7);
    font-size: 0.875rem;
  }
}

// 响应式设计
@media (max-width: 768px) {
  .navbar {
    padding: 0 1rem;

    .nav-menu {
      gap: 1rem;
    }

    .nav-link {
      padding: 0.5rem;
      font-size: 0.875rem;
    }
  }

  .main-content {
    padding: 1rem;
  }

  .hero-section {
    padding: 2rem 0;

    .hero-title {
      font-size: 2rem;
    }

    .hero-features {
      flex-direction: column;
      gap: 1rem;
    }

    .hero-actions {
      flex-direction: column;
      align-items: center;
    }
  }

  .stats-container {
    grid-template-columns: 1fr;
  }

  .upload-area {
    padding: 2rem 1rem;
  }
}
</style>