<script lang="ts">
import { Component, Vue, Watch, toNative } from 'vue-facing-decorator'

import api from '@/service/api'
import type { FilerFileInfo } from '@/service/types'

import { getImageMimeType } from '@/helper/utils'

import BaseModal from '@/component/modal.vue'

@Component({
    expose: ['show'],
    components: { BaseModal }
})
class PreviewModal extends Vue {
    // ─── 数据属性 ───
    isOpen = false
    filename = ''
    imageUrl = ''
    loading = false
    error = ''

    // ─── 监听器 ───
    @Watch('isOpen')
    onOpenChange(value: boolean) {
        if (!value) this.releaseImageUrl()
    }

    // ─── 方法 ───
    async show(file: FilerFileInfo) {
        this.releaseImageUrl()
        this.filename = file.name
        this.error = ''
        this.loading = true
        this.isOpen = true
        try {
            const blob = await api.filerDownload(file.path)
            const typed = new Blob([blob], { type: getImageMimeType(file.name) })
            this.imageUrl = URL.createObjectURL(typed)
        } catch {
            this.error = '图片加载失败'
        } finally {
            this.loading = false
        }
    }

    close() {
        this.isOpen = false
        this.releaseImageUrl()
    }

    releaseImageUrl() {
        if (this.imageUrl) {
            URL.revokeObjectURL(this.imageUrl)
            this.imageUrl = ''
        }
    }
}

export default toNative(PreviewModal)
</script>

<template>
  <BaseModal v-model="isOpen" :title="'预览: ' + filename" :show-footer="false" @cancel="close">
    <div class="flex min-h-[200px] items-center justify-center rounded-xl bg-slate-50 p-4">
      <div v-if="loading" class="flex flex-col items-center gap-3 py-10">
        <div class="w-12 h-12 spinner"></div>
        <span class="text-sm text-slate-500">加载中...</span>
      </div>

      <div v-else-if="error" class="flex flex-col items-center gap-3 py-10">
        <div class="w-16 h-16 rounded-lg bg-red-100 flex items-center justify-center">
          <i class="fas fa-circle-exclamation text-4xl text-red-400"></i>
        </div>
        <span class="text-sm text-red-500">{{ error }}</span>
      </div>

      <img
        v-else-if="imageUrl"
        :src="imageUrl"
        :alt="filename"
        class="max-h-[calc(100vh-12rem)] max-w-full object-contain rounded-lg shadow-sm select-none"
        draggable="false"
      />
    </div>
  </BaseModal>
</template>
