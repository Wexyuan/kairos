<script setup lang="ts">
import { Events, Window } from '@wailsio/runtime'
import { useTheme } from '@/composables'
import { useAppStore } from '@/stores/app'

interface WindowAction {
  key: 'minimize' | 'maximize' | 'close'
  label: string
  icon: string
  run: () => void
}

const EVENT_MAXIMISE = 'windows:WindowMaximise'
const EVENT_UN_MAXIMISE = 'windows:WindowUnMaximise'

const route = useRoute()
const appStore = useAppStore()
const { isDark, toggleDark } = useTheme()

const maximised = ref(false)
let stopMaximiseListener: (() => void) | undefined
let stopUnMaximiseListener: (() => void) | undefined

const pageTitle = computed(() => route.meta.title ?? '')
const themeLabel = computed(() => (isDark.value ? '切换到浅色模式' : '切换到深色模式'))
const themeIcon = computed(() => (isDark.value ? 'i-mdi-weather-sunny' : 'i-mdi-weather-night'))

const windowActions: WindowAction[] = [
  { key: 'minimize', label: '最小化', icon: 'i-mdi-window-minimize', run: minimiseWindow },
  { key: 'maximize', label: '最大化', icon: 'i-mdi-window-maximize', run: toggleMaximiseWindow },
  { key: 'close', label: '关闭', icon: 'i-mdi-close', run: closeWindow },
]

function minimiseWindow(): void {
  void Window.Minimise().catch(() => {})
}

function toggleMaximiseWindow(): void {
  void Window.ToggleMaximise().catch(() => {})
}

function closeWindow(): void {
  void Window.Close().catch(() => {})
}

async function syncMaximised(): Promise<void> {
  try {
    maximised.value = await Window.IsMaximised()
  }
  catch {
    maximised.value = false
  }
}

function actionLabel(action: WindowAction): string {
  return action.key === 'maximize' && maximised.value ? '还原' : action.label
}

function actionIcon(action: WindowAction): string {
  return action.key === 'maximize' && maximised.value ? 'i-mdi-window-restore' : action.icon
}

function handleToggleTheme(): void {
  toggleDark()
}

onMounted(async () => {
  await syncMaximised()
  stopMaximiseListener = Events.On(EVENT_MAXIMISE, () => {
    maximised.value = true
  })
  stopUnMaximiseListener = Events.On(EVENT_UN_MAXIMISE, () => {
    maximised.value = false
  })
})

onBeforeUnmount(() => {
  stopMaximiseListener?.()
  stopUnMaximiseListener?.()
})
</script>

<template>
  <header class="px-4 border-b border-ink/5 bg-surface flex shrink-0 h-16 items-center">
    <n-button quaternary size="small" class="icon-btn shrink-0" aria-label="切换菜单栏" title="切换菜单栏" @click="appStore.toggleSidebar">
      <template #icon>
        <span class="i-mdi-menu h-4 w-4" />
      </template>
    </n-button>

    <div class="px-2 flex flex-1 min-w-0 items-center self-stretch" style="--wails-non-client-region: caption">
      <span class="text-sm truncate">{{ pageTitle }}</span>
    </div>

    <div class="flex gap-1 items-center">
      <n-button quaternary size="small" class="icon-btn" :aria-label="themeLabel" :title="themeLabel" @click="handleToggleTheme">
        <template #icon>
          <span :class="themeIcon" class="h-4 w-4" />
        </template>
      </n-button>

      <span class="mx-1 bg-ink/15 h-4 w-px" />

      <n-button
        v-for="action in windowActions"
        :key="action.key"
        quaternary
        size="small"
        class="icon-btn"
        :style="{ '--wails-non-client-region': action.key }"
        :aria-label="actionLabel(action)"
        :title="actionLabel(action)"
        @click="action.run"
      >
        <template #icon>
          <span :class="actionIcon(action)" class="h-4 w-4" />
        </template>
      </n-button>
    </div>
  </header>
</template>
