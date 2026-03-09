const app = getApp()

Page({
  data: {
    userInfo: null,
    isLogin: false
  },

  onShow() {
    this.checkLoginStatus()
  },

  checkLoginStatus() {
    const isLogin = app.checkLogin()
    this.setData({
      isLogin,
      userInfo: app.globalData.userInfo
    })
  },

  onLogin() {
    app.wxLogin().then(data => {
      this.setData({
        isLogin: true,
        userInfo: data.user
      })
    }).catch(err => {
      console.error('登录失败', err)
    })
  },

  onLogout() {
    wx.showModal({
      title: '提示',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          app.logout()
          this.setData({
            isLogin: false,
            userInfo: null
          })
        }
      }
    })
  },

  onContact() {
    wx.makePhoneCall({
      phoneNumber: '400-000-0000'
    })
  }
})
