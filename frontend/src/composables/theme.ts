import type { GlobalTheme } from 'naive-ui'
import { createGlobalState } from '@vueuse/core'
import { darkTheme } from 'naive-ui'
import { WindowService } from '@/services/bindings'
import { buildThemeOverrides } from '@/theme'

// useTheme exposes naive-ui theme objects and dark mode toggle methods.
export const useTheme = createGlobalState(() => {
  const isDark = useDark({ storageKey: 'kairos-color-scheme' })
  const toggleDark = useToggle(isDark)

  const theme = computed<GlobalTheme | null>(() =>
    isDark.value ? darkTheme : null,
  )

  const themeOverrides = computed(() => buildThemeOverrides(isDark.value))

  watch(isDark, (dark) => {
    WindowService.SetTheme({ dark }).catch(() => {})
  }, { immediate: true })

  return { theme, themeOverrides, isDark, toggleDark }
})
