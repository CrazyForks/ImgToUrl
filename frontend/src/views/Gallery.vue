<template>
  <div class="gallery-container">
    <NavBar />

    <!-- 主要内容 -->
    <main class="main-content">
      <div class="gallery-wrapper">
        <!-- 页面标题和工具栏 -->
        <section class="header-section">
          <div class="title-area">
            <h1 class="page-title">
              <el-icon><Picture /></el-icon>
              图片画廊
            </h1>
            <p class="page-description">
              浏览和管理您上传的所有图片
            </p>
          </div>
          
          <div class="toolbar">
            <div class="search-bar">
              <el-input
                v-model="searchQuery"
                placeholder="搜索图片名称..."
                :prefix-icon="Search"
                clearable
                @input="handleSearch"
              />
            </div>
            
            <div class="view-controls">
              <el-radio-group v-model="viewMode" size="small">
                <el-radio-button label="grid">
                  <el-icon><Grid /></el-icon>
                  网格
                </el-radio-button>
                <el-radio-button label="list">
                  <el-icon><List /></el-icon>
                  列表
                </el-radio-button>
              </el-radio-group>
            </div>
            
            <div class="sort-controls">
              <el-select v-model="sortBy" placeholder="排序方式" size="small">
                <el-option label="上传时间" value="created_at" />
                <el-option label="文件名" value="original_name" />
                <el-option label="文件大小" value="file_size" />
              </el-select>
              
              <el-button 
                size="small" 
                @click="toggleSortOrder"
                :icon="sortOrder === 'desc' ? SortDown : SortUp"
              >
                {{ sortOrder === 'desc' ? '降序' : '升序' }}
              </el-button>
            </div>
          </div>
        </section>

        <!-- 统计信息 -->
        <section class="stats-section">
          <div class="stats-cards">
            <div class="stat-card">
              <div class="stat-icon">
                <el-icon :size="24" color="#409EFF">
                  <Picture />
                </el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ filteredImages.length }}</div>
                <div class="stat-label">图片总数</div>
              </div>
            </div>
            
            <div class="stat-card">
              <div class="stat-icon">
                <el-icon :size="24" color="#67C23A">
                  <FolderOpened />
                </el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ formatFileSize(totalSize) }}</div>
                <div class="stat-label">总存储量</div>
              </div>
            </div>
            
            <div class="stat-card">
              <div class="stat-icon">
                <el-icon :size="24" color="#E6A23C">
                  <Calendar />
                </el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ todayCount }}</div>
                <div class="stat-label">今日上传</div>
              </div>
            </div>
          </div>
        </section>

        <!-- 图片展示区域 -->
        <section class="images-section">
          <!-- 加载状态 -->
          <div v-if="loading" class="loading-container">
            <el-icon :size="48" class="loading-icon">
              <Loading />
            </el-icon>
            <p>加载中...</p>
          </div>
          
          <!-- 空状态 -->
          <div v-else-if="filteredImages.length === 0" class="empty-state">
            <el-icon :size="64" color="#909399">
              <Picture />
            </el-icon>
            <h3>暂无图片</h3>
            <p v-if="searchQuery">没有找到匹配的图片，请尝试其他关键词</p>
            <p v-else>还没有上传任何图片</p>
            <el-button type="primary" @click="$router.push('/upload')">
              <el-icon><Plus /></el-icon>
              上传图片
            </el-button>
          </div>
          
          <!-- 网格视图 -->
          <div v-else-if="viewMode === 'grid'" class="grid-view">
            <div 
              v-for="image in paginatedImages" 
              :key="image.uuid" 
              class="image-card"
              @click="previewImage(image)"
            >
              <div class="image-container">
                <img 
                  :src="image.public_url" 
                  :alt="image.original_name"
                  @load="handleImageLoad"
                  @error="handleImageError"
                />
                <div class="image-overlay">
                  <div class="overlay-actions">
                    <el-button circle size="small" @click.stop="previewImage(image)">
                      <el-icon><ZoomIn /></el-icon>
                    </el-button>
                    <el-button circle size="small" @click.stop="copyUrl(image.public_url)">
                      <el-icon><DocumentCopy /></el-icon>
                    </el-button>
                    <el-button circle size="small" @click.stop="downloadImage(image)">
                      <el-icon><Download /></el-icon>
                    </el-button>
                  </div>
                </div>
              </div>
              
              <div class="image-info">
                <div class="image-name" :title="image.original_name">
                  {{ image.original_name }}
                </div>
                <div class="image-meta">
                  <span class="file-size">{{ formatFileSize(image.file_size) }}</span>
                  <span class="dimensions">{{ image.width }}×{{ image.height }}</span>
                </div>
                <div class="upload-time">
                  {{ formatTime(image.created_at) }}
                </div>
              </div>
            </div>
          </div>
          
          <!-- 列表视图 -->
          <div v-else class="list-view">
            <el-table 
              :data="paginatedImages" 
              style="width: 100%"
              @row-click="previewImage"
            >
              <el-table-column width="80">
                <template #default="{ row }">
                  <div class="table-image">
                    <img 
                      :src="row.public_url" 
                      :alt="row.original_name"
                      @error="handleImageError"
                    />
                  </div>
                </template>
              </el-table-column>
              
              <el-table-column prop="original_name" label="文件名" min-width="200">
                <template #default="{ row }">
                  <div class="file-info">
                    <div class="file-name">{{ row.original_name }}</div>
                    <div class="file-type">{{ getFileExtension(row.original_name) }}</div>
                  </div>
                </template>
              </el-table-column>
              
              <el-table-column prop="file_size" label="大小" width="100">
                <template #default="{ row }">
                  {{ formatFileSize(row.file_size) }}
                </template>
              </el-table-column>
              
              <el-table-column label="尺寸" width="120">
                <template #default="{ row }">
                  {{ row.width }}×{{ row.height }}
                </template>
              </el-table-column>
              
              <el-table-column prop="created_at" label="上传时间" width="150">
                <template #default="{ row }">
                  {{ formatTime(row.created_at) }}
                </template>
              </el-table-column>
              
              <el-table-column label="操作" width="200" fixed="right">
                <template #default="{ row }">
                  <div class="table-actions">
                    <el-button size="small" @click.stop="previewImage(row)">
                      <el-icon><ZoomIn /></el-icon>
                    </el-button>
                    <el-button size="small" @click.stop="copyUrl(row.public_url)">
                      <el-icon><DocumentCopy /></el-icon>
                    </el-button>
                    <el-button size="small" @click.stop="downloadImage(row)">
                      <el-icon><Download /></el-icon>
                    </el-button>
                  </div>
                </template>
              </el-table-column>
            </el-table>
          </div>
          
          <!-- 分页 -->
          <div v-if="filteredImages.length > pageSize" class="pagination-container">
            <el-pagination
              v-model:current-page="currentPage"
              :page-size="pageSize"
              :total="filteredImages.length"
              layout="total, prev, pager, next, jumper"
              @current-change="handlePageChange"
            />
          </div>
        </section>
      </div>
    </main>

    <!-- 图片预览对话框 -->
    <el-dialog 
      v-model="previewVisible" 
      :title="currentPreviewImage?.original_name" 
      width="90%" 
      center
      class="preview-dialog"
      :modal-class="'preview-overlay'"
      :style="{
        background: '#0f1115',
        color: '#e5e7eb',
        '--el-bg-color': '#0f1115',
        '--el-bg-color-overlay': '#0f1115',
        '--el-fill-color-blank': '#0f1115',
        '--el-dialog-bg-color': '#0f1115',
        '--el-color-white': '#0f1115'
      }"
      append-to-body
      destroy-on-close
      close-on-click-modal
      close-on-press-escape
      lock-scroll
      @closed="onDialogClosed"
    >
      <div class="preview-container">
        <img 
          v-if="currentPreviewImage" 
          :src="currentPreviewImage.public_url" 
          :alt="currentPreviewImage.original_name" 
          class="preview-image"
        />
      </div>
      
      <template #footer>
        <div class="preview-footer">
          <div class="image-details">
            <div class="detail-item">
              <span class="label">文件名：</span>
              <span class="value">{{ currentPreviewImage?.original_name }}</span>
            </div>
            <div class="detail-item">
              <span class="label">大小：</span>
              <span class="value">{{ formatFileSize(currentPreviewImage?.file_size || 0) }}</span>
            </div>
            <div class="detail-item">
              <span class="label">尺寸：</span>
              <span class="value">{{ currentPreviewImage?.width }}×{{ currentPreviewImage?.height }}</span>
            </div>
            <div class="detail-item">
              <span class="label">上传时间：</span>
              <span class="value">{{ formatTime(currentPreviewImage?.created_at || '') }}</span>
            </div>
          </div>
          
          <div class="preview-actions">
            <el-button type="primary" @click="copyUrl(currentPreviewImage?.public_url || '')">
              <el-icon><DocumentCopy /></el-icon>
              复制链接
            </el-button>
            <el-button type="primary" @click="copyMarkdown(currentPreviewImage)">
              <el-icon><Document /></el-icon>
              复制 Markdown
            </el-button>
            <el-button type="primary" @click="downloadImage(currentPreviewImage)">
              <el-icon><Download /></el-icon>
              下载图片
            </el-button>
          </div>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, watch, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  Picture,
  House,
  Upload,
  DataAnalysis,
  Search,
  Grid,
  List,
  SortDown,
  SortUp,
  FolderOpened,
  Calendar,
  Loading,
  Plus,
  ZoomIn,
  DocumentCopy,
  Download,
  Document
} from '@element-plus/icons-vue'
import NavBar from '@/components/NavBar.vue'
import { useUploadStore, type ImageInfo } from '@/stores/upload'
import { listImages } from '@/api/images'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import relativeTime from 'dayjs/plugin/relativeTime'

