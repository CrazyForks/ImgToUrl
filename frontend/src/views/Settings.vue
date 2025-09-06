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
            <el-table-column prop="code" label="游客码" />
            <el-table-column label="过期时间">
              <template #default="{ row }">
                <span v-if="row.expires_at">{{ new Date(row.expires_at).toLocaleString() }}</span>
                <span v-else>永久</span>
              </template>
            </el-table-column>
            <el-table-column width="120" label="操作">
              <template #default="{ row }">
                <el-popconfirm title="删除该游客及其图片？" @confirm="onDeleteCode(row.id)">
                  <el-button size="small" type="danger">删除</el-button>
                </el-popconfirm>
              </template>
            </el-table-column>
          </el-table>
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

onMounted(fetchCodes)

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
</script>

<style scoped>
.main { max-width: 960px; margin: 24px auto; padding: 0 16px; }
.title { color: #fff; margin: 0 0 16px; }
.card { background: rgba(255,255,255,0.06); color: #e5e7eb; }
</style>