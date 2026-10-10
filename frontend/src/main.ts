import './assets/main.css'

import { createApp } from 'vue'
import { createPinia } from 'pinia'

import App from './App.vue'
import router from './router'
import { useAuthStore } from './stores/auth'

const app = createApp(App)

app.use(createPinia())
app.use(router)

// Resolve any persisted session before the first paint so the navbar and the
// route guards agree on the auth state from the very first frame.
async function bootstrap(): Promise<void> {
  await useAuthStore().initialize()
  app.mount('#app')
}

void bootstrap()
