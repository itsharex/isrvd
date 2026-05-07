<script lang="ts">
import { Component, Inject, Vue, toNative } from 'vue-facing-decorator'

import { FILER_STATE_KEY, FILER_ACTIONS_KEY } from '@/store/filer'
import type { FilerActions, FilerState } from '@/store/filer'

import api from '@/service/api'
import type { FilerFileInfo } from '@/service/types'

import BaseModal from '@/component/modal.vue'

@Component({
    expose: ['show'],
    components: { BaseModal }
})
class ZipModal extends Vue {
    @Inject({ from: FILER_STATE_KEY }) readonly filerState!: FilerState
    @Inject({ from: FILER_ACTIONS_KEY }) readonly filerActions!: FilerActions

    // ─── 数据属性 ───
    isOpen = false
    formData = { file: null as FilerFileInfo | null }

    // ─── 方法 ───
    show(file: FilerFileInfo) {
        this.formData.file = file
        this.isOpen = true
    }

    async handleConfirm() {
        await api.filerZip(this.formData.file?.path ?? '')
        this.filerActions.loadFiles()
        this.isOpen = false
    }
}

export default toNative(ZipModal)
</script>

<template>
  <BaseModal ref="modalRef" v-model="isOpen" title="压缩确认" :loading="filerState.loading" :confirm-disabled="!formData.file" @confirm="handleConfirm">
    <div v-if="formData.file" class="text-center py-6">
      <div class="w-16 h-16 rounded-lg bg-amber-400 flex items-center justify-center mx-auto mb-4 shadow-lg shadow-amber-500/30">
        <i class="fas fa-file-archive text-3xl text-white"></i>
      </div>
      <p class="text-lg text-slate-700 mb-2">
        确定要压缩 <strong class="text-slate-900">{{ formData.file.name }}</strong> 吗？
      </p>
      <p class="text-sm text-slate-500">压缩后的文件将保存在当前目录</p>
    </div>
    <template #confirm-text>
      {{ filerState.loading ? '压缩中...' : '开始压缩' }}
    </template>
  </BaseModal>
</template>
