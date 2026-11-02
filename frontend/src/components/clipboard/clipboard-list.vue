<script setup lang="ts">
import type { ClipboardEntryItem } from '@/services/bindings'
import { useAction } from '@/composables'
import { formatRelativeTime, useClipboardStore } from '@/stores/clipboard'

interface EntryRow {
  entry: ClipboardEntryItem
  segments: PreviewSegment[]
  charCount: number
  singleLine: boolean
}

interface PreviewSegment {
  text: string
  matched: boolean
}

const props = defineProps<{
  window?: boolean
}>()

const emit = defineEmits<{
  copied: [] // emitted after an entry is copied
  escape: [] // emitted when Escape is pressed
}>()

const EMPTY_PREVIEW = '（空白内容）'
const MATCH_CLASS = 'rounded-sm bg-brand/15 px-1 text-brand'

const store = useClipboardStore()
const { runAction, confirmAction } = useAction()
const message = useMessage()

const listContainerRef = ref<HTMLElement | null>(null)
const selectedIndex = ref(props.window ? 0 : -1)
let copying = false

const rows = computed<EntryRow[]>(() => store.entries.map(toEntryRow))
const selectedRow = computed(() => rows.value[selectedIndex.value])
const emptyText = computed(() => (store.keyword ? '没有匹配的剪贴板记录' : '暂无记录，复制点什么试试叭 (^v^)'))

watch(rows, () => {
  selectedIndex.value = props.window && rows.value.length ? 0 : -1
})

watch(selectedIndex, (index) => {
  if (index < 0)
    return
  nextTick(() => listContainerRef.value?.querySelector(`[data-index="${index}"]`)?.scrollIntoView({ block: 'nearest' }))
})

function toEntryRow(entry: ClipboardEntryItem): EntryRow {
  const content = entry.content.trim()
  const preview = content || EMPTY_PREVIEW
  return {
    entry,
    segments: toPreviewSegments(preview, store.keyword.trim()),
    charCount: content.length,
    singleLine: !content.includes('\n'),
  }
}

function toPreviewSegments(text: string, keyword: string): PreviewSegment[] {
  if (!keyword)
    return [{ text, matched: false }]
  const escaped = keyword.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
  const pattern = new RegExp(`(${escaped})`, 'gi')
  return text.split(pattern).map((part, index) => ({ text: part, matched: index % 2 === 1 }))
}

async function copyEntry(entry: ClipboardEntryItem): Promise<void> {
  if (copying)
    return
  copying = true
  try {
    if (await runAction('复制剪贴板内容', () => store.copyEntry(entry.id))) {
      message.success('已复制到剪贴板')
      emit('copied')
    }
  }
  finally {
    copying = false
  }
}

async function togglePinned(entry: ClipboardEntryItem): Promise<void> {
  const pinned = !entry.pinned
  if (await runAction(pinned ? '置顶' : '取消置顶', () => store.setEntryPinned(entry.id, pinned)))
    message.success(pinned ? '已置顶' : '已取消置顶')
}

function deleteEntry(entry: ClipboardEntryItem): void {
  confirmAction({
    title: '删除',
    content: '是否删除该条剪贴板记录？',
    action: () => store.deleteEntry(entry.id),
  })
}

const KEY_HANDLERS: Record<string, () => void> = {
  ArrowDown: () => moveSelection(1),
  ArrowUp: () => moveSelection(-1),
  Enter: () => copySelectedEntry(),
  Delete: () => deleteSelectedEntry(),
  Escape: () => emit('escape'),
}

const KEYS_ALLOWED_IN_INPUT = new Set(['Enter', 'Escape'])

if (props.window)
  useEventListener(window, 'keydown', handleKeydown)

function handleKeydown(event: KeyboardEvent): void {
  if (event.isComposing || (event.target instanceof HTMLElement && event.target.closest('.n-modal-container')))
    return
  if (!KEYS_ALLOWED_IN_INPUT.has(event.key) && event.target instanceof HTMLElement && event.target.closest('input, textarea, [contenteditable="true"]'))
    return
  const handler = KEY_HANDLERS[event.key]
  if (!handler)
    return
  event.preventDefault()
  handler()
}

function moveSelection(step: number): void {
  const last = rows.value.length - 1
  if (last < 0)
    return
  const next = selectedIndex.value < 0 ? (step > 0 ? 0 : last) : selectedIndex.value + step
  selectedIndex.value = Math.min(last, Math.max(0, next))
}

function copySelectedEntry(): void {
  const row = selectedRow.value ?? rows.value[0]
  if (row)
    void copyEntry(row.entry)
}

function deleteSelectedEntry(): void {
  if (selectedRow.value)
    deleteEntry(selectedRow.value.entry)
}
</script>

<template>
  <div ref="listContainerRef" class="flex flex-1 flex-col min-h-0">
    <div v-if="!rows.length" class="flex-center flex-1">
      <n-spin v-if="store.listLoading" :size="20" />
      <n-empty v-else :description="emptyText" />
    </div>

    <n-scrollbar v-else class="flex-1 min-h-0">
      <ul class="px-3 list-none flex flex-col gap-2">
        <li
          v-for="(row, index) in rows"
          :key="row.entry.id"
          :data-index="index"
          title="点击复制"
          class="px-3 py-2.5 border panel flex gap-3 cursor-pointer transition-colors items-center"
          :class="index === selectedIndex ? 'bg-brand/10 border-brand/20' : 'border-ink/5 hover:bg-ink/5'"
          @click="copyEntry(row.entry)"
        >
          <div class="text-ink/45 rounded-md bg-ink/5 flex-center shrink-0 h-7 w-7">
            <span class="i-mdi-text-short h-4 w-4" />
          </div>

          <div class="flex flex-1 flex-col gap-1 min-w-0">
            <p class="text-sm leading-relaxed whitespace-pre-wrap break-all" :class="row.singleLine ? 'line-clamp-1' : 'line-clamp-3'">
              <span v-for="(segment, i) in row.segments" :key="i" :class="segment.matched && MATCH_CLASS">{{ segment.text }}</span>
            </p>
            <span class="text-xs text-ink/45 truncate">{{ row.charCount }} 字符 · {{ formatRelativeTime(row.entry.last_used_at) }}</span>
          </div>

          <div class="flex-center shrink-0 gap-1">
            <n-button quaternary size="tiny" class="icon-btn" title="复制" @click.stop="copyEntry(row.entry)">
              <template #icon>
                <span class="i-mdi-content-copy h-4 w-4" />
              </template>
            </n-button>
            <n-button quaternary size="tiny" class="icon-btn" :title="row.entry.pinned ? '取消置顶' : '置顶'" @click.stop="togglePinned(row.entry)">
              <template #icon>
                <span :class="row.entry.pinned ? 'i-mdi-pin text-brand' : 'i-mdi-pin-outline'" class="h-4 w-4" />
              </template>
            </n-button>
            <n-button quaternary size="tiny" class="icon-btn" title="删除" @click.stop="deleteEntry(row.entry)">
              <template #icon>
                <span class="i-mdi-delete-outline h-4 w-4" />
              </template>
            </n-button>
          </div>
        </li>
      </ul>
    </n-scrollbar>
  </div>
</template>
