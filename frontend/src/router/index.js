import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../store/auth'

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/Login.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/',
    component: () => import('../views/Layout.vue'),
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'Dashboard',
        component: () => import('../views/Dashboard.vue')
      },
      {
        path: 'products',
        name: 'Products',
        component: () => import('../views/Products.vue')
      },
      {
        path: 'orders',
        name: 'Orders',
        component: () => import('../views/Orders.vue')
      },
      {
        path: 'create-order',
        name: 'CreateOrder',
        component: () => import('../views/CreateOrder.vue')
      },
      {
        path: 'banners',
        name: 'Banners',
        component: () => import('../views/Banners.vue')
      },
      {
        path: 'reports',
        name: 'Reports',
        component: () => import('../views/Reports.vue')
      },
      {
        path: 'qrcodes',
        name: 'QRCodes',
        component: () => import('../views/QRCodes.vue')
      }
    ]
  },
  // 商城H5页面
  {
    path: '/shop',
    name: 'Shop',
    component: () => import('../views/shop/Home.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/shop/product/:id',
    name: 'ShopProduct',
    component: () => import('../views/shop/Product.vue'),
    meta: { requiresAuth: false }
  },
  {
    path: '/shop/cart',
    name: 'ShopCart',
    component: () => import('../views/shop/Cart.vue'),
    meta: { requiresAuth: true }
  },
  {
    path: '/shop/orders',
    name: 'ShopOrders',
    component: () => import('../views/shop/OrderList.vue'),
    meta: { requiresAuth: true }
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

router.beforeEach(async (to, from, next) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    next('/login')
    return
  }

  if (to.path === '/login' && authStore.isAuthenticated) {
    next('/')
    return
  }

  // 如果已登录但没有用户信息,获取用户信息
  if (authStore.isAuthenticated && !authStore.user) {
    try {
      await authStore.fetchCurrentUser()
    } catch (error) {
      next('/login')
      return
    }
  }

  next()
})

export default router
