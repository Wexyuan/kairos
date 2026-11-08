<script setup lang="ts">
import type { InputInst } from 'naive-ui'
import { useAction } from '@/composables'
import { PAGE_SIZES, SEARCH_DELAY } from '@/stores/app'
import { useClipboardStore } from '@/stores/clipboard'

const props = withDefaults(defineProps<{
  mode?: 'page' | 'window'
}>(), {
  mode: 'page',
})

const emit = defineEmits<{
  copied: [] // emitted after an entry is copied
  escape: [] // emitted when Escape is pressed
}>()

const store = useClipboardStore()
const { runAction, confirmAction } = useAction()

const isWindow = computed(() => props.mode === 'window')
const searchInputRef = ref<InputInst | null>(null)
const searchValue = ref('')
const keyword = computed(() => searchValue.value.trim())
const pageCount = computed(() => Math.max(1, Math.ceil(store.total / store.pageSize)))

const applySearch = useDebounceFn(async () => {
  store.keyword = keyword.value
  await runAction('搜索剪贴板', () => store.searchEntries())
}, SEARCH_DELAY)

onMounted(() => void refresh())

watch(searchValue, () => void applySearch())

async function refresh(): Promise<void> {
  resetSearch()
  focusSearch()
  await reload()
}

function resetSearch(): void {
  searchValue.value = ''
  store.keyword = ''
}

function focusSearch(): void {
  if (!isWindow.value)
    return
  nextTick(() => {
    searchInputRef.value?.focus()
    searchInputRef.value?.select()
  })
}

async function reload(): Promise<void> {
  await runAction('加载剪贴板', () => store.listEntries())
}

function clearHistory(): void {
  confirmAction({
    title: '清空剪贴板',
    content: '是否清空全部剪贴板记录？',
    action: () => store.clearEntries(),
  })
}

function changePage(page: number): void {
  store.page = page
  void reload()
}

function changePageSize(pageSize: number): void {
  store.pageSize = pageSize
  changePage(1)
}

defineExpose({ refresh })
</script>

<template>
  <div class="flex flex-1 flex-col min-h-0" :class="!isWindow && 'gap-4'">
    <div class="flex shrink-0 gap-2 items-center justify-between" :class="isWindow ? 'px-3 py-2.5' : 'px-3'">
      <n-input
        ref="searchInputRef"
        v-model:value="searchValue"
        size="small"
        clearable
        placeholder="搜索剪贴板…"
        class="flex-1 min-w-0"
        :class="{ 'max-w-90': !isWindow }"
      >
        <template #prefix>
          <span class="i-mdi-magnify text-ink/40 h-4 w-4" />
        </template>
      </n-input>
      <div class="flex shrink-0 gap-2 items-center">
        <n-spin v-if="store.listLoading" :size="14" />
        <n-button quaternary size="small" class="icon-btn" title="刷新" @click="reload">
          <template #icon>
            <span class="i-mdi-refresh h-4 w-4" />
          </template>
        </n-button>
        <n-button quaternary size="small" class="icon-btn" title="清空历史" :disabled="!store.total" @click="clearHistory">
          <template #icon>
            <span class="i-mdi-delete-sweep-outline h-4 w-4" />
          </template>
        </n-button>
      </div>
    </div>

    <ClipboardList :window="isWindow" @copied="emit('copied')" @escape="emit('escape')" />

    <div v-if="!isWindow && store.total" class="px-3 flex shrink-0 gap-2 items-center justify-between">
      <span class="text-xs text-ink/45">第 {{ store.page }} / {{ pageCount }} 页</span>
      <n-pagination
        size="small"
        :page="store.page"
        :page-count="pageCount"
        :page-size="store.pageSize"
        :page-sizes="PAGE_SIZES"
        :page-slot="5"
        show-size-picker
        @update:page="changePage"
        @update:page-size="changePageSize"
      />
    </div>

    <footer v-if="isWindow" class="px-3 py-2.5 flex-center shrink-0 gap-2">
      <n-tag type="primary" size="small" round disabled>
        ↑↓ 选择
      </n-tag>
      <n-tag type="primary" size="small" round disabled>
        Enter 复制
      </n-tag>
      <n-tag type="primary" size="small" round disabled>
        Del 删除
      </n-tag>
      <n-tag type="primary" size="small" round disabled>
        Esc 关闭
      </n-tag>
    </footer>
  </div>
</template>
