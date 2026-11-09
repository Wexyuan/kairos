import type { GetSettingsResp } from '@/services/bindings'
import { SettingsService } from '@/services/bindings'

// STORAGE_KEYS lists the persistent storage keys for the application.
export const STORAGE_KEYS = {
  colorScheme: 'kairos-color-scheme',
  notesWorkspace: 'kairos-notes-workspace',
  noteMaxWidth: 'kairos-note-max-width',
} as const

// SEARCH_DELAY is the delay in milliseconds before triggering a search request.
export const SEARCH_DELAY = 250

// PAGE_SIZES is the list of available page sizes for list views.
export const PAGE_SIZES: number[] = [10, 20, 50, 100]

// DEFAULT_PAGE_SIZE is the default page size for list views.
export const DEFAULT_PAGE_SIZE = 20

// DEFAULT_NOTE_MAX_IMAGE_SIZE is the maximum note image size in bytes.
export const DEFAULT_NOTE_MAX_IMAGE_SIZE = 10 << 20

// DEFAULT_NOTE_MAX_WIDTH is the default maximum note width in pixels.
export const DEFAULT_NOTE_MAX_WIDTH = 960

// NOTE_MAX_WIDTH_RANGE is the range of supported note widths in pixels.
export const NOTE_MAX_WIDTH_RANGE = { min: 480, max: 1600, step: 40 } as const

// toNoteMaxWidth clips the note width to the supported range.
function toNoteMaxWidth(value: unknown): number {
  const width = typeof value === 'number' ? value : Number(value)
  if (!Number.isFinite(width))
    return DEFAULT_NOTE_MAX_WIDTH
  return Math.min(NOTE_MAX_WIDTH_RANGE.max, Math.max(NOTE_MAX_WIDTH_RANGE.min, Math.round(width)))
}

// useAppStore defines a store for managing global application state.
export const useAppStore = defineStore('app', () => {
  const sidebarCollapsed = ref(false)
  const settings = ref<GetSettingsResp | null>(null)

  const storedNoteMaxWidth = useLocalStorage(STORAGE_KEYS.noteMaxWidth, DEFAULT_NOTE_MAX_WIDTH)

  const noteMaxImageSize = computed(() => settings.value?.note_max_image_size ?? DEFAULT_NOTE_MAX_IMAGE_SIZE)

  const noteMaxWidth = computed(() => toNoteMaxWidth(storedNoteMaxWidth.value))

  function toggleSidebar(): void {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  function setNoteMaxWidth(value: number | null): void {
    storedNoteMaxWidth.value = toNoteMaxWidth(value)
  }

  // getSettings fetches the application settings from the backend.
  async function getSettings(): Promise<void> {
    try {
      settings.value = await SettingsService.GetSettings()
    }
    catch {
      settings.value = null
    }
  }

  return {
    sidebarCollapsed,
    settings,
    noteMaxImageSize,
    noteMaxWidth,
    toggleSidebar,
    setNoteMaxWidth,
    getSettings,
  }
})
