App({
  globalData: {
    apiBase: 'http://106.54.50.88:5102/api/v1',
    token: null,
    userInfo: null
  },

  onLaunch() {
    const token = wx.getStorageSync('token')
    const userInfo = wx.getStorageSync('userInfo')
    if (token) {
      this.globalData.token = token
    }
    if (userInfo) {
      this.globalData.userInfo = userInfo
    }
    if (!token) {
      this.wxLogin(false)
    }
  },

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
            this.globalData.token = null
            this.globalData.userInfo = null
            wx.removeStorageSync('token')
            wx.removeStorageSync('userInfo')
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

  setToken(token) {
    this.globalData.token = token
    wx.setStorageSync('token', token)
  },

  setUserInfo(user) {
    this.globalData.userInfo = user
    wx.setStorageSync('userInfo', user)
  },

  logout() {
    this.globalData.token = null
    this.globalData.userInfo = null
    wx.removeStorageSync('token')
    wx.removeStorageSync('userInfo')
  },

  checkLogin() {
    return !!this.globalData.token
  },

  wxLogin(showError) {
    return new Promise((resolve, reject) => {
      wx.login({
        success: (res) => {
          if (res.code) {
            this.request({
              url: '/auth/wechat-login',
              method: 'POST',
              data: { code: res.code },
              showLoading: false
            }).then(data => {
              this.setToken(data.token)
              this.setUserInfo(data.user)
              resolve(data)
            }).catch(err => {
              console.error('微信登录失败', err)
              if (showError) {
                wx.showToast({ title: '登录失败', icon: 'none' })
              }
              reject(err)
            })
          } else {
            reject(new Error('获取code失败'))
          }
        },
        fail: reject
      })
    })
  },

  updateProfile(nickname, avatarUrl) {
    return this.request({
      url: '/auth/update-profile',
      method: 'POST',
      data: { nickname, avatar: avatarUrl }
    }).then(user => {
      this.setUserInfo(user)
      return user
    })
  }
})
