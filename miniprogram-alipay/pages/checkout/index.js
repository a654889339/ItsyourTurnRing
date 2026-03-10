const app = getApp()

Page({
  data: {
    cartItems: [],
    addresses: [],
    selectedAddressId: 0,
    remark: '',
    totalPrice: 0,
    submitting: false,
    showAddressForm: false,
    addressForm: {
      name: '',
      phone: '',
      province: '',
      city: '',
      district: '',
      detail: ''
    }
  },

  onLoad() {
    this.loadCartItems()
    this.loadAddresses()
  },

  async loadCartItems() {
    const cartIds = my.getStorageSync({ key: 'checkoutCartIds' }).data || []
    if (cartIds.length === 0) {
      my.showToast({ content: '请选择商品' })
      setTimeout(() => my.navigateBack(), 1500)
      return
    }

    try {
      const items = await app.request({ url: '/cart', showLoading: false })
      const selected = (items || []).filter(item => cartIds.includes(item.id))
      let total = 0
      selected.forEach(item => {
        let price = item.product?.price || 0
        if (item.spec) price += item.spec.price_adjustment || 0
        total += price * item.quantity
      })
      this.setData({ cartItems: selected, totalPrice: total.toFixed(2) })
    } catch (err) {
      console.error('加载商品失败', err)
    }
  },

  async loadAddresses() {
    try {
      const addresses = await app.request({ url: '/addresses', showLoading: false })
      const list = addresses || []
      let selectedId = 0
      if (list.length > 0) {
        const def = list.find(a => a.is_default)
        selectedId = def ? def.id : list[0].id
      }
      this.setData({ addresses: list, selectedAddressId: selectedId })
    } catch (err) {
      console.error('加载地址失败', err)
    }
  },

  onSelectAddress(e) {
    this.setData({ selectedAddressId: Number(e.target.dataset.id) })
  },

  onRemarkInput(e) {
    this.setData({ remark: e.detail.value })
  },

  onShowAddressForm() {
    this.setData({ showAddressForm: true })
  },

  onHideAddressForm() {
    this.setData({ showAddressForm: false })
  },

  onAddressInput(e) {
    const field = e.target.dataset.field
    this.setData({ [`addressForm.${field}`]: e.detail.value })
  },

  async onSaveAddress() {
    const f = this.data.addressForm
    if (!f.name || !f.phone || !f.province || !f.city || !f.detail) {
      my.showToast({ content: '请填写完整地址' })
      return
    }
    try {
      const addr = await app.request({
        url: '/addresses',
        method: 'POST',
        data: { ...f, is_default: this.data.addresses.length === 0 }
      })
      await this.loadAddresses()
      if (addr && addr.id) {
        this.setData({ selectedAddressId: addr.id })
      }
      this.setData({
        showAddressForm: false,
        addressForm: { name: '', phone: '', province: '', city: '', district: '', detail: '' }
      })
      my.showToast({ content: '地址已保存' })
    } catch (err) {
      console.error('保存地址失败', err)
    }
  },

  async onSubmitOrder() {
    if (this.data.submitting) return

    if (this.data.cartItems.length === 0) {
      my.showToast({ content: '请选择商品' })
      return
    }
    if (!this.data.selectedAddressId) {
      my.showToast({ content: '请选择收货地址' })
      return
    }

    this.setData({ submitting: true })

    try {
      const cartIds = this.data.cartItems.map(item => item.id)
      const order = await app.request({
        url: '/orders',
        method: 'POST',
        data: {
          address_id: this.data.selectedAddressId,
          cart_ids: cartIds,
          remark: this.data.remark
        }
      })

      my.removeStorageSync({ key: 'checkoutCartIds' })
      await this.triggerPayment(order)
    } catch (err) {
      console.error('创建订单失败', err)
      this.setData({ submitting: false })
    }
  },

  async triggerPayment(order) {
    try {
      const result = await app.request({
        url: `/orders/${order.id}/prepay`,
        method: 'POST',
        data: { pay_method: 'alipay' }
      })

      if (!result.need_pay) {
        my.showToast({ content: '支付成功' })
        setTimeout(() => {
          my.switchTab({ url: '/pages/order/index' })
        }, 1500)
        return
      }

      const tradeNO = result.pay_params.tradeNO
      my.tradePay({
        tradeNO: tradeNO,
        success: (res) => {
          if (res.resultCode === '9000') {
            my.showToast({ content: '支付成功' })
            setTimeout(() => {
              my.switchTab({ url: '/pages/order/index' })
            }, 1500)
          } else {
            my.showToast({ content: '待支付，可在订单中继续付款' })
            setTimeout(() => {
              my.switchTab({ url: '/pages/order/index' })
            }, 2000)
          }
        },
        fail: () => {
          my.showToast({ content: '待支付，可在订单中继续付款' })
          setTimeout(() => {
            my.switchTab({ url: '/pages/order/index' })
          }, 2000)
        }
      })
    } catch (err) {
      console.error('发起支付失败', err)
      my.showToast({ content: '待支付，可在订单中继续付款' })
      setTimeout(() => {
        my.switchTab({ url: '/pages/order/index' })
      }, 2000)
    } finally {
      this.setData({ submitting: false })
    }
  }
})
