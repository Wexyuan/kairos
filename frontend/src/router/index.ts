import type { Router } from 'vue-router'
import { createRouter, createWebHistory } from 'vue-router'

declare module 'vue-router' {
  interface RouteMeta {
    title?: string
    icon?: string
  }
}

// setupRouter setups the router for the application.
export function setupRouter(): Router {
  return createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    routes: [
      {
        path: '/',
        redirect: '/notes',
      },
      {
        path: '/notes',
        name: 'notes',
        component: () => import('@/pages/notes/index.vue'),
        meta: { title: '笔记', icon: 'i-mdi-notebook-outline' },
      },
      {
        path: '/settings',
        name: 'settings',
        component: () => import('@/pages/settings/index.vue'),
        meta: { title: '设置', icon: 'i-mdi-cog' },
      },
    ],
  })
}
