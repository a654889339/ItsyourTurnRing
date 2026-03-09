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
    app.alipayLogin().then(data => {
      this.setData({
        isLogin: true,
        userInfo: data.user
      })
    })
  },

  onLogout() {
    my.confirm({
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          app.logout()
          this.setData({ isLogin: false, userInfo: null })
        }
      }
    })
  },

  onContact() {
    my.makePhoneCall({ number: '400-000-0000' })
  }
})
