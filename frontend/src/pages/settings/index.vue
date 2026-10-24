<script setup lang="ts">
import { useTheme } from '@/composables'

const { isDark, toggleDark } = useTheme()
</script>

<template>
  <div class="mx-auto p-6 gap-6 grid max-w-240 items-start md:grid-cols-2">
    <section class="panel">
      <div class="px-4 py-2.5 border-b border-ink/5 bg-brand/15 flex gap-2 items-center">
        <span class="i-mdi-palette-outline text-brand h-4 w-4" />
        <h2 class="title-md">
          外观
        </h2>
      </div>

      <div class="px-4">
        <div class="py-2.5 flex gap-4 items-center justify-between">
          <div class="min-w-0">
            <h3 class="title-sm">
              主题
            </h3>
            <p class="text-ink/60 mt-1">
              切换亮色 / 暗色主题界面
            </p>
          </div>
          <n-switch size="small" class="shrink-0" :value="isDark" @update:value="toggleDark" />
        </div>
      </div>
    </section>

    <section class="panel">
      <div class="px-4 py-2.5 border-b border-ink/5 bg-brand/15 flex gap-2 items-center">
        <span class="i-mdi-information-outline text-brand h-4 w-4" />
        <h2 class="title-md">
          关于
        </h2>
      </div>

      <div class="px-4">
        <div class="py-2.5 border-b border-ink/5 flex gap-4 items-center justify-between">
          <span class="text-ink/60">产品</span>
          <span class="title-sm">Kairos · 个人办公工作台</span>
        </div>
        <div class="py-2.5 flex gap-4 items-center justify-between">
          <span class="text-ink/60">版本</span>
          <span class="title-sm">0.1.0</span>
        </div>
      </div>
    </section>
  </div>
</template>
