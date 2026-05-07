<script lang="ts">
import { Component, Inject, Ref, Vue, toNative } from 'vue-facing-decorator'

import { APP_ACTIONS_KEY } from '@/store/state'
import type { AppActions } from '@/store/state'

import api from '@/service/api'
import type { DockerImageInfo, DockerRegistryInfo } from '@/service/types'

import { formatFileSize, formatTime } from '@/helper/utils'

import ImageBuildModal from './widget/image-build-modal.vue'
import ImagePullModal from './widget/image-pull-modal.vue'
import ImageTagModal from './widget/image-tag-modal.vue'
import RegistryPushModal from './widget/registry-push-modal.vue'

@Component({
    components: { ImagePullModal, ImageTagModal, ImageBuildModal, RegistryPushModal }
})
class Images extends Vue {
    @Inject({ from: APP_ACTIONS_KEY }) readonly actions!: AppActions

    // ─── Refs ───
    @Ref readonly pullModalRef!: InstanceType<typeof ImagePullModal>
    @Ref readonly tagModalRef!: InstanceType<typeof ImageTagModal>
    @Ref readonly buildModalRef!: InstanceType<typeof ImageBuildModal>
    @Ref readonly registryPushModalRef!: InstanceType<typeof RegistryPushModal>

    // ─── 数据属性 ───
    images: DockerImageInfo[] = []
    registries: DockerRegistryInfo[] = []
    loading = false
    showAllImages = false
    formatFileSize = formatFileSize
    formatTime = formatTime

    // ─── 方法 ───
    async loadImages() {
        this.loading = true
        try {
            const res = await api.dockerImageList(this.showAllImages)
            this.images = res.payload || []
        } catch {
            this.actions.showNotification('error', '加载镜像列表失败')
        }
        this.loading = false
    }

    async loadRegistries() {
        try {
            const res = await api.dockerRegistryList()
            this.registries = res.payload || []
        } catch {}
    }

    openPush(image: DockerImageInfo) {
        const tag = image.repoTags.find(t => t !== '<none>:<none>') || ''
        this.registryPushModalRef?.show(this.registries, null, tag)
    }

    // 提取镜像名称（去掉 registry host 和 tag）
    // 例: "registry.com/nginx:latest" → "nginx", "localhost:5000/nginx:v1.0" → "nginx"
    extractImageName(ref: string): string {
        if (!ref || ref === '<none>:<none>') return ''
        // 去掉 digest 或 tag（: 后是 tag，@ 后是 digest；端口号的 : 通过检查 / 来区分）
        let repo = ref
        const atIdx = ref.indexOf('@')
        if (atIdx > 0) {
            repo = ref.substring(0, atIdx)
        } else {
            const colonIdx = ref.lastIndexOf(':')
            if (colonIdx > 0 && !ref.substring(colonIdx).includes('/')) {
                repo = ref.substring(0, colonIdx)
            }
        }
        // 去掉 registry host（第一段含 . 或 : 即为 host）
        const slashIdx = repo.indexOf('/')
        if (slashIdx === -1) return repo
        const first = repo.substring(0, slashIdx)
        return (first.includes('.') || first.includes(':')) ? repo.substring(slashIdx + 1) : repo
    }

    // 获取镜像显示名称：优先用 repoTags，无标签时从 repoDigests 提取
    getImageName(img: DockerImageInfo): string {
        if (img.repoTags && img.repoTags.length > 0) {
            const tag = img.repoTags.find(t => t && t !== '<none>:<none>')
            if (tag) return this.extractImageName(tag)
        }
        if (img.repoDigests && img.repoDigests.length > 0) {
            const digest = img.repoDigests[0]
            const atIndex = digest.indexOf('@')
            if (atIndex > 0) {
                return this.extractImageName(digest.substring(0, atIndex))
            }
        }
        return '<none>'
    }

