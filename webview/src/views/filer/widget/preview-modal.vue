<script lang="ts">
import { Component, Vue, toNative } from 'vue-facing-decorator'

import api from '@/service/api'
import type { FilerFileInfo } from '@/service/types'

@Component({
    expose: ['show']
})
class PreviewModal extends Vue {
    // ─── 数据属性 ───
    isOpen = false
    filename = ''
    imageUrl = ''
    loading = false
    error = ''

    // ─── 方法 ───
    async show(file: FilerFileInfo) {
        this.filename = file.name
        this.imageUrl = ''
        this.error = ''
        this.loading = true
        this.isOpen = true
        try {
            const blob = await api.filerDownload(file.path)
            const ext = file.name.split('.').pop()?.toLowerCase() ?? ''
            const mimeMap: Record<string, string> = {
                jpg: 'image/jpeg', jpeg: 'image/jpeg', png: 'image/png',
                gif: 'image/gif', bmp: 'image/bmp', svg: 'image/svg+xml',
                webp: 'image/webp', ico: 'image/x-icon',
                tiff: 'image/tiff', tif: 'image/tiff'
            }
            const mime = mimeMap[ext] || 'image/*'
            const typed = new Blob([blob], { type: mime })
            this.imageUrl = URL.createObjectURL(typed)
        } catch {
            this.error = '图片加载失败'
        } finally {
            this.loading = false
        }
    }

    close() {
        this.isOpen = false
        if (this.imageUrl) {
            URL.revokeObjectURL(this.imageUrl)
            this.imageUrl = ''
        }
    }
}

export default toNative(PreviewModal)
</script>

<template>
  <Teleport to="body">
    <Transition name="modal-fade">
      <div
        v-if="isOpen"
        class="fixed inset-0 z-50 flex items-center justify-center p-4"
        @click.self="close"
      >
        <!-- Backdrop -->
        <div class="absolute inset-0 bg-black/70 backdrop-blur-sm" @click="close"></div>

        <!-- Panel -->
        <div class="relative z-10 flex flex-col bg-white rounded-2xl shadow-2xl max-w-4xl w-full max-h-[90vh] overflow-hidden">
          <!-- Header -->
          <div class="flex items-center justify-between px-5 py-3 border-b border-slate-200 flex-shrink-0">
            <div class="flex items-center gap-2 min-w-0">
              <i class="fas fa-image text-blue-400 flex-shrink-0"></i>
              <span class="text-sm font-medium text-slate-800 truncate">{{ filename }}</span>
            </div>
            <button
              class="w-8 h-8 rounded-lg flex items-center justify-center text-slate-400 hover:bg-slate-100 hover:text-slate-600 transition-colors flex-shrink-0"
              title="关闭"
              @click="close"
            >
              <i class="fas fa-xmark"></i>
            </button>
          </div>

          <!-- Body -->
          <div class="flex-1 overflow-auto flex items-center justify-center bg-slate-50 p-4" style="min-height: 200px;">
            <!-- Loading -->
            <div v-if="loading" class="flex flex-col items-center gap-3 py-10">
              <div class="w-10 h-10 spinner"></div>
              <span class="text-sm text-slate-500">加载中...</span>
            </div>

            <!-- Error -->
            <div v-else-if="error" class="flex flex-col items-center gap-3 py-10">
              <div class="w-14 h-14 rounded-xl bg-red-100 flex items-center justify-center">
                <i class="fas fa-circle-exclamation text-2xl text-red-400"></i>
              </div>
              <span class="text-sm text-red-500">{{ error }}</span>
            </div>

            <!-- Image -->
            <img
              v-else-if="imageUrl"
              :src="imageUrl"
              :alt="filename"
              class="max-w-full max-h-full object-contain rounded-lg shadow-sm select-none"
              style="max-height: calc(90vh - 120px);"
              draggable="false"
            />
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}
.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
</style>
