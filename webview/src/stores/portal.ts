import { defineStore, storeToRefs } from 'pinia'

import { interceptors } from '@/service/axios'

import { useAuthStore } from './auth'
import { useSystemStore } from './system'
import { useUIStore, type ConfirmOptions } from './ui'
import { useFilerStore } from './filer'

export type { ConfirmOptions }

/**
 * Portal Store - 统一入口
 * 
 * 组合 auth、system、ui、filer 四个子 store，提供统一的访问接口
 * 主要职责：
 * 1. 协调各子 store 的初始化
 * 2. 注册 axios 拦截器
 * 3. 提供统一的访问入口，隔离内部实现
 */
export const usePortal = defineStore('portal', () => {
    // ─── 引用子 Store ───
    
    const authStore = useAuthStore()
    const systemStore = useSystemStore()
    const uiStore = useUIStore()
    const filerStore = useFilerStore()

    // ─── 初始化 ───

    async function initialize() {
        try {
            // 1. 加载认证信息
            await authStore.loadAuth()

            // 2. 如果已登录，加载系统数据
            if (authStore.username) {
                await systemStore.initialize()
            } else {
                // 未登录也要标记为已初始化
                systemStore.initialized = true
            }
        } catch (e) {
            console.error('Portal initialize failed:', e)
            systemStore.initialized = true
            throw e
        }
    }

    // ─── 权限检查（组合 auth 和 system）───

    function hasPerm(module: string): boolean {
        return systemStore.hasPerm(module, authStore.founder, authStore.permissions)
    }

    // ─── 注册 Axios 拦截器 ───
    
    const stateProxy = {
        get token() { return authStore.token },
        set token(val: string | null) { authStore.token = val },
        get loading() { return uiStore.loading },
        set loading(val: boolean) { uiStore.loading = val }
    }
    
    interceptors(stateProxy, {
        showNotification: uiStore.showNotification,
        clearAuth: authStore.clearAuth
    })

    // ─── 导出统一接口 ───

    return {
        // Auth Store 状态和方法
        ...storeToRefs(authStore),
        setAuth: authStore.setAuth,
        clearAuth: authStore.clearAuth,
        isAuthenticated: authStore.isAuthenticated,

        // System Store 状态和方法
        ...storeToRefs(systemStore),
        loadSystemData: systemStore.loadSystemData,
        hasPerm,

        // UI Store 状态和方法
        ...storeToRefs(uiStore),
        showNotification: uiStore.showNotification,
        clearNotification: uiStore.clearNotification,
        showConfirm: uiStore.showConfirm,
        confirmLoading: uiStore.confirmLoading,
        closeConfirm: uiStore.closeConfirm,
        handleConfirm: uiStore.handleConfirm,

        // Filer Store 状态和方法
        ...storeToRefs(filerStore),
        loadFiles: filerStore.loadFiles,

        // Portal 方法
        initialize
    }
})

// ─── 类型导出 ───
export type Portal = ReturnType<typeof usePortal>
