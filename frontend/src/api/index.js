import axios from 'axios'

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 10000
})

// 请求拦截器
api.interceptors.request.use(
  config => {
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  error => Promise.reject(error)
)

// 响应拦截器
api.interceptors.response.use(
  response => {
    const data = response.data
    if (data.code === 0) {
      return data.data
    }
    return Promise.reject(new Error(data.message || '请求失败'))
  },
  error => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    const message = error.response?.data?.message || error.message || '网络错误'
    return Promise.reject(new Error(message))
  }
)

// 认证API
export const authAPI = {
  login: (data) => api.post('/auth/login', data),
  register: (data) => api.post('/auth/register', data),
  getCurrentUser: () => api.get('/auth/me')
}

// 商品API
export const productAPI = {
  list: (params) => api.get('/products', { params }),
  get: (id) => api.get(`/products/${id}`),
  create: (data) => api.post('/products', data),
  update: (id, data) => api.put(`/products/${id}`, data),
  delete: (id) => api.delete(`/products/${id}`),
  getChangeLogs: (id) => api.get(`/products/${id}/change-logs`)
}

// 分类API
export const categoryAPI = {
  list: () => api.get('/categories')
}

// 订单API
export const orderAPI = {
  list: (params) => api.get('/orders', { params }),
  get: (id) => api.get(`/orders/${id}`),
  create: (data) => api.post('/orders', data),
  updateStatus: (id, data) => api.put(`/orders/${id}/status`, data),
  cancel: (id) => api.post(`/orders/${id}/cancel`),
  pay: (id, data) => api.post(`/orders/${id}/pay`, data),
  receive: (id) => api.post(`/orders/${id}/receive`)
}

// 购物车API
export const cartAPI = {
  list: () => api.get('/cart'),
  add: (data) => api.post('/cart', data),
  update: (id, data) => api.put(`/cart/${id}`, data),
  remove: (id) => api.delete(`/cart/${id}`),
  clear: () => api.delete('/cart')
}

// 地址API
export const addressAPI = {
  list: () => api.get('/addresses'),
  get: (id) => api.get(`/addresses/${id}`),
  create: (data) => api.post('/addresses', data),
  update: (id, data) => api.put(`/addresses/${id}`, data),
  delete: (id) => api.delete(`/addresses/${id}`)
}

// 收藏API
export const favoriteAPI = {
  list: (params) => api.get('/favorites', { params }),
  add: (productId) => api.post('/favorites', { product_id: productId }),
  remove: (productId) => api.delete(`/favorites/${productId}`),
  check: (productId) => api.get(`/favorites/${productId}`)
}

// 评价API
export const reviewAPI = {
  list: (params) => api.get('/reviews', { params }),
  create: (data) => api.post('/reviews', data),
  getProductReviews: (productId, params) => api.get(`/reviews/product/${productId}`, { params })
}

// 轮播图API
export const bannerAPI = {
  list: () => api.get('/banners'),
  get: (id) => api.get(`/banners/${id}`),
  create: (data) => api.post('/banners', data),
  update: (id, data) => api.put(`/banners/${id}`, data),
  delete: (id) => api.delete(`/banners/${id}`)
}

// 二维码API
export const qrcodeAPI = {
  list: (params) => api.get('/qrcodes', { params }),
  get: (id) => api.get(`/qrcodes/${id}`),
  create: (data) => api.post('/qrcodes', data),
  delete: (id) => api.delete(`/qrcodes/${id}`),
  getPages: () => api.get('/qrcodes/pages')
}

// 报表API
export const reportAPI = {
  getDashboard: () => api.get('/reports/dashboard'),
  getSales: (params) => api.get('/reports/sales', { params }),
  getProductRank: (params) => api.get('/reports/products', { params })
}

// 上传API
export const uploadAPI = {
  uploadImage: (file) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/upload/image', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  uploadFile: (file) => {
    const formData = new FormData()
    formData.append('file', file)
    return api.post('/upload/file', formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
      timeout: 120000
    })
  },
  uploadBase64: (base64) => api.post('/upload/image', { image: base64 })
}

// 公开API (无需登录)
export const publicAPI = {
  getProducts: (params) => api.get('/public/products', { params }),
  getProduct: (id) => api.get(`/public/products/${id}`),
  getCategories: () => api.get('/public/categories'),
  getBanners: () => api.get('/public/banners'),
  getHome: () => api.get('/public/home')
}

export default api
