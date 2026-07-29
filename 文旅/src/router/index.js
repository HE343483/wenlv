import { createRouter, createWebHistory } from 'vue-router'
import Landing from '../views/Landing/Landing.vue'
import Login from '../views/Login/Login.vue'
import Main from '../views/Main/Main.vue'
import HomePage from '../views/HomePage/HomePage.vue'
import CollectionPage from '../views/CollectionPage/CollectionPage.vue'
import Sichuan3DMap from '../views/Sichuan3DMap/Sichuan3DMap.vue'
import ItineraryPage from '../views/ItineraryPage/ItineraryPage.vue'
import NotificationsPage from '../views/NotificationsPage/NotificationsPage.vue'
import ProfilePage from '../views/ProfilePage/ProfilePage.vue'
import { auth } from '../stores/auth.js'

const routes = [
  { path: '/', name: 'Landing', component: Landing },
  { path: '/login', name: 'Login', component: Login },
  {
    path: '/main',
    name: 'Main',
    component: Main,
    meta: { requiresAuth: true },
    children: [
      { path: '',              redirect: '/main/home' },
      { path: 'home',          name: 'Home',          component: HomePage },
      { path: 'collection',    name: 'Collection',    component: CollectionPage },
      { path: 'map',           name: 'Map',           component: Sichuan3DMap },
      { path: 'itinerary',     name: 'Itinerary',     component: ItineraryPage },
      { path: 'notifications', name: 'Notifications', component: NotificationsPage },
      { path: 'profile',       name: 'Profile',       component: ProfilePage }
    ]
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach((to, from, next) => {
  if (to.meta.requiresAuth && !auth.isLoggedIn) {
    next({ name: 'Login' })
  } else {
    next()
  }
})

export default router
