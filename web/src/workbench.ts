import { createApp, type App as VueApplication } from 'vue'
import { createPinia } from 'pinia'
import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import App from './App.vue'
import AuditPage from './pages/AuditPage.vue'
import BundlePage from './pages/BundlePage.vue'
import ConfigPage from './pages/ConfigPage.vue'

const workbenchRoutes: RouteRecordRaw[] = [
  { path: '/', redirect: '/bundles' },
  { path: '/bundles', name: 'bundle-validation', component: BundlePage },
  { path: '/configurations', name: 'configuration-impact', component: ConfigPage },
  { path: '/audits', name: 'release-audit', component: AuditPage }
]

export function mountConfigurationWorkbench(target: string): VueApplication {
  const navigation = createRouter({ history: createWebHistory(), routes: workbenchRoutes })
  const state = createPinia()
  const application = createApp(App)
  application.use(state)
  application.use(navigation)
  application.mount(target)
  return application
}

