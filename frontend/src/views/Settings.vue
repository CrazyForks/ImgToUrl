<template>
  <div class="settings-page">
    <NavBar />
    <main class="main">
      <h2 class="title">设置</h2>

      <!-- 修改密码 -->
      <el-card class="card">
        <template #header>修改密码</template>
        <el-form :model="form" label-width="96px" style="max-width:480px">
          <el-form-item label="旧密码">
            <el-input v-model="form.old_password" type="password" show-password />
          </el-form-item>
          <el-form-item label="新密码">
            <el-input v-model="form.new_password" type="password" show-password />
          </el-form-item>
          <el-button type="primary" :loading="loading" @click="onSubmit">保存</el-button>
        </el-form>
      </el-card>

      <!-- 游客码管理（仅 root 可见） -->
      <el-card v-if="isRoot" class="card" style="margin-top:16px">
        <template #header>游客码管理</template>
        <div style="max-width:520px">
          <el-form :model="guestForm" label-width="120px">
            <el-form-item label="有效时长">
              <el-select v-model="guestForm.mode" placeholder="选择时长">
                <el-option v-for="d in [1,2,3,4,5,6,7]" :key="d" :label="d + ' 天'" :value="d" />
                <el-option label="自定义(秒)" :value="0" />
                <el-option label="永久" :value="-1" />
              </el-select>
            </el-form-item>
            <el-form-item v-if="guestForm.mode===0" label="自定义秒数">
              <el-input v-model.number="guestForm.customSeconds" type="number" placeholder="例如 86400" />
            </el-form-item>
            <el-button type="primary" :loading="gLoading" @click="onCreateCode">生成游客码</el-button>
          </el-form>

          <el-table :data="codes" size="small" style="margin-top:12px">
            <el-table-column prop="id" label="ID" width="80" />
            <el-table-column label="游客码">
              <template #default="{ row }">
                <span class="copy-code" @click="copyCode(row.code)">{{ row.code }}</span>
              </template>
            </el-table-column>
            <el-table-column label="过期时间">
              <template #default="{ row }">
                <span v-if="row.expires_at">{{ new Date(row.expires_at).toLocaleString() }}</span>
                <span v-else>永久</span>
              </template>
            </el-table-column>
            <el-table-column width="140" label="操作">
              <template #default="{ row }">
                <el-popconfirm title="删除该游客及其图片？" @confirm="onDeleteCode(row.id)">
                  <template #reference>
                    <el-button size="small" type="danger">删除</el-button>
                  </template>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-card>

      <!-- 存储配置（仅 root 可见） -->
      <el-card v-if="isRoot" class="card" style="margin-top:16px">
        <template #header>存储配置</template>
        <div style="max-width:720px">
          <el-form :model="storage" label-width="140px">
            <el-form-item label="存储渠道">
              <el-radio-group v-model="storage.provider">
                <el-radio label="local">VPS 本地存储</el-radio>
                <el-radio label="r2">Cloudflare R2</el-radio>
                <el-radio label="oss">阿里云 OSS</el-radio>
              </el-radio-group>
            </el-form-item>

            <el-alert
              v-if="storage.provider === 'local'"
              title="本地存储说明"
              type="info"
              :closable="false"
              description="/uploads 映射到后端的 UPLOAD_PATH 目录。"
              style="margin-bottom: 12px"
            />

            <template v-if="storage.provider === 'r2'">
              <el-alert
                title="R2 配置说明"
                type="warning"
                :closable="false"
                description="请填写 R2 的 S3 兼容访问参数。当前版本若后端未启用 R2 SDK，会回退为本地存储。"
                style="margin-bottom: 12px"
              />
              <el-form-item label="Account ID">
                <el-input v-model="storage.r2.account_id" placeholder="例如：b1234567890abcdef" />
              </el-form-item>
              <el-form-item label="Access Key ID">
                <el-input v-model="storage.r2.access_key_id" placeholder="访问密钥 ID" />
              </el-form-item>
              <el-form-item label="Secret Access Key">
                <el-input v-model="storage.r2.secret_access_key" placeholder="访问密钥 Secret" type="password" show-password />
              </el-form-item>
              <el-form-item label="Bucket Name">
                <el-input v-model="storage.r2.bucket_name" placeholder="存储桶名称" />
              </el-form-item>
              <el-form-item label="Endpoint">
                <el-input v-model="storage.r2.endpoint" placeholder="例如：https://xxxxxxxxxxx.r2.cloudflarestorage.com" />
              </el-form-item>
              <el-form-item label="Region">
                <el-input v-model="storage.r2.region" placeholder="通常为 auto" />
              </el-form-item>
              <el-form-item label="Public URL 前缀">
                <el-input v-model="storage.r2.public_url" placeholder="用于生成公共访问 URL 的前缀（可选）" />
              </el-form-item>
            </template>

            <template v-if="storage.provider === 'oss'">
              <el-alert
                title="OSS 配置说明"
                type="warning"
                :closable="false"
                description="请填写阿里云 OSS 的接入参数。填写后保存即可切换到 OSS 存储。"
                style="margin-bottom: 12px"
              />
              <el-form-item label="Access Key ID">
                <el-input v-model="storage.oss.access_key_id" placeholder="访问密钥 ID" />
              </el-form-item>
              <el-form-item label="Secret Access Key">
                <el-input v-model="storage.oss.secret_access_key" placeholder="访问密钥 Secret" type="password" show-password />
              </el-form-item>
              <el-form-item label="Bucket Name">
                <el-input v-model="storage.oss.bucket_name" placeholder="存储桶名称" />
              </el-form-item>
              <el-form-item label="Endpoint">
                <el-input v-model="storage.oss.endpoint" placeholder="例如：https://oss-cn-hangzhou.aliyuncs.com 或自定义加速域" />
              </el-form-item>
              <el-form-item label="Region">
                <el-input v-model="storage.oss.region" placeholder="例如：oss-cn-hangzhou（可选）" />
              </el-form-item>
              <el-form-item label="Public URL 前缀">
                <el-input v-model="storage.oss.public_url" placeholder="用于生成公共访问 URL 的前缀（可选，建议填写 CDN 域名）" />
              </el-form-item>
            </template>

            <div style="display:flex; gap:8px; margin-top:8px">
              <el-button :loading="scLoading" type="primary" @click="onSaveStorage">保存配置</el-button>
              <el-button :loading="scLoading" @click="onReloadStorage">重新获取</el-button>
            </div>
          </el-form>
        </div>
      </el-card>
    </main>
  </div>
