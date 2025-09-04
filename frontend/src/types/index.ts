// 图片上传相关类型定义
export interface UploadedImage {
  id: string
  filename: string
  originalName: string
  url: string
  size: number
  mimeType: string
  uploadTime: string
  width?: number
  height?: number
  thumbnailUrl?: string
}

// 上传进度类型
export interface UploadProgress {
  file: File
  progress: number
  status: 'pending' | 'uploading' | 'success' | 'error'
  error?: string
  result?: UploadedImage
}

// 批量上传结果类型
export interface BatchUploadResult {
  success: UploadedImage[]
  failed: {
    filename: string
    error: string
  }[]
  total: number
  successCount: number
  failedCount: number
}

// 统计信息类型
export interface Stats {
  totalImages: number
  totalSize: number
  todayUploads: number
  systemUptime: number
  fileTypes: {
    [key: string]: number
  }
  uploadTrend: {
    date: string
    count: number
    size: number
  }[]
  storageUsage: {
    used: number
    total: number
    percentage: number
  }
  systemInfo: {
    version: string
    environment: string
    startTime: string
  }
}

// API 响应类型
export interface ApiResponse<T = any> {
  success: boolean
  message: string
  data?: T
  error?: string
  code?: number
}

// 分页参数类型
export interface PaginationParams {
  page: number
  pageSize: number
  sortBy?: string
  sortOrder?: 'asc' | 'desc'
  search?: string
  fileType?: string
}

// 分页响应类型
export interface PaginatedResponse<T> {
  items: T[]
  total: number
  page: number
  pageSize: number
  totalPages: number
  hasNext: boolean
  hasPrev: boolean
}

// 图片信息查询参数
export interface ImageQueryParams extends PaginationParams {
  startDate?: string
  endDate?: string
  minSize?: number
  maxSize?: number
  tags?: string[]
}

// 文件类型枚举
export enum FileType {
  IMAGE = 'image',
  JPEG = 'image/jpeg',
  PNG = 'image/png',
  GIF = 'image/gif',
  WEBP = 'image/webp',
  SVG = 'image/svg+xml',
  BMP = 'image/bmp',
  TIFF = 'image/tiff'
}

// 支持的图片格式
export const SUPPORTED_IMAGE_TYPES = [
  FileType.JPEG,
  FileType.PNG,
  FileType.GIF,
  FileType.WEBP,
  FileType.SVG,
  FileType.BMP,
  FileType.TIFF
]

// 文件大小限制（字节）
export const FILE_SIZE_LIMITS = {
  MAX_FILE_SIZE: 10 * 1024 * 1024, // 10MB
  MAX_BATCH_SIZE: 50 * 1024 * 1024, // 50MB
  MAX_BATCH_COUNT: 10 // 最多10个文件
}

// 上传状态枚举
export enum UploadStatus {
  PENDING = 'pending',
  UPLOADING = 'uploading',
  SUCCESS = 'success',
  ERROR = 'error',
  CANCELLED = 'cancelled'
}

// 视图模式枚举
export enum ViewMode {
  GRID = 'grid',
  LIST = 'list'
}

// 排序选项
export enum SortOption {
  UPLOAD_TIME_DESC = 'uploadTime_desc',
  UPLOAD_TIME_ASC = 'uploadTime_asc',
  SIZE_DESC = 'size_desc',
  SIZE_ASC = 'size_asc',
  NAME_ASC = 'name_asc',
  NAME_DESC = 'name_desc'
}

// 链接格式类型
export enum LinkFormat {
  DIRECT = 'direct',
  MARKDOWN = 'markdown',
  HTML = 'html',
  BBCode = 'bbcode'
}

// 主题模式
export enum ThemeMode {
  LIGHT = 'light',
  DARK = 'dark',
  AUTO = 'auto'
}

// 语言选项
export enum Language {
  ZH_CN = 'zh-cn',
  EN_US = 'en-us'
}

// 用户设置类型
export interface UserSettings {
  theme: ThemeMode
  language: Language
  defaultLinkFormat: LinkFormat
  autoCompress: boolean
  compressQuality: number
  enableNotifications: boolean
  defaultViewMode: ViewMode
  itemsPerPage: number
}

// 错误类型
export interface AppError {
  code: string
  message: string
  details?: any
  timestamp: string
}

// 通知类型
export interface Notification {
  id: string
  type: 'success' | 'warning' | 'error' | 'info'
  title: string
  message: string
  duration?: number
  timestamp: string
  read: boolean
}

