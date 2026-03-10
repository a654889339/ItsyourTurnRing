const app = getApp()

Page({
  data: {
    cartItems: [],
    selectedIds: [],
    totalPrice: 0,
    isLoggedIn: false
  },

  onShow() {
    this.setData({ isLoggedIn: app.checkLogin() })
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
      if (selectedIds.indexOf(item.id) >= 0) {
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

    if (selectedIds.indexOf(id) >= 0) {
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
    } else if (type === 'plus' && newQuantity < (item.product?.stock || 999)) {
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

    xhs.showModal({
      title: '提示',
      content: '确定要删除这个商品吗？',
      success: (res) => {
        if (res.confirm) {
          app.request({
            url: `/cart/${id}`,
            method: 'DELETE'
          }).then(() => {
            this.setData({
              cartItems: this.data.cartItems.filter(item => item.id !== id),
              selectedIds: this.data.selectedIds.filter(i => i !== id)
            })
            this.calculateTotal()
          }).catch(err => {
            console.error('删除失败', err)
          })
        }
      }
    })
  },

  onCheckout() {
    const { selectedIds, cartItems } = this.data

    if (selectedIds.length === 0) {
      xhs.showToast({ title: '请选择商品', icon: 'none' })
      return
    }

    xhs.setStorageSync('checkoutCartIds', selectedIds)
    xhs.navigateTo({ url: '/pages/order/index?action=create' })
  },

  onLogin() {
    app.xhsLogin(true).then(() => {
      this.setData({ isLoggedIn: true })
      this.loadCart()
    }).catch(err => {
      console.error('登录失败', err)
    })
  }
})
