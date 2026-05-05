<script lang="ts">
import { Component, Inject, Vue, toNative } from 'vue-facing-decorator'

import { APP_STATE_KEY } from '@/store/state'
import type { AppState } from '@/store/state'

import { FILER_STATE_KEY, FILER_ACTIONS_KEY } from '@/store/filer'
import type { FilerActions, FilerState } from '@/store/filer'

import api from '@/service/api'

import BaseModal from '@/component/modal.vue'

@Component({
    expose: ['show'],
    components: { BaseModal }
})
class CreateModal extends Vue {
    @Inject({ from: APP_STATE_KEY }) readonly appState!: AppState
    @Inject({ from: FILER_STATE_KEY }) readonly filerState!: FilerState
    @Inject({ from: FILER_ACTIONS_KEY }) readonly filerActions!: FilerActions

    // ─── 数据属性 ───
    isOpen = false
    formData = { name: '', content: '' }

    // ─── 方法 ───
    show() {
        this.formData = { name: '', content: '' }
        this.isOpen = true
    }

    async handleConfirm() {
        if (!this.formData.name.trim()) return
        await api.filerCreate(this.filerState.currentPath + '/' + this.formData.name, this.formData.content)
        this.filerActions.loadFiles()
        this.isOpen = false
    }
}

export default toNative(CreateModal)
</script>

<template>
  <BaseModal ref="modalRef" v-model="isOpen" title="新建文件" :loading="appState.loading" :confirm-disabled="!formData.name.trim()" @confirm="handleConfirm">
    <form class="space-y-5" @submit.prevent="handleConfirm">
      <div>
        <label for="fileName" class="block text-sm font-medium text-slate-700 mb-2">
          文件名称
        </label>
        <div class="relative">
          <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
            <i class="fas fa-file text-slate-400"></i>
          </div>
          <input 
            id="fileName" 
            v-model="formData.name" 
            type="text" 
            :disabled="appState.loading" 
            required
            class="input pl-11"
            placeholder="请输入文件名称"
          >
        </div>
      </div>
      <div>
        <label for="fileContent" class="block text-sm font-medium text-slate-700 mb-2">
          文件内容
        </label>
        <textarea 
          id="fileContent" 
          v-model="formData.content" 
          rows="10" 
          :disabled="appState.loading"
          class="input font-mono text-sm"
          placeholder="请输入文件内容..."
        ></textarea>
      </div>
    </form>
    <template #confirm-text>
      {{ appState.loading ? '创建中...' : '创建文件' }}
    </template>
  </BaseModal>
</template>
