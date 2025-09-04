<template>
  <div id="app">
    <!-- 固定的全局背景 -->
    

    <!-- 页面内容区域 -->
    <div class="page-content">
      <el-config-provider :locale="locale">
        <router-view />
      </el-config-provider>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import zhCn from 'element-plus/dist/locale/zh-cn.mjs'


const locale = ref(zhCn)
</script>

<style lang="scss">
html, body, #app {
  min-height: 100vh;
}

#app {
  position: relative;
  color: #ffffff;
}



/* 页面内容始终在背景之上，并可根据内容自然增高 */
.page-content {
  position: relative;
  z-index: 1;
  min-height: 100vh;
}

/* 可选：通用重置 */
* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>