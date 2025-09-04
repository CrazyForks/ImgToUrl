import { FileType, SUPPORTED_IMAGE_TYPES, FILE_SIZE_LIMITS } from '@/types'

/**
 * 格式化文件大小
 * @param bytes 字节数
 * @returns 格式化后的文件大小字符串
 */
export function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B'
  
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

/**
 * 格式化日期时间
 * @param date 日期字符串或Date对象
 * @param format 格式化模板
 * @returns 格式化后的日期字符串
 */
export function formatDate(date: string | Date, format: string = 'YYYY-MM-DD HH:mm:ss'): string {
  const d = new Date(date)
  
  if (isNaN(d.getTime())) {
    return '无效日期'
  }
  
  const year = d.getFullYear()
  const month = String(d.getMonth() + 1).padStart(2, '0')
  const day = String(d.getDate()).padStart(2, '0')
  const hours = String(d.getHours()).padStart(2, '0')
  const minutes = String(d.getMinutes()).padStart(2, '0')
  const seconds = String(d.getSeconds()).padStart(2, '0')
  
  return format
    .replace('YYYY', String(year))
    .replace('MM', month)
    .replace('DD', day)
    .replace('HH', hours)
    .replace('mm', minutes)
    .replace('ss', seconds)
}

/**
 * 验证文件是否符合要求
 * @param file 文件对象
 * @returns 验证结果
 */
export function validateFile(file: File): { valid: boolean; error?: string } {
  // 检查文件类型
  if (!SUPPORTED_IMAGE_TYPES.includes(file.type as FileType)) {
    return {
      valid: false,
      error: `不支持的文件类型: ${file.type}。支持的格式: ${SUPPORTED_IMAGE_TYPES.join(', ')}`
    }
  }
  
  // 检查文件大小
  if (file.size > FILE_SIZE_LIMITS.MAX_FILE_SIZE) {
    return {
      valid: false,
      error: `文件大小超出限制。最大允许: ${formatFileSize(FILE_SIZE_LIMITS.MAX_FILE_SIZE)}`
    }
  }
  
  // 检查文件名
  if (!file.name || file.name.trim() === '') {
    return {
      valid: false,
      error: '文件名不能为空'
    }
  }
  
  return { valid: true }
}

/**
 * 验证批量文件上传
 * @param files 文件数组
 * @returns 验证结果
 */
export function validateBatchFiles(files: File[]): { valid: boolean; error?: string } {
  // 检查文件数量
  if (files.length > FILE_SIZE_LIMITS.MAX_BATCH_COUNT) {
    return {
      valid: false,
      error: `文件数量超出限制。最多允许: ${FILE_SIZE_LIMITS.MAX_BATCH_COUNT} 个文件`
    }
  }
  
  // 检查总文件大小
  const totalSize = files.reduce((sum, file) => sum + file.size, 0)
  if (totalSize > FILE_SIZE_LIMITS.MAX_BATCH_SIZE) {
    return {
      valid: false,
      error: `批量上传总大小超出限制。最大允许: ${formatFileSize(FILE_SIZE_LIMITS.MAX_BATCH_SIZE)}`
    }
  }
  
  // 验证每个文件
  for (const file of files) {
    const result = validateFile(file)
    if (!result.valid) {
      return result
    }
  }
  
  return { valid: true }
}

/**
 * 生成唯一ID
 * @returns 唯一ID字符串
 */
export function generateId(): string {
  return Date.now().toString(36) + Math.random().toString(36).substr(2)
}

/**
 * 复制文本到剪贴板
 * @param text 要复制的文本
 * @returns 是否复制成功
 */
export async function copyToClipboard(text: string): Promise<boolean> {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text)
      return true
    } else {
      // 降级方案
      const textArea = document.createElement('textarea')
      textArea.value = text
      textArea.style.position = 'fixed'
      textArea.style.left = '-999999px'
      textArea.style.top = '-999999px'
      document.body.appendChild(textArea)
      textArea.focus()
      textArea.select()
      
      const success = document.execCommand('copy')
      document.body.removeChild(textArea)
      return success
    }
  } catch (error) {
    console.error('复制到剪贴板失败:', error)
    return false
  }
}

/**
 * 下载文件
 * @param url 文件URL
 * @param filename 文件名
 */
export function downloadFile(url: string, filename: string): void {
  try {
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    link.style.display = 'none'
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
  } catch (error) {
    console.error('下载文件失败:', error)
    // 降级方案：在新窗口打开
    window.open(url, '_blank')
  }
}

/**
 * 获取文件扩展名
 * @param filename 文件名
 * @returns 文件扩展名
 */
export function getFileExtension(filename: string): string {
  return filename.slice((filename.lastIndexOf('.') - 1 >>> 0) + 2).toLowerCase()
}

/**
 * 获取文件类型图标
 * @param mimeType MIME类型
 * @returns 图标名称
 */
