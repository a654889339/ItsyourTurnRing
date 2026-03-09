const app = getApp()

Page({
  data: {
    product: null,
    quantity: 1,
    showActionSheet: false,
    actionType: 'cart'
  },

  onLoad(options) {
    if (options.id) {
      this.loadProduct(options.id)
    }
  },

  async loadProduct(id) {
    try {
      const product = await app.request({
        url: `/public/products/${id}`
      })
      this.setData({ product })
    } catch (err) {
      wx.showToast({ title: '商品不存在', icon: 'none' })
      setTimeout(() => wx.navigateBack(), 1500)
    }
  },

  onQuantityChange(e) {
    const type = e.currentTarget.dataset.type
    let quantity = this.data.quantity

    if (type === 'minus' && quantity > 1) {
      quantity--
    } else if (type === 'plus' && quantity < this.data.product.stock) {
      quantity++
    }

    this.setData({ quantity })
  },

  onAddToCart() {
    if (!app.checkLogin()) {
      wx.showToast({ title: '请先登录', icon: 'none' })
      return
    }
    this.setData({
      actionType: 'cart',
      showActionSheet: true
    })
  },

  onBuyNow() {
    if (!app.checkLogin()) {
      wx.showToast({ title: '请先登录', icon: 'none' })
      return
    }
    this.setData({
      actionType: 'buy',
      showActionSheet: true
    })
  },

  onCloseActionSheet() {
    this.setData({ showActionSheet: false })
  },

  async onConfirmAction() {
    const { product, quantity, actionType } = this.data

    try {
      await app.request({
        url: '/cart',
        method: 'POST',
        data: {
          product_id: product.id,
          quantity: quantity
        }
      })

      this.setData({ showActionSheet: false })

      if (actionType === 'buy') {
        wx.switchTab({ url: '/pages/cart/index' })
      } else {
        wx.showToast({ title: '已加入购物车', icon: 'success' })
      }
    } catch (err) {
      console.error('操作失败', err)
    }
  }
})
