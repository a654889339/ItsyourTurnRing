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

    try {
      my.showLoading({ content: '发起支付...' })
      const result = await app.request({
        url: `/orders/${id}/prepay`,
        method: 'POST',
        data: { pay_method: 'alipay' },
        showLoading: false
      })
      my.hideLoading()

      if (!result.need_pay) {
        my.showToast({ content: '支付成功' })
        this.loadOrders()
        return
      }

      const tradeNO = result.pay_params.tradeNO
      my.tradePay({
        tradeNO: tradeNO,
        success: (res) => {
          if (res.resultCode === '9000') {
            my.showToast({ content: '支付成功' })
            this.loadOrders()
          } else {
            my.showToast({ content: '支付取消' })
          }
        },
        fail: () => {
          my.showToast({ content: '支付取消' })
        }
      })
    } catch (err) {
      my.hideLoading()
      console.error('发起支付失败', err)
    }
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