export function getFileTypeIcon(mimeType: string): string {
  const iconMap: Record<string, string> = {
    'image/jpeg': 'picture',
    'image/jpg': 'picture',
    'image/png': 'picture',
    'image/gif': 'picture',
    'image/webp': 'picture',
    'image/svg+xml': 'picture',
    'image/bmp': 'picture',
    'image/tiff': 'picture'
  }
  
  return iconMap[mimeType] || 'document'
}

/**
 * 防抖函数
 * @param func 要防抖的函数
 * @param wait 等待时间（毫秒）
 * @returns 防抖后的函数
 */
export function debounce<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let timeout: number | null = null
  
  return (...args: Parameters<T>) => {
    if (timeout) {
      clearTimeout(timeout)
    }
    
    timeout = setTimeout(() => {
      func.apply(null, args)
    }, wait)
  }
}

/**
 * 节流函数
 * @param func 要节流的函数
 * @param wait 等待时间（毫秒）
 * @returns 节流后的函数
 */
export function throttle<T extends (...args: any[]) => any>(
  func: T,
  wait: number
): (...args: Parameters<T>) => void {
  let inThrottle: boolean = false
  
  return (...args: Parameters<T>) => {
    if (!inThrottle) {
      func.apply(null, args)
      inThrottle = true
      setTimeout(() => {
        inThrottle = false
      }, wait)
    }
  }
}

/**
 * 深拷贝对象
 * @param obj 要拷贝的对象
 * @returns 拷贝后的对象
 */
export function deepClone<T>(obj: T): T {
  if (obj === null || typeof obj !== 'object') {
    return obj
  }
  
  if (obj instanceof Date) {
    return new Date(obj.getTime()) as T
  }
  
  if (obj instanceof Array) {
    return obj.map(item => deepClone(item)) as T
  }
  
  if (typeof obj === 'object') {
    const clonedObj = {} as T
    for (const key in obj) {
      if (obj.hasOwnProperty(key)) {
        clonedObj[key] = deepClone(obj[key])
      }
    }
    return clonedObj
  }
  
  return obj
}

/**
 * 检查是否为移动设备
 * @returns 是否为移动设备
 */
export function isMobile(): boolean {
  return /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
    navigator.userAgent
  )
}

/**
 * 获取设备类型
 * @returns 设备类型
 */
export function getDeviceType(): 'mobile' | 'tablet' | 'desktop' {
  const width = window.innerWidth
  
  if (width < 768) {
    return 'mobile'
  } else if (width < 1024) {
    return 'tablet'
  } else {
    return 'desktop'
  }
}

/**
 * 格式化数字
 * @param num 数字
 * @param decimals 小数位数
 * @returns 格式化后的数字字符串
 */
export function formatNumber(num: number, decimals: number = 0): string {
  return num.toLocaleString('zh-CN', {
    minimumFractionDigits: decimals,
    maximumFractionDigits: decimals
  })
}

/**
 * 计算百分比
 * @param value 当前值
 * @param total 总值
 * @param decimals 小数位数
 * @returns 百分比字符串
 */
export function calculatePercentage(
  value: number,
  total: number,
  decimals: number = 1
): string {
  if (total === 0) return '0%'
  const percentage = (value / total) * 100
  return `${percentage.toFixed(decimals)}%`
}

/**
 * 获取相对时间
 * @param date 日期
 * @returns 相对时间字符串
 */
export function getRelativeTime(date: string | Date): string {
  const now = new Date()
  const target = new Date(date)
  const diff = now.getTime() - target.getTime()
  
  const seconds = Math.floor(diff / 1000)
  const minutes = Math.floor(seconds / 60)
  const hours = Math.floor(minutes / 60)
  const days = Math.floor(hours / 24)
  const months = Math.floor(days / 30)
  const years = Math.floor(days / 365)
  
  if (years > 0) {
    return `${years}年前`
  } else if (months > 0) {
    return `${months}个月前`
  } else if (days > 0) {
    return `${days}天前`
  } else if (hours > 0) {
    return `${hours}小时前`
  } else if (minutes > 0) {
    return `${minutes}分钟前`
  } else if (seconds > 0) {
    return `${seconds}秒前`
  } else {
    return '刚刚'
  }
}

/**
 * 生成随机颜色
 * @returns 十六进制颜色字符串
 */
export function generateRandomColor(): string {
  return '#' + Math.floor(Math.random() * 16777215).toString(16).padStart(6, '0')
}

/**
 * 获取颜色的亮度
 * @param color 十六进制颜色字符串
 * @returns 亮度值 (0-255)
 */
export function getColorBrightness(color: string): number {
  const hex = color.replace('#', '')
  const r = parseInt(hex.substr(0, 2), 16)
  const g = parseInt(hex.substr(2, 2), 16)
  const b = parseInt(hex.substr(4, 2), 16)
  
  return (r * 299 + g * 587 + b * 114) / 1000
}

/**
 * 判断颜色是否为深色
 * @param color 十六进制颜色字符串
 * @returns 是否为深色
 */