</template>

<script setup lang="ts">
import NavBar from '@/components/NavBar.vue'
import { reactive, ref, computed, onMounted } from 'vue'
import { ElMessage } from 'element-plus'
import { changePassword, createGuestCode, listGuestCodes, deleteGuestCode } from '@/api/auth'
import { getStorageConfig, updateStorageConfig, type StorageConfigData } from '@/api/system'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

// 修改密码
const form = reactive({ old_password: '', new_password: '' })
const loading = ref(false)

const onSubmit = async () => {
  if (!form.old_password || !form.new_password) {
    ElMessage.error('请输入旧密码和新密码')
    return
  }
  loading.value = true
  try {
    const res = await changePassword(auth.token!, form.old_password, form.new_password)
    if (res.success) {
      ElMessage.success('密码已更新，请重新登录')
      auth.clearAuth()
      location.href = '/login'
    } else {
      ElMessage.error(res.error || '修改失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '修改失败')
  } finally {
    loading.value = false
  }
}

// 游客码管理（仅 root）
const isRoot = computed(() => auth.user?.username === 'root')

const guestForm = reactive<{ mode: number; customSeconds?: number }>({ mode: 1 })
const gLoading = ref(false)
const codes = ref<any[]>([])

const fetchCodes = async () => {
  if (!isRoot.value) return
  const res = await listGuestCodes(auth.token!)
  if (res.success) codes.value = res.data || []
}

// 存储配置
const storage = reactive<StorageConfigData>({
  provider: 'local',
  r2: {
    account_id: '',
    access_key_id: '',
    secret_access_key: '',
    bucket_name: '',
    endpoint: '',
    region: 'auto',
    public_url: ''
  },
  oss: {
    access_key_id: '',
    secret_access_key: '',
    bucket_name: '',
    endpoint: '',
    region: '',
    public_url: ''
  }
})
const scLoading = ref(false)

const loadStorage = async () => {
  if (!isRoot.value) return
  scLoading.value = true
  try {
    const res = await getStorageConfig()
    if (res.success && res.data) {
      storage.provider = (res.data.provider as 'local' | 'r2' | 'oss') || 'local'
      storage.r2 = {
        account_id: res.data.r2?.account_id || '',
        access_key_id: res.data.r2?.access_key_id || '',
        secret_access_key: res.data.r2?.secret_access_key || '',
        bucket_name: res.data.r2?.bucket_name || '',
        endpoint: res.data.r2?.endpoint || '',
        region: res.data.r2?.region || 'auto',
        public_url: res.data.r2?.public_url || ''
      }
      storage.oss = {
        access_key_id: res.data.oss?.access_key_id || '',
        secret_access_key: res.data.oss?.secret_access_key || '',
        bucket_name: res.data.oss?.bucket_name || '',
        endpoint: res.data.oss?.endpoint || '',
        region: res.data.oss?.region || '',
        public_url: res.data.oss?.public_url || ''
      }
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '获取存储配置失败')
  } finally {
    scLoading.value = false
  }
}

const onSaveStorage = async () => {
  if (!isRoot.value) {
    ElMessage.error('仅管理员可操作')
    return
  }
  scLoading.value = true
  try {
    const res = await updateStorageConfig({
      provider: storage.provider,
      r2: { ...storage.r2 },
      oss: { ...storage.oss }
    })
    if (res.success) {
      ElMessage.success('已保存存储配置')
    } else {
      ElMessage.error(res.error || '保存失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '保存失败')
  } finally {
    scLoading.value = false
  }
}

const onReloadStorage = () => loadStorage()

onMounted(fetchCodes)
onMounted(loadStorage)

const onCreateCode = async () => {
  gLoading.value = true
  try {
    let payload: any = {}
    if (guestForm.mode === -1) {
      payload.permanent = true
    } else if (guestForm.mode === 0) {
      if (!guestForm.customSeconds || guestForm.customSeconds <= 0) {
        ElMessage.error('请输入自定义秒数')
        gLoading.value = false
        return
      }
      payload.expires_at = Math.floor(Date.now() / 1000) + guestForm.customSeconds
    } else {
      payload.days = guestForm.mode
    }
    const res = await createGuestCode(auth.token!, payload)
    if (res.success) {
      ElMessage.success('已生成游客码')
      fetchCodes()
    } else {
      ElMessage.error(res.error || '生成失败')
    }
  } finally {
    gLoading.value = false
  }
}

const onDeleteCode = async (id: number) => {
  const res = await deleteGuestCode(auth.token!, id)
  if (res.success) {
    ElMessage.success('已删除')
    fetchCodes()
  } else {
    ElMessage.error(res.error || '删除失败')
  }
}
const copyCode = async (code: string) => {
  try {
    await navigator.clipboard.writeText(code)
    ElMessage.success('已复制')
  } catch (e) {
    try {
      const ta = document.createElement('textarea')
      ta.value = code
      ta.style.position = 'fixed'
      ta.style.opacity = '0'
      document.body.appendChild(ta)
      ta.select()
      document.execCommand('copy')
      document.body.removeChild(ta)
      ElMessage.success('已复制')
    } catch {
      ElMessage.error('复制失败')
    }
  }
}
</script>

<style scoped>
.main { max-width: 960px; margin: 24px auto; padding: 0 16px; }
.title { color: #fff; margin: 0 0 16px; }
/* 设置页卡片：去除糊感、禁用悬停位移/滤镜 */
.card {
  background: #111827;         /* 更清晰的深色纯色背景（tailwind slate-900） */
  color: #e5e7eb;
  border: 1px solid rgba(255,255,255,0.08);
  box-shadow: none;
  transform: none !important;
  filter: none !important;
  backdrop-filter: none !important;
}
/* 悬停不移动不变形 */
.card:hover {
  transform: none !important;
  filter: none !important;
  box-shadow: none;
}
.copy-code {
  cursor: pointer;
  color: #93c5fd; /* 提示可点击 */
  text-decoration: underline dotted;
}
.copy-code:hover {
  color: #bfdbfe;
  filter: none !important;
  transform: none !important;
}
</style>