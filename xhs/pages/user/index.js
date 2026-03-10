const app = getApp()

Page({
  data: {
    userInfo: null,
    isLogin: false,
    avatarUrl: '',
    nickname: '',
    userInitial: 'U'
  },

  onShow() {
    this.checkLoginStatus()
  },

  checkLoginStatus() {
    const isLogin = app.checkLogin()
    const userInfo = app.globalData.userInfo
    let avatarUrl = ''
    let nickname = ''
    let userInitial = 'U'

    if (userInfo) {
      avatarUrl = (userInfo.avatar && userInfo.avatar !== '') ? userInfo.avatar : ''
      nickname = userInfo.nickname || ''
      const displayName = nickname || userInfo.username || ''
      userInitial = displayName ? displayName[0].toUpperCase() : 'U'
    }

    this.setData({
      isLogin,
      userInfo,
      avatarUrl,
      nickname,
      userInitial
    })
  },

  onLogin() {
    app.xhsLogin(true).then(() => {
      this.checkLoginStatus()
    }).catch(err => {
      console.error('登录失败', err)
    })
  },

  onChooseAvatar() {
    xhs.chooseImage({
      count: 1,
      sizeType: ['compressed'],
      sourceType: ['album', 'camera'],
      success: (res) => {
        const tempUrl = res.tempFilePaths[0]
        if (!tempUrl) return

        this.setData({ avatarUrl: tempUrl })

        xhs.uploadFile({
          url: app.globalData.apiBase + '/upload/image',
          filePath: tempUrl,
          name: 'file',
          header: {
            'Authorization': app.globalData.token ? `Bearer ${app.globalData.token}` : ''
          },
          success: (uploadRes) => {
            try {
              const data = JSON.parse(uploadRes.data)
              if (data.code === 0 && data.data && data.data.url) {
                const imageUrl = data.data.url
                this.setData({ avatarUrl: imageUrl })
                app.updateProfile(this.data.nickname, imageUrl).then(() => {
                  xhs.showToast({ title: '头像已更新', icon: 'success' })
                })
              } else {
                xhs.showToast({ title: '上传失败', icon: 'none' })
              }
            } catch (err) {
              console.error('解析上传结果失败', err)
              xhs.showToast({ title: '上传失败', icon: 'none' })
            }
          },
          fail: (err) => {
            console.error('上传头像失败', err)
            xhs.showToast({ title: '上传失败', icon: 'none' })
          }
        })
      }
    })
  },

  onNicknameInput(e) {
    this.setData({ nickname: e.detail.value })
  },

  onNicknameBlur(e) {
    const nickname = e.detail.value
    if (!nickname || nickname === (this.data.userInfo && this.data.userInfo.nickname)) return

    this.setData({ nickname })

    app.updateProfile(nickname, this.data.avatarUrl).then(() => {
      xhs.showToast({ title: '昵称已更新', icon: 'success' })
    }).catch(err => {
      console.error('更新昵称失败', err)
    })
  },

  onLogout() {
    xhs.showModal({
      title: '提示',
      content: '确定要退出登录吗？',
      success: (res) => {
        if (res.confirm) {
          app.logout()
          this.setData({
            isLogin: false,
            userInfo: null,
            avatarUrl: '',
            nickname: '',
            userInitial: 'U'
          })
        }
      }
    })
  },

  onContact() {
    xhs.makePhoneCall({
      phoneNumber: '400-000-0000'
    })
  }
})
