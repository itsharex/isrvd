<script lang="ts">
import { Component, Inject, Vue, toNative } from 'vue-facing-decorator'

import { APP_ACTIONS_KEY } from '@/store/state'
import type { AppActions } from '@/store/state'

import api from '@/service/api'
import type { ApiTokenCreateResponse } from '@/service/types'

@Component
class ApiToken extends Vue {
    @Inject({ from: APP_ACTIONS_KEY }) readonly actions!: AppActions

    // ─── 数据属性 ───
    form = {
        name: '',
        expiresIn: 0 // 0 = 永不过期
    }
    loading = false
    newToken: ApiTokenCreateResponse | null = null

    // ─── 过期选项 ───
    expiryOptions = [
        { label: '永不过期', value: 0 },
        { label: '1 小时', value: 3600 },
        { label: '24 小时', value: 86400 },
        { label: '7 天', value: 604800 },
        { label: '30 天', value: 2592000 },
        { label: '90 天', value: 7776000 },
        { label: '365 天', value: 31536000 }
    ]

    // ─── 方法 ───
    async handleCreate() {
        if (!this.form.name.trim()) {
            this.actions.showNotification('error', '请输入令牌名称')
            return
        }
        this.loading = true
        try {
            const res = await api.createApiToken({
                name: this.form.name.trim(),
                expiresIn: this.form.expiresIn
            })
            this.newToken = res.payload ?? null
            this.form.name = ''
            this.actions.showNotification('success', '令牌创建成功')
        } catch {
            this.actions.showNotification('error', '令牌创建失败')
        }
        this.loading = false
    }

    copyToken(token: string) {
        navigator.clipboard.writeText(token).then(() => {
            this.actions.showNotification('success', '令牌已复制到剪贴板')
        }).catch(() => {
            this.actions.showNotification('error', '复制失败，请手动复制')
        })
    }

    dismissNewToken() {
        this.newToken = null
    }
}

export default toNative(ApiToken)
</script>

<template>
  <div>
    <div class="card mb-4">
      <!-- Toolbar Bar -->
      <div class="bg-slate-50 border-b border-slate-200 rounded-t-2xl px-4 md:px-6 py-3">
        <!-- 桌面端 -->
        <div class="hidden md:flex items-center gap-3">
          <div class="w-9 h-9 rounded-lg bg-emerald-500 flex items-center justify-center">
            <i class="fas fa-key text-white"></i>
          </div>
          <div>
            <h1 class="text-lg font-semibold text-slate-800">API 令牌</h1>
            <p class="text-xs text-slate-500">创建用于 API 访问的个人访问令牌</p>
          </div>
        </div>
        <!-- 移动端 -->
        <div class="flex md:hidden items-center gap-3">
          <div class="w-9 h-9 rounded-lg bg-emerald-500 flex items-center justify-center flex-shrink-0">
            <i class="fas fa-key text-white"></i>
          </div>
          <div class="min-w-0">
            <h1 class="text-lg font-semibold text-slate-800 truncate">API 令牌</h1>
            <p class="text-xs text-slate-500 truncate">创建个人访问令牌</p>
          </div>
        </div>
      </div>

      <!-- 新令牌提示 -->
      <div v-if="newToken" class="mx-4 md:mx-6 mt-4 rounded-xl border border-emerald-200 bg-emerald-50 p-4">
        <div class="flex items-start gap-3">
          <div class="w-10 h-10 rounded-lg bg-emerald-500 flex items-center justify-center flex-shrink-0">
            <i class="fas fa-check text-white"></i>
          </div>
          <div class="flex-1 min-w-0">
            <h3 class="text-sm font-semibold text-emerald-800 mb-1">令牌创建成功</h3>
            <p class="text-xs text-emerald-700 mb-3">请立即复制保存，此令牌仅显示一次：</p>
            <div class="flex items-center gap-2">
              <code class="flex-1 px-3 py-2 bg-white rounded-lg text-xs font-mono text-slate-700 break-all border border-emerald-200">{{ newToken.token }}</code>
              <button @click="copyToken(newToken.token)" class="flex-shrink-0 px-3 py-2 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-white text-xs font-medium flex items-center gap-1.5 transition-colors">
                <i class="fas fa-copy"></i>复制
              </button>
            </div>
          </div>
          <button @click="dismissNewToken()" class="flex-shrink-0 w-8 h-8 rounded-lg text-emerald-600 hover:bg-emerald-100 flex items-center justify-center transition-colors">
            <i class="fas fa-times"></i>
          </button>
        </div>
      </div>

      <!-- 创建表单 -->
      <div class="p-4 md:p-6">
        <form @submit.prevent="handleCreate" class="max-w-3xl space-y-4">
          <div>
            <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">令牌名称</label>
            <input
              v-model="form.name"
              type="text"
              class="input"
              placeholder="如：CI/CD Pipeline"
              maxlength="64"
            />
            <p class="text-xs text-slate-400 mt-1">用于标识令牌用途，最长 64 字符</p>
          </div>

          <div>
            <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">有效期</label>
            <select v-model="form.expiresIn" class="input">
              <option v-for="opt in expiryOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
            </select>
          </div>

          <div class="flex items-center gap-3 pt-2">
            <button
              type="submit"
              :disabled="loading || !form.name.trim()"
              class="px-4 py-2 rounded-lg bg-emerald-500 hover:bg-emerald-600 disabled:bg-slate-300 disabled:cursor-not-allowed text-white text-sm font-medium flex items-center gap-2 transition-colors"
            >
              <i class="fas fa-plus" v-if="!loading"></i>
              <i class="fas fa-spinner fa-spin" v-else></i>
              创建令牌
            </button>
          </div>
        </form>
      </div>

      <!-- 使用说明 -->
      <div class="mx-4 md:mx-6 mb-4 rounded-xl border border-slate-200 bg-slate-50 p-4">
        <h3 class="text-xs font-semibold text-slate-600 mb-3 flex items-center gap-1.5">
          <i class="fas fa-circle-info text-slate-400"></i>使用说明
        </h3>
        <div class="space-y-3">
          <div>
            <p class="text-xs font-medium text-slate-600 mb-1">调用方式</p>
            <p class="text-xs text-slate-500">在 HTTP 请求头中添加：<code class="px-1.5 py-0.5 bg-white rounded text-slate-700">Authorization: Bearer &lt;token&gt;</code></p>
          </div>
          <div>
            <p class="text-xs font-medium text-slate-600 mb-1">安全提示</p>
            <ul class="text-xs text-slate-500 space-y-0.5">
              <li>• 令牌创建后仅显示一次，请立即复制保存</li>
              <li>• 令牌基于 JWT 签发，无法撤销，请妥善保管</li>
              <li>• 如需更换令牌，请重新创建新令牌</li>
            </ul>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
