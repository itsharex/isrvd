<script lang="ts">
import { Component, Vue, toNative } from 'vue-facing-decorator'

import Dropdown from '@/component/dropdown.vue'

import { usePortal } from '@/stores'

@Component({
    components: { Dropdown }
})
class UserMenu extends Vue {
    portal = usePortal()

    // ─── 数据属性 ───
    menuOpen = false

    // ─── 方法 ───
    toggleTheme() {
        const isDark = document.documentElement.classList.toggle('dark')
        localStorage.setItem('app-theme', isDark ? 'dark' : 'light')
    }

    handleLogout() {
        this.portal.clearAuth()
    }
}

export default toNative(UserMenu)
</script>

<template>
  <!-- header 认证模式：仅显示用户名，无注销入口 -->
  <div
    v-if="portal.authMode === 'header'"
    class="px-2 py-2 text-sm font-medium text-slate-500 flex items-center gap-2 cursor-default select-none"
    :title="portal.username || '未登录'"
  >
    <i class="fas fa-user-tie"></i>
    <span class="hidden sm:inline">{{ portal.username }}</span>
  </div>

  <!-- jwt 认证模式：用户名 + 下拉菜单 -->
  <Dropdown v-else v-model:open="menuOpen" placement="bottom" align="right" :close-on-click="true" max-height="320px">
    <template #trigger="{ toggle }">
      <button
        class="px-2 py-2 text-sm font-medium text-slate-500 flex items-center gap-2 rounded-lg hover:bg-slate-100 transition-colors"
        :title="portal.username || '未登录'"
        @click="toggle"
      >
        <i class="fas fa-user-tie"></i>
        <span class="hidden sm:inline">{{ portal.username }}</span>
        <i class="fas fa-chevron-down text-xs text-slate-400 hidden sm:inline transition-transform duration-200" :class="{ 'rotate-180': menuOpen }"></i>
      </button>
    </template>

    <!-- 个人设置 -->
    <router-link
      to="/account/profile"
      class="w-full flex items-center gap-3 px-4 py-3 text-sm font-medium text-slate-600 hover:text-blue-600 hover:bg-blue-50 transition-colors"
      @click="menuOpen = false"
    >
      <i class="fas fa-user-circle"></i>
      个人设置
    </router-link>

    <!-- 暗黑模式切换 -->
    <button
      class="w-full flex items-center gap-3 px-4 py-3 text-sm font-medium text-slate-600 hover:text-blue-600 hover:bg-blue-50 transition-colors"
      @click="toggleTheme"
    >
      <i class="fas fa-moon"></i>
      <span>切换主题</span>
    </button>

    <!-- 分割线 -->
    <div class="border-t border-slate-100 my-1"></div>

    <!-- 注销选项 -->
    <button
      class="w-full flex items-center gap-3 px-4 py-3 text-sm font-medium text-slate-600 hover:text-red-600 hover:bg-red-50 transition-colors"
      @click="handleLogout"
    >
      <i class="fas fa-sign-out-alt"></i>
      退出
    </button>
  </Dropdown>
</template>
