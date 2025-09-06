<template>
  <div class="storage-container">
    <NavBar />

    <main class="main-content">
      <div class="page-wrapper">
        <section class="header-section">
          <div class="title-area">
            <h1 class="page-title">
              <el-icon><FolderOpened /></el-icon>
              存储内容
            </h1>
            <p class="page-description">查看并管理当前 VPS 上存储的图片，可删除释放空间</p>
          </div>
          <div class="actions">
            <el-input v-model="search" placeholder="搜索文件名..." clearable :prefix-icon="Search" class="search" />
            <el-button :loading="loading" @click="refresh" type="primary">
              <el-icon><Refresh /></el-icon>
              刷新
            </el-button>
          </div>
        </section>

        <section class="table-section">
          <el-table
            :data="pagedItems"
            v-loading="loading"
            style="width: 100%"
            :style="tableStyle"
            :cell-style="forceTransparentCellStyle"
            :header-cell-style="forceTransparentCellStyle"
            :row-style="forceTransparentRowStyle"
            :row-class-name="'no-hover-row'"
          >
            <el-table-column label="预览" width="90">
              <template #default="{ row }">
                <div class="thumb">
                  <img :src="toAbsolute(row.public_url)" :alt="row.original_name" @error="onImgError" />
                </div>
              </template>
            </el-table-column>

            <el-table-column label="文件名" min-width="240">
              <template #default="{ row }">
                <div class="file-info">
                  <div class="name" :title="row.original_name">{{ row.original_name }}</div>
                  <div class="meta">{{ row.mime_type.toUpperCase() }}</div>
                </div>
              </template>
            </el-table-column>

            <el-table-column label="大小" width="120">
              <template #default="{ row }">
                {{ formatFileSize(row.file_size) }}
              </template>
            </el-table-column>

            <el-table-column label="上传时间" width="180">
              <template #default="{ row }">
                {{ formatTime(row.created_at) }}
              </template>
            </el-table-column>

            <el-table-column label="操作" width="300" fixed="right">
              <template #default="{ row }">
                <div class="ops">
                  <el-button size="small" @click="copyLink(row)">
                    <el-icon><DocumentCopy /></el-icon>
                    复制链接
                  </el-button>
                  <el-button size="small" @click="download(row)">
                    <el-icon><Download /></el-icon>
                    下载
                  </el-button>
                  <el-popconfirm
                    title="确定要删除该图片吗？此操作不可恢复"
                    confirm-button-text="删除"
                    cancel-button-text="取消"
                    confirm-button-type="danger"
                    @confirm="remove(row)"
                  >
                    <template #reference>
                      <el-button size="small" type="danger">
                        <el-icon><Delete /></el-icon>
                        删除
                      </el-button>
                    </template>
                  </el-popconfirm>
                </div>
              </template>
            </el-table-column>
          </el-table>

          <div class="pagination" v-if="total > pageSize">
            <el-pagination
              v-model:current-page="page"
              :page-size="pageSize"
              :total="total"
              layout="total, prev, pager, next, jumper"
              @current-change="handlePageChange"
            />
          </div>
        </section>
      </div>
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { FolderOpened, Refresh, Search, DocumentCopy, Download, Delete } from '@element-plus/icons-vue'
import NavBar from '@/components/NavBar.vue'
import { listImages, deleteImage } from '@/api/images'
import dayjs from 'dayjs'
import 'dayjs/locale/zh-cn'
import relativeTime from 'dayjs/plugin/relativeTime'
import type { ImageInfo } from '@/stores/upload'

dayjs.locale('zh-cn')
dayjs.extend(relativeTime)

const loading = ref(false)
const items = ref<ImageInfo[]>([])
const total = ref(0)
const page = ref(1)
const pageSize = ref(20)
const search = ref('')

const filtered = computed(() => {
  if (!search.value) return items.value
  const q = search.value.toLowerCase()
  return items.value.filter(i => i.original_name.toLowerCase().includes(q))
})
const pagedItems = computed(() => {
  const start = (page.value - 1) * pageSize.value
  return filtered.value.slice(start, start + pageSize.value)
})

