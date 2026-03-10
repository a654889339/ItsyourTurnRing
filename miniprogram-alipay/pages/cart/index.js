const app = getApp()

Page({
  data: {
    cartItems: [],
    selectedIds: [],
    selectedMap: {},
    totalPrice: 0
  },

  onShow() {
    if (app.checkLogin()) {
      this.loadCart()
    }
  },

  async loadCart() {
    try {
      var items = await app.request({ url: '/cart', showLoading: false })
      var list = items || []
      this.setData({
        cartItems: list,
        selectedIds: list.map(function(item) { return Number(item.id) })
      })
      this.updateSelectedMap()
      this.calculateTotal()
    } catch (err) {
      console.error(err)
    }
  },

  updateSelectedMap() {
    var map = {}
    this.data.selectedIds.forEach(function(id) { map[id] = true })
    this.setData({ selectedMap: map })
  },

  calculateTotal() {
    var total = 0
    var selectedMap = this.data.selectedMap
    this.data.cartItems.forEach(function(item) {
      if (selectedMap[item.id]) {
        var price = (item.product && item.product.price) ? Number(item.product.price) : 0
        if (item.spec && item.spec.price_adjustment) {
          price += Number(item.spec.price_adjustment)
        }
        total += price * item.quantity
      }
    })
    this.setData({ totalPrice: total.toFixed(2) })
  },

  onSelectItem(e) {
    var id = Number(e.target.dataset.id)
    var selectedIds = this.data.selectedIds.slice()
    var idx = selectedIds.indexOf(id)
    if (idx > -1) {
      selectedIds.splice(idx, 1)
    } else {
      selectedIds.push(id)
    }
    this.setData({ selectedIds: selectedIds })
    this.updateSelectedMap()
    this.calculateTotal()
  },

  onSelectAll() {
    var cartItems = this.data.cartItems
    var selectedIds = this.data.selectedIds
    if (selectedIds.length === cartItems.length) {
      this.setData({ selectedIds: [] })
    } else {
      this.setData({ selectedIds: cartItems.map(function(item) { return Number(item.id) }) })
    }
    this.updateSelectedMap()
    this.calculateTotal()
  },

  async onQuantityChange(e) {
    var type = e.target.dataset.type
    var id = Number(e.target.dataset.id)
    var item = this.data.cartItems.find(function(i) { return i.id === id })
    if (!item) return

    let newQuantity = item.quantity
    if (type === 'minus' && newQuantity > 1) newQuantity--
    else if (type === 'plus' && item.product && newQuantity < item.product.stock) newQuantity++
    else return

    try {
      await app.request({ url: `/cart/${id}`, method: 'PUT', data: { quantity: newQuantity } })
      item.quantity = newQuantity
      this.setData({ cartItems: this.data.cartItems })
      this.calculateTotal()
    } catch (err) {
      console.error(err)
    }
  },

  async onDeleteItem(e) {
    var id = Number(e.target.dataset.id)
    my.confirm({
      content: '确定要删除这个商品吗？',
      success: async (res) => {
        if (res.confirm) {
          await app.request({ url: `/cart/${id}`, method: 'DELETE' })
          this.setData({
            cartItems: this.data.cartItems.filter(item => item.id !== id),
            selectedIds: this.data.selectedIds.filter(i => i !== id)
          })
          this.updateSelectedMap()
          this.calculateTotal()
        }
      }
    })
  },

  onCheckout() {
    if (this.data.selectedIds.length === 0) {
      my.showToast({ content: '请选择商品' })
      return
    }
    my.setStorageSync({ key: 'checkoutCartIds', data: this.data.selectedIds })
    my.navigateTo({ url: '/pages/checkout/index' })
  },

  onLogin() {
    app.alipayLogin().then(() => this.loadCart())
  }
})
