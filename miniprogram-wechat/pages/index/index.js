const app = getApp()

Page({
  data: {
    banners: [],
    categories: [],
    featuredProducts: [],
    newProducts: [],
    products: [],
    currentCategory: '',
    page: 1,
    hasMore: true,
    loading: false
  },

  onLoad() {
    this.loadHomeData()
    this.loadProducts()
  },

  onPullDownRefresh() {
    this.setData({ page: 1, hasMore: true })
    Promise.all([
      this.loadHomeData(),
      this.loadProducts()
    ]).then(() => {
      wx.stopPullDownRefresh()
    })
  },

  onReachBottom() {
    if (this.data.hasMore && !this.data.loading) {
      this.loadMore()
    }
  },

  async loadHomeData() {
    try {
      const data = await app.request({
        url: '/public/home',
        showLoading: false
      })
      this.setData({
        banners: data.banners || [],
        categories: data.categories || [],
        featuredProducts: data.featured || [],
        newProducts: data.new || []
      })
    } catch (err) {
      console.error('加载首页数据失败', err)
    }
  },

  async loadProducts() {
    if (this.data.loading) return

    this.setData({ loading: true })

    try {
      const result = await app.request({
        url: '/public/products',
        data: {
          page: this.data.page,
          page_size: 10,
          category: this.data.currentCategory
        },
        showLoading: false
      })

      const products = result.data || []
      this.setData({
        products: this.data.page === 1 ? products : [...this.data.products, ...products],
        hasMore: this.data.products.length + products.length < result.total,
        loading: false
      })
    } catch (err) {
      this.setData({ loading: false })
      console.error('加载商品失败', err)
    }
  },

  loadMore() {
    this.setData({ page: this.data.page + 1 })
    this.loadProducts()
  },

  onCategoryTap(e) {
    const code = e.currentTarget.dataset.code
    this.setData({
      currentCategory: this.data.currentCategory === code ? '' : code,
      page: 1,
      products: []
    })
    this.loadProducts()
  },

  onProductTap(e) {
    const id = e.currentTarget.dataset.id
    wx.navigateTo({
      url: `/pages/product/index?id=${id}`
    })
  },

  onBannerTap(e) {
    const banner = e.currentTarget.dataset.banner
    if (banner.link) {
      // 处理链接跳转
      if (banner.link.startsWith('/pages')) {
        wx.navigateTo({ url: banner.link })
      }
    }
  }
})
