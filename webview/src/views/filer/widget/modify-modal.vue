<script lang="ts">
import { css } from '@codemirror/lang-css'
import { go } from '@codemirror/lang-go'
import { html } from '@codemirror/lang-html'
import { javascript } from '@codemirror/lang-javascript'
import { json } from '@codemirror/lang-json'
import { markdown } from '@codemirror/lang-markdown'
import { python } from '@codemirror/lang-python'
import { sql } from '@codemirror/lang-sql'
import { xml } from '@codemirror/lang-xml'
import { yaml } from '@codemirror/lang-yaml'
import { Codemirror } from 'vue-codemirror'
import { Component, Inject, Vue, toNative } from 'vue-facing-decorator'

import { APP_STATE_KEY } from '@/store/state'
import type { AppState } from '@/store/state'

import { FILER_ACTIONS_KEY } from '@/store/filer'
import type { FilerActions } from '@/store/filer'

import api from '@/service/api'
import type { FilerFileInfo } from '@/service/types'

import BaseModal from '@/component/modal.vue'

@Component({
    expose: ['show'],
    components: { BaseModal, Codemirror }
})
class ModifyModal extends Vue {
    @Inject({ from: APP_STATE_KEY }) readonly appState!: AppState
    @Inject({ from: FILER_ACTIONS_KEY }) readonly filerActions!: FilerActions

    // ─── 数据属性 ───
    isOpen = false
    formData = { filename: '', content: '', path: '' }
    readonly extensions = [css(), go(), html(), javascript(), json(), markdown(), python(), sql(), xml(), yaml()]

    // ─── 方法 ───
    async show(file: FilerFileInfo) {
        const res = await api.filerRead(file.path)
        this.formData.path = file.path
        this.formData.filename = file.name
        this.formData.content = res.payload?.content ?? ''
        this.isOpen = true
    }

    async handleConfirm() {
        await api.filerModify(this.formData.path, this.formData.content)
        this.filerActions.loadFiles()
        this.isOpen = false
    }
}

export default toNative(ModifyModal)
</script>

<template>
  <BaseModal ref="modalRef" v-model="isOpen" :title="'编辑: ' + formData.filename" :loading="appState.loading" @confirm="handleConfirm">
    <div class="rounded-xl overflow-hidden border border-slate-200">
      <Codemirror v-model="formData.content" :style="{ height: '60vh' }" :extensions="extensions" :disabled="appState.loading" />
    </div>
    <template #confirm-text>
      {{ appState.loading ? '保存中...' : '保存文件' }}
    </template>
  </BaseModal>
</template>
