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
    app.alipayLogin().then(() => {
      this.checkLoginStatus()
    }).catch(err => {
      console.error('登录失败', err)
    })
  },

  onChooseAvatar() {
    my.chooseImage({
      count: 1,
      success: (res) => {
        const tempPath = res.apFilePaths[0]
        this.setData({ avatarUrl: tempPath })

        // 上传头像到服务器，再更新 profile
        my.uploadFile({
          url: app.globalData.apiBase + '/upload/image',
          fileType: 'image',
          fileName: 'file',
          filePath: tempPath,
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
                  my.showToast({ content: '头像已更新' })
                })
              }
            } catch (e) {
              console.error('上传头像失败', e)
            }
          },
          fail: (err) => {
            console.error('上传头像失败', err)
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
      my.showToast({ content: '昵称已更新' })
    }).catch(err => {
      console.error('更新昵称失败', err)
    })
  },

  onLogout() {
    my.confirm({
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
    my.makePhoneCall({ number: '400-000-0000' })
  }
})
