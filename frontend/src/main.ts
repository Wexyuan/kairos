import { createApp } from 'vue'
import { setupRouter } from '@/router'
import { setupStore } from '@/stores'
import App from './App.vue'
import 'virtual:uno.css'
import './styles/theme.css'

async function setupApp() {
  const app = createApp(App)

  const pinia = setupStore()
  app.use(pinia)

  const router = setupRouter()
  app.use(router)

  await router.isReady()

  app.mount('#app')
}

setupApp()