// 配置 dayjs
dayjs.locale('zh-cn')
dayjs.extend(relativeTime)

const router = useRouter()
const uploadStore = useUploadStore()

// 状态
const loading = ref(false)
const searchQuery = ref('')
const viewMode = ref<'grid' | 'list'>('grid')
const sortBy = ref('created_at')
const sortOrder = ref<'asc' | 'desc'>('desc')
const currentPage = ref(1)
const pageSize = ref(20)
const previewVisible = ref(false)
const currentPreviewImage = ref<ImageInfo | null>(null)
const galleryImages = ref<ImageInfo[]>([])

// 计算属性
const uploadedImages = computed(() => galleryImages.value)

// 过滤和排序后的图片
const filteredImages = computed(() => {
  let images = [...uploadedImages.value]
  
  // 搜索过滤
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    images = images.filter(image => 
      image.original_name.toLowerCase().includes(query)
    )
  }
  
  // 排序
  images.sort((a, b) => {
    let aValue: any = a[sortBy.value as keyof ImageInfo]
    let bValue: any = b[sortBy.value as keyof ImageInfo]
    
    if (sortBy.value === 'created_at') {
      aValue = new Date(aValue).getTime()
      bValue = new Date(bValue).getTime()
    } else if (typeof aValue === 'string') {
      aValue = aValue.toLowerCase()
      bValue = bValue.toLowerCase()
    }
    
    if (sortOrder.value === 'desc') {
      return bValue > aValue ? 1 : -1
    } else {
      return aValue > bValue ? 1 : -1
    }
  })
  
  return images
})

