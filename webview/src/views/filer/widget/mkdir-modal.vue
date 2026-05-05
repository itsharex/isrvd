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
class MkdirModal extends Vue {
    @Inject({ from: APP_STATE_KEY }) readonly appState!: AppState
    @Inject({ from: FILER_STATE_KEY }) readonly filerState!: FilerState
    @Inject({ from: FILER_ACTIONS_KEY }) readonly filerActions!: FilerActions

    // ─── 数据属性 ───
    isOpen = false
    formData = { name: '' }

    // ─── 方法 ───
    show() {
        this.formData.name = ''
        this.isOpen = true
    }

    async handleConfirm() {
        if (!this.formData.name.trim()) return
        await api.filerMkdir(this.filerState.currentPath + '/' + this.formData.name)
        this.filerActions.loadFiles()
        this.isOpen = false
    }
}

export default toNative(MkdirModal)
</script>

<template>
  <BaseModal ref="modalRef" v-model="isOpen" title="新建目录" :loading="appState.loading" :confirm-disabled="!formData.name.trim()" @confirm="handleConfirm">
    <form @submit.prevent="handleConfirm">
      <div>
        <label for="dirName" class="block text-sm font-medium text-slate-700 mb-2">
          目录名称
        </label>
        <div class="relative">
          <div class="absolute inset-y-0 left-0 pl-4 flex items-center pointer-events-none">
            <i class="fas fa-folder text-slate-400"></i>
          </div>
          <input 
            id="dirName" 
            v-model="formData.name" 
            type="text" 
            :disabled="appState.loading" 
            required
            class="input pl-11"
            placeholder="请输入目录名称"
          >
        </div>
      </div>
    </form>
    <template #confirm-text>
      {{ appState.loading ? '创建中...' : '创建目录' }}
    </template>
  </BaseModal>
</template>