// 系统配置类型
export interface SystemConfig {
  maxFileSize: number
  maxBatchSize: number
  maxBatchCount: number
  supportedFormats: string[]
  enableCompression: boolean
  compressionQuality: number
  enableThumbnails: boolean
  thumbnailSizes: number[]
  enableWatermark: boolean
  watermarkText?: string
  watermarkOpacity?: number
}

// 健康检查响应类型
export interface HealthCheckResponse {
  status: 'healthy' | 'unhealthy'
  timestamp: string
  version: string
  uptime: number
  services: {
    database: 'up' | 'down'
    storage: 'up' | 'down'
    redis?: 'up' | 'down'
  }
  metrics: {
    memoryUsage: number
    cpuUsage: number
    diskUsage: number
  }
}

// 导出工具函数类型
export type FormatFileSize = (bytes: number) => string
export type FormatDate = (date: string | Date, format?: string) => string
export type ValidateFile = (file: File) => { valid: boolean; error?: string }
export type GenerateId = () => string
export type CopyToClipboard = (text: string) => Promise<boolean>
export type DownloadFile = (url: string, filename: string) => void

// 组件 Props 类型
export interface ImageCardProps {
  image: UploadedImage
  viewMode: ViewMode
  showActions?: boolean
  onCopy?: (url: string, format: LinkFormat) => void
  onDownload?: (image: UploadedImage) => void
  onDelete?: (image: UploadedImage) => void
  onPreview?: (image: UploadedImage) => void
}

export interface UploadAreaProps {
  multiple?: boolean
  accept?: string
  maxSize?: number
  maxCount?: number
  disabled?: boolean
  onUpload?: (files: File[]) => void
  onError?: (error: string) => void
}

export interface ProgressBarProps {
  progress: number
  status: UploadStatus
  showText?: boolean
  color?: string
}

export interface StatsCardProps {
  title: string
  value: string | number
  icon?: string
  color?: string
  trend?: {
    value: number
    isUp: boolean
  }
}

// 路由元信息类型
export interface RouteMeta {
  title?: string
  requiresAuth?: boolean
  keepAlive?: boolean
  icon?: string
  hidden?: boolean
  roles?: string[]
}

// 菜单项类型
export interface MenuItem {
  id: string
  title: string
  path: string
  icon?: string
  children?: MenuItem[]
  meta?: RouteMeta
}

// 面包屑项类型
export interface BreadcrumbItem {
  title: string
  path?: string
  disabled?: boolean
}

// 表格列配置类型
export interface TableColumn {
  key: string
  title: string
  width?: number
  sortable?: boolean
  filterable?: boolean
  render?: (value: any, record: any, index: number) => any
}

// 表格配置类型
export interface TableConfig {
  columns: TableColumn[]
  pagination?: boolean
  selection?: boolean
  loading?: boolean
  size?: 'large' | 'default' | 'small'
}

// 表单验证规则类型
export interface ValidationRule {
  required?: boolean
  message?: string
  pattern?: RegExp
  min?: number
  max?: number
  validator?: (value: any) => boolean | string
}

// 表单字段类型
export interface FormField {
  name: string
  label: string
  type: 'input' | 'textarea' | 'select' | 'checkbox' | 'radio' | 'upload' | 'date'
  placeholder?: string
  options?: { label: string; value: any }[]
  rules?: ValidationRule[]
  disabled?: boolean
  hidden?: boolean
}

// 对话框配置类型
export interface DialogConfig {
  title: string
  width?: string | number
  height?: string | number
  modal?: boolean
  closable?: boolean
  maskClosable?: boolean
  keyboard?: boolean
  centered?: boolean
}

// 消息配置类型
export interface MessageConfig {
  type: 'success' | 'warning' | 'error' | 'info'
  message: string
  duration?: number
  showClose?: boolean
  center?: boolean
  onClose?: () => void
}

// 确认对话框配置类型
export interface ConfirmConfig {
  title?: string
  message: string
  type?: 'warning' | 'error' | 'info' | 'success'
  confirmText?: string
  cancelText?: string
  onConfirm?: () => void | Promise<void>
  onCancel?: () => void
}

// 加载配置类型
export interface LoadingConfig {
  text?: string
  spinner?: string
  background?: string
  customClass?: string
}

// 工具提示配置类型
export interface TooltipConfig {
  content: string
  placement?: 'top' | 'bottom' | 'left' | 'right'
  trigger?: 'hover' | 'click' | 'focus'
  disabled?: boolean
}

// 导出所有类型
export * from './index'