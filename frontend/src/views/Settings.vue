<template>
  <div class="settings-page">
    <NavBar />
    <main class="main">
      <h2 class="title">设置</h2>
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
    </main>
  </div>
</template>

<script setup lang="ts">
import NavBar from '@/components/NavBar.vue'
import { reactive, ref } from 'vue'
import { ElMessage } from 'element-plus'
import { changePassword } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()
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
</script>

<style scoped>
.main { max-width: 960px; margin: 24px auto; padding: 0 16px; }
.title { color: #fff; margin: 0 0 16px; }
.card { background: rgba(255,255,255,0.06); color: #e5e7eb; }
</style>