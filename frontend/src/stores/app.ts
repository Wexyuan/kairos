import type { GetSettingsResp } from '@/services/bindings'
import { SettingsService } from '@/services/bindings'

// STORAGE_KEYS lists the persistent storage keys for the application.
export const STORAGE_KEYS = {
  colorScheme: 'kairos-color-scheme',
  notesWorkspace: 'kairos-notes-workspace',
} as const

// SEARCH_DELAY is the delay in milliseconds before triggering a search request.
export const SEARCH_DELAY = 250

// PAGE_SIZES is the list of available page sizes for list views.
export const PAGE_SIZES: number[] = [10, 20, 50, 100]

// DEFAULT_PAGE_SIZE is the default page size for list views.
export const DEFAULT_PAGE_SIZE = 20

// DEFAULT_NOTE_MAX_IMAGE_SIZE is the maximum note image size in bytes.
export const DEFAULT_NOTE_MAX_IMAGE_SIZE = 10 << 20

// useAppStore defines a store for managing global application state.
export const useAppStore = defineStore('app', () => {
  const sidebarCollapsed = ref(false)
  const settings = ref<GetSettingsResp | null>(null)

  const noteMaxImageSize = computed(() => settings.value?.note_max_image_size ?? DEFAULT_NOTE_MAX_IMAGE_SIZE)

  function toggleSidebar(): void {
    sidebarCollapsed.value = !sidebarCollapsed.value
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
    toggleSidebar,
    getSettings,
  }
})
