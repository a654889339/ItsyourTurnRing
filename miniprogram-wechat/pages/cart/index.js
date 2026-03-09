const app = getApp()

Page({
  data: {
    cartItems: [],
    selectedIds: [],
    totalPrice: 0
  },

  onShow() {
    if (app.checkLogin()) {
      this.loadCart()
    }
  },

  async loadCart() {
    try {
      const items = await app.request({
        url: '/cart',
        showLoading: false
      })
      this.setData({
        cartItems: items || [],
        selectedIds: (items || []).map(item => item.id)
      })
      this.calculateTotal()
    } catch (err) {
      console.error('加载购物车失败', err)
    }
  },

  calculateTotal() {
    const { cartItems, selectedIds } = this.data
    let total = 0

    cartItems.forEach(item => {
      if (selectedIds.includes(item.id)) {
        let price = item.product?.price || 0
        if (item.spec) {
          price += item.spec.price_adjustment || 0
        }
        total += price * item.quantity
      }
    })

    this.setData({ totalPrice: total })
  },

  onSelectItem(e) {
    const id = e.currentTarget.dataset.id
    let { selectedIds } = this.data

    if (selectedIds.includes(id)) {
      selectedIds = selectedIds.filter(i => i !== id)
    } else {
      selectedIds.push(id)
    }

    this.setData({ selectedIds })
    this.calculateTotal()
  },

  onSelectAll() {
    const { cartItems, selectedIds } = this.data

    if (selectedIds.length === cartItems.length) {
      this.setData({ selectedIds: [] })
    } else {
      this.setData({ selectedIds: cartItems.map(item => item.id) })
    }
    this.calculateTotal()
  },

  async onQuantityChange(e) {
    const { id, type } = e.currentTarget.dataset
    const item = this.data.cartItems.find(i => i.id === id)
    if (!item) return

    let newQuantity = item.quantity
    if (type === 'minus' && newQuantity > 1) {
      newQuantity--
    } else if (type === 'plus' && newQuantity < item.product?.stock) {
      newQuantity++
    } else {
      return
    }

    try {
      await app.request({
        url: `/cart/${id}`,
        method: 'PUT',
        data: { quantity: newQuantity }
      })
      item.quantity = newQuantity
      this.setData({ cartItems: this.data.cartItems })
      this.calculateTotal()
    } catch (err) {
      console.error('更新数量失败', err)
    }
  },

  async onDeleteItem(e) {
    const id = e.currentTarget.dataset.id

    wx.showModal({
      title: '提示',
      content: '确定要删除这个商品吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await app.request({
              url: `/cart/${id}`,
              method: 'DELETE'
            })
            this.setData({
              cartItems: this.data.cartItems.filter(item => item.id !== id),
              selectedIds: this.data.selectedIds.filter(i => i !== id)
            })
            this.calculateTotal()
          } catch (err) {
            console.error('删除失败', err)
          }
        }
      }
    })
  },

  onCheckout() {
    const { selectedIds, cartItems } = this.data

    if (selectedIds.length === 0) {
      wx.showToast({ title: '请选择商品', icon: 'none' })
      return
    }

    // 存储选中的购物车ID
    wx.setStorageSync('checkoutCartIds', selectedIds)
    wx.navigateTo({ url: '/pages/order/create' })
  },

  onLogin() {
    app.wxLogin().then(() => {
      this.loadCart()
    }).catch(err => {
      console.error('登录失败', err)
    })
  }
})
