<script lang="ts">
import { Component, Vue, Watch, toNative } from 'vue-facing-decorator'

import api from '@/service/api'
import type { FilerFileInfo } from '@/service/types'

import { getPreviewMimeType, getPreviewType } from '@/helper/utils'
import type { PreviewFileType } from '@/helper/utils'

import BaseModal from '@/component/modal.vue'

import { usePortal } from '@/stores'

@Component({
    expose: ['show'],
    components: { BaseModal }
})
class PreviewModal extends Vue {
    portal = usePortal()
    // ─── 数据属性 ───
    isOpen = false
    filename = ''
    previewUrl = ''
    previewType: PreviewFileType = ''
    mimeType = ''
    loading = false
    error = ''

    // ─── 监听器 ───
    @Watch('isOpen')
    onOpenChange(value: boolean) {
        if (!value) this.resetPreview()
    }

    // ─── 方法 ───
    show(file: FilerFileInfo) {
        this.resetPreview()
        this.filename = file.name
        this.previewType = getPreviewType(file.name)
        this.mimeType = getPreviewMimeType(file.name)
        this.error = ''
        this.loading = true
        this.previewUrl = api.filerDownloadURL(file.path, this.portal.token || '', true)
        this.isOpen = true
    }

    close() {
        this.isOpen = false
        this.resetPreview()
    }

    handleLoaded() {
        this.loading = false
    }

    handleError() {
        this.loading = false
        this.error = '文件加载失败'
    }

    resetPreview() {
        this.previewUrl = ''
        this.previewType = ''
        this.mimeType = ''
        this.loading = false
        this.error = ''
    }
}

export default toNative(PreviewModal)
</script>

<template>
  <BaseModal v-model="isOpen" :title="'预览: ' + filename" :show-footer="false" max-width-class="max-w-5xl" @cancel="close">
    <div class="flex min-h-[240px] items-center justify-center rounded-xl bg-slate-50 p-4">
      <div v-if="loading" class="flex flex-col items-center gap-3 py-10">
        <div class="w-12 h-12 spinner"></div>
        <span class="text-sm text-slate-500">加载中...</span>
      </div>

      <div v-if="error" class="flex flex-col items-center gap-3 py-10">
        <div class="w-16 h-16 rounded-lg bg-red-100 flex items-center justify-center">
          <i class="fas fa-circle-exclamation text-4xl text-red-400"></i>
        </div>
        <span class="text-sm text-red-500">{{ error }}</span>
      </div>

      <img
        v-if="previewUrl && previewType === 'image' && !error"
        v-show="!loading"
        :src="previewUrl"
        :alt="filename"
        class="max-h-[calc(100vh-12rem)] max-w-full object-contain rounded-lg shadow-sm select-none"
        draggable="false"
        @load="handleLoaded"
        @error="handleError"
      />

      <audio
        v-if="previewUrl && previewType === 'audio' && !error"
        v-show="!loading"
        class="w-full"
        controls
        preload="metadata"
        @loadedmetadata="handleLoaded"
        @error="handleError"
      >
        <source :src="previewUrl" :type="mimeType" />
      </audio>

      <video
        v-if="previewUrl && previewType === 'video' && !error"
        v-show="!loading"
        class="max-h-[calc(100vh-12rem)] max-w-full rounded-lg bg-slate-900 shadow-sm"
        controls
        preload="metadata"
        @loadedmetadata="handleLoaded"
        @error="handleError"
      >
        <source :src="previewUrl" :type="mimeType" />
      </video>

      <iframe
        v-if="previewUrl && previewType === 'pdf' && !error"
        v-show="!loading"
        :src="previewUrl"
        class="h-[calc(100vh-12rem)] w-full rounded-lg border border-slate-200 bg-white"
        title="文件预览"
        @load="handleLoaded"
        @error="handleError"
      ></iframe>
    </div>
  </BaseModal>
</template>
