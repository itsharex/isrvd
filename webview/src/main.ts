import { createApp } from 'vue'

import App from '@/app.vue'
import router from '@/router'

// 导入全局样式
import '@fortawesome/fontawesome-free/css/all.min.css'
import '@xterm/xterm/css/xterm.css'

// 导入 Tailwind CSS 样式
import './assets/style.css'
import './assets/dark.css'

// 创建并挂载应用
const app = createApp(App)
app.use(router).mount('#app')
