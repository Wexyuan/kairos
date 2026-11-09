<script setup lang="ts">
import { useTheme } from '@/composables'
import { NOTE_MAX_WIDTH_RANGE, useAppStore } from '@/stores/app'

const MIB = 1024 * 1024

const { isDark, toggleDark } = useTheme()
const appStore = useAppStore()

const appName = computed(() => appStore.settings?.name ?? 'kairos')
const appVersion = computed(() => appStore.settings?.version ?? '-')
const clipboardShortcut = computed(() => appStore.settings?.clipboard_shortcut ?? '-')
const clipboardMaxEntries = computed(() => appStore.settings?.clipboard_max_entries ?? '-')
const clipboardMaxPinnedEntries = computed(() => appStore.settings?.clipboard_max_pinned_entries ?? '-')

const noteMaxImageSize = computed(() => {
  const bytes = appStore.settings?.note_max_image_size
  if (!bytes)
    return '-'
  const mib = bytes / MIB
  return `${Number.isInteger(mib) ? mib : mib.toFixed(1)} MiB`
})
</script>

<template>
  <div class="mx-auto p-6 gap-6 grid grid-cols-2 max-w-240 items-start">
    <div class="flex flex-col gap-6">
      <section class="panel">
        <div class="px-4 py-2.5 border-b border-ink/5 bg-brand/15 flex gap-2 items-center">
          <span class="i-mdi-palette-outline text-brand h-4 w-4" />
          <h2 class="title-md">
            外观
          </h2>
        </div>

        <div class="px-4">
          <div class="py-3 flex gap-4 items-center justify-between">
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
          <span class="i-mdi-clipboard-text-outline text-brand h-4 w-4" />
          <h2 class="title-md">
            剪贴板
          </h2>
        </div>

        <div class="px-4">
          <div class="py-3 border-b border-ink/5 flex gap-4 items-center justify-between">
            <div class="min-w-0">
              <h3 class="title-sm">
                全局快捷键
              </h3>
              <p class="text-ink/60 mt-1">
                唤起或隐藏剪贴板窗口
              </p>
            </div>
            <span class="title-sm shrink-0">{{ clipboardShortcut }}</span>
          </div>

          <div class="py-3 border-b border-ink/5 flex gap-4 items-center justify-between">
            <div class="min-w-0">
              <h3 class="title-sm">
                历史条数上限
              </h3>
              <p class="text-ink/60 mt-1">
                超出上限的记录会在下次复制时清理
              </p>
            </div>
            <span class="title-sm shrink-0">{{ clipboardMaxEntries }}</span>
          </div>

          <div class="py-3 flex gap-4 items-center justify-between">
            <div class="min-w-0">
              <h3 class="title-sm">
                置顶条数上限
              </h3>
              <p class="text-ink/60 mt-1">
                达到上限后无法继续置顶
              </p>
            </div>
            <span class="title-sm shrink-0">{{ clipboardMaxPinnedEntries }}</span>
          </div>
        </div>
      </section>
    </div>

    <div class="flex flex-col gap-6">
      <section class="panel">
        <div class="px-4 py-2.5 border-b border-ink/5 bg-brand/15 flex gap-2 items-center">
          <span class="i-mdi-file-image-outline text-brand h-4 w-4" />
          <h2 class="title-md">
            笔记
          </h2>
        </div>

        <div class="px-4">
          <div class="py-3 border-b border-ink/5 flex gap-4 items-center justify-between">
            <div class="min-w-0">
              <h3 class="title-sm">
                最大宽度
              </h3>
              <p class="text-ink/60 mt-1">
                笔记工具栏与正文的最大内容宽度
              </p>
            </div>
            <div class="flex shrink-0 gap-2 items-center">
              <n-input-number
                size="small"
                class="w-24"
                :value="appStore.noteMaxWidth"
                :min="NOTE_MAX_WIDTH_RANGE.min"
                :max="NOTE_MAX_WIDTH_RANGE.max"
                :step="NOTE_MAX_WIDTH_RANGE.step"
                @update:value="appStore.setNoteMaxWidth"
              />
              <span class="text-xs text-ink/60">px</span>
            </div>
          </div>

          <div class="py-3 flex gap-4 items-center justify-between">
            <div class="min-w-0">
              <h3 class="title-sm">
                图片大小上限
              </h3>
              <p class="text-ink/60 mt-1">
                粘贴或上传的单张图片最大体积
              </p>
            </div>
            <span class="title-sm shrink-0">{{ noteMaxImageSize }}</span>
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
          <div class="py-3 border-b border-ink/5 flex gap-4 items-center justify-between">
            <span class="text-ink/60">产品</span>
            <span class="title-sm truncate">{{ appName }}</span>
          </div>
          <div class="py-3 flex gap-4 items-center justify-between">
            <span class="text-ink/60">版本</span>
            <span class="title-sm">{{ appVersion }}</span>
          </div>
        </div>
      </section>
    </div>
  </div>
</template>
