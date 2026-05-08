import { defineStore } from 'pinia'
import { ref } from 'vue'

import api from '@/service/api'
import type { FilerFileInfo } from '@/service/types'

export const useFilerStore = defineStore('filer', () => {
    // ─── 状态定义 ───

    const loading = ref(false)
    const currentPath = ref('/')
    const files = ref<FilerFileInfo[]>([])

    // ─── 操作定义 ───

    async function loadFiles(path?: string) {
        const targetPath = path ?? currentPath.value
        loading.value = true
        try {
            const res = await api.filerList(targetPath)
            files.value = res.payload?.files || []
            currentPath.value = res.payload?.path ?? '/'
        } catch {
            files.value = []
        } finally {
            loading.value = false
        }
    }

    return {
        // 状态
        loading,
        currentPath,
        files,
        // 操作
        loadFiles
    }
})
