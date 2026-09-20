import { createRouter, createWebHistory } from 'vue-router'
import { hasToken } from '@/utils/token'

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
        {
          path: 'favorites',
          name: 'internal-favorites',
          component: () => import('@/views/FavoritesPage/FavoritesPage.vue'),
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
    {
      // AI 行程规划 — 首页(多城市行程表单 / 历史计划)
      path: '/trip',
      name: 'trip-planner',
      component: () => import('@/views/LandingView/LandingView.vue'),
    },
    {
      // AI 行程规划 — 结果页(地图 / 预算 / 知识图谱 / AI 问答 / 导出图片)
      path: '/trip/result',
      name: 'trip-result',
      component: () => import('@/views/ResultView/ResultView.vue'),
    },
  ],
})

export default router

// 全局前置守卫：内部页面需登录
router.beforeEach((to) => {
  const meta = to.meta as { guest?: boolean } | undefined
  // 登录/注册页无条件放行(未登录时的唯一去处,放行判断必须在 token 检查之前,否则死循环白屏)
  if (to.name === 'login' || to.name === 'register') {
    return hasToken() ? { name: 'home' } : true
  }
  if (to.name === 'home') return true
  if (to.path.startsWith('/scenic/') || to.path.startsWith('/food/')) return true
  if (!hasToken()) {
    return { name: 'login' }
  }
  if (meta?.guest) return { name: 'home' }
  return true
})
