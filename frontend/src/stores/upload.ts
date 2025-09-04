import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { UploadFile, UploadUserFile } from 'element-plus'
import { uploadImage, batchUploadImages, getImageInfo, getStats } from '@/api/upload'

export interface ImageInfo {
  id: number
  uuid: string
  original_name: string
  file_size: number
  mime_type: string
  width: number
  height: number
  public_url: string
  created_at: string
}

export interface UploadResult {
  success: boolean
  data?: ImageInfo
  error?: string
}

export interface BatchUploadResult {
  success: boolean
  data?: {
    successful: number
    failed: number
    results: ImageInfo[]
    errors: any[]
  }
  error?: string
}

export interface StatsInfo {
  total_images: number
  total_size: number
  today_images: number
}

export const useUploadStore = defineStore('upload', () => {
  // 状态
  const uploadedImages = ref<ImageInfo[]>([])
  const isUploading = ref(false)
  const uploadProgress = ref(0)
  const stats = ref<StatsInfo>({
    total_images: 0,
    total_size: 0,
    today_images: 0
  })

  // 计算属性
  const totalImages = computed(() => uploadedImages.value.length)
  const totalSize = computed(() => {
    return uploadedImages.value.reduce((sum, img) => sum + img.file_size, 0)
  })

  // 格式化文件大小
  const formatFileSize = (bytes: number): string => {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  // 单个文件上传
  const uploadSingleImage = async (file: File): Promise<UploadResult> => {
    try {
      isUploading.value = true
      uploadProgress.value = 0

      const formData = new FormData()
      formData.append('image', file)

      const response = await uploadImage(formData, (progress) => {
        uploadProgress.value = progress
      })

      if (response.success) {
        uploadedImages.value.unshift(response.data)
        await fetchStats() // 更新统计信息
      }

      return response
    } catch (error: any) {
      return {
        success: false,
        error: error.message || '上传失败'
      }
    } finally {
      isUploading.value = false
      uploadProgress.value = 0
    }
  }

  // 批量文件上传
  const uploadMultipleImages = async (files: File[]): Promise<BatchUploadResult> => {
    try {
      isUploading.value = true
      uploadProgress.value = 0

      const formData = new FormData()
      files.forEach(file => {
        formData.append('images', file)
      })

      const response = await batchUploadImages(formData, (progress) => {
        uploadProgress.value = progress
      })

      if (response.success && response.data) {
        // 将成功上传的图片添加到列表前面
        uploadedImages.value.unshift(...response.data.results)
        await fetchStats() // 更新统计信息
      }

      return response
    } catch (error: any) {
      return {
        success: false,
        error: error.message || '批量上传失败'
      }
    } finally {
      isUploading.value = false
      uploadProgress.value = 0
    }
  }

  // 获取图片信息
  const fetchImageInfo = async (uuid: string): Promise<ImageInfo | null> => {
    try {
      const response = await getImageInfo(uuid)
      if (response.success) {
        return response.data
      }
      return null
    } catch (error) {
      console.error('获取图片信息失败:', error)
      return null
    }
  }

  // 获取统计信息
  const fetchStats = async (): Promise<void> => {
    try {
      const response = await getStats()
      if (response.success) {
        stats.value = response.data
      }
    } catch (error) {
      console.error('获取统计信息失败:', error)
    }
  }

  // 复制链接到剪贴板
  const copyToClipboard = async (text: string): Promise<boolean> => {
    try {
      await navigator.clipboard.writeText(text)
      return true
    } catch (error) {
      // 降级方案
      const textArea = document.createElement('textarea')
      textArea.value = text
      document.body.appendChild(textArea)
      textArea.select()
      const success = document.execCommand('copy')
      document.body.removeChild(textArea)
      return success
    }
  }

  // 清空上传列表
  const clearUploadedImages = (): void => {
    uploadedImages.value = []
  }

  // 从列表中移除图片
  const removeImage = (uuid: string): void => {
    const index = uploadedImages.value.findIndex(img => img.uuid === uuid)
    if (index > -1) {
      uploadedImages.value.splice(index, 1)
    }
  }

  return {
    // 状态
    uploadedImages,
    isUploading,
    uploadProgress,
    stats,
    
    // 计算属性
    totalImages,
    totalSize,
    
    // 方法
    formatFileSize,
    uploadSingleImage,
    uploadMultipleImages,
    fetchImageInfo,
    fetchStats,
    copyToClipboard,
    clearUploadedImages,
    removeImage
  }
})