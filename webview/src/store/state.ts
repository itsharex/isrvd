import { reactive } from 'vue'

import { interceptors } from '@/service/axios'
import api from '@/service/api'
import type { FilerFileInfo, LinkConfig } from '@/service/types'

export const APP_STATE_KEY = 'app.state'
export const APP_ACTIONS_KEY = 'app.actions'

// ─── 类型定义 ───

interface Notification {
    id: number
    type: string
    message: string
    timer: ReturnType<typeof setTimeout>
}

export interface ConfirmOptions {
    title?: string
    message?: string
    icon?: string
    iconColor?: string
    confirmText?: string
    danger?: boolean
    onConfirm?: (() => void | Promise<void>) | null
}

interface ServiceAvailability {
    agent: boolean
    apisix: boolean
    docker: boolean
    swarm: boolean
    compose: boolean
}

export interface AppState {
    initialized: boolean
    initError: string | null
    authMode: 'jwt' | 'header' | null
    token: string | null
    username: string | null
    permissionsLoaded: boolean
    founder: boolean
    permissions: string[]
    serviceAvailability: ServiceAvailability
    toolbarLinks: LinkConfig[]
    loading: boolean
    currentPath: string
    files: FilerFileInfo[]
    notifications: Notification[]
    confirm: {
        show: boolean
        title: string
        message: string
        icon: string
        iconColor: string
        confirmText: string
        danger: boolean
        loading: boolean
        onConfirm: (() => void | Promise<void>) | null
    }
}

export interface AppActions {
    initialize(): Promise<void>
    loadAppData(): Promise<void>
    setAuth(data: { authMode: 'jwt' | 'header'; token: string; username: string }): void
    clearAuth(): void
    isAuthenticated(): boolean
    hasPerm(module: string): boolean
    loadFiles(path?: string): Promise<void>
    showNotification(type: string, message: string): void
    clearNotification(id: number): void
    showConfirm(options: ConfirmOptions): void
    confirmLoading(loading: boolean): void
    closeConfirm(): void
    handleConfirm(): Promise<void>
}

// ─── 初始状态 ───

const createInitialState = (): AppState => ({
    initialized: false,
    initError: null,
    authMode: null,
    token: null,
    username: null,
    permissionsLoaded: false,
    founder: false,
    permissions: [],
    serviceAvailability: { agent: false, apisix: false, docker: false, swarm: false, compose: false },
    toolbarLinks: [],
    loading: false,
    currentPath: '/',
    files: [],
    notifications: [],
    confirm: { show: false, title: '', message: '', icon: '', iconColor: 'blue', confirmText: '确认', danger: false, loading: false, onConfirm: null }
})

// ─── Provider ───

export const initProvider = () => {
    const state = reactive<AppState>(createInitialState())

    const actions: AppActions = {
        async initialize() {
            state.initialized = false
            state.initError = null

            try {
                // 恢复 token
                const token = localStorage.getItem('app-token')
                const username = localStorage.getItem('app-username')
                if (token && username) {
                    state.token = token
                    state.username = username
                    state.authMode = 'jwt'
                }

                // 验证认证
                const authRes = await api.accountInfo()
                const payload = authRes?.payload

                if (payload?.mode === 'header' && payload.username) {
                    this.setAuth({ authMode: 'header', token: '', username: payload.username })
                } else if (payload?.username && payload.member) {
                    state.permissionsLoaded = true
                    state.founder = payload.member.founder || false
                    state.permissions = payload.member.permissions || []
                } else {
                    this.clearAuth()
                }

                if (state.username) await this.loadAppData()
            } catch (e) {
                console.error('Initialize failed:', e)
                state.initError = e instanceof Error ? e.message : '初始化失败'
            } finally {
                state.initialized = true
            }
        },

        async loadAppData() {
            try {
                const [probeRes, configRes] = await Promise.all([
                    api.overviewProbe(),
                    api.systemConfig(),
                ])

                const probe = probeRes?.payload
                if (probe) {
                    state.serviceAvailability = {
                        agent: probe.agent?.available || false,
                        apisix: probe.apisix?.available || false,
                        docker: probe.docker?.available || false,
                        swarm: probe.swarm?.available || false,
                        compose: probe.compose?.available || false
                    }
                }

                state.toolbarLinks = configRes?.payload?.links || []
            } catch (e) {
                console.error('Load app data failed:', e)
            }
        },

        setAuth(data) {
            state.authMode = data.authMode
            state.username = data.username
            state.token = data.token || ''
            if (data.authMode === 'jwt') {
                localStorage.setItem('app-token', data.token)
                localStorage.setItem('app-username', data.username)
            }
        },

        clearAuth() {
            state.authMode = null
            state.token = null
            state.username = null
            state.permissionsLoaded = false
            state.founder = false
            state.permissions = []
            localStorage.removeItem('app-token')
            localStorage.removeItem('app-username')
        },

        isAuthenticated: () => !!state.username,

        hasPerm(module) {
            const checkAvailability = (seg: string): boolean => {
                const key = seg as keyof ServiceAvailability
                return !(key in state.serviceAvailability && !state.serviceAvailability[key])
            }

            // 精确路由匹配
            if (module.includes(' ')) {
                const path = module.split(' ')[1]
                const seg = path?.match(/^\/api\/([^/]+)/)?.[1]
                if (seg && !checkAvailability(seg)) return false
                return state.founder || state.permissions.includes(module)
            }

            // 模块匹配
            if (!checkAvailability(module)) return false
            if (state.founder) return true
            return state.permissions.some(key => {
                const path = key.split(' ')[1]
                return path && (path.startsWith(`/api/${module}/`) || path === `/api/${module}`)
            })
        },

        async loadFiles(path = state.currentPath) {
            console.log('loadFiles:', path)
        },

        showNotification(type, message) {
            if (!message) return
            const id = Date.now() + Math.random()
            const timer = setTimeout(() => this.clearNotification(id), 5000)
            state.notifications.push({ id, type, message, timer })
        },

        clearNotification(id) {
            const idx = state.notifications.findIndex(n => n.id === id)
            if (idx !== -1) {
                clearTimeout(state.notifications[idx].timer)
                state.notifications.splice(idx, 1)
            }
        },

        showConfirm(options) {
            state.confirm = {
                show: true,
                title: options.title || '确认操作',
                message: options.message || '',
                icon: options.icon || 'fa-question-circle',
                iconColor: options.iconColor || 'blue',
                confirmText: options.confirmText || '确认',
                danger: options.danger || false,
                loading: false,
                onConfirm: options.onConfirm || null
            }
        },

        confirmLoading(loading) {
            state.confirm.loading = loading
        },

        closeConfirm() {
            state.confirm.show = false
            state.confirm.loading = false
            state.confirm.onConfirm = null
        },

        async handleConfirm() {
            if (!state.confirm.onConfirm) return this.closeConfirm()
            state.confirm.loading = true
            try {
                await state.confirm.onConfirm()
            } finally {
                state.confirm.loading = false
            }
            this.closeConfirm()
        }
    }

    interceptors(state, actions)
    return { state, actions }
}
