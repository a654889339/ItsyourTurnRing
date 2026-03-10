const app = getApp()

Page({
  data: {
    isLogin: false,
    cartItems: [],
    selectedIds: [],
    selectedMap: {},
    totalPrice: 0
  },

  onShow() {
    const isLogin = app.checkLogin()
    this.setData({ isLogin })
    if (isLogin) {
      this.loadCart()
    }
  },

  async loadCart() {
    try {
      const items = await app.request({
        url: '/cart',
        showLoading: false
      })
      const list = items || []
      this.setData({
        cartItems: list,
        selectedIds: list.map(item => Number(item.id))
      })
      this.updateSelectedMap()
      this.calculateTotal()
    } catch (err) {
      console.error('加载购物车失败', err)
    }
  },

  updateSelectedMap() {
    const map = {}
    this.data.selectedIds.forEach(id => { map[id] = true })
    this.setData({ selectedMap: map })
  },

  calculateTotal() {
    const { cartItems, selectedMap } = this.data
    let total = 0

    cartItems.forEach(item => {
      if (selectedMap[item.id]) {
        let price = (item.product && item.product.price) ? Number(item.product.price) : 0
        if (item.spec && item.spec.price_adjustment) {
          price += Number(item.spec.price_adjustment)
        }
        total += price * item.quantity
      }
    })

    this.setData({ totalPrice: total.toFixed(2) })
  },

  onSelectItem(e) {
    const id = Number(e.currentTarget.dataset.id)
    let selectedIds = this.data.selectedIds.slice()

    const idx = selectedIds.indexOf(id)
    if (idx > -1) {
      selectedIds.splice(idx, 1)
    } else {
      selectedIds.push(id)
    }

    this.setData({ selectedIds })
    this.updateSelectedMap()
    this.calculateTotal()
  },

  onSelectAll() {
    const { cartItems, selectedIds } = this.data

    if (selectedIds.length === cartItems.length) {
      this.setData({ selectedIds: [] })
    } else {
      this.setData({ selectedIds: cartItems.map(item => Number(item.id)) })
    }
    this.updateSelectedMap()
    this.calculateTotal()
  },

  async onQuantityChange(e) {
    const { type } = e.currentTarget.dataset
    const id = Number(e.currentTarget.dataset.id)
    const item = this.data.cartItems.find(i => i.id === id)
    if (!item) return

    let newQuantity = item.quantity
    if (type === 'minus' && newQuantity > 1) {
      newQuantity--
    } else if (type === 'plus' && item.product && newQuantity < item.product.stock) {
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
    const id = Number(e.currentTarget.dataset.id)

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
            this.updateSelectedMap()
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
    wx.navigateTo({ url: '/pages/checkout/index' })
  },

  onLogin() {
    app.wxLogin(true).then(() => {
      this.setData({ isLogin: true })
      this.loadCart()
    }).catch(err => {
      console.error('登录失败', err)
    })
  }
})