    async pullImage(image: DockerImageInfo) {
        const tag = image.repoTags.find(t => t && t !== '<none>:<none>')
        if (!tag) {
            this.actions.showNotification('error', '镜像无标签，无法拉取')
            return
        }

        // 解析 registry host（用于匹配私有仓库携带认证信息）
        let tagHost = ''
        let imageName = tag
        const firstSlash = tag.indexOf('/')
        if (firstSlash > 0) {
            const firstPart = tag.substring(0, firstSlash)
            if (firstPart.includes('.') || firstPart.includes(':')) {
                tagHost = firstPart
                imageName = tag.substring(firstSlash + 1)
            }
        }
        const matchedRegistry = this.registries.find(r =>
            r.url === tagHost || r.url.replace(/^https?:\/\//, '') === tagHost
        )

        // 匹配到仓库：传 imageName 让后端拼接；否则传完整 tag
        const pullImageRef = matchedRegistry ? imageName : tag
        const registryUrl = matchedRegistry?.url || ''

        this.actions.showConfirm({
            title: '拉取镜像',
            message: `确定要重新拉取镜像 <strong class="text-slate-900">${tag || image.shortId}</strong> 吗？`,
            icon: 'fa-download',
            iconColor: 'blue',
            confirmText: '确认拉取',
            danger: false,
            onConfirm: async () => {
                try {
                    await api.dockerImagePull(pullImageRef, registryUrl, '')
                    this.actions.showNotification('success', '镜像拉取成功')
                    this.loadImages()
                } catch (e: unknown) {
                    this.actions.showNotification('error', (e instanceof Error ? e.message : '') || '镜像拉取失败')
                }
            }
        })
    }

    handleImageAction(image: DockerImageInfo, action: string) {
        const tag = image.repoTags.find(t => t && t !== '<none>:<none>')
        this.actions.showConfirm({
            title: '删除镜像',
            message: `确定要删除镜像 <strong class="text-slate-900">${tag || image.shortId}</strong> 吗？`,
            icon: 'fa-trash',
            iconColor: 'red',
            confirmText: '确认删除',
            danger: true,
            onConfirm: async () => {
                await api.dockerImageAction(image.id, action)
                this.actions.showNotification('success', '镜像删除成功')
                this.loadImages()
            }
        })
    }

    // ─── 生命周期 ───
    mounted() {
        this.loadImages()
        this.loadRegistries()
    }
}

export default toNative(Images)
</script>

<template>
  <div>
    <!-- Toolbar Bar -->
    <div class="card mb-4">
      <div class="bg-slate-50 border-b border-slate-200 rounded-t-2xl px-4 md:px-6 py-3">
        <!-- 桌面端 -->
        <div class="hidden md:flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-9 h-9 rounded-lg bg-blue-500 flex items-center justify-center">
              <i class="fas fa-compact-disc text-white"></i>
            </div>
            <div>
              <h1 class="text-lg font-semibold text-slate-800">镜像管理</h1>
              <p class="text-xs text-slate-500">管理 Docker 镜像</p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <div class="flex gap-1 bg-slate-100 p-1 rounded-lg">
              <button :class="['px-3 py-1 text-xs font-medium rounded-md transition-all duration-200 flex items-center gap-1.5', !showAllImages ? 'bg-white text-blue-600 shadow-sm' : 'text-slate-500 hover:text-slate-700']" @click="showAllImages = false; loadImages()">
                <i class="fas fa-cube"></i><span>顶层</span>
              </button>
              <button :class="['px-3 py-1 text-xs font-medium rounded-md transition-all duration-200 flex items-center gap-1.5', showAllImages ? 'bg-white text-blue-600 shadow-sm' : 'text-slate-500 hover:text-slate-700']" @click="showAllImages = true; loadImages()">
                <i class="fas fa-layer-group"></i><span>全部</span>
              </button>
            </div>
            <button class="px-3 py-1.5 rounded-lg bg-white border border-slate-200 hover:bg-slate-50 text-slate-700 text-xs font-medium flex items-center gap-1.5 transition-colors" @click="loadImages()">
              <i class="fas fa-rotate"></i>刷新
            </button>
            <button v-if="actions.hasPerm('POST /api/docker/image/:id/action')" class="px-3 py-1.5 rounded-lg bg-blue-500 hover:bg-blue-600 text-white text-xs font-medium flex items-center gap-1.5 transition-colors" @click="buildModalRef?.show()">
              <i class="fas fa-hammer"></i>构建
            </button>
            <button v-if="actions.hasPerm('POST /api/docker/image/pull')" class="px-3 py-1.5 rounded-lg bg-blue-500 hover:bg-blue-600 text-white text-xs font-medium flex items-center gap-1.5 transition-colors" @click="pullModalRef?.show()">
              <i class="fas fa-download"></i>拉取
            </button>
          </div>
        </div>
        <!-- 移动端 -->
        <div class="flex md:hidden items-center justify-between">
          <div class="flex items-center gap-3 min-w-0 flex-1">
            <div class="w-9 h-9 rounded-lg bg-blue-500 flex items-center justify-center flex-shrink-0">
              <i class="fas fa-compact-disc text-white"></i>
            </div>
            <div class="min-w-0">
              <h1 class="text-lg font-semibold text-slate-800 truncate">镜像管理</h1>
              <p class="text-xs text-slate-500 truncate">管理 Docker 镜像</p>
            </div>
          </div>
          <div class="flex items-center gap-1 flex-shrink-0">
            <div class="flex gap-1 bg-slate-100 p-1 rounded-lg mr-1">
              <button :class="['px-2 py-1 text-xs font-medium rounded-md transition-all duration-200 flex items-center gap-1', !showAllImages ? 'bg-white text-blue-600 shadow-sm' : 'text-slate-500 hover:text-slate-700']" @click="showAllImages = false; loadImages()">
                <i class="fas fa-cube"></i>
              </button>
              <button :class="['px-2 py-1 text-xs font-medium rounded-md transition-all duration-200 flex items-center gap-1', showAllImages ? 'bg-white text-blue-600 shadow-sm' : 'text-slate-500 hover:text-slate-700']" @click="showAllImages = true; loadImages()">
                <i class="fas fa-layer-group"></i>
              </button>
            </div>
            <button class="w-9 h-9 rounded-lg bg-white border border-slate-200 hover:bg-slate-50 flex items-center justify-center text-slate-600 transition-colors" title="刷新" @click="loadImages()">
              <i class="fas fa-rotate text-sm"></i>
            </button>
            <button v-if="actions.hasPerm('POST /api/docker/image/:id/action')" class="w-9 h-9 rounded-lg bg-blue-500 hover:bg-blue-600 flex items-center justify-center text-white transition-colors" title="构建" @click="buildModalRef?.show()">
              <i class="fas fa-hammer text-sm"></i>
            </button>
            <button v-if="actions.hasPerm('POST /api/docker/image/pull')" class="w-9 h-9 rounded-lg bg-blue-500 hover:bg-blue-600 flex items-center justify-center text-white transition-colors" title="拉取" @click="pullModalRef?.show()">
              <i class="fas fa-download text-sm"></i>
            </button>
          </div>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="loading" class="flex flex-col items-center justify-center py-20">
        <div class="w-12 h-12 spinner mb-3"></div>
        <p class="text-slate-500">加载中...</p>
      </div>