const tableStyle = computed(() => ({
  '--el-table-row-hover-bg-color': 'transparent',
  '--el-table-current-row-bg-color': 'transparent',
  '--el-table-bg-color': 'transparent',
  '--el-table-header-bg-color': 'transparent'
}))
const forceTransparentCellStyle = () => ({ backgroundColor: 'transparent' })
const forceTransparentRowStyle = () => ({ backgroundColor: 'transparent' })

const formatFileSize = (bytes: number): string => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}
const formatTime = (date: string) => dayjs(date).fromNow()

const toAbsolute = (url: string) => (url?.startsWith('http') ? url : `${window.location.origin}${url}`)

const fetchList = async () => {
  loading.value = true
  try {
    const res = await listImages(page.value, pageSize.value)
    if (res.success && res.data) {
      items.value = res.data.items
      total.value = res.data.total
    } else {
      ElMessage.error(res.error || '获取列表失败')
    }
  } finally {
    loading.value = false
  }
}

const refresh = async () => {
  await fetchList()
}

const handlePageChange = async (p: number) => {
  page.value = p
  await fetchList()
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

const copyLink = async (row: ImageInfo) => {
  const abs = toAbsolute(row.public_url)
  try {
    await navigator.clipboard.writeText(abs)
    ElMessage.success('已复制链接')
  } catch {
    ElMessage.error('复制失败')
  }
}

const download = (row: ImageInfo) => {
  const a = document.createElement('a')
  a.href = toAbsolute(row.public_url)
  a.download = row.original_name
  a.target = '_blank'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}

const remove = async (row: ImageInfo) => {
  const res = await deleteImage(row.uuid)
  if (res.success) {
    ElMessage.success('删除成功')
    await fetchList()
  } else {
    ElMessage.error(res.error || '删除失败')
  }
}

const onImgError = (e: Event) => {
  const img = e.target as HTMLImageElement
  img.src =
    'data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iOTAiIGhlaWdodD0iOTAiIHhtbG5zPSJodHRwOi8vd3d3LnczLm9yZy8yMDAwL3N2ZyI+PHJlY3Qgd2lkdGg9IjEwMCUiIGhlaWdodD0iMTAwJSIgZmlsbD0iI2VlZWVlZSIvPjx0ZXh0IHg9IjUwJSIgeT0iNTAlIiBmb250LXNpemU9IjEyIiB0ZXh0LWFuY2hvcj0ibWlkZGxlIiBmaWxsPSIjOTk5Ij7pooTmnKwiPC90ZXh0Pjwvc3ZnPg=='
}

onMounted(fetchList)
</script>

<style lang="scss" scoped>
.storage-container {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.main-content {
  flex: 1;
  padding: 2rem;
}
.page-wrapper {
  max-width: 1400px;
  margin: 0 auto;
}
.header-section {
  display: flex;
  justify-content: space-between;
  gap: 1rem;
  align-items: center;
  margin-bottom: 1.5rem;

  .page-title {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-size: 2rem;
    font-weight: 700;
    color: white;
    margin: 0;
    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.3);
  }

  .page-description {
    color: rgba(255, 255, 255, 0.75);
    margin-top: 0.25rem;
  }

  .actions {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    .search {
      width: 240px;
    }
  }
}

.table-section {
  background: rgba(255, 255, 255, 0.08);
  /* backdrop-filter removed to avoid flicker on hover */
  border: 1px solid rgba(255, 255, 255, 0.15);
  border-radius: 12px;
  padding: 1rem;

  .thumb {
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
    .name {
      color: white;
      font-weight: 600;
      margin-bottom: 0.25rem;
    }
    .meta {
      color: rgba(255, 255, 255, 0.7);
      font-size: 0.75rem;
    }
  }
  .ops {
    display: flex;
    gap: 8px;
    flex-wrap: nowrap;
    white-space: nowrap;
    overflow: visible;
  }
  .ops :deep(.el-button) {
    padding: 6px 8px;
    box-sizing: border-box;
    min-height: 28px;
  }
  .ops :deep(.el-button:hover),
  .ops :deep(.el-button:focus),
  .ops :deep(.el-button:active) {
    transform: none !important;
    box-shadow: none !important;
  }
  .ops :deep(.el-button.is-plain:hover),
  .ops :deep(.el-button:hover) {
    background-color: rgba(255, 255, 255, 0.08) !important;
    border-color: rgba(255, 255, 255, 0.25) !important;
  }
}

.pagination {
  display: flex;
  justify-content: center;
  padding: 1rem 0 0.5rem;
}

@media (max-width: 768px) {
  .main-content {
    padding: 1rem;
  }
}

/* Stabilize table layout and prevent hover flicker */
.table-section {
  /* contain: paint; */
}
.table-section .thumb,
.table-section .thumb img {
  will-change: transform, opacity;
  transform: translateZ(0);
  backface-visibility: hidden;
}
.el-table .cell {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.el-table__row {
  height: auto;
}
.el-table__header,
.el-table__body {
  table-layout: fixed !important;
}
.table-section .ops {
  white-space: nowrap;
  overflow: visible;
}
:deep(.el-table) {
  --el-table-row-hover-bg-color: transparent;
  --el-table-current-row-bg-color: rgba(255, 255, 255, 0.03);
  --el-table-bg-color: transparent;
  --el-table-header-bg-color: transparent;
}
:deep(.el-table__row > td) {
  background-color: transparent !important;
  transition: none !important;
}
:deep(.el-table__row:hover > td) {
  background-color: transparent !important;
  transition: none !important;
}
/* 强制覆盖所有行 hover 高亮（提升优先级，适配不同实现） */
:deep(.el-table--enable-row-hover .el-table__body tr:hover > td),
:deep(.el-table__body tr.hover-row > td),
:deep(.el-table__body tr:hover > td) {
  background-color: transparent !important;
}
/* 允许右侧固定列内容不被裁剪（Popconfirm/按钮完全可见） */
:deep(.el-table__fixed-right),
:deep(.el-table__fixed-right .el-table__body-wrapper) {
  overflow: visible !important;
  z-index: 2;
}
:deep(.el-table td.el-table__cell) {
  padding: 12px 8px !important;
  vertical-align: middle;
  white-space: nowrap;
}
.table-section .ops {
  position: relative;
  z-index: 3;
  display: inline-flex;
  align-items: center;
}
:deep(.el-popper) {
  z-index: 3000 !important;
}

/* 强力覆盖：任何情况下都不让表格行/列在 hover 或选中时变白 */
:deep(.el-table__inner-wrapper),
:deep(.el-table__body),
:deep(.el-table__body-wrapper),
:deep(.el-table__fixed),
:deep(.el-table__fixed-right),
:deep(.el-table__fixed-right .el-table__fixed-body-wrapper) {
  background: transparent !important;
}

:deep(.el-table__body tr > td),
:deep(.el-table__body tr:hover > td),
:deep(.el-table__body tr.hover-row > td),
:deep(.el-table__row.current-row > td) {
  background-color: transparent !important;
  transition: background-color 0s !important;
}

/* 稳定操作按钮：去除过渡，避免 hover 轻微位移 */
.table-section .ops :deep(.el-button) {
  transition: none !important;
  border-width: 1px !important;
}

/* 避免单元格内内容因hover产生重排 */
:deep(.el-table__body .el-table__cell) {
  white-space: nowrap;
}

/* 为每行加的类名，彻底禁用任何悬停/选中背景 */
:deep(.no-hover-row > td) {
  background-color: transparent !important;
  transition: none !important;
}

/* 防止内容/弹窗被裁剪 */
:deep(.el-table__body-wrapper),
:deep(.el-table__inner-wrapper) {
  overflow: visible !important;
}

</style>