// 分页后的图片
const paginatedImages = computed(() => {
  const start = (currentPage.value - 1) * pageSize.value
  const end = start + pageSize.value
  return filteredImages.value.slice(start, end)
})

// 统计信息
const totalSize = computed(() => {
  return filteredImages.value.reduce((sum, image) => sum + image.file_size, 0)
})

const todayCount = computed(() => {
  const today = dayjs().format('YYYY-MM-DD')
  return filteredImages.value.filter(image => 
    dayjs(image.created_at).format('YYYY-MM-DD') === today
  ).length
})

// 方法
const formatFileSize = (bytes: number): string => {
  return uploadStore.formatFileSize(bytes)
}

const formatTime = (dateString: string): string => {
  return dayjs(dateString).fromNow()
}

const getFileExtension = (filename: string): string => {
  return filename.split('.').pop()?.toUpperCase() || ''
}

const handleSearch = () => {
  currentPage.value = 1
}

const toggleSortOrder = () => {
  sortOrder.value = sortOrder.value === 'desc' ? 'asc' : 'desc'
}

const handlePageChange = (page: number) => {
  currentPage.value = page
  // 滚动到顶部
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const isClosing = ref(false)
const previewImage = (image: ImageInfo) => {
  if (isClosing.value) return
  currentPreviewImage.value = image
  previewVisible.value = true
}
const onDialogClosed = () => {
  // 使用 nextTick 确保 DOM 更新完成后再清理状态
  nextTick(() => {
    currentPreviewImage.value = null
    // 减少防抖时间，避免闪烁
    isClosing.value = true
    setTimeout(() => { isClosing.value = false }, 50)
  })
}

const toAbsolute = (url: string) => (url?.startsWith('http') ? url : `${window.location.origin}${url}`)
const copyUrl = async (url: string) => {
  const abs = toAbsolute(url)
  const success = await uploadStore.copyToClipboard(abs)
  if (success) {
    ElMessage.success('链接已复制到剪贴板')
  } else {
    ElMessage.error('复制失败')
  }
}

const copyMarkdown = async (image: ImageInfo | null) => {
  if (!image) return
  
  const markdown = `![${image.original_name}](${toAbsolute(image.public_url)})`
  const success = await uploadStore.copyToClipboard(markdown)
  if (success) {
    ElMessage.success('Markdown 格式已复制')
  } else {
    ElMessage.error('复制失败')
  }
}

const downloadImage = (image: ImageInfo | null) => {
  if (!image) return
  
  const link = document.createElement('a')
  link.href = toAbsolute(image.public_url)
  link.download = image.original_name
  link.target = '_blank'
  document.body.appendChild(link)
  link.click()
  document.body.removeChild(link)
}

const handleImageLoad = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.style.opacity = '1'
}

