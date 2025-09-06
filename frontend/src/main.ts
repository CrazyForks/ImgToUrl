import { createApp } from 'vue'
import { createPinia } from 'pinia'
import router from './router'
import App from './App.vue'

// Element Plus 样式
import 'element-plus/dist/index.css'
import 'element-plus/theme-chalk/dark/css-vars.css'

// 自定义样式
import './styles/index.scss'

const app = createApp(App)
const pinia = createPinia()

app.use(pinia)
app.use(router)

app.mount('#app')

// 全局：按钮点击放大→复原动画（事件委托）
// 说明：为所有按钮型元素添加一次性类名 .btn-click-pop，触发 CSS 动画
document.addEventListener('mousedown', (e) => {
  const target = (e.target as HTMLElement)?.closest(
    'button, .el-button, [role="button"], .el-pagination .btn-prev, .el-pagination .btn-next, .el-pagination .el-pager li'
  ) as HTMLElement | null
  if (!target) return
  // 允许快速连续点击时重新触发动画
  target.classList.remove('btn-click-pop')
  void target.offsetWidth // 强制回流以重启动画
  target.classList.add('btn-click-pop')
})