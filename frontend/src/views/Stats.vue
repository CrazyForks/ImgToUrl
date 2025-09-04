<template>
  <div class="stats-container">
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
          <router-link to="/upload" class="nav-link">
            <el-icon><Upload /></el-icon>
            上传图片
          </router-link>
          <router-link to="/gallery" class="nav-link">
            <el-icon><Picture /></el-icon>
            图片画廊
          </router-link>
        </div>
      </div>
    </nav>

    <!-- 主要内容 -->
    <main class="main-content">
      <div class="stats-wrapper">
        <!-- 页面标题 -->
        <section class="header-section">
          <div class="title-area">
            <h1 class="page-title">
              <el-icon><DataAnalysis /></el-icon>
              统计信息
            </h1>
            <p class="page-description">
              查看系统使用情况和数据分析
            </p>
          </div>
          
          <div class="refresh-area">
            <el-button 
              type="primary" 
              :loading="loading" 
              @click="refreshStats"
            >
              <el-icon><Refresh /></el-icon>
              刷新数据
            </el-button>
          </div>
        </section>

        <!-- 概览统计卡片 -->
        <section class="overview-section">
          <div class="stats-grid">
            <div class="stat-card primary">
              <div class="stat-icon">
                <el-icon :size="32" color="#409EFF">
                  <Picture />
                </el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ stats.total_images || 0 }}</div>
                <div class="stat-label">图片总数</div>
                <div class="stat-change" :class="{ positive: todayUploads > 0 }">
                  <el-icon><TrendCharts /></el-icon>
                  今日 +{{ todayUploads }}
                </div>
              </div>
            </div>
            
            <div class="stat-card success">
              <div class="stat-icon">
                <el-icon :size="32" color="#67C23A">
                  <FolderOpened />
                </el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ formatFileSize(stats.total_size || 0) }}</div>
                <div class="stat-label">总存储量</div>
                <div class="stat-change">
                  <el-icon><DataBoard /></el-icon>
                  平均 {{ formatFileSize(averageFileSize) }}/张
                </div>
              </div>
            </div>
            
            <div class="stat-card warning">
              <div class="stat-icon">
                <el-icon :size="32" color="#E6A23C">
                  <Calendar />
                </el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ todayUploads }}</div>
                <div class="stat-label">今日上传</div>
                <div class="stat-change">
                  <el-icon><Clock /></el-icon>
                  {{ formatFileSize(todaySize) }}
                </div>
              </div>
            </div>
            
            <div class="stat-card info">
              <div class="stat-icon">
                <el-icon :size="32" color="#909399">
                  <Monitor />
                </el-icon>
              </div>
              <div class="stat-content">
                <div class="stat-value">{{ systemUptime }}</div>
                <div class="stat-label">系统运行</div>
                <div class="stat-change">
                  <el-icon><CircleCheck /></el-icon>
                  状态正常
                </div>
              </div>
            </div>
          </div>
        </section>

        <!-- 详细统计 -->
        <section class="details-section">
          <div class="section-grid">
            <!-- 文件类型分布 -->
            <div class="detail-card">
              <div class="card-header">
                <h3 class="card-title">
                  <el-icon><PieChart /></el-icon>
                  文件类型分布
                </h3>
              </div>
              <div class="card-content">
                <div v-if="fileTypeStats.length === 0" class="empty-state">
                  <el-icon :size="48" color="#909399">
                    <Document />
                  </el-icon>
                  <p>暂无数据</p>
                </div>
                <div v-else class="file-type-list">
                  <div 
                    v-for="(item, index) in fileTypeStats" 
                    :key="item.type" 
                    class="file-type-item"
                  >
                    <div class="type-info">
                      <div class="type-icon" :style="{ backgroundColor: getTypeColor(index) }">
                        {{ item.type.toUpperCase() }}
                      </div>
                      <div class="type-details">
                        <div class="type-name">{{ item.type.toUpperCase() }} 格式</div>
                        <div class="type-count">{{ item.count }} 个文件</div>
                      </div>
                    </div>
                    <div class="type-stats">
                      <div class="percentage">{{ item.percentage }}%</div>
                      <div class="progress-bar">
                        <div 
                          class="progress-fill" 
                          :style="{ 
                            width: item.percentage + '%',
                            backgroundColor: getTypeColor(index)
                          }"
                        ></div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 上传趋势 -->
            <div class="detail-card">
              <div class="card-header">
                <h3 class="card-title">
                  <el-icon><TrendCharts /></el-icon>
                  上传趋势 (最近7天)
                </h3>
              </div>
              <div class="card-content">
                <div v-if="uploadTrend.length === 0" class="empty-state">
                  <el-icon :size="48" color="#909399">
                    <DataLine />
                  </el-icon>
                  <p>暂无数据</p>
                </div>
                <div v-else class="trend-chart">
                  <div class="chart-container">
                    <div 
                      v-for="(day, index) in uploadTrend" 
                      :key="day.date" 
                      class="chart-bar"
                    >
                      <div class="bar-container">
                        <div 
                          class="bar" 
                          :style="{ 
                            height: (day.count / maxDayCount * 100) + '%',
                            backgroundColor: getBarColor(day.count, maxDayCount)
                          }"
                        ></div>
                      </div>
                      <div class="bar-label">
                        <div class="day-name">{{ day.dayName }}</div>
                        <div class="day-count">{{ day.count }}</div>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 存储使用情况 -->
            <div class="detail-card">
              <div class="card-header">
                <h3 class="card-title">
                  <el-icon><DataBoard /></el-icon>
                  存储使用情况
                </h3>
              </div>
              <div class="card-content">
                <div class="storage-info">
                  <div class="storage-item">
                    <div class="storage-label">已使用存储</div>
                    <div class="storage-value">{{ formatFileSize(stats.total_size || 0) }}</div>
                  </div>
                  
                  <div class="storage-progress">
                    <div class="progress-info">
                      <span>存储使用率</span>
                      <span>{{ storageUsagePercentage }}%</span>
                    </div>
                    <el-progress 
                      :percentage="storageUsagePercentage" 
                      :stroke-width="12"
                      :color="getStorageColor(storageUsagePercentage)"
                    />
                  </div>
                  
                  <div class="storage-details">
                    <div class="detail-row">
                      <span class="detail-label">平均文件大小</span>
                      <span class="detail-value">{{ formatFileSize(averageFileSize) }}</span>
                    </div>
                    <div class="detail-row">
                      <span class="detail-label">最大文件大小</span>
                      <span class="detail-value">{{ formatFileSize(maxFileSize) }}</span>
                    </div>
                    <div class="detail-row">
                      <span class="detail-label">最小文件大小</span>
                      <span class="detail-value">{{ formatFileSize(minFileSize) }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>

            <!-- 系统信息 -->
            <div class="detail-card">
              <div class="card-header">
                <h3 class="card-title">
                  <el-icon><Monitor /></el-icon>
                  系统信息
                </h3>
              </div>
              <div class="card-content">
                <div class="system-info">
                  <div class="info-item">
                    <div class="info-icon">
                      <el-icon color="#67C23A"><CircleCheck /></el-icon>
                    </div>
                    <div class="info-content">
                      <div class="info-label">服务状态</div>
                      <div class="info-value status-online">在线</div>
                    </div>
                  </div>
                  
                  <div class="info-item">
                    <div class="info-icon">
                      <el-icon color="#409EFF"><Clock /></el-icon>
                    </div>
                    <div class="info-content">
                      <div class="info-label">运行时间</div>
                      <div class="info-value">{{ systemUptime }}</div>
                    </div>
                  </div>
                  
                  <div class="info-item">
                    <div class="info-icon">
                      <el-icon color="#E6A23C"><DataBoard /></el-icon>
                    </div>
                    <div class="info-content">
                      <div class="info-label">API 版本</div>
                      <div class="info-value">v1.0.0</div>
                    </div>
                  </div>
                  
                  <div class="info-item">
                    <div class="info-icon">
                      <el-icon color="#F56C6C"><Connection /></el-icon>
                    </div>
                    <div class="info-content">
                      <div class="info-label">数据库</div>
                      <div class="info-value status-online">已连接</div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import {
  Picture,
  House,
  Upload,
  DataAnalysis,
  Refresh,
  TrendCharts,
  FolderOpened,
  Calendar,
  Monitor,
  Clock,
  DataBoard,
  CircleCheck,
  PieChart,
  Document,
  DataLine,
  Connection
} from '@element-plus/icons-vue'
import { useUploadStore } from '@/stores/upload'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import relativeTime from 'dayjs/plugin/relativeTime'

// 配置 dayjs
dayjs.locale('zh-cn')
dayjs.extend(relativeTime)

const uploadStore = useUploadStore()

// 状态
const loading = ref(false)
const systemStartTime = ref(new Date())

// 计算属性
const stats = computed(() => uploadStore.stats)
const uploadedImages = computed(() => uploadStore.uploadedImages)

// 今日上传统计
const todayUploads = computed(() => {
  const today = dayjs().format('YYYY-MM-DD')
  return uploadedImages.value.filter(image => 
    dayjs(image.created_at).format('YYYY-MM-DD') === today
  ).length
})

const todaySize = computed(() => {
  const today = dayjs().format('YYYY-MM-DD')
  return uploadedImages.value
    .filter(image => dayjs(image.created_at).format('YYYY-MM-DD') === today)
    .reduce((sum, image) => sum + image.file_size, 0)
})

// 文件大小统计
const averageFileSize = computed(() => {
  const images = uploadedImages.value
  if (images.length === 0) return 0
  return Math.round(images.reduce((sum, image) => sum + image.file_size, 0) / images.length)
})

const maxFileSize = computed(() => {
  const images = uploadedImages.value
  if (images.length === 0) return 0
  return Math.max(...images.map(image => image.file_size))
})

const minFileSize = computed(() => {
  const images = uploadedImages.value
  if (images.length === 0) return 0
  return Math.min(...images.map(image => image.file_size))
})

// 存储使用率 (假设总容量为 10GB)
const storageUsagePercentage = computed(() => {
  const totalCapacity = 10 * 1024 * 1024 * 1024 // 10GB
  const usedSpace = stats.value.total_size || 0
  return Math.min(Math.round((usedSpace / totalCapacity) * 100), 100)
})

// 系统运行时间
const systemUptime = computed(() => {
  return dayjs(systemStartTime.value).fromNow(true)
})

// 文件类型统计
const fileTypeStats = computed(() => {
  const images = uploadedImages.value
  if (images.length === 0) return []
  
  const typeCount: Record<string, number> = {}
  
  images.forEach(image => {
    const extension = image.original_name.split('.').pop()?.toLowerCase() || 'unknown'
    typeCount[extension] = (typeCount[extension] || 0) + 1
  })
  
  const total = images.length
  return Object.entries(typeCount)
    .map(([type, count]) => ({
      type,
      count,
      percentage: Math.round((count / total) * 100)
    }))
    .sort((a, b) => b.count - a.count)
})

// 上传趋势 (最近7天)
const uploadTrend = computed(() => {
  const days = []
  for (let i = 6; i >= 0; i--) {
    const date = dayjs().subtract(i, 'day')
    const dateStr = date.format('YYYY-MM-DD')
    const dayName = date.format('MM/DD')
    
    const count = uploadedImages.value.filter(image => 
      dayjs(image.created_at).format('YYYY-MM-DD') === dateStr
    ).length
    
    days.push({
      date: dateStr,
      dayName,
      count
    })
  }
  return days
})

const maxDayCount = computed(() => {
  return Math.max(...uploadTrend.value.map(day => day.count), 1)
})

// 方法
const formatFileSize = (bytes: number): string => {
  return uploadStore.formatFileSize(bytes)
}

const refreshStats = async () => {
  loading.value = true
  try {
    await uploadStore.fetchStats()
    ElMessage.success('数据已刷新')
  } catch (error) {
    ElMessage.error('刷新失败')
  } finally {
    loading.value = false
  }
}

const getTypeColor = (index: number): string => {
  const colors = ['#409EFF', '#67C23A', '#E6A23C', '#F56C6C', '#909399']
  return colors[index % colors.length]
}

const getBarColor = (count: number, maxCount: number): string => {
  const ratio = count / maxCount
  if (ratio > 0.8) return '#67C23A'
  if (ratio > 0.5) return '#E6A23C'
  if (ratio > 0.2) return '#409EFF'
  return '#909399'
}

const getStorageColor = (percentage: number): string => {
  if (percentage > 80) return '#F56C6C'
  if (percentage > 60) return '#E6A23C'
  return '#67C23A'
}

// 组件挂载时加载数据
onMounted(async () => {
  loading.value = true
  try {
    await uploadStore.fetchStats()
  } catch (error) {
    console.error('加载统计数据失败:', error)
  } finally {
    loading.value = false
  }
})
</script>

<style lang="scss" scoped>
.stats-container {
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

.stats-wrapper {
  max-width: 1400px;
  margin: 0 auto;
}

// 页面头部
.header-section {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 2rem;

  .title-area {
    .page-title {
      display: flex;
      align-items: center;
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
}

// 概览统计
.overview-section {
  margin-bottom: 3rem;

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 1.5rem;
  }

  .stat-card {
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 16px;
    padding: 2rem;
    display: flex;
    align-items: center;
    gap: 1.5rem;
    transition: transform 0.3s ease;

    &:hover {
      transform: translateY(-4px);
    }

    &.primary {
      border-color: rgba(64, 158, 255, 0.3);
    }

    &.success {
      border-color: rgba(103, 194, 58, 0.3);
    }

    &.warning {
      border-color: rgba(230, 162, 60, 0.3);
    }

    &.info {
      border-color: rgba(144, 147, 153, 0.3);
    }

    .stat-icon {
      width: 64px;
      height: 64px;
      border-radius: 16px;
      background: rgba(255, 255, 255, 0.1);
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .stat-content {
      flex: 1;

      .stat-value {
        font-size: 2rem;
        font-weight: 700;
        color: white;
        margin-bottom: 0.5rem;
      }

      .stat-label {
        color: rgba(255, 255, 255, 0.7);
        font-size: 1rem;
        margin-bottom: 0.5rem;
      }

      .stat-change {
        display: flex;
        align-items: center;
        gap: 0.25rem;
        color: rgba(255, 255, 255, 0.6);
        font-size: 0.875rem;

        &.positive {
          color: #67C23A;
        }
      }
    }
  }
}

// 详细统计
.details-section {
  .section-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
    gap: 2rem;
  }

  .detail-card {
    background: rgba(255, 255, 255, 0.1);
    backdrop-filter: blur(10px);
    border: 1px solid rgba(255, 255, 255, 0.2);
    border-radius: 16px;
    overflow: hidden;

    .card-header {
      padding: 1.5rem;
      border-bottom: 1px solid rgba(255, 255, 255, 0.1);

      .card-title {
        display: flex;
        align-items: center;
        gap: 0.5rem;
        font-size: 1.25rem;
        font-weight: 600;
        color: white;
        margin: 0;
      }
    }

    .card-content {
      padding: 1.5rem;
    }

    .empty-state {
      text-align: center;
      padding: 2rem;
      color: rgba(255, 255, 255, 0.7);

      p {
        margin-top: 1rem;
      }
    }
  }
}

// 文件类型分布
.file-type-list {
  .file-type-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);

    &:last-child {
      border-bottom: none;
    }

    .type-info {
      display: flex;
      align-items: center;
      gap: 1rem;

      .type-icon {
        width: 40px;
        height: 40px;
        border-radius: 8px;
        display: flex;
        align-items: center;
        justify-content: center;
        color: white;
        font-size: 0.75rem;
        font-weight: 600;
      }

      .type-details {
        .type-name {
          color: white;
          font-weight: 600;
          margin-bottom: 0.25rem;
        }

        .type-count {
          color: rgba(255, 255, 255, 0.7);
          font-size: 0.875rem;
        }
      }
    }

    .type-stats {
      text-align: right;
      min-width: 80px;

      .percentage {
        color: white;
        font-weight: 600;
        margin-bottom: 0.5rem;
      }

      .progress-bar {
        width: 60px;
        height: 4px;
        background: rgba(255, 255, 255, 0.2);
        border-radius: 2px;
        overflow: hidden;

        .progress-fill {
          height: 100%;
          border-radius: 2px;
          transition: width 0.3s ease;
        }
      }
    }
  }
}

// 上传趋势图表
.trend-chart {
  .chart-container {
    display: flex;
    justify-content: space-between;
    align-items: end;
    height: 200px;
    padding: 1rem 0;

    .chart-bar {
      flex: 1;
      display: flex;
      flex-direction: column;
      align-items: center;
      gap: 0.5rem;

      .bar-container {
        height: 150px;
        width: 24px;
        display: flex;
        align-items: end;

        .bar {
          width: 100%;
          min-height: 4px;
          border-radius: 2px 2px 0 0;
          transition: height 0.3s ease;
        }
      }

      .bar-label {
        text-align: center;

        .day-name {
          color: rgba(255, 255, 255, 0.7);
          font-size: 0.75rem;
          margin-bottom: 0.25rem;
        }

        .day-count {
          color: white;
          font-weight: 600;
          font-size: 0.875rem;
        }
      }
    }
  }
}

// 存储信息
.storage-info {
  .storage-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;

    .storage-label {
      color: rgba(255, 255, 255, 0.7);
    }

    .storage-value {
      color: white;
      font-weight: 600;
      font-size: 1.125rem;
    }
  }

  .storage-progress {
    margin-bottom: 1.5rem;

    .progress-info {
      display: flex;
      justify-content: space-between;
      margin-bottom: 0.5rem;
      color: rgba(255, 255, 255, 0.7);
      font-size: 0.875rem;
    }
  }

  .storage-details {
    .detail-row {
      display: flex;
      justify-content: space-between;
      padding: 0.5rem 0;
      border-bottom: 1px solid rgba(255, 255, 255, 0.1);

      &:last-child {
        border-bottom: none;
      }

      .detail-label {
        color: rgba(255, 255, 255, 0.7);
      }

      .detail-value {
        color: white;
        font-weight: 500;
      }
    }
  }
}

// 系统信息
.system-info {
  .info-item {
    display: flex;
    align-items: center;
    gap: 1rem;
    padding: 1rem 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);

    &:last-child {
      border-bottom: none;
    }

    .info-icon {
      width: 40px;
      height: 40px;
      border-radius: 8px;
      background: rgba(255, 255, 255, 0.1);
      display: flex;
      align-items: center;
      justify-content: center;
    }

    .info-content {
      flex: 1;

      .info-label {
        color: rgba(255, 255, 255, 0.7);
        margin-bottom: 0.25rem;
      }

      .info-value {
        color: white;
        font-weight: 600;

        &.status-online {
          color: #67C23A;
        }
      }
    }
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

  .header-section {
    flex-direction: column;
    gap: 1rem;
    align-items: flex-start;
  }

  .stats-grid {
    grid-template-columns: 1fr;
  }

  .section-grid {
    grid-template-columns: 1fr;
  }

  .stat-card {
    padding: 1.5rem;

    .stat-icon {
      width: 48px;
      height: 48px;
    }

    .stat-value {
      font-size: 1.5rem;
    }
  }

  .chart-container {
    height: 150px;

    .bar-container {
      height: 100px;
      width: 20px;
    }
  }
}
</style>