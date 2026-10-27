import { createApp, h } from 'vue'
import AppProviders from '@/components/layout/app-providers.vue'
import { setupStore } from '@/stores'
import App from './clipboard-app.vue'
import 'virtual:uno.css'
import './styles/theme.css'

document.addEventListener('contextmenu', event => event.preventDefault())

function setupApp(): void {
  const app = createApp({
    render: () => h(AppProviders, null, { default: () => h(App) }),
  })

  const pinia = setupStore()
  app.use(pinia)

  app.mount('#app')
}

setupApp()
