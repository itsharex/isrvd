<script lang="ts">
import { Component, Vue, toNative } from 'vue-facing-decorator'

import api from '@/service/api'
import type { FilerFileInfo } from '@/service/types'

import BaseModal from '@/component/modal.vue'

import { usePortal } from '@/stores'

@Component({
    expose: ['show'],
    components: { BaseModal }
})
class UnzipModal extends Vue {
    portal = usePortal()

    // ─── 数据属性 ───
    isOpen = false
    formData = { file: null as FilerFileInfo | null }

    // ─── 方法 ───
    show(file: FilerFileInfo) {
        this.formData.file = file
        this.isOpen = true
    }

    async handleConfirm() {
        await api.filerUnzip(this.formData.file?.path ?? '')
        this.portal.loadFiles()
        this.isOpen = false
    }
}

export default toNative(UnzipModal)
</script>

<template>
  <BaseModal ref="modalRef" v-model="isOpen" title="解压确认" :loading="portal.loading" :confirm-disabled="!formData.file" @confirm="handleConfirm">
    <div v-if="formData.file" class="text-center py-6">
      <div class="w-16 h-16 rounded-lg bg-amber-400 flex items-center justify-center mx-auto mb-4 shadow-lg shadow-amber-500/30">
        <i class="fas fa-expand-arrows-alt text-3xl text-white"></i>
      </div>
      <p class="text-lg text-slate-700 mb-2">
        确定要解压 <strong class="text-slate-900">{{ formData.file.name }}</strong> 吗？
      </p>
      <p class="text-sm text-slate-500">文件将解压到当前目录</p>
    </div>

    <template #confirm-text>
      {{ portal.loading ? '解压中...' : '开始解压' }}
    </template>
  </BaseModal>
</template>
