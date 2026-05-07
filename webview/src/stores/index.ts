/**
 * Store 统一入口
 * 
 * ⚠️ 重要：禁止直接引用子 store 文件
 * ❌ 错误：import { useAuthStore } from './stores/auth'
 * ✅ 正确：import { usePortal } from './stores'
 * 
 * 所有功能都通过 Portal 访问，确保：
 * 1. 统一的初始化流程
 * 2. 统一的拦截器注册
 * 3. 一致的访问接口
 */

export { usePortal, type Portal, type ConfirmOptions } from './portal'

// 禁止导出子 store，强制通过 Portal 访问
// export { useAuthStore } from './auth'      // ❌ 禁止
// export { useSystemStore } from './system'  // ❌ 禁止
// export { useUIStore } from './ui'          // ❌ 禁止
// export { useFilerStore } from './filer'    // ❌ 禁止
