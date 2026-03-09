App({
  globalData: {
    apiBase: 'https://ring-api.itsyourturn.top/api/v1',
    token: null,
    userInfo: null
  },

  onLaunch() {
    // 从缓存读取token
    const token = wx.getStorageSync('token')
    if (token) {
      this.globalData.token = token
    }
  },

  // API请求封装
  request(options) {
    const { url, method = 'GET', data, showLoading = true } = options

    return new Promise((resolve, reject) => {
      if (showLoading) {
        wx.showLoading({ title: '加载中...' })
      }

      wx.request({
        url: this.globalData.apiBase + url,
        method,
        data,
        header: {
          'Content-Type': 'application/json',
          'Authorization': this.globalData.token ? `Bearer ${this.globalData.token}` : ''
        },
        success: (res) => {
          if (showLoading) wx.hideLoading()

          if (res.data.code === 0) {
            resolve(res.data.data)
          } else if (res.statusCode === 401) {
            // 登录过期
            this.globalData.token = null
            wx.removeStorageSync('token')
            wx.showToast({ title: '请先登录', icon: 'none' })
            reject(new Error('未登录'))
          } else {
            wx.showToast({ title: res.data.message || '请求失败', icon: 'none' })
            reject(new Error(res.data.message))
          }
        },
        fail: (err) => {
          if (showLoading) wx.hideLoading()
          wx.showToast({ title: '网络错误', icon: 'none' })
          reject(err)
        }
      })
    })
  },

  // 设置登录token
  setToken(token) {
    this.globalData.token = token
    wx.setStorageSync('token', token)
  },

  // 登出
  logout() {
    this.globalData.token = null
    this.globalData.userInfo = null
    wx.removeStorageSync('token')
  },

  // 检查登录状态
  checkLogin() {
    return !!this.globalData.token
  },

  // 微信登录
  wxLogin() {
    return new Promise((resolve, reject) => {
      wx.login({
        success: (res) => {
          if (res.code) {
            // 发送code到后端换取token
            this.request({
              url: '/auth/wechat-login',
              method: 'POST',
              data: { code: res.code }
            }).then(data => {
              this.setToken(data.token)
              this.globalData.userInfo = data.user
              resolve(data)
            }).catch(reject)
          } else {
            reject(new Error('登录失败'))
          }
        },
        fail: reject
      })
    })
  }
})
