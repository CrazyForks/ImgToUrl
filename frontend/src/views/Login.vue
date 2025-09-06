<template>
  <div class="login-page">
    <div class="login-card">
      <h1 class="title">登录</h1>
      <p class="subtitle">默认账号 root / 123456（请登录后尽快修改密码）</p>

      <el-tabs v-model="tab">
        <el-tab-pane label="账号登录" name="account">
          <el-form :model="form" @keyup.enter="onSubmit">
            <el-form-item>
              <el-input v-model="form.username" placeholder="用户名" clearable />
            </el-form-item>
            <el-form-item>
              <el-input v-model="form.password" type="password" placeholder="密码" show-password />
            </el-form-item>
            <el-button type="primary" class="full" :loading="loading" @click="onSubmit">登录</el-button>
          </el-form>
        </el-tab-pane>

        <el-tab-pane label="游客模式" name="guest">
          <el-form :model="guest" @keyup.enter="onGuestLogin">
            <el-form-item>
              <el-input v-model="guest.code" placeholder="请输入游客码" clearable />
            </el-form-item>
            <el-button type="primary" class="full" :loading="gLoading" @click="onGuestLogin">游客登录</el-button>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { login, guestLogin } from '@/api/auth'
import { useAuthStore } from '@/stores/auth'

const tab = ref<'account' | 'guest'>('account')
const form = reactive({ username: 'root', password: '123456' })
const guest = reactive({ code: '' })
const loading = ref(false)
const gLoading = ref(false)
const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const onSubmit = async () => {
  if (!form.username || !form.password) {
    ElMessage.error('请输入用户名和密码')
    return
  }
  loading.value = true
  try {
    const res = await login(form.username, form.password)
    if (res.success && res.data) {
      auth.setAuth(res.data.token, res.data.username)
      ElMessage.success('登录成功')
      const redirect = (route.query.redirect as string) || '/'
      router.replace(redirect)
    } else {
      ElMessage.error(res.error || '登录失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '登录失败')
  } finally {
    loading.value = false
  }
}
const onGuestLogin = async () => {
  if (!guest.code) {
    ElMessage.error('请输入游客码')
    return
  }
  gLoading.value = true
  try {
    const res = await guestLogin(guest.code)
    if (res.success && res.data) {
      auth.setAuth(res.data.token, res.data.username)
      ElMessage.success('游客登录成功')
      const redirect = (route.query.redirect as string) || '/'
      router.replace(redirect)
    } else {
      ElMessage.error(res.error || '游客登录失败')
    }
  } catch (e: any) {
    ElMessage.error(e?.message || '游客登录失败')
  } finally {
    gLoading.value = false
  }
}
</script>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex; align-items: center; justify-content: center;
  background: radial-gradient(1200px 800px at 50% -20%, rgba(255,255,255,0.15), transparent),
              linear-gradient(160deg, #0b1220, #0c0f14 60%, #0a0c10);
}
.login-card {
  width: 360px; padding: 28px;
  background: rgba(255,255,255,0.06);
  border: 1px solid rgba(255,255,255,0.12);
  border-radius: 16px;
  box-shadow: 0 12px 32px rgba(0,0,0,0.35);
  color: #e5e7eb;
}
.title { margin: 0 0 6px; font-size: 22px; font-weight: 700; }
.subtitle { margin: 0 0 16px; color: #9ca3af; font-size: 13px; }
.full { width: 100%; }
</style>