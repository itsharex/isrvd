<script lang="ts">
import { Component, Inject, Vue, toNative } from 'vue-facing-decorator'

import { APP_STATE_KEY } from '@/store/state'
import type { AppState } from '@/store/state'

import Dropdown from '@/component/dropdown.vue'

@Component({
    components: { Dropdown }
})
class ToolbarLinks extends Vue {
    @Inject({ from: APP_STATE_KEY }) readonly state!: AppState

    menuOpen = false
}

export default toNative(ToolbarLinks)
</script>

<template>
  <!-- 桌面端：横向按钮列表 -->
  <div class="hidden md:flex items-center gap-2 overflow-x-auto ml-auto mr-2">
    <a
      v-for="link in state.toolbarLinks"
      :key="link.url"
      :href="link.url"
      target="_blank"
      rel="noopener noreferrer"
      class="btn-ghost px-4 py-2 text-sm font-medium rounded-lg transition-all duration-200 flex items-center gap-2 text-slate-600 hover:text-primary-600 hover:bg-primary-50 whitespace-nowrap"
    >
      <i v-if="link.icon" :class="link.icon.includes(' ') ? link.icon : `fas ${link.icon}`"></i>
      <span class="whitespace-nowrap">{{ link.label }}</span>
    </a>
  </div>

  <!-- 手机端：图标下拉菜单 -->
  <div class="flex md:hidden items-center ml-auto mr-2">
    <Dropdown v-model:open="menuOpen" placement="bottom" align="right" :close-on-click="true" max-height="320px">
      <template #trigger="{ toggle }">
        <button
          class="p-2 rounded-lg hover:bg-slate-100 transition-colors text-slate-600"
          title="快捷链接"
          @click="toggle"
        >
          <i class="fas fa-star"></i>
        </button>
      </template>

      <template v-if="state.toolbarLinks.length === 0">
        <div class="px-4 py-3 text-sm text-slate-400">无快捷链接</div>
      </template>

      <a
        v-for="link in state.toolbarLinks"
        :key="link.url"
        :href="link.url"
        target="_blank"
        rel="noopener noreferrer"
        class="w-full flex items-center gap-3 px-4 py-3 text-sm font-medium text-slate-600 hover:text-blue-600 hover:bg-blue-50 transition-colors"
        @click="menuOpen = false"
      >
        <i v-if="link.icon" :class="link.icon.includes(' ') ? link.icon : `fas ${link.icon}`"></i>
        <span>{{ link.label }}</span>
      </a>
    </Dropdown>
  </div>
</template>
