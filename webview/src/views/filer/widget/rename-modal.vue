<script lang="ts">
import { Component, Inject, Vue, toNative } from 'vue-facing-decorator'

import { FILER_ACTIONS_KEY } from '@/store/filer'
import type { FilerActions } from '@/store/filer'
import { APP_STATE_KEY } from '@/store/state'
import type { AppState } from '@/store/state'

import api from '@/service/api'
import type { FilerFileInfo } from '@/service/types'

import BaseModal from '@/component/modal.vue'

@Component({
    expose: ['show'],
    components: { BaseModal }
})
class RenameModal extends Vue {
    @Inject({ from: APP_STATE_KEY }) readonly appState!: AppState
    @Inject({ from: FILER_ACTIONS_KEY }) readonly filerActions!: FilerActions

    // ─── 数据属性 ───
    isOpen = false
    formData = { name: '', file: null as FilerFileInfo | null }

    // ─── 方法 ───
    show(file: FilerFileInfo) {
        this.formData.file = file
        this.formData.name = file.name
        this.isOpen = true
    }

    async handleConfirm() {
        if (!this.formData.name.trim() || !this.formData.file) return
        await api.filerRename(this.formData.file.path, this.formData.name)
        this.filerActions.loadFiles()
        this.isOpen = false
    }
}

export default toNative(RenameModal)
</script>

<template>
  <BaseModal ref="modalRef" v-model="isOpen" title="重命名" :loading="appState.loading" :confirm-disabled="!formData.name.trim()" @confirm="handleConfirm">
    <form @submit.prevent="handleConfirm">
      <div>
        <label for="target" class="block text-sm font-medium text-slate-700 mb-2">
          新名称
        </label>
        <div class="relative">
          <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
            <i class="fas fa-pen text-slate-400"></i>
          </div>
          <input 
            id="target" 
            v-model="formData.name" 
            type="text" 
            :disabled="appState.loading" 
            required
            class="input pl-11"
            placeholder="请输入新名称"
          >
        </div>
      </div>
    </form>
    <template #confirm-text>
      {{ appState.loading ? '重命名中...' : '确认重命名' }}
    </template>
  </BaseModal>
</template>
