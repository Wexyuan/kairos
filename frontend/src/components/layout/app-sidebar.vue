<script setup lang="ts">
import type { MenuOption } from 'naive-ui'
import { h } from 'vue'
import { useRoute, useRouter } from 'vue-router'

const route = useRoute()
const router = useRouter()
const collapsed = ref(false)

function buildMenuOptions(): MenuOption[] {
  return router
    .getRoutes()
    .filter(record => Boolean(record.meta.title))
    .map(record => ({
      key: record.name as string,
      label: record.meta.title,
      icon: () => h('div', { class: `h-5 w-5 ${record.meta.icon}` }),
    }))
}

const menuOptions = buildMenuOptions()

const activeKey = computed<string | null>(() =>
  typeof route.name === 'string' ? route.name : null,
)

function handleMenuSelect(key: string): void {
  router.push({ name: key })
}
</script>

<template>
  <n-layout-sider
    v-model:collapsed="collapsed"
    collapse-mode="width"
    :collapsed-width="64"
    :width="220"
    show-trigger
  >
    <div class="flex flex-col h-full">
      <RouterLink to="/" class="p-4 flex shrink-0 gap-3 cursor-pointer items-center overflow-hidden">
        <img src="/logo.svg" alt="kairos" class="shrink-0 h-8 w-8">
        <img src="/title.svg" alt="" class="shrink-0 h-6 transition-opacity duration-300" :class="collapsed ? 'opacity-0' : 'opacity-100'">
      </RouterLink>
      <n-menu
        :collapsed="collapsed"
        :collapsed-width="64"
        :indent="20"
        :options="menuOptions"
        :value="activeKey"
        class="flex-1"
        @update:value="handleMenuSelect"
      />
    </div>
  </n-layout-sider>
</template>

<style scoped>
:deep(.n-menu .n-menu-item-content::before) {
  left: 14px;
  right: 14px;
}
</style>
