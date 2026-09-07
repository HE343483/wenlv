import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      // 公开首页 — 品牌展示 + 景点探索
      path: '/',
      name: 'home',
      component: () => import('@/views/HomeView/HomeView.vue'),
    },
    {
      // 登录页 — 基本表单
      path: '/login',
      name: 'login',
      component: () => import('@/views/LoginView/LoginView.vue'),
    },
    {
      // 注册页 — 基本表单
      path: '/register',
      name: 'register',
      component: () => import('@/views/RegisterView/RegisterView.vue'),
    },
    {
      // 内部页面 — 登录后主框架
      path: '/home',
      component: () => import('@/views/HomeLayout/HomeLayout.vue'),
      redirect: '/home/index',
      children: [
        {
          path: 'index',
          name: 'internal-home',
          component: () => import('@/views/HomeLayout/HomeDashboard.vue'),
        },
        {
          path: 'explore',
          name: 'internal-explore',
          component: () => import('@/views/ExplorePage/ExplorePage.vue'),
        },
        {
          path: 'food',
          name: 'internal-food',
          component: () => import('@/views/FoodPage/FoodPage.vue'),
        },
        {
          path: 'routes',
          name: 'internal-routes',
          component: () => import('@/views/RoutesPage/RoutesPage.vue'),
        },
        {
          path: 'profile',
          name: 'internal-profile',
          component: () => import('@/views/ProfilePage/ProfilePage.vue'),
        },
      ],
    },
    {
      // 景点详情页
      path: '/scenic/:id',
      name: 'scenic-detail',
      component: () => import('@/views/ScenicDetail/ScenicDetail.vue'),
    },
    {
      // 美食详情页
      path: '/food/:id',
      name: 'food-detail',
      component: () => import('@/views/FoodDetail/FoodDetail.vue'),
    },
  ],
})

export default router
