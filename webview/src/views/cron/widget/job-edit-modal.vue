<script lang="ts">
import { Codemirror } from 'vue-codemirror'
import { Component, Vue, toNative } from 'vue-facing-decorator'

import api from '@/service/api'
import type { CronJob, CronJobCreate, CronTypeInfo } from '@/service/types'

import BaseModal from '@/component/modal.vue'

import { usePortal } from '@/stores'

const defaultFormData = (type = 'SHELL'): CronJobCreate => ({
    name: '',
    schedule: '',
    type: type as CronJobCreate['type'],
    content: '',
    workDir: '',
    timeout: 0,
    enabled: true,
    description: ''
})

@Component({
    expose: ['show'],
    components: { BaseModal, Codemirror },
    emits: ['success']
})
class JobEditModal extends Vue {
    portal = usePortal()

    isOpen = false
    modalLoading = false
    isEditMode = false
    jobID = ''
    types: CronTypeInfo[] = []
    formData = defaultFormData()

    show(job: CronJob | null = null, types: CronTypeInfo[] = []) {
        this.types = types
        this.isEditMode = !!job
        if (job) {
            this.jobID = job.id
            this.formData = {
                name: job.name,
                schedule: job.schedule,
                type: job.type,
                content: job.content,
                workDir: job.workDir,
                timeout: job.timeout,
                enabled: job.enabled,
                description: job.description
            }
        } else {
            this.jobID = ''
            this.formData = defaultFormData(types[0]?.value)
        }
        this.isOpen = true
    }

    async handleConfirm() {
        if (!this.formData.name || !this.formData.schedule || !this.formData.content) {
            this.portal.showNotification('error', '请填写必填项：名称、Cron 表达式、脚本内容')
            return
        }
        this.modalLoading = true
        try {
            if (this.isEditMode) {
                await api.cronJobUpdate(this.jobID, this.formData)
                this.portal.showNotification('success', '任务已更新')
            } else {
                await api.cronJobCreate(this.formData)
                this.portal.showNotification('success', '任务已创建')
            }
            this.isOpen = false
            this.$emit('success')
        } catch {
            this.portal.showNotification('error', this.isEditMode ? '更新任务失败' : '创建任务失败')
        } finally {
            this.modalLoading = false
        }
    }
}

export default toNative(JobEditModal)
</script>

<template>
  <BaseModal
    v-model="isOpen"
    :title="isEditMode ? '编辑计划任务' : '新建计划任务'"
    :loading="modalLoading"
    max-width-class="max-w-2xl"
    @confirm="handleConfirm"
  >
    <template #confirm-text>{{ isEditMode ? '保存' : '创建' }}</template>

    <div class="space-y-4 p-1">
      <div>
        <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">任务名称 <span class="text-red-500">*</span></label>
        <input v-model="formData.name" type="text" class="input" placeholder="如：每日备份" />
      </div>

      <div>
        <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">描述</label>
        <textarea v-model="formData.description" rows="2" class="input resize-none" placeholder="可选说明"></textarea>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">
            Cron 表达式 <span class="text-red-500">*</span>
            <a href="https://crontab.guru" target="_blank" rel="noreferrer" class="ml-1 text-primary-500 hover:underline normal-case font-normal">参考</a>
          </label>
          <input v-model="formData.schedule" type="text" class="input font-mono" placeholder="如：0 2 * * *" />
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">脚本类型 <span class="text-red-500">*</span></label>
          <select v-model="formData.type" class="input">
            <option v-for="t in types" :key="t.value" :value="t.value">{{ t.label }}</option>
          </select>
        </div>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">工作目录</label>
          <input v-model="formData.workDir" type="text" class="input font-mono" placeholder="可选，默认当前目录" />
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">超时时间（秒）</label>
          <input v-model.number="formData.timeout" type="number" min="0" class="input" placeholder="0 表示不限制" />
        </div>
      </div>

      <div>
        <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">脚本内容 <span class="text-red-500">*</span></label>
        <div class="rounded-xl overflow-hidden border border-slate-200">
          <Codemirror
            v-model="formData.content"
            :style="{ height: '240px' }"
            :disabled="modalLoading"
            placeholder="输入脚本内容，如：#!/bin/bash&#10;echo &quot;hello&quot;"
          />
        </div>
      </div>

      <div>
        <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">状态</label>
        <select v-model="formData.enabled" class="input">
          <option :value="true">启用</option>
          <option :value="false">禁用</option>
        </select>
      </div>
    </div>
  </BaseModal>
</template>