      <!-- Image List -->
      <div v-else-if="images.length > 0" class="space-y-3">
        <!-- 桌面端表格视图 -->
        <div class="hidden md:block overflow-x-auto">
          <table class="w-full border-collapse">
            <thead>
              <tr class="bg-slate-50 border-b border-slate-200">
                <th class="px-4 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">镜像</th>
                <th class="px-4 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">标签</th>
                <th class="w-32 px-4 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">大小</th>
                <th class="w-36 px-4 py-3 text-left text-xs font-semibold text-slate-600 uppercase tracking-wider">创建时间</th>
                <th class="w-40 px-4 py-3 text-right text-xs font-semibold text-slate-600 uppercase tracking-wider">操作</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-slate-100">
              <tr v-for="img in images" :key="img.id" class="hover:bg-slate-50 transition-colors">
                <td class="px-4 py-3 max-w-[280px]">
                  <div class="flex items-center gap-2 min-w-0">
                    <div class="w-8 h-8 rounded-lg bg-blue-400 flex items-center justify-center flex-shrink-0">
                      <i class="fas fa-compact-disc text-white text-sm"></i>
                    </div>
                    <div class="min-w-0">
                      <span class="font-medium text-slate-800 truncate block">{{ getImageName(img) }}</span>
                      <code class="text-xs text-slate-400 font-mono truncate block mt-0.5">{{ img.shortId }}</code>
                    </div>
                  </div>
                </td>
                <td class="px-4 py-3">
                  <div v-if="img.repoTags && img.repoTags.length > 0" class="flex flex-wrap gap-1">
                    <span v-for="(tag, idx) in img.repoTags" :key="idx" class="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-mono bg-blue-50 text-blue-600">{{ tag }}</span>
                  </div>
                  <span v-else class="text-sm text-slate-400">-</span>
                </td>
                <td class="px-4 py-3 text-sm text-slate-600">{{ formatFileSize(img.size) }}</td>
                <td class="px-4 py-3 whitespace-nowrap text-sm text-slate-600">{{ formatTime(new Date(img.created * 1000).toISOString()) }}</td>
                <td class="px-4 py-3">
                  <div class="flex justify-end items-center gap-1">
                    <button v-if="actions.hasPerm('GET /api/docker/image/:id')" class="btn-icon text-slate-600 hover:bg-slate-50" title="查看详情" @click="$router.push('/docker/image/' + img.id)">
                      <i class="fas fa-circle-info text-xs"></i>
                    </button>
                    <button v-if="actions.hasPerm('POST /api/docker/image/:id/action')" class="btn-icon text-blue-600 hover:bg-blue-50" title="打标签" @click="tagModalRef?.show(img)">
                      <i class="fas fa-tag text-xs"></i>
                    </button>
                    <button v-if="actions.hasPerm('POST /api/docker/image/pull')" class="btn-icon text-blue-600 hover:bg-blue-50" title="拉取（更新）" @click="pullImage(img)">
                      <i class="fas fa-download text-xs"></i>
                    </button>
                    <button v-if="actions.hasPerm('POST /api/docker/image/:id/action')" :disabled="registries.length === 0" class="btn-icon text-indigo-600 hover:bg-indigo-50 disabled:opacity-40 disabled:cursor-not-allowed" :title="registries.length === 0 ? '暂无可用私有仓库' : '推送到仓库'" @click="openPush(img)">
                      <i class="fas fa-upload text-xs"></i>
                    </button>
                    <button v-if="actions.hasPerm('POST /api/docker/image/:id/action')" class="btn-icon text-red-600 hover:bg-red-50" title="删除" @click="handleImageAction(img, 'remove')">
                      <i class="fas fa-trash text-xs"></i>
                    </button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- 移动端卡片视图 -->
        <div class="md:hidden space-y-3 p-4">
          <div 
            v-for="img in images" 
            :key="img.id"
            class="rounded-xl border border-slate-200 bg-white p-4 transition-all hover:shadow-sm"
          >
            <!-- 顶部：镜像信息和图标 -->
            <div class="flex items-center gap-3 min-w-0 flex-1 mb-3">
              <div class="w-10 h-10 rounded-lg bg-blue-400 flex items-center justify-center flex-shrink-0">
                <i class="fas fa-compact-disc text-white text-base"></i>
              </div>
              <div class="min-w-0">
                <span class="font-medium text-slate-800 text-sm truncate block">{{ getImageName(img) }}</span>
                <code class="text-xs text-slate-400 font-mono truncate block mt-0.5">{{ img.shortId }}</code>
              </div>
            </div>

