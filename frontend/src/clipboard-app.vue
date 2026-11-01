<script setup lang="ts">
import type ClipboardPanelComponent from '@/components/clipboard/clipboard-panel.vue'
import { Events } from '@wailsio/runtime'
import { WindowService } from '@/services/bindings'
import { useClipboardStore } from '@/stores/clipboard'

const EVENT_WINDOW_SHOW = 'windows:WindowShow'

const store = useClipboardStore()
const panelRef = ref<InstanceType<typeof ClipboardPanelComponent> | null>(null)
let stopShowListener: (() => void) | undefined

onMounted(() => {
  stopShowListener = Events.On(EVENT_WINDOW_SHOW, refresh)
})

onBeforeUnmount(() => stopShowListener?.())

function refresh(): void {
  void panelRef.value?.refresh()
}

function hideWindow(): void {
  WindowService.HideClipboardWindow().catch(() => {})
}
</script>

<template>
  <div class="bg-surface flex flex-col h-screen">
    <header class="px-3 border-b border-ink/5 flex shrink-0 gap-2 h-12 items-center">
      <span class="i-mdi-clipboard-text-outline text-brand shrink-0 h-4 w-4" />
      <div class="flex flex-1 gap-2 min-w-0 items-center self-stretch" style="--wails-draggable: drag">
        <span class="title-md truncate">剪贴板</span>
        <span v-if="store.total" class="text-xs text-ink/45 shrink-0">{{ store.total }} 条</span>
      </div>
      <n-button quaternary size="tiny" class="icon-btn shrink-0" title="关闭（Esc）" @click="hideWindow">
        <template #icon>
          <span class="i-mdi-close h-4 w-4" />
        </template>
      </n-button>
    </header>

    <ClipboardPanel ref="panelRef" mode="window" @copied="hideWindow" @escape="hideWindow" />
  </div>
</template>
