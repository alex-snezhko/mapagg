import AddMap from '@/components/AddMap.vue'
import Main from '@/components/Main.vue'
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/add-map',
      name: 'add-map',
      component: AddMap
    },
    {
      path: '/',
      name: 'home',
      component: Main,
    },
  ],
})

export default router
