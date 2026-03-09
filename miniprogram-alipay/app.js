App({
  globalData: {
    apiBase: 'https://ring-api.itsyourturn.top/api/v1',
    token: null,
    userInfo: null
  },

  onLaunch() {
    const token = my.getStorageSync({ key: 'token' }).data
    if (token) {
      this.globalData.token = token
    }
  },

  request(options) {
    const { url, method = 'GET', data, showLoading = true } = options

    return new Promise((resolve, reject) => {
      if (showLoading) {
        my.showLoading({ content: '加载中...' })
      }

      my.request({
        url: this.globalData.apiBase + url,
        method,
        data,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': this.globalData.token ? `Bearer ${this.globalData.token}` : ''
        },
        success: (res) => {
          if (showLoading) my.hideLoading()

          if (res.data.code === 0) {
            resolve(res.data.data)
          } else if (res.status === 401) {
            this.globalData.token = null
            my.removeStorageSync({ key: 'token' })
            my.showToast({ content: '请先登录' })
            reject(new Error('未登录'))
          } else {
            my.showToast({ content: res.data.message || '请求失败' })
            reject(new Error(res.data.message))
          }
        },
        fail: (err) => {
          if (showLoading) my.hideLoading()
          my.showToast({ content: '网络错误' })
          reject(err)
        }
      })
    })
  },

  setToken(token) {
    this.globalData.token = token
    my.setStorageSync({ key: 'token', data: token })
  },

  logout() {
    this.globalData.token = null
    this.globalData.userInfo = null
    my.removeStorageSync({ key: 'token' })
  },

  checkLogin() {
    return !!this.globalData.token
  },

  alipayLogin() {
    return new Promise((resolve, reject) => {
      my.getAuthCode({
        scopes: 'auth_user',
        success: (res) => {
          this.request({
            url: '/auth/alipay-login',
            method: 'POST',
            data: { code: res.authCode }
          }).then(data => {
            this.setToken(data.token)
            this.globalData.userInfo = data.user
            resolve(data)
          }).catch(reject)
        },
        fail: reject
      })
    })
  }
})
