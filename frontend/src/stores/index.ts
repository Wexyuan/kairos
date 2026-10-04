import type { Pinia } from 'pinia'
import { createPinia } from 'pinia'

// setupStore setups the pinia instance for the application.
export function setupStore(): Pinia {
  return createPinia()
}
