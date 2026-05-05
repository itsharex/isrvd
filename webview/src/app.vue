<script lang="ts">
import { Component, Provide, Ref, Vue, toNative } from 'vue-facing-decorator'

import { APP_ACTIONS_KEY, APP_STATE_KEY, initProvider } from '@/store/state'

import { setRouterGuard } from '@/router'

import ConfirmModal from '@/component/confirm.vue'
import NavigationBar from '@/component/navigation.vue'
import NotificationManager from '@/component/notification.vue'
import PageAgent from '@/component/page-agent.vue'
import UserMenu from '@/component/user-menu.vue'

import AuthLogin from '@/views/account/login.vue'

const { state, actions } = initProvider()
setRouterGuard(actions.hasPerm, () => state.permissionsLoaded, actions.isAuthenticated)

@Component({
    components: { ConfirmModal, NavigationBar, NotificationManager, PageAgent, UserMenu, AuthLogin }
})
class App extends Vue {
    @Provide(APP_STATE_KEY) state = state
    @Provide(APP_ACTIONS_KEY) actions = actions
    sidebarCollapsed = false

    @Ref readonly navigationRef!: InstanceType<typeof NavigationBar>

    toggleMobileMenu() {
        this.navigationRef?.toggleMobileSidebar()
    }

    async mounted() {
        await this.actions.initialize()
    }
}

export default toNative(App)
</script>

<template>
  <div class="min-h-screen bg-slate-50">
    <!-- 初始化加载状态 -->
    <div v-if="!state.initialized" class="flex items-center justify-center min-h-screen">
      <div class="flex flex-col items-center gap-4">
        <div class="w-12 h-12 border-4 border-slate-200 border-t-blue-500 rounded-full animate-spin"></div>
        <span class="text-slate-500 text-sm">正在初始化...</span>
      </div>
    </div>

    <!-- 主内容 -->
    <template v-else-if="state.username">
      <!-- 移动端顶部菜单栏 -->
      <header
        class="fixed top-0 left-0 right-0 h-16 bg-white/80 backdrop-blur-xl border-b border-slate-200/50 z-40 flex items-center justify-between px-4 transition-all duration-300"
        :class="sidebarCollapsed ? 'lg:left-16' : 'lg:left-64'"
      >
        <!-- 移动端菜单切换按钮 -->
        <button
          class="lg:hidden p-2 rounded-lg hover:bg-slate-100 transition-colors"
          @click="toggleMobileMenu"
        >
          <i class="fas fa-bars text-slate-600"></i>
        </button>

        <!-- 工具栏按钮区域 -->
        <div class="flex items-center gap-2 overflow-x-auto ml-auto">
          <a
            v-for="link in state.toolbarLinks"
            :key="link.url"
            :href="link.url"
            target="_blank"
            rel="noopener noreferrer"
            class="flex items-center gap-2 px-2 py-2 rounded-lg text-sm font-medium text-slate-500 hover:bg-slate-100 transition-colors whitespace-nowrap flex-shrink-0"
          >
            <i v-if="link.icon" :class="link.icon.includes(' ') ? link.icon : `fas ${link.icon}`"></i>
            <span>{{ link.label }}</span>
          </a>
        </div>

        <!-- 用户信息 -->
        <div class="flex items-center gap-1">
          <PageAgent v-if="actions.hasPerm('agent')" />
          <div v-if="actions.hasPerm('agent')" class="hidden sm:block w-px h-5 bg-slate-200 mx-1"></div>
          <UserMenu />
        </div>
      </header>

      <NavigationBar ref="navigationRef" v-model:collapsed="sidebarCollapsed" />
      <main class="px-4 py-6 pt-20 transition-all duration-300" :class="sidebarCollapsed ? 'lg:ml-16' : 'lg:ml-64'">
        <router-view />
      </main>
    </template>

    <!-- 登录页面 -->
    <AuthLogin v-else />

    <NotificationManager />
    <ConfirmModal />
  </div>
</template>
