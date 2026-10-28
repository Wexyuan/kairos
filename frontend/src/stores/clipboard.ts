import type { ClipboardEntryItem } from '@/services/bindings'
import { ClipboardService } from '@/services/bindings'

// toClipboardEntries converts nullable entries into a non-null clipboard entry array.
export function toClipboardEntries(entries: (ClipboardEntryItem | null)[] | null | undefined): ClipboardEntryItem[] {
  return (entries ?? []).filter((entry): entry is ClipboardEntryItem => entry !== null)
}

// formatRelativeTime formats a time string as a relative time Chinese label.
export function formatRelativeTime(value: string): string {
  const time = new Date(value).getTime()
  if (Number.isNaN(time))
    return ''
  const diff = Date.now() - time
  if (diff < 60 * 1000)
    return '刚刚'
  if (diff < 60 * 60 * 1000)
    return `${Math.floor(diff / (60 * 1000))} 分钟前`
  if (diff < 24 * 60 * 60 * 1000)
    return `${Math.floor(diff / (60 * 60 * 1000))} 小时前`
  if (diff < 7 * 24 * 60 * 60 * 1000)
    return `${Math.floor(diff / (24 * 60 * 60 * 1000))} 天前`
  return new Date(time).toLocaleString('zh-CN', { hour12: false })
}

// useClipboardStore defines a store for managing clipboard history.
export const useClipboardStore = defineStore('clipboard', () => {
  const entries = ref<ClipboardEntryItem[]>([])
  const total = ref(0)
  const page = ref(1)
  const pageSize = ref(20)
  const keyword = ref('')
  const listLoading = ref(false)

  // searchEntries searches clipboard entries.
  async function searchEntries(): Promise<void> {
    page.value = 1
    await listEntries()
  }

  // copyEntry copies an entry back to the system clipboard.
  async function copyEntry(id: string): Promise<void> {
    await ClipboardService.CopyClipboardEntry({ id })
    await listEntries()
  }

  // deleteEntry deletes clipboard entry.
  async function deleteEntry(id: string): Promise<void> {
    await ClipboardService.DeleteClipboardEntry({ id })
    await listEntries()
  }

  // clearEntries deletes all clipboard entries.
  async function clearEntries(): Promise<void> {
    await ClipboardService.ClearClipboardEntries()
    await listEntries()
  }

  // listEntries lists clipboard entries.
  async function listEntries(): Promise<void> {
    listLoading.value = true
    try {
      const result = await ClipboardService.ListClipboardEntries({
        page: page.value,
        page_size: pageSize.value,
        keyword: keyword.value.trim(),
      })
      entries.value = toClipboardEntries(result?.list)
      total.value = result?.total ?? 0
      page.value = result?.page ?? page.value
      pageSize.value = result?.page_size ?? pageSize.value

      const lastPage = Math.max(1, Math.ceil(total.value / pageSize.value))
      if (!entries.value.length && page.value > lastPage) {
        page.value = lastPage
        await listEntries()
      }
    }
    finally {
      listLoading.value = false
    }
  }

  return {
    entries,
    total,
    page,
    pageSize,
    keyword,
    listLoading,
    searchEntries,
    copyEntry,
    deleteEntry,
    clearEntries,
    listEntries,
  }
})
