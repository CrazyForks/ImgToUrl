<template>
  <div class="upload-container">
    <!-- 导航栏 -->
    <nav class="navbar">
      <div class="nav-content">
        <router-link to="/" class="logo">
          <el-icon :size="32" color="#fff">
            <Picture />
          </el-icon>
          <span class="logo-text">图床转换系统</span>
        </router-link>
        <div class="nav-menu">
          <router-link to="/" class="nav-link">
            <el-icon><House /></el-icon>
            首页
          </router-link>
          <router-link to="/gallery" class="nav-link">
            <el-icon><Picture /></el-icon>
            图片画廊
          </router-link>
          <router-link to="/stats" class="nav-link">
            <el-icon><DataAnalysis /></el-icon>
            统计信息
          </router-link>
        </div>
      </div>
    </nav>

    <!-- 主要内容 -->
    <main class="main-content">
      <div class="upload-wrapper">
        <!-- 上传区域 -->
        <section class="upload-section">
          <h1 class="page-title">图片上传</h1>
          
          <!-- 上传组件 -->
          <div class="upload-area" 
               :class="{ 'drag-over': isDragOver }"
               @drop="handleDrop" 
               @dragover="handleDragOver" 
               @dragleave="handleDragLeave">
            <div v-if="!isUploading" class="upload-content">
              <el-icon :size="64" color="#909399">
                <UploadFilled />
              </el-icon>
              <h3 class="upload-title">拖拽图片到此处或点击上传</h3>
              <p class="upload-description">
                支持 JPG、PNG、GIF、WebP 格式<br>
                单个文件最大 10MB，最多同时上传 10 个文件
              </p>
              <div class="upload-buttons">
                <el-button type="primary" size="large" @click="selectFiles">
                  <el-icon><Plus /></el-icon>
                  选择文件
                </el-button>
                <el-button size="large" @click="selectMultipleFiles">
                  <el-icon><FolderAdd /></el-icon>
                  批量上传
                </el-button>
              </div>
            </div>
            
            <!-- 上传进度 -->
            <div v-else class="upload-progress">
              <el-icon :size="48" color="#409EFF" class="loading-icon">
                <Loading />
              </el-icon>
              <h3>正在上传...</h3>
              <el-progress :percentage="uploadProgress" :stroke-width="8" />
            </div>
          </div>
          
          <input ref="fileInput" 
                 type="file" 
                 multiple 
                 accept="image/*" 
                 @change="handleFileSelect" 
                 style="display: none;" />
        </section>

        <!-- 上传历史 -->
        <section v-if="uploadedImages.length > 0" class="history-section">
          <div class="section-header">
            <h2 class="section-title">
              <el-icon><Clock /></el-icon>
              上传历史 ({{ uploadedImages.length }})
            </h2>
            <div class="section-actions">
              <el-button size="small" @click="clearHistory">
                <el-icon><Delete /></el-icon>
                清空历史
              </el-button>
            </div>
          </div>
          
          <div class="image-grid">
            <div v-for="image in uploadedImages" 
                 :key="image.uuid" 
                 class="image-card">
              <div class="image-preview">
                <img :src="image.public_url" 
                     :alt="image.original_name" 
                     @load="handleImageLoad"
                     @error="handleImageError" />
                <div class="image-overlay">
                  <el-button circle size="small" @click="previewImage(image)">
                    <el-icon><ZoomIn /></el-icon>
                  </el-button>
                </div>
              </div>
              
              <div class="image-info">
                <div class="image-name" :title="image.original_name">
                  {{ image.original_name }}
                </div>
                <div class="image-meta">
                  <span class="image-size">{{ formatFileSize(image.file_size) }}</span>
                  <span class="image-dimensions">{{ image.width }}×{{ image.height }}</span>
                </div>
                <div class="image-time">
                  {{ formatTime(image.created_at) }}
                </div>
              </div>
              
              <div class="image-actions">
                <div class="url-input-group">
                  <el-input :model-value="toAbsoluteUrl(image.public_url)" 
                           readonly 
                           size="small" 
                           placeholder="图片链接">
                    <template #append>
                      <el-button @click="copyUrl(image.public_url)">
                        <el-icon><DocumentCopy /></el-icon>
                      </el-button>
                    </template>
                  </el-input>
                </div>
                
                <div class="action-buttons">
                  <el-dropdown trigger="click">
                    <el-button size="small">
                      复制链接
                      <el-icon><ArrowDown /></el-icon>
                    </el-button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item @click="copyUrl(image.public_url)">
                          <el-icon><Link /></el-icon>
                          直链
                        </el-dropdown-item>
                        <el-dropdown-item @click="copyMarkdown(image)">
                          <el-icon><Document /></el-icon>
                          Markdown
                        </el-dropdown-item>
                        <el-dropdown-item @click="copyHtml(image)">
                          <el-icon><DocumentCopy /></el-icon>
                          HTML
                        </el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                  
                  <el-button size="small" @click="removeImage(image.uuid)">
                    <el-icon><Delete /></el-icon>
                  </el-button>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>
    </main>

    <!-- 图片预览对话框 -->
    <el-dialog v-model="previewVisible" 
               title="图片预览" 
               width="80%" 
               center>
      <div class="preview-container">
        <img v-if="currentPreviewImage" 
             :src="currentPreviewImage?.public_url" 
             :alt="currentPreviewImage?.original_name" 
             class="preview-image" />
      </div>
      <template #footer>
        <div class="preview-actions">
          <el-button @click="copyUrl(currentPreviewImage?.public_url || '')">
            <el-icon><DocumentCopy /></el-icon>
            复制链接
          </el-button>
          <el-button @click="downloadImage(currentPreviewImage)">
            <el-icon><Download /></el-icon>
            下载图片
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  Picture,
  House,
  DataAnalysis,
  UploadFilled,
  Plus,
  FolderAdd,
  Loading,
  Clock,
  Delete,
  ZoomIn,
  DocumentCopy,
  ArrowDown,
  Link,
  Document,
  Download
} from '@element-plus/icons-vue'
import { useUploadStore, type ImageInfo } from '@/stores/upload'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import relativeTime from 'dayjs/plugin/relativeTime'

