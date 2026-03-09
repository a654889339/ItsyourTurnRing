import { defineStore } from 'pinia'
import { authAPI } from '../api'

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem('token') || null,
    user: null
  }),

  getters: {
    isAuthenticated: (state) => !!state.token,
    currentUser: (state) => state.user
  },

  actions: {
    async login(username, password) {
      try {
        const result = await authAPI.login({ username, password })
        this.token = result.token
        this.user = result.user
        localStorage.setItem('token', result.token)
        return result
      } catch (error) {
        throw error
      }
    },

    async register(data) {
      try {
        const result = await authAPI.register(data)
        this.token = result.token
        this.user = result.user
        localStorage.setItem('token', result.token)
        return result
      } catch (error) {
        throw error
      }
    },

    async fetchCurrentUser() {
      try {
        const user = await authAPI.getCurrentUser()
        this.user = user
        return user
      } catch (error) {
        this.logout()
        throw error
      }
    },

    logout() {
      this.token = null
      this.user = null
      localStorage.removeItem('token')
    }
  }
})
