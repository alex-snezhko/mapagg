import About from '@/components/About.vue'
import AddMapChoropleth from '@/components/AddMapChoropleth.vue'
import AddMapLinks from '@/components/AddMapLinks.vue'
import AddMapPointsOfInterest from '@/components/AddMapPointsOfInterest.vue'
import Main from '@/components/Main.vue'
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/about',
      name: 'about',
      component: About
    },
    {
      path: '/add-map',
      name: 'add-map',
      component: AddMapLinks
    },
    {
      path: '/add-map/choropleth',
      name: 'add-map-choropleth',
      component: AddMapChoropleth
    },
    {
      path: '/add-map/points-of-interest',
      name: 'add-map-points-of-interest',
      component: AddMapPointsOfInterest,
    },
    {
      path: '/',
      name: 'home',
      component: Main,
    },
  ],
})

export default router
