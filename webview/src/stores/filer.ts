import { defineStore } from 'pinia'
import { ref } from 'vue'

import api from '@/service/api'
import type { FilerFileInfo } from '@/service/types'

export const useFilerStore = defineStore('filer', () => {
    // ─── 状态定义 ───

    const currentPath = ref('/')
    const files = ref<FilerFileInfo[]>([])
    const isLoading = ref(false)

    // ─── 操作定义 ───

    async function loadFiles(path?: string) {
        const targetPath = path ?? currentPath.value
        isLoading.value = true
        try {
            const res = await api.filerList(targetPath)
            files.value = res.payload?.files || []
            currentPath.value = res.payload?.path ?? '/'
        } catch {
            files.value = []
        } finally {
            isLoading.value = false
        }
    }

    return {
        // 状态
        currentPath,
        files,
        isLoading,
        // 操作
        loadFiles
    }
})
