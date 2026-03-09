const app = getApp()

Page({
  data: {
    product: null,
    quantity: 1,
    showActionSheet: false,
    actionType: 'cart'
  },

  onLoad(query) {
    if (query.id) {
      this.loadProduct(query.id)
    }
  },

  async loadProduct(id) {
    try {
      const product = await app.request({ url: `/public/products/${id}` })
      this.setData({ product })
    } catch (err) {
      my.showToast({ content: '商品不存在' })
      setTimeout(() => my.navigateBack(), 1500)
    }
  },

  onQuantityChange(e) {
    const type = e.target.dataset.type
    let quantity = this.data.quantity
    if (type === 'minus' && quantity > 1) quantity--
    else if (type === 'plus' && quantity < this.data.product.stock) quantity++
    this.setData({ quantity })
  },

  onAddToCart() {
    if (!app.checkLogin()) {
      my.showToast({ content: '请先登录' })
      return
    }
    this.setData({ actionType: 'cart', showActionSheet: true })
  },

  onBuyNow() {
    if (!app.checkLogin()) {
      my.showToast({ content: '请先登录' })
      return
    }
    this.setData({ actionType: 'buy', showActionSheet: true })
  },

  onCloseActionSheet() {
    this.setData({ showActionSheet: false })
  },

  async onConfirmAction() {
    try {
      await app.request({
        url: '/cart',
        method: 'POST',
        data: { product_id: this.data.product.id, quantity: this.data.quantity }
      })
      this.setData({ showActionSheet: false })
      if (this.data.actionType === 'buy') {
        my.switchTab({ url: '/pages/cart/index' })
      } else {
        my.showToast({ content: '已加入购物车' })
      }
    } catch (err) {
      console.error(err)
    }
  }
})
