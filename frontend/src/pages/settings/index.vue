<script setup lang="ts">
import { useTheme } from '@/composables'

const { isDark, toggleDark } = useTheme()

const themeOptions = [
  { value: 'dark', label: '深色', icon: 'i-mdi-weather-night' },
  { value: 'light', label: '浅色', icon: 'i-mdi-weather-sunny' },
]

const themeValue = computed({
  get: () => (isDark.value ? 'dark' : 'light'),
  set: (value: string) => {
    if ((value === 'dark') !== isDark.value)
      toggleDark()
  },
})
</script>

<template>
  <div class="mx-auto pa-4 max-w-3xl">
    <h3 class="title-sm text-ink/60 p-2">
      外观
    </h3>

    <n-card size="small">
      <template #header>
        <div class="title-sm">
          主题
        </div>
        <div class="title-sm text-ink/60">
          切换应用为深色或浅色主题，设置后立即生效
        </div>
      </template>

      <template #header-extra>
        <n-radio-group v-model:value="themeValue" size="small">
          <n-radio-button v-for="option in themeOptions" :key="option.value" :value="option.value">
            <span class="flex-center gap-2">
              <span class="h-4 w-4" :class="option.icon" />
              <span>{{ option.label }}</span>
            </span>
          </n-radio-button>
        </n-radio-group>
      </template>
    </n-card>
  </div>
</template>
