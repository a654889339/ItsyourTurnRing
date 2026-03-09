const app = getApp()

Page({
  data: {
    orders: [],
    currentStatus: ''
  },

  onShow() {
    if (app.checkLogin()) {
      this.loadOrders()
    }
  },

  onPullDownRefresh() {
    this.loadOrders().then(() => my.stopPullDownRefresh())
  },

  async loadOrders() {
    try {
      const result = await app.request({
        url: '/orders',
        data: { page: 1, page_size: 50, status: this.data.currentStatus }
      })
      this.setData({ orders: result.data || [] })
    } catch (err) {
      console.error(err)
    }
  },

  onTabChange(e) {
    this.setData({ currentStatus: e.target.dataset.status })
    this.loadOrders()
  },

  async onPay(e) {
    const id = e.target.dataset.id
    my.confirm({
      content: '确认支付？（模拟支付）',
      success: async (res) => {
        if (res.confirm) {
          await app.request({ url: `/orders/${id}/pay`, method: 'POST', data: { pay_method: 'alipay' } })
          my.showToast({ content: '支付成功' })
          this.loadOrders()
        }
      }
    })
  },

  async onCancel(e) {
    const id = e.target.dataset.id
    my.confirm({
      content: '确定要取消订单吗？',
      success: async (res) => {
        if (res.confirm) {
          await app.request({ url: `/orders/${id}/cancel`, method: 'POST' })
          this.loadOrders()
        }
      }
    })
  },

  async onReceive(e) {
    const id = e.target.dataset.id
    my.confirm({
      content: '确认已收到商品？',
      success: async (res) => {
        if (res.confirm) {
          await app.request({ url: `/orders/${id}/receive`, method: 'POST' })
          this.loadOrders()
        }
      }
    })
  },

  onLogin() {
    app.alipayLogin().then(() => this.loadOrders())
  }
})
