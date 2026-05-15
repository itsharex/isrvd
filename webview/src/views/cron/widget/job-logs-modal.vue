<script lang="ts">
import { Component, Vue, toNative } from 'vue-facing-decorator'

import api from '@/service/api'
import type { CronJob, CronJobLog } from '@/service/types'

import BaseModal from '@/component/modal.vue'

import { usePortal } from '@/stores'

@Component({
    expose: ['show'],
    components: { BaseModal }
})
class JobLogsModal extends Vue {
    portal = usePortal()

    isOpen = false
    loading = false
    job: CronJob | null = null
    logs: CronJobLog[] = []

    async show(job: CronJob) {
        this.job = job
        this.logs = []
        this.isOpen = true
        this.loading = true
        try {
            const res = await api.cronJobLogs(job.id, 50)
            this.logs = res.payload?.logs || []
        } catch {
            this.portal.showNotification('error', '获取执行日志失败')
        } finally {
            this.loading = false
        }
    }

    formatTime(t?: string): string {
        if (!t) return '-'
        return new Date(t).toLocaleString('zh-CN')
    }

    formatDuration(ms: number): string {
        return ms < 1000 ? `${ms}ms` : `${(ms / 1000).toFixed(2)}s`
    }
}

export default toNative(JobLogsModal)
</script>

<template>
  <BaseModal v-model="isOpen" :show-footer="false" max-width-class="max-w-3xl">
    <template #title>
      <div class="flex items-center gap-2">
        <i class="fas fa-list-ul text-blue-500"></i>
        <span class="text-base font-semibold">执行历史 — {{ job?.name }}</span>
      </div>
    </template>

    <div v-if="loading" class="flex items-center justify-center py-12">
      <div class="w-8 h-8 spinner"></div>
    </div>
    <div v-else-if="logs.length === 0" class="flex flex-col items-center justify-center py-12">
      <i class="fas fa-list-ul text-3xl text-slate-300 mb-2"></i>
      <p class="text-slate-500 text-sm">暂无执行记录</p>
    </div>
    <div v-else class="space-y-3">
      <div
        v-for="(log, idx) in logs"
        :key="log.runId || idx"
        class="rounded-xl border p-4"
        :class="log.success ? 'border-emerald-100 bg-emerald-50/30' : 'border-red-100 bg-red-50/30'"
      >
        <div class="flex items-center justify-between mb-2">
          <div class="flex items-center gap-2">
            <span v-if="log.success" class="inline-flex items-center gap-1 text-xs font-medium text-emerald-600">
              <i class="fas fa-circle-check"></i>成功
            </span>
            <span v-else class="inline-flex items-center gap-1 text-xs font-medium text-red-600">
              <i class="fas fa-circle-xmark"></i>失败
            </span>
            <span class="text-xs text-slate-400">{{ formatDuration(log.duration) }}</span>
          </div>
          <span class="text-xs text-slate-400 whitespace-nowrap">{{ formatTime(log.startTime) }}</span>
        </div>
        <pre v-if="log.output" class="text-xs text-slate-700 font-mono bg-slate-900/5 rounded-lg p-3 overflow-x-auto whitespace-pre-wrap break-all max-h-40">{{ log.output }}</pre>
        <pre v-if="log.error" class="mt-2 text-xs text-red-700 font-mono bg-red-900/5 rounded-lg p-3 overflow-x-auto whitespace-pre-wrap break-all max-h-24">{{ log.error }}</pre>
      </div>
    </div>
  </BaseModal>
</template>
