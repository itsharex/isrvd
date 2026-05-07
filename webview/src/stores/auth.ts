import { defineStore } from 'pinia'
import { ref } from 'vue'

import api from '@/service/api'
import type { AuthInfo } from '@/service/types'

export const useAuthStore = defineStore('auth', () => {
    // ─── 状态定义 ───

    const authMode = ref<'jwt' | 'header' | null>(null)
    const token = ref<string | null>(null)
    const username = ref<string | null>(null)
    const permissionsLoaded = ref(false)
    const founder = ref(false)
    const permissions = ref<string[]>([])

    // ─── 操作定义 ───

    function setAuth(data: { authMode: 'jwt' | 'header'; token: string; username: string }) {
        authMode.value = data.authMode
        username.value = data.username
        token.value = data.token || ''
        if (data.authMode === 'jwt') {
            localStorage.setItem('app-token', data.token)
            localStorage.setItem('app-username', data.username)
        }
    }

    function clearAuth() {
        authMode.value = null
        token.value = null
        username.value = null
        permissionsLoaded.value = false
        founder.value = false
        permissions.value = []
        localStorage.removeItem('app-token')
        localStorage.removeItem('app-username')
    }

    function isAuthenticated() {
        return !!username.value
    }

    async function loadAuth() {
        // 恢复 token
        const savedToken = localStorage.getItem('app-token')
        const savedUsername = localStorage.getItem('app-username')
        if (savedToken && savedUsername) {
            token.value = savedToken
            username.value = savedUsername
            authMode.value = 'jwt'
        }

        // 验证认证
        const authRes = await api.accountInfo()
        const payload = authRes?.payload || {} as AuthInfo

        // 核心原则：无 username 或无 member = 无权限，直接清理
        if (!payload?.username || !payload?.member) {
            clearAuth()
        } else {
            // 认证模式处理
            if (payload.mode === 'header') {
                setAuth({ authMode: 'header', token: '', username: payload.username })
            }
            // 权限赋值内聚在一处，确保有 member 才写入
            permissionsLoaded.value = true
            founder.value = payload.member.founder || false
            permissions.value = payload.member.permissions || []
        }
    }

    return {
        // 状态
        authMode,
        token,
        username,
        permissionsLoaded,
        founder,
        permissions,
        // 操作
        setAuth,
        clearAuth,
        isAuthenticated,
        loadAuth
    }
})

// ─── 类型导出 ───
export type AuthStore = ReturnType<typeof useAuthStore>
