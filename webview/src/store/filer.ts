import { reactive } from 'vue'

import api from '@/service/api'
import type { FilerFileInfo } from '@/service/types'

export const FILER_STATE_KEY = 'filer.state'
export const FILER_ACTIONS_KEY = 'filer.actions'

// ─── 类型定义 ───

export interface FilerState {
    currentPath: string
    files: FilerFileInfo[]
    loading: boolean
}

export interface FilerActions {
    loadFiles(path?: string): Promise<void>
}

// ─── 初始状态 ───

const createInitialState = (): FilerState => ({
    currentPath: '/',
    files: [],
    loading: false,
})

// ─── Provider ───

export const initFilerProvider = () => {
    const state = reactive<FilerState>(createInitialState())

    const actions: FilerActions = {
        async loadFiles(path = state.currentPath) {
            state.loading = true
            try {
                const res = await api.filerList(path)
                state.files = res.payload?.files || []
                state.currentPath = res.payload?.path ?? '/'
            } catch {
                state.files = []
            } finally {
                state.loading = false
            }
        },
    }

    return { state, actions }
}