            <!-- 标签 -->
            <div v-if="img.repoTags && img.repoTags.length > 0" class="flex items-start gap-2 mb-3">
              <span class="text-xs text-slate-400 flex-shrink-0 mt-0.5">标签</span>
              <div class="flex flex-wrap gap-1">
                <span v-for="(tag, idx) in img.repoTags" :key="idx" class="inline-flex items-center px-1.5 py-0.5 rounded text-xs font-mono bg-blue-50 text-blue-600">{{ tag }}</span>
              </div>
            </div>

            <!-- 创建时间 -->
            <div class="flex items-center gap-2 mb-3">
              <span class="text-xs text-slate-400 flex-shrink-0">创建</span>
              <span class="text-xs text-slate-500">{{ formatTime(new Date(img.created * 1000).toISOString()) }}</span>
            </div>
            <!-- 大小 -->
            <div class="flex items-center gap-2 mb-3">
              <span class="text-xs text-slate-400 flex-shrink-0">大小</span>
              <span class="text-xs text-slate-500">{{ formatFileSize(img.size) }}</span>
            </div>
          
            <!-- 底部：操作按钮 -->
            <div class="flex flex-wrap gap-1.5 pt-2 border-t border-slate-100">
              <button v-if="actions.hasPerm('GET /api/docker/image/:id')" class="btn-icon text-slate-600 hover:bg-slate-50" title="查看详情" @click="$router.push('/docker/image/' + img.id)">
                <i class="fas fa-circle-info text-xs"></i><span class="text-xs ml-1">详情</span>
              </button>
              <button v-if="actions.hasPerm('POST /api/docker/image/:id/action')" class="btn-icon text-blue-600 hover:bg-blue-50" title="打标签" @click="tagModalRef?.show(img)">
                <i class="fas fa-tag text-xs"></i><span class="text-xs ml-1">标签</span>
              </button>
              <button v-if="actions.hasPerm('POST /api/docker/image/pull')" class="btn-icon text-blue-600 hover:bg-blue-50" title="拉取（更新）" @click="pullImage(img)">
                <i class="fas fa-download text-xs"></i><span class="text-xs ml-1">拉取</span>
              </button>
              <button v-if="actions.hasPerm('POST /api/docker/image/:id/action')" :disabled="registries.length === 0" class="btn-icon text-indigo-600 hover:bg-indigo-50 disabled:opacity-40 disabled:cursor-not-allowed" :title="registries.length === 0 ? '暂无可用私有仓库' : '推送到仓库'" @click="openPush(img)">
                <i class="fas fa-upload text-xs"></i><span class="text-xs ml-1">推送</span>
              </button>
              <button v-if="actions.hasPerm('POST /api/docker/image/:id/action')" class="btn-icon text-red-600 hover:bg-red-50" title="删除" @click="handleImageAction(img, 'remove')">
                <i class="fas fa-trash text-xs"></i><span class="text-xs ml-1">删除</span>
              </button>
            </div>
          </div>
        </div>
      </div>

      <div v-else class="flex flex-col items-center justify-center py-20">
        <div class="w-16 h-16 rounded-lg bg-slate-100 flex items-center justify-center mb-4">
          <i class="fas fa-compact-disc text-4xl text-slate-300"></i>
        </div>
        <p class="text-slate-600 font-medium mb-1">暂无镜像</p>
        <p class="text-sm text-slate-400">点击「拉取镜像」从 Registry 获取镜像</p>
      </div>
    </div>

    <ImagePullModal ref="pullModalRef" @success="loadImages" />
    <ImageTagModal ref="tagModalRef" @success="loadImages" />
    <ImageBuildModal ref="buildModalRef" @success="loadImages" />
    <RegistryPushModal ref="registryPushModalRef" />
  </div>
</template>
