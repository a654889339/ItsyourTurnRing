const app = getApp()

Page({
  data: {
    orders: [],
    currentStatus: '',
    statusTabs: [
      { label: '全部', value: '' },
      { label: '待付款', value: 'pending' },
      { label: '已付款', value: 'paid' },
      { label: '已发货', value: 'shipped' },
      { label: '已完成', value: 'completed' }
    ]
  },

  onShow() {
    if (app.checkLogin()) {
      this.loadOrders()
    }
  },

  onPullDownRefresh() {
    this.loadOrders().then(() => {
      wx.stopPullDownRefresh()
    })
  },

  async loadOrders() {
    try {
      const result = await app.request({
        url: '/orders',
        data: {
          page: 1,
          page_size: 50,
          status: this.data.currentStatus
        }
      })
      this.setData({ orders: result.data || [] })
    } catch (err) {
      console.error('加载订单失败', err)
    }
  },

  onTabChange(e) {
    const status = e.currentTarget.dataset.status
    this.setData({ currentStatus: status })
    this.loadOrders()
  },

  getStatusText(status) {
    const map = {
      pending: '待付款',
      paid: '已付款',
      shipped: '已发货',
      received: '已收货',
      completed: '已完成',
      cancelled: '已取消'
    }
    return map[status] || status
  },

  getTotalCount(order) {
    return (order.items || []).reduce((sum, item) => sum + item.quantity, 0)
  },

  async onPay(e) {
    const id = e.currentTarget.dataset.id

    wx.showModal({
      title: '提示',
      content: '确认支付？（模拟支付）',
      success: async (res) => {
        if (res.confirm) {
          try {
            await app.request({
              url: `/orders/${id}/pay`,
              method: 'POST',
              data: { pay_method: 'wechat' }
            })
            wx.showToast({ title: '支付成功', icon: 'success' })
            this.loadOrders()
          } catch (err) {
            console.error('支付失败', err)
          }
        }
      }
    })
  },

  async onCancel(e) {
    const id = e.currentTarget.dataset.id

    wx.showModal({
      title: '提示',
      content: '确定要取消订单吗？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await app.request({
              url: `/orders/${id}/cancel`,
              method: 'POST'
            })
            this.loadOrders()
          } catch (err) {
            console.error('取消失败', err)
          }
        }
      }
    })
  },

  async onReceive(e) {
    const id = e.currentTarget.dataset.id

    wx.showModal({
      title: '提示',
      content: '确认已收到商品？',
      success: async (res) => {
        if (res.confirm) {
          try {
            await app.request({
              url: `/orders/${id}/receive`,
              method: 'POST'
            })
            this.loadOrders()
          } catch (err) {
            console.error('确认收货失败', err)
          }
        }
      }
    })
  },

  onLogin() {
    app.wxLogin().then(() => {
      this.loadOrders()
    })
  }
})