const handleImageError = (event: Event) => {
  const img = event.target as HTMLImageElement
  img.src = 'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjAwIiBoZWlnaHQ9IjIwMCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjVmNWY1Ii8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCwgc2Fucy1zZXJpZiIgZm9udC1zaXplPSIxNCIgZmlsbD0iIzk5OSIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPuWKoOi9veWksei0pTwvdGV4dD48L3N2Zz4='
}

// 监听搜索和排序变化，重置页码
watch([searchQuery, sortBy, sortOrder], () => {
  currentPage.value = 1
})

// 组件挂载时加载数据
onMounted(async () => {
  loading.value = true
  try {
    // 从后端拉取画廊数据，独立于上传历史
    const res = await listImages(1, 1000)
    if (res.success && res.data) {
      galleryImages.value = res.data.items || []
    }
    await uploadStore.fetchStats()
  } catch (error) {
    console.error('加载数据失败:', error)
  } finally {
    loading.value = false
  }
})
</script>

<style lang="scss" scoped>
.gallery-container {
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

.gallery-wrapper {
  max-width: 1400px;
  margin: 0 auto;
}

// 页面头部
.header-section {
  margin-bottom: 2rem;

  .title-area {
    text-align: center;
    margin-bottom: 2rem;

    .page-title {
      display: flex;
      align-items: center;
      justify-content: center;
      gap: 0.5rem;
      font-size: 2.5rem;
      font-weight: 700;
      color: white;
      margin-bottom: 0.5rem;
      text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
    }

    .page-description {
      color: rgba(255, 255, 255, 0.7);
      font-size: 1.125rem;
    }
  }

  .toolbar {
    display: flex;
    gap: 1rem;
    align-items: center;
    flex-wrap: wrap;
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    border-radius: 12px;
    padding: 1rem;

    .search-bar {
      flex: 1;
      min-width: 200px;
    }

    .view-controls,
    .sort-controls {
      display: flex;
      gap: 0.5rem;
      align-items: center;
    }
  }
}

// 统计信息
.stats-section {
  margin-bottom: 2rem;

  .stats-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
  }

  .stat-card {
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 12px;
    padding: 1.5rem;
    display: flex;
    align-items: center;
    gap: 1rem;

    .stat-icon {
      width: 48px;
      height: 48px;
      border-radius: 12px;
      background: rgba(255, 255, 255, 0.1);
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .stat-content {
      .stat-value {
        font-size: 1.5rem;
        font-weight: 700;
        color: white;
        margin-bottom: 0.25rem;
      }

      .stat-label {
        color: rgba(255, 255, 255, 0.7);
        font-size: 0.875rem;
      }
    }
  }
}

// 图片展示区域
.images-section {
  .loading-container {
    text-align: center;
    padding: 4rem;
    color: white;

    .loading-icon {
      animation: rotate 2s linear infinite;
      margin-bottom: 1rem;
    }
  }

  .empty-state {
    text-align: center;
    padding: 4rem;
    color: white;

    h3 {
      margin: 1rem 0;
      font-size: 1.5rem;
    }

    p {
      color: rgba(255, 255, 255, 0.7);
      margin-bottom: 2rem;
    }
  }

  // 网格视图
  .grid-view {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: 1.5rem;
    margin-bottom: 2rem;

    .image-card {
      background: rgba(255, 255, 255, 0.1);
      backdrop-filter: blur(10px);
      border: 1px solid rgba(255, 255, 255, 0.2);
      border-radius: 12px;
      overflow: hidden;
      cursor: pointer;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
      transition: box-shadow 0.3s ease;

      &:hover {
        box-shadow: 0 10px 24px rgba(0, 0, 0, 0.28);
      }

      .image-container {
        position: relative;
        height: 200px;
        overflow: hidden;

        img {
          width: 100%;
          height: 100%;
          object-fit: cover;
          opacity: 0;
          transition: opacity 0.3s ease;
          pointer-events: none;
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

          .overlay-actions {
            display: flex;
            gap: 0.5rem;
          }
        }

        /* hover 交给外层 .image-card，避免靠边抖动 */
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

        .upload-time {
          color: rgba(255, 255, 255, 0.5);
          font-size: 0.75rem;
        }
      }
    }
  }

  // 列表视图
  .list-view {
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    border-radius: 12px;
    overflow: hidden;
    margin-bottom: 2rem;

    .table-image {
      width: 60px;
      height: 60px;
      border-radius: 8px;
      overflow: hidden;

      img {
        width: 100%;
        height: 100%;
        object-fit: cover;
      }
    }

    .file-info {
      .file-name {
        font-weight: 600;
        margin-bottom: 0.25rem;
      }

      .file-type {
        font-size: 0.75rem;
        color: rgba(255, 255, 255, 0.7);
        background: rgba(255, 255, 255, 0.1);
        padding: 0.125rem 0.5rem;
        border-radius: 4px;
        display: inline-block;
      }
    }

    .table-actions {
      display: flex;
      gap: 0.5rem;
    }
  }

  // 分页
  .pagination-container {
    display: flex;
    justify-content: center;
    padding: 2rem 0;
  }
}

// 预览对话框
.preview-dialog {
  .preview-container {
    text-align: center;
    max-height: 70vh;
    overflow: auto;

    .preview-image {
      max-width: 100%;
      max-height: 70vh;
      object-fit: contain;
    }
  }

  .preview-footer {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 2rem;

    .image-details {
      flex: 1;
      text-align: left;

      .detail-item {
        margin-bottom: 0.5rem;

        .label {
          font-weight: 600;
          margin-right: 0.5rem;
        }

        .value {
          color: rgba(255, 255, 255, 0.8);
        }
      }
    }

    .preview-actions {
      display: flex;
      gap: 0.5rem;
      flex-shrink: 0;
    }
  }
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

  .toolbar {
    flex-direction: column;
    align-items: stretch;

    .search-bar {
      order: -1;
    }

    .view-controls,
    .sort-controls {
      justify-content: center;
    }
  }

  .grid-view {
    grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  }

  .preview-footer {
    flex-direction: column;
    gap: 1rem;

    .preview-actions {
      justify-content: center;
    }
  }
}
/* 扩大 hover 区域并启用 overlay 点击，避免靠边抖动 */
.grid-view .image-card:hover .image-container .image-overlay {
  opacity: 1;
  pointer-events: auto;
}

/* 预览弹窗暗色主题与遮罩 - 与主题色协调 */
:deep(.preview-overlay) {
  background: rgba(15, 17, 21, 0.95);
  backdrop-filter: blur(8px);
}

:deep(.preview-dialog .el-dialog) {
  background: linear-gradient(135deg, #1a1d29 0%, #2d1b69 100%);
  color: #ffffff;
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.8);
  border-radius: 16px;
}

:deep(.preview-dialog .el-dialog__header) {
  background: rgba(255, 255, 255, 0.05);
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px 16px 0 0;
}

:deep(.preview-dialog .el-dialog__title) {
  color: #ffffff;
  font-weight: 600;
}

:deep(.preview-dialog .el-dialog__headerbtn .el-dialog__close) {
  color: #ffffff;
  transition: all 0.3s ease;
}

:deep(.preview-dialog .el-dialog__headerbtn .el-dialog__close:hover) {
  color: #409eff;
  transform: scale(1.1);
}

:deep(.preview-dialog .el-dialog__body) {
  background: transparent;
  padding: 2rem;
}

:deep(.preview-dialog .el-dialog__footer) {
  background: rgba(255, 255, 255, 0.05);
  border-top: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 0 0 16px 16px;
}

:deep(.preview-dialog) {
  /* 覆盖对话框本体为深色，解决默认白底 */
  background: linear-gradient(135deg, #0f1115 0%, #1a1f2b 60%, #1f2a4a 100%);
  color: #ffffff;
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.8);
  border-radius: 16px;
}

:deep(.preview-dialog .el-dialog__body) {
  background: transparent;
}

:deep(.preview-dialog .el-dialog__header),
:deep(.preview-dialog .el-dialog__footer) {
  background: rgba(255, 255, 255, 0.05);
  border-color: rgba(255, 255, 255, 0.1);
}

/* 彻底去除预览弹窗内部白底（body/滚动容器/自定义容器） */
:deep(.el-dialog.preview-dialog) {
  background: linear-gradient(135deg, #0f1115 0%, #1a1f2b 60%, #1f2a4a 100%) !important;
  color: #fff;
  border: 1px solid rgba(255,255,255,0.08);
}
:deep(.el-dialog.preview-dialog .el-dialog__body),
:deep(.el-dialog.preview-dialog .el-scrollbar__wrap),
:deep(.el-dialog.preview-dialog .el-scrollbar__view),
:deep(.el-dialog.preview-dialog .preview-container) {
  background: transparent !important;
}
:deep(.el-dialog.preview-dialog .el-dialog__header),
:deep(.el-dialog.preview-dialog .el-dialog__footer) {
  background: rgba(255,255,255,0.05) !important;
  border-color: rgba(255,255,255,0.1) !important;
}

/* 预览图片区域也统一深色，以免图片边缘看见白色留白 */
:deep(.el-dialog.preview-dialog .preview-container) {
  padding: 0; /* 减少白边观感 */
}
:deep(.el-dialog.preview-dialog .preview-image) {
  display: block;
  margin: 0 auto;
  background: transparent !important;
}

/* 统一在预览对话框内切换到深色主题，并覆盖默认按钮为深色 */
:deep(.el-dialog.preview-dialog) {
  /* 深色背景与文本 */
  background: #0f1115 !important;
  color: #e5e7eb !important;

  /* Element Plus 主题变量在对话框作用域内覆盖 */
  --el-bg-color: #0f1115;
  --el-bg-color-overlay: #0f1115;
  --el-text-color-primary: #ffffff;
  --el-text-color-regular: #e5e7eb;
  --el-border-color: rgba(255,255,255,0.12);
  --el-border-color-lighter: rgba(255,255,255,0.12);
}

/* 去除 body/滚动容器白底 */
:deep(.el-dialog.preview-dialog .el-dialog__body),
:deep(.el-dialog.preview-dialog .el-scrollbar__wrap),
:deep(.el-dialog.preview-dialog .el-scrollbar__view),
:deep(.el-dialog.preview-dialog .preview-container) {
  background: transparent !important;
}

/* header/footer 亦为深色 */
:deep(.el-dialog.preview-dialog .el-dialog__header),
:deep(.el-dialog.preview-dialog .el-dialog__footer) {
  background: rgba(255,255,255,0.05) !important;
  border-color: rgba(255,255,255,0.1) !important;
}

/* 兜底：将默认型按钮在该弹窗内统一为深色风格，避免出现白底按钮 */
:deep(.el-dialog.preview-dialog .el-button:not(.el-button--primary):not(.is-link):not(.is-text)) {
  background-color: #1f2a4a !important;
  border-color: #334155 !important;
  color: #e5e7eb !important;
}
:deep(.el-dialog.preview-dialog .el-button:not(.el-button--primary):not(.is-link):not(.is-text):hover) {
  background-color: #2b3a63 !important;
  border-color: #475569 !important;
}

/* primary 按钮在暗色下的悬停反馈 */
:deep(.el-dialog.preview-dialog .el-button.el-button--primary:hover) {
  filter: brightness(1.05);
}

/* 彻底兜底：同时覆盖三种层级，消除预览内部白底 */
:deep(.el-dialog.preview-dialog),
:deep(.preview-dialog .el-dialog),
:deep(.el-overlay-dialog.preview-dialog .el-dialog) {
  background: #0f1115 !important;
  color: #e5e7eb !important;
  --el-bg-color: #0f1115;
  --el-bg-color-overlay: #0f1115;
  --el-text-color-primary: #ffffff;
  --el-text-color-regular: #e5e7eb;
  --el-border-color: rgba(255,255,255,0.12);
  --el-border-color-lighter: rgba(255,255,255,0.12);
}

/* body/滚动容器必须也改为深色，避免内部仍是白色 */
:deep(.el-dialog.preview-dialog .el-dialog__body),
:deep(.preview-dialog .el-dialog .el-dialog__body),
:deep(.el-overlay-dialog.preview-dialog .el-dialog .el-dialog__body),
:deep(.preview-dialog .el-scrollbar__wrap),
:deep(.preview-dialog .el-scrollbar__view) {
  background: #0f1115 !important;
}

/* 预览容器本身也给深色，避免透明时透出白底 */
:deep(.preview-dialog .preview-container) {
  background: #0f1115 !important;
}

/* 进一步覆盖 Element Plus 白色系变量，彻底去除预览内部白底 */
:deep(.el-dialog.preview-dialog),
:deep(.preview-dialog .el-dialog),
:deep(.el-overlay-dialog.preview-dialog .el-dialog) {
  /* 统一暗色变量（关键：这两个常导致白底） */
  --el-bg-color: #0f1115 !important;
  --el-fill-color-blank: #0f1115 !important;
  --el-color-white: #0f1115 !important;
  --el-fill-color-light: #111827 !important;
  --el-dialog-bg-color: #0f1115 !important;
  --el-text-color-primary: #ffffff !important;
  --el-text-color-regular: #e5e7eb !important;
  --el-border-color: rgba(255,255,255,0.12) !important;
}

/* body/滚动容器强制暗色（避免变量未生效时的兜底） */
:deep(.el-dialog.preview-dialog .el-dialog__body),
:deep(.preview-dialog .el-dialog .el-dialog__body),
:deep(.el-overlay-dialog.preview-dialog .el-dialog .el-dialog__body),
:deep(.preview-dialog .el-scrollbar__wrap),
:deep(.preview-dialog .el-scrollbar__view),
:deep(.preview-dialog .preview-container) {
  background: #0f1115 !important;
}

/* header/footer 同步暗色，避免边框/底色反白 */
:deep(.el-dialog.preview-dialog .el-dialog__header),
:deep(.el-dialog.preview-dialog .el-dialog__footer) {
  background: #111827 !important;
  border-color: rgba(255,255,255,0.1) !important;
}

</style>