// 配置 dayjs
dayjs.locale('zh-cn')
dayjs.extend(relativeTime)

const router = useRouter()
const uploadStore = useUploadStore()
const fileInput = ref<HTMLInputElement>()

// 状态
const isDragOver = ref(false)
const previewVisible = ref(false)
const currentPreviewImage = ref<ImageInfo | null>(null)

// 计算属性
const uploadedImages = computed(() => uploadStore.uploadedImages)
const isUploading = computed(() => uploadStore.isUploading)
const uploadProgress = computed(() => uploadStore.uploadProgress)

// 格式化文件大小
const formatFileSize = (bytes: number): string => {
  return uploadStore.formatFileSize(bytes)
}

// 格式化时间
const formatTime = (dateString: string): string => {
  return dayjs(dateString).fromNow()
}

// 文件选择
const selectFiles = () => {
  if (fileInput.value) {
    fileInput.value.multiple = false
    fileInput.value.click()
  }
}

const selectMultipleFiles = () => {
  if (fileInput.value) {
    fileInput.value.multiple = true
    fileInput.value.click()
  }
}

// 处理文件选择
const handleFileSelect = async (event: Event) => {
  const target = event.target as HTMLInputElement
  const files = target.files
  if (files && files.length > 0) {
    await handleFiles(Array.from(files))
  }
  // 清空 input 值，允许重复选择同一文件
  target.value = ''
}

// 拖拽处理
const handleDragOver = (event: DragEvent) => {
  event.preventDefault()
  event.dataTransfer!.dropEffect = 'copy'
  isDragOver.value = true
}

const handleDragLeave = (event: DragEvent) => {
  event.preventDefault()
  isDragOver.value = false
}

const handleDrop = async (event: DragEvent) => {
  event.preventDefault()
  isDragOver.value = false
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
      if ('data' in result && result.data) {
        // 单个文件上传成功
        ElMessage.success('上传成功！')
      } else if ('data' in result && result.data && 'successful' in result.data) {
        // 批量上传结果
        const { successful, failed } = result.data
        if (failed > 0) {
          ElMessage.warning(`上传完成：成功 ${successful} 个，失败 ${failed} 个`)
        } else {
          ElMessage.success(`批量上传成功：${successful} 个文件`)
        }
      }
    } else {
      ElMessage.error(result.error || '上传失败')
    }
  } catch (error: any) {
    ElMessage.error(error.message || '上传失败')
  }
}

 // 复制链接
const toAbsoluteUrl = (u: string) => {
  try {
    return new URL(u, window.location.origin).href
  } catch {
    return u
  }
}
const copyUrl = async (url: string) => {
  const abs = toAbsoluteUrl(url)
  const success = await uploadStore.copyToClipboard(abs)
  if (success) {
    ElMessage.success('链接已复制到剪贴板')
  } else {
    ElMessage.error('复制失败')
  }
}

 // 复制 Markdown 格式
const copyMarkdown = async (image: ImageInfo) => {
  const abs = toAbsoluteUrl(image.public_url)
  const markdown = `![${image.original_name}](${abs})`
  const success = await uploadStore.copyToClipboard(markdown)
  if (success) {
    ElMessage.success('Markdown 格式已复制')
  } else {
    ElMessage.error('复制失败')
  }
}

 // 复制 HTML 格式
const copyHtml = async (image: ImageInfo) => {
  const abs = toAbsoluteUrl(image.public_url)
  const html = `<img src="${abs}" alt="${image.original_name}" />`
  const success = await uploadStore.copyToClipboard(html)
  if (success) {
    ElMessage.success('HTML 格式已复制')
  } else {
    ElMessage.error('复制失败')
  }
}