export function isDarkColor(color: string): boolean {
  return getColorBrightness(color) < 128
}

/**
 * 压缩图片
 * @param file 图片文件
 * @param quality 压缩质量 (0-1)
 * @param maxWidth 最大宽度
 * @param maxHeight 最大高度
 * @returns 压缩后的Blob
 */
export function compressImage(
  file: File,
  quality: number = 0.8,
  maxWidth: number = 1920,
  maxHeight: number = 1080
): Promise<Blob> {
  return new Promise((resolve, reject) => {
    const canvas = document.createElement('canvas')
    const ctx = canvas.getContext('2d')
    const img = new Image()
    
    img.onload = () => {
      // 计算新的尺寸
      let { width, height } = img
      
      if (width > maxWidth) {
        height = (height * maxWidth) / width
        width = maxWidth
      }
      
      if (height > maxHeight) {
        width = (width * maxHeight) / height
        height = maxHeight
      }
      
      canvas.width = width
      canvas.height = height
      
      // 绘制图片
      ctx?.drawImage(img, 0, 0, width, height)
      
      // 转换为Blob
      canvas.toBlob(
        (blob) => {
          if (blob) {
            resolve(blob)
          } else {
            reject(new Error('图片压缩失败'))
          }
        },
        file.type,
        quality
      )
    }
    
    img.onerror = () => {
      reject(new Error('图片加载失败'))
    }
    
    img.src = URL.createObjectURL(file)
  })
}

/**
 * 获取图片尺寸
 * @param file 图片文件
 * @returns 图片尺寸
 */
export function getImageDimensions(file: File): Promise<{ width: number; height: number }> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    
    img.onload = () => {
      resolve({
        width: img.naturalWidth,
        height: img.naturalHeight
      })
      URL.revokeObjectURL(img.src)
    }
    
    img.onerror = () => {
      reject(new Error('无法获取图片尺寸'))
      URL.revokeObjectURL(img.src)
    }
    
    img.src = URL.createObjectURL(file)
  })
}

/**
 * 创建缩略图
 * @param file 图片文件
 * @param size 缩略图尺寸
 * @returns 缩略图Blob
 */
export function createThumbnail(file: File, size: number = 200): Promise<Blob> {
  return compressImage(file, 0.7, size, size)
}

/**
 * 本地存储工具
 */
export const storage = {
  /**
   * 设置本地存储
   * @param key 键
   * @param value 值
   */
  set(key: string, value: any): void {
    try {
      localStorage.setItem(key, JSON.stringify(value))
    } catch (error) {
      console.error('设置本地存储失败:', error)
    }
  },
  
  /**
   * 获取本地存储
   * @param key 键
   * @param defaultValue 默认值
   * @returns 存储的值
   */
  get<T>(key: string, defaultValue?: T): T | null {
    try {
      const item = localStorage.getItem(key)
      return item ? JSON.parse(item) : defaultValue || null
    } catch (error) {
      console.error('获取本地存储失败:', error)
      return defaultValue || null
    }
  },
  
  /**
   * 删除本地存储
   * @param key 键
   */
  remove(key: string): void {
    try {
      localStorage.removeItem(key)
    } catch (error) {
      console.error('删除本地存储失败:', error)
    }
  },
  
  /**
   * 清空本地存储
   */
  clear(): void {
    try {
      localStorage.clear()
    } catch (error) {
      console.error('清空本地存储失败:', error)
    }
  }
}

/**
 * URL 工具
 */
export const urlUtils = {
  /**
   * 获取URL参数
   * @param name 参数名
   * @returns 参数值
   */
  getParam(name: string): string | null {
    const urlParams = new URLSearchParams(window.location.search)
    return urlParams.get(name)
  },
  
  /**
   * 设置URL参数
   * @param params 参数对象
   */
  setParams(params: Record<string, string>): void {
    const url = new URL(window.location.href)
    Object.entries(params).forEach(([key, value]) => {
      url.searchParams.set(key, value)
    })
    window.history.replaceState({}, '', url.toString())
  },
  
  /**
   * 删除URL参数
   * @param names 参数名数组
   */
  removeParams(names: string[]): void {
    const url = new URL(window.location.href)
    names.forEach(name => {
      url.searchParams.delete(name)
    })
    window.history.replaceState({}, '', url.toString())
  }
}

/**
 * 导出所有工具函数
 */
export default {
  formatFileSize,
  formatDate,
  validateFile,
  validateBatchFiles,
  generateId,
  copyToClipboard,
  downloadFile,
  getFileExtension,
  getFileTypeIcon,
  debounce,
  throttle,
  deepClone,
  isMobile,
  getDeviceType,
  formatNumber,
  calculatePercentage,
  getRelativeTime,
  generateRandomColor,
  getColorBrightness,
  isDarkColor,
  compressImage,
  getImageDimensions,
  createThumbnail,
  storage,
  urlUtils
}