// 预览图片
const previewImage = (image: ImageInfo) => {
  currentPreviewImage.value = image
  previewVisible.value = true
}

// 下载图片
const downloadImage = (image: ImageInfo | null) => {
  if (!image) return
  
  const link = document.createElement('a')
  const abs = toAbsoluteUrl(image.public_url)
  link.href = abs
  link.download = image.original_name
  link.target = '_blank'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

// 移除图片
const removeImage = (uuid: string) => {
  uploadStore.removeImage(uuid)
  ElMessage.success('已从列表中移除')
}

// 清空历史
const clearHistory = async () => {
  try {
    await ElMessageBox.confirm(
      '确定要清空所有上传历史吗？此操作不可恢复。',
      '确认清空',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    uploadStore.clearUploadedImages()
    ElMessage.success('历史记录已清空')
  } catch {
    // 用户取消操作
  }
}

// 图片加载处理
const handleImageLoad = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.style.opacity = '1'
}

const handleImageError = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjVmNWY1Ii8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCwgc2Fucy1zZXJpZiIgZm9udC1zaXplPSIxNCIgZmlsbD0iIzk5OSIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPuWKoOi9veWksei0pTwvdGV4dD48L3N2Zz4='
}

// 组件挂载时加载统计信息
onMounted(() => {
  uploadStore.fetchStats()
})
</script>

<style lang="scss" scoped>
.upload-container {
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
    text-decoration: none;
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

.upload-wrapper {
  max-width: 1200px;
  margin: 0 auto;
}

// 上传区域
.upload-section {
  margin-bottom: 3rem;

  .page-title {
    font-size: 2.5rem;
    font-weight: 700;
    color: white;
    text-align: center;
    margin-bottom: 2rem;
    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
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
    min-height: 300px;
    display: flex;
    align-items: center;
    justify-content: center;

    &:hover,
    &.drag-over {
      border-color: #409EFF;
      background: rgba(64, 158, 255, 0.1);
    }

    .upload-content {
      .upload-title {
        font-size: 1.5rem;
        color: white;
        margin: 1rem 0;
        font-weight: 600;
      }

      .upload-description {
        color: rgba(255, 255, 255, 0.7);
        margin-bottom: 2rem;
        line-height: 1.6;
      }

      .upload-buttons {
        display: flex;
        gap: 1rem;
        justify-content: center;
      }
    }

    .upload-progress {
      width: 100%;
      max-width: 400px;

      .loading-icon {
        animation: rotate 2s linear infinite;
      }

      h3 {
        color: white;
        margin: 1rem 0 2rem;
        font-size: 1.25rem;
      }
    }
  }
}

// 历史记录区域
.history-section {
  .section-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 2rem;

    .section-title {
      display: flex;
      align-items: center;
      gap: 0.5rem;
      font-size: 1.5rem;
      font-weight: 600;
      color: white;
    }
  }

  .image-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
    gap: 1.5rem;
  }

  .image-card {
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 12px;
    overflow: hidden;
    transition: transform 0.3s ease;

    &:hover {
      transform: translateY(-4px);
    }

    .image-preview {
      position: relative;
      height: 200px;
      overflow: hidden;

      img {
        width: 100%;
        height: 100%;
        object-fit: cover;
        opacity: 0;
        transition: opacity 0.3s ease;
      }

      .image-overlay {
        position: absolute;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0, 0, 0, 0.5);
        display: flex;
        align-items: center;
        justify-content: center;
        opacity: 0;
        transition: opacity 0.3s ease;
      }

      &:hover .image-overlay {
        opacity: 1;
      }
    }

    .image-info {
      padding: 1rem;

      .image-name {
        font-weight: 600;
        color: white;
        margin-bottom: 0.5rem;
        overflow: hidden;
        text-overflow: ellipsis;
        white-space: nowrap;
      }

      .image-meta {
        display: flex;
        gap: 1rem;
        color: rgba(255, 255, 255, 0.7);
        font-size: 0.875rem;
        margin-bottom: 0.5rem;
      }

      .image-time {
        color: rgba(255, 255, 255, 0.5);
        font-size: 0.75rem;
      }
    }

    .image-actions {
      padding: 0 1rem 1rem;

      .url-input-group {
        margin-bottom: 1rem;
      }

      .action-buttons {
        display: flex;
        gap: 0.5rem;
      }
    }
  }
}

// 预览对话框
.preview-container {
  text-align: center;

  .preview-image {
    max-width: 100%;
    max-height: 70vh;
    object-fit: contain;
  }
}

.preview-actions {
  display: flex;
  gap: 1rem;
  justify-content: center;
}

// 动画
@keyframes rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
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

  .upload-area {
    padding: 2rem 1rem;

    .upload-buttons {
      flex-direction: column;
      align-items: center;
    }
  }

  .image-grid {
    grid-template-columns: 1fr;
  }

  .section-header {
    flex-direction: column;
    gap: 1rem;
    align-items: flex-start;
  }

  .action-buttons {
    flex-direction: column;
  }
}
</style>